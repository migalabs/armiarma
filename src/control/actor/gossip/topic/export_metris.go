package topic

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/cortze/rumor/p2p/track"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/metrics/exporter"
	"github.com/protolambda/rumor/metrics/prometheus"
)

type TopicExportMetricsCmd struct {
	*base.Base
	PeerStore *metrics.PeerStore
}

func (c *TopicExportMetricsCmd) Default() {
}

func (c *TopicExportMetricsCmd) Help() string {
	return "Exports the Gossip Metrics to the given file path"
}

func (c *TopicExportMetricsCmd) Run(ctx context.Context, args ...string) error {
	// TODO: Placing this here as a quick solution.
	prometheusRunner := prometheus.NewPrometheusRunner(c.PeerStore)
	err := prometheusRunner.Run(context.Background())
	if err != nil {
		return errors.Wrap(err, "could not start prometheus runner")
	}

	csvExporter := exporter.NewExporter(c.PeerStore)
	err = csvExporter.Run(context.Background())
	if err != nil {
		return errors.Wrap(err, "could not run csv exporter")
	}

	c.Control.RegisterStop(func(ctx context.Context) error {
		c.Log.Infof("Stoped Exporting")
		return nil
	})

	return nil
}

// Convert from rumor PeerAllData to our Peer. Note that
// some external data is fetched and some fields are parsed
func fetchPeerExtraInfo(peerData *track.PeerAllData) metrics.Peer {
	client, version := utils.FilterClientType(peerData.UserAgent)
	address, err := utils.GetFullAddress(peerData.Addrs)
	if err != nil {
		log.Error("error when getting public multiaddress for peer", err)
	}
	ip, country, city, err := utils.GetIpAndLocationFromAddrs(address)
	if err != nil {
		log.Error("error when fetching country/city from ip", err)
	}

	peer := metrics.Peer{
		PeerId:        peerData.PeerID.String(),
		NodeId:        peerData.NodeID.String(),
		UserAgent:     peerData.UserAgent,
		ClientName:    client,
		ClientVersion: version,
		ClientOS:      "TODO",
		Pubkey:        peerData.Pubkey,
		Addrs:         address,
		Ip:            ip,
		Country:       country,
		City:          city,
		Latency:       float64(peerData.Latency/time.Millisecond) / 1000,
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
