// Package paths resolves the locations of hue-lighter's per-user files.
//
// Files live in the user's XDG config directory (~/.config/hue-lighter). When
// running as Flatpak, XDG_CONFIG_HOME points into the sandbox, so the same
// code resolves to ~/.var/app/<app-id>/config/hue-lighter. Environment
// variables override each location, e.g. for containers.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const appDirName = "hue-lighter"

// ConfigDir returns the directory holding hue-lighter's per-user files.
func ConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine user config directory: %w", err)
	}
	return filepath.Join(userConfigDir, appDirName), nil
}

// ConfigFile returns the path of the config file, overridable via CONFIG_PATH.
func ConfigFile() (string, error) {
	return resolve("CONFIG_PATH", "config.yaml")
}

// APIKeyFile returns the path of the file based API key store, overridable via
// HUE_API_KEY_STORE_PATH.
func APIKeyFile() (string, error) {
	return resolve("HUE_API_KEY_STORE_PATH", "api-keys.json")
}

func resolve(envName string, fileName string) (string, error) {
	if path := os.Getenv(envName); path != "" {
		return path, nil
	}

	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}
