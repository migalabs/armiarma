package discovery

import (
	"context"
	"net"
	"time"

	psql "github.com/migalabs/armiarma/pkg/db/postgresql"

	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/migalabs/armiarma/pkg/utils/apis"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/sirupsen/logrus"
)

/*
This file implements the discovery5 service using the go-ethereum library
With this implementation, you can create a Discovery5 Service and inititate the service itself.

*/

var (
	ModuleName = "DISC"
	log        = logrus.WithField(
		"module", ModuleName,
	)
)

const (
	minIterTime = 150 * time.Millisecond
)

type PeerDiscovery interface {
	Start()
	Stop()
	Next() bool
	Peer() (*models.HostInfo, bool)
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
}

// NewDiscovery
func NewDiscovery(ctx context.Context, discServ PeerDiscovery, db *psql.DBClient, ipLoc *apis.IpLocator) *Discovery {
	// return the Discovery object
	return &Discovery{
		ctx:         ctx,
		DiscService: discServ,
		DBClient:    db,
		IpLocator:   ipLoc,
	}
}

// Start
func (d *Discovery) Start() {
	log.Info("starting peer discovery service")
	d.DiscService.Start()
	// get next peer from the DiscService
	go func() {
		log.Info("launching peer reader")
		ticker := time.NewTicker(minIterTime)
		// check if the DiscPeer Obj has a new peer to read
		for {
			if d.DiscService.Next() {
				log.Debugf("next peer avail")
				// check if the ctx has been closed
				if d.ctx.Err() != nil {
					log.Info("closing the peer reader")
					return
				}
				// retrieve the next peer and check if it fine
				newPeer, ok := d.DiscService.Peer()
				if !ok {
					continue
				}
				log.Debugf("new peer discovered: %s\n", newPeer.ID.String())
				d.peerHandler(newPeer)
			}

			select {
			case <-ticker.C:
				// wait untill min time has passed
				ticker.Reset(minIterTime)
			case <-d.ctx.Done():
				log.Info("shutdown detected in discovery service, shutting down")
				ticker.Stop()
				// TODO: add STOP To -> d.DiscService.Stop()
				return
			}
		}
	}()
}

// peer handler for the discovered peers
func (d *Discovery) peerHandler(hInfo *models.HostInfo) {
	// Persist to DB the hInfo
	d.DBClient.PersistToDB(hInfo)
	// if public, req location
	if utils.IsIPPublic(net.ParseIP(hInfo.IP)) {
		// get location from the received peer
		d.IpLocator.LocateIP(hInfo.IP)
	} else {
		log.Debugf("new peer %s had a non-public IP %s", hInfo.ID.String(), hInfo.IP)
	}
	// iter through all the attributes of the Node to persit them
	for key, att := range hInfo.Attr {
		log.Debugf("found attr %s on peer %s", key, hInfo.ID.String())
		d.DBClient.PersistToDB(att)
	}
}
