package host

import (
	"context"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/control/actor/peer/metadata"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/track"
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
	logrus.Info("connection detected: ", conn.RemotePeer().String())
	h, _ := c.Host()
	// Request the Host Metadata
	hInfo := ReqHostInfo(context.Background(), h, conn)
	// Request the BeaconMetadata
	bMetadata, err := ReqBeaconMetadata(context.Background(), h, conn.RemotePeer())
	if err != nil {
		log.Warn(err)
	}
	// request BeaconStatus metadata as we connect to a peer
	bStatus, err := ReqBeaconStatus(context.Background(), h, conn.RemotePeer())
	if err != nil {
		log.Warn(err)
	}
	var peer metrics.Peer
	// Read ENR of the Peer from the generated enode
	n := c.Store.LatestENR(conn.RemotePeer())
	if n.ID().String() == "" {
		// TODO: If the peer wasn't discovered via dv5 "n" will be empty
		// we could compose new ENR with the beacon metadata and the host info
		log.Warn("Peer ENR not found")
	}
	hInfo.NodeID = n.ID().String()
	// fetch all the info gathered from the peer into a new Peer struct
	peer.FetchHostInfo(hInfo)
	// TODO: Switch from Update ot Add in the future, when implementing metrics versions
	peer.UpdateBeaconStatus(bStatus)
	peer.UpdateBeaconMetadata(bMetadata)
	// So far, just act like if we got new info, Update or Aggregate new info from a peer already on the Peerstore
	c.PeerStore.AddPeer(peer)
	c.PeerStore.AddConnectionEvent(conn.RemotePeer().String(), "Connection")

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
<<<<<<< HEAD
=======

// DEPRECATE for Libp2p2 network.Direction.String() https://github.com/libp2p/go-libp2p-core/blob/094b0d3f8ba2934339cb35e1a875b11ab6d08839/network/network.go#L38
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
>>>>>>> Organize Beacon Info of peer + Add new methods to fetch info from a peer + Adding test for fetching
