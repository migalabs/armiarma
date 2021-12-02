package utils

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// App default configurations
var (
	DefaultLoglvl    = logrus.InfoLevel
	DefaultLogOutput = os.Stdout
	DefaultFormater  = &logrus.TextFormatter{}
)

// Select Log Level from string
func ParseLogLevel(lvl string) logrus.Level {
	switch lvl {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
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
func ParseLogFormatter(lvl string) logrus.Formatter {
	switch lvl {
	case "text":
		return &logrus.TextFormatter{}
	default:
		return DefaultFormater
	}
}

/*
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
*/
