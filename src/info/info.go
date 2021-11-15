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
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/config"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/gossipsub/blockchaintopics"
	"github.com/migalabs/armiarma/src/utils"
	log "github.com/sirupsen/logrus"

	"github.com/migalabs/armiarma/src/onchaindata/eth2"
	"github.com/migalabs/armiarma/src/onchaindata/eth2/endpoint"
)

var (
	PkgName string = "INFO"
)

// define constant variables
var (
	DefaultIP            string = "0.0.0.0"
	DefaultTcpPort       int    = 9020
	DefaultUdpPort       int    = 9020
	DefaultNetwork       string = "mainnet"
	DefaultUserAgent     string = "bsc_crawler"
	DefaultLogLevel      string = "info"
	DefaultOutputPath    string = "./peerstore"
	DefaultDBType        string = "bolt"
	DefaultBootNodesFile string = "./src/discovery/bootnodes_mainnet.json"

	MinPort           int      = 0
	MaxPort           int      = 65000
	PossibleLogLevels []string = []string{"info", "debug"}
)

type InfoData struct {
	localLogger   log.FieldLogger
	iP            net.IP
	tcpPort       int
	udpPort       int
	userAgent     string
	topicArray    []string
	network       string
	forkDigest    string
	eth2endpoint  string
	logLevel      string
	privateKey    *crypto.Secp256k1PrivateKey
	bootNodesFile string
	OutputPath    string
	dBType        string
}

// NewDefaultInfoData
// * This method will create an empty InfoData object
// * This method will create an InfoData object
// * using default values from config
// @param stdOpts (meaning, without the mod name and the level)
// @return An InfoData object
func NewDefaultInfoData(stdOpts base.LogOpts) InfoData {

	config_object := config.NewEmptyConfigData(stdOpts)

	info_object := InfoData{}

	info_object.importFromConfig(config_object, stdOpts)

	return info_object
}

// NewCustomInfoData
// * This method will create an InfoData object
// * using imported values from config
// @param input_file which to give to the ConfigData object
// @param stdOpts (meaning, mod name and the level will be added here)
// @return An InfoData object
func NewCustomInfoData(input_file string, stdOpts base.LogOpts) *InfoData {

	config_object := config.NewEmptyConfigData(stdOpts)
	config_object.ReadFromJSON(input_file)

	info_object := InfoData{}
	info_object.importFromConfig(config_object, stdOpts)

	return &info_object
}

