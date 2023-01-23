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

	cli "github.com/urfave/cli/v2"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/migalabs/armiarma/pkg/config"
	"github.com/migalabs/armiarma/pkg/gossipsub/blockchaintopics"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/sirupsen/logrus"

	rendp "github.com/migalabs/armiarma/pkg/networks/ethereum/remoteendpoint"
)

var (
	log = logrus.WithField(
		"module", "INFO",
	)
)

type Eth2InfoData struct {
	IP            net.IP
	TcpPort       int
	UdpPort       int
	UserAgent     string
	TopicArray    []string
	Network       string
	ForkDigest    string
	DbEndpoint    string
	Eth2endpoint  string
	LogLevel      string
	PrivateKey    *crypto.Secp256k1PrivateKey
	BootNodesFile string
	OutputPath    string
}

func InitEth2(c *cli.Context) (Eth2InfoData, error) {
	// read the config file given by the user
	var confFile string
	if c.IsSet("config-file") {
		confFile = c.String("config-file")
	} else {
		confFile = DefaultEth2ConfigFile
	}
	// check the config file given by the user, or check the defautl one
	// give error if not
	configObj := config.NewConfigFromFile(confFile)

	// parse the config file and generate the info struct
	infObj := Eth2infoFromConfig(configObj)

	return infObj, nil
}

// importFromConfig:
// This method will import all data from the given ConfigData object.
// As soon as we read the log level from the config object
// we create the logger object.
// @param inputConfig object to import data from.
// @param stdOpts base logging options.
func Eth2infoFromConfig(inputConfig config.ConfigData) Eth2InfoData {
	i := Eth2InfoData{}
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
		i.TcpPort = DefaultTcpPort
		log.Warnf("Setting default TcpPort: %d", DefaultTcpPort)
	} else {
		i.SetTcpPort(inputConfig.TcpPort)
	}

	if !checkValidPort(inputConfig.UdpPort) {
		i.UdpPort = DefaultUdpPort
		log.Warnf("Setting default UdpPort: %d", DefaultUdpPort)
	} else {
		i.SetUdpPort(inputConfig.UdpPort)
	}

	// UserAgent
	if inputConfig.UserAgent == "" {
		i.UserAgent = DefaultUserAgent
		log.Warnf("Setting default UserAgent: %s", DefaultUserAgent)
	} else {
		i.UserAgent = inputConfig.UserAgent
	}

	// Nework
	if inputConfig.Network == "" {
		i.Network = DefaultEth2Network
		log.Warnf("Setting default Network: %s", DefaultEth2Network)
	} else {
		i.Network = inputConfig.Network
	}

	// Eth2 Endpoint
	// Check if any Eth2Endpoint was given to get the ForkDigest
	if inputConfig.Eth2Endpoint == "" {
		// some endpoint was given
		log.Warnf("No Eth2 Endpoint was given")
	} else {
		i.Eth2endpoint = inputConfig.Eth2Endpoint
	}

	// Eth2 Endpoint
	// Check if any Eth2Endpoint was given to get the ForkDigest
	if inputConfig.DBEndpoint != "" {
		i.DbEndpoint = inputConfig.DBEndpoint
	}

	// Fork digest
	valid := i.SetForkDigest(inputConfig.ForkDigest)
	if !valid {
		// Check if any Eth2Endpoint was given to get the ForkDigest
		if i.Eth2endpoint != "" {
			infuraCli, err := rendp.NewInfuraClient(i.Eth2endpoint)
			if err != nil {
				log.Warnf("unable to genereate the eth2 endpoint from the given one. %s", err.Error())
				_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
				log.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
			} else {
				ctx, _ := context.WithCancel(context.Background())
				//defer cancel()
				forkdigest, err := rendp.GetForkDigetsOfEth2Head(ctx, &infuraCli)
				if err != nil {
					log.Warnf("unable to compute the fork digest from the eth2 endpoint. %s", err.Error())
					_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
					log.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
				} else {
					valid = i.SetForkDigest(forkdigest.String())
					if !valid {
						log.Warnf("unable to set the computed fork digest. %s", forkdigest.String())
						_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
						log.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
					}
				}
			}
		} else {
			log.Warnf("invalid fork digest and no endpoint given")
			_ = i.SetForkDigest(blockchaintopics.DefaultForkDigest)
			log.Warnf("Setting default ForkDigest to latest in mainnet: %s", blockchaintopics.DefaultForkDigest)
		}
	}
	log.Info("fork digest:", i.ForkDigest)

	// make sure we have already configured the ForkDigest

	//Topic
	valid = i.SetTopicArray(inputConfig.TopicArray)
	if !valid {
		defaultTopicList := blockchaintopics.MessageTypes
		i.SetTopicArray(defaultTopicList)
		log.Warnf("Setting default TopicArray: %s", defaultTopicList)
	}

	// Private Key
	err := i.SetPrivKeyFromString(inputConfig.PrivateKey)
	if err != nil {
		log.Warnf("%s. Generating a new one", err.Error())
		i.SetPrivKeyFromString(utils.GeneratePrivKey())
	}
	log.Infof("Private Key of the host: %s", i.GetPrivKeyString())

	// BootNodesFile
	if !utils.CheckFileExists(inputConfig.BootNodesFile) {
		// file does not exist
		i.BootNodesFile = DefaultEth2BootNodesFile
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

func (i *Eth2InfoData) SetTcpPort(input_port int) {
	if !checkValidPort(input_port) {
		log.Debugf("TCP port not valid: %d", input_port)
		return
	}
	i.TcpPort = input_port
}

func (i Eth2InfoData) GetUdpPortString() string {
	return fmt.Sprintf("%d", i.UdpPort)
}
func (i *Eth2InfoData) SetUdpPort(input_port int) {
	if !checkValidPort(input_port) {
		log.Debugf("UDP port not valid: %d", input_port)
		return
	}
	i.UdpPort = input_port
}

func checkValidPort(input_port int) bool {
	// we put greater than min port, as 0 is default when no value was set
	if input_port > MinPort && input_port <= MaxPort {
		return true
	}
	return false
}

// SetTopicArray:
// This method loops over the given array and validate that topics exist before setting the array.
// We need at least one valid.
// @return boolean in case any topic in the list was applied (true) or none was applied (false).
func (i *Eth2InfoData) SetTopicArray(inputList []string) bool {
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
func (i *Eth2InfoData) SetTopicArrayFromString(input_list string) bool {
	topicStringArray := strings.Split(input_list, ",")
	return i.SetTopicArray(topicStringArray)
}

func (i *Eth2InfoData) SetForkDigest(inputString string) bool {
	new_fork_digest, valid := blockchaintopics.CheckValidForkDigest(inputString)
	if valid {
		i.ForkDigest = new_fork_digest
		return true
	}
	return false

}

func (i Eth2InfoData) checkValidLogLevel(input_level string) bool {
	for _, log_level := range PossibleLogLevels {
		if strings.ToLower(input_level) == strings.ToLower(log_level) {
			return true
		}
	}
	return false
}

func (i Eth2InfoData) GetPrivKeyString() string {
	return utils.PrivKeyToString(i.PrivateKey)
}

func (i *Eth2InfoData) SetPrivKeyFromString(input_key string) error {
	parsed_key, err := utils.ParsePrivateKey(input_key)

	if err != nil {
		error_string := "Could not parse Private Key"
		return errors.New(error_string)
	}
	i.PrivateKey = parsed_key
	return nil
}
