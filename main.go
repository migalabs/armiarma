/*
	Copyright © 2021 Miga Labs
*/
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/migalabs/armiarma/cmd"
	"github.com/migalabs/armiarma/src/config"
	"github.com/migalabs/armiarma/src/utils"
	log "github.com/sirupsen/logrus"
)

var (
	Version      = "v0.0.0\n"
	WellcomeText = "Welcome to the Armiarma network monitoring tool."
	SpecifyText  = "List of available flags:"
)

func main() {
	// read arguments from the command line
	PrintVersion()

	// generate new config for the crawler
	crawlerConfig, help := config.NewConfigFromArgs()
	if help {
		CliHelp()
		os.Exit(0)
	}
	// read the log settings from the config

	// Set the general log configurations for the entire tool
	log.SetFormatter(utils.ParseLogFormatter("text"))
	log.SetOutput(utils.ParseLogOutput("terminal"))
	// read the log level from the config-file / default = info
	log.SetLevel(utils.ParseLogLevel(crawlerConfig.GetLogLevel()))

	// generate the crawler
	crawler, err := cmd.NewCrawler(context.Background(), crawlerConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// launch crawler service
	crawler.Run()

	// register the shutdown signal
	signal_channel := make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	<-signal_channel
	// End up app, finishing everything
	log.Info("SHUTDOWN DETECTED!")
	// TODO: Shutdown all the services (manually to let them exit in a controled way)
	crawler.Close()

	os.Exit(0)

}

func CliHelp() {
	fmt.Println(WellcomeText)
	fmt.Println(SpecifyText)
	fmt.Println(cmd.CrawlerHelp())
}

func PrintVersion() {
	fmt.Println("Armirma " + Version)
}
