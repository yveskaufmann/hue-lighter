package logging

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func NewLogger() *log.Entry {
	logger := log.New()
	logger.SetFormatter(newFormatter())
	logger.SetLevel(getLogLevelByEnvironment())
	return log.NewEntry(logger)
}

func getLogLevelByEnvironment() log.Level {
	defaultLevel := log.InfoLevel
	parsedLevel := defaultLevel

	if lvlFromEnv, ok := os.LookupEnv("LOG_LEVEL"); ok {
		level, err := log.ParseLevel(strings.ToLower(lvlFromEnv))
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid log level '%s', defaulting to %s\n", lvlFromEnv, defaultLevel.String())
			parsedLevel = defaultLevel
		} else {
			parsedLevel = level
		}
	}

	return parsedLevel
}

func newFormatter() log.Formatter {
	defaultFormatType := "text"

	formatType, ok := os.LookupEnv("LOG_FORMAT")
	if !ok {
		formatType = defaultFormatType
	}

	formatType = strings.TrimSpace(formatType)

	switch formatType {
	case "json":
	case "text":
	default:
		fmt.Fprintf(os.Stderr, "invalid log format '%s', defaulting to %s\n", formatType, defaultFormatType)
		formatType = defaultFormatType
	}

	if formatType == "json" {
		return &log.JSONFormatter{}
	}

	return &log.TextFormatter{
		FullTimestamp: true,
	}
}
