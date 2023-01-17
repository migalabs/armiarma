package postgresql

import "github.com/pkg/errors"

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



		PRIMARY KEY(id),
		FOREIGN KEY(peer_id) REFERENCES peer_info(peer_id)
		)
	`)

	if err != nil {
		return errors.Wrap(err, "unable to create peer_control_info table")
	}

	return nil
}

func (c *DBClient) InsertNewPeerControlInfo(cInfo *models.ControlInfo) error {

	return nil
}
