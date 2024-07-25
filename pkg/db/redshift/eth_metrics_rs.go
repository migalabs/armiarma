package redshift

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	_ "github.com/lib/pq"
)


// GetNodePerForkDistribution fetches node distribution per fork from Redshift
func (db *DBClient) GetNodePerForkDistribution() (map[string]interface{}, error) {
	log.Debug("fetching node per fork distribution")
	nodeDist := make(map[string]interface{})

	rows, err := db.psqlPool.QueryContext(
		db.ctx,
		`
		SELECT 
			aux.fork_digest,
			count(aux.fork_digest) as cnt
		FROM (
			SELECT 
				CURRENT_TIMESTAMP as c_t,
				timestamp 'epoch' + timestamp * interval '1 second' as t_s,
				fork_digest
			FROM eth_nodes
			WHERE fork_digest IS NOT NULL and timestamp 'epoch' + timestamp * interval '1 second' > CURRENT_TIMESTAMP - INTERVAL '1 DAY'
		) as aux
		GROUP BY fork_digest
		ORDER BY cnt DESC;
		`,
	)
	if err != nil {
		return nodeDist, errors.Wrap(err, "unable to fetch node per fork distribution")
	}
	defer rows.Close()

	for rows.Next() {
		var forkName string
		var count int
		err = rows.Scan(&forkName, &count)
		if err != nil {
			return nodeDist, errors.Wrap(err, "unable to parse fetched node per fork distribution")
		}
		nodeDist[forkName] = count
	}

	return nodeDist, nil
}

// GetAttnetsDistribution fetches attnets distribution from Redshift
func (db *DBClient) GetAttnetsDistribution() (map[string]interface{}, error) {
	log.Debug("fetching attnets distribution")
	nodeDist := make(map[string]interface{})

	rows, err := db.psqlPool.QueryContext(
		db.ctx,
		`
		SELECT aux.attnets_number as attnets,
			count(aux.attnets_number) as cnt
		FROM (
		SELECT 
				CURRENT_TIMESTAMP as c_t,
				timestamp 'epoch' + timestamp * interval '1 second' as t,
				fork_digest,
				attnets_number
			FROM eth_nodes
			WHERE fork_digest IS NOT NULL and timestamp 'epoch' + timestamp * interval '1 second' > CURRENT_TIMESTAMP - INTERVAL '1 DAY'
		) as aux
		GROUP BY attnets
		ORDER BY cnt DESC;	
		`,
	)
	if err != nil {
		return nodeDist, errors.Wrap(err, "unable to fetch attnet distribution")
	}
	defer rows.Close()

	for rows.Next() {
		var attnets int
		var count int
		err = rows.Scan(&attnets, &count)
		if err != nil {
			return nodeDist, errors.Wrap(err, "unable to parse fetched attnets distribution")
		}
		nodeDist[fmt.Sprintf("%d", attnets)] = count
	}

	return nodeDist, nil
}

// GetDeprecatedNodes fetches the count of deprecated nodes from Redshift
func (db *DBClient) GetDeprecatedNodes() (int, error) {
	log.Debug("fetching deprecated node count")

	var deprecatedCount int
	err := db.psqlPool.QueryRowContext(
		db.ctx,
		`
		SELECT
			count(deprecated)
		FROM peer_info
		WHERE deprecated='true';
		`).Scan(
		&deprecatedCount,
	)
	if err != nil {
		return deprecatedCount, errors.Wrap(err, "unable to fetch deprecated node count")
	}

	return deprecatedCount, nil
}
