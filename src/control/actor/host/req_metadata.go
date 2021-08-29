package host

import (
	"context"
	"fmt"
	"time"
	"github.com/pkg/errors"
	
	log "github.com/sirupsen/logrus"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/protolambda/rumor/p2p/rpc/methods"
	"github.com/protolambda/rumor/p2p/rpc/reqresp"
	"github.com/protolambda/zrnt/eth2/beacon"

	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/libp2p/go-libp2p-core/network"
)

var (
	timeout time.Duration = 7 * time.Second
)

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
				return errors.New(fmt.Sprintf("error requesting BeaconState RPC: %s", msg))
			case reqresp.SuccessCode:
				var stat beacon.Status
				if err := chunk.ReadObj(&stat); err != nil {
					return err
				}
				data = stat
			default:
				return errors.New("unexpected result code")
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
	err = methods.MetaDataRPCv1.RunRequest(ctx, h.NewStream, peerID, comp, reqresp.RequestSSZInput{Obj: &data}, 1,
		func() error {
			fmt.Println("Error here?")
			return nil
		},
		func(chunk reqresp.ChunkedResponseHandler) error {
			resCode = chunk.ResultCode()
			switch resCode {
			case reqresp.ServerErrCode, reqresp.InvalidReqCode:
				msg, err := chunk.ReadErrMsg()
				if err != nil {
					return errors.New(fmt.Sprintf("error requesting BeaconMetadata RPC: %s", msg))
				}
			case reqresp.SuccessCode:
				var meta beacon.MetaData
				if err := chunk.ReadObj(&meta); err != nil {
					return err
				}
				data = meta
			default:
				return errors.New("unexpected result code")
			}
			return nil
		})
	return
}

// basic Host info that will be requested from the identification of a libp2p peer
type BasicHostInfo struct {
	TimeStamp time.Time
	// Peer Host/Node Info
	PeerID string
	UserAgent string
	ProtocolVersion string
	Addrs string
	PubKey string
	RTT time.Duration
	Protocols []string
	// Information regarding the metadata exchange
	Direction string
	// Metadata requested
	MetadataRequest bool
	MetadataSucceed bool 
}

type HostWithIDService interface {
	IDService() *identify.IDService
}

// request the host infomartion regarding a given peer, from the libp2p perspective
// return empty struct and error if failure 
// TODO: So far, both request, ping request and identify request has been deplyed on the same func
// 		 RTT can be also measured from identify request (a bit les accurate) which can leave us to remove the ping request
// 		 Still leaving it there for Understanding purposes. Some Clients don't support /ipfs/ping/1.0.0, but all support "/eth2/beacon_chain/req/ping/1/" instead
// DISCUSS the if the returning error should just be asociated to the identify request [Currently removed since we always have few info about the peer]
func ReqHostInfo(ctx context.Context, h host.Host , conn network.Conn) BasicHostInfo{
	// time out for ping
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	tout := timeoutCtx.Done()
	defer cancel()
	peerID := conn.RemotePeer()

	var hInfo BasicHostInfo
	hInfo.TimeStamp = time.Now()
	
	// Identify Peer to access main data
	// convert host to IDService
	withIdentify, ok := h.(HostWithIDService)
	if !ok {
		log.Error("host does not support libp2p identify protocol")
		return hInfo
	}
	t := time.Now()
	idService := withIdentify.IDService()
	if idService == nil {
		log.Error("libp2p identify not enabled on this host")
		return hInfo
	}
	hInfo.MetadataRequest = true
	var rtt time.Duration
	select {
		case <-idService.IdentifyWait(conn):
			hInfo.MetadataSucceed = true
			rtt = time.Since(t)
			log.Info("completed identification")
		case <-tout:
			log.Info("awaiting identification timed out")
	}
	// Fulfill the hInfo struct
	ua, err := h.Peerstore().Get(peerID, "AgentVersion")
	if err == nil {
		hInfo.UserAgent = ua.(string)
	}
	pv, err := h.Peerstore().Get(peerID, "ProtocolVersion")
	if err == nil {
		hInfo.ProtocolVersion = pv.(string)
	} 
	pubk, err := conn.RemotePublicKey().Raw()
	if err == nil {
		hInfo.PubKey = string(pubk)
	}
	prot, err := h.Peerstore().GetProtocols(peerID)
	if err == nil {
		hInfo.Protocols = prot
	}

	hInfo.RTT = rtt
	hInfo.PeerID = peerID.String()
	hInfo.Addrs = conn.RemoteMultiaddr().String()
	hInfo.Direction = conn.Stat().Direction.String()
	
	return hInfo
}