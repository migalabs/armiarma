package config

import (
    "fmt"
    "os"
	"strings"

    "github.com/joho/godotenv"
    "github.com/pkg/errors"
	"github.com/migalabs/armiarma/pkg/utils"
)

var (
    // Crawler defaults
    DefaultLogLevel                  string = getEnv("LOG_LEVEL", "info")
    DefaultPrivKey                   string = getEnv("PRIV_KEY", "")
    DefaultIP                        string = getEnv("DEFAULT_IP", "0.0.0.0")
    DefaultMetricsIP                 string = getEnv("METRICS_IP", "0.0.0.0")
    DefaultSSEIP                     string = getEnv("SSE_IP", "0.0.0.0")
    DefaultPort                      int    = 9020
    DefaultMetricsPort               int    = 9080
    DefaultSSEPort                   int    = 9099
    DefaultUserAgent                 string = "Armiarma Crawler"
	DefaultDatabaseType 			 string = "redshift"
    DefaultPSQLEndpoint              string = constructPSQLEndpoint()
    DefaultRedShiftEndpoint          string = constructRedShiftEndpoint()
    DefaultActivePeersBackupInterval string = "12h"
    DefaultPersistConnEvents         bool   = true

    DefaultAttestationBufferSize = 10000

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

func init() {
    if err := godotenv.Load(); err != nil {
        fmt.Println("No .env file found")
    }
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}

func constructPSQLEndpoint() string {
    user := getEnv("DB_USER", "default_user")
    pass := getEnv("DB_PASSWORD", "default_password")
    host := getEnv("DB_HOST", "localhost")
    port := getEnv("DB_PORT", "5432")
    dbname := getEnv("DB_NAME", "armiarmadb")
    return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, dbname)
}

func constructRedShiftEndpoint() string {
    user := getEnv("REDSHIFT_USER", "default_user")
    pass := getEnv("REDSHIFT_PASSWORD", "default_password")
    host := getEnv("REDSHIFT_HOST", "default_host")
    port := getEnv("REDSHIFT_PORT", "5439") // Default Redshift port
    dbname := getEnv("REDSHIFT_DB", "default_db")
    return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require", user, pass, host, port, dbname)
}

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
