package utils

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

type NetworkType string
type ClientName string
type ClientOS string
type ClientArch string

const (
	// Libp2p Available Networks
	EthereumNetwork NetworkType = "Ethereum CL"
	IpfsNetwork     NetworkType = "IPFS"
	FilecoinNetwork NetworkType = "Filecoin"

	// Ethereum Consensus-Layer Clients
	Prysm      ClientName = "prysm"
	Lighthouse ClientName = "lighthouse"
	Teku       ClientName = "teku"
	Nimbus     ClientName = "nimbus"
	Lodestar   ClientName = "lodestar"
	Grandine   ClientName = "grandine"
	Cortex     ClientName = "cortze"
	Trinity    ClientName = "trinity"
	Erigon     ClientName = "erigon"

	// IPFS Client
	Kubo         ClientName = "kubo"
	GoIpfs       ClientName = "go-ipfs"
	HydraBooster ClientName = "hydra-booster"
	Storm        ClientName = "storm"
	Ioi          ClientName = "ioi"
	Punchr       ClientName = "punchr"

	// Filecoin
	Lotus ClientName = "lotus"

	// Others
	Others ClientName = "Others"

	// OSes
	Mac     ClientOS = "mac"
	Windows ClientOS = "windows"
	Linux   ClientOS = "linux"

	// Arch
	Arm    ClientArch = "arm"
	X86_64 ClientArch = "x86_64"

	Unknown string = "unknown"
)

// Ethereum CL CLients
var EthCLClients map[ClientName][]string = map[ClientName][]string{
	Prysm:      {"prysm"},
	Lighthouse: {"lighthouse"},
	Teku:       {"teku"},
	Nimbus:     {"nimbus", "nim-libp2p"},
	Lodestar:   {"lodestar", "js-libp2p"},
	Grandine:   {"grandine", "rust-libp2p"},
	Cortex:     {"cortex"},
	Trinity:    {"trinity"},
	Erigon:     {"erigon", "erigon/lightclient"},
}

// IPFS Clients
var IpfsClients map[ClientName][]string = map[ClientName][]string{
	Kubo:         {"kubo"},
	GoIpfs:       {"go-ipfs"},
	HydraBooster: {"hydra-booster"},
	Storm:        {"storm"},
	Ioi:          {"ioi"},
	Punchr:       {"punchr"},
}

// Filecoin Clients
var FilecoinClients map[ClientName][]string = map[ClientName][]string{
	Lotus: {"lotus"},
}

// Valid OS
var ValidOs map[ClientOS][]string = map[ClientOS][]string{
	Mac:     {"macos", "freebsd"},
	Windows: {"win", "windows"},
	Linux:   {"linux", "ubuntu"},
}

// Valid Architectures
var ValidArchs map[ClientArch][]string = map[ClientArch][]string{
	Arm:    {"aarch64", "aarch", "aarch_64"},
	X86_64: {"x86_64"},
}

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

func ParseClientType(network NetworkType, userAgent string) (cliName string, cliVersion string, cliOs string, cliArch string) {

	// split the UserAgent into chunks divided by '/'
	splUserAgent := strings.Split(userAgent, "/")

	// check the which of the clients is present in the userAgent
	switch network {
	case EthereumNetwork:
		// parse client name from Ethereum Valid Clients
		client := ClientNameParser(EthCLClients, splUserAgent[0])

		// stract the version from the user
		var version string
		switch client {
		case Prysm, Lighthouse, Lodestar, Grandine, Nimbus, Cortex, Trinity:
			version = cleanVersion(getVersionIfAny(splUserAgent, 1))
		case Teku:
			// teku
			version = cleanVersion(getVersionIfAny(splUserAgent, 2))

		default:
			log.Errorf("unable to determine client name for UserAgent %s", userAgent)
			version = Unknown
		}

		cliName = string(client)
		cliVersion = version

	case IpfsNetwork:
		// parse client name from Ethereum Valid Clients
		client := ClientNameParser(IpfsClients, splUserAgent[0])

		// stract the version from the user
		var version string
		switch client {
		case GoIpfs, Kubo, Ioi, Storm, HydraBooster:
			version = cleanVersion(getVersionIfAny(splUserAgent, 1))
		default:
			log.Errorf("unable to determine client name for UserAgent %s", userAgent)
			version = Unknown
		}

		cliName = string(client)
		cliVersion = version

	case FilecoinNetwork:
		// parse client name from Ethereum Valid Clients
		client := ClientNameParser(FilecoinClients, splUserAgent[0])

		// stract the version from the user
		var version string
		switch client {
		case Lotus:
			version = cleanVersion(cleanVersionLotus(splUserAgent[0]))
		default:
			log.Errorf("unable to determine client name for UserAgent %s", userAgent)
			version = Unknown
		}

		cliName = string(client)
		cliVersion = version

	default:
		log.Error("unable to retrieve the user_agent from network", network)
	}

	os := ClientOSParser(ValidOs, userAgent)
	arch := ClientArchParser(ValidArchs, userAgent)

	cliOs = string(os)
	cliArch = string(arch)

	return
}

func ClientNameParser(validNames map[ClientName][]string, parsingName string) ClientName {
	defaultName := ClientName(Unknown)

	// iter over the possibilities for the ethereum consensus layer
	for cName, subCliNames := range validNames {
		// iter through sub-cli names (e.g. lodestar and js-libp2p)
		for _, subValidCli := range subCliNames {
			if strContainsLowerCaps(string(parsingName), string(subValidCli)) {
				return cName
			}
		}
	}
	return defaultName
}

func ClientOSParser(validNames map[ClientOS][]string, parsingName string) ClientOS {
	defaultName := ClientOS(Unknown)

	// iter over the possibilities for the CPU architecture
	for os, subOS := range validNames {
		// iter through sub-cli names (e.g. windows and linux)
		for _, subValidOS := range subOS {
			if strContainsLowerCaps(string(parsingName), string(subValidOS)) {
				return os
			}
		}
	}
	return defaultName
}

func ClientArchParser(validNames map[ClientArch][]string, parsingName string) ClientArch {
	defaultName := ClientArch(Unknown)

	// iter over the possibilities for the ethereum consensus layer
	for arch, subArchNames := range validNames {
		// iter through sub-cli names (e.g. lodestar and js-libp2p)
		for _, subValidOS := range subArchNames {
			if strContainsLowerCaps(string(parsingName), string(subValidOS)) {
				return arch
			}
		}
	}
	return defaultName
}

func strContainsLowerCaps(s string, subStr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(subStr))
}

func getVersionIfAny(fields []string, index int) string {
	if index > (len(fields) - 1) {
		return Unknown
	} else {
		return fields[index]
	}
}

func cleanVersion(version string) string {
	cleaned := strings.Split(version, "+")[0]
	cleaned = strings.Split(cleaned, "-")[0]
	return cleaned
}

func cleanVersionLotus(version string) string {
	cleaned := strings.Split(version, "+")[0]
	cleaned = strings.Split(cleaned, "-")[1]
	return cleaned
}
