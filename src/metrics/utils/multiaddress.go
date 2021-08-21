package utils

import (
	"fmt"
	"net"
	"strings"

	"github.com/pkg/errors"
)

// Get the Real Ip Address from the multi Address list
// TODO: -- IP could be returned after filtering it
func GetFullAddress(multiAddrs []string) (string, error) {
	// TODO: It's pretty unlikely that we don't have the multiaddress of a discovered peer
	//       however, should we leave it empty if we don't have it? All the fields where filled
	//       before for the post analysis in the Python Script, does it still make sense?
	var address string = "/ip4/127.0.0.1/tcp/9000"
	if len(multiAddrs) > 0 {
		fmt.Println(multiAddrs)
		address = multiAddrs[0]
		// TODO: Missing net.IP from multiaddress
		for _, addrs := range multiAddrs {
			// parse IP from Addrs
			ip, err := GetIpFromMultiAddr(addrs)
			if err != nil {
				continue
			}
			if IsPublic(ip) {
				address = addrs
				break
			} else {
				continue
			}
		}
		return address, nil
	}
	return address, errors.New("Given set of multiAddresses is empty")
}

// Extract IP from Multiaddress
func GetIpFromMultiAddr(addrs string) (net.IP, error) {
	if len(addrs) > 0 {
		sIP := strings.Split(addrs, "/")[2]
		if len(sIP) > 0 {
			return net.ParseIP(sIP), nil
		}
		return net.ParseIP(sIP), errors.New("Error, the given multiAddrs is incomplete")
	}
	return net.ParseIP(""), errors.New("Error, the given multiAddrs is empty")
}

var PrivateIPNetworks = []net.IPNet{
	net.IPNet{
		IP:   net.ParseIP("10.0.0.0"),
		Mask: net.CIDRMask(8, 32),
	},
	net.IPNet{
		IP:   net.ParseIP("172.16.0.0"),
		Mask: net.CIDRMask(12, 32),
	},
	net.IPNet{
		IP:   net.ParseIP("192.168.0.0"),
		Mask: net.CIDRMask(16, 32),
	},
}

func IsPublic(ip net.IP) bool {
	for _, ipNet := range PrivateIPNetworks {
		if ipNet.Contains(ip) || ip.IsLoopback() || ip.IsUnspecified() {
			return false
		}
	}
	return true
}
