package host

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/control/actor/peer/metadata"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/metrics/utils"
	"github.com/protolambda/rumor/p2p/track"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type HostNotifyCmd struct {
	*base.Base
	*metrics.PeerStore
	*metadata.PeerMetadataState
	Store track.ExtendedPeerstore
}

func (c *HostNotifyCmd) Help() string {
	return "Get notified of specific events, as long as the command runs."
}

func (c *HostNotifyCmd) HelpLong() string {
	return `
Args: <event-types>...
Network event notifications.
Valid event types:
 - listen (listen_open listen_close)
 - connection (connection_open connection_close)
 - stream (stream_open stream_close)
 - all
Notification logs will have keys: "event" - one of the above detailed event types, e.g. listen_close.
- "peer": peer ID
- "direction": "inbound"/"outbound"/"unknown", for connections and streams
- "extra": stream/connection extra data
- "protocol": protocol ID for streams.
`
}

func (c *HostNotifyCmd) listenF(net network.Network, addr ma.Multiaddr) {
	c.Log.WithFields(logrus.Fields{"event": "listen_open", "addr": addr.String()}).Debug("opened network listener")
}

func (c *HostNotifyCmd) listenCloseF(net network.Network, addr ma.Multiaddr) {
	c.Log.WithFields(logrus.Fields{"event": "listen_close", "addr": addr.String()}).Debug("closed network listener")
}

func (c *HostNotifyCmd) connectedF(net network.Network, conn network.Conn) {
	logrus.WithFields(logrus.Fields{
		"EVENT":     "Connection detected",
		"DIRECTION": conn.Stat().Direction.String(),
	}).Info("Peer: ", conn.RemotePeer().String())
	// try to request metadata for the peer
	peerData, err := PollPeerMetadata(conn.RemotePeer(), c.Base, c.PeerMetadataState, c.Store, c.PeerStore)
	// request BeaconStatus metadata as we connect to a peer
	h, err := c.Host()
	if err != nil {
		log.Error(err)
	}
	bState, err2 := nodemetadata.ReqBeaconStatus(context.Background(), h, conn.RemotePeer())
	if err2 != nil {
		log.Warn(err)
	}
	logrus.Info(&bState)
	var peer metrics.Peer
	// TODO: so far, If we didn't manage to exchange metadata, we asume that the peer didn't exchange Status neither
	// 		  IMPORTANT: Metadata doen't include any kind of info related to the Host/Node
	// 				     It just includes SeqNumber, Attnets
	if err == nil {
		peer = fetchPeerExtraInfo(peerData, bState)	
		// temp
		log.Info(peer)
	} else {
		log.Info("could not get metadata for peer: ", conn.RemotePeer().String(), " err: ", err)
		// TODO: Add also IP taken from ENR of discovered peers
		peer = metrics.Peer{
			PeerId: conn.RemotePeer().String(),
		}
	}
	ma := conn.RemoteMultiaddr().String()
	fmt.Println("Multiaddress of the peer", ma)
	peer, err := fetchPeerExtraInfo(peerData, ma)
	if err != nil {
		log.Error("Could not fetch peer data for: " + conn.RemotePeer().String())
	}
	logrus.WithFields(logrus.Fields{
		"EVENT": "Metadata request OK",
	}).Info("Peer: ", conn.RemotePeer().String())
	// store the peer and record that the connection was ok
	c.PeerStore.StoreOrUpdatePeer(peer)
	c.PeerStore.ConnectionEvent(conn.RemotePeer().String(), conn.Stat().Direction.String())
	// End of metric traces to track the connections and disconnections
	c.Log.WithFields(logrus.Fields{
		"event": "connection_open", "peer": conn.RemotePeer().String(),
		"direction": conn.Stat().Direction.String(),
	}).Debug("new peer connection")
}

func (c *HostNotifyCmd) disconnectedF(net network.Network, conn network.Conn) {
	logrus.WithFields(logrus.Fields{
		"EVENT": "Disconnection detected",
	}).Info("Peer: ", conn.RemotePeer().String())
	c.PeerStore.DisconnectionEvent(conn.RemotePeer().String())
	// End of metric traces to track the connections and disconnections
	c.Log.WithFields(logrus.Fields{
		"event": "connection_close", "peer": conn.RemotePeer().String(),
		"direction": conn.Stat().Direction.String(),
	}).Debug("peer disconnected")
}

func (c *HostNotifyCmd) openedStreamF(net network.Network, str network.Stream) {
	c.Log.WithFields(logrus.Fields{
		"event": "stream_open", "peer": str.Conn().RemotePeer().String(),
		"direction": str.Stat().Direction.String(),
		"protocol":  str.Protocol(),
	}).Debug("opened stream")
}

