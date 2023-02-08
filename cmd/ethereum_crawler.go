/*
Copyright © 2021 Miga Labs
*/
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	"github.com/migalabs/armiarma/pkg/config"
	"github.com/migalabs/armiarma/pkg/crawler"
)

// CrawlCommand contains the crawl sub-command configuration.
var Eth2CrawlerCommand = &cli.Command{
	Name:   "eth2",
	Usage:  "crawl the eth2 network with the given configuration in the conf-file",
	Action: LaunchEth2Crawler,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Usage:       "Verbosity level for the Crawler's logs",
			EnvVars:     []string{"ARMIARMA_LOG_LEVEL"},
			DefaultText: config.DefaultLogLevel,
		},
		&cli.StringFlag{
			Name:    "priv-key",
			Usage:   "String representation of the PrivateKey to be used by the crawler",
			EnvVars: []string{"ARMIARMA_PRIV_KEY"},
		},
		&cli.StringFlag{
			Name:        "ip",
			Usage:       "IP in the machine that we want to asign to the crawler",
			EnvVars:     []string{"ARMIARMA_IP"},
			DefaultText: config.DefaultIP,
		},
		&cli.IntFlag{
			Name:        "port",
			Usage:       "TCP and UDP port that the crawler with advertise to establish connections",
			EnvVars:     []string{"ARMIARMA_PORT"},
			DefaultText: fmt.Sprintf("%d", config.DefaultPort),
		},
		&cli.StringFlag{
			Name:        "user-agent",
			Usage:       "Agent name that will identify the crawler in the network",
			EnvVars:     []string{"ARMIARMA_USER_AGENT"},
			DefaultText: config.DefaultUserAgent,
		},
		&cli.StringFlag{
			Name:        "psql-endpoint",
			Usage:       "PSQL enpoint where the crwaler will submit the all the gathered info",
			EnvVars:     []string{"ARMIARMA_PSQL"},
			DefaultText: config.DefaultPSQLEndpoint,
		},
		&cli.StringSliceFlag{
			Name:        "gossip-topic",
			Usage:       "List of gossipsub topics that the crawler will subscribe to",
			EnvVars:     []string{"ARMIARMA_GOSSIP_TOPICS"},
			DefaultText: "One --gossip-topic <topic> per topic",
		},
		&cli.StringFlag{
			Name:        "remote-cl-endpoint",
			Usage:       "Remote Ethereum Consensus Layer Client to request metadata (experimental)",
			EnvVars:     []string{"ARMIARMA_REMOTE_CL_ENDPOINT"},
			DefaultText: config.DefaultCLRemoteEndpoint,
		},
		&cli.StringFlag{
			Name:        "fork-digest",
			Usage:       "Fork Digest of the Ethereum Consensus Layer network that we want to crawl",
			EnvVars:     []string{"ARMIARMA_FORK_DIGEST"},
			DefaultText: config.DefaultMainnetForkDigest,
		},
		&cli.StringSliceFlag{
			Name:        "bootnode",
			Usage:       "List of boondes that the crawler will use to discover more peers in the network",
			EnvVars:     []string{"ARMIARMA_BOOTNODES"},
			DefaultText: "One --bootnode <bootnode> per bootnode",
		},
		&cli.StringFlag{
			Name:        "local-peerstore",
			Usage:       "Path to the local folder that the crawler will use to register the Addrs-Book of discovered peers",
			EnvVars:     []string{"ARMIARMA_LOCAL_PEERSTORE"},
			DefaultText: config.DefaultLocalPeerstorePath,
		},
		&cli.StringSliceFlag{
			Name:        "subnet",
			Usage:       "List of subnets (gossipsub topics) that we want to subscribe the crawler to",
			EnvVars:     []string{"ARMIARMA_SUBNETS"},
			DefaultText: "One --subnet <subnet_id> per subnet",
		},
		&cli.StringFlag{
			Name:        "val-pubkeys",
			Usage:       "Path of the file that has the pubkeys of those validators that we want to track",
			EnvVars:     []string{"ARMIARMA_VAL_PUBKEYS"},
			DefaultText: "./validator_pubkeys.txt",
		},
	},
}

// CrawlAction is the function that is called when running `eth2`.
func LaunchEth2Crawler(c *cli.Context) error {
	log.Infoln("Starting Ethereum Crawler...")

	conf := config.NewEthereumCrawlerConfig()
	conf.Apply(c)

	// Generate the Eth2 crawler struct
	ethCrawler, err := crawler.NewEthereumCrawler(c, *conf)
	if err != nil {
		return err
	}

	// launch the subroutines
	ethCrawler.Run()

	// check the shutdown signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)

	// keep the app running until syscall.SIGTERM
	sig := <-sigs
	log.Printf("Received %s signal - Stopping...\n", sig.String())
	signal.Stop(sigs)
	ethCrawler.Close()

	return nil
}
