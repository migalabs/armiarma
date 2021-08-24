package utils

import (
	"strings"
)

// Client utils
// Main function that will analyze the client type and verion out of the Peer UserAgent
// return the Client Type and it's verison (if determined)
func FilterClientType(userAgent string) (string, string) {
	var client string
	var version string
	// get the UserAgent in lowercases
	userAgent = strings.ToLower(userAgent)
	// check the client type
	if strings.Contains(userAgent, "lighthouse") { // the client is from Lighthouse
		// Lighthouse UserAgent Example: "Lighthouse/v1.0.3-65dcdc3/x86_64-linux"
		client = "Lighthouse"
		// Extract version
		s := strings.Split(userAgent, "/")
		aux := strings.Split(s[1], "-")
		version = aux[0]
	} else if strings.Contains(userAgent, "prysm") { // the client is from Prysm
		// Prysm UserAgent Example: "Prysm/v1.1.0/9b367b36fc12ecf565ad649209aa2b5bba8c7797"
		client = "Prysm"
		// Extract version
		s := strings.Split(userAgent, "/")
		version = s[1]
	} else if strings.Contains(userAgent, "teku") { // the client is from Prysm
		// Prysm UserAgent Example: "Prysm/v1.1.0/9b367b36fc12ecf565ad649209aa2b5bba8c7797"
		client = "Teku"
		// Extract version
		s := strings.Split(userAgent, "/")
		aux := strings.Split(s[2], "+")
		version = aux[0]
	} else if strings.Contains(userAgent, "nimbus") {
		client = "Nimbus"
		version = "Unknown"
	} else if strings.Contains(userAgent, "js-libp2p") {
		client = "Lodestar"
		s := strings.Split(userAgent, "/")
		version = s[1]
	} else if strings.Contains(userAgent, "unknown") {
		client = "Unknown"
		version = "Unknown"
	} else {
		client = "Unknown"
		version = "Unknown"
	}
	return client, version
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
