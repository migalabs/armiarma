package hosts

import (
	"context"
	"sync"
	"time"

	"github.com/migalabs/armiarma/pkg/db/models"
	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
	"github.com/migalabs/armiarma/pkg/utils"

	"github.com/libp2p/go-libp2p/core/network"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	log "github.com/sirupsen/logrus"
)

/*
	File that includes the methods to set the custom modification channels for the Libp2p host
*/

type IdentificationEvent struct {
	HostInfo  *models.HostInfo
	Timestamp time.Time // Timestamp of when was the attempt done
}

func (c *BasicLibp2pHost) standardListenF(net network.Network, addr ma.Multiaddr) {
	log.Trace("Listen")
}

func (c *BasicLibp2pHost) standardListenCloseF(net network.Network, addr ma.Multiaddr) {
	log.Trace("Close listen")
}

func (c *BasicLibp2pHost) standardConnectF(net network.Network, conn network.Conn) {
	// get timestamp fo the event
	t := time.Now()

	log.WithFields(log.Fields{
		"EVENT":     "Connection detected",
		"DIRECTION": conn.Stat().Direction.String(),
	}).Debug("Peer: ", conn.RemotePeer().String())

	// Only locate new IP if the connection is "Inbound"
	// if it's outbound - we should already have it in the DB
	if conn.Stat().Direction.String() == "Inbound" {
		ip := utils.ExtractIPFromMAddr(conn.RemoteMultiaddr()).String()
		c.IpLocator.LocateIP(ip)
	}

	// since se only have one multiaddress, gen the array
	mAddrs := make([]ma.Multiaddr, 0)
	mAddrs = append(mAddrs, conn.RemoteMultiaddr())

	// create new HostInfo
	hInfo := models.NewHostInfo(
		conn.RemotePeer(),
		c.NetworkNode.Network(),
		models.WithMultiaddress(mAddrs),
	)

	// Aggregate timeout context for the different
	mainCtx, cancel := context.WithTimeout(c.Ctx(), 5*time.Second)
	defer cancel()
	// set sync group and error groups to handle different reqresps
	var wg sync.WaitGroup

	// request the Host Metadata
	h := c.Host()

	// for Eth2
	var bStatus common.Status
	var bMetadata common.MetaData

	var hinfoErr error
	var statusErr, metadataErr error

	wg.Add(1)
	go ReqHostInfo(mainCtx, &wg, h, c.IpLocator, conn, hInfo, &hinfoErr)

	switch c.NetworkNode.(type) {
	case (*eth.LocalEthereumNode):
		ethNet := c.NetworkNode.(*eth.LocalEthereumNode)
		// request BeaconStatus metadata as we connect to a peer
		wg.Add(1)
		go ethNet.ReqBeaconStatus(mainCtx, &wg, h, conn.RemotePeer(), &bStatus, &statusErr)
		// request the BeaconMetadata
		wg.Add(1)
		go ethNet.ReqBeaconMetadata(mainCtx, &wg, h, conn.RemotePeer(), &bMetadata, &metadataErr)
	default:
	}

	wg.Wait()
	// Parse the errors from the different go routines,
	// if there wasn't anything in the channel, or if the err is nil fetch peer info
	if hinfoErr != nil {
		// if error, cancel the timeout and stop ReqMetadata and ReqStatus
		log.WithFields(log.Fields{
			"ERROR": hinfoErr.Error(),
		}).Debug("ReqHostInfo Peer: ", conn.RemotePeer().String())
	} else {
		log.Debug("peer identified, succeed")
	}

	// If the network was eth2, wait for the metadata echange to reply
	switch c.NetworkNode.(type) {
	case (*eth.LocalEthereumNode):
		// Beacon Status reqresp error check
		// if there is an error  in the channel, print error
		if statusErr != nil {
			log.WithFields(log.Fields{
				"ERROR": statusErr.Error(),
			}).Debug("ReqStatus Peer: ", conn.RemotePeer().String())
		} else {
			log.Debug("peer status req, succeed", bStatus)
			hInfo.AddAtt("beacon-status", eth.NewBeaconStatus(conn.RemotePeer(), bStatus))
		}
		// // Beacon Metadata reqresp error check
		// // if if there is an error  in the channel, print error
		if metadataErr != nil {
			log.WithFields(log.Fields{
				"ERROR": metadataErr.Error(),
			}).Debug("ReqMetadata Peer: ", conn.RemotePeer().String())
		} else {
			log.Debug("peer metadata req, succeed", bMetadata)
			hInfo.AddAtt("beaconmetadata", eth.NewBeaconMetadata(conn.RemotePeer(), bMetadata))
		}
	default:
	}

	identStat := IdentificationEvent{
		HostInfo:  hInfo,
		Timestamp: t,
	}

	// Add info about identification into connEvent
	connEvent := &models.ConnInfo{
		Direction:  models.ConnDirection(conn.Stat().Direction), // Se share the same int8 structure as network.Direction
		ConnTime:   t,
		Latency:    hInfo.PeerInfo.Latency,
		Identified: hInfo.IsHostIdentified(),
		Error:      hinfoErr.Error(),
	}

	// Record the connectino event
	c.RecConnEvent(&models.EventTrace{
		PeerID: conn.RemotePeer(),
		Event:  connEvent,
	})

	// Send the new connection status
	c.RecIdentEvent(identStat)
}

func (c *BasicLibp2pHost) standardDisconnectF(net network.Network, conn network.Conn) {
	t := time.Now()
	log.WithFields(log.Fields{
		"EVENT":     "Disconnection detected",
		"DIRECTION": conn.Stat().Direction.String(),
	}).Debug("Peer: ", conn.RemotePeer().String())
	// compose the disconnection event
	disconEvent := &models.EndConnInfo{
		DiscTime: t,
	}
	// Send the new disconnection status
	c.RecConnEvent(&models.EventTrace{
		PeerID: conn.RemotePeer(),
		Event:  disconEvent,
	})
}

func (c *BasicLibp2pHost) standardOpenedStreamF(net network.Network, str network.Stream) {
	log.Trace("Open Stream")
}

func (c *BasicLibp2pHost) standardClosedF(net network.Network, str network.Stream) {
	log.Trace("Close Stream")
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
