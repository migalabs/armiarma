/*
	Copyright © 2021 Miga Labs
*/
package cmd

import (
	"context"
	"time"

	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/config"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/discovery"
	"github.com/migalabs/armiarma/src/enode"
	"github.com/migalabs/armiarma/src/gossipsub"
	"github.com/migalabs/armiarma/src/gossipsub/blockchaintopics"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"
	"github.com/migalabs/armiarma/src/peering"
	"github.com/migalabs/armiarma/src/prometheus"
	"github.com/migalabs/armiarma/src/utils/apis"
)

var (
	IpCacheSize = 4000
)

func CrawlerHelp() string {
	return "-crawler\tLaunch the Network Crawler on the given network"
}

// crawler status containing the main basemodule and info that the app will ConnectedF
type Crawler struct {
	*base.Base
	Host             *hosts.BasicLibp2pHost
	Node             *enode.LocalNode
	DB               *db.PeerStore
	Dv5              *discovery.Discovery
	Peering          peering.PeeringService
	Gs               *gossipsub.GossipSub
	Info             *info.InfoData
	IpLocalizer      apis.PeerLocalizer
	PrometheusRunner *prometheus.PrometheusRunner
}

func NewCrawler(ctx context.Context, config config.ConfigData) (*Crawler, error) {
	// TODO: just hardcoded
	logOpts := base.LogOpts{
		ModName:   "CRAWLER MAIN",
		Output:    "terminal",
		Formatter: "text",
		Level:     "debug",
	}
	// generate a base for the crawler app
	b, err := base.NewBase(
		base.WithContext(ctx),
		base.WithLogger(logOpts),
	)
	if err != nil {
		return nil, err
	}

	stdOpts := base.LogOpts{
		Output:    "terminal",
		Formatter: "text",
	}

	info_tmp := info.NewCustomInfoData(config, stdOpts)
	stdOpts.Level = info_tmp.GetLogLevel()

	// TODO: just harcoded
	baseOpts := base.LogOpts{
		ModName:   "libp2p host",
		Output:    "terminal",
		Formatter: "text",
		Level:     info_tmp.GetLogLevel(),
	}

	// TODO: generate a new DB
	db := db.NewPeerStore(info_tmp.GetDBType(), info_tmp.GetOutputPath())

	// IpLocalizer
	ipLocalizer := apis.NewPeerLocalizer(b.Ctx(), IpCacheSize)

	hostOpts := hosts.BasicLibp2pHostOpts{
		Info_obj:  *info_tmp,
		LogOpts:   baseOpts,
		IpLocator: &ipLocalizer,
		PeerStore: &db,
	}

	// generate libp2pHost
	host, err := hosts.NewBasicLibp2pHost(b.Ctx(), hostOpts)
	if err != nil {
		return nil, err
	}

	// generate local Enode and DV5
	node_tmp := enode.NewLocalNode(b.Ctx(), info_tmp, stdOpts)
	//node_tmp.AddEntries()
	dv5_tmp := discovery.NewDiscovery(b.Ctx(), node_tmp, &db, &ipLocalizer, info_tmp, 9006, stdOpts)

	// GossipSup
	gs_tmp := gossipsub.NewGossipSub(b.Ctx(), host, &db, stdOpts)

	// Generate the PeeringService
	peeringOpts := &peering.PeeringOpts{
		InfoObj: info_tmp,
		LogOpts: stdOpts,
	}
	// generate the peering strategy
	prunOpts := peering.PruningOpts{
		AggregatedDelay: 24 * time.Hour, // Hardcoded, still using the Default Delay
		LogOpts:         stdOpts,
	}

	pStrategy, err := peering.NewPruningStrategy(b.Ctx(), &db, prunOpts)
	peeringServ, err := peering.NewPeeringService(b.Ctx(), host, &db, peeringOpts,
		peering.WithPeeringStrategy(&pStrategy),
	)

	prometheusRunner := prometheus.NewPrometheusRunner()

	// generate the CrawlerBase
	crawler := &Crawler{
		Base:             b,
		Host:             host,
		Info:             info_tmp,
		DB:               &db,
		Node:             node_tmp,
		Dv5:              dv5_tmp,
		Peering:          peeringServ,
		Gs:               gs_tmp,
		IpLocalizer:      ipLocalizer,
		PrometheusRunner: &prometheusRunner,
	}
	return crawler, nil
}

// generate new CrawlerBase
func (c *Crawler) Run() {
	// initialization secuence for the crawler
	mainctx := c.Ctx()

	c.PrometheusRunner.Start(mainctx)
	c.IpLocalizer.Run()
	c.Host.Start()
	c.Dv5.Start_dv5()
	go c.Dv5.FindRandomNodes()

	topics := blockchaintopics.ReturnTopics(c.Info.GetForkDigest(), c.Info.GetTopicArray())
	for _, topic := range topics {
		c.Gs.JoinAndSubscribe(topic)
	}

	go c.Peering.Run()

	c.Gs.ServeMetrics()
	// Generate a Peering Service (so far with default peering strategy)
	c.DB.ServeMetrics(mainctx)

	go c.DB.ExportLoop(mainctx, c.Info.GetOutputPath())
}

// generate new CrawlerBases
func (c *Crawler) Close() {
	// initialization secuence for the crawler
	c.Log.Info("stoping crawler client")
	c.Host.Stop()
	c.Peering.Close()
	c.IpLocalizer.Close()
}
