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
	log.Debug("Listen")
}

func (c *BasicLibp2pHost) standardListenCloseF(net network.Network, addr ma.Multiaddr) {
	log.Debug("Close listen")
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

	log.WithFields(logrus.Fields{
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

	// Error channels
	hinfoErr := make(chan error, 1)

	// Request the Host Metadata
	h := c.Host()

	// for Eth2
	var bStatus common.Status
	var bMetadata common.MetaData
	metadataErr := make(chan error, 1)
	statusErr := make(chan error, 1)

	wg.Add(1)
	go ReqHostInfo(mainCtx, &wg, h, c.IpLocator, conn, &peer, hinfoErr)

	if c.Network == "eth2" {

		// Request the BeaconMetadata
		wg.Add(1)
		go ReqBeaconMetadata(mainCtx, &wg, h, conn.RemotePeer(), &bMetadata, metadataErr)

		// request BeaconStatus metadata as we connect to a peer
		wg.Add(1)
		go ReqBeaconStatus(mainCtx, &wg, h, conn.RemotePeer(), &bStatus, statusErr)
	}

	wg.Wait()
	// Parse the errors from the different go routines,

	// check if an error was sent into the channel,
	// if there wasn't anything in the channel, or if the err is nil fetch peer info
	// if if there is an error  in the channel, print error
	if err, _ := <-hinfoErr; err != nil {
		// if error, cancel the timeout and stop ReqMetadata and ReqStatus
		log.WithFields(logrus.Fields{
			"ERROR": err.Error(),
		}).Debug("ReqHostInfo Peer: ", conn.RemotePeer().String())
	} else {
		log.Debug("peer identified, succeed")
	}

	// If the network was eth2, wait for the metadata echange to reply
	if c.Network == "eth2" {
		// Beacon Status request error check
		// if if there is an error  in the channel, print error
		if err := <-metadataErr; err != nil {
			log.WithFields(logrus.Fields{
				"ERROR": err.Error(),
			}).Debug("ReqMetadata Peer: ", conn.RemotePeer().String())
		} else {
			log.Debug("peer metadata req, succeed")
			peer.SetAtt("beaconmetadata", models.NewBeaconMetadata(bMetadata))
		}

		// Beacon Status request error check
		// if if there is an error  in the channel, print error
		if err := <-statusErr; err != nil {
			log.WithFields(logrus.Fields{
				"ERROR": err.Error(),
			}).Debug("ReqStatus Peer: ", conn.RemotePeer().String())
		} else {
			log.Debug("peer status req, succeed")
			peer.SetAtt("beaconstatus", models.NewBeaconStatus(bStatus))
		}
		close(metadataErr)
		close(statusErr)
	}

	// close channels
	close(hinfoErr)

	log.Debug("sending identification event of peer", peer.PeerId)
	identStat := IdentificationEvent{
		Peer:      peer,
		Timestamp: t,
	}

	// Send the new connection status
	c.RecIdentEvent(identStat)
}

func (c *BasicLibp2pHost) standardDisconnectF(net network.Network, conn network.Conn) {
	log.Debugf("disconnected from peer %s", conn.RemotePeer().String())
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
	log.Debug("Open Stream")
}

func (c *BasicLibp2pHost) standardClosedF(net network.Network, str network.Stream) {
	log.Debug("Close")
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
