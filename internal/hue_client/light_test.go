package hueclient

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"com.github.yveskaufmann/hue-lighter/internal/testutils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test client with mock certificate
func createTestClient(t *testing.T, bridgeID, serverURL string, apiKeyStore APIKeyStore) (*Client, func()) {
	// Generate test certificate
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test CA",
			Organization: []string{"Test CA Org"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	caPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertBytes,
	})

	tmpFile, err := os.CreateTemp("", "test-ca-*.pem")
	require.NoError(t, err)

	_, err = tmpFile.Write(caPEM)
	require.NoError(t, err)
	tmpFile.Close()

	// Strip http:// or https:// from server URL to get just the address
	serverAddr := serverURL
	if strings.HasPrefix(serverAddr, "http://") {
		serverAddr = strings.TrimPrefix(serverAddr, "http://")
	} else if strings.HasPrefix(serverAddr, "https://") {
		serverAddr = strings.TrimPrefix(serverAddr, "https://")
	}

	logger := logrus.New().WithField("test", "light-tests")
	
	// Create client but bypass TLS for testing with HTTP servers
	client := &Client{
		deviceName:  "test-device",
		baseURL:     serverURL, // Use the full URL including protocol
		apiKeyStore: apiKeyStore,
		client:      &http.Client{}, // Use default client for HTTP testing
		bridgeID:    bridgeID,
		logger:      logger,
	}

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	return client, cleanup
}

func TestGetAllLights(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse interface{}
		statusCode     int
		wantErr        bool
		errorMsg       string
		validateResult func(t *testing.T, lights *LightList)
	}{
		{
			name:       "successfully retrieves all lights",
			statusCode: http.StatusOK,
			serverResponse: LightList{
				Data: []LightListItem{
					{
						ID:   "light-1",
						Type: "light",
						On:   LightOnState{On: true},
						Meta: LightMeta{Name: "Living Room Light"},
					},
					{
						ID:   "light-2",
						Type: "light",
						On:   LightOnState{On: false},
						Meta: LightMeta{Name: "Bedroom Light"},
					},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, lights *LightList) {
				require.NotNil(t, lights)
				assert.Len(t, lights.Data, 2)
				assert.Equal(t, "light-1", lights.Data[0].ID)
				assert.Equal(t, "Living Room Light", lights.Data[0].Meta.Name)
				assert.True(t, lights.Data[0].On.On)
				assert.Equal(t, "light-2", lights.Data[1].ID)
				assert.False(t, lights.Data[1].On.On)
			},
		},
		{
			name:       "returns empty list when no lights available",
			statusCode: http.StatusOK,
			serverResponse: LightList{
				Data: []LightListItem{},
			},
			wantErr: false,
			validateResult: func(t *testing.T, lights *LightList) {
				require.NotNil(t, lights)
				assert.Len(t, lights.Data, 0)
			},
		},
		{
			name:           "handles server error response",
			statusCode:     http.StatusInternalServerError,
			serverResponse: `{"error": "internal server error"}`,
			wantErr:        true,
			errorMsg:       "request failed with status code",
		},
		{
			name:           "handles network timeout",
			statusCode:     http.StatusGatewayTimeout,
			serverResponse: `{"error": "timeout"}`,
			wantErr:        true,
			errorMsg:       "request failed with status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testutils.MockHTTPResponse(tt.statusCode, tt.serverResponse)
			defer server.Close()

			apiKeyStore := newMockAPIKeyStore()
			apiKeyStore.Set("test-bridge#test-device", "test-api-key")

			client, cleanup := createTestClient(t, "test-bridge", server.URL, apiKeyStore)
			defer cleanup()

			lights, err := client.GetAllLights()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, lights)
			}
		})
	}
}

