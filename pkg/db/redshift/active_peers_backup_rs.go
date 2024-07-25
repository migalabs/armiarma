package redshift

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	_ "github.com/lib/pq"
)

var (
	LastActivityValidRange = 180 // 6 Months
)

// DropActivePeersTable drops the active_peers table in Redshift
func (c *DBClient) DropActivePeersTable() error {
	log.Info("dropping table active_peers")
	_, err := c.psqlPool.ExecContext(
		c.ctx,
		`
		DROP TABLE IF EXISTS active_peers;
		`,
	)
	return err
}

// InitActivePeersTable initializes the active_peers table in Redshift
func (c *DBClient) InitActivePeersTable() error {
	log.Info("init active_peers table")

	_, err := c.psqlPool.ExecContext(
		c.ctx,
		`
		CREATE TABLE IF NOT EXISTS active_peers(
			id INTEGER IDENTITY(1,1),
			timestamp TIMESTAMP,
			peers BIGINT[],

			PRIMARY KEY(timestamp)			
		);
		`,
	)
	return err
}

// getActivePeers retrieves active peer IDs from the peer_info table
func (c *DBClient) getActivePeers() ([]int, error) {
	activePeers := make([]int, 0)

	rows, err := c.psqlPool.QueryContext(
		c.ctx,
		`
		SELECT
			id,
			peer_id
		FROM peer_info
		WHERE deprecated = 'false' and attempted = 'true' and client_name IS NOT NULL and timestamp 'epoch' + last_activity * interval '1 second' > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		`,
		LastActivityValidRange,
	)
	if err != nil {
		return activePeers, errors.Wrap(err, "unable to retrieve active peer's ids")
	}
	defer rows.Close()

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

// activePeersBackup backs up the list of active peers in the active_peers table
func (c *DBClient) activePeersBackup() error {
	log.Debug("making backup in DB of the actual active peers")

	activePeers, err := c.getActivePeers()
	if err != nil {
		return errors.Wrap(err, "unable to backup active peers")
	}
	if len(activePeers) <= 0 {
		log.Infof("tried to persist %d active peers (skipped)", len(activePeers))
		return nil
	}

	// backup the list of active peers
	_, err = c.psqlPool.ExecContext(
		c.ctx,
		`
		INSERT INTO active_peers(
			timestamp,
			peers)
		VALUES ($1, $2)
		`,
		time.Now(),
		activePeers,
	)

	return err
}
