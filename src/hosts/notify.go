package hosts

import (
	"context"
	"time"

	eth_node "github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/utils"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

/*
	File that includes the methods to set the custom modification channels for the Libp2p host
*/

// ConnectionStatus
// * It is the struct that compiles the data of a received connection from the host
// * The struct will be shared between peering and strategy.
type ConnectionStatus struct {
	Peer      db.Peer   // TODO: right now just sending the entire info about the peer, (recheck after Peer struct subdivision)
	Timestamp time.Time // Timestamp of when was the attempt done
	// TODO: More things to add in te future
}

// DisconnectionStatus
// * It is the struct that compiles the data of a detected disconnection from the host
// * The struct will be shared between peering and strategy.
type DisconnectionStatus struct {
	Peer      db.Peer
	Timestamp time.Time // Timestamp of when was the attempt done
}

func (c *BasicLibp2pHost) standardListenF(net network.Network, addr ma.Multiaddr) {
	c.Log.Debug("Listen")
}

func (c *BasicLibp2pHost) standardListenCloseF(net network.Network, addr ma.Multiaddr) {
	c.Log.Debug("Close listen")
}

func (c *BasicLibp2pHost) standardConnectF(net network.Network, conn network.Conn) {
	c.Log.WithFields(logrus.Fields{
		"EVENT":     "Connection detected",
		"DIRECTION": conn.Stat().Direction.String(),
	}).Debug("Peer: ", conn.RemotePeer().String())
	// Generate new peer to aggregate new data
	peer := db.NewPeer(conn.RemotePeer().String())
	h := c.Host()
	t := time.Now()
	// Request the Host Metadata
	err := ReqHostInfo(context.Background(), h, conn, &peer)
	if err != nil {
		peer.MetadataSucceed = false
		c.Log.WithFields(logrus.Fields{
			"ERROR": err,
		}).Warn("Peer: ", conn.RemotePeer().String())
	}
	// Request the BeaconMetadata
	bMetadata, err := ReqBeaconMetadata(context.Background(), h, conn.RemotePeer())
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"ERROR": err,
		}).Debug("Peer: ", conn.RemotePeer().String())
	} else {
		peer.UpdateBeaconMetadata(bMetadata)
	}
	// request BeaconStatus metadata as we connect to a peer
	bStatus, err := ReqBeaconStatus(context.Background(), h, conn.RemotePeer())
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"ERROR": err,
		}).Debug("Peer: ", conn.RemotePeer().String())
	} else {
		peer.UpdateBeaconStatus(bStatus)
	}

	// Read ENR of the Peer from the generated enode
	n, err := c.PeerStore.GetENR(conn.RemotePeer().String())
	if err != nil || n == nil {
		n.ID().String()
		// We can only get the node.ID if the ENR of the peer was already in the PeerStore fromt dv5
		peer.NodeId = n.ID().String()
	} else {
		edcsakey, err := utils.ParsePubkey(peer.Pubkey)
		if err != nil {
			c.Log.WithFields(logrus.Fields{
				"ERROR": "parsing pubkey",
			}).Debug("Peer: ", conn.RemotePeer().String())
		}
		peer.NodeId = eth_node.PubkeyToIDV4(edcsakey).String()
	}

	// Add new connection event
	peer.ConnectionEvent(conn.Stat().Direction.String(), t)
	connStat := ConnectionStatus{
		Peer:      peer,
		Timestamp: t,
	}
	// Send the new connection status
	c.RecNewConn(connStat)
}

func (c *BasicLibp2pHost) standardDisconnectF(net network.Network, conn network.Conn) {
	c.Log.Debugf("disconnected from peer %s", conn.RemotePeer().String())
	peer := db.NewPeer(conn.RemotePeer().String())
	t := time.Now()
	peer.DisconnectionEvent(t)
	disconnStat := ConnectionStatus{
		Peer:      peer,
		Timestamp: t,
	}
	// Send the new disconnection status
	c.RecNewConn(disconnStat)
}

func (c *BasicLibp2pHost) standardOpenedStreamF(net network.Network, str network.Stream) {
	c.Log.Debug("Open Stream")
}

func (c *BasicLibp2pHost) standardClosedF(net network.Network, str network.Stream) {
	c.Log.Debug("Close")
}

//
func (c *BasicLibp2pHost) SetCustomNotifications() error {
	// generate empty bundle to set custom notifiers
	bundle := &network.NotifyBundle{
		ListenF:       c.standardListenF,
		ListenCloseF:  c.standardListenCloseF,
		ConnectedF:    c.standardConnectF,
		DisconnectedF: c.standardDisconnectF,
		OpenedStreamF: c.standardOpenedStreamF,
		ClosedStreamF: c.standardClosedF,
	}
	// read host from main struct
	h := c.Host()
	// set the custom notifiers to the host
	h.Network().Notify(bundle)
	return nil
}
