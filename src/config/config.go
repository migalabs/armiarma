/**

This package contains the structs and methods needed to import configuration
parameters from a config file.
It also contains default configuration in case some of the parameters were wrong


*/

package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/migalabs/armiarma/src/base"
	log "github.com/sirupsen/logrus"
)

// define constant variables
var (
	PkgName string = "CONFIG"
)

type ConfigData struct {
	localLogger   log.FieldLogger
	IP            string   `json:"IP"`
	TcpPort       int      `json:"TcpPort"`
	UdpPort       int      `json:"UdpPort"`
	UserAgent     string   `json:"UserAgent"`
	TopicArray    []string `json:"TopicArray"`
	Network       string   `json:"Network"`
	Eth2Endpoint  string   `json:"Eth2Endpoint`
	ForkDigest    string   `json:"ForkDigest"`
	LogLevel      string   `json:"LogLevel"`
	PrivateKey    string   `json:"PrivateKey"`
	BootNodesFile string   `json:"BootNodesFile"`
	OutputPath    string   `json:"OutputPath"`
	DBType        string   `json:"DBType"`
}

// NewEmptyConfig
// * This method will create a ConfigData empty object
// @return A ConfigData object
func NewEmptyConfig() ConfigData {
	// generate an empty configuration (will call later on the default info)
	config := ConfigData{
		localLogger: base.CreateLogger(defaultConfigLoggerOpts(base.LogOpts{})),
	}

	return config
}

// NewConfigFromArgs
// * This method will create a ConfigData from the given args flags
// @return A ConfigData object
// @return whether the help was requested or not
func NewConfigFromArgs() (ConfigData, bool) {
	// Parse the arguments looking for the config-file
	var help bool
	var configFile string
	flag.BoolVar(&help, "help", false, "display available commands and flags to run the armiarma crawler")
	flag.StringVar(&configFile, "config-file", "./config-files/config.json", "config-file with all the available configurations. Find an example at ./config-files/config.json")
	flag.Parse()

	// generate an empty configuration (will call later on the default info)
	config := ConfigData{
		localLogger: base.CreateLogger(defaultConfigLoggerOpts(base.LogOpts{})),
	}

	// check if the help was requested
	if help {
		return config, help
	}

	// check if a file was given
	if configFile != "" {
		config.ReadFromJSON(configFile)
	}

	return config, help
}

// ReadFromJSON
// *This method will parse a Configuration file and retrieve the data
// * into the current ConfigData object
// @param input_file where to read configuration from
func (c *ConfigData) ReadFromJSON(input_file string) {
	c.localLogger.Infof("Reading configuration from file: %s", input_file)

	if _, err := os.Stat(input_file); os.IsNotExist(err) {
		c.localLogger.Warnf("File does not exist or is corrupted")
	} else {

		file, err := ioutil.ReadFile(input_file)
		if err == nil {
			err := json.Unmarshal([]byte(file), c)
			if err != nil {
				c.localLogger.Warnf("Could not Unmarshal Config file content: %s", err)
			}
		} else {
			c.localLogger.Warnf("Could not read Config file: %s", err)
		}
	}
}

// defaultConfigLoggerOpts
// * This method will apply logging options on top of an existing
// * logging object for a ConfigData.
// @param opts the base logging object
// @return the modified logging object adjusted to ConfigData
func defaultConfigLoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PkgName
	input_opts.Level = "info"
	return input_opts
}

// getters and setters

func (c *ConfigData) GetTcpPort() int {
	return c.TcpPort
}
func (c *ConfigData) SetTcpPort(input_port int) {
	c.TcpPort = input_port
}

func (c *ConfigData) GetUdpPort() int {
	return c.UdpPort
}
func (c *ConfigData) SetUdpPort(input_port int) {
	c.UdpPort = input_port
}

func (c *ConfigData) GetIP() string {
	return c.IP
}

func (c *ConfigData) SetIP(input_ip string) {
	c.IP = input_ip
}

func (c *ConfigData) GetUserAgent() string {
	return c.UserAgent
}
func (c *ConfigData) SetUserAgent(input_string string) {
	c.UserAgent = input_string
}

func (c *ConfigData) GetTopicArray() []string {
	return c.TopicArray
}
func (c *ConfigData) SetTopicArray(input_list []string) {
	c.TopicArray = input_list
}
func (c *ConfigData) SetTopicArrayFromString(input_list string) {
	c.TopicArray = strings.Split(input_list, ",")
}

func (c *ConfigData) GetNetwork() string {
	return c.Network
}
func (c *ConfigData) SetNetwork(input_string string) {
	c.Network = input_string
}

func (c *ConfigData) GetEth2Endpoint() string {
	return c.Eth2Endpoint
}
func (c *ConfigData) SetEth2Endpoint(input_string string) {
	c.Eth2Endpoint = input_string
}

func (c *ConfigData) GetForkDigest() string {
	return c.ForkDigest
}
func (c *ConfigData) SetForkDigest(input_string string) {
	c.ForkDigest = input_string
}

func (c *ConfigData) GetLogLevel() string {
	return c.LogLevel
}
func (c *ConfigData) SetLogLevel(input_string string) {
	c.LogLevel = input_string
}

func (c *ConfigData) GetPrivKey() string {
	return c.PrivateKey
}
func (c *ConfigData) SetPrivKey(input_string string) {
	c.PrivateKey = input_string
}

func (c *ConfigData) GetBootNodesFile() string {
	return c.BootNodesFile
}
func (c *ConfigData) SetBootNodesFile(input_string string) {
	c.BootNodesFile = input_string
}

func (c *ConfigData) GetOutputPath() string {
	return c.OutputPath
}
func (c *ConfigData) SetOutputPath(input_string string) {
	c.OutputPath = input_string
}

func (c *ConfigData) GetDBType() string {
	return c.DBType
}
func (c *ConfigData) SetDBType(input_string string) {
	c.DBType = input_string
}
