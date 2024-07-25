package redshift

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	LastActivityValidRange = 180 // 6 Months
)


// GetClientDistribution fetches client distribution metrics from Redshift
func (db *DBClient) GetClientDistribution() (map[string]interface{}, error) {
	log.Debug("fetching client distribution metrics")
	cliDist := make(map[string]interface{})

	rows, err := db.psqlPool.QueryContext(
		db.ctx,
		`
		SELECT 
			client_name, count(client_name) as count
		FROM peer_info
		WHERE 
			deprecated = 'false' and 
		    attempted = 'true' and 
		    client_name IS NOT NULL and 
		    timestamp 'epoch' + last_activity * interval '1 second' > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		GROUP BY client_name
		ORDER BY count DESC;
		`,
		LastActivityValidRange,
	)
	defer rows.Close()
	if err != nil {
		fmt.Print("\n", err.Error())
		return cliDist, errors.Wrap(err, "unable to fetch client distribution")
	}

	for rows.Next() {
		var cliName string
		var count int
		err = rows.Scan(&cliName, &count)
		if err != nil {
			return cliDist, errors.Wrap(err, "unable to parse fetched client distribution")
		}
		cliDist[cliName] = count
	}

	return cliDist, nil
}

// GetVersionDistribution fetches version distribution metrics from Redshift
func (db *DBClient) GetVersionDistribution() (map[string]interface{}, error) {
	log.Debug("fetching client distribution metrics")
	verDist := make(map[string]interface{})

	rows, err := db.psqlPool.QueryContext(
		db.ctx,
		`
		SELECT client_name,
			client_version, 
			count(client_version) as cnt
		FROM peer_info
		WHERE 
			deprecated = 'false' and 
			attempted = 'true' and 
			client_name IS NOT NULL and 
			timestamp 'epoch' + last_activity * interval '1 second' > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		GROUP BY client_name, client_version
		ORDER BY client_name DESC, cnt DESC;
		`,
		LastActivityValidRange,
	)
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
			return verDist, errors.Wrap(err, "unable to parse fetched client distribution")
		}
		verDist[cliName+"_"+cliVersion] = count
	}

	return verDist, nil
}

// GetGeoDistribution fetches geo distribution metrics from Redshift
func (db *DBClient) GetGeoDistribution() (map[string]interface{}, error) {
	log.Debug("fetching geo distribution metrics")
	geoDist := make(map[string]interface{})

	rows, err := db.psqlPool.QueryContext(
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
			WHERE deprecated = 'false' and 
			      attempted = 'true' and 
			      client_name IS NOT NULL and 
			      timestamp 'epoch' + last_activity * interval '1 second' > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		) as aux 
		GROUP BY country_code
		ORDER BY cnt DESC;
		`,
		LastActivityValidRange,
	)
	defer rows.Close()
	if err != nil {
		return geoDist, errors.Wrap(err, "unable to fetch geo distribution")
	}

	for rows.Next() {
		var country string
		var count int
		err = rows.Scan(&country, &count)
		if err != nil {
			return geoDist, errors.Wrap(err, "unable to parse fetched geo distribution")
		}
		geoDist[country] = count
	}

	return geoDist, nil
}

// GetOsDistribution fetches OS distribution metrics from Redshift
func (db *DBClient) GetOsDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{})
	rows, err := db.psqlPool.QueryContext(
		db.ctx,
		`
		SELECT
			client_os,
			count(client_os) as nodes
		FROM peer_info
		WHERE deprecated='false' and 
		      attempted='true' and 
		      client_name IS NOT NULL and 
		      timestamp 'epoch' + last_activity * interval '1 second' > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		GROUP BY client_os
		ORDER BY nodes DESC;
		`,
		LastActivityValidRange,
	)
	defer rows.Close()
	if err != nil {
		return summary, err
	}
	for rows.Next() {
		var os string
		var count int
		err = rows.Scan(&os, &count)
		if err != nil {
			return summary, err
		}
		summary[os] = count
	}
	return summary, nil
}

// GetArchDistribution fetches architecture distribution metrics from Redshift
func (db *DBClient) GetArchDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{})
	rows, err := db.psqlPool.QueryContext(
		db.ctx,
		`
		SELECT
			client_arch,
			count(client_arch) as nodes
		FROM peer_info
		WHERE deprecated='false' and 
		      attempted='true' and 
		      client_name IS NOT NULL and 
		      timestamp 'epoch' + last_activity * interval '1 second' > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		GROUP BY client_arch
		ORDER BY nodes DESC;
		`,
		LastActivityValidRange,
	)
	defer rows.Close()
	if err != nil {
		return summary, err
	}
	for rows.Next() {
		var arch string
		var count int
		err = rows.Scan(&arch, &count)
		if err != nil {
			return summary, err
		}
		summary[arch] = count
	}
	return summary, nil
}

// GetHostingDistribution fetches hosting distribution metrics from Redshift
func (db *DBClient) GetHostingDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// get the number of mobile hosts
	var mobile int
	err := db.psqlPool.QueryRowContext(
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
			WHERE pi.deprecated='false' and 
			      attempted = 'true' and 
			      client_name IS NOT NULL and 
			      ips.mobile='true' and 
			      timestamp 'epoch' + last_activity * interval '1 second' > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		) as aux
		`,
		LastActivityValidRange,
	).Scan(&mobile)
	if err != nil {
		return summary, err
	}
	summary["mobile_ips"] = mobile

	// get the number of proxy peers
	var proxy int
	err = db.psqlPool.QueryRowContext(
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
			WHERE pi.deprecated='false' and 
			      attempted = 'true' and 
			      client_name IS NOT NULL and 
			      ips.proxy='true' and 
			      timestamp 'epoch' + last_activity * interval '1 second' > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		) as aux
		`,
		LastActivityValidRange,
	).Scan(&proxy)
	if err != nil {
		return summary, err
	}
	summary["under_proxy"] = proxy

	// get the number of hosted IPs
	var hosted int
	err = db.psqlPool.QueryRowContext(
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
			WHERE pi.deprecated='false' and 
			      attempted = 'true' and 
			      client_name IS NOT NULL and 
			      ips.hosting='true' and 
			      timestamp 'epoch' + last_activity * interval '1 second' > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		) as aux		
		`,
		LastActivityValidRange,
	).Scan(&hosted)
	if err != nil {
		return summary, err
	}
	summary["hosted_ips"] = hosted
	return summary, nil
}

