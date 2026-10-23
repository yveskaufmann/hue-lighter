package hueclient

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationAPIKeyStore_InMemoryPrimaryKeepsKey(t *testing.T) {
	logger := logrus.New().WithField("test", "in-memory-primary")
	persistentStore := newMockAPIKeyStore()
	persistentStore.store["bridge-123"] = "secret"
	store := NewMigrationAPIKeyStore(NewInMemoryAPIKeyStore(logger), persistentStore, logger)

	apiKey, err := store.Get("bridge-123")
	require.NoError(t, err)
	assert.Equal(t, "secret", apiKey)

	// Simulate the keyring losing its D-Bus connection during shutdown.
	persistentStore.getErr = errors.New("dial unix /run/flatpak/bus: connect: connection refused")

	apiKey, err = store.Get("bridge-123")
	require.NoError(t, err)
	assert.Equal(t, "secret", apiKey)
}
