package topic

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	//"github.com/libp2p/go-libp2p-core/host"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/metrics/prometheus"
	"github.com/protolambda/rumor/metrics/utils"
	"github.com/protolambda/rumor/p2p/track"
	log "github.com/sirupsen/logrus"
)

// TODO: This will be shared by both armiarma-client and server, so move to /metrics

type TopicExportMetricsCmd struct {
	*base.Base
	PeerStore         *metrics.PeerStore
	GossipState       *metrics.GossipState
	Store             track.ExtendedPeerstore
	ExportPeriod      time.Duration `ask:"--export-period" help:"Requets the frecuency in witch the Metrics will be exported to the files"`
	BackupPeriod      time.Duration `ask:"--backup-period" help:"Requets the frecuency in witch the Backup of the Metrics will be exported"`
	MetricsFolderPath string        `ask:"--metrics-folder-path" help:"The path of the folder where to export the metrics."`
	RawFilePath       string
	PeerstorePath     string
	CsvPath           string
	CustomMetricsPath string
}

func (c *TopicExportMetricsCmd) Defaul() {
	c.ExportPeriod = 24 * time.Hour
	c.BackupPeriod = 10 * time.Second
	c.RawFilePath = "gossip-metrics.json"
	c.CustomMetricsPath = "custom-metrics.json"
	c.PeerstorePath = "peerstore.json"
	c.CsvPath = "metrics.csv"
}

func (c *TopicExportMetricsCmd) Help() string {
	return "Exports the Gossip Metrics to the given file path"
}

func (c *TopicExportMetricsCmd) Run(ctx context.Context, args ...string) error {
	if c.GossipState.GsNode == nil {
		return NoGossipErr
	}

	// TODO: Placing this here as a quick solution.
	prometheusRunner := prometheus.NewPrometheusRunner(c.PeerStore)
	err := prometheusRunner.Run(context.Background())
	if err != nil {
		log.Fatal("TODO Error starting prometheus")
	}

	// Generate the custom Metrics to export in Json at end of execution
	//customMetrics := custom.NewCustomMetrics()
	c.Log.Info("Checking for existing Metrics on Project ...")
	// Check if Previous PeerStore were generated
	err, fileExists := c.PeerStore.ImportMetrics(c.MetricsFolderPath)
	if fileExists && err != nil {
		c.Log.Error("Error Importing the metrics from the previous file", err)
	}
	if !fileExists {
		c.Log.Info("Not previous metrics found, generating new ones")
	}
	c.Log.Infof("Exporting Every %d , with a backup every %d", c.ExportPeriod, c.BackupPeriod)
	fmt.Println("Exporting Every ", c.ExportPeriod, " with a backup every", c.BackupPeriod)
	stopping := false
	go func() {
		t := time.Now()
		fmt.Println("Initial time:", t)
		c.UpdateFilesAndFolders(t)

		// loop to export the metrics every Backup and Period time
		for {
			if stopping {
				_ = c.PeerStore.ExportMetrics(c.RawFilePath, c.PeerstorePath, c.CsvPath, c.Store)
				c.Log.Infof("Metrics Export Stopped")
				return
			}

			log.Info("Backup Export of the Metrics")
			for _, peerId := range c.Store.Peers() {
				peerData := c.Store.GetAllData(peerId)
				peer := fetchPeerExtraInfo(peerData)
				c.PeerStore.AddPeer(peer)
			}
			err := c.PeerStore.ExportMetrics(c.RawFilePath, c.PeerstorePath, c.CsvPath, c.Store)
			if err != nil {
				fmt.Println("ERROR TODO:")
			}

			c.UpdateFilesAndFolders(t)
			c.PeerStore.ResetDynamicMetrics()
			// Force Memmory Free from the Garbage Collector
			debug.FreeOSMemory()

			time.Sleep(c.BackupPeriod)
		}
	}()
	c.Control.RegisterStop(func(ctx context.Context) error {
		stopping = true
		c.Log.Infof("Stoped Exporting")
		return nil
	})

	return nil
}

// Convert from rumor PeerAllData to our Peer. Note that
// some external data is fetched and some fields are parsed
func fetchPeerExtraInfo(peerData *track.PeerAllData) metrics.Peer {
	client, version := utils.FilterClientType(peerData.UserAgent)
	address := utils.GetFullAddress(peerData.Addrs)

	ip, country, city, err := utils.GetIpAndLocationFromAddrs(address)
	if err != nil {
		log.Error("error when fetching country/city from ip", err)
	}

	peer := metrics.Peer {
		PeerId: peerData.PeerID.String(),
		NodeId: peerData.NodeID.String(),
		UserAgent: peerData.UserAgent,
		ClientName: client,
		ClientVersion: version,
		ClientOS: "TODO",
		Pubkey: peerData.Pubkey,
		Addrs: address,
		Ip: ip,
		Country: country,
		City: city,
		Latency: float64(peerData.Latency/time.Millisecond) / 1000,
	}

	return peer
}

func (c *TopicExportMetricsCmd) UpdateFilesAndFolders(t time.Time) {
	year := strconv.Itoa(t.Year())
	month := t.Month().String()
	day := strconv.Itoa(t.Day())
	h := t.Hour()
	var hour string
	if h < 10 {
		hour = "0" + strconv.Itoa(h)
	} else {
		hour = strconv.Itoa(h)
	}
	m := t.Minute()
	var minute string
	if m < 10 {
		minute = "0" + strconv.Itoa(m)
	} else {
		minute = strconv.Itoa(m)
	}
	date := year + "-" + month + "-" + day + "-" + hour + ":" + minute
	folderName := c.MetricsFolderPath + "/" + "metrics" + "/" + date
	// generate new metrics folder
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		c.Log.Warnf("making folder:", folderName)
		os.Mkdir(folderName, 0770)
	}
	// complete file path
	c.RawFilePath = folderName + "/" + "gossip-metrics.json"
	c.CustomMetricsPath = folderName + "/" + "custom-metrics.json"
	c.PeerstorePath = folderName + "/" + "peerstore.json"
	c.CsvPath = folderName + "/" + "metrics.csv"
	c.Log.Warnf("New exporting folder:", folderName)

	// Update the last checkpoint
	err := GenCheckpointFile(c.MetricsFolderPath, date)
	if err != nil {
		c.Log.Warn(err)
		fmt.Println(err)
	}
}

// Function that writes in a file the folder name of the last checkpoint generated in the project
// DOUBT: Write path relative or absolute? dunno
func GenCheckpointFile(cpPath string, lastCP string) error {
	cp := metrics.Checkpoint{
		Checkpoint: lastCP,
	}
	cpFile := cpPath + "/metrics/checkpoint-folder.json"
	fmt.Println("Checkpoint File:", cpFile)
	jb, err := json.Marshal(cp)
	if err != nil {
		fmt.Println("Error Marshalling last Checkpoint")
		return err
	}
	err = ioutil.WriteFile(cpFile, jb, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", cpFile)
		return err
	}
	return nil
}
