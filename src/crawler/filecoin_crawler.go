/*
	Copyright © 2021 Miga Labs
*/
package crawler

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/migalabs/armiarma/src/config"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/exporters"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"
	"github.com/migalabs/armiarma/src/utils"

	"github.com/migalabs/armiarma/src/db/postgresql"
	"github.com/migalabs/armiarma/src/utils/apis"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"

	ma "github.com/multiformats/go-multiaddr"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
)

// TEMPORARY data for the running the filecoin demo
var (
	workers = 25

	bootstrapNodes = []string{
		"/ip4/3.224.142.21/tcp/1347/p2p/12D3KooWCVe8MmsEMes2FzgTpt9fXtmCY7wrq91GRiaC8PHSCCBj",
		"/ip4/107.23.112.60/tcp/1347/p2p/12D3KooWCwevHg1yLCvktf2nvLu7L9894mcrJR4MsBCcm4syShVc",
		"/ip4/100.25.69.197/tcp/1347/p2p/12D3KooWEWVwHGn2yR36gKLozmb4YjDJGerotAPGxmdWZx2nxMC4",
		"/ip4/3.123.163.135/tcp/1347/p2p/12D3KooWKhgq8c7NQ9iGjbyK7v7phXvG6492HQfiDaGHLHLQjk7R",
		"/ip4/18.198.196.213/tcp/1347/p2p/12D3KooWL6PsFNPhYftrJzGgF5U18hFoaVhfGk7xwzD8yVrHJ3Uc",
		"/ip4/18.195.111.146/tcp/1347/p2p/12D3KooWLFynvDQiUpXoHroV1YxKHhPJgysQGH2k3ZGwtWzR4dFH",
		"/ip4/52.77.116.139/tcp/1347/p2p/12D3KooWP5MwCiqdMETF9ub1P3MbCvQCcfconnYHbWg6sUJcDRQQ",
		"/ip4/18.136.2.101/tcp/1347/p2p/12D3KooWRs3aY1p3juFjPy8gPN95PEQChm2QKGUCAdcDCC4EBMKf",
		"/ip4/13.250.155.222/tcp/1347/p2p/12D3KooWScFR7385LTyR4zU1bYdzSiiAb5rnNABfVahPvVSzyTkR",
		"/ip4/47.115.22.33/tcp/41778/p2p/12D3KooWDqaZkm3oSczUm3dvAJ5aL2rdSeQ5VQbnHRTQNEFShhmc",
		"/ip4/61.147.123.111/tcp/12757/p2p/12D3KooWGhufNmZHF3sv48aQeS13ng5XVJZ9E6qy2Ms4VzqeUsHk",
		"/ip4/61.147.123.121/tcp/12757/p2p/12D3KooWDgQrcyZpcMAkbEFSJJYV2qXEMwXX67WTbqpNdbifHaEq",
		"/ip4/3.129.112.217/tcp/1235/p2p/12D3KooWBF8cpp65hp2u9LK5mh19x67ftAam84z9LsfaquTDSBpt",
		"/ip4/36.103.232.198/tcp/34721/p2p/12D3KooWQnwEGNqcM2nAcPtRR9rAX8Hrg4k9kJLCHoTR5chJfz6d",
		"/ip4/36.103.232.198/tcp/34723/p2p/12D3KooWMKxMkD5DMpSWsW7dBddKxKT7L2GgbNuckz9otxvkvByP",
		"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/ip4/104.236.151.122/tcp/4001/ipfs/QmSoLju6m7xTh3DuokvT3886QRYqxAzb1kShaanJgW36yx",
		"/ip4/104.236.176.52/tcp/4001/ipfs/QmSoLnSGccFuZQJzRadHn95W2CrSFmZuTdDWP8HXaHca9z",
		"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLpPVmHKQ4XTPdz8tjDFgdeRFkpV8JgYq8JVJ69RrZm",
		"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
		"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
		"/ip4/162.243.248.213/tcp/4001/ipfs/QmSoLueR4xBeUbY9WZ9xGUUxunbKWcrNFTDAadQJmocnWm",
		"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
		"/ip4/178.62.61.185/tcp/4001/ipfs/QmSoLMeWqB7YGVLJN3pNLQpmmEk35v6wYtsMGLzSr5QBU3",
	}
	protocols = []string{
		"/ipfs/kad/1.0.0",
		"/ipfs/kad/2.0.0",
		"/dnsaddr/bootstrap.libp2p.io",
	}
)

