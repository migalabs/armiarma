package utils

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetIpFromMultiaddress(t *testing.T) {
	MultiAddresses := []string{
		"/ip4/127.0.0.1/tcp/9000",
		"/ip4/0.0.0.0/tcp/9000",
		"/ip4/98.7.3.23/tcp/9000",
	}

	ip, err := GetIpFromMultiAddr(MultiAddresses[0])
	require.Equal(t, err, nil)
	require.Equal(t, ip.String(), "127.0.0.1")

	ip, err = GetIpFromMultiAddr(MultiAddresses[1])
	require.Equal(t, err, nil)
	require.Equal(t, ip.String(), "0.0.0.0")

	ip, err = GetIpFromMultiAddr(MultiAddresses[2])
	require.Equal(t, err, nil)
	require.Equal(t, ip.String(), "98.7.3.23")

}

func Test_IPPublic(t *testing.T) {
	ips := []string{
		"127.0.0.1",
		"0.0.0.0",
		"178.0.0.1",
		"10.2.25.1",
		"98.7.3.23",
	}

	ispublic := IsPublic(net.ParseIP(ips[0]))
	require.Equal(t, IsPublic(net.ParseIP(ips[0])), false)

	ispublic = IsPublic(net.ParseIP(ips[1]))
	require.Equal(t, IsPublic(net.ParseIP(ips[0])), false)

	ispublic = IsPublic(net.ParseIP(ips[2]))
	require.Equal(t, IsPublic(net.ParseIP(ips[0])), false)

	ispublic = IsPublic(net.ParseIP(ips[3]))
	require.Equal(t, IsPublic(net.ParseIP(ips[0])), false)

	ispublic = IsPublic(net.ParseIP(ips[4]))
	require.Equal(t, ispublic, true)

}

func Test_PublicMultiAddrsSelector(t *testing.T) {
	ipv4MultiAddresses_1 := []string{
		"/ip4/127.0.0.1/tcp/9000",
		"/ip4/192.168.35.1/tcp/9000",
		"/ip4/0.0.0.0/tcp/9000",
		"/ip4/172.16.0.1/tcp/9000",
		"/ip4/10.2.25.1/tcp/9000",
		"/ip4/98.7.3.23/tcp/9000",
	}
	ipv4MultiAddresses_2 := []string{}
	ipv4MultiAddresses_3 := []string{
		"/ip4/192.168.35.1/tcp/9000",
		"/ip4/127.0.0.1/tcp/9000",
		"/ip4/0.0.0.0/tcp/9000",
		"/ip4/172.16.0.1/tcp/9000",
		"/ip4/10.2.25.1/tcp/9000",
		"",
	}

	addr, err := GetFullAddress(ipv4MultiAddresses_1)
	require.Equal(t, err, nil)
	require.Equal(t, addr, "/ip4/98.7.3.23/tcp/9000")

	addr, err = GetFullAddress(ipv4MultiAddresses_2)
	require.Equal(t, addr, "/ip4/127.0.0.1/tcp/9000")

	addr, err = GetFullAddress(ipv4MultiAddresses_3)
	require.Equal(t, err, nil)
	require.Equal(t, addr, "/ip4/192.168.35.1/tcp/9000")

}
