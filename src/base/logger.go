package base

import(
	"io"
	log "github.com/sirupsen/logrus"

)

// App default configurations
var(
	DefaultLoglvl = logrus.InfoLevel
	DefaultLogOutput  = os.Stdout
	DefaultFormater   = &log.TextFormater{}
)

// LogOpts defines the basic struct to start a new logger
type LogOpts struct {
	ModName string
	Output string
	Formatter strng
	Level string
}

// NewDefaultLogger generates a simple Module Logger
func NewDefaultLogger() log.Logger {
	logger := log.WithFields(log.Fields{
			"module": "Module",
		})
	logger.SetFormatter(DefaultFormater)
	logger.SetOutput(DefaultLogOutput)
	logger.SetLevel(DefaultLoglvl)
	return logger
}

// Select Log Level from string
func ParseLogLevel(lvl string) log.Formatter {
	switch lvl {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "error":
		return log.ErrorLevel
	default:
		return DefaultLoglvl
	}
}

// parse Formatter from string
func ParseLogOutput(lvl string) log.Formatter {
	switch lvl {
	case "terminal":
		return os.Stdout
	default:
		return DefaultLogOutput
	}
}

// parse Formatter from string
func ParseLogFormatter(lvl string) log.Formatter {
	switch lvl {
	case "text":
		return log.TextFormater{}
	default:
		return DefaultFormater
	}
}
