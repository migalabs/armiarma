/*
Copyright © 2021 Miga Labs
*/
package crawler

import (
	"context"
	"crypto/ecdsa"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	cli "github.com/urfave/cli/v2"

	"github.com/migalabs/armiarma/pkg/config"
	psql "github.com/migalabs/armiarma/pkg/db/postgresql"
	"github.com/migalabs/armiarma/pkg/discovery"
	"github.com/migalabs/armiarma/pkg/discovery/dv5"
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
	EthNode   *eth.LocalEthereumNode
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

	ctx, cancel := context.WithCancel(mainCtx.Context)
	var err error

	// parse or create a private key for the host
	var gethPrivKey *ecdsa.PrivateKey
	var libp2pPrivKey *crypto.Secp256k1PrivateKey
	if conf.PrivateKey == "" {
		gethPrivKey, err = utils.GenerateECDSAPrivKey()
		if err != nil {
			cancel()
			return nil, err
		}
	} else {
		gethPrivKey, err = utils.ParseECDSAPrivateKey(conf.PrivateKey)
		if err != nil {
			cancel()
			return nil, err
		}
	}
	libp2pPrivKey, err = utils.AdaptSecp256k1FromECDSA(gethPrivKey)
	if err != nil {
		cancel()
		return nil, err
	}

	// generate local node for the ethereum network
	ethNode := eth.NewLocalEthereumNode(
		ctx,
		gethPrivKey,
		eth.ComposeQuickBeaconStatus(conf.ForkDigest),
		eth.ComposeQuickBeaconMetaData(),
	)
	// subscribre to all attestnets and set forkdigest
	ethNode.SetAttNetworks("ffffffffffffffff")
	ethNode.SetForkDigest(strings.Trim(conf.ForkDigest, "0x"))

	// generate the central exporting service
	promethMetrics := metrics.NewPrometheusMetrics(ctx)

	// generate/connect to PSQL Database
	backupInterval, err := time.ParseDuration(conf.ActivePeersBackupInterval)
	if err != nil {
		cancel()
		return nil, err
	}
	dbClient, err := psql.NewDBClient(ctx, ethNode.Network(), conf.PsqlEndpoint, backupInterval, true)
	if err != nil {
		cancel()
		return nil, err
	}

	// create an ip-locator instance
	ipLocator := apis.NewIpLocator(ctx, dbClient)

	// generate libp2pHostd
	host, err := hosts.NewBasicLibp2pEth2Host(
		ctx,
		conf.IP,
		conf.Port,
		libp2pPrivKey,
		conf.UserAgent,
		ethNode, // ethereum local node
		ipLocator,
	)
	if err != nil {
		cancel()
		return nil, err
	}

	// create a new discovery5 service to discover peers in the Ethereum network
	dv5, err := dv5.NewDiscovery5(
		ctx,
		ethNode,
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

	// create a gossipsub routing
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
	// subcribe to attestation subnets
	for _, subnet := range conf.Subnets {
		subTopics := eth.ComposeAttnetsTopic(conf.ForkDigest, subnet)
		gs.JoinAndSubscribe(subTopics, ethMsgHandler.SubnetMessageHandler)
	}

	// generate the peering strategy
	pStrategy, err := peering.NewPruningStrategy(
		ctx,
		ethNode.Network(),
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

	// generate the CrawlerBase
	crawler := &EthereumCrawler{
		ctx:       ctx,
		cancel:    cancel,
		Host:      host,
		DB:        dbClient,
		EthNode:   ethNode,
		Disc:      disc,
		Peering:   peeringServ,
		Gossipsub: gs,
		IpLocator: ipLocator,
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
	c.EthNode.ServeBeaconPing(c.Host.Host())
	c.EthNode.ServeBeaconStatus(c.Host.Host())
	c.EthNode.ServeBeaconMetadata(c.Host.Host())

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
