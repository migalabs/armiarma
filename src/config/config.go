/**

This package contains the structs and methods needed to import configuration
parameters from a config file.
It also contains default configuration in case some of the parameters were wrong


*/

package config

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/utils"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_IP string = "0.0.0.0"
const DEFAULT_TCP_PORT int = 9000
const DEFAULT_UDP_PORT int = 9001
const DEFAULT_TOPIC_ARRAY string = "hola,adios" // parse and split by comma to obtain the array
const DEFAULT_NETWORK string = "mainnet"
const DEFAULT_FORK_DIGEST string = "0xffff"
const DEFAULT_USER_AGENT string = "bsc_crawler"
const DEFAULT_LOG_LEVEL string = "debug"

const DEFAULT_EMPTY_INT int = 0
const DEFAULT_EMPTY_STRING string = ""
const DEFAULT_DB_PATH string = ""
const DEFAULT_DB_TYPE string = ""

const PKG_NAME string = "Config"

type ConfigData struct {
	localLogger   log.FieldLogger
	IP            string   `json:"IP"`
	TcpPort       int      `json:"TcpPort"`
	UdpPort       int      `json:"UdpPort"`
	UserAgent     string   `json:"UserAgent"`
	TopicArray    []string `json:"TopicArray"`
	Network       string   `json:"Network"`
	ForkDigest    string   `json:"ForkDigest"`
	LogLevel      string   `json:"LogLevel"`
	PrivateKey    string   `json:"PrivateKey"`
	BootNodesFile string   `json:"BootNodesFile"`
	DBPath        string   `json:"DBPath"`
	DBType        string   `json:"DBType"`
}

// NewEmptyConfig
// * This method will create a ConfigData empty object
// @param opts The parameter including the logging options
// @return A ConfigData object
func NewEmptyConfigData(opts base.LogOpts) *ConfigData {
	opts = defaultConfigLoggerOpts(opts)
	return &ConfigData{
		localLogger: base.CreateLogger(opts),
	}
}

// NewDefaultConfigData
// * This method will create a ConfigData object using Default parameters
// * hardoded in the code
// @param opts The parameter including the logging options
// @return A ConfigData object
func NewDefaultConfigData(opts base.LogOpts) *ConfigData {

	opts = defaultConfigLoggerOpts(opts)

	return &ConfigData{
		localLogger: base.CreateLogger(opts),
		IP:          DEFAULT_IP,
		TcpPort:     DEFAULT_TCP_PORT,
		UdpPort:     DEFAULT_UDP_PORT,

		UserAgent: DEFAULT_USER_AGENT,

		TopicArray: strings.Split(DEFAULT_TOPIC_ARRAY, ","),
		Network:    DEFAULT_NETWORK,
		ForkDigest: DEFAULT_FORK_DIGEST,

		LogLevel:   DEFAULT_LOG_LEVEL,
		PrivateKey: "",
		DBPath:     DEFAULT_DB_PATH,
		DBType:     DEFAULT_DB_TYPE,
	}
}

// ReadFromJSON
// *This method will parse a Configuration file and retrieve the data
// * into the current ConfigData object
// @param input_file where to read configuration from
func (c *ConfigData) ReadFromJSON(input_file string) {
	c.localLogger.Infof("Reading configuration from: ", input_file)

	if _, err := os.Stat(input_file); os.IsNotExist(err) {
		c.localLogger.Debugf("Could not read file")
	} else {

		file, err := ioutil.ReadFile(input_file)
		if err == nil {
			err := json.Unmarshal([]byte(file), c)
			if err != nil {
				c.localLogger.Debugf("Could not Unmarshal Config file: %s", err)
			}
		} else {
			c.localLogger.Debugf("Could not read Config file: %s", err)
		}
	}

	c.checkEmptyFields() // this function will check any field that was not read and apply the default

}

