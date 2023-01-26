/**

This package contains the structs and methods needed to import configuration
parameters from a config file.
It also contains default configuration in case some of the parameters were wrong


*/

package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

// define constant variables
var (
	ModuleName string = "CONFIG"
)

type ConfigData struct {
	IP            string   `json:"IP"`
	TcpPort       int      `json:"TcpPort"`
	UdpPort       int      `json:"UdpPort"`
	UserAgent     string   `json:"UserAgent"`
	TopicArray    []string `json:"TopicArray"`
	Network       string   `json:"Network"`
	Eth2Endpoint  string   `json:"Eth2Endpoint"`
	DBEndpoint    string   `json:"DBEndpoint"`
	ForkDigest    string   `json:"ForkDigest"`
	LogLevel      string   `json:"LogLevel"`
	PrivateKey    string   `json:"PrivateKey"`
	BootNodesFile string   `json:"BootNodesFile"`
	OutputPath    string   `json:"OutputPath"`
}

// NewConfigFromArgs
// * This method will create a ConfigData from the given args flags
// @return A ConfigData object
// @return whether the help was requested or not
func NewConfigFromFile(configFile string) ConfigData {
	// generate an empty configuration
	config := ConfigData{}

	// check if a file was given
	if configFile != "" {
		log.Debug("config file was provided")
		config.ReadFromJSON(configFile)
	}

	return config
}

// ReadFromJSON
// *This method will parse a Configuration file and retrieve the data
// * into the current ConfigData object
// @param input_file where to read configuration from
func (c *ConfigData) ReadFromJSON(input_file string) {
	log.Infof("Reading configuration from file: %s", input_file)

	if _, err := os.Stat(input_file); os.IsNotExist(err) {
		log.Warnf("File does not exist or is corrupted")
	} else {

		file, err := ioutil.ReadFile(input_file)
		if err == nil {
			err := json.Unmarshal([]byte(file), c)
			if err != nil {
				log.Warnf("Could not Unmarshal Config file content: %s", err)
			}
		} else {
			log.Warnf("Could not read Config file: %s", err)
		}
	}
}
