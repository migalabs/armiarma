package postgresql

import (
	"github.com/migalabs/armiarma/pkg/db/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (c *DBClient) InitPeerControlTable() error {
	log.Debug("initializing peer_control_info table in db")

	_, err := c.psqlPool.Exec(c.ctx, `
		CREATE TABLE IF NOT EXISTS peer_control_info(
			id SERIAL
			peer_id TEXT NOT NULL,
			deprecated BOOL NOT NULL,
			left_network BOOL NOT NULL,
			ident_state TEXT NOT NULL,
			last_activity BIGINT NOT NULL, 
			last_conn_attempt BIGINT NOT NULL,
			last_error TEXT,
			next_conn_delay BIGINT NOT NULL,

			PRIMARY KEY(id))
	`)

	if err != nil {
		return errors.Wrap(err, "unable to create peer_control_info table")
	}

	return nil
}

func (c *DBClient) UpsertPeerControlInfo(cInfo *models.ControlInfo) (query string, args []interface{}) {
	// Compose query
	query = `
		INSERT INTO peer_control_info(
			peer_id,
			deprecated,
			left_network,
			ident_state,
			last_activity,
			last_conn_attempt,
			last_error,
			next_conn_delay)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT ON CONSTRAINT peer_id
			SET UPDATE
			deprecated = excluded.deprecated,
			left_network = excluded.left_network,
			ident_state = excluded.ident_state,
			last_activity = excluded.last_activity,
		`
	args = append(args, cInfo.LastActivity)
	args = append(args, cInfo.LastActivity)
	args = append(args, cInfo.LastActivity)
	args = append(args, cInfo.LastActivity)
	args = append(args, cInfo.LastActivity)

	return query, args
}
