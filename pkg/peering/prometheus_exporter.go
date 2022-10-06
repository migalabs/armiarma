package peering

import (
	"time"

	promth "github.com/migalabs/armiarma/pkg/exporters"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ServeMetrics:
// This method will serve the global peerstore values to the
// local prometheus instance.
func (c *PeeringService) ServeMetrics() {
	// Generate new ticker
	ticker := time.NewTicker(promth.MetricLoopInterval)
	// register variables
	prometheus.MustRegister(PrunedErrorDistribution)
	prometheus.MustRegister(ErrorAttemptDistribution)
	prometheus.MustRegister(PeersAttemptedInLastIteration)
	prometheus.MustRegister(PeerstoreIterTime)

	// routine to loop
	go func() {
		for {
			select {
			case <-ticker.C:

				iterTime := c.strategy.LastIterTime()
				peersPeriter := c.strategy.AttemptedPeersSinceLastIter()

				controlDist := c.strategy.ControlDistribution()
				cntDist := make(map[string]int)
				errorAttemptDist := c.strategy.GetErrorAttemptDistribution()
				errAttdist := make(map[string]int)

				// get new values
				PeerstoreIterTime.Set(iterTime) // Float in seconds
				PeersAttemptedInLastIteration.Set(float64(peersPeriter))

				// generate the distribution
				controlDist.Range(func(key, value interface{}) bool {
					cntDist[key.(string)] = value.(int)
					PrunedErrorDistribution.WithLabelValues(key.(string)).Set(float64(value.(int)))
					return true
				})

				// generate the distribution
				errorAttemptDist.Range(func(key, value interface{}) bool {
					errAttdist[key.(string)] = value.(int)
					ErrorAttemptDistribution.WithLabelValues(key.(string)).Set(float64(value.(int)))
					return true
				})

				log.WithFields(logrus.Fields{
					"LastIterTime(secs)":          iterTime,
					"AttemptedPeersSinceLastIter": peersPeriter,
					//"IterForcingNextConnTime":         peerIterForcingTime,
					"ControlDistribution":        cntDist,
					"ControlAttemptDistribution": errAttdist,
				}).Info("peering metrics summary")

			case <-c.ctx.Done():
				log.Info("Closing the prometheus metrics export service")
				// closing the routine in a ordened way
				ticker.Stop()
				return
			}
		}
	}()
}