func TestGetOneLightById(t *testing.T) {
	tests := []struct {
		name           string
		lightID        string
		serverResponse interface{}
		statusCode     int
		wantErr        bool
		errorMsg       string
		validateResult func(t *testing.T, light *LightListItem)
	}{
		{
			name:       "successfully retrieves light by ID",
			lightID:    "light-123",
			statusCode: http.StatusOK,
			serverResponse: LightList{
				Data: []LightListItem{
					{
						ID:   "light-123",
						Type: "light",
						On:   LightOnState{On: true},
						Meta: LightMeta{Name: "Test Light"},
						Dimming: &LightDimmingState{
							Dimming: 75.5,
						},
					},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, light *LightListItem) {
				require.NotNil(t, light)
				assert.Equal(t, "light-123", light.ID)
				assert.Equal(t, "Test Light", light.Meta.Name)
				assert.True(t, light.On.On)
				assert.NotNil(t, light.Dimming)
				assert.Equal(t, float32(75.5), light.Dimming.Dimming)
			},
		},
		{
			name:       "returns nil when light not found",
			lightID:    "nonexistent-light",
			statusCode: http.StatusOK,
			serverResponse: LightList{
				Data: []LightListItem{},
			},
			wantErr: false,
			validateResult: func(t *testing.T, light *LightListItem) {
				assert.Nil(t, light)
			},
		},
		{
			name:       "returns error when API returns error",
			lightID:    "invalid-light",
			statusCode: http.StatusOK,
			serverResponse: LightList{
				Errors: []struct {
					Description string `json:"description,omitempty"`
				}{
					{Description: "resource not available"},
				},
			},
			wantErr:  true,
			errorMsg: "failed to fetch light",
		},
		{
			name:           "handles server error",
			lightID:        "light-456",
			statusCode:     http.StatusInternalServerError,
			serverResponse: `{"error": "server error"}`,
			wantErr:        true,
			errorMsg:       "request failed with status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testutils.MockHTTPResponse(tt.statusCode, tt.serverResponse)
			defer server.Close()

			apiKeyStore := newMockAPIKeyStore()
			apiKeyStore.Set("test-bridge#test-device", "test-api-key")

			client, cleanup := createTestClient(t, "test-bridge", server.URL, apiKeyStore)
			defer cleanup()

			light, err := client.GetOneLightById(tt.lightID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, light)
			}
		})
	}
}

