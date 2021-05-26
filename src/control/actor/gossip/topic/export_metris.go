package topic

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/metrics/custom"
	"github.com/protolambda/rumor/metrics/export"
	"github.com/protolambda/rumor/p2p/track"
)

type TopicExportMetricsCmd struct {
	*base.Base
	GossipMetrics     *metrics.GossipMetrics
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
	c.BackupPeriod = 30 * time.Minute
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
	// Generate the custom Metrics to export in Json at end of execution
	customMetrics := custom.NewCustomMetrics()
	c.Log.Info("Checking for existing Metrics on Project ...")
	// Check if Previous GossipMetrics were generated
	err, fileExists := c.GossipMetrics.ImportMetrics(c.RawFilePath)
	if fileExists && err != nil {
		c.Log.Error("Error Importing the metrics from the previous file", err)
	}
	if !fileExists {
		c.Log.Info("Not previous metrics found, generating new ones")
	}
	c.Log.Infof("Exporting Every %d , with a backup every %d", c.ExportPeriod, c.BackupPeriod)
	stopping := false
	go func() {
		t := time.Now()
		c.UpdateFilesAndFolders(t)

		// loop to export the metrics every Backup and Period time
		for {
			if stopping {
				_ = c.GossipMetrics.ExportMetrics(c.RawFilePath, c.PeerstorePath, c.CsvPath, c.Store)
				c.Log.Infof("Metrics Export Stopped")
				h, _ := c.Host()
				// Exporting the CustomMetrics for last time (still don't know which is the best place where to put this call)
				err := FilCustomMetrics(c.GossipMetrics, c.Store, &customMetrics, h)
				if err != nil {
					fmt.Println(err)
					return
				}
				// export the CustomMetrics into a json
				err = customMetrics.ExportJson(c.CustomMetricsPath)
				if err != nil {
					fmt.Println(err)
					return
				}
				return
			}

			start := time.Now()
			c.Log.Infof("Backup Export of the Metrics")
			c.ExportSecuence(t, &customMetrics)

			// Check Backup period to wait for next round
			exportStepDuration := time.Since(start)
			if exportStepDuration < c.BackupPeriod {
				time.Sleep(c.ExportPeriod - exportStepDuration)
			}
			// Check if the Export Period has been accomplished (generate new forlde for the metrics)
			tnow := time.Since(t)
			if tnow >= c.ExportPeriod {
				c.Log.Infof("Exporting Metrics changing to Folder")
				t = time.Now()
				c.UpdateFilesAndFolders(t)
				c.ExportSecuence(t, &customMetrics)
				c.GossipMetrics.ResetDynamicMetrics()
			}
		}
	}()
	c.Control.RegisterStop(func(ctx context.Context) error {
		stopping = true
		c.Log.Infof("Stoped Exporting")
		return nil
	})

	return nil
}

// fulfil the info from the Custom Metrics
func FilCustomMetrics(gm *metrics.GossipMetrics, ps track.ExtendedPeerstore, cm *custom.CustomMetrics, h host.Host) error {
	// TODO: - Generate and do the client version stuff

	// Get total peers in peerstore
	peerstoreLen := custom.TotalPeers(h)
	// get the connection status for each of the peers in the extra-metrics
	succeed, failed, notattempted := gm.GetConnectionMetrics(h)
	// Analyze the reported error by the connection attempts
	resetbypeer, timeout, dialtoself, dialbackoff, uncertain := gm.GetErrorCounter(h)
	// Filter peers on peerstore by port
	x, y, z := custom.GetPeersWithPorts(h, ps)
	// Generate the MetricsDataFrame of the Current Metrics
	mdf := export.NewMetricsDataFrame(gm.GossipMetrics)
	lig := mdf.AnalyzeClientType("Lighthouse")
	tek := mdf.AnalyzeClientType("Teku")
	nim := mdf.AnalyzeClientType("Nimbus")
	pry := mdf.AnalyzeClientType("Prysm")
	lod := mdf.AnalyzeClientType("Lodestar")
	unk := mdf.AnalyzeClientType("Unknown")

	// read client versions from Metrics
	cm.PeerStore.SetTotal(peerstoreLen)
	cm.PeerStore.SetPort13000(x)
	cm.PeerStore.SetPort9000(y)
	cm.PeerStore.SetPortDiff(z)
	cm.PeerStore.SetNotAttempted(notattempted)
	cm.PeerStore.ConnectionFailed.SetTotal(failed)
	cm.PeerStore.ConnectionFailed.SetResetByPeer(resetbypeer)
	cm.PeerStore.ConnectionFailed.SetTimeOut(timeout)
	cm.PeerStore.ConnectionFailed.SetDialToSelf(dialtoself)
	cm.PeerStore.ConnectionFailed.SetDialBackOff(dialbackoff)
	cm.PeerStore.ConnectionFailed.SetUncertain(uncertain)

	// fill the CustomMetrics with the readed information
	cm.PeerStore.ConnectionSucceed.SetTotal(succeed)
	cm.PeerStore.ConnectionSucceed.Lighthouse = lig
	cm.PeerStore.ConnectionSucceed.Teku = tek
	cm.PeerStore.ConnectionSucceed.Nimbus = nim
	cm.PeerStore.ConnectionSucceed.Prysm = pry
	cm.PeerStore.ConnectionSucceed.Lodestar = lod
	cm.PeerStore.ConnectionSucceed.Unknown = unk

	return nil
}

func (c *TopicExportMetricsCmd) UpdateFilesAndFolders(t time.Time) {
	year := strconv.Itoa(t.Year())
	month := t.Month().String()
	day := strconv.Itoa(t.Day())
	hour := strconv.Itoa(t.Hour())
	m := t.Minute()
	var minute string
	if m < 10 {
		minute = "0" + strconv.Itoa(m)
	} else {
		minute = strconv.Itoa(m)
	}
	folderName := c.MetricsFolderPath + "/" + "metrics" + "/" + year + "-" + month + "-" + day + "-" + hour + ":" + minute
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
}

func (c *TopicExportMetricsCmd) ExportSecuence(start time.Time, cm *custom.CustomMetrics) {
	// Export The metrics
	c.GossipMetrics.FillMetrics(c.Store)
	err := c.GossipMetrics.ExportMetrics(c.RawFilePath, c.PeerstorePath, c.CsvPath, c.Store)
	if err != nil {
		c.Log.Infof("Problems exporting the Metrics to the given file path")
	} else {
		ed := time.Since(start)
		log := "Metrics Exported, time to export:" + ed.String()
		c.Log.Infof(log)
	}
	// Export the Custom metrics
	h, _ := c.Host()
	// Exporting the CustomMetrics for last time (still don't know which is the best place where to put this call)
	err = FilCustomMetrics(c.GossipMetrics, c.Store, cm, h)
	if err != nil {
		c.Log.Warn(err)
	}
	// export the CustomMetrics into a json
	err = cm.ExportJson(c.CustomMetricsPath)
	if err != nil {
		c.Log.Warn(err)
	}
}
