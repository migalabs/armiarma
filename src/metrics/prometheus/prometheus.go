package prometheus

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/push"
    "github.com/eminetto/clean-architecture-go/config"
)


func NewPrometheusService(path string, port string) (*Service, error) {
	
    

}