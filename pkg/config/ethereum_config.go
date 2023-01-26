package config

import (
	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var (
	// Ethereum Mainnet
	DefaultMainnetForkDigest string = eth.ForkDigests[eth.BellatrixKey]

	// Genosis Mainnet
	DefaultGnosisForkDigest string = eth.ForkDigests[eth.GnosisBellatrixKey]

	// GossipSub Topics
	DefaultEthereumGossipTopics []string = []string{}
	AllEthereumGossipTopics     []string = eth.MessageTypes

	// remote Infura endpoint
	DefaultCLRemoteEndpoint = ""
)

type EthereumCrawlerConfig struct {
	LogLevel            string   `json:"log-level"`
	PrivateKey          string   `json:"priv-key"`
	IP                  string   `json:"ip"`
	Port                int      `json:"port"`
	UserAgent           string   `json:"user-agent"`
	GossipTopics        []string `json:"gossip-topics"`
	EthCLRemoteEndpoint string   `json:"remote-cl-endpoint"`
	PsqlEndpoint        string   `json:"psql-endpoint"`
	ForkDigest          string   `json:"fork-digest"`
	Bootnodes           []string `json:"bootnodes"`
	LocalPeerstorePath  string   `json:"local-peerstore-path"`
}

// TODO: read from config-file

func NewEthereumCrawlerConfig() *EthereumCrawlerConfig {
	// Return Default values for the ethereum configuration
	return &EthereumCrawlerConfig{
		LogLevel:            DefaultLogLevel,
		PrivateKey:          DefaultPrivKey,
		IP:                  DefaultIP,
		Port:                DefaultPort,
		UserAgent:           DefaultUserAgent,
		GossipTopics:        DefaultEthereumGossipTopics,
		EthCLRemoteEndpoint: DefaultCLRemoteEndpoint,
		PsqlEndpoint:        DefaultPSQLEndpoint,
		ForkDigest:          DefaultMainnetForkDigest,
		Bootnodes:           DefaultEthereumBootnodes,
		LocalPeerstorePath:  DefaultLocalPeerstorePath,
	}
}

func (c *EthereumCrawlerConfig) Apply(ctx *cli.Context) {
	// apply to the existing Default configuration the set flags
	// log level
	if ctx.IsSet("log-level") {
		c.LogLevel = ctx.String("log-level")
	}
	// private key
	if ctx.IsSet("priv-key") {
		c.PrivateKey = ctx.String("priv-key")
	}
	// ip
	if ctx.IsSet("ip") {
		c.IP = ctx.String("ip")
	}
	// port
	if ctx.IsSet("port") {
		port := ctx.Int("port")
		if checkValidPort(port) {
			c.Port = port
		}
	}
	// user agent
	if ctx.IsSet("user-agent") {
		c.UserAgent = ctx.String("user-agent")
	}
	// gossip topics
	if ctx.IsSet("gossip-topic") {
		c.GossipTopics = ctx.StringSlice("gossip-topic")
	}
	// eth cl remote endpoint
	if ctx.IsSet("remote-cl-endpoint") {
		c.EthCLRemoteEndpoint = ctx.String("remote-cl-endpoint")
	}
	// postgresql endpoint
	if ctx.IsSet("psql-endpoint") {
		c.PsqlEndpoint = ctx.String("psql-endpoint")
	}
	// fork digest
	if ctx.IsSet("fork-digest") {
		forkDigest := ctx.String("fork-digest")
		validForkDigest, valid := eth.CheckValidForkDigest(forkDigest)
		if valid {
			c.ForkDigest = validForkDigest
		}
	}
	// bootnodes
	if ctx.IsSet("bootnode") {
		c.Bootnodes = ctx.StringSlice("bootnode")
	}
	// local peerstore path
	if ctx.IsSet("local-peerstore") {
		c.LocalPeerstorePath = ctx.String("local-peerstore")
	}
	err := validateOrCreatePeerstore(c.LocalPeerstorePath)
	if err != nil {
		log.Panic("unable to create folder for local-peerstore" + err.Error())
	}
	log.WithFields(log.Fields{
		"log-level":     c.LogLevel,
		"priv-key":      c.PrivateKey,
		"ip":            c.IP,
		"port":          c.Port,
		"user-agent":    c.UserAgent,
		"psql":          c.PsqlEndpoint,
		"gossip-topics": c.GossipTopics,
		"cl-endpoint":   c.EthCLRemoteEndpoint,
		"bootnodes":     c.Bootnodes,
		"peerstore":     c.LocalPeerstorePath,
	}).Info("config for the Ethereum crawler")
}
