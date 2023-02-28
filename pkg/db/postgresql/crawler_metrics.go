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

func (db *DBClient) GetOsDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{}, 0)
	rows, err := db.psqlPool.Query(
		db.ctx,
		`
		SELECT
			client_os,
			count(client_os) as nodes
		FROM peer_info
		WHERE deprecated='false' and attempted='true' and client_name IS NOT NULL
		GROUP BY client_os
		ORDER BY nodes DESC;
		`)
	if err != nil {
		return summary, err
	}
	for rows.Next() {
		var os string
		var count int
		err = rows.Scan(&os, &count)
		summary[os] = count
	}
	return summary, nil
}

func (db *DBClient) GetArchDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{}, 0)
	rows, err := db.psqlPool.Query(
		db.ctx,
		`
		SELECT
			client_arch,
			count(client_arch) as nodes
		FROM peer_info
		WHERE deprecated='false' and attempted='true' and client_name IS NOT NULL
		GROUP BY client_arch
		ORDER BY nodes DESC;
		`)
	if err != nil {
		return summary, err
	}
	for rows.Next() {
		var arch string
		var count int
		err = rows.Scan(&arch, &count)
		summary[arch] = count
	}
	return summary, nil
}

func (db *DBClient) GetHostingDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{})
	// get the number of mobile hosts
	var mobile int
	err := db.psqlPool.QueryRow(
		db.ctx,
		`
		SELECT 
			count(aux.mobile) as mobile
		FROM (
			SELECT
				pi.peer_id,
				pi.attempted,
				pi.client_name,
				pi.deprecated,
				pi.ip,
				ips.mobile
			FROM peer_info as pi
			INNER JOIN ips ON pi.ip=ips.ip
			WHERE pi.deprecated='false' and attempted = 'true' and client_name IS NOT NULL and ips.mobile='true'
		) as aux
		`).Scan(&mobile)
	if err != nil {
		return summary, err
	}
	summary["mobile_ips"] = mobile

	// get the number of proxy peers
	var proxy int
	err = db.psqlPool.QueryRow(
		db.ctx,
		`
		SELECT 
			count(aux.proxy) as under_proxy
		FROM (
			SELECT
				pi.peer_id,
				pi.attempted,
				pi.client_name,
				pi.deprecated,
				pi.ip,
				ips.proxy
			FROM peer_info as pi
			INNER JOIN ips ON pi.ip=ips.ip
			WHERE pi.deprecated='false' and attempted = 'true' and client_name IS NOT NULL and ips.proxy='true'
		) as aux
		`).Scan(&proxy)
	if err != nil {
		return summary, err
	}
	summary["under_proxy"] = proxy

	// get the number of hosted IPs
	var hosted int
	err = db.psqlPool.QueryRow(
		db.ctx,
		`
		SELECT 
			count(aux.hosting) as hosted_ips
		FROM (
			SELECT
				pi.peer_id,
				pi.attempted,
				pi.client_name,
				pi.deprecated,
				pi.ip,
				ips.hosting
			FROM peer_info as pi
			INNER JOIN ips ON pi.ip=ips.ip
			WHERE pi.deprecated='false' and attempted = 'true' and client_name IS NOT NULL and ips.hosting='true'
		) as aux		
		`).Scan(&hosted)
	if err != nil {
		return summary, err
	}
	summary["hosted_ips"] = hosted
	return summary, nil
}
