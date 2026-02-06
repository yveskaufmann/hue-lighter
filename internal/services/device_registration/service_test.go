package device_registration

import (
	"errors"
	"testing"
	"time"

	hueclient "com.github.yveskaufmann/hue-lighter/internal/hue_client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock HueClient for testing
type mockHueClient struct {
	bridgeID         string
	deviceName       string
	registerResponse *hueclient.DeviceRegistrationResponse
	registerError    error
}

func (m *mockHueClient) BridgeID() string {
	return m.bridgeID
}

func (m *mockHueClient) DeviceName() string {
	return m.deviceName
}

func (m *mockHueClient) RegisterDevice(name string) (*hueclient.DeviceRegistrationResponse, error) {
	if m.registerError != nil {
		return nil, m.registerError
	}
	return m.registerResponse, nil
}

// Mock APIKeyStore for testing
type mockAPIKeyStore struct {
	keys      map[string]string
	getError  error
	setError  error
}

func newMockAPIKeyStore() *mockAPIKeyStore {
	return &mockAPIKeyStore{
		keys: make(map[string]string),
	}
}

func (m *mockAPIKeyStore) Get(identifier string) (string, error) {
	if m.getError != nil {
		return "", m.getError
	}
	key, exists := m.keys[identifier]
	if !exists {
		return "", hueclient.ErrMissingAPIKey
	}
	return key, nil
}

func (m *mockAPIKeyStore) Set(identifier string, apiKey string) error {
	if m.setError != nil {
		return m.setError
	}
	m.keys[identifier] = apiKey
	return nil
}

func (m *mockAPIKeyStore) Remove(identifier string) error {
	delete(m.keys, identifier)
	return nil
}

func TestNewService(t *testing.T) {
	t.Run("creates new registration service", func(t *testing.T) {
		logger := logrus.New().WithField("test", "new-service")
		client := &mockHueClient{
			bridgeID:   "test-bridge",
			deviceName: "test-device",
		}
		apiKeyStore := newMockAPIKeyStore()

		service := NewService(client, apiKeyStore, logger)

		assert.NotNil(t, service)
		assert.NotNil(t, service.client)
		assert.NotNil(t, service.apiKeyStore)
		assert.NotNil(t, service.logger)
	})
}

func TestService_RegisterDevice_AlreadyRegistered(t *testing.T) {
	t.Run("skips registration if device already registered", func(t *testing.T) {
		logger := logrus.New().WithField("test", "already-registered")
		client := &mockHueClient{
			bridgeID:   "bridge-123",
			deviceName: "device-1",
		}
		apiKeyStore := newMockAPIKeyStore()
		
		// Pre-populate the API key store with existing registration
		apiKeyStore.Set("bridge-123#test-device", "existing-api-key")

		service := NewService(client, apiKeyStore, logger)

		// This should return immediately without calling RegisterDevice
		err := service.RegisterDevice("test-device")
		
		require.NoError(t, err)
		
		// Verify the API key is still the same
		key, err := apiKeyStore.Get("bridge-123#test-device")
		require.NoError(t, err)
		assert.Equal(t, "existing-api-key", key)
	})
}

func TestService_RegisterDevice_Success(t *testing.T) {
	t.Run("successfully registers new device", func(t *testing.T) {
		logger := logrus.New().WithField("test", "register-success")
		
		client := &mockHueClient{
			bridgeID:   "bridge-456",
			deviceName: "device-2",
			registerResponse: &hueclient.DeviceRegistrationResponse{
				Success: &struct {
					Username  string `json:"username,omitempty"`
					ClientKey string `json:"clientkey,omitempty"`
				}{
					Username:  "new-api-key-123",
					ClientKey: "new-client-key-456",
				},
			},
		}
		apiKeyStore := newMockAPIKeyStore()

		service := NewService(client, apiKeyStore, logger)

		// Start registration in a goroutine since it waits 15 seconds
		done := make(chan error, 1)
		go func() {
			done <- service.RegisterDevice("new-device")
		}()

		// Wait a bit more than the 15 second delay
		select {
		case err := <-done:
			require.NoError(t, err)
			
			// Verify API key was stored
			key, err := apiKeyStore.Get("bridge-456#device-2")
			require.NoError(t, err)
			assert.Equal(t, "new-api-key-123", key)
			
		case <-time.After(20 * time.Second):
			t.Fatal("Registration took too long")
		}
	})
}

func TestService_RegisterDevice_LinkButtonNotPressed(t *testing.T) {
	t.Run("returns error when link button not pressed", func(t *testing.T) {
		logger := logrus.New().WithField("test", "link-button-not-pressed")
		
		client := &mockHueClient{
			bridgeID:   "bridge-789",
			deviceName: "device-3",
			registerResponse: &hueclient.DeviceRegistrationResponse{
				Error: &struct {
					Type        int    `json:"type,omitempty"`
					Address     string `json:"address,omitempty"`
					Description string `json:"description,omitempty"`
				}{
					Type:        hueclient.HueErrorTypeLinkButtonNotPressed,
					Description: "link button not pressed",
				},
			},
		}
		apiKeyStore := newMockAPIKeyStore()

		service := NewService(client, apiKeyStore, logger)

		// Start registration in a goroutine
		done := make(chan error, 1)
		go func() {
			done <- service.RegisterDevice("test-device")
		}()

		// Wait for result
		select {
		case err := <-done:
			require.Error(t, err)
			assert.Contains(t, err.Error(), "link button not pressed")
			
			// Verify API key was NOT stored
			_, err = apiKeyStore.Get("bridge-789#device-3")
			assert.Error(t, err)
			
		case <-time.After(20 * time.Second):
			t.Fatal("Registration took too long")
		}
	})
}

func TestService_RegisterDevice_RegistrationAPIError(t *testing.T) {
	t.Run("handles registration API error", func(t *testing.T) {
		logger := logrus.New().WithField("test", "api-error")
		
		client := &mockHueClient{
			bridgeID:      "bridge-error",
			deviceName:    "device-error",
			registerError: errors.New("network error"),
		}
		apiKeyStore := newMockAPIKeyStore()

		service := NewService(client, apiKeyStore, logger)

		// Start registration in a goroutine
		done := make(chan error, 1)
		go func() {
			done <- service.RegisterDevice("test-device")
		}()

		// Wait for result
		select {
		case err := <-done:
			require.Error(t, err)
			assert.Contains(t, err.Error(), "network error")
			
		case <-time.After(20 * time.Second):
			t.Fatal("Registration took too long")
		}
	})
}

func TestService_RegisterDevice_StoreError(t *testing.T) {
	t.Run("returns error when API key storage fails", func(t *testing.T) {
		logger := logrus.New().WithField("test", "store-error")
		
		client := &mockHueClient{
			bridgeID:   "bridge-store",
			deviceName: "device-store",
			registerResponse: &hueclient.DeviceRegistrationResponse{
				Success: &struct {
					Username  string `json:"username,omitempty"`
					ClientKey string `json:"clientkey,omitempty"`
				}{
					Username:  "api-key-store-test",
					ClientKey: "client-key-store-test",
				},
			},
		}
		apiKeyStore := newMockAPIKeyStore()
		apiKeyStore.setError = errors.New("storage error")

		service := NewService(client, apiKeyStore, logger)

		// Start registration in a goroutine
		done := make(chan error, 1)
		go func() {
			done <- service.RegisterDevice("test-device")
		}()

		// Wait for result
		select {
		case err := <-done:
			require.Error(t, err)
			assert.Contains(t, err.Error(), "storage error")
			
		case <-time.After(20 * time.Second):
			t.Fatal("Registration took too long")
		}
	})
}

func TestService_RegisterDevice_OtherHueError(t *testing.T) {
	t.Run("handles other Hue API errors", func(t *testing.T) {
		logger := logrus.New().WithField("test", "other-hue-error")
		
		client := &mockHueClient{
			bridgeID:   "bridge-other-error",
			deviceName: "device-other",
			registerResponse: &hueclient.DeviceRegistrationResponse{
				Error: &struct {
					Type        int    `json:"type,omitempty"`
					Address     string `json:"address,omitempty"`
					Description string `json:"description,omitempty"`
				}{
					Type:        999,
					Description: "unknown error",
				},
			},
		}
		apiKeyStore := newMockAPIKeyStore()

		service := NewService(client, apiKeyStore, logger)

		// Start registration in a goroutine
		done := make(chan error, 1)
		go func() {
			done <- service.RegisterDevice("test-device")
		}()

		// Wait for result
		select {
		case err := <-done:
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unknown error")
			
		case <-time.After(20 * time.Second):
			t.Fatal("Registration took too long")
		}
	})
}

func TestService_RegisterDevice_WithExistingKeyDifferentDevice(t *testing.T) {
	t.Run("registers new device even if another device is registered", func(t *testing.T) {
		logger := logrus.New().WithField("test", "different-device")
		
		client := &mockHueClient{
			bridgeID:   "bridge-multi",
			deviceName: "device-new",
			registerResponse: &hueclient.DeviceRegistrationResponse{
				Success: &struct {
					Username  string `json:"username,omitempty"`
					ClientKey string `json:"clientkey,omitempty"`
				}{
					Username:  "new-device-api-key",
					ClientKey: "new-device-client-key",
				},
			},
		}
		apiKeyStore := newMockAPIKeyStore()
		
		// Pre-register a different device
		apiKeyStore.Set("bridge-multi#old-device", "old-api-key")

		service := NewService(client, apiKeyStore, logger)

		// Start registration in a goroutine
		done := make(chan error, 1)
		go func() {
			done <- service.RegisterDevice("device-new")
		}()

		// Wait for result
		select {
		case err := <-done:
			require.NoError(t, err)
			
			// Verify new device API key was stored
			key, err := apiKeyStore.Get("bridge-multi#device-new")
			require.NoError(t, err)
			assert.Equal(t, "new-device-api-key", key)
			
			// Verify old device key still exists
			oldKey, err := apiKeyStore.Get("bridge-multi#old-device")
			require.NoError(t, err)
			assert.Equal(t, "old-api-key", oldKey)
			
		case <-time.After(20 * time.Second):
			t.Fatal("Registration took too long")
		}
	})
}
