package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"com.github.yveskaufmann/hue-lighter/internal/paths"
	"gopkg.in/yaml.v3"
)

// LoadConfigFromDefaultPath loads the config from CONFIG_PATH or, if unset,
// from the user's config directory (see paths.ConfigFile).
func LoadConfigFromDefaultPath() (*Config, error) {

	configPath, err := paths.ConfigFile()
	if err != nil {
		return nil, err
	}

	return LoadConfig(configPath)
}

func LoadConfig(path string) (*Config, error) {

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %q\n\n"+
				"Please create your config file by copying the example:\n"+
				"  mkdir -p %q && cp configs/config.example.yaml %q\n"+
				"Then edit it with your location and light IDs.\n"+
				"See README.md for detailed setup instructions", path, filepath.Dir(path), path)
		}
		return nil, fmt.Errorf("failed to open config file %q: %w", path, err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file %q: %w", path, err)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid config in file %q: %w", path, err)
	}

	return &config, nil
}

func (c *Config) validate() error {
	if c == nil {
		return errors.New("config is nil")
	}

	if c.Location.Latitude < -90 || c.Location.Latitude > 90 ||
		c.Location.Longitude < -180 || c.Location.Longitude > 180 {
		return errors.New("invalid location coordinates")
	}

	for _, light := range c.Lights {
		if light.ID == nil && light.Name == nil {
			return errors.New("light must have either ID or Name")
		}
	}

	return nil
}
