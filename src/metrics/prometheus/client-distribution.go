package prometheus

import (
	"fmt"
	//"time"
	"context"
	"net/http"

	"github.com/protolambda/rumor/control/actor/base"

	"github.com/prometheus/client_golang/prometheus"
)

// TODO Merge the prometheus metrics that are somewhere else here


clientDistribution = prometheus.NewHistogramVec(prometheus.HistogramOpts {
		Name: "crawler_observed_client_distribution",
		Help: "The client distribution observed of the peers with we tried to exchange the peer metadata.",
		Buckets: prometheus.LinearBuckets(20, 5, 5),  // 5 buckets, each 5 centigrade wide. (I still don't really get this)
	},
	[]string{"channel"}, nil,
   )
