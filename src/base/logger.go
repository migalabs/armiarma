package base

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

// App default configurations
var (
	DefaultLoglvl    = log.InfoLevel
	DefaultLogOutput = os.Stdout
	DefaultFormater  = &log.TextFormatter{}
)

// LogOpts defines the basic struct to start a new logger
type LogOpts struct {
	ModName   string
	Output    string
	Formatter string
	Level     string
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

func CreateLogger(opts LogOpts) log.FieldLogger {
	logger := log.New()
	logger.SetFormatter(ParseLogFormatter(opts.Formatter))
	logger.SetOutput(ParseLogOutput(opts.Output))
	logger.SetLevel(ParseLogLevel(opts.Level))
	l := logger.WithField("module", opts.ModName)

	return l
}

func CreateLogOpts(input_mod_name string, input_output string, input_formatter string, input_level string) LogOpts {
	return LogOpts{
		ModName:   input_mod_name,
		Output:    input_output,
		Formatter: input_formatter,
		Level:     input_level,
	}
}

func CreateStdLogOpts(input_output string, input_formatter string) LogOpts {
	return LogOpts{
		Output:    input_output,
		Formatter: input_formatter,
	}
}
