package postgresql

import (
	"time"
    "strings"

	"github.com/jackc/pgx/v4"
	"github.com/migalabs/armiarma/pkg/db/models"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var ispCleaner = []struct {
	substr, name string
}{
    {"amazon", "Amazon"},
    {"google", "Google"},
    {"microsoft", "Microsoft"},
    {"oracle", "Oracle"},
    {"at&t", "AT&T"},
    {"vodafone", "Vodafone"},
    {"orange", "Orange"},
    {"china mobile", "China Mobile"},
    {"china telecom", "China Telecom"},
    {"alibaba", "Alibaba"},
    {"pt comunicacoes", "PT Comunicacoes"},
    {"swisscom", "Swisscom"},
    {"sony", "Sony"},
    {"telecom argentina", "Telecom Argentina"},
    {"ovh", "OVH"},
    {"t-mobile", "T-Mobile"},
    {"hetzner", "Hetzner"},
    {"digitalocean", "DigitalOcean"},
    {"verizon", "Verizon"},
    {"virgin media", "Virgin Media"},
    {"hostinger", "Hostinger"},
    {"telefonica", "Telefonica"},
    {"contabo", "Contabo"},
    {"mevspace", "Mevspace"},
    {"chinanet", "Chinanet"},
    {"kamatera", "Kamatera"},
    {"teraswitch", "TeraSwitch"},
    {"emirates telecommunications", "Etisalat"},
    {"emirates integrated telecommunications", "du"},
    {"centurylink", "CenturyLink"},
    {"huawei", "Huawei"},
    {"frontier communications", "Frontier Communications"},
    {"charter communications", "Charter Communications"},
    {"digi ", "DIGI"},
    {"akamai", "Akamai Technologies"},
    {"china unicom", "China Unicom"},
    {"telus communications", "TELUS Communications"},
    {"datacamp limited", "DataCamp"},
    {"limestone", "Limestone Networks"},
    {"hong kong telecommunications", "Hong Kong Telecommunications"},
    {"velia.net", "velia.net"},
    {"comcast", "Comcast"},
    {"init7", "Init7"},
    {"hivelocity", "Hivelocity"},
    {"leaseweb", "LeaseWeb"},
    {"fornex hosting", "Fornex Hosting"},
    {"servers.com", "Servers.com"},
    {"nos comunicacoes", "NOS Comunicacoes"},
}

func CleanISP(isp string) (string) {
	for _, ispClean := range ispCleaner{
		if strings.Contains(strings.ToLower(isp), ispClean.substr){
			return ispClean.name
		}
	}
	return isp
}

func (c *DBClient) InitIpTable() error {
	log.Debug("init ips table in psql-db")
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

// UpsertIP attemtps to insert IP in the DB - or Updates the data info if they where already there
func (c *DBClient) UpsertIpInfo(ipInfo models.IpInfo) (query string, args []interface{}) {
	log.Trace("upsert ip_info in psql-db")
	// compose query
	query = `
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
		ON CONFLICT (ip)
		DO UPDATE SET
			expiration_time = excluded.expiration_time,
			continent = excluded.continent,
			continent_code = excluded.continent_code,
			country = excluded.country,
			country_code = excluded.country_code,
			region = excluded.region,
			region_name = excluded.region_name,
			city = excluded.city,
			zip = excluded.zip,
			lat = excluded.lat,
			lon = excluded.lon,
			isp = excluded.isp,
			org = excluded.org,
			as_raw = excluded.as_raw,
			asname = excluded.asname,
			mobile = excluded.mobile,
			proxy = excluded.proxy,
			hosting = excluded.hosting;
		`

	args = append(args, ipInfo.IP)
	args = append(args, ipInfo.ExpirationTime)
	args = append(args, ipInfo.Continent)
	args = append(args, ipInfo.ContinentCode)
	args = append(args, ipInfo.Country)
	args = append(args, ipInfo.CountryCode)
	args = append(args, ipInfo.Region)
	args = append(args, ipInfo.RegionName)
	args = append(args, ipInfo.City)
	args = append(args, ipInfo.Zip)
	args = append(args, ipInfo.Lat)
	args = append(args, ipInfo.Lon)
	args = append(args, CleanISP(ipInfo.Isp))
	args = append(args, ipInfo.Org)
	args = append(args, ipInfo.As)
	args = append(args, ipInfo.AsName)
	args = append(args, ipInfo.Mobile)
	args = append(args, ipInfo.Proxy)
	args = append(args, ipInfo.Hosting)

	return query, args
}

// ReadIpInfo reads all the information available for that specific IP in the DB
func (c *DBClient) ReadIpInfo(ip string) (models.IpInfo, error) {
	log.Tracef("reading ip_info for ip %s from psql-db", ip)
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

// GetExpiredIpInfo returns all the IP whos' TTL has already expired
func (c *DBClient) GetExpiredIpInfo() ([]string, error) {
	log.Trace("fetching expired ips from psql-db")
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

// CheckIpRecords checks if a given IP is already stored in the DB as whether its TTL has expired
func (c *DBClient) CheckIpRecords(ip string) (exists bool, expired bool, err error) {
	log.Tracef("checking if ip %s exists in ips table", ip)
	var readIp string
	var expTime time.Time

	row := c.psqlPool.QueryRow(c.ctx, `
		SELECT 
			ip,
			expiration_time
		FROM ips
		WHERE ip=$1;
	`, ip)

	err = row.Scan(&readIp, &expTime)
	if err == pgx.ErrNoRows {
		return false, false, nil
	} else if err != nil {
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
