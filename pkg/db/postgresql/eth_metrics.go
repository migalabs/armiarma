package postgresql

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (db *DBClient) GetNodePerForkDistribution() (map[string]interface{}, error) {
	log.Debug("fetching node per fork distribution")
	nodeDist := make(map[string]interface{})

	rows, err := db.psqlPool.Query(
		db.ctx,
		`
		SELECT aux.fork_digest,
			count(aux.fork_digest) as cnt
		FROM (
			SELECT pi.peer_id,
				pi.ip,
				pi.port,
				pi.deprecated,
				pi.attempted,
				eth.fork_digest
			FROM peer_info as pi
			LEFT JOIN eth_nodes as eth ON pi.ip=eth.ip and pi.port=eth.tcp
			WHERE pi.deprecated='false' and pi.attempted='true' and eth.fork_digest IS NOT NULL
		) as aux
		GROUP BY fork_digest
		ORDER BY cnt DESC;
		`,
	)
	// make sure we close the rows and we free the connection/session
	defer rows.Close()
	if err != nil {
		return nodeDist, errors.Wrap(err, "unable to fetch node per fork distribution")
	}

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

func (db *DBClient) GetAttnetsDistribution() (map[string]interface{}, error) {
	log.Debug("fetching attnets distribution")
	nodeDist := make(map[string]interface{})

	rows, err := db.psqlPool.Query(
		db.ctx,
		`
		SELECT aux.attnets_number as attnets,
			count(aux.attnets_number) as cnt
		FROM (
			SELECT pi.peer_id,
				pi.ip,
				pi.port,
				pi.deprecated,
				pi.attempted,
				eth.attnets_number
			FROM peer_info as pi
			LEFT JOIN eth_nodes as eth ON pi.ip=eth.ip and pi.port=eth.tcp
			WHERE pi.deprecated='false' and pi.attempted='true' and eth.attnets_number >= 0
		) as aux
		GROUP BY attnets
		ORDER BY attnets ASC;
		`,
	)
	// make sure we close the rows and we free the connection/session
	defer rows.Close()
	if err != nil {
		return nodeDist, errors.Wrap(err, "unable to fetch attnet distribution")
	}

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
