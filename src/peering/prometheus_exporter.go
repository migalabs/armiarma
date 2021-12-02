package peering

import (
	"time"

	promth "github.com/migalabs/armiarma/src/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ServeMetrics
// * This method will serve the global peerstore values to the
// * local prometheus instance
func (c *PeeringService) ServeMetrics() {
	// Generate new ticker
	ticker := time.NewTicker(promth.MetricLoopInterval)
	// register variables
	prometheus.MustRegister(PrunedErrorDistribution)
	prometheus.MustRegister(ErrorAttemptDistribution)
	prometheus.MustRegister(PeersAttemptedInLastIteration)
	prometheus.MustRegister(PeerstoreIterTime)
	//prometheus.MustRegister(IterForcingNextConnTime)

	// routine to loop
	go func() {
		for {
			select {
			case <-ticker.C:

				iterTime := c.strategy.LastIterTime()
				peersPeriter := c.strategy.AttemptedPeersSinceLastIter()
				//peerIterForcingTime := c.strategy.IterForcingNextConnTime()
				controlDist := c.strategy.ControlDistribution()
				errorAttemptDist := c.strategy.GetErrorAttemptDistribution()
				// get new values
				PeerstoreIterTime.Set(iterTime) // Float in seconds
				PeersAttemptedInLastIteration.Set(float64(peersPeriter))
				//IterForcingNextConnTime.Set(peerIterForcingTime)

				// generate the distribution
				for key, value := range controlDist {
					PrunedErrorDistribution.WithLabelValues(key).Set(float64(value))
				}
				// generate the distribution
				for key, value := range errorAttemptDist {
					ErrorAttemptDistribution.WithLabelValues(key).Set(float64(value))
				}

				Log.WithFields(logrus.Fields{
					"LastIterTime(secs)":          iterTime,
					"AttemptedPeersSinceLastIter": peersPeriter,
					//"IterForcingNextConnTime":         peerIterForcingTime,
					"ControlDistribution":        controlDist,
					"ControlAttemptDistribution": errorAttemptDist,
				}).Info("peering metrics summary")

			case <-c.ctx.Done():
				Log.Info("Closing the prometheus metrics export service")
				// closing the routine in a ordened way
				ticker.Stop()
				return
			}
		}
	}()
}
