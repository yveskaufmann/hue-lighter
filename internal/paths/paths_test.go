package paths

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigFile(t *testing.T) {
	t.Run("defaults to the XDG config directory", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/home/test/.config")
		t.Setenv("CONFIG_PATH", "")

		path, err := ConfigFile()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/home/test/.config", "hue-lighter", "config.yaml"), path)
	})

	t.Run("uses CONFIG_PATH when set", func(t *testing.T) {
		t.Setenv("CONFIG_PATH", "/etc/hue-lighter/config.yaml")

		path, err := ConfigFile()
		require.NoError(t, err)
		assert.Equal(t, "/etc/hue-lighter/config.yaml", path)
	})
}

func TestAPIKeyFile(t *testing.T) {
	t.Run("defaults to the XDG config directory", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/home/test/.config")
		t.Setenv("HUE_API_KEY_STORE_PATH", "")

		path, err := APIKeyFile()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join("/home/test/.config", "hue-lighter", "api-keys.json"), path)
	})

	t.Run("uses HUE_API_KEY_STORE_PATH when set", func(t *testing.T) {
		t.Setenv("HUE_API_KEY_STORE_PATH", "/data/api-keys.json")

		path, err := APIKeyFile()
		require.NoError(t, err)
		assert.Equal(t, "/data/api-keys.json", path)
	})
}
