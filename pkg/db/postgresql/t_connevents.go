package postgresql

import (
	"fmt"

	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/pkg/errors"
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
	// FOREIGN KEY(peer_id) REFERENCES peer_info(peer_id)

	if err != nil {
		return errors.Wrap(err, "initializing conn_events table")
	}

	return nil
}

func (c *DBClient) InsertNewConnEvent(connEv *models.ConnEvent) error {
	// we might have to check whether there is already a peer_id in the table peer_info
	fmt.Println("inserting", connEv)
	// add a simple new row
	tag, err := c.psqlPool.Exec(c.ctx, `
		INSERT INTO conn_events (
			peer_id,
			direction,
			conn_time, 
			latency,
			disconn_time,
			identified,
			error)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`,
		connEv.PeerID.String(),
		models.DirectionIndexToString(connEv.Direction),
		connEv.ConnTime.Unix(),
		connEv.Latency.Milliseconds(),
		connEv.DiscTime.Unix(),
		connEv.Identified,
		connEv.Error,
	)
	if err != nil {
		return errors.Wrap(err, "unable to insert conn event")
	}

	fmt.Println(tag)
	return nil
}
