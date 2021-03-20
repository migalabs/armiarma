package plotter

import (
    "github.com/protolambda/rumor/control/actor/base"
    "github.com/protolambda/rumor/metrics/"
)

type MetricsPlotterCmd struct {
    *base.Base
    GossipMetrics *metrics.GossipMetrics

    // Host related info
    Port        string  `ask:"--port" help:"Listening port where to visualize the metrics"`
    TimeInterv  int     `ask:"--refresh-interval" help:"Time interval when the plots will be updated with the lastest metrics (hours)"`
}

func (c *MetricsPlotterCmd) Help() string{
    return "Start the metrics visualizer on the specified port number (Web Browser accessible)"
}

func (c *MetricsPlotterCmd) Default() {
    c.Port = "8080"
    c.TimeInterv = 1 // every 1 hour
}

func (c *MetricsPlotterCmd) Run() error{
    fmt.Println("Ploter")
    // check if the host has been already initialized

    // generate the metrics dataframe into the proper Plotting format

    // initialize the host as the http request handlers

    // start a loop that given certain time, will update the metrics dataframe with the newest metrics

    // check untill the exit signal to stop both, the host and the loop/go routine
}

