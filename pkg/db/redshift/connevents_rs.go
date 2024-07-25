package redshift

import (
	"context"
	"database/sql"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	_ "github.com/lib/pq"
)


// InitConnEventTable initializes the conn_events table in Redshift
func (c *DBClient) InitConnEventTable() error {
	log.Debugf("init conn_events table in Redshift\n")

	_, err := c.psqlPool.ExecContext(c.ctx, `
		CREATE TABLE IF NOT EXISTS conn_events(
			id INTEGER IDENTITY(1,1),
			peer_id TEXT NOT NULL,
			direction TEXT NOT NULL,
			conn_time BIGINT NOT NULL, 
			latency BIGINT,
			disconn_time BIGINT NOT NULL,
			identified BOOL,
			error TEXT NOT NULL,

			PRIMARY KEY (id)
		);
		`)

	if err != nil {
		return errors.Wrap(err, "initializing conn_events table")
	}

	return nil
}

// InsertNewConnEvent inserts a new connection event into the conn_events table
func (c *DBClient) InsertNewConnEvent(connEv *models.ConnEvent) (query string, args []interface{}) {
	log.Trace("inserting new connection event to conn_event in Redshift")
	// compose query
	query = `
		INSERT INTO conn_events (
			peer_id,
			direction,
			conn_time, 
			latency,
			disconn_time,
			identified,
			error)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`

	args = append(args, connEv.PeerID.String())
	args = append(args, models.DirectionIndexToString(connEv.Direction))
	args = append(args, connEv.ConnTime.Unix())
	args = append(args, connEv.Latency.Milliseconds())
	args = append(args, connEv.DiscTime.Unix())
	args = append(args, connEv.Identified)
	args = append(args, connEv.Error)

	return query, args
}
