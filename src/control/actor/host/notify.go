package host

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	log.WithFields(logrus.Fields{
		"EVENT":     "Connection detected",
		"DIRECTION": conn.Stat().Direction.String(),
	}).Info("Peer: ", conn.RemotePeer().String())
	// Generate new peer to aggregate new data
	peer := metrics.NewPeer(conn.RemotePeer().String())
	h, _ := c.Host()
	// Request the Host Metadata
	err := ReqHostInfo(context.Background(), h, conn, &peer)
	if err != nil {
		peer.MetadataSucceed = false
		log.WithFields(logrus.Fields{
			"ERROR": err,
		}).Warn("Peer: ", conn.RemotePeer().String())
	}
	// Request the BeaconMetadata
	bMetadata, err := ReqBeaconMetadata(context.Background(), h, conn.RemotePeer())
	if err != nil {
		log.WithFields(logrus.Fields{
			"ERROR": err,
		}).Warn("Peer: ", conn.RemotePeer().String())
	} else {
		peer.UpdateBeaconMetadata(bMetadata)
	}
	// request BeaconStatus metadata as we connect to a peer
	bStatus, err := ReqBeaconStatus(context.Background(), h, conn.RemotePeer())
	if err != nil {
		log.WithFields(logrus.Fields{
			"ERROR": err,
		}).Warn("Peer: ", conn.RemotePeer().String())
	} else {
		peer.UpdateBeaconStatus(bStatus)
	}
	// Read ENR of the Peer from the generated enode
	n := c.Store.LatestENR(conn.RemotePeer())
	if n != nil {
		n.ID().String()
		// We can only get the node.ID if the ENR of the peer was already in the PeerStore fromt dv5
		peer.NodeId = n.ID().String()
	} else {
		// TODO: If the peer wasn't discovered via dv5 "n" will be empty
		log.WithFields(logrus.Fields{
			"ERROR": "Peer ENR not found",
		}).Warn("Peer: ", conn.RemotePeer().String())
	}
	// Add new connection event
	peer.ConnectionEvent(conn.Stat().Direction.String(), time.Now())
	// Add new peer or aggregate info to existing peer
	c.PeerStore.StoreOrUpdatePeer(peer)
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
