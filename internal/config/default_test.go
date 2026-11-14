package config

import (
	"os"
	"path/filepath"
	"testing"

	"com.github.yveskaufmann/hue-lighter/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault_IsValid(t *testing.T) {
	cfg := Default()
	require.NoError(t, cfg.validate())
	assert.Empty(t, cfg.Lights)
	assert.NotEmpty(t, cfg.Meta.Name)
}

func TestLoadOrCreateDefault(t *testing.T) {
	t.Run("creates the default config if there is none", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "hue-lighter", "config.yaml")
		t.Setenv("CONFIG_PATH", configPath)

		cfg, created, err := LoadOrCreateDefault()
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, Default(), cfg)

		saved, err := LoadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, Default().Meta, saved.Meta)
		assert.Equal(t, Default().Location, saved.Location)
		assert.Empty(t, saved.Lights)

		info, err := os.Stat(configPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})

	t.Run("loads an existing config without overwriting it", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.yaml")
		content := []byte(testutils.ValidHueConfigYAML())
		require.NoError(t, os.WriteFile(configPath, content, 0644))
		t.Setenv("CONFIG_PATH", configPath)

		cfg, created, err := LoadOrCreateDefault()
		require.NoError(t, err)
		assert.False(t, created)
		assert.NotEmpty(t, cfg.Lights)

		after, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Equal(t, content, after)
	})

	t.Run("fails on an invalid config instead of replacing it", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.yaml")
		content := []byte("location:\n  latitude: 123\n  longitude: 0\n")
		require.NoError(t, os.WriteFile(configPath, content, 0644))
		t.Setenv("CONFIG_PATH", configPath)

		_, created, err := LoadOrCreateDefault()
		require.Error(t, err)
		assert.False(t, created)

		after, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Equal(t, content, after)
	})
}
