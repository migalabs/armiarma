/**

This package contains the structs and methods needed to import configuration
parameters from a config file.
It also contains default configuration in case some of the parameters were wrong


*/

package config

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	gcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/migalabs/armiarma/src/base"
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

type ConfigData struct {
	base base.Base

	IP      string `json:"IP"`
	TcpPort int    `json:"TcpPort"`
	UdpPort int    `json:"UdpPort"`

	UserAgent string `json:"UserAgent"`

	TopicArray []string `json:"TopicArray"`
	Network    string   `json:"Network"`
	ForkDigest string   `json:"ForkDigest"`

	LogLevel   string `json:"LogLevel"`
	PrivateKey string `json:"PrivateKey"`
}

// Will create an empty object
func NewEmptyConfigData(opts base.LogOpts) *ConfigData {
	new_base, err := base.NewBase(
		base.WithLogger(base.LogOpts{
			ModName:   opts.ModName,
			Output:    opts.Output,
			Formatter: opts.Formatter,
			Level:     opts.Level,
		}),
	)
	if err != nil {
		log.Panicf("Could not create base object %s", err)
	}

	return &ConfigData{
		base: *new_base,
	}
}

// Will create an object using default parameters
func NewDefaultConfigData(opts base.LogOpts) *ConfigData {

	new_base, err := base.NewBase(
		base.WithLogger(base.LogOpts{
			ModName:   opts.ModName,
			Output:    opts.Output,
			Formatter: opts.Formatter,
			Level:     opts.Level,
		}),
	)
	if err != nil {
		log.Panicf("Could not create base object %s", err)
	}

	return &ConfigData{
		base:    *new_base,
		IP:      DEFAULT_IP,
		TcpPort: DEFAULT_TCP_PORT,
		UdpPort: DEFAULT_UDP_PORT,

		UserAgent: DEFAULT_USER_AGENT,

		TopicArray: strings.Split(DEFAULT_TOPIC_ARRAY, ","),
		Network:    DEFAULT_NETWORK,
		ForkDigest: DEFAULT_FORK_DIGEST,

		LogLevel:   DEFAULT_LOG_LEVEL,
		PrivateKey: "",
	}
}

// Receives an input file where to read configuration from and imports into
// the current object
func (c *ConfigData) ReadFromJSON(input_file string) {
	c.base.Log.Debugf("Reading configuration from: ", input_file)
	file, _ := ioutil.ReadFile(input_file)

	err := json.Unmarshal([]byte(file), c)

	c.checkEmptyFields()

	if err != nil {
		fmt.Println(err)
	}
}

// iterate over the fields of
func (c *ConfigData) checkEmptyFields() {
	if c.GetIP() == "" {
		c.SetIP(DEFAULT_IP)
		c.base.Log.Debugf("Setting default IP: %s", DEFAULT_IP)
	}

	if c.GetTcpPort() == 0 {
		c.SetTcpPort(DEFAULT_TCP_PORT)
		c.base.Log.Debugf("Setting default TcpPort: %d", DEFAULT_TCP_PORT)
	}

	if c.GetUdpPort() == 0 {
		c.SetUdpPort(DEFAULT_UDP_PORT)
		c.base.Log.Debugf("Setting default UdpPort: %d", DEFAULT_UDP_PORT)
	}

	if c.GetUserAgent() == "" {
		c.SetUserAgent(DEFAULT_USER_AGENT)
		c.base.Log.Debugf("Setting default UserAgent: %s", DEFAULT_USER_AGENT)
	}

	if len(c.GetTopicArray()) == 0 {
		c.SetTopicArrayFromString(DEFAULT_TOPIC_ARRAY)
		c.base.Log.Debugf("Setting default TopicArray: %s", DEFAULT_TOPIC_ARRAY)
	}

	if c.GetNetwork() == "" {
		c.SetNetwork(DEFAULT_NETWORK)
		c.base.Log.Debugf("Setting default Network: %s", DEFAULT_NETWORK)
	}

	if c.GetForkDigest() == "" {
		c.SetForkDigest(DEFAULT_FORK_DIGEST)
		c.base.Log.Debugf("Setting default ForkDigest: %s", DEFAULT_FORK_DIGEST)
	}

	if c.GetPrivKey() == "" {
		c.base.Log.Debugf("Could not read private key from config file")
		c.generate_privKey()

	}

}

func (c *ConfigData) generate_privKey() {

	key, err := ecdsa.GenerateKey(gcrypto.S256(), rand.Reader)

	if err != nil {
		c.base.Log.Panicf("failed to generate key: %v", err)
	}

	secpKey := (*crypto.Secp256k1PrivateKey)(key)

	keyBytes, err := secpKey.Raw()

	if err != nil {
		c.base.Log.Panicf("failed to serialize key: %v", err)
	}

	c.SetPrivKey(hex.EncodeToString(keyBytes))
	c.base.Log.Debugf("Generated Key!: ", hex.EncodeToString(keyBytes))
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
