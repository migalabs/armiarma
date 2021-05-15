package topic

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
    "github.com/protolambda/rumor/metrics/custom"
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
    CustomMetricsPath string `ask:"--custom-metrics-path" help:"The path to the json file where the custom metrics will be exported"`
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
    // Generate the custom Metrics to export in Json at end of execution
    customMetrics := custom.NewCustomMetrics() 
    c.Log.Info("Checking for existing Metrics on Project ...")
    // Check if Previous GossipMetrics were generated
    err, fileExists := c.GossipMetrics.ImportMetrics(c.FilePath)
    if fileExists && err != nil {
        c.Log.Error("Error Importing the metrics from the previous file", err)
    }
    if !fileExists {
        c.Log.Info("Not previous metrics found, generating new ones")
    }
    // Check if Previous ExtraMetrics were generated
    err, fileExists = c.GossipMetrics.ExtraMetrics.ImportMetrics(c.ExtraMetricsPath)
    if fileExists && err != nil {
        c.Log.Error("Error Importing the extra-metrics from the previous file", err)
    }
    if !fileExists {
        c.Log.Info("Not previous extra-metrics found, generating new ones")
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
        h, _ := c.Host()
        // Exporting the CustomMetrics for last time (still don't know which is the best place where to put this call)
        err := FilCustomMetrics(c.GossipMetrics, c.Store, &customMetrics, h)
        if err != nil {
            return err
        }
        // export the CustomMetrics into a json
        err = customMetrics.ExportJson(c.CustomMetricsPath)
        if err != nil {
            return err   
        }
		return nil
	})

	return nil
}


// fulfil the info from the Custom Metrics
func FilCustomMetrics(gm *metrics.GossipMetrics, ps track.ExtendedPeerstore, cm *custom.CustomMetrics, h host.Host) error{
    // TODO: - Generate and do the client version stuff

    // Get total peers in peerstore
    peerstoreLen := custom.TotalPeers(h)
    // get the connection status for each of the peers in the extra-metrics
    succeed, failed, notattempted := gm.ExtraMetrics.GetConnectionMetrics(h)
    // Analyze the reported error by the connection attempts
    resetbypeer, timeout, dialtoself, dialbackoff, uncertain := gm.ExtraMetrics.GetErrorCounter(h)
    // Filter peers on peerstore by port
    x, y, z := custom.GetPeersWithPorts(h, ps)
    // Generate the MetricsDataFrame of the Current Metrics
	mdf := metrics.NewMetricsDataFrame(gm.GossipMetrics)
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
