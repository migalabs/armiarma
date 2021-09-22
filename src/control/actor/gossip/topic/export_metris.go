package topic

import (
	"context"

	"github.com/pkg/errors"

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
