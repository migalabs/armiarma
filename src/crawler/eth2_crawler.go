/*
	Copyright © 2021 Miga Labs
*/
package crawler

import (
	"crypto/ecdsa"

	cli "github.com/urfave/cli/v2"

	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/db/postgresql"
	"github.com/migalabs/armiarma/src/discovery"
	"github.com/migalabs/armiarma/src/discovery/dv5"
	"github.com/migalabs/armiarma/src/enode"
	"github.com/migalabs/armiarma/src/exporters"
	"github.com/migalabs/armiarma/src/gossipsub"
	"github.com/migalabs/armiarma/src/gossipsub/blockchaintopics"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"
	"github.com/migalabs/armiarma/src/peering"
	"github.com/migalabs/armiarma/src/utils/apis"

	"github.com/sirupsen/logrus"
)

var (
	IpCacheSize = 5000
	// logging variables
	eth2log = logrus.WithField(
		"module", "ETH2_CRAWLER",
	)
)

// crawler status containing the main basemodule and info that the app will ConnectedF
type Eth2Crawler struct {
	Host            *hosts.BasicLibp2pHost
	Node            *enode.LocalNode
	DB              *db.PeerStore
	Disc            *discovery.Discovery
	Peering         peering.PeeringService
	Gs              *gossipsub.GossipSub
	Info            *info.Eth2InfoData
	IpLocalizer     apis.PeerLocalizer
	ExporterService *exporters.ExporterService
}

func NewEth2Crawler(ctx *cli.Context, infObj info.Eth2InfoData) (*Eth2Crawler, error) {
	// generate the central exporting service
	exporterService := exporters.NewExporterService(ctx.Context)

	// generate Eth2 network model for the PSQL
	ethmodel := postgresql.NewEth2Model("eth2")

	// Generate new DB for the peerstore
	db := db.NewPeerStore(ctx.Context, exporterService, infObj.OutputPath, infObj.DbEndpoint, &ethmodel)

	// IpLocalizer
	ipLocalizer := apis.NewPeerLocalizer(ctx.Context, IpCacheSize)

	// generate libp2pHost
	host, err := hosts.NewBasicLibp2pEth2Host(ctx.Context, infObj, &ipLocalizer, &db)
	if err != nil {
		return nil, err
	}

	// generate local Enode and DV5
	node := enode.NewLocalNode(ctx.Context, &infObj)

	// read Eth2 bootnodes
	dv5bootnodes, err := dv5.ReadEth2BootnodeFile(infObj.BootNodesFile)
	if err != nil {
		return nil, err
	}

	dv5, err := dv5.NewDiscovery(
		ctx.Context,
		node,
		(*ecdsa.PrivateKey)(infObj.PrivateKey),
		dv5bootnodes,
		infObj.ForkDigest,
		9006)
	if err != nil {
		return nil, err
	}

	disc := discovery.NewDiscovery(ctx.Context, dv5, &db, &ipLocalizer)

	// GossipSup
	gs := gossipsub.NewGossipSub(ctx.Context, exporterService, host, &db)
	// generate the peering strategy
	pStrategy, err := peering.NewPruningStrategy(ctx.Context, "eth2", &db)
	if err != nil {
		return nil, err
	}
	// Generate the PeeringService
	peeringServ, err := peering.NewPeeringService(ctx.Context, host, &db,
		peering.WithPeeringStrategy(&pStrategy),
	)
	if err != nil {
		return nil, err
	}

	// generate the CrawlerBase
	crawler := &Eth2Crawler{
		Host:            host,
		Info:            &infObj,
		DB:              &db,
		Node:            node,
		Disc:            disc,
		Peering:         peeringServ,
		Gs:              gs,
		IpLocalizer:     ipLocalizer,
		ExporterService: exporterService,
	}
	return crawler, nil
}

// generate new CrawlerBase
func (c *Eth2Crawler) Run() {
	// initialization secuence for the crawler
	c.ExporterService.Run()
	c.IpLocalizer.Run()
	c.Host.Start()
	c.Disc.Start()
	topics := blockchaintopics.ReturnTopics(c.Info.ForkDigest, c.Info.TopicArray)
	for _, topic := range topics {
		c.Gs.JoinAndSubscribe(topic)
	}
	c.Peering.Run()
	c.Gs.ServeMetrics()
	c.DB.ServeMetrics()
}
