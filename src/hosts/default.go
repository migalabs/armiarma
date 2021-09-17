package hosts

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Include all the host default configurations
var (
	DefaultIP       string = "0.0.0.0"
	DefaultTCP      string = "12345"
	DefaultUDP      string = "12345"
	DefaulUserAgent string = "BSC-Armiarma-Crawler"
	// Related to the Log Level
	DevelopmentLogLvl = logrus.DebugLevel
	ProductionLogLvl  = logrus.ErrorLevel
	DefaultLogOutput  = os.Stdout
	DefaultFormater   = logrus.TextFormater{}
)
