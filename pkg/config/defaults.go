package config

import (
	"os"
	"strings"

	"github.com/migalabs/armiarma/pkg/utils"

	"github.com/pkg/errors"
)

var (
	// Crawler
	DefaultLogLevel                  string = "info"
	DefaultPrivKey                   string = ""
	DefaultIP                        string = "0.0.0.0"
	DefaultMetricsIP                 string = "0.0.0.0"
	DefaultPort                      int    = 9020
	DefaultMetricsPort               int    = 9080
	DefaultUserAgent                 string = "Armiarma Crawler"
	DefaultPSQLEndpoint              string = "postgres://user:password@localhost:5432/armiarmadb"
	DefaultActivePeersBackupInterval string = "12h"

	Ipfsprotocols = []string{
		"/ipfs/kad/1.0.0",
		"/ipfs/kad/2.0.0",
	}
	Filecoinprotocols = []string{
		"/fil/kad/testnetnet/kad/1.0.0",
	}

	// Control
	MinPort           int      = 0
	MaxPort           int      = 65000
	PossibleLogLevels []string = []string{"trace", "debug", "info", "warn", "error"}
)

func checkValidLogLevel(logLevel string) bool {
	for _, availLevel := range PossibleLogLevels {
		if strings.ToLower(availLevel) == strings.ToLower(logLevel) {
			return true
		}
	}
	return false
}

func checkValidPort(inputPort int) bool {
	// we put greater than min port, as 0 is default when no value was set
	if inputPort > MinPort && inputPort <= MaxPort {
		return true
	}
	return false
}

func validateOrCreatePeerstore(outputPath string) error {
	// Check if the folder already exists, or generate one
	if !utils.CheckFileExists(outputPath) {
		// folder does not exist, generate a new one
		err := os.Mkdir(outputPath, 0755)
		if err != nil {
			return errors.Wrap(err, "unable to create folder for local peertore "+outputPath)
		}
	}
	return nil
}
