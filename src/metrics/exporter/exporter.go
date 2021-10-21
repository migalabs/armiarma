package exporter

import (
	"context"
	"time"

	"github.com/protolambda/rumor/metrics"

	log "github.com/sirupsen/logrus"
)

type Exporter struct {
	PeerStore      *metrics.PeerStore
	CsvFileName    string
	ExportInterval time.Duration
}

func NewExporter(gm *metrics.PeerStore) Exporter {
	return Exporter{
		PeerStore:      gm,
		CsvFileName:    "metrics.csv",
		ExportInterval: 10 * time.Minute,
	}
}

func (c *Exporter) Run(ctx context.Context) error {

	go func() {
		for {
			err := c.PeerStore.ExportToCSV(c.CsvFileName)
			if err != nil {
				log.Error("could not export peerstore to csv: ", err)
			}

			time.Sleep(c.ExportInterval)
		}
	}()

	return nil
}