func TestUpdateOneLightById(t *testing.T) {
	tests := []struct {
		name           string
		lightID        string
		lightUpdate    *LightBodyUpdate
		serverResponse interface{}
		statusCode     int
		wantErr        bool
		errorMsg       string
		validateResult func(t *testing.T, result *ResourceIdentifier)
	}{
		{
			name:    "successfully updates light state",
			lightID: "light-789",
			lightUpdate: &LightBodyUpdate{
				On: &LightOnState{On: true},
			},
			statusCode: http.StatusOK,
			serverResponse: LightUpdateResponse{
				Data: []ResourceIdentifier{
					{
						Action: struct {
							Identity string `json:"identity,omitempty"`
						}{Identity: "light-789"},
					},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *ResourceIdentifier) {
				require.NotNil(t, result)
				assert.Equal(t, "light-789", result.Action.Identity)
			},
		},
		{
			name:    "successfully updates light brightness",
			lightID: "light-456",
			lightUpdate: &LightBodyUpdate{
				Dimming: &LightDimmingState{
					Dimming: 50.0,
				},
			},
			statusCode: http.StatusOK,
			serverResponse: LightUpdateResponse{
				Data: []ResourceIdentifier{
					{
						Action: struct {
							Identity string `json:"identity,omitempty"`
						}{Identity: "light-456"},
					},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *ResourceIdentifier) {
				require.NotNil(t, result)
			},
		},
		{
			name:    "returns error when update fails with API error",
			lightID: "light-error",
			lightUpdate: &LightBodyUpdate{
				On: &LightOnState{On: true},
			},
			statusCode: http.StatusOK,
			serverResponse: LightUpdateResponse{
				Errors: []struct {
					Description string `json:"description,omitempty"`
				}{
					{Description: "light not reachable"},
				},
			},
			wantErr:  true,
			errorMsg: "failed to update light",
		},
		{
			name:    "returns nil when no data in response",
			lightID: "light-empty",
			lightUpdate: &LightBodyUpdate{
				On: &LightOnState{On: false},
			},
			statusCode: http.StatusOK,
			serverResponse: LightUpdateResponse{
				Data: []ResourceIdentifier{},
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *ResourceIdentifier) {
				assert.Nil(t, result)
			},
		},
		{
			name:    "handles server error",
			lightID: "light-500",
			lightUpdate: &LightBodyUpdate{
				On: &LightOnState{On: true},
			},
			statusCode:     http.StatusInternalServerError,
			serverResponse: `{"error": "server error"}`,
			wantErr:        true,
			errorMsg:       "failed to update light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := testutils.MockHTTPResponse(tt.statusCode, tt.serverResponse)
			defer server.Close()

			apiKeyStore := newMockAPIKeyStore()
			apiKeyStore.Set("test-bridge#test-device", "test-api-key")

			client, cleanup := createTestClient(t, "test-bridge", server.URL, apiKeyStore)
			defer cleanup()

			result, err := client.UpdateOneLightById(tt.lightID, tt.lightUpdate)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestTurnOnLightById(t *testing.T) {
	tests := []struct {
		name           string
		lightID        string
		serverResponse interface{}
		statusCode     int
		wantErr        bool
		errorMsg       string
	}{
		{
			name:       "successfully turns on light",
			lightID:    "light-on-test",
			statusCode: http.StatusOK,
			serverResponse: LightUpdateResponse{
				Data: []ResourceIdentifier{
					{
						Action: struct {
							Identity string `json:"identity,omitempty"`
						}{Identity: "light-on-test"},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "returns error when turn on fails",
			lightID: "light-fail",
			statusCode: http.StatusOK,
			serverResponse: LightUpdateResponse{
				Errors: []struct {
					Description string `json:"description,omitempty"`
				}{
					{Description: "light unreachable"},
				},
			},
			wantErr:  true,
			errorMsg: "failed to update light",
		},
		{
			name:           "handles server error",
			lightID:        "light-error",
			statusCode:     http.StatusBadRequest,
			serverResponse: `{"error": "bad request"}`,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that validates the request body
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request method and path
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, tt.lightID)

				// Decode and verify the request body
				var update LightBodyUpdate
				err := json.NewDecoder(r.Body).Decode(&update)
				require.NoError(t, err)
				
				// Verify that On field is set to true
				require.NotNil(t, update.On)
				assert.True(t, update.On.On)

				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			apiKeyStore := newMockAPIKeyStore()
			apiKeyStore.Set("test-bridge#test-device", "test-api-key")

			client, cleanup := createTestClient(t, "test-bridge", server.URL, apiKeyStore)
			defer cleanup()

			err := client.TurnOnLightById(tt.lightID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTurnOffLightById(t *testing.T) {
	tests := []struct {
		name           string
		lightID        string
		serverResponse interface{}
		statusCode     int
		wantErr        bool
		errorMsg       string
	}{
		{
			name:       "successfully turns off light",
			lightID:    "light-off-test",
			statusCode: http.StatusOK,
			serverResponse: LightUpdateResponse{
				Data: []ResourceIdentifier{
					{
						Action: struct {
							Identity string `json:"identity,omitempty"`
						}{Identity: "light-off-test"},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "returns error when turn off fails",
			lightID: "light-fail-off",
			statusCode: http.StatusOK,
			serverResponse: LightUpdateResponse{
				Errors: []struct {
					Description string `json:"description,omitempty"`
				}{
					{Description: "operation failed"},
				},
			},
			wantErr:  true,
			errorMsg: "failed to update light",
		},
		{
			name:           "handles HTTP error",
			lightID:        "light-http-error",
			statusCode:     http.StatusServiceUnavailable,
			serverResponse: `{"error": "service unavailable"}`,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that validates the request body
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request method and path
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, tt.lightID)

				// Decode and verify the request body
				var update LightBodyUpdate
				err := json.NewDecoder(r.Body).Decode(&update)
				require.NoError(t, err)
				
				// Verify that On field is set to false
				require.NotNil(t, update.On)
				assert.False(t, update.On.On)

				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			apiKeyStore := newMockAPIKeyStore()
			apiKeyStore.Set("test-bridge#test-device", "test-api-key")

			client, cleanup := createTestClient(t, "test-bridge", server.URL, apiKeyStore)
			defer cleanup()

			err := client.TurnOffLightById(tt.lightID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLightOperations_Integration(t *testing.T) {
	t.Run("complete light control workflow", func(t *testing.T) {
		// Create a mock server that simulates a real Hue Bridge
		lightState := map[string]bool{
			"light-1": false,
			"light-2": true,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// Handle GET all lights
			if r.Method == http.MethodGet && r.URL.Path == "/clip/v2/resource/light" {
				lights := LightList{
					Data: []LightListItem{
						{
							ID:   "light-1",
							Type: "light",
							On:   LightOnState{On: lightState["light-1"]},
							Meta: LightMeta{Name: "Living Room"},
						},
						{
							ID:   "light-2",
							Type: "light",
							On:   LightOnState{On: lightState["light-2"]},
							Meta: LightMeta{Name: "Bedroom"},
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(lights)
				return
			}

			// Handle PUT (update light)
			if r.Method == http.MethodPut {
				var update LightBodyUpdate
				json.NewDecoder(r.Body).Decode(&update)
				
				// Extract light ID from path
				// In real scenario, would parse the path properly
				response := LightUpdateResponse{
					Data: []ResourceIdentifier{{
						Action: struct {
							Identity string `json:"identity,omitempty"`
						}{Identity: "light-1"},
					}},
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
				return
			}

			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		apiKeyStore := newMockAPIKeyStore()
		apiKeyStore.Set("test-bridge#test-device", "test-api-key")

		client, cleanup := createTestClient(t, "test-bridge", server.URL, apiKeyStore)
		defer cleanup()

		// Test: Get all lights
		lights, err := client.GetAllLights()
		require.NoError(t, err)
		assert.Len(t, lights.Data, 2)

		// Test: Turn on a light
		err = client.TurnOnLightById("light-1")
		require.NoError(t, err)

		// Test: Turn off a light
		err = client.TurnOffLightById("light-2")
		require.NoError(t, err)
	})
}

func TestLightOperations_WithComplexUpdate(t *testing.T) {
	t.Run("updates light with color and brightness", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var update LightBodyUpdate
			json.NewDecoder(r.Body).Decode(&update)

			// Verify complex update structure
			assert.NotNil(t, update.Dimming)
			assert.NotNil(t, update.Color)
			assert.Equal(t, float32(80.0), update.Dimming.Dimming)
			assert.NotNil(t, update.Color.XY)
			assert.Equal(t, float32(0.3), update.Color.XY.X)
			assert.Equal(t, float32(0.4), update.Color.XY.Y)

			response := LightUpdateResponse{
				Data: []ResourceIdentifier{{
					Action: struct {
						Identity string `json:"identity,omitempty"`
					}{Identity: "light-color"},
				}},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		apiKeyStore := newMockAPIKeyStore()
		apiKeyStore.Set("test-bridge#test-device", "test-api-key")

		client, cleanup := createTestClient(t, "test-bridge", server.URL, apiKeyStore)
		defer cleanup()

		update := &LightBodyUpdate{
			Dimming: &LightDimmingState{Dimming: 80.0},
			Color: &LightColor{
				XY: &struct {
					X float32 `json:"x,omitempty"`
					Y float32 `json:"y,omitempty"`
				}{
					X: 0.3,
					Y: 0.4,
				},
			},
		}

		result, err := client.UpdateOneLightById("light-color", update)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}
