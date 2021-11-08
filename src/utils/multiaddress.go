package utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/libp2p/go-libp2p-core/crypto"
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

// checkvalidIP
// * This method checks whether the IP can be parsed or not
func CheckValidIP(input_IP string) bool {
	parse_IP := net.ParseIP(input_IP)
	if parse_IP != nil {
		return true
	}
	return false
}

func ParsePubkey(v string) (*ecdsa.PublicKey, error) {
	if strings.HasPrefix(v, "0x") {
		v = v[2:]
	}
	pubKeyBytes, err := hex.DecodeString(v)
	if err != nil {
		return nil, fmt.Errorf("cannot parse public key, expected hex string: %v", err)
	}
	var pub crypto.PubKey
	pub, err = crypto.UnmarshalSecp256k1PublicKey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse public key, invalid public key (Secp256k1): %v", err)
	}
	return (*ecdsa.PublicKey)((pub).(*crypto.Secp256k1PublicKey)), nil
}
