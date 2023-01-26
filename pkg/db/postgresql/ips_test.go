package postgresql

import (
	"context"
	"testing"

	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/migalabs/armiarma/pkg/utils/apis"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	validTestIps = []string{
		"180.215.214.130",
		"86.85.31.80",
		"81.174.250.125",
		"34.107.92.185",
		"180.150.120.141",
		"159.196.123.200",
		"34.96.169.237",
		"71.244.131.129",
		"178.128.149.36",
		"213.89.168.210",
		"104.248.233.81",
		"124.148.222.208",
		"84.10.105.207",
		"185.125.111.100",
		"18.198.21.237",
		"34.203.178.51",
		"92.6.61.195",
		"157.245.255.33",
		"104.173.236.54",
		"83.148.225.33",
		"18.138.231.16",
		"73.153.58.102",
		"143.110.239.135",
		"92.116.77.117",
		"58.96.228.225",
		"13.36.230.182",
		"104.175.205.114",
		"35.234.106.26",
		"172.105.110.9",
		"144.91.122.45",
		"47.157.209.183",
		"47.28.73.5",
		"45.129.183.26",
		"115.236.175.123",
		"50.88.27.211",
		"93.138.104.226",
		"50.35.78.231",
		"54.92.10.151",
		"34.139.96.113",
		"178.128.253.77",
		"101.127.67.207",
		"54.229.169.184",
		"34.92.177.243",
		"173.212.226.174",
		"91.173.185.10",
		"157.230.87.172",
		"43.231.209.121",
		"72.206.111.193",
		"70.244.32.109",
		"100.25.98.222",
		"35.193.91.168",
		"172.103.79.177",
		"5.173.204.2",
		"52.202.245.254",
		"162.255.89.192",
		"104.248.199.44",
		"165.22.84.144",
		"34.230.90.108",
		"3.87.117.41",
		"34.146.226.24",
		"18.212.179.142",
		"100.26.52.161",
		"212.95.50.85",
		"44.208.35.134",
		"54.167.125.23",
		"188.172.228.127",
		"138.201.120.161",
		"177.106.145.30",
		"67.209.54.149",
		"178.39.190.120",
		"84.246.85.71",
		"74.70.109.33",
		"45.119.155.42",
		"34.90.241.78",
		"52.91.29.254",
		"51.195.234.11",
		"130.44.150.189",
		"65.109.60.42",
		"172.105.21.150",
		"54.88.78.205",
		"35.194.125.249",
		"107.23.211.226",
		"72.50.194.201",
		"108.50.218.17",
		"135.181.56.50",
		"52.200.193.4",
		"142.132.140.50",
		"34.234.97.125",
		"81.83.6.224",
		"144.76.18.69",
		"178.63.77.34",
		"35.79.212.192",
		"172.105.66.39",
		"50.16.14.183",
		"3.145.88.212",
		"181.94.225.237",
		"54.162.188.62",
		"35.194.124.209",
		"95.217.87.121",
		"24.99.221.90",
		"86.80.132.169",
		"176.9.50.93",
		"35.171.23.239",
		"54.198.17.90",
		"89.39.106.192",
		"142.132.155.124",
		"65.108.71.187",
		"35.230.63.210",
		"35.196.123.187",
		"212.71.250.196",
		"112.2.231.236",
		"173.249.42.23",
		"116.202.234.116",
		"82.64.241.190",
		"54.145.57.0",
		"213.239.197.194",
		"103.85.38.143",
		"34.244.194.73",
		"82.64.121.18",
		"97.121.197.191",
		"156.146.44.237",
		"209.90.107.213",
		"84.17.38.158",
		"3.87.176.192",
		"164.107.116.121",
		"13.215.67.61",
		"188.40.141.251",
		"211.199.167.48",
		"65.109.68.102",
		"51.210.118.58",
		"54.208.87.95",
		"185.229.191.152",
		"185.86.10.133",
		"99.246.130.70",
		"98.38.120.152",
		"34.234.63.64",
		"45.72.108.162",
		"13.38.25.126",
		"185.132.133.129",
		"185.100.212.49",
		"146.190.235.31",
		"35.240.0.83",
		"69.230.152.160",
		"63.251.87.47",
		"116.88.132.125",
		"141.95.33.156",
		"172.104.111.251",
		"34.239.187.121",
		"65.109.66.238",
		"89.58.47.137",
		"82.66.160.125",
		"221.223.25.99",
		"185.242.11.191",
		"40.127.209.81",
		"172.56.105.14",
		"67.250.102.122",
		"132.145.49.100",
		"146.59.116.113",
		"104.57.185.124",
		"45.32.132.77",
		"63.33.70.163",
		"47.215.136.252",
		"51.195.234.146",
		"176.9.58.130",
		"18.141.232.118",
		"51.79.228.226",
		"51.77.120.45",
		"162.19.88.160",
		"162.55.53.170",
		"119.224.66.248",
		"176.9.65.94",
		"35.205.14.119",
		"157.90.176.112",
		"62.227.95.249",
		"79.137.113.17",
		"100.24.114.157",
		"143.177.112.71",
		"18.234.24.204",
		"66.27.123.186",
		"65.30.42.134",
		"119.123.206.55",
		"34.235.117.9",
		"35.194.125.249",
		"54.198.160.85",
		"34.146.226.24",
		"62.212.75.3",
		"146.190.42.212",
		"150.136.221.48",
		"107.23.211.5",
		"167.235.34.105",
		"95.111.253.25",
		"101.191.162.212",
		"121.36.207.94",
		"107.6.94.205",
		"193.164.249.99",
		"65.108.120.223",
		"18.193.122.40",
		"34.122.254.225",
		"54.234.178.195",
		"18.184.212.84",
		"88.99.103.98",
		"3.1.230.234",
		"69.10.37.122",
	}
	privTestIps = []string{
		"192.168.0.1",
		"127.0.0.1",
	}
)

