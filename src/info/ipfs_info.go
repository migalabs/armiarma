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
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	cli "github.com/urfave/cli/v2"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/migalabs/armiarma/src/config"
	"github.com/migalabs/armiarma/src/gossipsub/blockchaintopics"
	"github.com/migalabs/armiarma/src/utils"
)

type IpfsInfoData struct {
	IP        net.IP
	TcpPort   int
	UdpPort   int
	UserAgent string
	Network   string

	TopicArray    []string
	DbEndpoint    string
	LogLevel      string
	PrivateKey    *crypto.Secp256k1PrivateKey
	BootNodesFile string
	OutputPath    string
}

func InitIpfs(c *cli.Context) (IpfsInfoData, error) {
	// read the config file given by the user
	var confFile string
	if c.IsSet("config-file") {
		confFile = c.String("config-file")
	} else {
		confFile = DefaultIpfsConfigFile
	}
	// check the config file given by the user, or check the defautl one
	// give error if not
	configObj := config.NewConfigFromFile(confFile)

	// parse the config file and generate the info struct
	infoObj := IPFSinfoFromConfig(configObj)

	return infoObj, nil
}

// importFromConfig:
// This method will import all data from the given ConfigData object.
// As soon as we read the log level from the config object
// we create the logger object.
// @param inputConfig object to import data from.
// @param stdOpts base logging options.
func IPFSinfoFromConfig(inputConfig config.ConfigData) IpfsInfoData {
	i := IpfsInfoData{}
	// first of all import the log level
	if !i.checkValidLogLevel(inputConfig.LogLevel) {
		i.LogLevel = DefaultLogLevel
	} else {
		i.LogLevel = inputConfig.LogLevel
	}

	logrus.SetLevel(utils.ParseLogLevel(i.LogLevel))

	// start full import
	log.Infof("Importing Configuration...")
	log.Infof("setting logs to %s", i.LogLevel)

	//IP
	if utils.CheckValidIP(inputConfig.IP) {
		i.IP = net.ParseIP(inputConfig.IP)

	} else {
		i.IP = net.ParseIP(DefaultIP)
		log.Warnf("Setting default IP: %s", DefaultIP)
	}
	// Ports

	if !checkValidPort(inputConfig.TcpPort) {
		i.SetTcpPort(DefaultTcpPort)
		log.Warnf("Setting default TcpPort: %d", DefaultTcpPort)
	} else {
		i.SetTcpPort(inputConfig.TcpPort)
	}

	if !checkValidPort(inputConfig.UdpPort) {
		i.SetUdpPort(DefaultUdpPort)
		log.Warnf("Setting default UdpPort: %d", DefaultUdpPort)
	} else {
		i.SetUdpPort(inputConfig.UdpPort)
	}

	// Network
	if inputConfig.Network == "" {
		i.Network = DefaultIpfsNetwork
		log.Warnf("Setting default network: %d", DefaultIpfsNetwork)
	} else {
		i.Network = inputConfig.Network
	}

	// UserAgent
	if inputConfig.UserAgent == "" {
		i.UserAgent = DefaultUserAgent
		log.Warnf("Setting default UserAgent: %s", DefaultUserAgent)
	} else {
		i.UserAgent = inputConfig.UserAgent
	}

	// Eth2 Endpoint
	// Check if any Eth2Endpoint was given to get the ForkDigest
	if inputConfig.DBEndpoint != "" {
		i.DbEndpoint = inputConfig.DBEndpoint
	}

	//Topic
	_ = i.SetTopicArray(inputConfig.TopicArray)

	// Private Key
	err := i.SetPrivKeyFromString(inputConfig.PrivateKey)
	if err != nil {
		log.Warnf("%s. Generating a new one", err.Error())
		i.PrivateKey = utils.GeneratePrivKey()
	}
	log.Infof("Private Key of the host: %s", i.GetPrivKeyString())

	// BootNodesFile
	if !utils.CheckFileExists(inputConfig.BootNodesFile) {
		// file does not exist
		i.BootNodesFile = DefaultIpfsBootNodesFile
		log.Warnf("Could not find bootnodes file, applying default...")

	} else {
		i.BootNodesFile = inputConfig.BootNodesFile
	}

	// TODO: pending db type and path
	if inputConfig.OutputPath == "" {
		log.Warnf("Setting default Output Path: %s", DefaultOutputPath)
		i.OutputPath = DefaultOutputPath
	} else {
		i.OutputPath = inputConfig.OutputPath
	}

	// Check if the folder already exists, or generate one
	if !utils.CheckFileExists(i.OutputPath) {
		// folder does not exist, generate a new one
		log.Infof("Generating new folder in path %s", i.OutputPath)
		err := os.Mkdir(i.OutputPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Infof("Imported!")
	return i
}

// getters and setters

func (i *IpfsInfoData) SetTcpPort(input_port int) {
	if !checkValidPort(input_port) {
		log.Debugf("TCP port not valid: %d", input_port)
		return
	}
	i.TcpPort = input_port
}

func (i IpfsInfoData) GetUdpPortString() string {
	return fmt.Sprintf("%d", i.UdpPort)
}
func (i *IpfsInfoData) SetUdpPort(input_port int) {
	if !checkValidPort(input_port) {
		log.Debugf("UDP port not valid: %d", input_port)
		return
	}
	i.UdpPort = input_port
}

// SetTopicArray:
// This method loops over the given array and validate that topics exist before setting the array.
// We need at least one valid.
// @return boolean in case any topic in the list was applied (true) or none was applied (false).
func (i *IpfsInfoData) SetTopicArray(inputList []string) bool {
	resultTopicList := make([]string, 0)
	if len(inputList) > 0 {
		for _, inputTopic := range inputList {
			if utils.ExistsInArray(blockchaintopics.MessageTypes, inputTopic) {
				// topic exists
				resultTopicList = append(resultTopicList, inputTopic)
				continue // go to next inputTopic
			}
			log.Warnf("Could not validate topic: %s", inputTopic)
		}

		if len(resultTopicList) > 0 {
			i.TopicArray = resultTopicList
			return true
		}
		return false // no topic was applied

	} else {
		log.Warnf("Empty topic list")
		return false
	}

}

func (i *IpfsInfoData) SetTopicArrayFromString(input_list string) bool {
	topicStringArray := strings.Split(input_list, ",")
	return i.SetTopicArray(topicStringArray)
}

func (i IpfsInfoData) checkValidLogLevel(input_level string) bool {
	for _, log_level := range PossibleLogLevels {
		if strings.ToLower(input_level) == strings.ToLower(log_level) {
			return true
		}
	}
	return false
}

func (i IpfsInfoData) GetPrivKeyString() string {
	return utils.PrivKeyToString(i.PrivateKey)
}

func (i *IpfsInfoData) SetPrivKeyFromString(input_key string) error {
	parsed_key, err := utils.ParsePrivateKey(input_key)

	if err != nil {
		error_string := "Could not parse Private Key"
		return errors.New(error_string)
	}
	i.PrivateKey = parsed_key
	return nil
}
