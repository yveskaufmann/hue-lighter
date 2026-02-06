package hueclient

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryAPIKeyStore(t *testing.T) {
	logger := logrus.New().WithField("test", "inmemory")

	t.Run("NewInMemoryAPIKeyStore", func(t *testing.T) {
		store := NewInMemoryAPIKeyStore(logger)

		require.NotNil(t, store)
		assert.NotNil(t, store.store)
		assert.NotNil(t, store.logger)
	})

	t.Run("Set and Get API key", func(t *testing.T) {
		store := NewInMemoryAPIKeyStore(logger)

		// Set an API key
		err := store.Set("bridge-123", "api-key-123")
		assert.NoError(t, err)

		// Get the API key
		apiKey, err := store.Get("bridge-123")
		assert.NoError(t, err)
		assert.Equal(t, "api-key-123", apiKey)
	})

	t.Run("Get non-existent API key", func(t *testing.T) {
		store := NewInMemoryAPIKeyStore(logger)

		apiKey, err := store.Get("non-existent-bridge")
		assert.Error(t, err)
		assert.Equal(t, ErrMissingAPIKey, err)
		assert.Empty(t, apiKey)
	})

	t.Run("Remove API key", func(t *testing.T) {
		store := NewInMemoryAPIKeyStore(logger)

		// Set an API key
		err := store.Set("bridge-456", "api-key-456")
		require.NoError(t, err)

		// Verify it exists
		apiKey, err := store.Get("bridge-456")
		require.NoError(t, err)
		assert.Equal(t, "api-key-456", apiKey)

		// Remove the API key
		err = store.Remove("bridge-456")
		assert.NoError(t, err)

		// Verify it's gone
		apiKey, err = store.Get("bridge-456")
		assert.Error(t, err)
		assert.Equal(t, ErrMissingAPIKey, err)
		assert.Empty(t, apiKey)
	})

	t.Run("Multiple API keys", func(t *testing.T) {
		store := NewInMemoryAPIKeyStore(logger)

		// Set multiple API keys
		testKeys := map[string]string{
			"bridge-1": "api-key-1",
			"bridge-2": "api-key-2",
			"bridge-3": "api-key-3",
		}

		for bridgeID, apiKey := range testKeys {
			err := store.Set(bridgeID, apiKey)
			require.NoError(t, err)
		}

		// Verify all keys can be retrieved
		for bridgeID, expectedKey := range testKeys {
			apiKey, err := store.Get(bridgeID)
			require.NoError(t, err)
			assert.Equal(t, expectedKey, apiKey)
		}
	})

	t.Run("Overwrite API key", func(t *testing.T) {
		store := NewInMemoryAPIKeyStore(logger)

		// Set initial API key
		err := store.Set("bridge-overwrite", "initial-key")
		require.NoError(t, err)

		// Overwrite with new key
		err = store.Set("bridge-overwrite", "new-key")
		require.NoError(t, err)

		// Verify new key is returned
		apiKey, err := store.Get("bridge-overwrite")
		require.NoError(t, err)
		assert.Equal(t, "new-key", apiKey)
	})
}