func (c *HostNotifyCmd) closedStreamF(net network.Network, str network.Stream) {
	c.Log.WithFields(logrus.Fields{
		"event": "stream_close", "peer": str.Conn().RemotePeer().String(),
		"direction": str.Stat().Direction.String(),
		"protocol":  str.Protocol(),
	}).Debug("closed stream")
}

func (c *HostNotifyCmd) Run(ctx context.Context, args ...string) error {
	h, err := c.Host()
	if err != nil {
		return err
	}
	bundle := &network.NotifyBundle{}
	for _, notifyType := range args {
		notifyType = strings.TrimSpace(notifyType)
		if notifyType == "" {
			continue
		}
		switch notifyType {
		case "listen_open":
			bundle.ListenF = c.listenF
		case "listen_close":
			bundle.ListenCloseF = c.listenCloseF
		case "connection_open":
			bundle.ConnectedF = c.connectedF
		case "connection_close":
			bundle.DisconnectedF = c.disconnectedF
		case "stream_open":
			bundle.OpenedStreamF = c.openedStreamF
		case "stream_close":
			bundle.ClosedStreamF = c.closedStreamF
		case "listen":
			bundle.ListenF = c.listenF
			bundle.ListenCloseF = c.listenCloseF
		case "connection":
			bundle.ConnectedF = c.connectedF
			bundle.DisconnectedF = c.disconnectedF
		case "stream":
			bundle.OpenedStreamF = c.openedStreamF
			bundle.ClosedStreamF = c.closedStreamF
		case "all":
			bundle.ListenF = c.listenF
			bundle.ListenCloseF = c.listenCloseF
			bundle.ConnectedF = c.connectedF
			bundle.DisconnectedF = c.disconnectedF
			bundle.OpenedStreamF = c.openedStreamF
			bundle.ClosedStreamF = c.closedStreamF
		default:
			return fmt.Errorf("unrecognized notification type: %s", notifyType)
		}
	}
	h.Network().Notify(bundle)
	c.Control.RegisterStop(func(ctx context.Context) error {
		h.Network().StopNotify(bundle)
		return nil
	})
	return nil
}

/* Dismissed since stream.Stat().Direction includes method .String()
func fmtDirection(d network.Direction) string {
	switch d {
	case network.DirInbound:
		return "inbound"
	case network.DirOutbound:
		return "outbound"
	case network.DirUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}
*/

// Convert from rumor PeerAllData to our Peer. Note that
// some external data is fetched and some fields are parsed
<<<<<<< HEAD
func fetchPeerExtraInfo(peerData *track.PeerAllData, addr string) (metrics.Peer, error) {
	client, version := utils.FilterClientType(peerData.UserAgent)
	// Obtaining the IP or MultiAddrs from the ENR increases the chances to get a Restringed IP
	// therefore, this one can be obtained directly from the libp2p connection stream
	ip := strings.Split(addr, "/")[2]
	country, city, err := utils.GetLocationFromIp(ip)
=======
func fetchPeerExtraInfo(peerData *track.PeerAllData, bStatus beacon.Status) metrics.Peer {
	client, version := utils.FilterClientType(peerData.UserAgent)
	// TODO: temporary fix untill IP clear IP is adjusted to database
	// 		 GetIpAndLocationFromAddrs should be deprecated to GetLocationFromIP
	address := utils.GetFullAddress(peerData.Addrs)
	ip, country, city, err := utils.GetIpAndLocationFromAddrs(address)
>>>>>>> Add Beacon.Status to Peer + ReqBeaconStatus()
	if err != nil {
		return metrics.Peer{}, errors.Wrap(err, "could not get location from ip")
	}
<<<<<<< HEAD
	peer := metrics.NewPeer(peerData.PeerID.String())
	peer.NodeId = peerData.NodeID.String()
	peer.UserAgent = peerData.UserAgent
	peer.ClientName = client
	peer.ClientVersion = version
	peer.ClientOS = "TODO"
	peer.Pubkey = peerData.Pubkey
	peer.Addrs = addr
	peer.Ip = ip
	peer.Country = country
	peer.City = city
	peer.Latency = float64(peerData.Latency/time.Millisecond) / 1000

	return peer, nil
=======

	peer := metrics.Peer{
		PeerId:        peerData.PeerID.String(),
		NodeId:        peerData.NodeID.String(),
		UserAgent:     peerData.UserAgent,
		ClientName:    client,
		ClientVersion: version,
		ClientOS:      "TODO",
		Pubkey:        peerData.Pubkey,
		Addrs:         address,
		Ip:            ip,
		Country:       country,
		City:          city,
		Latency:       float64(peerData.Latency/time.Millisecond) / 1000,
	}
	peer.UpdateBeaconStatus(bStatus)
	return peer
>>>>>>> Add Beacon.Status to Peer + ReqBeaconStatus()
}
