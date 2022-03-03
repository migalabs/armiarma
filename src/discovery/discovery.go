package discovery

import (
	"context"

	"github.com/migalabs/armiarma/src/db"

	"github.com/migalabs/armiarma/src/utils"
	"github.com/migalabs/armiarma/src/utils/apis"

	"github.com/migalabs/armiarma/src/db/models"
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

type PeerDiscovery interface {
	Start()
	Next() bool
	Peer() (models.Peer, bool)
}

type BootNodeListString struct {
	BootNodes []string `json:"bootNodes"`
}

type Discovery struct {
	// Service control variables
	ctx context.Context

	DiscService PeerDiscovery
	PeerStore   *db.PeerStore
	IpLocator   *apis.PeerLocalizer
}

// NewDiscovery
func NewDiscovery(ctx context.Context, discServ PeerDiscovery, db *db.PeerStore, ipLoc *apis.PeerLocalizer) *Discovery {
	// return the Discovery object
	return &Discovery{
		ctx:         ctx,
		DiscService: discServ,
		PeerStore:   db,
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
		// check if the DiscPeer Obj has a new peer to read
		for d.DiscService.Next() {
			log.Debugf("next peer avail")
			// check if the ctx has been closed
			if d.ctx.Err() != nil {
				log.Info("closing the peer reader")
				return
			}
			// retrieve the next peer and check if it fine
			p, ok := d.DiscService.Peer()
			if !ok {
				continue
			}
			log.Debugf("new peer discovered: %s\n", p.PeerId)
			d.peerHandler(&p)
		}
	}()

}

// peer handler for the discovered peers
func (d *Discovery) peerHandler(pb *models.Peer) {
	// get Pub IP from MAddrs
	for _, maddr := range pb.MAddrs {
		// get the IP
		ip := utils.ExtractIPFromMAddr(maddr)
		// if public, req location
		if utils.IsIPPublic(ip) {
			pb.Ip = ip.String()
			// get location from the received peer
			locResp, err := d.IpLocator.LocateIP(pb.Ip)
			if err != nil {
				log.Debugf("could not get location from ip: %s  error: %s", pb.Ip, err)
			} else {
				pb.Country = locResp.Country
				pb.CountryCode = locResp.CountryCode
				pb.City = locResp.City
			}
		}
	}
	// store the basic peer into de PSQL DB
	d.PeerStore.StoreOrUpdatePeer(*pb)
}
