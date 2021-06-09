package prometheus

import (
	"fmt"

	"github.com/protolambda/rumor/control/actor/base"
	"github.com/protolambda/rumor/metrics"
)

type PrometheusCmd struct {
	*base.Base
	*metrics.GossipMetrics
}

func (c *PrometheusCmd) Cmd(route string) (cmd interface{}, err error) {
	switch route {
	case "start":
		cmd = &PrometheusStartCmd{Base: c.Base, GossipMetrics: c.GossipMetrics}
	default:
		return nil, fmt.Errorf("no command was assigned")
	}
	return cmd, nil
}

func (c *PrometheusCmd) Routes() []string {
	return []string{"start"}
}

func (c *PrometheusCmd) Help() string {
	return "Manage the Prometheus interaction with the crawler"
}
