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
	"github.com/migalabs/armiarma/src/db/models"
	"github.com/migalabs/armiarma/src/rpc/methods"
	"github.com/migalabs/armiarma/src/rpc/reqresp"
	"github.com/migalabs/armiarma/src/utils"
	"github.com/migalabs/armiarma/src/utils/apis"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/protolambda/zrnt/eth2/beacon/common"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
)

var ()

// ReqBeaconStatus:
// Function that opens a new Stream from the given host to send a RPC requesting the BeaconStatus of the given peer.ID.
// Returns the BeaconStatus of the given peer if succeed, error if failed.
// @param ctx: parent context.
// @param wg: wait group to notify when function has been done.
// @param h: host to use to connect.
// @param peerID: who to connect to.
// @param stat: where to deserialize the content of the beacon status: output.
// @param finErr: error channel.
func ReqBeaconStatus(ctx context.Context, wg *sync.WaitGroup, h host.Host, peerID peer.ID, stat *common.Status, finErr chan error) {
	defer wg.Done()
	// Generate the compression
	comp := reqresp.SnappyCompression{}
	// Generate the Server Error Code
	var resCode reqresp.ResponseCode // error by default
	// record the error into the error channel
	finErr <- methods.StatusRPCv1.RunRequest(ctx, h.NewStream, peerID, comp,
		reqresp.RequestSSZInput{Obj: stat}, 1,
		func() error {
			return nil
		},
		func(chunk reqresp.ChunkedResponseHandler) error {
			resCode = chunk.ResultCode()
			switch resCode {
			case reqresp.ServerErrCode, reqresp.InvalidReqCode:
				msg, err := chunk.ReadErrMsg()
				if err != nil {
					return err
				}
				return errors.Errorf("error requesting BeaconStatus RPC: %s", msg)
			case reqresp.SuccessCode:
				if err := chunk.ReadObj(stat); err != nil {
					return errors.Wrap(err, "from requesting BeaconMetadata RPC")
				}
			default:
				return errors.New("unexpected result code for BeaconStatus RPC request")
			}
			return nil
		})
}

// ReqBeaconMetadata:
// Function that opens a new Stream from the given host to send a RPC requesting the BeaconStatus of the given peer.ID.
// Returns the BeaconStatus of the given peer if succeed, error if failed.
// @param ctx: parent context.
// @param wg: wait group to notify when function has been done.
// @param h: host to use to connect.
// @param peerID: who to connect to.
// @param meta: where to deserialize the content of the beacon metadata: output.
// @param finErr: error channel.
func ReqBeaconMetadata(ctx context.Context, wg *sync.WaitGroup, h host.Host, peerID peer.ID, meta *common.MetaData, finErr chan error) {
	defer wg.Done()
	// Generate the compression
	comp := reqresp.SnappyCompression{}
	// Generate the Server Error Code
	var resCode reqresp.ResponseCode // error by default
	// record the error into the error channel
	finErr <- methods.MetaDataRPCv1.RunRequest(ctx, h.NewStream, peerID, comp, reqresp.RequestSSZInput{Obj: nil}, 1,
		func() error {
			return nil
		},
		func(chunk reqresp.ChunkedResponseHandler) error {
			resCode = chunk.ResultCode()
			switch resCode {
			case reqresp.ServerErrCode, reqresp.InvalidReqCode:
				msg, err := chunk.ReadErrMsg()
				if err != nil {
					return errors.Errorf("error requesting BeaconMetadata RPC: %s", msg)
				}
			case reqresp.SuccessCode:
				if err := chunk.ReadObj(meta); err != nil {
					return errors.Wrap(err, "from requesting BeaconMetadata RPC")
				}
			default:
				return errors.New("unexpected result code for BeaconMetadata RPC request")
			}
			return nil
		})
}

// Identify the peer from the Libp2p Identify Service

type HostWithIDService interface {
	IDService() *identify.IDService
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

func ReqHostInfo(ctx context.Context, wg *sync.WaitGroup, h host.Host, ipLoc *apis.PeerLocalizer, conn network.Conn, peer *models.Peer, errIdent chan error) {
	defer wg.Done()

	peerID := conn.RemotePeer()

	var finErr error
	// Identify Peer to access main data
	// convert host to IDService
	withIdentify, ok := h.(HostWithIDService)
	if !ok {
		errIdent <- errors.Errorf("host does not support libp2p identify protocol")
		return
	}
	t := time.Now()
	idService := withIdentify.IDService()
	if idService == nil {
		errIdent <- errors.Errorf("libp2p identify not enabled on this host")
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
		errIdent <- errors.Errorf("identification error caused by timed out")
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
		errIdent <- errors.Wrap(err, "unable to compose the maddrs")
		fmt.Println("unable to extract multiaddrs")
		return
	}
	// generate array of MAddr to fit the models.Peer struct
	mAddrs := make([]ma.Multiaddr, 0)
	mAddrs = append(mAddrs, multiAddr)
	peer.MAddrs = mAddrs

	// Update, request location from a peer only when the connection is inbound
	// if the connection is outbound, we already had the IP located from the ENR
	/*
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
	*/
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
	errIdent <- finErr
}
