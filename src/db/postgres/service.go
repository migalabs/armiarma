package postgres

import (
	"context"
	"strings"

	pgx "github.com/jackc/pgx/v4"
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
	psqlConn      *pgx.Conn
	// Metric Variables
	// TODO: missing some sort of local-crawler identification fields
	// 		 like: location, IP, ID, etc
}

func ConnectToDB(ctx context.Context, url string) (*PostgresDBService, error) {
	mainCtx, cancel := context.WithCancel(ctx)
	// spliting the url to don't share any confidential information on logs
	log.Infof("Conneting to postgres DB %s", url)
	if strings.Contains(url, "@") {
		log.Debugf("Connecting to PostgresDB at %s", strings.Split(url, "@")[1])
	}
	psqlConn, err := pgx.Connect(mainCtx, url)
	if err != nil {
		return nil, err
	}
	if strings.Contains(url, "@") {
		log.Infof("PostgresDB %s succesfully connected", strings.Split(url, "@")[1])
	}
	psqlDB := &PostgresDBService{
		ctx:           mainCtx,
		cancel:        cancel,
		connectionUrl: url,
		psqlConn:      psqlConn,
	}
	// init the psql db
	err = psqlDB.init()
	if err != nil {
		return psqlDB, errors.Wrap(err, "error initializing the tables of the psqldb")
	}
	return psqlDB, err
}

// TODO: missing:
// 				- create tables
//				- insert/store item
// 				- read/load item

// Initialize all the DBs creating tables and making sure that everything is ready to start crawling
func (p *PostgresDBService) init() error {

	// IMPORTANT: !!!!! When the table is initialized, the peer connected need to be disconnected
	// TODO:
	err := p.createPeerTable()
	if err != nil {
		return err
	}
	return nil

}

func (p *PostgresDBService) Type() string {
	return PsqlType
}

func (p *PostgresDBService) Close() {
	log.Debug("Closing ProstgresDB")
	p.psqlConn.Close(p.ctx)
	p.cancel()
}
