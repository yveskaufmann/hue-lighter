package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSave_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hue-lighter", "config.yaml")

	config := &Config{
		Location: Location{Latitude: 52.5609, Longitude: 13.5105},
		Lights: []Light{
			{ID: stringPtr("408429bc-b135-4d4a-9e32-c877a0608425"), Name: stringPtr("Arbeitszimmer Hue Play L")},
			{ID: stringPtr("8f7c4128-5031-4c08-ad8b-b0a33f405f97"), Name: stringPtr("Arbeitszimmer Hue Play R")},
		},
	}
	config.Meta.Version = "1"
	config.Meta.Name = "Hue Lighter Automation"

	require.NoError(t, Save(path, config))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	loaded, err := LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, config, loaded)

	// No temporary files are left behind.
	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestSave_RejectsInvalidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	err := Save(path, &Config{Location: Location{Latitude: 91}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location coordinates")
	assert.NoFileExists(t, path)
}

func TestConfig_Clone(t *testing.T) {
	config := &Config{Lights: []Light{{ID: stringPtr("light-1"), Name: stringPtr("Light 1")}}}

	clone := config.Clone()
	*clone.Lights[0].Name = "Changed"
	clone.Lights = append(clone.Lights, Light{ID: stringPtr("light-2")})

	assert.Equal(t, "Light 1", *config.Lights[0].Name)
	assert.Len(t, config.Lights, 1)
}