func TestIpTable(t *testing.T) {

	var testIp string = "186.154.213.155"
	var err error

	// gen new db without initializing
	dbCli, err := NewDBClient(context.Background(), utils.EthereumNetwork, loginStr, false)
	require.NoError(t, err)

	// Test 1 -> test table creation
	err = dbCli.InitIpTable()
	require.NoError(t, err)

	// Test 2 -> insert the new IP

	ipInfo, _, _, err := apis.CallIpApi(testIp)
	require.NoError(t, err)

	q, args := dbCli.UpsertIpInfo(ipInfo)
	_, err = dbCli.SingleQuery(q, args...)
	require.NoError(t, err)

	// Test 6* -> check ip records
	exists, isExpired, err := dbCli.CheckIpRecords(ipInfo.IP)
	require.NoError(t, err)
	require.Equal(t, true, exists)
	require.Equal(t, false, isExpired)

	// Test 3 -> read ip info
	readIpInfo, err := dbCli.ReadIpInfo(ipInfo.IP)
	require.NoError(t, err)
	require.Equal(t, ipInfo.Continent, readIpInfo.Continent)
	//require.Equal(t, ipInfo.ExpirationTime, readIpInfo.ExpirationTime) // We asume that the extra nanosecods are not necesary
	require.Equal(t, ipInfo.ContinentCode, readIpInfo.ContinentCode)
	require.Equal(t, ipInfo.Country, readIpInfo.Country)
	require.Equal(t, ipInfo.CountryCode, readIpInfo.CountryCode)
	require.Equal(t, ipInfo.Region, readIpInfo.Region)
	require.Equal(t, ipInfo.RegionName, readIpInfo.RegionName)
	require.Equal(t, ipInfo.City, readIpInfo.City)
	require.Equal(t, ipInfo.Zip, readIpInfo.Zip)
	//require.Equal(t, ipInfo.Lat, readIpInfo.Lat)
	//require.Equal(t, ipInfo.Lon, readIpInfo.Lon)
	require.Equal(t, ipInfo.Isp, readIpInfo.Isp)
	require.Equal(t, ipInfo.Org, readIpInfo.Org)
	require.Equal(t, ipInfo.As, readIpInfo.As)
	require.Equal(t, ipInfo.AsName, readIpInfo.AsName)
	require.Equal(t, ipInfo.Mobile, readIpInfo.Mobile)
	require.Equal(t, ipInfo.Proxy, readIpInfo.Proxy)
	require.Equal(t, ipInfo.Hosting, readIpInfo.Hosting)

	log.Info("expTime ", readIpInfo.ExpirationTime)
	log.Info("force expire ", readIpInfo.ExpirationTime.AddDate(0, -1, -1))
	// Test 4 -> update peer
	readIpInfo.ExpirationTime = readIpInfo.ExpirationTime.AddDate(0, -1, -1) // rest 31 days
	q, args = dbCli.UpsertIpInfo(readIpInfo)
	_, err = dbCli.SingleQuery(q, args...)
	require.NoError(t, err)

	// Test 5 -> get expired ip info
	expired, err := dbCli.GetExpiredIpInfo()
	require.NoError(t, err)
	require.Equal(t, 1, len(expired))

	// Test 6* -> check ip records
	exists, isExpired, err = dbCli.CheckIpRecords(readIpInfo.IP)
	require.NoError(t, err)
	require.Equal(t, true, exists)
	require.Equal(t, true, isExpired)

}

// test the requestCache individually
func TestApiCall(t *testing.T) {

	// request the 50 Public IPs
	for i := 0; i < 10; i++ {
		ipInfo, _, _, err := apis.CallIpApi(validTestIps[i])
		require.NoError(t, err)
		require.Equal(t, validTestIps[i], ipInfo.IP)
	}
}

// test the requestCache individually
func TestIpLocator(t *testing.T) {

	loginStr := "postgresql://test:password@localhost:5432/armiarmadb"

	// create db and only initialize the ip table
	dbCli, err := NewDBClient(context.Background(), utils.EthereumNetwork, loginStr, false)
	require.NoError(t, err)
	err = dbCli.InitIpTable()
	require.NoError(t, err)

	ipLocator := apis.NewIpLocator(context.Background(), dbCli)

	ipLocator.Run()
	defer ipLocator.Close()

	// request the 202 Public IPs
	for _, value := range validTestIps {
		ipLocator.LocateIP(value)
	}

	// request the 2 Private IPs
	for _, value := range privTestIps {
		ipLocator.LocateIP(value)
	}

	// request the 202 Public IPs again (they should be in Cache)
	for _, value := range validTestIps {
		ipLocator.LocateIP(value)
	}
}
