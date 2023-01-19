package postgresql

import (
	"context"
	"fmt"
	"sync"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

const (
	bufferSize    = 2048
	maxPersisters = 1
)

type DBClient struct {
	// Control Variables
	ctx context.Context
	m   sync.RWMutex

	// Network that we are Crawling
	Network utils.NetworkType

	// Pgx Postgres variables
	loginStr string
	psqlPool *pgxpool.Pool

	// Request channels

	persistC chan interface{}
}

func NewDBClient(
	ctx context.Context,
	NetworkType utils.NetworkType,
	loginStr string,
	initialized bool) (*DBClient, error) {

	logEntry := logrus.WithField("module", "db-client")
	logEntry.WithFields(logrus.Fields{"endpoint": loginStr}).Debug("attempt connection to DB")

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
	persistC := make(chan interface{}, bufferSize)

	// compose the DBClient
	dbClient := &DBClient{
		ctx:      ctx,
		Network:  NetworkType,
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
	go dbClient.spawnPersisters()

	return dbClient, nil
}

func (c *DBClient) initTables() error {
	// initialize all the necesary tables to perform the crawl

	var err error

	// peer_info table
	err = c.initPeerInfoTable()
	if err != nil {
		return errors.Wrap(err, "initializing peer_info table")
	}

	// ip table

	return err
}

func (c *DBClient) spawnPersisters() {

	var totPersisters int

	// spaw as many persisters as defined in `maxPersisters`
	for persister := 1; persister <= maxPersisters; persister++ {
		c.launchPersister(persister)
		totPersisters++
	}

	logrus.Debugf("spawned total of %d db persister", totPersisters)
}

func (c *DBClient) launchPersister(persisterID int) error {
	logEntry := logrus.WithFields(logrus.Fields{"persisterID": persisterID})

	go func() {
		logEntry.Info("inititalizing persister")

		// check with higher priority if the main-ctx died
		select {
		case <-c.ctx.Done(): // check if the context of the tool died
			logEntry.Info("context died, clossing persister")
			return

		case obj := <-c.persistC: // persist any kind of item
			switch obj.(type) {
			case (*models.PeerInfo):
				peerInfo := obj.(*models.PeerInfo)
				logrus.Debugf("persisting peer %s", peerInfo.ID.String())

			case (*models.ConnEvent):
				connEvent := obj.(*models.ConnEvent)
				logrus.Debugf("persisting peer %s", connEvent.PeerID.String())

			default:
				logEntry.Error("unrecognized type of object received to persist into DB", obj)
			}
		}

	}()

	return nil

}

func (c *DBClient) Close() {
	// close safelly the connection with PSQL
	c.psqlPool.Close()

	// close all the exisiting channels
	close(c.persistC)
}

func (c *DBClient) PersistToDB(persItem interface{}) {
	fmt.Println("persisting ", persItem)
	c.persistC <- persItem
}
