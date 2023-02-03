package postgresql

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// this file contains all the list of queries to extract the metrics from the Crawler (as agnostic as possible from the network)

// Basic call over the whole list of non-deprecated peers
func (db *DBClient) GetClientDistribution() (map[string]interface{}, error) {
	log.Debug("fetching client distribution metrics")
	cliDist := make(map[string]interface{}, 0)

	rows, err := db.psqlPool.Query(
		db.ctx,
		`
		SELECT 
			client_name, count(client_name) as count
		FROM peer_info
		WHERE 
			deprecated = 'false' and attempted = 'true' and client_name IS NOT NULL
		GROUP BY client_name
		ORDER BY count DESC;
		`,
	)
	// make sure we close the rows and we free the connection/session
	defer rows.Close()
	if err != nil {
		return cliDist, errors.Wrap(err, "unable to fetch client distribution")
	}

	for rows.Next() {
		var cliName string
		var count int
		err = rows.Scan(&cliName, &count)
		if err != nil {
			return cliDist, errors.Wrap(err, "unable to parse fetch client distribution")
		}
		cliDist[cliName] = count
	}

	return cliDist, nil
}

// Basic call over the whole list of non-deprecated peers
func (db *DBClient) GetVersionDistribution() (map[string]interface{}, error) {
	log.Debug("fetching client distribution metrics")
	verDist := make(map[string]interface{}, 0)

	rows, err := db.psqlPool.Query(
		db.ctx,
		`
		SELECT client_name,
			client_version, 
			count(client_version) as cnt
		FROM peer_info
		WHERE 
			deprecated = 'false' and attempted = 'true' and client_name IS NOT NULL
		GROUP BY client_name, client_version
		ORDER BY client_name DESC, cnt DESC;
		`,
	)
	// make sure we close the rows and we free the connection/session
	defer rows.Close()
	if err != nil {
		return verDist, errors.Wrap(err, "unable to fetch client distribution")
	}

	for rows.Next() {
		var cliName string
		var cliVersion string
		var count int
		err = rows.Scan(&cliName, &cliVersion, &count)
		if err != nil {
			return verDist, errors.Wrap(err, "unable to parse fetch client distribution")
		}
		verDist[cliName+"_"+cliVersion] = count
	}

	return verDist, nil
}

// Basic call over the whole list of non-deprecated peers
func (db *DBClient) GetGeoDistribution() (map[string]interface{}, error) {
	log.Debug("fetching client distribution metrics")
	geoDist := make(map[string]interface{}, 0)

	rows, err := db.psqlPool.Query(
		db.ctx,
		`
		SELECT 
			aux.country_code as country_code,
			count(aux.country_code) as cnt
		FROM (
			SELECT peer_info.peer_id, 
				ips.ip,
				ips.country_code
			FROM peer_info
			RIGHT JOIN ips on peer_info.ip = ips.ip
			WHERE deprecated = 'false' and attempted = 'true' and client_name IS NOT NULL
		) as aux 
		GROUP BY country_code
		ORDER BY cnt DESC;
		`,
	)
	// make sure we close the rows and we free the connection/session
	defer rows.Close()
	if err != nil {
		return geoDist, errors.Wrap(err, "unable to fetch client distribution")
	}

	for rows.Next() {
		var country string
		var count int
		err = rows.Scan(&country, &count)
		if err != nil {
			return geoDist, errors.Wrap(err, "unable to parse fetch client distribution")
		}
		geoDist[country] = count
	}

	return geoDist, nil
}
