package hosts

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/migalabs/armiarma/src/db"
	db_utils "github.com/migalabs/armiarma/src/db/utils"
	"github.com/migalabs/armiarma/src/utils"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
)

var (
	timeout time.Duration = 7 * time.Second
)

/*
// Function that opens a new Stream from the given host to send a RPC requesting the BeaconStatus of the given peer.ID
// Returns the BeaconStatus of the given peer if succeed, error if failed
func ReqBeaconStatus(ctx context.Context, h host.Host, peerID peer.ID) (data beacon.Status, err error) {
	// Generate the compression
	comp := reqresp.SnappyCompression{}
	// Generate the Server Error Code
	var resCode reqresp.ResponseCode // error by default
	err = methods.StatusRPCv1.RunRequest(ctx, h.NewStream, peerID, comp,
		reqresp.RequestSSZInput{Obj: &data}, 1,
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
				var stat beacon.Status
				if err := chunk.ReadObj(&stat); err != nil {
					return errors.Wrap(err, "from requesting BeaconMetadata RPC")
				}
				data = stat
			default:
				return errors.New("unexpected result code for BeaconStatus RPC request")
			}
			return nil
		})
	return
}

func ReqBeaconMetadata(ctx context.Context, h host.Host, peerID peer.ID) (data beacon.MetaData, err error) {
	// Generate the compression
	comp := reqresp.SnappyCompression{}
	// Generate the Server Error Code
	var resCode reqresp.ResponseCode // error by default
	err = methods.MetaDataRPCv1.RunRequest(ctx, h.NewStream, peerID, comp, reqresp.RequestSSZInput{Obj: nil}, 1,
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
				var meta beacon.MetaData
				if err := chunk.ReadObj(&meta); err != nil {
					return errors.Wrap(err, "from requesting BeaconMetadata RPC")
				}
				data = meta
			default:
				return errors.New("unexpected result code for BeaconMetadata RPC request")
			}
			return nil
		})
	return
}
*/
// Identify the peer from the Libp2p Identify Service

type HostWithIDService interface {
	IDService() *identify.IDService
}

// ReqHostInfo returns the basic host information regarding a given peer, from the libp2p perspective
// it aggregates the info from the libp2p Identify protocol adding some extra info such as RTT between local host and remote peer
// return empty struct and error if failure on the identify process
func ReqHostInfo(ctx context.Context, h host.Host, conn network.Conn, peer *db.Peer) (err_ident error) {
	// time out for ping
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	tout := timeoutCtx.Done()
	defer cancel()
	peerID := conn.RemotePeer()

	// Identify Peer to access main data
	// convert host to IDService
	withIdentify, ok := h.(HostWithIDService)
	if !ok {
		return errors.Errorf("host does not support libp2p identify protocol")
	}
	t := time.Now()
	idService := withIdentify.IDService()
	if idService == nil {
		return errors.Errorf("libp2p identify not enabled on this host")
	}
	peer.MetadataRequest = true
	var rtt time.Duration
	select {
	case <-idService.IdentifyWait(conn):
		peer.MetadataSucceed = true
		rtt = time.Since(t)
	case <-tout:
		err_ident = errors.Errorf("identification error caused by timed out")
	}
	var err error
	/* Not defined yet on the Peer struct
	pv, err := h.Peerstore().Get(peerID, "ProtocolVersion")
	if err == nil {
		hInfo.ProtocolVersion = pv.(string)
	}
	prot, err := h.Peerstore().GetProtocols(peerID)
	if err == nil {
		//hInfo.Protocols = prot
		log.Infof("peer on protocol %s", prot)
	}
	*/
	// Update the values of the
	peer.Latency = float64(rtt/time.Millisecond) / 1000
	peer.PeerId = peerID.String()
	peer.ConnectedDirection = conn.Stat().Direction.String()

	multiAddrStr := conn.RemoteMultiaddr().String() + "/p2p/" + peerID.String()
	multiAddr, err := ma.NewMultiaddr(multiAddrStr)
	if err != nil {
		return fmt.Errorf("error composing the maddrs from peer", err)
	}
	// generate array of MAddr to fit the db.Peer struct
	mAddrs := make([]ma.Multiaddr, 0)
	mAddrs = append(mAddrs, multiAddr)
	peer.MAddrs = mAddrs
	peer.Ip = utils.ExtractIPFromMAddr(multiAddr).String()
	if err != nil {
		// Almost impossible, when we are connected to a peer, we will always have a complete Multiaddrs after the Identify req
		// leaving it emtpy to spot the problem, IP-Api request already makes a parse of the IP before making server petition
		// TODO: think about a better idea to integrate a logger into this functions
		//log.Error(err)
	}
	peer.Country, peer.City, peer.CountryCode, err = db_utils.GetLocationFromIp(peer.Ip)
	if err != nil {
		// TODO: think about a better idea to integrate a logger into this functions
		//log.Error("error when fetching country/city from ip", err)
	}
	// Fulfill the hInfo struct
	ua, err := h.Peerstore().Get(peerID, "AgentVersion")
	if err == nil {
		peer.UserAgent = ua.(string)
		// Extract Client type and version
		peer.ClientName, peer.ClientVersion = db_utils.FilterClientType(peer.UserAgent)
		peer.ClientOS = "TODO"
	} else {
		// EDGY CASE: when peers refuse the connection, the callback gets called and the identify protocol
		// returns an empty struct (we are unable to identify them)
		err_ident = errors.Errorf("identification error caused by connection refuse")
	}
	pubk, err := conn.RemotePublicKey().Raw()
	if err == nil {
		peer.Pubkey = hex.EncodeToString(pubk)
	}
	return err_ident
}
