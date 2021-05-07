package topic

import (
	"context"
	"time"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
	"github.com/protolambda/rumor/p2p/track"
)

type TopicExportMetricsCmd struct {
	*base.Base
	GossipMetrics   *metrics.GossipMetrics
    GossipState     *metrics.GossipState
    Store           track.ExtendedPeerstore
    ExportPeriod    time.Duration `ask:"--export-period" help:"Requets the frecuency in witch the Metrics will be exported to the files"`
	FilePath        string `ask:"--file-path" help:"The path of the file where to export the metrics."`
	PeerstorePath   string `ask:"--peerstore-path" help:"The path of the file where to export the peerstore."`
    CsvPath         string `ask:"--csv-file" help:"The path where the csv file will be exported"`
    ExtraMetricsPath string `ask:"--extra-metrics-path" help:"The path to the csv file where the extra metrics will be exported"`
}

func (c *TopicExportMetricsCmd) Defaul() {
    c.ExportPeriod = 60 * time.Second
}

func (c *TopicExportMetricsCmd) Help() string {
	return "Exports the Gossip Metrics to the given file path"
}

func (c *TopicExportMetricsCmd) Run(ctx context.Context, args ...string) error {
    if c.GossipState.GsNode == nil {
        return NoGossipErr
    }
    c.Log.Info("Checking for existing Metrics on Project ...")
    err, fileExists := c.GossipMetrics.ImportMetrics(c.FilePath)
    if fileExists && err != nil {
        c.Log.Error("Error Importing the metrics from the previous file", err)
    }
    if !fileExists {
        c.Log.Info("Not previous metrics found, generating new ones")
    }
    stopping := false
	go func() {
		for {
            if stopping {
                _ =  c.GossipMetrics.ExportMetrics(c.FilePath, c.PeerstorePath, c.CsvPath, c.ExtraMetricsPath, c.Store)
                c.Log.Infof("Metrics Export Stopped")
                return
            }
			start := time.Now()
            c.Log.Infof("Exporting Metrics")
            c.GossipMetrics.FillMetrics(c.Store)
	        err := c.GossipMetrics.ExportMetrics(c.FilePath, c.PeerstorePath, c.CsvPath, c.ExtraMetricsPath, c.Store)
            if err != nil {
                c.Log.Infof("Problems exporting the Metrics to the given file path")
            } else {
                ed := time.Since(start)
                log := "Metrics Exported, time to export:" + ed.String()
                c.Log.Infof(log)
            }
            exportStepDuration := time.Since(start)
			if exportStepDuration < c.ExportPeriod{
				time.Sleep(c.ExportPeriod - exportStepDuration)
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
