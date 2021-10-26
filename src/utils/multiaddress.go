package utils

import (
	"net"
	"strings"

	ma "github.com/multiformats/go-multiaddr"
)

const MADDR_SEPARATOR string = "/"

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

func IsIPPublic(ip net.IP) bool {
	for _, ipNet := range PrivateIPNetworks {
		if ipNet.Contains(ip) || ip.IsLoopback() || ip.IsUnspecified() {
			return false
		}
	}
	return true
}

func ExtractIPFromMAddr(input_addr ma.Multiaddr) net.IP {
	string_addr := input_addr.String() // extract in string
	// remember that the first position is "", as for having an initial /
	// /ipX/<ip>/<transport_protocol>/<port>/p2p/<peerID>
	string_addr_splitted := strings.Split(string_addr, MADDR_SEPARATOR)
	if len(string_addr_splitted) < 3 {
		return nil // finish returning nil
	}

	extracted_ip := string_addr_splitted[2] // the IP is in the third position

	return net.ParseIP(extracted_ip)
}
