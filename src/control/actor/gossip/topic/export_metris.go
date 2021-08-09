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
	"github.com/protolambda/rumor/p2p/track"
	log "github.com/sirupsen/logrus"
)

// TODO: This will be shared by both armiarma-client and server, so move to /metrics

type TopicExportMetricsCmd struct {
	*base.Base
	PeerStore     *metrics.PeerStore
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
				//h, _ := c.Host()
				// Exporting the CustomMetrics for last time (still don't know which is the best place where to put this call)
				/*
					err := FilCustomMetrics(c.PeerStore, c.Store, &customMetrics, h)
					if err != nil {
						fmt.Println(err)
						return
					}
					// export the CustomMetrics into a json
					err = customMetrics.ExportJson(c.CustomMetricsPath)
					if err != nil {
						fmt.Println(err)
						return
					}*/
				return
			}

			start := time.Now()
			c.Log.Infof("Backup Export of the Metrics")
			//c.ExportSecuence(t, &customMetrics)
			c.PeerStore.FillMetrics(c.Store)
			err := c.PeerStore.ExportMetrics(c.RawFilePath, c.PeerstorePath, c.CsvPath, c.Store)
			if err != nil {
				fmt.Println("ERROR TODO:")
			}

			// Check Backup period to wait for next round
			exportStepDuration := time.Since(start)
			if exportStepDuration < c.BackupPeriod {
				fmt.Println("Waiting to run new backup export")
				wt := c.BackupPeriod - exportStepDuration
				fmt.Println("Waiting time:", wt)
				time.Sleep(wt)
			}
			// Check if the Export Period has been accomplished (generate new folder for the metrics)
			tnow := time.Since(t)
			if tnow >= c.ExportPeriod {
				c.Log.Infof("Exporting Metrics changing to Folder")
				t = time.Now()
				c.UpdateFilesAndFolders(t)
				//c.ExportSecuence(t, &customMetrics)
				c.PeerStore.ResetDynamicMetrics()
				// Force Memmory Free from the Garbage Collector
				debug.FreeOSMemory()
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

//func FilCustomMetrics(gm *metrics.PeerStore, ps track.ExtendedPeerstore, cm *custom.CustomMetrics, h host.Host) error {
// Get total peers in peerstore
//peerstoreLen := custom.TotalPeers(h)
// get the connection status for each of the peers in the extra-metrics
//succeed, failed, notattempted := gm.GetConnectionMetrics(h)
// Analyze the reported error by the connection attempts
//resetbypeer, timeout, dialtoself, dialbackoff, uncertain := gm.GetErrorCounter(h)
// Filter peers on peerstore by port
//x, y, z := custom.GetPeersWithPorts(h, ps)
// Generate the MetricsDataFrame of the Current Metrics
//mdf := export.NewMetricsDataFrame(&gm.PeerStore)

//_ = mdf
/*
	lig := mdf.AnalyzeClientType("Lighthouse")
	tek := mdf.AnalyzeClientType("Teku")
	nim := mdf.AnalyzeClientType("Nimbus")
	pry := mdf.AnalyzeClientType("Prysm")
	lod := mdf.AnalyzeClientType("Lodestar")
	unk := mdf.AnalyzeClientType("Unknown")*/
/*
	lig := custom.NewClient()
	tek := custom.NewClient()
	nim := custom.NewClient()
	pry := custom.NewClient()
  lod := custom.NewClient()
	unk := custom.NewClient()
*/

// read client versions from Metrics
/*
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
	cm.PeerStore.ConnectionFailed.SetUncertain(uncertain)*/

// fill the CustomMetrics with the readed information
/*
	cm.PeerStore.ConnectionSucceed.SetTotal(succeed)
	cm.PeerStore.ConnectionSucceed.Lighthouse = lig
	cm.PeerStore.ConnectionSucceed.Teku = tek
	cm.PeerStore.ConnectionSucceed.Nimbus = nim
	cm.PeerStore.ConnectionSucceed.Prysm = pry
	cm.PeerStore.ConnectionSucceed.Lodestar = lod
	cm.PeerStore.ConnectionSucceed.Unknown = unk*/

// fill the json with client distribution from those peers we got the metadata request from
/*
	mtlig := mdf.AnalyzeClientTypeIfMetadataRequested("Lighthouse")
	mttek := mdf.AnalyzeClientTypeIfMetadataRequested("Teku")
	mtnim := mdf.AnalyzeClientTypeIfMetadataRequested("Nimbus")
	mtpry := mdf.AnalyzeClientTypeIfMetadataRequested("Prysm")
	mtlod := mdf.AnalyzeClientTypeIfMetadataRequested("Lodestar")
	mtunk := mdf.AnalyzeClientTypeIfMetadataRequested("Unknown")
*/
/*
	mtlig := custom.NewClient()
	mttek := custom.NewClient()
	mtnim := custom.NewClient()
	mtpry := custom.NewClient()
	mtlod := custom.NewClient()
	mtunk := custom.NewClient()



	tot := mtlig.Total + mttek.Total + mtnim.Total + mtpry.Total + mtlod.Total + mtunk.Total

	// fill the CustomMetrics with the readed information
	cm.PeerStore.MetadataRequested.SetTotal(tot)
	cm.PeerStore.MetadataRequested.Lighthouse = mtlig
	cm.PeerStore.MetadataRequested.Teku = mttek
	cm.PeerStore.MetadataRequested.Nimbus = mtnim
	cm.PeerStore.MetadataRequested.Prysm = mtpry
	cm.PeerStore.MetadataRequested.Lodestar = mtlod
	cm.PeerStore.MetadataRequested.Unknown = mtunk
*/

//	return nil
//}

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

/*
func (c *TopicExportMetricsCmd) ExportSecuence(start time.Time, cm *custom.CustomMetrics) {
	// Export The metrics
	fmt.Println("exporting metrics")
	c.PeerStore.FillMetrics(c.Store)
	err := c.PeerStore.ExportMetrics(c.RawFilePath, c.PeerstorePath, c.CsvPath, c.Store)
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
	err = FilCustomMetrics(c.PeerStore, c.Store, cm, h)
	if err != nil {
		c.Log.Warn(err)
	}
	// export the CustomMetrics into a json
	err = cm.ExportJson(c.CustomMetricsPath)
	if err != nil {
		c.Log.Warn(err)
	}
}*/

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
