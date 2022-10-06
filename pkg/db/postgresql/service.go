package postgresql

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Static postgres queries, for each modification in the tables, the table needs to be reseted
var (
	// logrus associated with the postgres db
	PsqlType = "postgres-db"
	log      = logrus.WithField(
		"module", PsqlType,
	)
)

type PostgresDBService struct {
	// Control Variables
	ctx           context.Context
	cancel        context.CancelFunc
	connectionUrl string // the url might not be necessary (better to remove it?¿)
	psqlPool      *pgxpool.Pool
	// Network DB Model
	netModel NetworkModel
}

// Connect to the PostgreSQL Database and get the multithread-proof connection
// from the given url-composed credentials
func ConnectToDB(ctx context.Context, url string, dbmodel NetworkModel) (*PostgresDBService, error) {
	mainCtx, cancel := context.WithCancel(ctx)
	// spliting the url to don't share any confidential information on logs
	log.Infof("Conneting to postgres DB %s", url)
	if strings.Contains(url, "@") {
		log.Debugf("Connecting to PostgresDB at %s", strings.Split(url, "@")[1])
	}
	psqlPool, err := pgxpool.Connect(mainCtx, url)
	if err != nil {
		return nil, err
	}
	if strings.Contains(url, "@") {
		log.Infof("PostgresDB %s succesfully connected", strings.Split(url, "@")[1])
	}
	// filter the type of network that we are filtering

	psqlDB := &PostgresDBService{
		ctx:           mainCtx,
		cancel:        cancel,
		connectionUrl: url,
		psqlPool:      psqlPool,
		netModel:      dbmodel,
	}
	// init the psql db
	err = psqlDB.netModel.init(ctx, psqlDB.psqlPool)
	if err != nil {
		return psqlDB, errors.Wrap(err, "error initializing the tables of the psqldb")
	}
	return psqlDB, err
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) Close() {
}

type NetworkModel interface {
	init(context.Context, *pgxpool.Pool) error
	// Peer model related
	StorePeer(context.Context, *pgxpool.Pool, string, models.Peer)
	LoadPeer(context.Context, *pgxpool.Pool, string) (models.Peer, bool)
	DeletePeer(context.Context, *pgxpool.Pool, string)
	GetPeers(context.Context, *pgxpool.Pool) []peer.ID
	CheckPeersSummaryTableStatus(context.Context, *pgxpool.Pool) bool
	GetConnectedPeers(context.Context, *pgxpool.Pool) ([]string, error)
	GetNumberOfPeers(context.Context, *pgxpool.Pool) (int, error)
	GetLastActivityTime(context.Context, *pgxpool.Pool) (time.Time, error)
	// Client diversity related model
	StoreClientDiversitySnapshot(context.Context, *pgxpool.Pool, models.ClientDiversity) error
	LoadClientDiversitySnapshot(context.Context, *pgxpool.Pool, time.Time) (models.ClientDiversity, error)
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) StorePeer(id string, pm models.Peer) {
	p.netModel.StorePeer(p.ctx, p.psqlPool, id, pm)
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) LoadPeer(id string) (models.Peer, bool) {
	return p.netModel.LoadPeer(p.ctx, p.psqlPool, id)
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) DeletePeer(id string) {
	p.netModel.DeletePeer(p.ctx, p.psqlPool, id)
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) GetPeers() []peer.ID {
	return p.netModel.GetPeers(p.ctx, p.psqlPool)
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) CheckPeersSummaryTableStatus() bool {
	return p.netModel.CheckPeersSummaryTableStatus(p.ctx, p.psqlPool)
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) GetConnectedPeers() ([]string, error) {
	return p.netModel.GetConnectedPeers(p.ctx, p.psqlPool)
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) GetNumberOfPeers() (int, error) {
	return p.netModel.GetNumberOfPeers(p.ctx, p.psqlPool)
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) GetLastActivityTime() (time.Time, error) {
	return p.netModel.GetLastActivityTime(p.ctx, p.psqlPool)
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) StoreClientDiversitySnapshot(cliDiver models.ClientDiversity) error {
	return p.netModel.StoreClientDiversitySnapshot(p.ctx, p.psqlPool, cliDiver)
}

// Close the connection with the PostgreSQL
func (p *PostgresDBService) LoadClientDiversitySnapshot(qTime time.Time) (models.ClientDiversity, error) {
	return p.netModel.LoadClientDiversitySnapshot(p.ctx, p.psqlPool, qTime)
}
