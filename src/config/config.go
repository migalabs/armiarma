package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

const DEFAULT_IP string = "0.0.0.0"
const DEFAULT_TCP_PORT int = 9000
const DEFAULT_UDP_PORT int = 9001
const DEFAULT_TOPIC_ARRAY string = "hola,adios" // parse and split by comma to obtain the array
const DEFAULT_NETWORK string = "mainnet"
const DEFAULT_FORK_DIGEST string = "0xffff"
const DEFAULT_USER_AGENT string = "bsc_crawler"
const DEFAULT_LOG_LEVEL string = "debug"

type ConfigData struct {
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

func NewEmptyConfigData() *ConfigData {
	return &ConfigData{}
}

func NewDefaultConfigData() *ConfigData {
	return &ConfigData{
		IP:      DEFAULT_IP,
		TcpPort: DEFAULT_TCP_PORT,
		UdpPort: DEFAULT_UDP_PORT,

		UserAgent: DEFAULT_USER_AGENT,

		TopicArray: strings.Split(DEFAULT_TOPIC_ARRAY, ","),
		Network:    DEFAULT_NETWORK,
		ForkDigest: DEFAULT_FORK_DIGEST,

		LogLevel: DEFAULT_LOG_LEVEL,
	}
}

func (c *ConfigData) ReadFromJSON(input_file string) error {

	file, _ := ioutil.ReadFile(input_file)

	err := json.Unmarshal([]byte(file), c)

	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
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
