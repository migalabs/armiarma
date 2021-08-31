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
	logrus.WithFields(logrus.Fields{
		"EVENT": "Connection detected",
		"DIRECTION": fmtDirection(conn.Stat().Direction),
	}).Info("Peer: ", conn.RemotePeer().String())

	// create the peer
	peer := metrics.NewPeer(conn.RemotePeer().String())

	// try to request metadata for the peer
	peerData, err := PollPeerMetadata(conn.RemotePeer(), c.Base, c.PeerMetadataState, c.Store, c.PeerStore)

	// double check that the peerData is not empty
	if err == nil && peerData.String() != "no data available" {
		peer = fetchPeerExtraInfo(peerData)
		logrus.WithFields(logrus.Fields{
			"EVENT": "Metadata request OK",
		}).Info("Peer: ", conn.RemotePeer().String())
	} else {
		logrus.WithFields(logrus.Fields{
			"EVENT": "Metadata request NOK",
		}).Info("Peer: ", conn.RemotePeer().String())
	}

	// store the peer and record that the connection was ok
	c.PeerStore.StoreOrUpdatePeer(peer)
	c.PeerStore.ConnectionEvent(conn.RemotePeer().String(), fmtDirection(conn.Stat().Direction))

	// End of metric traces to track the connections and disconnections
	c.Log.WithFields(logrus.Fields{
		"event": "connection_open", "peer": conn.RemotePeer().String(),
		"direction": fmtDirection(conn.Stat().Direction),
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
	ip := utils.GetIpFromMultiAddress(address)

	country, city, err := utils.GetLocationFromIp(ip)
	if err != nil {
		log.Warn("error when fetching country/city from ip", err)
	}

	peer := metrics.NewPeer(peerData.PeerID.String())
	peer.NodeId = peerData.NodeID.String()
	peer.UserAgent = peerData.UserAgent
	peer.ClientName = client
	peer.ClientVersion = version
	peer.ClientOS = "TODO"
	peer.Pubkey = peerData.Pubkey
	peer.Addrs = address
	peer.Ip = ip
	peer.Country = country
	peer.City = city
	peer.Latency = float64(peerData.Latency/time.Millisecond) / 1000

	return peer
}
