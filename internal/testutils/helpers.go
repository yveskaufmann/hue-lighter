package testutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockFileSystem creates an in-memory filesystem for testing
func MockFileSystem(t *testing.T) afero.Fs {
	return afero.NewMemMapFs()
}

// CreateTempFile creates a temporary file in the mock filesystem with content
func CreateTempFile(t *testing.T, fs afero.Fs, path, content string) {
	err := afero.WriteFile(fs, path, []byte(content), 0644)
	require.NoError(t, err)
}

// MockHTTPResponse creates a mock HTTP response for testing
func MockHTTPResponse(statusCode int, body interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if body != nil {
			switch v := body.(type) {
			case string:
				w.Write([]byte(v))
			default:
				json.NewEncoder(w).Encode(v)
			}
		}
	}))
}

// MockHueBridgeResponse creates a mock Hue Bridge API response
func MockHueBridgeResponse(statusCode int, data interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if data != nil {
			json.NewEncoder(w).Encode(data)
		}
	}))
}

// MockHueErrorResponse creates a mock Hue Bridge error response
func MockHueErrorResponse(errorType, description string) *httptest.Server {
	errorResponse := []map[string]interface{}{
		{
			"error": map[string]interface{}{
				"type":        1,
				"address":     "/",
				"description": description,
			},
		},
	}
	return MockHueBridgeResponse(400, errorResponse)
}

// SetEnv sets environment variable and returns cleanup function
func SetEnv(t *testing.T, key, value string) func() {
	original := os.Getenv(key)
	os.Setenv(key, value)
	return func() {
		if original == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, original)
		}
	}
}

// AssertErrorContains checks that error contains expected message
func AssertErrorContains(t *testing.T, err error, expectedMessage string) {
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedMessage)
}

// AssertNoError is a convenience wrapper for assert.NoError
func AssertNoError(t *testing.T, err error) {
	assert.NoError(t, err)
}

// FixedTimeProvider provides a fixed time for testing time-dependent code
type FixedTimeProvider struct {
	fixedTime time.Time
}

func NewFixedTimeProvider(fixedTime time.Time) *FixedTimeProvider {
	return &FixedTimeProvider{fixedTime: fixedTime}
}

func (f *FixedTimeProvider) Now() time.Time {
	return f.fixedTime
}

// ValidHueConfig returns a valid configuration for testing
func ValidHueConfig() map[string]interface{} {
	return map[string]interface{}{
		"location": map[string]interface{}{
			"latitude":  52.5,
			"longitude": 13.4,
		},
		"lights": []map[string]interface{}{
			{
				"id":   "light-1",
				"name": "Test Light 1",
			},
			{
				"id":   "light-2",
				"name": "Test Light 2",
			},
		},
	}
}

// ValidHueConfigYAML returns a valid YAML config string
func ValidHueConfigYAML() string {
	return `location:
  latitude: 52.5
  longitude: 13.4
lights:
  - id: "light-1"
    name: "Test Light 1"
  - id: "light-2"
    name: "Test Light 2"`
}

// InvalidHueConfigYAML returns invalid YAML config for testing error cases
func InvalidHueConfigYAML(errorType string) string {
	switch errorType {
	case "invalid-latitude":
		return `location:
  latitude: 91.0
  longitude: 13.4
lights: []`
	case "invalid-longitude":
		return `location:
  latitude: 52.5
  longitude: 181.0
lights: []`
	case "missing-location":
		return `lights:
  - id: "light-1"
    name: "Test Light"`
	case "malformed-yaml":
		return `location:
  latitude: 52.5
  longitude: [invalid`
	default:
		return ""
	}
}

// MockAPIKeyStore creates a mock API key store for testing
func MockAPIKeyStore(fs afero.Fs) string {
	apiKeyData := `{
  "bridge-id-1": {
    "username": "mock-api-key-1",
    "clientkey": "mock-client-key-1"
  }
}`
	path := "/tmp/test-api-keys.json"
	err := afero.WriteFile(fs, path, []byte(apiKeyData), 0644)
	if err != nil {
		panic(err)
	}
	return path
}
