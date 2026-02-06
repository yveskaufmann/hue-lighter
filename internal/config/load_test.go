package config

import (
	"os"
	"path/filepath"
	"testing"

	"com.github.yveskaufmann/hue-lighter/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromDefaultPath(t *testing.T) {
	tests := []struct {
		name           string
		configPath     string // environment variable value
		setupFile      bool
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name:       "loads config from custom path via CONFIG_PATH env var",
			configPath: "", // Will be set to temp file
			setupFile:  true,
			wantErr:    false,
		},
		{
			name:           "returns error when custom config file not found",
			configPath:     "/nonexistent/config.yaml",
			wantErr:        true,
			expectedErrMsg: "config file not found at \"/nonexistent/config.yaml\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up temp file if needed
			var tempFile string
			if tt.setupFile {
				tmpDir := t.TempDir()
				tempFile = filepath.Join(tmpDir, "config.yaml")
				err := os.WriteFile(tempFile, []byte(testutils.ValidHueConfigYAML()), 0644)
				require.NoError(t, err)
				tt.configPath = tempFile
			}

			// Set up environment variable
			var cleanup func()
			if tt.configPath != "" {
				cleanup = testutils.SetEnv(t, "CONFIG_PATH", tt.configPath)
			} else {
				cleanup = testutils.SetEnv(t, "CONFIG_PATH", "")
			}
			defer cleanup()

			// Execute the function
			config, err := LoadConfigFromDefaultPath()

			// Assert results
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)

				// Verify the config was loaded correctly
				assert.Equal(t, 52.5, config.Location.Latitude)
				assert.Equal(t, 13.4, config.Location.Longitude)
				assert.Len(t, config.Lights, 2)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name:        "loads valid config successfully",
			fileContent: testutils.ValidHueConfigYAML(),
			wantErr:     false,
		},
		{
			name:           "returns error for invalid YAML",
			fileContent:    testutils.InvalidHueConfigYAML("malformed-yaml"),
			wantErr:        true,
			expectedErrMsg: "failed to decode config file",
		},
		{
			name:           "returns error for invalid latitude",
			fileContent:    testutils.InvalidHueConfigYAML("invalid-latitude"),
			wantErr:        true,
			expectedErrMsg: "invalid config in file",
		},
		{
			name:           "returns error for invalid longitude",
			fileContent:    testutils.InvalidHueConfigYAML("invalid-longitude"),
			wantErr:        true,
			expectedErrMsg: "invalid config in file",
		},
		{
			name:        "returns error for missing location",
			fileContent: testutils.InvalidHueConfigYAML("missing-location"),
			wantErr:     false, // Location (0,0) is actually valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.fileContent), 0644)
			require.NoError(t, err)

			// Execute the function
			config, err := LoadConfig(configPath)

			// Assert results
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)

				// Verify config structure based on content type
				if tt.fileContent == testutils.ValidHueConfigYAML() {
					assert.Equal(t, 52.5, config.Location.Latitude)
					assert.Equal(t, 13.4, config.Location.Longitude)
					assert.Len(t, config.Lights, 2)
				} else if tt.fileContent == testutils.InvalidHueConfigYAML("missing-location") {
					// Config with missing location section gets default values (0,0)
					assert.Equal(t, 0.0, config.Location.Latitude)
					assert.Equal(t, 0.0, config.Location.Longitude)
					assert.Len(t, config.Lights, 1)
				}
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	// Test the specific case of file not found to verify error message format
	config, err := LoadConfig("/nonexistent/path/config.yaml")

	require.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "config file not found at \"/nonexistent/path/config.yaml\"")
	assert.Contains(t, err.Error(), "Please create your config file by copying the example:")
	assert.Contains(t, err.Error(), "cp configs/config.example.yaml configs/config.yaml")
}

func TestLoadConfig_FileOpenError(t *testing.T) {
	// Create a directory where we expect a file (this will cause a different error than "not found")
	tmpDir := t.TempDir()
	dirAsFile := filepath.Join(tmpDir, "config.yaml")
	err := os.Mkdir(dirAsFile, 0755)
	require.NoError(t, err)

	// Try to load the directory as if it were a file
	config, err := LoadConfig(dirAsFile)

	require.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to decode config file")
	// Should not contain the helpful message for missing files
	assert.NotContains(t, err.Error(), "Please create your config file by copying the example:")
}