func TestFileAPIKeyStore(t *testing.T) {
	logger := logrus.New().WithField("test", "file")

	t.Run("NewFileAPIKeyStore with non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "api-keys.json")

		store, err := NewFileAPIKeyStore(filePath, logger)

		require.NoError(t, err)
		require.NotNil(t, store)
		assert.Equal(t, filePath, store.filePath)
	})

	t.Run("NewFileAPIKeyStore with existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "api-keys.json")

		// Create existing file with data
		existingData := `{"bridge-existing":"existing-key"}`
		err := os.WriteFile(filePath, []byte(existingData), 0600)
		require.NoError(t, err)

		store, err := NewFileAPIKeyStore(filePath, logger)
		require.NoError(t, err)

		// Verify existing data was loaded
		apiKey, err := store.Get("bridge-existing")
		require.NoError(t, err)
		assert.Equal(t, "existing-key", apiKey)
	})

	t.Run("NewFileAPIKeyStore with invalid JSON file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "api-keys.json")

		// Create file with invalid JSON
		err := os.WriteFile(filePath, []byte(`{invalid json`), 0600)
		require.NoError(t, err)

		store, err := NewFileAPIKeyStore(filePath, logger)
		require.Error(t, err)
		assert.Nil(t, store)
	})

	t.Run("Set and Get API key with file persistence", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "api-keys.json")

		store, err := NewFileAPIKeyStore(filePath, logger)
		require.NoError(t, err)

		// Set an API key
		err = store.Set("bridge-file-test", "file-api-key")
		require.NoError(t, err)

		// Verify file was created and contains data
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Contains(t, string(data), "bridge-file-test")
		assert.Contains(t, string(data), "file-api-key")

		// Get the API key
		apiKey, err := store.Get("bridge-file-test")
		require.NoError(t, err)
		assert.Equal(t, "file-api-key", apiKey)
	})

	t.Run("Remove API key with file persistence", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "api-keys.json")

		store, err := NewFileAPIKeyStore(filePath, logger)
		require.NoError(t, err)

		// Set and then remove an API key
		err = store.Set("bridge-to-remove", "remove-me")
		require.NoError(t, err)

		err = store.Remove("bridge-to-remove")
		require.NoError(t, err)

		// Verify key is gone
		apiKey, err := store.Get("bridge-to-remove")
		assert.Error(t, err)
		assert.Equal(t, ErrMissingAPIKey, err)
		assert.Empty(t, apiKey)

		// Verify file doesn't contain the removed key
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.NotContains(t, string(data), "bridge-to-remove")
	})

	t.Run("Load timestamp and refresh interval", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "api-keys.json")

		// Pre-create file with data
		err := os.WriteFile(filePath, []byte(`{"bridge-refresh":"refresh-key"}`), 0600)
		require.NoError(t, err)

		store, err := NewFileAPIKeyStore(filePath, logger)
		require.NoError(t, err)

		// First load sets timestamp
		initialTimestamp := store.lastLoadTimestamp
		assert.False(t, initialTimestamp.IsZero())

		// Immediate subsequent get should use cached data (no reload)
		apiKey, err := store.Get("bridge-refresh")
		require.NoError(t, err)
		assert.Equal(t, "refresh-key", apiKey)

		// Timestamp should be the same (no reload occurred)
		assert.Equal(t, initialTimestamp, store.lastLoadTimestamp)
	})

	t.Run("File creation with directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "nested", "path", "api-keys.json")

		store, err := NewFileAPIKeyStore(filePath, logger)
		require.NoError(t, err)

		// Set a key - should create nested directories
		err = store.Set("bridge-nested", "nested-key")
		require.NoError(t, err)

		// Verify file exists in nested path
		_, err = os.Stat(filePath)
		require.NoError(t, err)
	})

	t.Run("File permissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "permissions", "api-keys.json")

		store, err := NewFileAPIKeyStore(filePath, logger)
		require.NoError(t, err)

		// Set a key
		err = store.Set("bridge-perms", "perms-key")
		require.NoError(t, err)

		// Check file permissions (should be 0600)
		fileInfo, err := os.Stat(filePath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), fileInfo.Mode().Perm())

		// Check directory permissions (should be 0700)
		dirInfo, err := os.Stat(filepath.Dir(filePath))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0700), dirInfo.Mode().Perm())
	})

	t.Run("Multiple FileAPIKeyStore instances sharing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "shared-api-keys.json")

		// Create first store and set a key
		store1, err := NewFileAPIKeyStore(filePath, logger)
		require.NoError(t, err)

		err = store1.Set("bridge-shared", "shared-key-1")
		require.NoError(t, err)

		// Create second store (should load existing data)
		store2, err := NewFileAPIKeyStore(filePath, logger)
		require.NoError(t, err)

		// Second store should see data from first store
		apiKey, err := store2.Get("bridge-shared")
		require.NoError(t, err)
		assert.Equal(t, "shared-key-1", apiKey)

		// Update from second store
		err = store2.Set("bridge-shared", "shared-key-2")
		require.NoError(t, err)

		// First store should load updated data on next access
		// Force reload by resetting timestamp
		store1.lastLoadTimestamp = time.Time{}

		apiKey, err = store1.Get("bridge-shared")
		require.NoError(t, err)
		assert.Equal(t, "shared-key-2", apiKey)
	})
}

func TestErrMissingAPIKey(t *testing.T) {
	assert.NotNil(t, ErrMissingAPIKey)
	assert.Contains(t, ErrMissingAPIKey.Error(), "missing API key")
}
