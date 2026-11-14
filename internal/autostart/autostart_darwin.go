package autostart

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// Enable writes a LaunchAgent that starts hue-lighter on login, if running
// from an app bundle. The agent is not loaded right away, since the app is
// already running; launchd picks it up on the next login.
func Enable(_ context.Context, logger *log.Entry) {
	logger = logger.WithField("component", "Autostart")

	executable, err := os.Executable()
	if err == nil {
		executable, err = filepath.EvalSymlinks(executable)
	}
	if err != nil {
		logger.WithError(err).Warn("Failed to determine executable, skipping autostart registration")
		return
	}

	if !isAppBundleExecutable(executable) {
		logger.Debug("Not running from an app bundle, skipping autostart registration")
		return
	}

	if err := writeLaunchAgent(executable, logger); err != nil {
		logger.WithError(err).Warn("Failed to register autostart")
	}
}

func writeLaunchAgent(executable string, logger *log.Entry) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine home directory: %w", err)
	}

	agentPath := filepath.Join(home, "Library", "LaunchAgents", LaunchAgentLabel+".plist")
	logFile := filepath.Join(home, "Library", "Logs", "hue-lighter.log")
	content := launchAgentPlist(executable, logFile)

	if existing, err := os.ReadFile(agentPath); err == nil && bytes.Equal(existing, content) {
		logger.Debug("Autostart LaunchAgent is up to date")
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(agentPath), 0o755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}
	if err := os.WriteFile(agentPath, content, 0o644); err != nil {
		return fmt.Errorf("failed to write LaunchAgent %q: %w", agentPath, err)
	}

	logger.Infof("Autostart on login enabled via LaunchAgent %s", agentPath)
	return nil
}
