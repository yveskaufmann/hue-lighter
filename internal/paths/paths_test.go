package paths

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeUserConfigDir points the environment variables os.UserConfigDir relies
// on to a temporary directory and returns the config directory expected on
// the current OS.
func fakeUserConfigDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	switch runtime.GOOS {
	case "windows":
		t.Setenv("AppData", root)
		return root
	case "darwin", "ios":
		t.Setenv("HOME", root)
		return filepath.Join(root, "Library", "Application Support")
	case "plan9":
		t.Setenv("home", root)
		return filepath.Join(root, "lib")
	default:
		configHome := filepath.Join(root, ".config")
		t.Setenv("XDG_CONFIG_HOME", configHome)
		return configHome
	}
}

func TestConfigDir(t *testing.T) {
	t.Run("defaults to the user config directory", func(t *testing.T) {
		userConfigDir := fakeUserConfigDir(t)

		dir, err := ConfigDir()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(userConfigDir, "hue-lighter"), dir)
	})
}

func TestConfigFile(t *testing.T) {
	t.Run("defaults to the user config directory", func(t *testing.T) {
		userConfigDir := fakeUserConfigDir(t)
		t.Setenv("CONFIG_PATH", "")

		path, err := ConfigFile()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(userConfigDir, "hue-lighter", "config.yaml"), path)
	})

	t.Run("uses CONFIG_PATH when set", func(t *testing.T) {
		want := filepath.Join(t.TempDir(), "config.yaml")
		t.Setenv("CONFIG_PATH", want)

		path, err := ConfigFile()
		require.NoError(t, err)
		assert.Equal(t, want, path)
	})
}

func TestAPIKeyFile(t *testing.T) {
	t.Run("defaults to the user config directory", func(t *testing.T) {
		userConfigDir := fakeUserConfigDir(t)
		t.Setenv("HUE_API_KEY_STORE_PATH", "")

		path, err := APIKeyFile()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(userConfigDir, "hue-lighter", "api-keys.json"), path)
	})

	t.Run("uses HUE_API_KEY_STORE_PATH when set", func(t *testing.T) {
		want := filepath.Join(t.TempDir(), "api-keys.json")
		t.Setenv("HUE_API_KEY_STORE_PATH", want)

		path, err := APIKeyFile()
		require.NoError(t, err)
		assert.Equal(t, want, path)
	})
}
