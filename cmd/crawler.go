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
	"time"

	"github.com/migalabs/armiarma/src/base"
	"github.com/migalabs/armiarma/src/db"
	"github.com/migalabs/armiarma/src/discovery"
	"github.com/migalabs/armiarma/src/enode"
	"github.com/migalabs/armiarma/src/gossipsub"
	"github.com/migalabs/armiarma/src/gossipsub/blockchaintopics"
	"github.com/migalabs/armiarma/src/hosts"
	"github.com/migalabs/armiarma/src/info"
	"github.com/migalabs/armiarma/src/peering"
	"github.com/migalabs/armiarma/src/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// crawler status containing the main basemodule and info that the app will ConnectedF
type CrawlerBase struct {
	*base.Base
	Host             *hosts.BasicLibp2pHost
	Node             *enode.LocalNode
	DB               *db.PeerStore
	Dv5              *discovery.Discovery
	Peering          peering.PeeringService
	Gs               *gossipsub.GossipSub
	Info             *info.InfoData
	PrometheusRunner *prometheus.PrometheusRunner
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

		// TODO: generate a new DB
		db := db.NewPeerStore(info_tmp.GetDBType(), info_tmp.GetDBPath())
		// Generate a Peering Service (so far with default peering strategy)

		hostOpts := hosts.BasicLibp2pHostOpts{
			Info_obj:  *info_tmp,
			LogOpts:   baseOpts,
			PeerStore: &db,
		}

		// generate libp2pHost
		host, err := hosts.NewBasicLibp2pHost(b.Ctx(), hostOpts)
		if err != nil {
			log.Panic(err)
		}

		node_tmp := enode.NewLocalNode(b.Ctx(), info_tmp, stdOpts)
		//node_tmp.AddEntries()
		dv5_tmp := discovery.NewDiscovery(b.Ctx(), node_tmp, &db, info_tmp, 9006, stdOpts)
		gs_tmp := gossipsub.NewGossipSub(b.Ctx(), host, &db, stdOpts)
		// Generate the PeeringService
		peeringOpts := &peering.PeeringOpts{
			InfoObj: info_tmp,
			LogOpts: stdOpts,
		}
		// generate the peering strategy
		prunOpts := peering.PruningOpts{
			AggregatedDelay: 24 * time.Hour, // Hardcoded, still using the Default Delay
			LogOpts:         stdOpts,
		}

		pStrategy, err := peering.NewPruningStrategy(b.Ctx(), &db, prunOpts)
		peeringServ, err := peering.NewPeeringService(b.Ctx(), host, &db, peeringOpts,
			peering.WithPeeringStrategy(&pStrategy),
		)

		prometheusRunner := prometheus.NewPrometheusRunner(&db)
		prometheusRunner.Start(b.Ctx())

		// generate the CrawlerBase
		crawler := CrawlerBase{
			Base:             b,
			Host:             host,
			Info:             info_tmp,
			DB:               &db,
			Node:             node_tmp,
			Dv5:              dv5_tmp,
			Peering:          peeringServ,
			Gs:               gs_tmp,
			PrometheusRunner: &prometheusRunner,
		}

		// Initialization Phase for the crawler
		err = crawler.Run()
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
		// TODO: Shutdown all the services (manually to let them exit in a controled way)
		crawler.Close()
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
func (c *CrawlerBase) Run() error {
	// initialization secuence for the crawler

	c.Host.Start()
	c.Dv5.Start_dv5()
	go c.Dv5.FindRandomNodes()
	go c.Peering.Run()

	topics := blockchaintopics.ReturnAllTopics(c.Info.GetForkDigest())
	for _, topic := range topics {
		c.Gs.JoinAndSubscribe(topic)
	}

	select {}
	// return nil
}

// generate new CrawlerBase
func (c *CrawlerBase) Close() {
	// initialization secuence for the crawler
	c.Log.Info("stoping crawler client")
	c.Host.Stop()
	c.Peering.Close()
}
