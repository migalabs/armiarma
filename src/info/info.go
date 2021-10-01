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
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/config"
	"github.com/migalabs/armiarma/src/utils"
	log "github.com/sirupsen/logrus"
)

const PKG_NAME string = "INFO"

type InfoData struct {
	localLogger   log.FieldLogger
	iP            net.IP
	tcpPort       int
	udpPort       int
	userAgent     string
	topicArray    []string
	network       string
	forkDigest    string
	logLevel      string
	privateKey    *crypto.Secp256k1PrivateKey
	bootNodesFile string
}

// Will create an InfoData object using default values from config
// Receive stdOpts (meaning, without the mod name and the level)
func NewDefaultInfoData(stdOpts base.LogOpts) *InfoData {

	config_object := config.NewEmptyConfigData(stdOpts)

	info_object := InfoData{}

	info_object.importFromConfig(*config_object, stdOpts)

	return &info_object
}

// Will create an InfoData object using imported values from config
func NewCustomInfoData(input_file string, stdOpts base.LogOpts) *InfoData {

	config_object := config.NewEmptyConfigData(stdOpts)
	config_object.ReadFromJSON(input_file)

	info_object := InfoData{}
	info_object.importFromConfig(*config_object, stdOpts)

	return &info_object
}

// This function will import the config values into the current InfoData
// object
func (i *InfoData) importFromConfig(input_config config.ConfigData, stdOpts base.LogOpts) {

	i.SetLogLevel(input_config.GetLogLevel())
	//set the local logger usign the stadOpts and the custom info opts
	infoLogOpts := i.infoLoggerOpts(stdOpts)
	i.localLogger = base.CreateLogger(infoLogOpts)
	i.localLogger.Infof("Importing from Config into Info...")
	i.SetIPFromString(input_config.GetIP())
	i.SetTcpPort(input_config.GetTcpPort())
	i.SetUdpPort(input_config.GetUdpPort())
	i.SetUserAgent(input_config.GetUserAgent())
	i.SetTopicArray(input_config.GetTopicArray())
	i.SetNetwork(input_config.GetNetwork())
	i.SetForkDigest(input_config.GetForkDigest())
	i.SetLogLevel(input_config.GetLogLevel())

	i.SetPrivKeyFromString(input_config.GetPrivKey())
	i.SetBootNodeFile(input_config.GetBootNodesFile())
	i.localLogger.Infof("Imported!")

}

func (i *InfoData) infoLoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PKG_NAME
	input_opts.Level = i.GetLogLevel()

	return input_opts
}

// getters and setters

func (i *InfoData) GetTcpPort() int {
	return i.tcpPort
}
func (i *InfoData) GetTcpPortString() string {

	return fmt.Sprintf("%d", i.tcpPort)
}
func (i *InfoData) SetTcpPort(input_port int) {
	if input_port > 65000 || input_port < 0 {
		i.localLogger.Debugf("TCP port not valid, applying default %d", config.DEFAULT_TCP_PORT)
		i.tcpPort = config.DEFAULT_TCP_PORT
		return
	}
	i.tcpPort = input_port
}

func (i *InfoData) GetUdpPort() int {

	return i.udpPort
}
func (i *InfoData) GetUdpPortString() string {

	return fmt.Sprintf("%d", i.udpPort)
}
func (i *InfoData) SetUdpPort(input_port int) {
	if input_port > 65000 || input_port < 0 {
		i.localLogger.Debugf("UDP port not valid, applying default %d", config.DEFAULT_UDP_PORT)
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
		i.localLogger.Panicf("Could not parse Private Key %s", input_key)
		return
	}
	i.privateKey = parsed_key
}

func (i *InfoData) GetBootNodeFile() string {
	return i.bootNodesFile
}
func (i *InfoData) SetBootNodeFile(input_string string) {
	i.bootNodesFile = input_string
}
