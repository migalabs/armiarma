package base

import(
	"os"
	"io"
	log "github.com/sirupsen/logrus"

)

// App default configurations
var(
	DefaultLoglvl = log.InfoLevel
	DefaultLogOutput  = os.Stdout
	DefaultFormater   = &log.TextFormatter{}
)

// LogOpts defines the basic struct to start a new logger
type LogOpts struct {
	ModName string
	Output string
	Formatter string
	Level string
}

// NewDefaultLogger generates a simple Module Logger
func NewDefaultLogger() log.FieldLogger {
	logger := log.New()
	logger.SetFormatter(DefaultFormater)
	logger.SetOutput(DefaultLogOutput)
	logger.SetLevel(DefaultLoglvl)
	l := logger.WithField("module", "default-module")
	return l
}

// Select Log Level from string
func ParseLogLevel(lvl string) log.Level {
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
func ParseLogOutput(lvl string) io.Writer {
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
		return &log.TextFormatter{}
	default:
		return DefaultFormater
	}
}
