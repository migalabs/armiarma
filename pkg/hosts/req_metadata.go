package hosts

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/rpc/methods"
	"github.com/migalabs/armiarma/pkg/rpc/reqresp"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/migalabs/armiarma/pkg/utils/apis"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/protolambda/zrnt/eth2/beacon/common"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
)

var ()

// ReqBeaconStatus:
// Function that opens a new Stream from the given host to send a RPC reqresping the BeaconStatus of the given peer.ID.
// Returns the BeaconStatus of the given peer if succeed, error if failed.
// @param ctx: parent context.
// @param wg: wait group to notify when function has been done.
// @param h: host to use to connect.
// @param peerID: who to connect to.
// @param stat: where to deserialize the content of the beacon status: output.
// @param finErr: error channel.
func ReqBeaconStatus(ctx context.Context, wg *sync.WaitGroup, h host.Host, peerID peer.ID, resultStatus *common.Status, finErr *error) {

	defer wg.Done()

	frkDgst := new(common.ForkDigest)
	b, err := hex.DecodeString("4a26c58b")
	if err != nil {
		log.Panic("unable to decode ForkDigest", err.Error())
	}
	frkDgst.UnmarshalText(b)

	status := &common.Status{
		ForkDigest:     *frkDgst,
		FinalizedRoot:  common.Root{},
		FinalizedEpoch: 0,
		HeadRoot:       common.Root{},
		HeadSlot:       0,
	}
	var resCode reqresp.ResponseCode // error by default
	err = methods.StatusRPCv1.RunRequest(ctx, h.NewStream, peerID, new(reqresp.SnappyCompression),
		reqresp.RequestSSZInput{Obj: status}, 1,
		func() error {
			return nil
		},
		func(chunk reqresp.ChunkedResponseHandler) error {
			resCode = chunk.ResultCode()
			switch resCode {
			case reqresp.ServerErrCode, reqresp.InvalidReqCode:
				msg, err := chunk.ReadErrMsg()
				if err != nil {
					return fmt.Errorf("%s: %w", msg, err)
				}
			case reqresp.SuccessCode:
				var stat common.Status
				if err := chunk.ReadObj(&stat); err != nil {
					return err
				}
				fmt.Println(stat)
				*resultStatus = stat
			default:
				return errors.New("unexpected result code")
			}
			return nil
		})
	*finErr = err
}

// ReqBeaconMetadata:
// Function that opens a new Stream from the given host to send a RPC reqresping the BeaconStatus of the given peer.ID.
// Returns the BeaconStatus of the given peer if succeed, error if failed.
// @param ctx: parent context.
// @param wg: wait group to notify when function has been done.
// @param h: host to use to connect.
// @param peerID: who to connect to.
// @param meta: where to deserialize the content of the beacon metadata: output.
// @param finErr: error channel.
func ReqBeaconMetadata(ctx context.Context, wg *sync.WaitGroup, h host.Host, peerID peer.ID, meta *common.MetaData, finErr *error) {
	defer wg.Done()
	// Generate the Server Error Code
	var resCode reqresp.ResponseCode // error by default
	// record the error into the error channel
	err := methods.MetaDataRPCv1.RunRequest(ctx, h.NewStream, peerID, reqresp.SnappyCompression{}, reqresp.RequestSSZInput{Obj: nil}, 1,
		func() error {
			return nil
		},
		func(chunk reqresp.ChunkedResponseHandler) error {
			resCode = chunk.ResultCode()
			switch resCode {
			case reqresp.ServerErrCode, reqresp.InvalidReqCode:
				msg, err := chunk.ReadErrMsg()
				if err != nil {
					return errors.Errorf("error reqresping BeaconMetadata RPC: %s", msg)
				}
			case reqresp.SuccessCode:
				if err := chunk.ReadObj(meta); err != nil {
					return errors.Wrap(err, "from reqresping BeaconMetadata RPC")
				}
			default:
				return errors.New("unexpected result code for BeaconMetadata RPC reqresp")
			}
			return nil
		})
	*finErr = err
}

// Identify the peer from the Libp2p Identify Service

type HostWithIDService interface {
	IDService() identify.IDService
}

// ReqBeaconMetadata:
// ReqHostInfo returns the basic host information regarding a given peer, from the libp2p perspective
// it aggregates the info from the libp2p Identify protocol adding some extra info such as RTT between local host and remote peer
// return empty struct and error if failure on the identify process.
// Returns the BeaconStatus of the given peer if succeed, error if failed.
// @param ctx: parent context.
// @param wg: wait group to notify when function has been done.
// @param h: host to use to connect.
// @param peerID: who to connect to.
// @param meta: where to deserialize the content of the beacon metadata: output.
// @param finErr: error channel.

