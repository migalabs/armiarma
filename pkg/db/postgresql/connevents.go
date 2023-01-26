package postgresql

import (
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (c *DBClient) InitConnEventTable() error {
	log.Debugf("init conn_events table in psql-db\n")

	_, err := c.psqlPool.Exec(c.ctx, `
		CREATE TABLE IF NOT EXISTS conn_events(
			id SERIAL,
			peer_id TEXT NOT NULL,
			direction TEXT NOT NULL,
			conn_time BIGINT NOT NULL, 
			latency BIGINT,
			disconn_time BIGINT NOT NULL,
			identified BOOL,
			error TEXT NOT NULL,

			PRIMARY KEY (id),
			UNIQUE(peer_id, conn_time)
		);
		`)

	if err != nil {
		return errors.Wrap(err, "initializing conn_events table")
	}

	return nil
}

func (c *DBClient) InsertNewConnEvent(connEv *models.ConnEvent) (query string, args []interface{}) {
	log.Trace("inserting new connection event to conn_event in psql-db")
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
