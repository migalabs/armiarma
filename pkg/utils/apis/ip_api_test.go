package apis

import (
	"context"
	"testing"

	"github.com/migalabs/armiarma/pkg/db/postgresql"
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

// test the requestCache individually
func TestApiCall(t *testing.T) {

	// request the 50 Public IPs
	for i := 0; i < 10; i++ {
		ipInfo, _, err := CallIpApi(validTestIps[i])
		require.NoError(t, err)
		require.Equal(t, validTestIps[i], ipInfo.Query)
	}
}

// test the requestCache individually
func TestIpLocator(t *testing.T) {

	loginStr := "postgresql://test:password@localhost:5432/armiarmadb"

	// create db and only initialize the ip table
	dbCli, err := postgresql.NewDBClient(context.Background(), loginStr, false)
	require.NoError(t, err)
	err = dbCli.InitIpTable()
	require.NoError(t, err)

	ipLocator := NewIpLocator(context.Background(), dbCli)

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
