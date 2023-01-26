/*
	Copyright © 2021 Miga Labs
*/
package crawler

import (
	"context"
	"crypto/ecdsa"

	cli "github.com/urfave/cli/v2"

	"github.com/migalabs/armiarma/pkg/discovery"
	"github.com/migalabs/armiarma/pkg/discovery/dv5"
	"github.com/migalabs/armiarma/pkg/enode"
	"github.com/migalabs/armiarma/pkg/exporters"
	"github.com/migalabs/armiarma/pkg/gossipsub"
	"github.com/migalabs/armiarma/pkg/utils"

	//"github.com/migalabs/armiarma/pkg/gossipsub/blockchaintopics"
	psql "github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/migalabs/armiarma/pkg/hosts"
	"github.com/migalabs/armiarma/pkg/info"
	"github.com/migalabs/armiarma/pkg/peering"
	"github.com/migalabs/armiarma/pkg/utils/apis"

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
	ctx             context.Context
	cancel          context.CancelFunc
	Host            *hosts.BasicLibp2pHost
	Node            *enode.LocalNode
	DB              *psql.DBClient
	Disc            *discovery.Discovery
	Peering         peering.PeeringService
	Gs              *gossipsub.GossipSub
	Info            *info.Eth2InfoData
	IpLocator       *apis.IpLocator
	ExporterService *exporters.ExporterService
}

func NewEth2Crawler(mainCtx *cli.Context, infObj info.Eth2InfoData) (*Eth2Crawler, error) {
	ctx, cancel := context.WithCancel(mainCtx.Context)

	network := utils.EthereumNetwork

	// generate the central exporting service
	exporterService := exporters.NewExporterService(ctx)

	// Generate/connect to PSQL Database
	dbClient, err := psql.NewDBClient(
		ctx,
		network,
		infObj.DbEndpoint,
		true, // we want the DB intitialized
	)
	if err != nil {
		cancel()
		return nil, err
	}

	// IpLocalizer
	ipLocator := apis.NewIpLocator(ctx, dbClient)

	// generate libp2pHost
	// TODO: pass only strictly necessary info (IP, PORT, PrivKey)
	host, err := hosts.NewBasicLibp2pEth2Host(
		ctx,
		infObj,
		network,
		ipLocator,
		dbClient,
	)
	if err != nil {
		cancel()
		return nil, err
	}

	// generate local Enode and DV5
	node := enode.NewLocalNode(ctx, (*ecdsa.PrivateKey)(infObj.PrivateKey))

	// read Eth2 bootnodes
	dv5bootnodes, err := dv5.ReadEth2BootnodeFile(infObj.BootNodesFile)
	if err != nil {
		cancel()
		return nil, err
	}

	dv5, err := dv5.NewDiscovery5(
		ctx,
		node,
		(*ecdsa.PrivateKey)(infObj.PrivateKey),
		dv5bootnodes,
		infObj.ForkDigest,
		9006)
	if err != nil {
		cancel()
		return nil, err
	}

	disc := discovery.NewDiscovery(
		ctx,
		dv5,
		dbClient,
		ipLocator,
	)

	// GossipSup
	gs := gossipsub.NewGossipSub(ctx, exporterService, host, dbClient)
	// generate the peering strategy
	pStrategy, err := peering.NewPruningStrategy(
		ctx,
		network,
		".peerstore", // TODO: remove hardcoded
		dbClient,
	)
	if err != nil {
		cancel()
		return nil, err
	}
	// Generate the PeeringService
	peeringServ, err := peering.NewPeeringService(
		ctx,
		host,
		dbClient,
		peering.WithPeeringStrategy(pStrategy),
	)
	if err != nil {
		cancel()
		return nil, err
	}

	// generate the CrawlerBase
	crawler := &Eth2Crawler{
		ctx:             ctx,
		cancel:          cancel,
		Host:            host,
		Info:            &infObj,
		DB:              dbClient,
		Node:            node,
		Disc:            disc,
		Peering:         peeringServ,
		Gs:              gs,
		IpLocator:       ipLocator,
		ExporterService: exporterService,
	}
	return crawler, nil
}

// generate new CrawlerBase
func (c *Eth2Crawler) Run() {
	// initialization secuence for the crawler
	c.ExporterService.Run()
	c.IpLocator.Run()
	c.Host.Start()
	c.Disc.Start()
	//topics := blockchaintopics.ReturnTopics(c.Info.ForkDigest, c.Info.TopicArray)
	//for _, topic := range topics {
	//	c.Gs.JoinAndSubscribe(topic)
	//}
	c.Peering.Run()
	// c.Gs.ServeMetrics()
	//c.DB.ServeMetrics() // TODO: Missing
}

func (c *Eth2Crawler) Close() {
	c.Host.Host().Close()
	c.Disc.Stop()
	c.DB.Close()
	c.cancel()

}
