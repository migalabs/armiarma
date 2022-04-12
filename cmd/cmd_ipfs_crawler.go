/*
	Copyright © 2021 Miga Labs
*/
package cmd

import (
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	"github.com/migalabs/armiarma/src/crawler"
	"github.com/migalabs/armiarma/src/info"
)

// CrawlCommand contains the crawl sub-command configuration.
var IpfsCrawlerCommand = &cli.Command{
	Name:   "ipfs",
	Usage:  "crawl the ipfs network with the given configuration in the conf-file",
	Action: LaunchIpfsCrawler,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "config-file",
			Usage:       "path to the <config.json> file used to configure the crawler",
			EnvVars:     []string{"ARMIARMA_CONFIG_FILE_NAME"},
			DefaultText: info.DefaultIpfsConfigFile,
			Value:       info.DefaultIpfsConfigFile,
		},
	},
}

// CrawlAction is the function that is called when running `ipfs`.
func LaunchIpfsCrawler(c *cli.Context) error {
	log.Infoln("Starting IPFS crawler...")

	// Load configuration file
	infObj, err := info.InitIpfs(c)
	if err != nil {
		return err
	}

	// Generate the Eth2 crawler struct
	ipfsc, err := crawler.NewIpfsCrawler(c, infObj)
	if err != nil {
		return err
	}

	// launch the subroutines
	ipfsc.Run()
	return nil
}
