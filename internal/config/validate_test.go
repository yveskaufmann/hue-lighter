package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config is nil",
		},
		{
			name: "valid config with valid coordinates",
			config: &Config{
				Location: struct {
					Latitude  float64 `yaml:"latitude"`
					Longitude float64 `yaml:"longitude"`
				}{
					Latitude:  52.5,
					Longitude: 13.4,
				},
				Lights: []struct {
					ID   *string `yaml:"id"`
					Name *string `yaml:"name"`
				}{
					{ID: stringPtr("light-1")},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with edge case coordinates",
			config: &Config{
				Location: struct {
					Latitude  float64 `yaml:"latitude"`
					Longitude float64 `yaml:"longitude"`
				}{
					Latitude:  90.0,
					Longitude: 180.0,
				},
				Lights: []struct {
					ID   *string `yaml:"id"`
					Name *string `yaml:"name"`
				}{
					{Name: stringPtr("test-light")},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with negative edge case coordinates",
			config: &Config{
				Location: struct {
					Latitude  float64 `yaml:"latitude"`
					Longitude float64 `yaml:"longitude"`
				}{
					Latitude:  -90.0,
					Longitude: -180.0,
				},
				Lights: []struct {
					ID   *string `yaml:"id"`
					Name *string `yaml:"name"`
				}{
					{ID: stringPtr("light-1"), Name: stringPtr("light-name")},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid latitude too high",
			config: &Config{
				Location: struct {
					Latitude  float64 `yaml:"latitude"`
					Longitude float64 `yaml:"longitude"`
				}{
					Latitude:  91.0,
					Longitude: 0.0,
				},
			},
			wantErr: true,
			errMsg:  "invalid location coordinates",
		},
		{
			name: "invalid latitude too low",
			config: &Config{
				Location: struct {
					Latitude  float64 `yaml:"latitude"`
					Longitude float64 `yaml:"longitude"`
				}{
					Latitude:  -91.0,
					Longitude: 0.0,
				},
			},
			wantErr: true,
			errMsg:  "invalid location coordinates",
		},
		{
			name: "invalid longitude too high",
			config: &Config{
				Location: struct {
					Latitude  float64 `yaml:"latitude"`
					Longitude float64 `yaml:"longitude"`
				}{
					Latitude:  0.0,
					Longitude: 181.0,
				},
			},
			wantErr: true,
			errMsg:  "invalid location coordinates",
		},
		{
			name: "invalid longitude too low",
			config: &Config{
				Location: struct {
					Latitude  float64 `yaml:"latitude"`
					Longitude float64 `yaml:"longitude"`
				}{
					Latitude:  0.0,
					Longitude: -181.0,
				},
			},
			wantErr: true,
			errMsg:  "invalid location coordinates",
		},
		{
			name: "light with neither ID nor name",
			config: &Config{
				Location: struct {
					Latitude  float64 `yaml:"latitude"`
					Longitude float64 `yaml:"longitude"`
				}{
					Latitude:  52.5,
					Longitude: 13.4,
				},
				Lights: []struct {
					ID   *string `yaml:"id"`
					Name *string `yaml:"name"`
				}{
					{}, // Neither ID nor Name set
				},
			},
			wantErr: true,
			errMsg:  "light must have either ID or Name",
		},
		{
			name: "valid config with multiple lights",
			config: &Config{
				Location: struct {
					Latitude  float64 `yaml:"latitude"`
					Longitude float64 `yaml:"longitude"`
				}{
					Latitude:  52.5,
					Longitude: 13.4,
				},
				Lights: []struct {
					ID   *string `yaml:"id"`
					Name *string `yaml:"name"`
				}{
					{ID: stringPtr("light-1")},
					{Name: stringPtr("light-2")},
					{ID: stringPtr("light-3"), Name: stringPtr("light-3-name")},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with empty lights array",
			config: &Config{
				Location: struct {
					Latitude  float64 `yaml:"latitude"`
					Longitude float64 `yaml:"longitude"`
				}{
					Latitude:  52.5,
					Longitude: 13.4,
				},
				Lights: []struct {
					ID   *string `yaml:"id"`
					Name *string `yaml:"name"`
				}{},
			},
			wantErr: false,
		},
		{
			name: "mixed valid and invalid lights",
			config: &Config{
				Location: struct {
					Latitude  float64 `yaml:"latitude"`
					Longitude float64 `yaml:"longitude"`
				}{
					Latitude:  52.5,
					Longitude: 13.4,
				},
				Lights: []struct {
					ID   *string `yaml:"id"`
					Name *string `yaml:"name"`
				}{
					{ID: stringPtr("light-1")},
					{}, // Invalid light
				},
			},
			wantErr: true,
			errMsg:  "light must have either ID or Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create string pointers for testing
func stringPtr(s string) *string {
	return &s
}
