/*
	Copyright © 2021 Miga Labs
*/
package cmd

import (
	"context"

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
	"github.com/sirupsen/logrus"
)

var (
	IpCacheSize = 400

	ModuleName = "CRAWLER"
	log        = logrus.WithField(
		"module", ModuleName,
	)
)

func CrawlerHelp() string {
	return "\t--config-file\tconfig-file with all the available configurations. Find an example at ./config-files/config.json"
}

// crawler status containing the main basemodule and info that the app will ConnectedF
type Crawler struct {
	ctx    context.Context
	cancel context.CancelFunc

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
	mainCtx, cancel := context.WithCancel(ctx)
	info_tmp := info.NewCustomInfoData(config)
	// Generate new DB for the peerstore
	db := db.NewPeerStore(mainCtx, info_tmp.GetDBType(), info_tmp.GetOutputPath())
	// IpLocalizer
	ipLocalizer := apis.NewPeerLocalizer(mainCtx, IpCacheSize)
	// generate libp2pHost
	host, err := hosts.NewBasicLibp2pHost(mainCtx, *info_tmp, &ipLocalizer, &db)
	if err != nil {
		return nil, err
	}
	// generate local Enode and DV5
	node_tmp := enode.NewLocalNode(mainCtx, info_tmp)
	//node_tmp.AddEntries()
	dv5_tmp := discovery.NewDiscovery(mainCtx, node_tmp, &db, &ipLocalizer, info_tmp, 9006)
	// GossipSup
	gs_tmp := gossipsub.NewGossipSub(mainCtx, host, &db)
	// generate the peering strategy
	pStrategy, err := peering.NewPruningStrategy(mainCtx, &db)
	if err != nil {
		return nil, err
	}
	// Generate the PeeringService
	peeringServ, err := peering.NewPeeringService(mainCtx, host, &db, info_tmp,
		peering.WithPeeringStrategy(&pStrategy),
	)
	if err != nil {
		return nil, err
	}
	prometheusRunner := prometheus.NewPrometheusRunner()

	// generate the CrawlerBase
	crawler := &Crawler{
		ctx:              mainCtx,
		cancel:           cancel,
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
	c.PrometheusRunner.Start()
	c.IpLocalizer.Run()
	c.Host.Start()
	c.Dv5.Start()
	c.Dv5.FindRandomNodes()
	topics := blockchaintopics.ReturnTopics(c.Info.GetForkDigest(), c.Info.GetTopicArray())
	for _, topic := range topics {
		c.Gs.JoinAndSubscribe(topic)
	}
	c.Peering.Run()
	c.Gs.ServePrometheusMetrics()
	c.DB.ServePrometheusMetrics()
	c.DB.ExportCsvService(c.Info.GetOutputPath())
}

// generate new CrawlerBases
func (c *Crawler) Close() {
	defer c.cancel()
	// initialization secuence for the crawler
	log.Info("stoping crawler client")
	c.Dv5.CloseFindingNodes()
	c.DB.Close()
	c.Gs.Close()
	c.Peering.Close()
	c.Host.Stop()
	c.IpLocalizer.Close()
}
