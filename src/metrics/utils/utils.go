package utils

import (
	"strings"
	pgossip "github.com/protolambda/rumor/p2p/gossip"
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
	} else {
		log.Warn("Could not get client from userAgent: ", userAgent)
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
	} else if strings.Contains(err, "i/o timeout") {
		errorPretty = "i/o timeout"
	} else if strings.Contains(err, "dial to self attempted") {
		errorPretty = "dial to self attempted"
	} else if strings.Contains(err, "dial backoff") {
		errorPretty = "dial backoff"
	}

	return errorPretty
}

func ShortToFullTopicName(topicName string) string {
	switch topicName {
	case "BeaconBlock":
		return pgossip.BeaconBlock
	case "BeaconAggregateProof":
		return pgossip.BeaconAggregateProof
	case "VoluntaryExit":
		return pgossip.VoluntaryExit
	case "ProposerSlashing":
		return pgossip.ProposerSlashing
	case "AttesterSlashing":
		return pgossip.AttesterSlashing
	default:
		return ""
	}
}

// Get the Real Ip Address from the multi Address list
// TODO: Implement the Private IP filter in a better way
func GetFullAddress(multiAddrs []string) string {
	var address string
	if len(multiAddrs) > 0 {
		for _, element := range multiAddrs {
			if strings.Contains(element, "/ip4/192.168.") || strings.Contains(element, "/ip4/127.0.") || strings.Contains(element, "/ip6/") || strings.Contains(element, "/ip4/172.") || strings.Contains(element, "0.0.0.0") {
				continue
			} else {
				address = element
				break
			}
		}
	} else {
		address = "/ip4/127.0.0.1/tcp/9000"
	}
	return address
}
