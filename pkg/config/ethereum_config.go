package config

import (
	"strconv"

	eth "github.com/migalabs/armiarma/pkg/networks/ethereum"
	rendp "github.com/migalabs/armiarma/pkg/networks/ethereum/remoteendpoint"
	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var (
	// GossipSub Topics
	DefaultEthereumGossipTopics []string = []string{}
	AllEthereumGossipTopics     []string = eth.MessageTypes
	DefaultSubnets              []int    = []int{}
	DefaultValPubkeys           []string = []string{}

	// remote RemoteEth client's endpoint
	DefaultCLRemoteEndpoint = ""
)

type EthereumCrawlerConfig struct {
	LogLevel                  string   `json:"log-level"`
	PrivateKey                string   `json:"priv-key"`
	IP                        string   `json:"ip"`
	Port                      int      `json:"port"`
	MetricsIP                 string   `json:"metrics-ip"`
	MetricsPort               int      `json:"metrics-port"`
	UserAgent                 string   `json:"user-agent"`
	EthCLRemoteEndpoint       string   `json:"remote-cl-endpoint"`
	PsqlEndpoint              string   `json:"psql-endpoint"`
	ActivePeersBackupInterval string   `json:ActivePeersBackupInterval`
	ForkDigest                string   `json:"fork-digest"`
	Bootnodes                 []string `json:"bootnodes"`
	GossipTopics              []string `json:"gossip-topics"`
	Subnets                   []int    `json:"subnets"`
	PersistMsgs               bool     `json:"persist-msgs"`
	ValPubkeys                []string `json:"val-pubkeys"`
}

// TODO: read from config-file
func NewEthereumCrawlerConfig() *EthereumCrawlerConfig {
	// Return Default values for the ethereum configuration
	return &EthereumCrawlerConfig{
		LogLevel:                  DefaultLogLevel,
		PrivateKey:                DefaultPrivKey,
		IP:                        DefaultIP,
		Port:                      DefaultPort,
		MetricsIP:                 DefaultMetricsIP,
		MetricsPort:               DefaultMetricsPort,
		UserAgent:                 DefaultUserAgent,
		EthCLRemoteEndpoint:       DefaultCLRemoteEndpoint,
		PsqlEndpoint:              DefaultPSQLEndpoint,
		ActivePeersBackupInterval: DefaultActivePeersBackupInterval,
		ForkDigest:                eth.DefaultForkDigest,
		Bootnodes:                 DefaultEthereumBootnodes,
		Subnets:                   DefaultSubnets,
		GossipTopics:              DefaultEthereumGossipTopics,
		PersistMsgs:               false,
		ValPubkeys:                DefaultValPubkeys,
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
	// metrics-ip (pprof + prometheus)
	if ctx.IsSet("metrics-ip") {
		c.MetricsIP = ctx.String("metrics-ip")

	}
	// metrics-port (pprof + prometheus)
	if ctx.IsSet("metrics-port") {
		mPort := ctx.Int("metrics-port")
		if checkValidPort(mPort) {
			c.MetricsPort = mPort
		}
	}
	// user agent
	if ctx.IsSet("user-agent") {
		c.UserAgent = ctx.String("user-agent")
	}

	// fork digest
	if ctx.IsSet("fork-digest") {
		forkDigest := ctx.String("fork-digest")
		validForkDigest, valid := eth.CheckValidForkDigest(forkDigest)
		if valid {
			c.ForkDigest = validForkDigest
		}
		// check if fork-digest is not empty -> eth-cl endpoint
		if forkDigest == "" && ctx.IsSet("remote-cl-endpoint") {
			c.EthCLRemoteEndpoint = ctx.String("remote-cl-endpoint")
			log.Warnf("fork_digest not provided - fetching latest one from %s", c.EthCLRemoteEndpoint)
			clEndp, err := rendp.NewInfuraClient(c.EthCLRemoteEndpoint)
			if err != nil {
				log.Panic(errors.Wrap(err, "unable to determine the latest fork_digest"))
			}
			forkD, err := rendp.GetForkDigetsOfEth2Head(ctx.Context, &clEndp)
			if err != nil {
				log.Panic(errors.Wrap(err, "unable to retreive the fork_digests from given rndp"))
			}
			c.ForkDigest = forkD.String()
		}
	}

	// postgresql endpoint
	if ctx.IsSet("psql-endpoint") {
		c.PsqlEndpoint = ctx.String("psql-endpoint")
	}

	// active peers' backup interval
	if ctx.IsSet("peers-backup") {
		c.ActivePeersBackupInterval = ctx.String("peers-backup")
	}

	// bootnodes
	if ctx.IsSet("bootnode") {
		c.Bootnodes = ctx.StringSlice("bootnode")
	}

	// gossip topics
	if ctx.IsSet("gossip-topic") {
		c.GossipTopics = ctx.StringSlice("gossip-topic")
	}

	// Subnets
	if ctx.IsSet("subnet") {
		subnets := ctx.StringSlice("subnet")
		allF := false
		for _, subn := range subnets {
			if subn == "all" { // check if the all flag was set
				allF = true
				break
			} else { // add the set flags otherwise
				subNum, err := strconv.Atoi(subn)
				if err != nil {
					log.Panic(errors.Wrap(err, "invalid subnet index"))
				}
				c.Subnets = append(c.Subnets, subNum)
			}
		}
		if allF {
			for i := 1; i <= 64; i++ {
				c.Subnets = append(c.Subnets, i)
			}
		}
	}

	// check if we want to track the Msgs in the SQL database
	if ctx.IsSet("persist-msgs") {
		c.PersistMsgs = ctx.Bool("persist-msgs")
	}

	// read validator-pubkeys .csv file if it exists
	if ctx.IsSet("val-pubkeys") {
		filePath := ctx.String("val-pubkeys")
		valKeys, err := utils.ReadFilePerRows(filePath, ",")
		if err != nil {
			log.Panic(errors.Wrap(err, "unable to read file with val-pubkeys"))
		}
		c.ValPubkeys = append(c.ValPubkeys, valKeys...)
	}

	log.WithFields(log.Fields{
		"log-level":       c.LogLevel,
		"priv-key":        c.PrivateKey,
		"ip":              c.IP,
		"port":            c.Port,
		"user-agent":      c.UserAgent,
		"psql":            c.PsqlEndpoint,
		"backup-interval": c.ActivePeersBackupInterval,
		"fork-digest":     c.ForkDigest,
		"cl-endpoint":     c.EthCLRemoteEndpoint,
		"bootnodes":       c.Bootnodes,
		"gossip-topics":   c.GossipTopics,
		"subnets":         c.Subnets,
		"persist-msgs":    c.PersistMsgs,
		"val-pubkeys":     len(c.ValPubkeys),
	}).Info("config for the Ethereum crawler")
}