// crawler status containing the main basemodule and info that the app will ConnectedF
type FilecoinCrawler struct {
	ctx    context.Context
	cancel context.CancelFunc

	Host            *hosts.BasicLibp2pHost
	pm              *pb.ProtocolMessenger
	DB              *postgresql.PostgresDBService
	Info            *info.InfoData
	IpLocalizer     apis.PeerLocalizer
	ExporterService *exporters.ExporterService
}

func NewFilecoinCrawler(ctx context.Context, config config.ConfigData) (*FilecoinCrawler, error) {
	mainCtx, cancel := context.WithCancel(ctx)
	infoObj := info.NewCustomInfoData(config)

	ipLocalizer := apis.NewPeerLocalizer(mainCtx, IpCacheSize)
	exporterService := exporters.NewExporterService(mainCtx)
	db := db.NewPeerStore(mainCtx, exporterService, infoObj.GetOutputPath(), infoObj.GetDBEndpoint())
	// Neccessary secuence for setting up the network crawler
	// 1. Create Host
	log.Info("creating host")
	host, err := hosts.NewBasicLibp2pFilecoin2Host(mainCtx, *infoObj, &ipLocalizer, &db)
	if err != nil {
		return nil, err
	}

	h := host.Host()

	// generate the PostgresDBService
	psql, err := postgresql.ConnectToDB(ctx, infoObj.GetDBEndpoint())
	if err != nil {
		return nil, err
	}
	// 3. Create the Exporting Service
	// exporterService := exporters.NewExporterService(mainCtx)

	// Generate necessary messenger for requesting near peers
	ms := &msgSender{
		h:         h,
		protocols: protocol.ConvertFromStrings(protocols),
		timeout:   10 * time.Second,
	}

	pm, err := pb.NewProtocolMessenger(ms)
	if err != nil {
		return nil, err
	}

	// generate the CrawlerBase
	crawler := &FilecoinCrawler{
		ctx:             mainCtx,
		cancel:          cancel,
		Host:            host,
		pm:              pm,
		Info:            infoObj,
		DB:              psql,
		IpLocalizer:     ipLocalizer,
		ExporterService: exporterService,
	}
	return crawler, nil
}

// generate new CrawlerBase
func (c *FilecoinCrawler) Run() {
	// initialization secuence for the crawler
	c.ExporterService.Run()
	c.IpLocalizer.Run()
	c.Host.Start()
	c.crawlNetwork()
	//c.DB.ServeMetrics()
	//c.DB.ExportCsvService(c.Info.GetOutputPath())
}

// generate new CrawlerBases
func (c *FilecoinCrawler) Close() {
	defer c.cancel()
	// initialization secuence for the crawler
	log.Info("stoping crawler client")
	c.Host.Stop()
	c.IpLocalizer.Close()
	c.ExporterService.Close()
}

