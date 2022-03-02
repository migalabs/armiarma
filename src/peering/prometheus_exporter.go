package peering

import (
	"sync/atomic"
	"time"

	promth "github.com/migalabs/armiarma/src/exporters"
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
				cntDist := make(map[string]int64)
				errorAttemptDist := c.strategy.GetErrorAttemptDistribution()
				errAttdist := make(map[string]int64)

				// get new values
				PeerstoreIterTime.Set(iterTime) // Float in seconds
				PeersAttemptedInLastIteration.Set(float64(peersPeriter))

				// generate the distribution
				for key, value := range controlDist {
					cntDist[key] = atomic.LoadInt64(value)
					PrunedErrorDistribution.WithLabelValues(key).Set(float64(atomic.LoadInt64(value)))
				}
				// generate the distribution
				for key, value := range errorAttemptDist {
					errAttdist[key] = atomic.LoadInt64(value)
					ErrorAttemptDistribution.WithLabelValues(key).Set(float64(atomic.LoadInt64(value)))
				}

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