// importFromConfig
// * This method will import all data from the given ConfigData object
// * As soon as we read the log level from the config object
// * we create the logger object
// @param input_config object to import data from
// @param stdOpts base logging options
func (i *InfoData) importFromConfig(input_config config.ConfigData, stdOpts base.LogOpts) {

	// first of all import the log level
	default_log_level := false
	if !i.checkValidLogLevel(input_config.GetLogLevel()) {
		i.SetLogLevel(DefaultLogLevel)
		default_log_level = true
	} else {
		i.SetLogLevel(input_config.GetLogLevel())
	}

	//set the local logger using the stadOpts and the custom info opts
	infoLogOpts := i.infoLoggerOpts(stdOpts)
	i.localLogger = base.CreateLogger(infoLogOpts)
	if default_log_level {
		i.localLogger.Warnf("Setting default LogLevel: %s", DefaultLogLevel)
	}

	// start full import
	i.localLogger.Infof("Importing Configuration...")
	i.localLogger.Infof("setting logs to %s", i.GetLogLevel())
	//IP
	if utils.CheckValidIP(input_config.GetIP()) {
		i.SetIPFromString(input_config.GetIP())

	} else {
		i.SetIPFromString(DefaultIP)
		i.localLogger.Warnf("Setting default IP: %s", DefaultIP)
	}
	// Ports

	if !checkValidPort(input_config.GetTcpPort()) {
		i.SetTcpPort(DefaultTcpPort)
		i.localLogger.Warnf("Setting default TcpPort: %d", DefaultTcpPort)
	} else {
		i.SetTcpPort(input_config.GetTcpPort())
	}

	if !checkValidPort(input_config.GetUdpPort()) {
		i.SetUdpPort(DefaultUdpPort)
		i.localLogger.Warnf("Setting default UdpPort: %d", DefaultUdpPort)
	} else {
		i.SetUdpPort(input_config.GetUdpPort())
	}

	// UserAgent
	if input_config.GetUserAgent() == "" {
		i.SetUserAgent(DefaultUserAgent)
		i.localLogger.Warnf("Setting default UserAgent: %s", DefaultUserAgent)
	} else {
		i.SetUserAgent(input_config.GetUserAgent())
	}

	// Nework
	if input_config.GetNetwork() == "" {
		i.SetNetwork(DefaultNetwork)
		i.localLogger.Warnf("Setting default Network: %s", DefaultNetwork)
	} else {
		i.SetNetwork(input_config.GetNetwork())
	}

	// Eth2 Endpoint
	// Check if any Eth2Endpoint was given to get the ForkDigest
	if input_config.GetEth2Endpoint() == "" {
		// some endpoint was given
		i.localLogger.Warnf("No Eth2 Endpoint was given")
	} else {
		i.SetEth2Endpoint(input_config.GetEth2Endpoint())
	}

	// Fork digest
	valid := i.SetForkDigest(input_config.GetForkDigest())
	if !valid {
		// Check if any Eth2Endpoint was given to get the ForkDigest
		if i.GetEth2Endpoint() != "" {
			infuraCli, err := endpoint.NewInfuraClient(i.GetEth2Endpoint())
			if err != nil {
				i.localLogger.Warnf("unable to genereate the eth2 endpoint from the given one. %s", err.Error())
				_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
				i.localLogger.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
			} else {
				ctx, _ := context.WithCancel(context.Background())
				//defer cancel()
				forkdigest, err := eth2.GetForkDigetsOfEth2Head(ctx, &infuraCli)
				if err != nil {
					i.localLogger.Warnf("unable to compute the fork digest from the eth2 endpoint. %s", err.Error())
					_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
					i.localLogger.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
				} else {
					valid = i.SetForkDigest(forkdigest.String())
					if !valid {
						i.localLogger.Warnf("unable to set the computed fork digest. %s", forkdigest.String())
						_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
						i.localLogger.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
					}
				}
			}
		} else {
			i.localLogger.Warnf("invalid fork digest and no endpoint given")
			_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
			i.localLogger.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
		}
	}
	i.localLogger.Info("fork digest:", i.GetForkDigest())

	// make sure we have already configured the ForkDigest

	//Topic
	valid = i.SetTopicArray(input_config.GetTopicArray())
	if !valid {
		defaultTopicList := blockchaintopics.MessageTypes
		i.SetTopicArray(defaultTopicList)
		i.localLogger.Warnf("Setting default TopicArray: %s", defaultTopicList)
	}

	// Private Key
	err := i.SetPrivKeyFromString(input_config.GetPrivKey())
	if err != nil {
		i.localLogger.Warnf("%s. Generating a new one", err.Error())
		i.SetPrivKeyFromString(utils.Generate_privKey())
	}
	i.localLogger.Infof("Private Key of the host: %s", i.GetPrivKeyString())

	// BootNodesFile
	if !utils.CheckFileExists(input_config.GetBootNodesFile()) {
		// file does not exist
		i.SetBootNodeFile(DefaultBootNodesFile)
		i.localLogger.Warnf("Could not find bootnodes file, applying default...")

	} else {
		i.SetBootNodeFile(input_config.GetBootNodesFile())
	}

	// TODO: pending db type and path
	if input_config.GetOutputPath() == "" {
		i.localLogger.Warnf("Setting default Output Path: %s", DefaultOutputPath)
		i.SetOutputPath(DefaultOutputPath)
	} else {
		i.SetOutputPath(input_config.GetOutputPath())
	}

	// Check if the folder already exists, or generate one
	if !utils.CheckFileExists(i.GetOutputPath()) {
		// folder does not exist, generate a new one
		i.localLogger.Infof("Generating new folder in path %s", i.GetOutputPath())
		err := os.Mkdir(i.GetOutputPath(), 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	if _, ok := db.DBTypes[input_config.GetDBType()]; !ok {
		// type not okay, does not exist in our local hasmap
		i.SetDBType(DefaultDBType)
		i.localLogger.Warnf("Setting default DB Type: %s", DefaultDBType)
	} else {
		i.SetDBType(input_config.GetDBType())
	}

	i.localLogger.Infof("Imported!")
}

// infoLoggerOpts
// * This method will modify logging options accordingly for the InfoData object
// @param input_opts the base logging options
// @return the mordified logging options from the input
func (i InfoData) infoLoggerOpts(input_opts base.LogOpts) base.LogOpts {
	input_opts.ModName = PkgName
	input_opts.Level = i.GetLogLevel()

	return input_opts
}

// getters and setters

func (i InfoData) GetTcpPort() int {
	return i.tcpPort
}
func (i InfoData) GetTcpPortString() string {

	return fmt.Sprintf("%d", i.tcpPort)
}
func (i *InfoData) SetTcpPort(input_port int) {
	if !checkValidPort(input_port) {
		i.localLogger.Debugf("TCP port not valid: %d", input_port)
		return
	}
	i.tcpPort = input_port
}

func (i InfoData) GetUdpPort() int {

	return i.udpPort
}
func (i InfoData) GetUdpPortString() string {

	return fmt.Sprintf("%d", i.udpPort)
}
func (i *InfoData) SetUdpPort(input_port int) {
	if !checkValidPort(input_port) {
		i.localLogger.Debugf("UDP port not valid: %d", input_port)
		return
	}
	i.udpPort = input_port
}

func checkValidPort(input_port int) bool {
	// we put greater than min port, as 0 is default when no value was set
	if input_port > MinPort && input_port <= MaxPort {
		return true
	}
	return false
}

func (i InfoData) GetIP() net.IP {
	return i.iP
}
func (i InfoData) GetIPToString() string {
	return i.GetIP().String()
}
func (i *InfoData) SetIP(inputIp net.IP) {
	i.iP = inputIp
}
func (i *InfoData) SetIPFromString(inputIp string) {
	i.iP = net.ParseIP(inputIp)

}

func (i InfoData) GetUserAgent() string {
	return i.userAgent
}
func (i *InfoData) SetUserAgent(inputString string) {
	i.userAgent = inputString
}

func (i InfoData) GetTopicArray() []string {
	return i.topicArray
}

// SetTopicArray
// * This method loops over the given array and validate that topics exist before setting the array.
// * We need at least one valid
// @return boolean in case any topic in the list was applied (true) or none was applied (false)
func (i *InfoData) SetTopicArray(inputList []string) bool {
	resultTopicList := make([]string, 0)
	if len(inputList) > 0 {
		for _, inputTopic := range inputList {
			if utils.ExistsInArray(blockchaintopics.MessageTypes, inputTopic) {
				// topic exists
				resultTopicList = append(resultTopicList, inputTopic)
				continue // go to next inputTopic
			}
			i.localLogger.Warnf("Could not validate topic: %s", inputTopic)
		}

		if len(resultTopicList) > 0 {
			i.topicArray = resultTopicList
			return true
		}
		return false // no topic was applied

	} else {
		i.localLogger.Warnf("Empty topic list")
		return false
	}

}
func (i *InfoData) SetTopicArrayFromString(input_list string) bool {
	topicStringArray := strings.Split(input_list, ",")
	return i.SetTopicArray(topicStringArray)
}

func (i InfoData) GetNetwork() string {
	return i.network
}
func (i *InfoData) SetNetwork(input_string string) {
	i.network = input_string
}

func (i InfoData) GetEth2Endpoint() string {
	return i.eth2endpoint
}
func (i *InfoData) SetEth2Endpoint(input_string string) {
	i.eth2endpoint = input_string
}

func (i InfoData) GetForkDigest() string {
	return i.forkDigest
}
func (i *InfoData) SetForkDigest(input_string string) bool {
	new_fork_digest, valid := blockchaintopics.CheckValidForkDigest(input_string)
	if valid {
		i.forkDigest = new_fork_digest
		return true
	}
	return false

}

func (i InfoData) GetLogLevel() string {
	return i.logLevel
}
func (i *InfoData) SetLogLevel(input_string string) {
	i.logLevel = input_string
}
func (i InfoData) checkValidLogLevel(input_level string) bool {
	for _, log_level := range PossibleLogLevels {
		if strings.ToLower(input_level) == strings.ToLower(log_level) {
			return true
		}
	}
	return false
}

func (i InfoData) GetPrivKey() *crypto.Secp256k1PrivateKey {
	return i.privateKey
}
func (i InfoData) GetPrivKeyString() string {
	return utils.PrivKeyToString(i.GetPrivKey())
}
func (i *InfoData) SetPrivKey(input_key *crypto.Secp256k1PrivateKey) {
	i.privateKey = input_key
}
func (i *InfoData) SetPrivKeyFromString(input_key string) error {
	parsed_key, err := utils.ParsePrivateKey(input_key)

	if err != nil {
		error_string := "Could not parse Private Key"
		return errors.New(error_string)
	}
	i.privateKey = parsed_key
	return nil
}

func (i InfoData) GetBootNodeFile() string {
	return i.bootNodesFile
}
func (i *InfoData) SetBootNodeFile(input_string string) {
	i.bootNodesFile = input_string
}

func (i InfoData) GetOutputPath() string {
	return i.OutputPath
}
func (i *InfoData) SetOutputPath(input_string string) {
	i.OutputPath = input_string
}

func (i InfoData) GetDBType() string {
	return i.dBType
}
func (i *InfoData) SetDBType(input_string string) {
	i.dBType = input_string
}
