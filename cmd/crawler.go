/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/hosts"
)

// crawler status containing the main basemodule and info that the app will ConnectedF
type CrawlerBase struct {
	*base.Base
	Host *hosts.BasicLibp2pHost
}


// crawlerCmd represents the crawler command
var crawlerCmd = &cobra.Command{
	Use:   "crawler",
	Short: "Launch the Network Crawler on the given network",
	Long: `Launch the Network Crawler on the given network`,
	Run: func(cmd *cobra.Command, args []string) {
		mainCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// TODO: just hardcoded
		logOpts := base.LogOpts{
			ModName: "crawler app",
			Output: "terminal",
			Formatter: "text",
			Level: "debug",
		}
		// generata a base for the crawler app
		base, err := base.NewBase(
			base.WithContext(),
			base.WithLogger(logOpts)
		)
		if err != nil {
			log.Panic(err)
		}
		// TODO: just harcoded
		hostOpts := BasicLibp2pHostOpts{
			IP: "127.0.0.1",
			TCP: "9018",
			UDP: "9018",
			UserAgent: "BSC-Armiarma-Crawler",
			PrivKey: "FA44444DE",
			LogOpts: base.LogOpts{
				ModName: "libp2p host",
				Output: "terminal",
				Formatter: "text",
				Level: "debug",
			},
		}
		// generate libp2pHost
		host, err := hosts.NewBasicLibp2pHost(base.Ctx(), hostOpts)
		if err != nil {
			log.Panic(err)
		}

		// generate the CrawlerBase
		crawler := CrawlerBase{
			base,
			Host: host,
		}

		// Initialization Phase for the crawler
		err = crawler.InitCrawler()
		if err != nil {
			crawler.Log.Panic(err)
		}
		// register the shutdown signal
		var signal_channel chan os.Signal
		signal_channel = make(chan os.Signal, 1)
		signal.Notify(signal_channel, os.Interrupt)
		<-signal_channel
		// End up app, finishing everything
		crawler.Log.Info("SHUTDOWN DETECTED!")
		crawler.Host.Stop()
	},
}

func init() {
	rootCmd.AddCommand(crawlerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// crawlerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// crawlerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	var loglvl string
	crawlerCmd.Flags().StringVarP(&loglvl, "log-level", "log-lvl", "debug", "Set the log level of the App")
	fmt.Println("---> parsed flags", loglvl)
}


// generate new CrawlerBase
func (c *CrawlerBase) InitCrawler() error {
	// initialization secuence for the crawler
	c.Host.Start()
	return nil
}
