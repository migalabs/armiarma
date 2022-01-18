package hosts

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/migalabs/armiarma/src/db/models"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/sirupsen/logrus"
)

/*
	File that includes the methods to set the custom modification channels for the Libp2p host
*/

// ConnectionEvent
// It is the struct that compiles the data of a received a connection from the host
// The struct will be shared between peering and strategy.
type ConnectionEvent struct {
	ConnType  int8        // (0)-Empty (1)-Connection (2)-Disconnection
	Peer      models.Peer // TODO: right now just sending the entire info about the peer, (recheck after Peer struct subdivision)
	Timestamp time.Time   // Timestamp of when was the attempt done
	// TODO: More things to add in te future
}

type IdentificationEvent struct {
	Peer      models.Peer // TODO: right now just sending the entire info about the peer, (recheck after Peer struct subdivision)
	Timestamp time.Time   // Timestamp of when was the attempt done
}

func (c *BasicLibp2pHost) standardListenF(net network.Network, addr ma.Multiaddr) {
	Log.Debug("Listen")
}

func (c *BasicLibp2pHost) standardListenCloseF(net network.Network, addr ma.Multiaddr) {
	Log.Debug("Close listen")
}

func (c *BasicLibp2pHost) standardConnectF(net network.Network, conn network.Conn) {

	// Add new connection event
	// ConnectionEvent and Disconnection
	t := time.Now()
	p := models.NewPeer(conn.RemotePeer().String())
	// TODO: Add the connection ID to the connection and disconnection events
	p.ConnectionEvent(conn.Stat().Direction.String(), t)
	cs := ConnectionEvent{
		ConnType:  int8(1),
		Peer:      p,
		Timestamp: t,
	}

	c.RecConnEvent(cs)

	Log.WithFields(logrus.Fields{
		"EVENT":     "Connection detected",
		"DIRECTION": conn.Stat().Direction.String(),
	}).Debug("Peer: ", conn.RemotePeer().String())

	//  Make synchrony among the different requests that we have to do

	// identify everything that we can about the peer
	// not adding the connection 2 times
	peer := models.NewPeer(conn.RemotePeer().String())

	// Aggregate timeout context for the different
	mainCtx, cancel := context.WithTimeout(c.Ctx(), 5*time.Second)
	defer cancel()
	// set sync group and error groups to handle different requests
	var wg sync.WaitGroup

	wg.Add(3) // there will be 3

	// Error channels
	hinfoErr := make(chan error, 1)
	metadataErr := make(chan error, 1)
	statusErr := make(chan error, 1)

	// Request the Host Metadata
	h := c.Host()
	go ReqHostInfo(mainCtx, &wg, h, c.IpLocator, conn, &peer, hinfoErr)

	var bMetadata common.MetaData
	// Request the BeaconMetadata
	go ReqBeaconMetadata(mainCtx, &wg, h, conn.RemotePeer(), &bMetadata, metadataErr)

	var bStatus common.Status
	// request BeaconStatus metadata as we connect to a peer
	go ReqBeaconStatus(mainCtx, &wg, h, conn.RemotePeer(), &bStatus, statusErr)

	wg.Wait()
	// Parse the errors from the different go routines,

	// check if an error was sent into the channel,
	// if there wasn't anything in the channel, or if the err is nil fetch peer info
	// if if there is an error  in the channel, print error
	if err, _ := <-hinfoErr; err != nil {
		// if error, cancel the timeout and stop ReqMetadata and ReqStatus
		Log.WithFields(logrus.Fields{
			"ERROR": err.Error(),
		}).Debug("ReqHostInfo Peer: ", conn.RemotePeer().String())
	} else {
		Log.Debug("peer identified, succeed")
	}

	// Beacon Status request error check
	// if if there is an error  in the channel, print error
	if err := <-metadataErr; err != nil {
		Log.WithFields(logrus.Fields{
			"ERROR": err.Error(),
		}).Debug("ReqMetadata Peer: ", conn.RemotePeer().String())
	} else {
		Log.Debug("peer metadata req, succeed")
		peer.UpdateBeaconMetadata(bMetadata)
	}

	// Beacon Status request error check
	// if if there is an error  in the channel, print error
	if err := <-statusErr; err != nil {
		Log.WithFields(logrus.Fields{
			"ERROR": err.Error(),
		}).Debug("ReqStatus Peer: ", conn.RemotePeer().String())
	} else {
		Log.Debug("peer status req, succeed")
		peer.UpdateBeaconStatus(bStatus)
	}
	// close channels
	close(hinfoErr)
	close(metadataErr)
	close(statusErr)

	Log.Debug("sending identification event of peer", peer.PeerId)
	identStat := IdentificationEvent{
		Peer:      peer,
		Timestamp: t,
	}

	// Send the new connection status
	c.RecIdentEvent(identStat)

}

func (c *BasicLibp2pHost) standardDisconnectF(net network.Network, conn network.Conn) {
	Log.Debugf("disconnected from peer %s", conn.RemotePeer().String())
	// TEMP

	peer := models.NewPeer(conn.RemotePeer().String())
	t := time.Now()
	peer.DisconnectionEvent(t)
	disconnStat := ConnectionEvent{
		ConnType:  int8(2),
		Peer:      peer,
		Timestamp: t,
	}
	// Send the new disconnection status
	c.RecConnEvent(disconnStat)
}

func (c *BasicLibp2pHost) standardOpenedStreamF(net network.Network, str network.Stream) {
	Log.Debug("Open Stream")
}

func (c *BasicLibp2pHost) standardClosedF(net network.Network, str network.Stream) {
	Log.Debug("Close")
}

// SetCustomNotifications:
// Set all notification handlers
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
