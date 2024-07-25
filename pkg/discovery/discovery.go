package discovery

import (
	"context"
	"net"
	"sync"
	"time"

//	psql "github.com/migalabs/armiarma/pkg/db/postgresql"
	psql "github.com/hdser/armiarma/pkg/db/redshift"

	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/migalabs/armiarma/pkg/utils/apis"

	"github.com/migalabs/armiarma/pkg/db/models"
	log "github.com/sirupsen/logrus"
)

/*
This file implements the discovery5 service using the go-ethereum library
With this implementation, you can create a Discovery5 Service and inititate the service itself.

*/

var (
	ModuleName = "DISC"
)

const (
	minIterTime = 100 * time.Millisecond
)

type PeerDiscovery interface {
	Start() chan *models.HostInfo
	Stop()
}

type BootNodeListString struct {
	BootNodes []string `json:"bootNodes"`
}

type Discovery struct {
	// Service control variables
	ctx context.Context

	DiscService PeerDiscovery
	DBClient    *psql.DBClient
	IpLocator   *apis.IpLocator

	wg    sync.WaitGroup
	doneC chan struct{}
}

// NewDiscovery generates a new module to discover peers in the given network with the given PeerDiscovery submodule
func NewDiscovery(ctx context.Context, discServ PeerDiscovery, db *psql.DBClient, ipLoc *apis.IpLocator) *Discovery {
	// return the Discovery object
	return &Discovery{
		ctx:         ctx,
		DiscService: discServ,
		DBClient:    db,
		IpLocator:   ipLoc,
		doneC:       make(chan struct{}),
	}
}

// Start spawns the discovery service in a separate go-routine
func (d *Discovery) Start() {
	nodeNotC := d.DiscService.Start()

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		// check if the DiscPeer Obj has a new peer to read
		for {
			select {
			case hInfo := <-nodeNotC:
				d.peerHandler(hInfo)

			case <-d.doneC:
				log.Info("shutdown detected in discovery service, shutting down")
				return

			case <-d.ctx.Done():
				log.Info("shutdown detected in discovery service, shutting down")
				return
			}
		}
	}()
}

func (d *Discovery) Stop() {
	d.DiscService.Stop()
	d.doneC <- struct{}{}
}

// peer handler for the discovered peers
func (d *Discovery) peerHandler(hInfo *models.HostInfo) {
	log.WithFields(log.Fields{
		"peer_id": hInfo.ID.String(),
		"ip":      hInfo.IP,
		"attrs":   hInfo.Attr,
	}).Debugf("discovered new peer")
	// if the peer

	// Persist to DB the hInfo
	d.DBClient.PersistToDB(hInfo)
	// if public, req location
	if utils.IsIPPublic(net.ParseIP(hInfo.IP)) {
		// get location from the received peer
		d.IpLocator.LocateIP(hInfo.IP)
	} else {
		log.Warnf("new peer %s had a non-public IP %s", hInfo.ID.String(), hInfo.IP)
	}
	log.Trace("done handling peer")
}
