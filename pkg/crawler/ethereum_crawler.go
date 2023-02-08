/*
Copyright © 2021 Miga Labs
*/
package crawler

import (
	"context"
	"crypto/ecdsa"
	"strings"

	"github.com/libp2p/go-libp2p-core/crypto"
	cli "github.com/urfave/cli/v2"

	"github.com/migalabs/armiarma/pkg/config"
	psql "github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/migalabs/armiarma/pkg/discovery"
	"github.com/migalabs/armiarma/pkg/discovery/dv5"
	"github.com/migalabs/armiarma/pkg/enode"
	"github.com/migalabs/armiarma/pkg/gossipsub"
	"github.com/migalabs/armiarma/pkg/hosts"
	"github.com/migalabs/armiarma/pkg/metrics"
	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
	"github.com/migalabs/armiarma/pkg/peering"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/migalabs/armiarma/pkg/utils/apis"
	log "github.com/sirupsen/logrus"
)

// crawler status containing the main basemodule and info that the app will ConnectedF
type EthereumCrawler struct {
	ctx       context.Context
	cancel    context.CancelFunc
	Host      *hosts.BasicLibp2pHost
	EthNet    *eth.EthereumNetwork
	Node      *enode.LocalNode
	DB        *psql.DBClient
	Disc      *discovery.Discovery
	Peering   peering.PeeringService
	Gossipsub *gossipsub.GossipSub
	IpLocator *apis.IpLocator
	Metrics   *metrics.PrometheusMetrics
}

func NewEthereumCrawler(mainCtx *cli.Context, conf config.EthereumCrawlerConfig) (*EthereumCrawler, error) {
	// Setup the configuration
	log.SetLevel(utils.ParseLogLevel(conf.LogLevel))

	// --- build up all the necesary modules ---
	ctx, cancel := context.WithCancel(mainCtx.Context)
	var err error

	// generate confg for EthNetwork // TODO: should be moved to enode
	ethNet := eth.NewEthereumNetwork(
		ctx,
		eth.ComposeQuickBeaconStatus(conf.ForkDigest),
		eth.ComposeQuickBeaconMetaData(),
	)
	// serve

	// Private Key
	var gethPrivKey *ecdsa.PrivateKey
	var libp2pPrivKey *crypto.Secp256k1PrivateKey
	if conf.PrivateKey == "" {
		gethPrivKey, err = utils.GenerateECDSAPrivKey()
		if err != nil {
			return nil, err
		}
	} else {
		gethPrivKey, err = utils.ParseECDSAPrivateKey(conf.PrivateKey)
		if err != nil {
			return nil, err
		}
	}
	libp2pPrivKey, err = utils.AdaptSecp256k1FromECDSA(gethPrivKey)
	if err != nil {
		return nil, err
	}

	// generate the central exporting service
	promethMetrics := metrics.NewPrometheusMetrics(ctx)

	// Generate/connect to PSQL Database
	dbClient, err := psql.NewDBClient(ctx, ethNet.NetworkType(), conf.PsqlEndpoint, true)
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
		conf.IP,
		conf.Port,
		libp2pPrivKey,
		conf.UserAgent,
		ethNet,
		ipLocator,
	)
	if err != nil {
		cancel()
		return nil, err
	}

	// generate local Enode and DV5 service
	node := enode.NewLocalNode(ctx, gethPrivKey)
	// subscribre to all attestnets and set forkdigest
	node.SetAttNetworks("ffffffffffffffff")
	node.SetForkDigest(strings.Trim(conf.ForkDigest, "0x"))

	dv5, err := dv5.NewDiscovery5(
		ctx,
		node,
		gethPrivKey,
		dv5.ParseBootnodesFromStringSlice(conf.Bootnodes),
		conf.ForkDigest,
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
	gs := gossipsub.NewGossipSub(ctx, host.Host(), dbClient)

	// generate a new subnets-handler
	ethMsgHandler, err := eth.NewEthMessageHandler(eth.GoerliGenesis, conf.ValPubkeys)
	if err != nil {
		cancel()
		return nil, err
	}
	// subscribe the topics
	for _, top := range conf.GossipTopics {
		var msgHandler gossipsub.MessageHandler
		switch top {
		case eth.BeaconBlockTopicBase:
			msgHandler = ethMsgHandler.BeaconBlockMessageHandler
		default:
			log.Error("untraceable gossipsub topic", top)
			continue
		}
		topic := eth.ComposeTopic(conf.ForkDigest, top)
		gs.JoinAndSubscribe(topic, msgHandler)
	}
	// ---- Subcribe to Subnets ----
	for _, subnet := range conf.Subnets {
		subTopics := eth.ComposeAttnetsTopic(conf.ForkDigest, subnet)
		gs.JoinAndSubscribe(subTopics, ethMsgHandler.SubnetMessageHandler)
	}

	// generate the peering strategy
	pStrategy, err := peering.NewPruningStrategy(
		ctx,
		ethNet.NetworkType(),
		conf.LocalPeerstorePath,
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

	// init the metrics for all the modules

	// generate the CrawlerBase
	crawler := &EthereumCrawler{
		ctx:       ctx,
		cancel:    cancel,
		Host:      host,
		DB:        dbClient,
		Node:      node,
		Disc:      disc,
		Peering:   peeringServ,
		Gossipsub: gs,
		IpLocator: ipLocator,
		EthNet:    ethNet,
		Metrics:   promethMetrics,
	}

	// Register the metrics for the crawler and submodules
	crawlMetricsMod := crawler.GetMetrics()
	promethMetrics.AddMeticsModule(crawlMetricsMod)

	pruneMetricsMod := peeringServ.GetMetrics()
	promethMetrics.AddMeticsModule(pruneMetricsMod)

	discoveryMetricsMod := disc.GetEthereumMetrics()
	promethMetrics.AddMeticsModule(discoveryMetricsMod)

	hostMetricsMod := host.GetMetrics()
	promethMetrics.AddMeticsModule(hostMetricsMod)

	gossipMetricsMod := gs.GetMetrics()
	promethMetrics.AddMeticsModule(gossipMetricsMod)

	return crawler, nil
}

// generate new CrawlerBase
func (c *EthereumCrawler) Run() {
	// init all the eth_protocols
	c.EthNet.ServeBeaconPing(c.Host.Host())
	c.EthNet.ServeBeaconStatus(c.Host.Host())
	c.EthNet.ServeBeaconMetadata(c.Host.Host())

	// initialization secuence for the crawler
	c.IpLocator.Run()
	c.Host.Start()
	c.Disc.Start()
	c.Peering.Run()
	c.Metrics.Start()
}

func (c *EthereumCrawler) Close() {
	c.Host.Host().Close()
	c.Disc.Stop()
	c.DB.Close()
	c.Metrics.Close()
	c.cancel()

}
