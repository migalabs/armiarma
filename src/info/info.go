/**

This package contains all needed structs and functions to
create an object of type InfoData.
InfoData will be considered the main source of parameter information
for all other packages in this project.
This way, we have a centralized information object where to get
information from.
This way we make sure the information is only stored once.

*/

package info

import (
	"log"
	"net"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/config"
	"github.com/migalabs/armiarma/src/utils"
)

type InfoData struct {
	base    base.Base
	iP      net.IP
	tcpPort int
	udpPort int

	userAgent string

	topicArray []string
	network    string
	forkDigest string

	logLevel   string
	privateKey *crypto.Secp256k1PrivateKey
}

// Will create an InfoData object using default values from config
func NewDefaultInfoData(opts base.LogOpts) *InfoData { //input_log log.Logger) *InfoData {

	baseConfigOpts := base.LogOpts{
		ModName:   "Config",
		Output:    "terminal",
		Formatter: "text",
		Level:     "debug",
	}

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

	config_object := config.NewEmptyConfigData(baseConfigOpts)

	info_object := InfoData{
		base: *new_base,
	}

	info_object.importFromConfig(*config_object)

	return &info_object
}

// Will create an InfoData object using imported values from config
func NewCustomInfoData(input_file string, opts base.LogOpts) *InfoData { //, input_log log.Logger) *InfoData {

	baseConfigOpts := base.LogOpts{
		ModName:   "Config",
		Output:    "terminal",
		Formatter: "text",
		Level:     "debug",
	}

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

	config_object := config.NewEmptyConfigData(baseConfigOpts)
	config_object.ReadFromJSON(input_file)

	info_object := InfoData{
		base: *new_base,
	}
	info_object.importFromConfig(*config_object)
	return &info_object
}

// This function will import the config values into the current InfoData
// object
func (i *InfoData) importFromConfig(input_config config.ConfigData) {

	i.base.Log.Debugf("Importing from Config into Info...")
	i.SetIPFromString(input_config.GetIP())
	i.SetTcpPort(input_config.GetTcpPort())
	i.SetUdpPort(input_config.GetUdpPort())
	i.SetUserAgent(input_config.GetUserAgent())
	i.SetTopicArray(input_config.GetTopicArray())
	i.SetNetwork(input_config.GetNetwork())
	i.SetForkDigest(input_config.GetForkDigest())
	i.SetLogLevel(input_config.GetLogLevel())
	i.SetPrivKeyFromString(input_config.GetPrivKey())
	i.base.Log.Debugf("Imported!")

}

// getters and setters

func (i *InfoData) GetTcpPort() int {
	return i.tcpPort
}
func (i *InfoData) SetTcpPort(input_port int) {
	if input_port > 65000 || input_port < 0 {
		// i.logging.Printf("TCP port not valid, applying default %d", config.DEFAULT_TCP_PORT)
		i.tcpPort = config.DEFAULT_TCP_PORT
		return
	}
	i.tcpPort = input_port
}

func (i *InfoData) GetUdpPort() int {

	return i.udpPort
}
func (i *InfoData) SetUdpPort(input_port int) {
	if input_port > 65000 || input_port < 0 {
		// i.logging.Printf("UDP port not valid, applying default %d", config.DEFAULT_UDP_PORT)
		i.udpPort = config.DEFAULT_UDP_PORT
		return
	}
	i.udpPort = input_port
}

func (i *InfoData) GetIP() net.IP {
	return i.iP
}
func (i *InfoData) GetIPToString() string {
	return i.GetIP().String()
}
func (i *InfoData) SetIP(input_ip net.IP) {
	i.iP = input_ip
}
func (i *InfoData) SetIPFromString(input_ip string) {
	i.iP = net.ParseIP(input_ip)
}

func (i *InfoData) GetUserAgent() string {
	return i.userAgent
}
func (i *InfoData) SetUserAgent(input_string string) {
	i.userAgent = input_string
}

func (i *InfoData) GetTopicArray() []string {
	return i.topicArray
}
func (i *InfoData) SetTopicArray(input_list []string) {
	i.topicArray = input_list
}

func (i *InfoData) GetNetwork() string {
	return i.network
}
func (i *InfoData) SetNetwork(input_string string) {
	i.network = input_string
}

func (i *InfoData) GetForkDigest() string {
	return i.forkDigest
}
func (i *InfoData) SetForkDigest(input_string string) {
	i.forkDigest = input_string
}

func (i *InfoData) GetLogLevel() string {
	return i.logLevel
}
func (i *InfoData) SetLogLevel(input_string string) {
	i.logLevel = input_string
}

func (i *InfoData) GetPrivKey() *crypto.Secp256k1PrivateKey {
	return i.privateKey
}
func (i *InfoData) GetPrivKeyString() string {
	return utils.PrivKeyToString(i.GetPrivKey())
}
func (i *InfoData) SetPrivKey(input_key *crypto.Secp256k1PrivateKey) {
	i.privateKey = input_key
}
func (i *InfoData) SetPrivKeyFromString(input_key string) {
	parsed_key, err := utils.ParsePrivateKey(input_key)

	if err != nil {
		// i.logging.Panicf("Wrong Private Key parsing %s", input_key)
		return
	}

	i.privateKey = parsed_key

}