// GetRTTDistribution fetches RTT distribution metrics from Redshift
func (db *DBClient) GetRTTDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	rows, err := db.psqlPool.QueryContext(
		db.ctx,
		`
		SELECT 
			t.latency as latency_range,
			count(*) as nodes 
		FROM (
			SELECT
				CASE
					WHEN latency between 0 and 100 THEN ' 0-100ms'
					WHEN latency between 101 and 200 THEN '101-200ms'
					WHEN latency between 201 and 300 THEN '201-300ms'
					WHEN latency between 301 and 400 THEN '301-400ms'     
					WHEN latency between 401 and 500 THEN '401-500ms'     
					WHEN latency between 501 and 600 THEN '501-600ms'      
					WHEN latency between 601 and 700 THEN '601-700ms'     
					WHEN latency between 701 and 800 THEN '701-800ms'
					WHEN latency between 801 and 900 THEN '801-900ms'
					WHEN latency between 901 and 1000 THEN '901-1000ms'     
					ELSE '+1s' 
				END as latency    
			FROM peer_info 
			WHERE deprecated=false and 
			      client_name IS NOT NULL and 
			      timestamp 'epoch' + last_activity * interval '1 second' > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
		) as t 
		GROUP BY t.latency 
		ORDER BY nodes DESC;	
		`,
		LastActivityValidRange,
	)
	defer rows.Close()
	if err != nil {
		return summary, err
	}

	for rows.Next() {
		var rttRange string
		var rttValue int
		err = rows.Scan(
			&rttRange,
			&rttValue,
		)
		if err != nil {
			return summary, err
		}
		summary[rttRange] = rttValue
	}
	return summary, nil
}

// GetIPDistribution fetches IP distribution metrics from Redshift
func (db *DBClient) GetIPDistribution() (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	rows, err := db.psqlPool.QueryContext(
		db.ctx,
		`
		SELECT 
			nodes as nodes_per_ip, 
			count(t.nodes) as number_of_ips 
		FROM ( 
			SELECT 
				ip, 
				count(ip) as nodes 
			FROM peer_info 
			WHERE deprecated = false and 
			      client_name IS NOT NULL and 
			      timestamp 'epoch' + last_activity * interval '1 second' > CURRENT_TIMESTAMP - ($1 * INTERVAL '1 DAY')
			GROUP BY ip 
			ORDER BY nodes DESC 
		) as t 
		GROUP BY nodes 
		ORDER BY number_of_ips DESC;	
		`,
		LastActivityValidRange,
	)
	defer rows.Close()
	if err != nil {
		return summary, err
	}

	for rows.Next() {
		var nodesPerIP int
		var ips int
		err = rows.Scan(
			&nodesPerIP,
			&ips,
		)
		if err != nil {
			return summary, err
		}
		summary[fmt.Sprintf("%d", nodesPerIP)] = ips
	}
	return summary, nil
}
