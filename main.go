/*
	Copyright © 2021 Miga Labs
*/
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/migalabs/armiarma/src/utils"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	"github.com/migalabs/armiarma/cmd"
)

var (
	Version = "v1.1.0\n"
	// logging variables
	log = logrus.WithField(
		"module", "ARMIARMA",
	)
)

func main() {
	// read arguments from the command line
	PrintVersion()

	ctx, cancel := context.WithCancel(context.Background())

	// Set the general log configurations for the entire tool
	logrus.SetFormatter(utils.ParseLogFormatter("text"))
	logrus.SetOutput(utils.ParseLogOutput("terminal"))

	app := &cli.App{
		Name:      "armiarma",
		Usage:     "A libp2p DHT crawler, monitor, and measurement tool that exposes timely information about DHT networks.",
		UsageText: "armiarma [commands] [arguments...]",
		Authors: []*cli.Author{
			{
				Name:  "Miga Labs",
				Email: "migalabs@protonmail.com",
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			cmd.Eth2CrawlerCommand,
			cmd.IpfsCrawlerCommand,
		},
	}

	// generate the crawler
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Errorf("error: %v\n", err)
		os.Exit(1)
	}

	// check the shutdown signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)

	// keep the app running until syscall.SIGTERM
	sig := <-sigs
	log.Printf("Received %s signal - Stopping...\n", sig.String())
	signal.Stop(sigs)
	cancel()

}

func PrintVersion() {
	fmt.Println("Armirma_" + Version)
}
