package info

import (
	"fmt"
	"net"

	"github.com/migalabs/armiarma/src/config"
)

type InfoData struct {
	iP      net.IP
	tcpPort int
	udpPort int

	userAgent string

	topicArray []string
	network    string
	forkDigest string

	logLevel   string
	privateKey string
}

func NewDefaultInfoData() *InfoData {

	config_object := config.NewDefaultConfigData()

	info_object := InfoData{}

	err := info_object.importFromConfig(*config_object)
	if err != nil {
		fmt.Println(err)
	}

	return &info_object
}

func NewCustomInfoData(input_file string) *InfoData {

	config_object := config.NewDefaultConfigData()
	err := config_object.ReadFromJSON(input_file)
	if err != nil {
		fmt.Println(err)
	}

	info_object := InfoData{}
	err = info_object.importFromConfig(*config_object)
	if err != nil {
		fmt.Println(err)
	}

	return &info_object
}

func (i *InfoData) importFromConfig(input_config config.ConfigData) error {

	i.SetIPFromString(input_config.GetIP())
	i.SetTcpPort(input_config.GetTcpPort())
	i.SetUdpPort(input_config.GetUdpPort())

	i.SetUserAgent(input_config.GetUserAgent())

	i.SetTopicArray(input_config.GetTopicArray())
	i.SetNetwork(input_config.GetNetwork())
	i.SetForkDigest(input_config.GetForkDigest())
	i.SetLogLevel(input_config.GetLogLevel())

}

// getters and setters

func (i *InfoData) GetTcpPort() int {
	return i.tcpPort
}
func (i *InfoData) SetTcpPort(input_port int) {
	i.tcpPort = input_port
}

func (i *InfoData) GetUdpPort() int {
	return i.udpPort
}
func (i *InfoData) SetUdpPort(input_port int) {
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
