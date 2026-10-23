package config

import (
	"fmt"
	"os"
	"path/filepath"

	"com.github.yveskaufmann/hue-lighter/internal/paths"
	"gopkg.in/yaml.v3"
)

// SaveToDefaultPath saves the config to CONFIG_PATH or, if unset, to the
// user's config directory (see paths.ConfigFile).
func SaveToDefaultPath(config *Config) error {
	configPath, err := paths.ConfigFile()
	if err != nil {
		return err
	}
	return Save(configPath, config)
}

// Save validates the config and writes it atomically to path. Comments of a
// hand-edited config file are not preserved.
func Save(path string, config *Config) error {
	if err := config.validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory %q: %w", dir, err)
	}

	// Write to a temporary file first, so a crash never leaves a half written config.
	tmpFile, err := os.CreateTemp(dir, ".config-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary config file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write config: %w", err)
	}
	if err := tmpFile.Chmod(0600); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to set config permissions: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), path); err != nil {
		return fmt.Errorf("failed to replace config %q: %w", path, err)
	}
	return nil
}
