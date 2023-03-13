package postgresql

import (
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (c *DBClient) DropActivePeersTable() error {
	log.Info("dropping table active_peers")
	_, err := c.psqlPool.Exec(
		c.ctx,
		`
		DROP TABLE active_peers;
		`,
	)
	return err
}

func (c *DBClient) InitActivePeersTable() error {
	log.Info("init active_peers table")

	_, err := c.psqlPool.Exec(
		c.ctx,
		`
			CREATE TABLE IF NOT EXISTS active_peers(
				id SERIAL,
				timestamp TIMESTAMP,
				peers BIGINT[],

				PRIMARY KEY(timestamp)			
			);
		`,
	)
	return err
}

func (c *DBClient) getActivePeers() ([]int, error) {
	activePeers := make([]int, 0)

	rows, err := c.psqlPool.Query(
		c.ctx,
		`
		SELECT
			id,
			peer_id
		FROM peer_info
		WHERE deprecated='false'
		`,
	)
	if err != nil {
		return activePeers, errors.Wrap(err, "unable to retrieve active peer's ids")
	}

	for rows.Next() {
		var id int
		var pId string
		err = rows.Scan(&id, &pId)
		if err != nil {
			return activePeers, errors.Wrap(err, "unable to retrieve active peer's ids")
		}
		activePeers = append(activePeers, id)
	}

	return activePeers, nil
}

func (c *DBClient) activePeersBackup() error {
	log.Debug("making backup in DB of the actual active peers")

	activePeers, err := c.getActivePeers()
	if err != nil {
		return errors.Wrap(err, "unable to backup active peers")
	}

	// backup the list of active peers
	_, err = c.psqlPool.Exec(
		c.ctx,
		`
			INSERT INTO active_peers(
				timestamp,
				peers)
			VALUES ($1,$2)
		`,
		time.Now(),
		activePeers,
	)

	return err
}