func ReqHostInfo(ctx context.Context, wg *sync.WaitGroup, h host.Host, ipLoc *apis.PeerLocalizer, conn network.Conn, peer *models.Peer, errIdent *error) {
	defer wg.Done()

	peerID := conn.RemotePeer()

	var finErr error
	// Identify Peer to access main data
	// convert host to IDService
	withIdentify, ok := h.(HostWithIDService)
	if !ok {
		finErr = errors.Errorf("host does not support libp2p identify protocol")
		*errIdent = finErr
		return
	}
	t := time.Now()
	idService := withIdentify.IDService()
	if idService == nil {
		finErr = errors.Errorf("libp2p identify not enabled on this host")
		*errIdent = finErr
		return
	}
	peer.MetadataRequest = true
	var rtt time.Duration
	select {
	case <-idService.IdentifyWait(conn):
		peer.MetadataSucceed = true
		peer.LastIdentifyTimestamp = time.Now()
		rtt = time.Since(t)
	case <-ctx.Done():
		finErr = errors.Errorf("identification error caused by timed out")
		*errIdent = finErr
		return
	}
	// Fill the the metrics
	// Into the new peer to fetch
	pv, err := h.Peerstore().Get(peerID, "ProtocolVersion")
	if err == nil {
		peer.ProtocolVersion = pv.(string)
	}

	prot, err := h.Peerstore().GetProtocols(peerID)
	if err == nil {
		peer.Protocols = prot
	}
	// Update the values of the
	peer.Latency = float64(rtt/time.Millisecond) / 1000
	peer.PeerId = peerID.String()
	multiAddrStr := conn.RemoteMultiaddr().String() + "/p2p/" + peerID.String()
	multiAddr, err := ma.NewMultiaddr(multiAddrStr)
	if err != nil {
		finErr = errors.Wrap(err, "unable to compose the maddrs")
		*errIdent = finErr
		fmt.Println("unable to extract multiaddrs")
		return
	}
	// generate array of MAddr to fit the models.Peer struct
	mAddrs := make([]ma.Multiaddr, 0)
	mAddrs = append(mAddrs, multiAddr)
	peer.MAddrs = mAddrs

	// Update, reqresp location from a peer only when the connection is inbound
	// if the connection is outbound, we already had the IP located from the ENR

	if conn.Stat().Direction.String() == "Inbound" {
		peer.Ip = utils.ExtractIPFromMAddr(multiAddr).String()
		locResp, err := ipLoc.LocateIP(peer.Ip)
		if err != nil {
			// TODO: think about a better idea to integrate a logger into this functions
			log.Warnf("error when fetching country/city from ip %s. %s", peer.Ip, err.Error())
		} else {
			peer.Country = locResp.Country
			peer.City = locResp.City
			peer.CountryCode = locResp.CountryCode
		}
	}

	peer.Ip = utils.ExtractIPFromMAddr(multiAddr).String()

	// locResp, err := ipLoc.LocateIP(peer.Ip)
	// if err != nil {
	// 	// TODO: think about a better idea to integrate a logger into this functions
	// 	log.Warnf("error when fetching country/city from ip %s. %s", peer.Ip, err.Error())
	// } else {
	// 	peer.Country = locResp.Country
	// 	peer.City = locResp.City
	// 	peer.CountryCode = locResp.CountryCode
	// }

	// Fulfill the hInfo struct
	ua, err := h.Peerstore().Get(peerID, "AgentVersion")
	if err == nil {
		peer.UserAgent = ua.(string)
		// Extract Client type and version
		peer.ClientName, peer.ClientVersion = utils.FilterClientType(peer.UserAgent)
		peer.ClientOS = "TODO"
	} else {
		// EDGY CASE: when peers refuse the connection, the callback gets called and the identify protocol
		// returns an empty struct (we are unable to identify them)
		finErr = errors.Errorf("unable to identify peer")
		peer.MetadataSucceed = false
	}
	pubk, err := conn.RemotePublicKey().Raw()
	if err == nil {
		peer.SetAtt("pubkey", hex.EncodeToString(pubk))
	}
	// return the erro defined in the top
	// nil if we could identify it, ident error if we couldnt line 181
	*errIdent = finErr
}
