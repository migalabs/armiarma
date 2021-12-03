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
	"github.com/migalabs/armiarma/src/config"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/gossipsub/blockchaintopics"
	"github.com/migalabs/armiarma/src/utils"
	"github.com/sirupsen/logrus"

	"github.com/migalabs/armiarma/src/onchaindata/eth2"
	"github.com/migalabs/armiarma/src/onchaindata/eth2/endpoint"
)

var (
	ModuleName = "INFO"
	Log        = logrus.WithField(
		"module", ModuleName,
	)
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
func NewDefaultInfoData() InfoData {

	configObj := config.NewEmptyConfig()

	infoObj := InfoData{}

	infoObj.importFromConfig(configObj)

	return infoObj
}

// NewCustomInfoData
// * This method will create an InfoData object
// * using imported values from givem config.ConfigData
// @param input ConfigData object
// @param stdOpts (meaning, mod name and the level will be added here)
// @return An InfoData object
func NewCustomInfoData(configObj config.ConfigData) *InfoData {

	infoObj := InfoData{}
	infoObj.importFromConfig(configObj)

	return &infoObj
}

// importFromConfig
// * This method will import all data from the given ConfigData object
// * As soon as we read the log level from the config object
// * we create the logger object
// @param inputConfig object to import data from
// @param stdOpts base logging options
func (i *InfoData) importFromConfig(inputConfig config.ConfigData) {

	// first of all import the log level
	if !i.checkValidLogLevel(inputConfig.GetLogLevel()) {
		i.SetLogLevel(DefaultLogLevel)
	} else {
		i.SetLogLevel(inputConfig.GetLogLevel())
	}

	// start full import
	Log.Infof("Importing Configuration...")
	Log.Infof("setting logs to %s", i.GetLogLevel())
	//IP
	if utils.CheckValidIP(inputConfig.GetIP()) {
		i.SetIPFromString(inputConfig.GetIP())

	} else {
		i.SetIPFromString(DefaultIP)
		Log.Warnf("Setting default IP: %s", DefaultIP)
	}
	// Ports

	if !checkValidPort(inputConfig.GetTcpPort()) {
		i.SetTcpPort(DefaultTcpPort)
		Log.Warnf("Setting default TcpPort: %d", DefaultTcpPort)
	} else {
		i.SetTcpPort(inputConfig.GetTcpPort())
	}

	if !checkValidPort(inputConfig.GetUdpPort()) {
		i.SetUdpPort(DefaultUdpPort)
		Log.Warnf("Setting default UdpPort: %d", DefaultUdpPort)
	} else {
		i.SetUdpPort(inputConfig.GetUdpPort())
	}

	// UserAgent
	if inputConfig.GetUserAgent() == "" {
		i.SetUserAgent(DefaultUserAgent)
		Log.Warnf("Setting default UserAgent: %s", DefaultUserAgent)
	} else {
		i.SetUserAgent(inputConfig.GetUserAgent())
	}

	// Nework
	if inputConfig.GetNetwork() == "" {
		i.SetNetwork(DefaultNetwork)
		Log.Warnf("Setting default Network: %s", DefaultNetwork)
	} else {
		i.SetNetwork(inputConfig.GetNetwork())
	}

	// Eth2 Endpoint
	// Check if any Eth2Endpoint was given to get the ForkDigest
	if inputConfig.GetEth2Endpoint() == "" {
		// some endpoint was given
		Log.Warnf("No Eth2 Endpoint was given")
	} else {
		i.SetEth2Endpoint(inputConfig.GetEth2Endpoint())
	}

	// Fork digest
	valid := i.SetForkDigest(inputConfig.GetForkDigest())
	if !valid {
		// Check if any Eth2Endpoint was given to get the ForkDigest
		if i.GetEth2Endpoint() != "" {
			infuraCli, err := endpoint.NewInfuraClient(i.GetEth2Endpoint())
			if err != nil {
				Log.Warnf("unable to genereate the eth2 endpoint from the given one. %s", err.Error())
				_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
				Log.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
			} else {
				ctx, _ := context.WithCancel(context.Background())
				//defer cancel()
				forkdigest, err := eth2.GetForkDigetsOfEth2Head(ctx, &infuraCli)
				if err != nil {
					Log.Warnf("unable to compute the fork digest from the eth2 endpoint. %s", err.Error())
					_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
					Log.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
				} else {
					valid = i.SetForkDigest(forkdigest.String())
					if !valid {
						Log.Warnf("unable to set the computed fork digest. %s", forkdigest.String())
						_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
						Log.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
					}
				}
			}
		} else {
			Log.Warnf("invalid fork digest and no endpoint given")
			_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
			Log.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
		}
	}
	Log.Info("fork digest:", i.GetForkDigest())

	// make sure we have already configured the ForkDigest

	//Topic
	valid = i.SetTopicArray(inputConfig.GetTopicArray())
	if !valid {
		defaultTopicList := blockchaintopics.MessageTypes
		i.SetTopicArray(defaultTopicList)
		Log.Warnf("Setting default TopicArray: %s", defaultTopicList)
	}

	// Private Key
	err := i.SetPrivKeyFromString(inputConfig.GetPrivKey())
	if err != nil {
		Log.Warnf("%s. Generating a new one", err.Error())
		i.SetPrivKeyFromString(utils.Generate_privKey())
	}
	Log.Infof("Private Key of the host: %s", i.GetPrivKeyString())

	// BootNodesFile
	if !utils.CheckFileExists(inputConfig.GetBootNodesFile()) {
		// file does not exist
		i.SetBootNodeFile(DefaultBootNodesFile)
		Log.Warnf("Could not find bootnodes file, applying default...")

	} else {
		i.SetBootNodeFile(inputConfig.GetBootNodesFile())
	}

	// TODO: pending db type and path
	if inputConfig.GetOutputPath() == "" {
		Log.Warnf("Setting default Output Path: %s", DefaultOutputPath)
		i.SetOutputPath(DefaultOutputPath)
	} else {
		i.SetOutputPath(inputConfig.GetOutputPath())
	}

	// Check if the folder already exists, or generate one
	if !utils.CheckFileExists(i.GetOutputPath()) {
		// folder does not exist, generate a new one
		Log.Infof("Generating new folder in path %s", i.GetOutputPath())
		err := os.Mkdir(i.GetOutputPath(), 0755)
		if err != nil {
			Log.Fatal(err)
		}
	}

	if _, ok := db.DBTypes[inputConfig.GetDBType()]; !ok {
		// type not okay, does not exist in our local hasmap
		i.SetDBType(DefaultDBType)
		Log.Warnf("Setting default DB Type: %s", DefaultDBType)
	} else {
		i.SetDBType(inputConfig.GetDBType())
	}

	Log.Infof("Imported!")
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
		Log.Debugf("TCP port not valid: %d", input_port)
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
		Log.Debugf("UDP port not valid: %d", input_port)
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
			Log.Warnf("Could not validate topic: %s", inputTopic)
		}

		if len(resultTopicList) > 0 {
			i.topicArray = resultTopicList
			return true
		}
		return false // no topic was applied

	} else {
		Log.Warnf("Empty topic list")
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
