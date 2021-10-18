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
	"context"
	"os"
	"os/signal"

	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/discovery"
	"github.com/migalabs/armiarma/src/enode"
	"github.com/migalabs/armiarma/src/gossipsub"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// crawler status containing the main basemodule and info that the app will ConnectedF
type CrawlerBase struct {
	*base.Base
	Host *hosts.BasicLibp2pHost
	Node *enode.LocalNode
	Dv5  *discovery.Discovery
	Gs   *gossipsub.GossipSub
	Info *info.InfoData
}

// variable to be used as a flag from command line
var inputconfigFile string

// crawlerCmd represents the crawler command
var crawlerCmd = &cobra.Command{
	Use:   "crawler",
	Short: "Launch the Network Crawler on the given network",
	Long:  `Launch the Network Crawler on the given network`,
	Run: func(cmd *cobra.Command, args []string) {
		mainCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// TODO: just hardcoded
		logOpts := base.LogOpts{
			ModName:   "crawler app",
			Output:    "terminal",
			Formatter: "text",
			Level:     "debug",
		}
		// generate a base for the crawler app
		b, err := base.NewBase(
			base.WithContext(mainCtx),
			base.WithLogger(logOpts),
		)
		if err != nil {
			log.Panic(err)
		}

		stdOpts := base.LogOpts{
			Output:    "terminal",
			Formatter: "text",
		}

		info_tmp := info.NewCustomInfoData(inputconfigFile, stdOpts)
		stdOpts.Level = info_tmp.GetLogLevel()

		// TODO: just harcoded
		baseOpts := base.LogOpts{
			ModName:   "libp2p host",
			Output:    "terminal",
			Formatter: "text",
			Level:     info_tmp.GetLogLevel(),
		}

		hostOpts := hosts.BasicLibp2pHostOpts{
			Info_obj: *info_tmp,
			LogOpts:  baseOpts,
		}
		// generate libp2pHost
		host, err := hosts.NewBasicLibp2pHost(b.Ctx(), hostOpts)
		if err != nil {
			log.Panic(err)
		}

		node_tmp := enode.NewLocalNode(b.Ctx(), info_tmp, stdOpts)
		//node_tmp.AddEntries()
		dv5_tmp := discovery.NewDiscovery(b.Ctx(), node_tmp, info_tmp, 9006, stdOpts)
		gs_tmp := gossipsub.NewGossipSub(b.Ctx(), *host, stdOpts)
		// generate the CrawlerBase
		crawler := CrawlerBase{
			Base: b,
			Host: host,
			Info: info_tmp,
			Node: node_tmp,
			Dv5:  dv5_tmp,
			Gs:   gs_tmp,
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

	crawlerCmd.Flags().StringVar(&inputconfigFile, "config-file", "", "Set the configuration file to import")
}

// generate new CrawlerBase
func (c *CrawlerBase) InitCrawler() error {
	// initialization secuence for the crawler

	c.Host.Start()
	c.Dv5.Start_dv5()
	go c.Dv5.FindRandomNodes(*c.Host)

	c.Gs.JoinAndSubscribe("/eth2/b5303f2a/beacon_block/ssz_snappy")

	select {}
	// return nil
}