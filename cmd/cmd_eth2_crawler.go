/*
	Copyright © 2021 Miga Labs
*/
package cmd

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	"github.com/migalabs/armiarma/pkg/crawler"
	"github.com/migalabs/armiarma/pkg/info"
)

// CrawlCommand contains the crawl sub-command configuration.
var Eth2CrawlerCommand = &cli.Command{
	Name:   "eth2",
	Usage:  "crawl the eth2 network with the given configuration in the conf-file",
	Action: LaunchEth2Crawler,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "config-file",
			Usage:       "path to the <config.json> file used to configure the crawler",
			EnvVars:     []string{"ARMIARMA_CONFIG_FILE_NAME"},
			DefaultText: info.DefaultEth2ConfigFile,
			Value:       info.DefaultEth2ConfigFile,
		}},
}

// CrawlAction is the function that is called when running `eth2`.
func LaunchEth2Crawler(c *cli.Context) error {
	log.Infoln("Starting Eth2 crawler...")

	// Load configuration file
	infObj, err := info.InitEth2(c)
	if err != nil {
		return err
	}

	// Generate the Eth2 crawler struct
	eth2c, err := crawler.NewEth2Crawler(c, infObj)
	if err != nil {
		return err
	}

	// launch the subroutines
	eth2c.Run()

	// check the shutdown signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)

	// keep the app running until syscall.SIGTERM
	sig := <-sigs
	log.Printf("Received %s signal - Stopping...\n", sig.String())
	signal.Stop(sigs)
	eth2c.Close()

	return nil
}
