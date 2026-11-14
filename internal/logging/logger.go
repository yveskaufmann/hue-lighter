package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func NewLogger() *log.Entry {
	logger := log.New()
	logger.SetOutput(newOutput())
	logger.SetFormatter(newFormatter())
	logger.SetLevel(getLogLevelByEnvironment())
	return log.NewEntry(logger)
}

// newOutput returns the file named by LOG_FILE (appended to, created if
// needed, a leading "~/" is expanded) or stderr if LOG_FILE is unset, empty
// or cannot be opened.
//
// On macOS, LOG_FILE is set by the app bundle (LSEnvironment) and the login
// LaunchAgent, since macOS discards stderr of apps started from Finder.
func newOutput() io.Writer {
	path := strings.TrimSpace(os.Getenv("LOG_FILE"))
	if path == "" {
		return os.Stderr
	}

	file, err := openLogFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file, logging to stderr instead: %v\n", err)
		return os.Stderr
	}
	return file
}

func openLogFile(path string) (*os.File, error) {
	path, err := expandHome(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create log directory for %q: %w", path, err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %q: %w", path, err)
	}
	return file, nil
}

// expandHome replaces a leading "~" with the user's home directory, since
// values from plist files are not expanded by a shell.
func expandHome(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to expand %q: %w", path, err)
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
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
