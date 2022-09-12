package utils

import (
	"strings"
)

// Gets the client and version for a given userAgent.
// TODO: Perhaps use some regex
func FilterClientType(userAgent string) (string, string) {
	// Examples:
	// Teku: teku/teku/v21.8.2/linux-x86_64/corretto-java-16
	// Teku: teku/teku/v21.7.0+9-g77b4b9e/linux-x86_64/-ubuntu-openjdk64bitservervm-java-11
	// Prysm: Prysm/v1.4.3/8bca66ac6408a03af52d65541f58384007ed50ef
	// Prysm: Prysm/v1.3.8-hotfix+6c0942/6c09424feb3141b96016bed817d7ade1cd75deb7
	// Lighthouse: Lighthouse/v1.5.1-b0ac346/x86_64-linux
	// Nimbus: nimbus
	// go-ipfs: go-ipfs/0.8.0/48f94e2
	// hydra-boost: hydra-booster/0.7.4
	// storm: storm
	// lotus: lotus-1.13.0+mainnet+git.7a55e8e8
	userAgentLower := strings.ToLower(userAgent)
	fields := strings.Split(userAgentLower, "/")
	// Eth2
	if strings.Contains(userAgentLower, "lighthouse") {
		return "Lighthouse", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "prysm") {
		return "Prysm", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "teku") {
		return "Teku", cleanVersion(getVersionIfAny(fields, 2))
	} else if strings.Contains(userAgentLower, "nimbus") {
		return "Nimbus", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "js-libp2p") || strings.Contains(userAgentLower, "lodestar") {
		return "Lodestar", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "rust-libp2p") || strings.Contains(userAgentLower, "grandine") {
		return "Grandine", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "eth2-crawler") {
		return "NodeWatch", ""
	} else if strings.Contains(userAgentLower, "bsc") || strings.Contains(userAgentLower, "armiarma") {
		return "BSC-Crawler", ""
	} else if strings.Contains(userAgentLower, "go-ipfs") { // IPFS
		return "go-ipgs", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "hydra") { // IPFS
		return "hydra-boost", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "storm") { // IPFS
		return "storm", cleanVersion(getVersionIfAny(fields, 1))
	} else if strings.Contains(userAgentLower, "lotus") { // IPFS
		// TODO: wont work, needs to be fixed to get the real Version
		return "lotus", cleanVersionAux(getVersionIfAny(fields, 0))
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

func cleanVersionAux(version string) string {
	cleaned := strings.Split(version, "+")[0]
	cleaned = strings.Split(cleaned, "-")[1]
	return cleaned
}
