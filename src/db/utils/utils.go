package utils

import (
	"strings"

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
	} else if strings.Contains(userAgentLower, "BSC-Eth2-Crawler") || strings.Contains(userAgentLower, "BSC-Armiarma") {
		return "BSC-Eth2-Crawler", ""
	} else if userAgentLower == "" {
		return "NotIdentified", ""
	} else {
		log.Debugf("Could not get client from userAgent: %s", userAgent)
		return "Others", ""
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
	err = strings.ToLower(err)
	errorPretty := "Uncertain"
	// filter the error type
	if strings.Contains(err, "connection reset by peer") {
		// The peer that we tried to connect resets/drops the connection
		errorPretty = "connection reset by peer"
	} else if strings.Contains(err, "i/o timeout") {
		// When trying to connect a peer, the timeout waiting for stablishing the connection was triggered
		errorPretty = "i/o timeout"
	} else if strings.Contains(err, "dial to self attempted") {
		// When the host tries to connect to itself
		errorPretty = "dial to self attempted"
	} else if strings.Contains(err, "dial backoff") {
		errorPretty = "dial backoff"
	} else if strings.Contains(err, "connection refused") {
		// The peer that we tried to connect refuses/drops the connection
		errorPretty = "connection refused"
	} else if strings.Contains(err, "context deadline exceeded") {
		// When the crawler is able to stablish the connection, but we are unable to identify it
		// h.Connect() internally calls/identifies the peer which reports the error
		errorPretty = "context deadline exceeded"
	} else if strings.Contains(err, "no route to host") {
		// Unable to find a host in that IP
		errorPretty = "no route to host"
	} else if strings.Contains(err, "network is unreachable") || strings.Contains(err, "unreachable network") {
		errorPretty = "unreachable network"
	} else if strings.Contains(err, "peer id mismatch") {
		// Dialing a peer that does not longer exist
		// Hoever there is a new one with another peerID
		errorPretty = "peer id mismatch"
	} else if strings.Contains(err, "none") {
		errorPretty = "none"
	} else if strings.Contains(err, "error requesting metadata") {
		errorPretty = "metadata error"
	} else {
		// Uncertain (not tracked one)
		log.Errorf("uncertain error: %s", err)
	}
	// TODO: Further encountered errors:
	// 		 - stream reset
	// 		 - no good address

	return errorPretty
}