func (c *FilecoinCrawler) crawlNetwork() {
	// 2. Create Kademlia DHT service
	h := c.Host.Host()

	log.Info("generating dht")
	kdht, err := dht.New(c.ctx, h)
	if err != nil {
		log.Error(err)
	}
	// bootstrap
	log.Info("setting the bootstrap to dht")
	err = kdht.Bootstrap(c.ctx)
	if err != nil {
		log.Error(err)
	}

	// Peer Discovery
	connectablePeers := NewDiscoveryPeers(c.DB)

	// Fill with bootstrap nodes
	log.Info("connecting to the bootstrap nodes")
	for _, peerAddr := range bootstrapNodes {
		maddr, _ := utils.UnmarshalMaddr(peerAddr)
		peerInfo, _ := peer.AddrInfoFromP2pAddr(maddr)
		// Load it to the sync map
		connectablePeers.Store(peerInfo.ID.String(), *peerInfo)
	}

	for i := 0; i < workers; i++ {
		go func() {
			// make sure that the
			for {
				if c.ctx.Err() != nil {
					log.Info("closing kdth discovery")
					return
				}
				// iterate through the peers
				count := 0

				peer, ok := connectablePeers.Next()
				if !ok {
					time.Sleep(1 * time.Second)
					continue
				}
				count++
				log.Debugf(" connecting", peer.ID.String())
				if err := h.Connect(c.ctx, peer); err != nil {
					log.Debug(err.Error())
					// remove unreacheable node from the list
					connectablePeers.Blacklist(peer.ID.String())
				} else {
					log.Debug("Connection established with bootstrap node:" + peer.ID.String())
					// If peer was connectable, req all the possible info from the peer and save it in the PSQL
					fpeer := c.ExtractHostInfo(peer)
					c.DB.StoreFilecoinPeer(fpeer.PeerId, fpeer)
					// try to request neighbors to connected peer
					neighborsRt, err := c.fetchNeighbors(c.ctx, peer)
					if err != nil {
						log.Debugf("unable to request neibors to peer. %s", err.Error())
					}
					// add peer to connectable list
					for _, newPeer := range neighborsRt.Neighbors {
						// neihbors is an array of AddrsInfo already have
						// add them to the peerQ
						connectablePeers.Store(newPeer.ID.String(), newPeer)
					}
				}
			}
		}()
	}

	// Print summary
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				// count blacklisted peers
				blacklisted := 0
				connectablePeers.blacklist.Range(func(key, value interface{}) bool {
					blacklisted++
					return true
				})
				connpeers := c.DB.GetFilecoinPeers()
				log.Infof("SUMMARY: pointer = %d", connectablePeers.p)
				log.Infof("SUMMARY: %d discovered peers, %d connectable, %d blacklisted", len(connectablePeers.pArray), len(connpeers), blacklisted)
			case <-c.ctx.Done():
				log.Info("closing routing")
				return
			}
		}
	}()

	log.Info("announcing ourselves...")
	//routingDiscovery := disc.NewRoutingDiscovery(kdht)
	log.Debug("successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.

	//wait untill the CNTL + C is recorded
	// register the shutdown signal
	signal_channel := make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	<-signal_channel
	c.cancel()
	// End up app, finishing everything
	log.Info("SHUTDOWN DETECTED!")

}

// TODO: initialize struct from PSQL DB

type discoveredPeers struct {
	pMap      sync.Map
	m         sync.Mutex
	pArray    []*peer.AddrInfo
	blacklist sync.Map
	p         int64
}

func NewDiscoveryPeers(db *postgresql.PostgresDBService) discoveredPeers {
	dp := discoveredPeers{
		pArray: make([]*peer.AddrInfo, 0),
	}
	// poblate the dp with peers in the DB
	peers := db.GetFilecoinPeers()
	for _, pID := range peers {
		p, ok := db.LoadFilecoinPeer(pID.String())
		if !ok {
			continue
		}
		// compose the addr info
		maddr := peer.AddrInfo{
			ID:    pID,
			Addrs: make([]ma.Multiaddr, 0),
		}
		maddr.Addrs = append(maddr.Addrs, p.MAddrs[:]...)
		dp.Store(pID.String(), maddr)
	}
	return dp
}

func (d *discoveredPeers) Next() (peer.AddrInfo, bool) {
	log.Debugf("next peer requested")
	if len(d.pArray) != 0 {
		log.Debugf("getting next peer")
		d.m.Lock()
		pinfo := d.pArray[d.p]
		d.p++
		// check d.p
		if d.p >= int64(len(d.pArray)) {
			d.p = 0
		}
		d.m.Unlock()
		if d.isBlacklisted(pinfo.ID.String()) {
			log.Warnf("peer blacklisted, try again")
			return peer.AddrInfo{}, false
		}
		log.Debugf("next peer: %d", pinfo.ID.String())
		return *pinfo, true
	} else {
		log.Debugf("array not init")
		return peer.AddrInfo{}, false
	}
}

func (d *discoveredPeers) Store(peerID string, p peer.AddrInfo) {
	log.Debugf("storing %d", peerID)
	_, ok := d.pMap.Load(peerID)
	// update results always
	d.pMap.Store(peerID, &p)
	if ok {
		return
	}
	// only if peer is not already in the array, we add it
	d.m.Lock()
	d.pArray = append(d.pArray, &p)
	d.m.Unlock()
	log.Debugf("done storing %d", peerID)
}

func (d *discoveredPeers) Blacklist(peerID string) {
	d.blacklist.Store(peerID, struct{}{})
}

func (d *discoveredPeers) isBlacklisted(peerID string) bool {
	d.m.Lock()
	// get pointer of the peerID
	_, ok := d.blacklist.Load(peerID)
	d.m.Unlock()
	return ok
}
