package postgresql

import (
	"context"
	"testing"

	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/migalabs/armiarma/pkg/utils/apis"
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

	ipInfo, _, err := apis.CallIpApi(testIp)
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
		ipInfo, _, err := apis.CallIpApi(validTestIps[i])
		require.NoError(t, err)
		require.Equal(t, validTestIps[i], ipInfo.Query)
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

	// request the 50 Public IPs
	for _, value := range validTestIps {
		ipLocator.LocateIP(value)
	}

	// request the 2 Private IPs
	for _, value := range privTestIps {
		ipLocator.LocateIP(value)
	}

	// request the 50 Public IPs again (they should be in Cache)
	for _, value := range validTestIps {
		ipLocator.LocateIP(value)
	}
}
