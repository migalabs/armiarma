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

	"github.com/migalabs/armiarma/src/db/postgresql"
	"github.com/migalabs/armiarma/src/utils/apis"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	ma "github.com/multiformats/go-multiaddr"
)

// TEMPORARY data for the running the filecoin demo
var (
	workers     = 100
	minWaitTime = 5 * time.Second
	protocols   = []string{
		"/ipfs/kad/1.0.0",
		"/ipfs/kad/2.0.0",
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
	connectablePeers := NewDiscoveryPeers(c.ctx, c.DB)
	npeer := connectablePeers.Run()

	// get official bootstrap peers
	bootstrapNodes := dht.GetDefaultBootstrapPeerAddrInfos()

	// Fill with bootstrap nodes
	log.Info("connecting to the bootstrap nodes")
	for _, peerInfo := range bootstrapNodes {
		// Load it to the sync map
		p := c.ExtractHostInfo(peerInfo)
		c.DB.StoreFilecoinPeer(peerInfo.ID.String(), p)
	}

	for i := 0; i < workers; i++ {
		go func() {
			workerid := i
			// request next peer
			connectablePeers.ReqNextPeer()
			// make sure that the
			for {
				select {
				case pid := <-npeer:
					// read the next peer id
					pID, _ := peer.Decode(pid)
					if err != nil {
						log.Errorf("error decoding PeerID string into peer.ID %s", err.Error())
					}
					// load the peer
					p, ok := c.DB.LoadFilecoinPeer(pid)
					if !ok {
						log.Error("error loading peer from db")
					}

					// get the multiaddress
					maddr := peer.AddrInfo{
						ID:    pID,
						Addrs: make([]ma.Multiaddr, 0),
					}
					maddr.Addrs = append(maddr.Addrs, p.MAddrs[:]...)
					// try to connect

					if len(maddr.Addrs) == 0 {
						time.Sleep(1 * time.Second)
						connectablePeers.ReqNextPeer()
						continue
					}
					log.Debugf(" connecting", pid)
					if err := h.Connect(c.ctx, maddr); err != nil {
						log.Error(err.Error())
						// remove unreacheable node from the list
						connectablePeers.Blacklist(pid)

					} else {
						log.Debug("Connection established with bootstrap node:" + pid)
						// If peer was connectable, req all the possible info from the peer and save it in the PSQL
						fpeer := c.ExtractHostInfo(maddr)
						c.DB.StoreFilecoinPeer(fpeer.PeerId, fpeer)
						// try to request neighbors to connected peer
						neighborsRt, err := c.fetchNeighbors(c.ctx, maddr)
						if err != nil {
							log.Debugf("unable to request neibors to peer. %s", err.Error())
						}
						// add peer to connectable list
						for _, newPeer := range neighborsRt.Neighbors {
							// neihbors is an array of AddrsInfo already have
							// generate new peer struct from AddrInfo
							fpeer := c.ExtractHostInfo(newPeer)
							c.DB.StoreFilecoinPeer(fpeer.PeerId, fpeer)
						}
					}
					// always after finishing, request new peer
					connectablePeers.ReqNextPeer()

				case <-c.ctx.Done():
					log.Infof("closing kdth discovery, worker %d", workerid)
					return
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
				log.Infof("SUMMARY: %d discovered peers, %d blacklisted", len(connpeers), blacklisted)
			case <-c.ctx.Done():
				log.Info("closing routing")
				return
			}
		}
	}()

	//log.Info("announcing ourselves...")
	//routingDiscovery := disc.NewRoutingDiscovery(kdht)
	//log.Debug("successfully announced!")

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
	ctx       context.Context
	pArray    []string
	blacklist sync.Map
	nPeerChan chan string
	nPeerReq  chan struct{}

	db *postgresql.PostgresDBService
}

func NewDiscoveryPeers(ctx context.Context, db *postgresql.PostgresDBService) discoveredPeers {
	dp := discoveredPeers{
		ctx:       ctx,
		pArray:    make([]string, 0),
		nPeerChan: make(chan string, workers),
		nPeerReq:  make(chan struct{}, workers),
		db:        db,
	}
	dp.refreshPeerList()
	return dp
}

func (d *discoveredPeers) refreshPeerList() {
	log.Debug("refreshing peer list")
	cnt := 0
	// poblate the dp with peers in the DB
	peers := d.db.GetFilecoinPeers()
	for _, pID := range peers {
		_, ok := d.db.LoadFilecoinPeer(pID.String())
		if !ok {
			continue
		}
		// add the
		d.pArray = append(d.pArray, pID.String())
		cnt++
	}
	log.Debugf("refreshed peerstore with %d peers", cnt)
}

func (d *discoveredPeers) Run() chan string {
	pointer := 0
	refreshFlag := false
	go func() {
		for {
			select {

			case <-d.nPeerReq:
				log.Debugf("peer requested")
				// check if array is empty
				if len(d.pArray) != 0 {
					pid := d.pArray[pointer]
					pointer++
					if d.isBlacklisted(pid) {
						d.nPeerReq <- struct{}{}
						continue
					}
					// return the pid to the worker
					log.Debugf("next peer: %d", pid)
					d.nPeerChan <- pid
				} else {
					log.Debugf("empty array")
					time.Sleep(minWaitTime)
					d.ReqNextPeer()
					refreshFlag = true
				}
				if refreshFlag || pointer >= len(d.pArray) {
					d.refreshPeerList()
				}

			case <-d.ctx.Done():
				log.Info("shutting down peer feeder")
				close(d.nPeerChan)
				close(d.nPeerReq)
				return
			}
		}
	}()
	return d.nPeerChan
}

func (d *discoveredPeers) ReqNextPeer() {
	d.nPeerReq <- struct{}{}
}

func (d *discoveredPeers) Blacklist(peerID string) {
	d.blacklist.Store(peerID, struct{}{})
}

func (d *discoveredPeers) isBlacklisted(peerID string) bool {
	// get pointer of the peerID
	_, ok := d.blacklist.Load(peerID)
	return ok
}
