package postgresql

import (
	"context"
	"testing"

	"github.com/migalabs/armiarma/pkg/utils/apis"
	"github.com/stretchr/testify/require"
)

func TestIpTable(t *testing.T) {

	var testIp string = "186.154.213.155"
	var err error

	// gen new db without initializing
	dbCli, err := NewDBClient(context.Background(), loginStr, false)
	require.NoError(t, err)

	// Test 1 -> test table creation
	err = dbCli.InitIpTable()
	require.NoError(t, err)

	// Test 2 -> insert the new IP

	ipInfo, _, err := apis.CallIpApi(testIp)
	require.NoError(t, err)

	err = dbCli.InsertNewIP(ipInfo)
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
	err = dbCli.UpdateIP(readIpInfo)
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
