/*
	Copyright © 2021 Miga Labs
*/
package crawler

import (
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/db/postgresql"
	"github.com/migalabs/armiarma/src/discovery"
	"github.com/migalabs/armiarma/src/discovery/kdht"
	"github.com/migalabs/armiarma/src/exporters"
	"github.com/migalabs/armiarma/src/gossipsub"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"
	"github.com/migalabs/armiarma/src/peering"
	"github.com/migalabs/armiarma/src/utils/apis"

	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

// TEMPORARY data for the running the filecoin demo
var (
	workers       = 100
	minWaitTime   = 5 * time.Second
	timeout       = 10 * time.Second
	ipfsprotocols = []string{
		"/ipfs/kad/1.0.0",
		"/ipfs/kad/2.0.0",
	}
	filecoinprotocols = []string{
		"/fil/kad/testnetnet/kad/1.0.0",
	}
	// logging variables
	ipfslog = logrus.WithField(
		"module", "IPFS_CRAWLER",
	)
)

// crawler status containing the main basemodule and info that the app will ConnectedF
type IpfsCrawler struct {
	Host            *hosts.BasicLibp2pHost
	DB              *db.PeerStore
	Disc            *discovery.Discovery
	Peering         peering.PeeringService
	Gs              *gossipsub.GossipSub
	Info            *info.IpfsInfoData
	IpLocalizer     apis.PeerLocalizer
	ExporterService *exporters.ExporterService
}

func NewIpfsCrawler(ctx *cli.Context, infObj info.IpfsInfoData) (*IpfsCrawler, error) {

	exporterService := exporters.NewExporterService(ctx.Context)

	// generate Eth2 network model for the PSQL
	ipfsmodel := postgresql.NewIpfsPeerModel("ipfs")

	// Generate new DB for the peerstore
	db := db.NewPeerStore(ctx.Context, exporterService, infObj.OutputPath, infObj.DbEndpoint, &ipfsmodel)

	// IpLocalizer
	ipLocalizer := apis.NewPeerLocalizer(ctx.Context, IpCacheSize)

	// Host
	host, err := hosts.NewBasicLibp2pIpfsHost(ctx.Context, infObj, &ipLocalizer, &db)
	if err != nil {
		return nil, err
	}

	// IPFS protocols
	protocols := make([]string, 0)
	// filter the network to select needed protocols
	switch infObj.Network {
	case "ipfs":
		protocols = info.Ipfsprotocols
	case "filecoin":
		protocols = info.Filecoinprotocols
	case "":
		ipfslog.Warn("network not defined. setting ipfs by default")
		protocols = info.Ipfsprotocols
	default:
		ipfslog.Warnf("network not recognized. network=%s. setting ipfs by default", infObj.Network)
		protocols = info.Ipfsprotocols
	}
	ipfslog.Infoln("running peer discovery with protocols:", protocols)

	// discovery nodes
	bootnodes, err := kdht.ReadIpfsBootnodeFile(infObj.BootNodesFile)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retreive the bootnodes")
	}
	// generate KDHT peer discovery
	kdhtd := kdht.NewIPFSDiscService(ctx.Context, host.Host(), protocols, bootnodes, timeout)

	// Peer discovery
	disc := discovery.NewDiscovery(ctx.Context, &kdhtd, &db, &ipLocalizer)

	// GossipSup
	gs := gossipsub.NewGossipSub(ctx.Context, exporterService, host, &db)

	// generate the peering strategy

	pStrategy, err := peering.NewPruningStrategy(ctx.Context, "ipfs", &db)
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
	crawler := &IpfsCrawler{
		Host:            host,
		Info:            &infObj,
		DB:              &db,
		Disc:            disc,
		Peering:         peeringServ,
		Gs:              gs,
		IpLocalizer:     ipLocalizer,
		ExporterService: exporterService,
	}
	return crawler, nil
}

// generate new CrawlerBase
func (c *IpfsCrawler) Run() {
	// IMPORTANT
	// Set the VENV Variable for handling too many opened connections
	os.Setenv("LIBP2P_SWARM_FD_LIMIT", "10000")

	// initialization secuence for the crawler
	c.ExporterService.Run()
	c.IpLocalizer.Run()
	c.Host.Start()
	c.Disc.Start()
	/*
		// no topics to join so far
		for _, topic := range topics {
			c.Gs.JoinAndSubscribe(topic)
		}
	*/
	c.Peering.Run()
	c.Gs.ServeMetrics()
	c.DB.ServeMetrics()
}
