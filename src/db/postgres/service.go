package postgres

import (
	"context"
	"strings"

	pgx "github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
)

// Static postgres queries, for each modification in the tables, the table needs to be reseted
var (
	// logrus associated with the postgres db
	ModuleName = "POSTGRES-DB"
	log        = logrus.WithField(
		"module", ModuleName,
	)
)

type PostgresDBService struct {
	// Control Variables
	ctx           context.Context
	cancel        context.CancelFunc
	connectionUrl string // the url might not be necessary (better to remove it?¿)
	conn          *pgx.Conn
	// Metric Variables
	// TODO: missing some sort of local-crawler identification fields
	// 		 like: location, IP, ID, etc
}

func ConnectToDB(ctx context.Context, url string) (*PostgresDBService, error) {
	mainCtx, cancel := context.WithCancel(ctx)
	log.Debugf("Connecting to PostgresDB at %s", strings.Split(url, "@")[1])
	psqlConn, err := pgx.Connect(mainCtx, url)
	if err != nil {
		return nil, err
	}
	log.Infof("PostgresDB %s succesfully connected", strings.Split(url, "@")[1])
	psqlDB := &PostgresDBService{
		ctx:           mainCtx,
		cancel:        cancel,
		connectionUrl: url,
		conn:          psqlConn,
	}
	return psqlDB, err
}

// TODO: missing:
// 				- create tables
//				- insert/store item
// 				- read/load item

func (p *PostgresDBService) Close() {
	log.Debug("Closing ProstgresDB")
	p.conn.Close()
	p.cancel()
}
