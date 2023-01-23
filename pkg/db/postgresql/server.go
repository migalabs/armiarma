package postgresql

import (
	"context"
	"sync"
	"time"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils"
	log "github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

const (
	batchFlushingTimeout = 1 * time.Second
	batchSize            = 128
	maxPersisters        = 1

	noQueryError = "no error"
)

type DBClient struct {
	// Control Variables
	ctx context.Context
	m   sync.RWMutex

	// Network that we are Crawling
	Network utils.P2pNetwork

	// Pgx Postgres variables
	loginStr string
	psqlPool *pgxpool.Pool

	// Request channels
	persistC chan interface{}
}

func NewDBClient(
	ctx context.Context,
	p2pNetwork utils.P2pNetwork,
	loginStr string,
	initialized bool) (*DBClient, error) {

	logEntry := log.WithField("module", "db-client")
	logEntry.WithFields(log.Fields{"endpoint": loginStr}).Debug("attempt connection to DB")

	// check if the login string has enough len
	if len(loginStr) == 0 {
		return nil, errors.New("empty db-endpoint provided")
	}

	// try connecting to the DB from the given logingStr
	pPool, err := pgxpool.Connect(ctx, loginStr)
	if err != nil {
		return nil, err
	}

	// check if the connection is successful
	err = pPool.Ping(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to ping db")
	}

	// generate all the necessary/control channels
	persistC := make(chan interface{}, batchSize)

	// compose the DBClient
	dbClient := &DBClient{
		ctx:      ctx,
		Network:  p2pNetwork,
		loginStr: loginStr,
		psqlPool: pPool,
		persistC: persistC,
	}

	// initialize all the tables
	if initialized {
		err = dbClient.initTables()
		if err != nil {
			return nil, errors.Wrap(err, "unable to initialize the SQL tables at "+loginStr)
		}
	}

	// run the db persisters
	go dbClient.launchPersister()

	return dbClient, nil
}

func (c *DBClient) initTables() error {
	// initialize all the necesary tables to perform the crawl

	var err error

	// peer_info table
	err = c.InitPeerInfoTable()
	if err != nil {
		return errors.Wrap(err, "initializing peer_info table")
	}

	// conn_event
	err = c.InitConnEventTable()
	if err != nil {
		return errors.Wrap(err, "initializing peer_info table")
	}

	// ip table
	err = c.InitIpTable()
	if err != nil {
		return errors.Wrap(err, "initializing peer_info table")
	}

	// INIT all the tables - Separate Networks

	return err
}

func (c *DBClient) launchPersister() {
	log.Info("inititalizing db persister")
	go func() {
		// batch to aggregate all the queries
		batch := &pgx.Batch{}
		isReadyToPersistFn := func(batch *pgx.Batch) bool {
			return batch.Len() >= batchSize
		}
		batchQueryFn := func(batch *pgx.Batch, query string, args ...interface{}) {
			batch.Queue(query, args)
		}
		persistBatchFn := func(batch *pgx.Batch) error {
			log.Tracef("persisting batch of queries with len(%d)", batch.Len())
			t := time.Now()
			// begin pgx.Tx
			tx, err := c.psqlPool.Begin()
			if err != nil {
				return errors.Wrap(err, "unable to perist batch")
			}
			// Add batch to TX
			batchResults := tx.SendBatch(c.ctx, batch)

			// Exec the queries
			var qerr error
			var rows pgx.Rows
			for qerr == nil {
				rows, qerr = batchResults.Query()
				rows.Close()
			}
			// check if there was any error
			if qerr.Error() != noQueryError {
				return errors.Wrap(err, "unable to persist batch")
			}

			// after peristing the batch, we can already flush all the
			batch = &pgx.Batch{}

			log.Tracef("batch persisted in %s", time.Since(t))
			return nil
		}

		// batch flushing ticker
		ticker := time.NewTicker(batchFlushingTimeout)

		for {
			// check with higher priority if the main-ctx died
			select {
			case <-c.ctx.Done(): // check if the context of the tool died
				log.Info("context died, clossing persister")
				return
			default:
			}

			// load  or flush after
			select {
			case obj := <-c.persistC: // persist any kind of item
				// Every item/SQL query  has to return (string. []interfaces)
				switch obj.(type) {
				case (*models.HostInfo):
					hostInfo := obj.(*models.HostInfo)
					log.Tracef("persisting peer %s\n", hostInfo.ID.String())

					// add raw new HostInfo
					q, args := c.UpsertHostInfo(hostInfo)
					batchQueryFn(batch, q, args...)

					// check if the peerInfo needs to update anything else
					if hostInfo.IsHostIdentified() {
						q, args = c.UpsertPeerInfo()
						batchQueryFn(batch, q, args...)
					}

				case (*models.ConnEvent):
					connEvent := obj.(*models.ConnEvent)
					log.Tracef("persisting conn_event for peer %s\n", connEvent.PeerID.String())
					q, args := c.InsertNewConnEvent(connEvent)
					batchQueryFn(batch, q, args...)

				case (*models.ConnAttempt):
					connAttempt := obj.(*models.ConnAttempt)
					log.Tracerf("persisting conn_attempt")

				case (*models.IpInfo):
					ipInfo := obj.(*models.IpInfo)
					log.Tracef("persisting ip_info %s\n", ipInfo.IP)
					q, args := c.UpsertIpInfo(*ipInfo)
					batchQueryFn(batch, q, args...)

				default:
					log.Error("unrecognized type of object received to persist into DB", obj)
				}
				// after adding whatever query we got check if we need to persist the batch
				if isReadyToPersistFn(batch) {
					err := persistBatchFn(batch)
					if err != nil {
						log.Error(err)
					}
				}

			case <-ticker.C:
				log.Trace("ticker jumped - flushing content of query-batch")
				// flush the batched queries
				err := persistBatchFn(batch)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}()
}

func (c *DBClient) Close() {
	// close safelly the connection with PSQL
	c.psqlPool.Close()

	// close all the exisiting channels
	close(c.persistC)
}

func (c *DBClient) PersistToDB(persItem interface{}) {
	log.Tracef("persisting item: %T\n", persItem)
	c.persistC <- persItem
}

func (c *DBClient) SingleQuery(query string, args ...interface{}) (interface{}, error) {
	return c.psqlPool.Exec(c.ctx, query, args...)
}
