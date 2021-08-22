package host

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/protolambda/rumor/metrics/utils"
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

	// request metadata as soon as we connect to a peer
	peerData, err := PollPeerMetadata(conn.RemotePeer(), c.Base, c.PeerMetadataState, c.Store, c.PeerStore)
	var peer metrics.Peer
	if err == nil {
		peer = fetchPeerExtraInfo(peerData)
		_ = peer
	} else {
		log.Info("could not get metadata for peer: ", conn.RemotePeer().String(), " err: ", err)
		// TODO: Add also IP taken from ENR of discovered peers
		peer = metrics.Peer {
			PeerId: conn.RemotePeer().String(),
		}
	}
	c.PeerStore.AddPeer(peer)
	c.PeerStore.AddConnectionEvent(conn.RemotePeer().String(), "Connection")

	// End of metric traces to track the connections and disconnections
	c.Log.WithFields(logrus.Fields{
		"event": "connection_open", "peer": conn.RemotePeer().String(),
		"direction": fmtDirection(conn.Stat().Direction),
	}).Debug("new peer connection")
}

func (c *HostNotifyCmd) disconnectedF(net network.Network, conn network.Conn) {
	c.PeerStore.AddConnectionEvent(conn.RemotePeer().String(), "Disconnection")
	logrus.Info("disconnection detected", conn.RemotePeer().String())
	// End of metric traces to track the connections and disconnections
	c.Log.WithFields(logrus.Fields{
		"event": "connection_close", "peer": conn.RemotePeer().String(),
		"direction": fmtDirection(conn.Stat().Direction),
	}).Debug("peer disconnected")
}

func (c *HostNotifyCmd) openedStreamF(net network.Network, str network.Stream) {
	c.Log.WithFields(logrus.Fields{
		"event": "stream_open", "peer": str.Conn().RemotePeer().String(),
		"direction": fmtDirection(str.Stat().Direction),
		"protocol":  str.Protocol(),
	}).Debug("opened stream")
}

func (c *HostNotifyCmd) closedStreamF(net network.Network, str network.Stream) {
	c.Log.WithFields(logrus.Fields{
		"event": "stream_close", "peer": str.Conn().RemotePeer().String(),
		"direction": fmtDirection(str.Stat().Direction),
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

// Convert from rumor PeerAllData to our Peer. Note that
// some external data is fetched and some fields are parsed
func fetchPeerExtraInfo(peerData *track.PeerAllData) metrics.Peer {
	client, version := utils.FilterClientType(peerData.UserAgent)
	address := utils.GetFullAddress(peerData.Addrs)

	ip, country, city, err := utils.GetIpAndLocationFromAddrs(address)
	if err != nil {
		log.Error("error when fetching country/city from ip", err)
	}

	peer := metrics.Peer {
		PeerId: peerData.PeerID.String(),
		NodeId: peerData.NodeID.String(),
		UserAgent: peerData.UserAgent,
		ClientName: client,
		ClientVersion: version,
		ClientOS: "TODO",
		Pubkey: peerData.Pubkey,
		Addrs: address,
		Ip: ip,
		Country: country,
		City: city,
		Latency: float64(peerData.Latency/time.Millisecond) / 1000,
	}

	return peer
}
