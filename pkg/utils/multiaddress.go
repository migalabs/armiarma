package utils

import (
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/libp2p/go-libp2p/core/peer"
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

func UnmarshalMaddr(inputAddr string) (ma.Multiaddr, error) {
	new_ma, err := ma.NewMultiaddr(inputAddr)

	if err != nil {
		return nil, err
	}

	return new_ma, nil
}

func IsIPPublic(ip net.IP) bool {
	for _, ipNet := range PrivateIPNetworks {
		if ipNet.Contains(ip) || ip.IsLoopback() || ip.IsUnspecified() {
			return false
		}
	}
	return true
}

func CompAddrInfo(pid string, maddrs []ma.Multiaddr) (peer.AddrInfo, error) {
	peerid, err := peer.Decode(pid)
	if err != nil {
		return peer.AddrInfo{}, errors.Wrap(err, "unable to compose addr info related to peer")
	}
	addrinfo := peer.AddrInfo{
		ID:    peerid,
		Addrs: make([]ma.Multiaddr, 0),
	}
	addrinfo.Addrs = append(addrinfo.Addrs, maddrs...)
	return addrinfo, nil
}

func ExtractIPFromMAddr(maddr ma.Multiaddr) net.IP {
	// check if MAddrs is empty
	if maddr == nil {
		return nil
	}
	// remember that the first position is "", as for having an initial /
	// /ipX/<ip>/<transport_protocol>/<port>/p2p/<peerID>
	spltAddr := strings.Split(maddr.String(), MADDR_SEPARATOR)
	if len(spltAddr) < 3 {
		return nil // finish returning nil
	}

	ip := spltAddr[2] // the IP is in the third position

	return net.ParseIP(ip)
}

func GetPortFromMaddrs(maddr ma.Multiaddr) int {
	// check if MAddrs is empty
	if maddr == nil {
		return -1
	}
	// remember that the first position is "", as for having an initial /
	// /ipX/<ip>/<transport_protocol>/<port>/p2p/<peerID>
	spltAddr := strings.Split(maddr.String(), MADDR_SEPARATOR)
	if len(spltAddr) < 5 {
		return -1 // finish returning nil
	}

	portStr := spltAddr[4] // the IP is in the third position
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return -1
	}
	return port
}

// checkvalidIP
// * This method checks whether the IP can be parsed or not
func CheckValidIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}

func GetPublicAddrsFromAddrArray(mAddrs []ma.Multiaddr) ma.Multiaddr {
	// loop to check if which is the public ip
	var finalAddr ma.Multiaddr
	for _, addr := range mAddrs {
		ip := ExtractIPFromMAddr(addr)
		if IsIPPublic(ip) {
			finalAddr = addr
			break
		}
	}
	return finalAddr
}
