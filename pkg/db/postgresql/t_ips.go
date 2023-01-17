package postgresql

import (
	"fmt"
	"time"

	"github.com/migalabs/armiarma/pkg/db/models"

	"github.com/pkg/errors"
)

func (c *DBClient) InitIpTable() error {
	log.Debug("init ips table")

	_, err := c.psqlPool.Exec(c.ctx, `
		CREATE TABLE IF NOT EXISTS ips(
			id SERIAL,
			ip TEXT NOT NULL,
			expiration_time TIMESTAMP NOT NULL,
			continent TEXT NOT NULL,
			continent_code TEXT NOT NULL,
			country TEXT NOT NULL,
			country_code TEXT NOT NULL,
			region TEXT NOT NULL,
			region_name TEXT NOT NULL,
			city TEXT NOT NULL,
			zip TEXT NOT NULL,
			lat REAL NOT NULL,
			lon REAL NOT NULL,
			isp TEXT NOT NULL,
			org TEXT NOT NULL,
			as_raw TEXT NOT NULL,
			asname TEXT NOT NULL,
			mobile BOOL NOT NULL,
			proxy BOOL NOT NULL,
			hosting BOOL NOT NULL,

			PRIMARY KEY (ip)
		);
	`)
	if err != nil {
		return errors.Wrap(err, "error init ips table")
	}
	return nil
}

func (c *DBClient) InsertNewIP(ipInfo models.IpInfo) error {
	log.Debugf("insert ip %s to ips table", ipInfo.IP)

	_, err := c.psqlPool.Exec(c.ctx, `
		INSERT INTO ips(
			ip,
			expiration_time,
			continent,
			continent_code,
			country,
			country_code,
			region,
			region_name,
			city,
			zip,
			lat,
			lon,
			isp,
			org,
			as_raw,
			asname,
			mobile,
			proxy,
			hosting)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
	`,
		ipInfo.IP,
		ipInfo.ExpirationTime,
		ipInfo.Continent,
		ipInfo.ContinentCode,
		ipInfo.Country,
		ipInfo.CountryCode,
		ipInfo.Region,
		ipInfo.RegionName,
		ipInfo.City,
		ipInfo.Zip,
		ipInfo.Lat,
		ipInfo.Lon,
		ipInfo.Isp,
		ipInfo.Org,
		ipInfo.As,
		ipInfo.AsName,
		ipInfo.Mobile,
		ipInfo.Proxy,
		ipInfo.Hosting,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error inserting ip %s to ips table", ipInfo.IP))
	}
	return nil
}

func (c *DBClient) UpdateIP(ipInfo models.IpInfo) error {
	log.Debugf("update ip %s to ips table", ipInfo.IP)

	_, err := c.psqlPool.Exec(c.ctx, `
		UPDATE ips SET
			expiration_time = $2,
			continent = $3,
			continent_code = $4,
			country = $5,
			country_code = $6,
			region = $7,
			region_name = $8,
			city = $9,
			zip = $10,
			lat = $11,
			lon = $12,
			isp = $13,
			org = $14,
			as_raw = $15,
			asname = $16,
			mobile = $17,
			proxy = $18,
			hosting = $19
		WHERE ip=$1;
	`,
		ipInfo.IP,
		ipInfo.ExpirationTime,
		ipInfo.Continent,
		ipInfo.ContinentCode,
		ipInfo.Country,
		ipInfo.CountryCode,
		ipInfo.Region,
		ipInfo.RegionName,
		ipInfo.City,
		ipInfo.Zip,
		ipInfo.Lat,
		ipInfo.Lon,
		ipInfo.Isp,
		ipInfo.Org,
		ipInfo.As,
		ipInfo.AsName,
		ipInfo.Mobile,
		ipInfo.Proxy,
		ipInfo.Hosting,
	)
	if err != nil {
		return errors.Wrap(err, "error updating ips table")
	}
	return nil
}

func (c *DBClient) ReadIpInfo(ip string) (models.IpInfo, error) {
	log.Debugf("reading ip info for ip %s", ip)

	var ipInfo models.IpInfo

	err := c.psqlPool.QueryRow(c.ctx, `
		SELECT 
			ip,
			expiration_time,
			continent,
			continent_code,
			country,
			country_code,
			region,
			region_name,
			city,
			zip,
			lat,
			lon,
			isp,
			org,
			as_raw,
			asname,
			mobile,
			proxy,
			hosting
		FROM ips
		WHERE ip=$1
	`, ip).Scan(
		&ipInfo.IP,
		&ipInfo.ExpirationTime,
		&ipInfo.Continent,
		&ipInfo.ContinentCode,
		&ipInfo.Country,
		&ipInfo.CountryCode,
		&ipInfo.Region,
		&ipInfo.RegionName,
		&ipInfo.City,
		&ipInfo.Zip,
		&ipInfo.Lat,
		&ipInfo.Lon,
		&ipInfo.Isp,
		&ipInfo.Org,
		&ipInfo.As,
		&ipInfo.AsName,
		&ipInfo.Mobile,
		&ipInfo.Proxy,
		&ipInfo.Hosting,
	)
	if err != nil {
		return models.IpInfo{}, err
	}

	return ipInfo, nil

}

func (c *DBClient) GetExpiredIpInfo() ([]string, error) {

	expIps := make([]string, 0)

	ipRows, err := c.psqlPool.Query(c.ctx, `
		SELECT ip 
		FROM ips
		WHERE expiration_time < NOW();
	`)
	if err != nil {
		return expIps, errors.Wrap(err, "unable to get expired ip records")
	}

	defer ipRows.Close()

	for ipRows.Next() {
		var ip string
		err := ipRows.Scan(&ip)
		if err != nil {
			return expIps, errors.Wrap(err, "error parsing readed row for expired ip records")
		}
		expIps = append(expIps, ip)
	}

	return expIps, nil
}

func (c *DBClient) CheckIpRecords(ip string) (exists bool, expired bool, err error) {

	log.Debugf("checkign if ip %s exists in ips table", ip)

	var expTime time.Time
	var readIp string

	err = c.psqlPool.QueryRow(c.ctx, `
		SELECT ip, expiration_time
		FROM ips
		WHERE ip=$1;
	`, ip).Scan(&readIp, &expTime)

	if err != nil {
		return
	}

	if readIp == ip {
		exists = true
	}
	if expTime.Before(time.Now()) {
		expired = true
	}
	return
}
