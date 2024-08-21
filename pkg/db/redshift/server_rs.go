package redshift

import (
	"context"
	"database/sql"
	//"strconv"
	//"strings"
	"sync"
	"time"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/gossipsub"
	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
	"github.com/migalabs/armiarma/pkg/utils"
	log "github.com/sirupsen/logrus"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	batchFlushingTimeout = 1 * time.Second
	batchSize            = 512
	maxPersisters        = 2
)

var (
	noQueryError  string = "no error"
	noQueryResult string = "no result"
)

type DBClient struct {
	// Control Variables
	ctx                 context.Context
	dailyBackupInterval time.Duration

	// Network that we are Crawling
	Network utils.NetworkType

	// SQL DB variables
	loginStr   string
	redshiftDB *sql.DB

	// Request channels
	persistC chan interface{}
	doneC    chan struct{}
	wg       *sync.WaitGroup

	// Control Variables
	persistConnEvents bool
}

func NewDBClient(
	ctx context.Context,
	p2pNetwork utils.NetworkType,
	loginStr string,
	dailyBackupInt time.Duration,
	options ...DBOption) (*DBClient, error) {
	// check if the login string has enough len
	if len(loginStr) == 0 {
		return nil, errors.New("empty db-endpoint provided")
	}

	// try connecting to the DB from the given loginStr
	redshiftDB, err := sql.Open("postgres", loginStr)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{"endpoint": loginStr}).Debug("successful connection to DB")

	// check if the connection is successful
	err = redshiftDB.PingContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to ping db through dbWriter")
	}

	// generate all the necessary/control channels
	persistC := make(chan interface{}, batchSize)

	var wg sync.WaitGroup

	// compose the DBClient
	dbClient := &DBClient{
		ctx:                 ctx,
		dailyBackupInterval: dailyBackupInt,
		Network:             p2pNetwork,
		loginStr:            loginStr,
		redshiftDB:          redshiftDB,
		persistC:            persistC,
		doneC:               make(chan struct{}),
		wg:                  &wg,
		persistConnEvents:   true,
	}

	// Check for all the available options
	for _, opt := range options {
		err := opt(dbClient)
		if err != nil {
			return nil, err
		}
	}

	// run the db persisters
	for i := 0; i < maxPersisters; i++ {
		go dbClient.launchPersister()
	}
	// launch the daily backup heartbeat
	go dbClient.dailyBackupheartbeat()
	return dbClient, nil
}

func (c *DBClient) initTables() error {
	// initialize all the necessary tables to perform the crawl

	var err error

	// peer_info table
	err = c.InitPeerInfoTable()
	if err != nil {
		return errors.Wrap(err, "initializing peer_info table")
	}

	// conn_event
	err = c.InitConnEventTable()
	if err != nil {
		return errors.Wrap(err, "initializing conn_events table")
	}

	// ip table
	err = c.InitIpTable()
	if err != nil {
		return errors.Wrap(err, "initializing ips table")
	}

	// active peers' backup
	err = c.InitActivePeersTable()
	if err != nil {
		return errors.Wrap(err, "initializing active_peers backup")
	}

	switch c.Network {
	// ETHEREUM
	case utils.EthereumNetwork:
		// eth_nodes table
		err = c.InitEthNodesTable()
		if err != nil {
			return errors.Wrap(err, "initializing eth_nodes table")
		}

		// eth_status table
		err = c.InitEthereumNodeStatus()
		if err != nil {
			return errors.Wrap(err, "initializing eth_status table")
		}

		// gossipsub messages
		// eth_attestation
		err = c.InitEthereumAttestationsTable()
		if err != nil {
			return errors.Wrap(err, "initializing eth_attestations table")
		}
		// eth blocks
		err = c.InitEthereumBeaconBlocksTable()
		if err != nil {
			return errors.Wrap(err, "initializing eth_blocks table")
		}
	//IPFS
	// FILECOIN
	default:

	}

	return err
}