// checkEmptyFields
// * This method will iterate over all fields in the current
// * ConfigData object and check if any is empty.
// *If so, apply the default
func (c *ConfigData) checkEmptyFields() {
	if c.checkValidIP() {
		c.SetIP(DEFAULT_IP)
		c.localLogger.Debugf("Setting default IP: %s", DEFAULT_IP)
	}

	if c.checkValidTcpPort() {
		c.SetTcpPort(DEFAULT_TCP_PORT)
		c.localLogger.Debugf("Setting default TcpPort: %d", DEFAULT_TCP_PORT)
	}

	if c.checkValidUdpPort() {
		c.SetUdpPort(DEFAULT_UDP_PORT)
		c.localLogger.Debugf("Setting default UdpPort: %d", DEFAULT_UDP_PORT)
	}

	if c.GetUserAgent() == "" {
		c.SetUserAgent(DEFAULT_USER_AGENT)
		c.localLogger.Debugf("Setting default UserAgent: %s", DEFAULT_USER_AGENT)
	}

	if len(c.GetTopicArray()) == 0 {
		c.SetTopicArrayFromString(DEFAULT_TOPIC_ARRAY)
		c.localLogger.Debugf("Setting default TopicArray: %s", DEFAULT_TOPIC_ARRAY)
	}

	if c.GetNetwork() == "" {
		c.SetNetwork(DEFAULT_NETWORK)
		c.localLogger.Debugf("Setting default Network: %s", DEFAULT_NETWORK)
	}

	if c.GetForkDigest() == "" {
		c.SetForkDigest(DEFAULT_FORK_DIGEST)
		c.localLogger.Debugf("Setting default ForkDigest: %s", DEFAULT_FORK_DIGEST)
	}

	if c.GetPrivKey() == "" {
		c.localLogger.Debugf("Could not read private key from config file")
		c.SetPrivKey(utils.Generate_privKey())
	}

	if c.GetBootNodesFile() == "" {
		c.localLogger.Debugf("Could not find bootnodes file")
	}

	if c.GetLogLevel() == "" {
		c.SetLogLevel(DEFAULT_LOG_LEVEL)
		c.localLogger.Debugf("Setting default LogLevel: %s", DEFAULT_LOG_LEVEL)
	}

	if c.GetDBPath() == "" {
		c.SetDBPath(DEFAULT_DB_PATH)
		c.localLogger.Debugf("Setting default DB Path: %s", DEFAULT_DB_PATH)
	}

	if c.GetDBType() == "" {
		c.SetDBType(DEFAULT_DB_TYPE)
		c.localLogger.Debugf("Setting default DB Type: %s", DEFAULT_DB_TYPE)
	}

}

// defaultConfigLoggerOpts
// * This method will apply logging options on top of an existing
// * logging object for a ConfigData.
// @param opts the base logging object
// @return the modified logging object adjusted to ConfigData
func defaultConfigLoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PKG_NAME
	input_opts.Level = DEFAULT_LOG_LEVEL

	return input_opts
}

// getters and setters

func (c *ConfigData) GetTcpPort() int {
	return c.TcpPort
}
func (c *ConfigData) SetTcpPort(input_port int) {
	c.TcpPort = input_port
}
func (c *ConfigData) checkValidTcpPort() bool {
	return checkValidPort(c.GetTcpPort())
}

func (c *ConfigData) GetUdpPort() int {
	return c.UdpPort
}
func (c *ConfigData) SetUdpPort(input_port int) {
	c.UdpPort = input_port
}
func (c *ConfigData) checkValidUdpPort() bool {
	return checkValidPort(c.GetUdpPort())
}

func checkValidPort(input_port int) bool {
	// we put greater than min port, as 0 is default when no value was set
	if input_port > utils.MIN_PORT_NUM && input_port <= utils.MAX_PORT_NUM {
		return true
	}
	return false
}

func (c *ConfigData) GetIP() string {
	return c.IP
}

func (c *ConfigData) SetIP(input_ip string) {
	c.IP = input_ip
}

func (c *ConfigData) checkValidIP() bool {
	input_IP := c.GetIP()
	if input_IP != "" && utils.IsIPPublic(net.ParseIP(input_IP)) {
		return true
	}
	return false
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

func (c *ConfigData) GetDBPath() string {
	return c.DBPath
}
func (c *ConfigData) SetDBPath(input_string string) {
	c.DBPath = input_string
}

func (c *ConfigData) GetDBType() string {
	return c.DBPath
}
func (c *ConfigData) SetDBType(input_string string) {
	c.DBType = input_string
}
