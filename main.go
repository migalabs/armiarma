/*
Copyright © 2021 Miga Labs
*/
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/migalabs/armiarma/pkg/utils"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	"github.com/migalabs/armiarma/cmd"
)

var (
	Version = "v2.0.0\n"
	// logging variables
	log = logrus.WithField(
		"module", "ARMIARMA",
	)
)

func main() {
	// read arguments from the command line
	PrintVersion()

	// Set the general log configurations for the entire tool
	logrus.SetFormatter(utils.ParseLogFormatter("text"))
	logrus.SetOutput(utils.ParseLogOutput("terminal"))

	app := &cli.App{
		Name:      "armiarma",
		Usage:     "Distributed libp2p crawler that monitors, measures, and exposes the gathered information about libp2p network's overlays.",
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
			// cmd.IpfsCrawlerCommand,
		},
	}

	// generate the crawler
	if err := app.RunContext(context.Background(), os.Args); err != nil {
		log.Errorf("error: %v\n", err)
		os.Exit(1)
	}
}

func helpInArgs(args []string) bool {
	help := false
	for _, b := range args {
		switch b {
		case "--help":
			help = true
			break
		case "-h":
			help = true
			break
		case "h":
			help = true
			break
		case "help":
			help = true
			break
		}
	}
	return help
}

func PrintVersion() {
	fmt.Println("Armiarma_" + Version)
}