func (c *DBClient) launchPersister() {
	logEntry := log.WithFields(log.Fields{
		"mod": "db-persister",
	})
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		// batch to aggregate all the queries
		batch := NewQueryBatch(c.ctx, c.redshiftDB, batchSize)

		// batch flushing ticker
		ticker := time.NewTicker(batchFlushingTimeout)

		var readyToFinish bool

	persistingLoop:
		for {
			if readyToFinish && len(c.persistC) == 0 {
				break persistingLoop
			}

			// check with higher priority if the main-ctx died
			select {
			case <-c.ctx.Done(): // check if the context of the tool died
				logEntry.Info("context died, closing persister")
				readyToFinish = true
			case <-c.doneC:
				logEntry.Info("closed detected, closing persister")
				readyToFinish = true
			default:
			}

			// load  or flush after
			select {
			case obj := <-c.persistC: // persist any kind of item
				// Every item/SQL query  has to return (string, []interface{})
				switch obj.(type) {
				case (*models.HostInfo):
					hostInfo := obj.(*models.HostInfo)
					logEntry.Tracef("persisting host_info %s\n", hostInfo.ID.String())
					// // double-check when are we rewriting hInfo without IP, and port
					// if hostInfo.IP == "" {
					// 	log.Error("error trying to add host info without IP and ports", hostInfo)
					// }
					// add raw new HostInfo
					q, args := c.UpsertHostInfo(hostInfo)
					batch.AddQuery(q, args...)

					// check if the peerInfo needs to update anything else
					if hostInfo.IsHostIdentified() {
						logEntry.Tracef("host_info has peer_info %s\n", hostInfo.PeerInfo.RemotePeer.String())
						q, args = c.UpdatePeerInfo(&hostInfo.PeerInfo)
						batch.AddQuery(q, args...)
					}
					// Read all the Attributes in hInfo
					for attName, att := range hostInfo.Attr {
						log.Debugf("detected attribute %s on peer", attName)
						switch att.(type) {
						case eth.BeaconStatusStamped:
							bstatus := att.(eth.BeaconStatusStamped)
							q, args = c.UpsertEthereumNodeStatus(bstatus)
							batch.AddQuery(q, args...)
						case eth.BeaconMetadataStamped:
							bmetadata := att.(eth.BeaconMetadataStamped)
							q, args = c.UpsertEthereumNodeMetadata(bmetadata)
							batch.AddQuery(q, args...)
						case (*eth.EnrNode):
							enrNode := att.(*eth.EnrNode)
							logEntry.Tracef("persisting eth node_info %s\n", enrNode.ID.String())
							q, args := c.UpsertEnrInfo(enrNode)
							batch.AddQuery(q, args...)
						default:
							log.Warnf("not yet recognized type for attr %s - %T - %+v", attName, att, att)
						}
					}

				case (*models.PeerInfo):
					peerInfo := obj.(*models.PeerInfo)
					logEntry.Tracef("persisting new peer_info %s\n", peerInfo.RemotePeer.String())
					q, args := c.UpdatePeerInfo(peerInfo)
					batch.AddQuery(q, args...)

				case (*models.ConnectionAttempt):
					connAttempt := obj.(*models.ConnectionAttempt)
					logEntry.Tracef("persisting conn_attempt")
					q, args := c.UpdateConnAttempt(connAttempt)
					batch.AddQuery(q, args...)

				case (*models.ConnEvent):
					connEvent := obj.(*models.ConnEvent)
					logEntry.Tracef("persisting conn_event for peer %s\n", connEvent.PeerID.String())
					if c.persistConnEvents {
						q, args := c.InsertNewConnEvent(connEvent)
						batch.AddQuery(q, args...)
					}
					// Control Info LastActivity based on last disconnection
					// get the disconnection time to update the LastActivity timestamp in the peer_info table
					q, args := c.UpdateLastActivityTimestamp(connEvent.PeerID, connEvent.DiscTime)
					batch.AddQuery(q, args...)

				case (models.IpInfo):
					ipInfo := obj.(models.IpInfo)
					logEntry.Tracef("persisting ip_info %s\n", ipInfo.IP)
					q, args := c.UpsertIpInfo(ipInfo)
					batch.AddQuery(q, args...)

				// GossipSub Messages
				case (gossipsub.PersistableMsg):
					prsMsg := obj.(gossipsub.PersistableMsg)
					// select the type of message inside the list of messages
					switch prsMsg.(type) {
					case (*eth.TrackedAttestation):
						attMsg := prsMsg.(*eth.TrackedAttestation)
						log.Tracef("persisting eth_attestation %s", attMsg.MsgID)
						q, args := c.InsertNewEthereumAttestation(attMsg)
						batch.AddQuery(q, args...)
					case (*eth.TrackedBeaconBlock):
						bblockMsg := prsMsg.(*eth.TrackedBeaconBlock)
						log.Tracef("persisting eth_block %s", bblockMsg.MsgID)
						q, args := c.InsertNewEthereumBeaconBlock(bblockMsg)
						batch.AddQuery(q, args...)
					}
				default:
					logEntry.Errorf("unrecognized type of object received to persist into DB %T", obj)
					logEntry.Error(obj)
				}

				// after adding whatever query we got check if we need to persist the batch
				if batch.IsReadyToPersist() {
					logEntry.Debug("batch-query full, ready to persist")
					err := batch.PersistBatch()
					if err != nil {
						log.Error(err)
					}
				}

			case <-ticker.C:
				logEntry.Trace("ticker jumped - flushing content of query-batch")
				// flush the batched queries
				err := batch.PersistBatch()
				if err != nil {
					log.Error(err)
				}
			}
		}
	}()
}

func (c *DBClient) dailyBackupheartbeat() {
	// make a first backup of the active peers(if any)
	err := c.activePeersBackup()
	if err != nil {
		log.Error(err)
	}
	ticker := time.NewTicker(c.dailyBackupInterval)
	for {
		select {
		case <-ticker.C:
			err := c.activePeersBackup()
			if err != nil {
				log.Error(err)
			}
		case <-c.ctx.Done():
			return
		}
	}

}

func (c *DBClient) Close() {
	// Let the persister finish cleaning the batch
	c.doneC <- struct{}{}
	c.wg.Wait()

	err := c.activePeersBackup()
	if err != nil {
		log.Error(err)
	}
	// close safely the connection with Redshift
	c.redshiftDB.Close()

	// close all the existing channels
	close(c.persistC)
}

func (c *DBClient) PersistToDB(persItem interface{}) {
	c.persistC <- persItem
}

func (c *DBClient) SingleQuery(query string, args ...interface{}) (interface{}, error) {
	return c.redshiftDB.ExecContext(c.ctx, query, args...)
}


