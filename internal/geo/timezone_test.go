package geo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZoneLocation(t *testing.T) {
	tests := []struct {
		zone      string
		latitude  float64
		longitude float64
	}{
		{zone: "Europe/Berlin", latitude: 52.5, longitude: 13.3667},
		{zone: "America/New_York", latitude: 40.7142, longitude: -74.0064},
		{zone: "Australia/Sydney", latitude: -33.8667, longitude: 151.2167},
	}

	for _, tt := range tests {
		t.Run(tt.zone, func(t *testing.T) {
			location, err := zoneLocation(tt.zone)
			require.NoError(t, err)
			location = rounded(location)
			assert.Equal(t, tt.latitude, location.Latitude)
			assert.Equal(t, tt.longitude, location.Longitude)
			assert.Equal(t, SourceTimeZone, location.Source)
		})
	}

	t.Run("unknown zone", func(t *testing.T) {
		_, err := zoneLocation("Mars/Olympus_Mons")
		require.Error(t, err)
	})
}

func TestSystemTimeZone(t *testing.T) {
	t.Run("uses TZ when set", func(t *testing.T) {
		zone, err := systemTimeZone(":Europe/Paris", "/nonexistent")
		require.NoError(t, err)
		assert.Equal(t, "Europe/Paris", zone)
	})

	t.Run("resolves the localtime symlink", func(t *testing.T) {
		dir := t.TempDir()
		zoneFile := filepath.Join(dir, "zoneinfo", "Europe", "Berlin")
		require.NoError(t, os.MkdirAll(filepath.Dir(zoneFile), 0755))
		require.NoError(t, os.WriteFile(zoneFile, nil, 0644))
		localtime := filepath.Join(dir, "localtime")
		require.NoError(t, os.Symlink(zoneFile, localtime))

		zone, err := systemTimeZone("", localtime)
		require.NoError(t, err)
		assert.Equal(t, "Europe/Berlin", zone)
	})

	t.Run("fails when localtime is no zoneinfo symlink", func(t *testing.T) {
		localtime := filepath.Join(t.TempDir(), "localtime")
		require.NoError(t, os.WriteFile(localtime, nil, 0644))

		_, err := systemTimeZone("", localtime)
		require.Error(t, err)
	})
}
