package utils

import (
	"fmt"
	"net"
	"strings"

	"github.com/migalabs/armiarma/src/gossipsub"
	log "github.com/sirupsen/logrus"
)

// Gets the client and version for a given userAgent
// TODO: Perhaps use some regex
func FilterClientType(userAgent string) (string, string) {
	// Examples:
	// Teku: teku/teku/v21.8.2/linux-x86_64/corretto-java-16
	// Teku: teku/teku/v21.7.0+9-g77b4b9e/linux-x86_64/-ubuntu-openjdk64bitservervm-java-11
	// Prysm: Prysm/v1.4.3/8bca66ac6408a03af52d65541f58384007ed50ef
	// Prysm: Prysm/v1.3.8-hotfix+6c0942/6c09424feb3141b96016bed817d7ade1cd75deb7
	// Lighthouse: Lighthouse/v1.5.1-b0ac346/x86_64-linux
	// Nimbus: nimbus
	userAgentLower := strings.ToLower(userAgent)
	fields := strings.Split(userAgentLower, "/")
	if strings.Contains(userAgentLower, "lighthouse") {
		return "Lighthouse", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "prysm") {
		return "Prysm", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "teku") {
		return "Teku", cleanVersion(getVersionIfAny(fields, 2))
	} else if strings.Contains(userAgentLower, "nimbus") {
		return "Nimbus", "Unknown"
	} else if strings.Contains(userAgentLower, "js-libp2p") {
		return "Lodestar", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "rust-libp2p") {
		return "Grandine", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "eth2-crawler") {
		return "NodeWatch", ""
	} else if strings.Contains(userAgentLower, "armiarma-crawler") {
		return "BSC-Armiarma", ""
	} else {
		log.Debugf("Could not get client from userAgent: %s", userAgent)
		return "Unknown", "Unknown"
	}
}

func getVersionIfAny(fields []string, index int) string {
	if index > (len(fields) - 1) {
		return "Unknown"
	} else {
		return fields[index]
	}
}

func cleanVersion(version string) string {
	cleaned := strings.Split(version, "+")[0]
	cleaned = strings.Split(cleaned, "-")[0]
	return cleaned
}

// funtion that formats the error into a Pretty understandable (standard) way
// also important to cohesionate the extra-metrics output csv
func FilterError(err string) string {
	errorPretty := "Uncertain"
	// filter the error type
	if strings.Contains(err, "connection reset by peer") {
		errorPretty = "Connection reset by peer"
	} else if strings.Contains(err, "i/o timeout") || strings.Contains(err, "context deadline exceeded") {
		errorPretty = "i/o timeout"
	} else if strings.Contains(err, "dial to self attempted") {
		errorPretty = "dial to self attempted"
	} else if strings.Contains(err, "dial backoff") {
		errorPretty = "dial backoff"
	} else if strings.Contains(err, "connection refused") {
		errorPretty = "connection refused"
	} else if strings.Contains(err, "no route to host") {
		errorPretty = "no route to host"
	} else if strings.Contains(err, "network is unreachable") {
		errorPretty = "unreachable network"
	} else if strings.Contains(err, "peer id mismatch") {
		errorPretty = "peer id mismatch"
	} else {
		log.Errorf("uncertain error: %s", err)
	}
	// TODO: Further encountered errors:
	// 		 - stream reset
	// 		 - no good address

	return errorPretty
}

func ShortToFullTopicName(topicName string) string {
	switch topicName {
	case "BeaconBlock":
		return gossipsub.BeaconBlock
	case "BeaconAggregateProof":
		return gossipsub.BeaconAggregateProof
	case "VoluntaryExit":
		return gossipsub.VoluntaryExit
	case "ProposerSlashing":
		return gossipsub.ProposerSlashing
	case "AttesterSlashing":
		return gossipsub.AttesterSlashing
	default:
		return ""
	}
}

func GetIPfromMultiaddress(multiaddr string) (ip string, err error) {
	s := strings.Split(multiaddr, "/")
	if len(s) < 3 {
		return ip, fmt.Errorf("multiaddress doesn't include an IP: %s", multiaddr)
	}
	return s[2], nil
}

// IP public filtering
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
