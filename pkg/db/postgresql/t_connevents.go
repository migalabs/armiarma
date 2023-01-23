package postgresql

import (
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (c *DBClient) InitConnEventTable() error {
	log.Debug("initializing conn_events table in db")

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
