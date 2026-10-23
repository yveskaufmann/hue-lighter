package hueclient

import (
	"fmt"

	"com.github.yveskaufmann/hue-lighter/internal/paths"
	log "github.com/sirupsen/logrus"
)

func NewAPIKeyStore(logger *log.Entry) (APIKeyStore, error) {

	// TODO: Support to use different API key stores implementations based on configuration

	apiKeyFileStorePath, err := paths.APIKeyFile()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve API key store path: %w", err)
	}

	apiKeyFileStore, err := NewFileAPIKeyStore(apiKeyFileStorePath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create file API key store: %w", err)
	}

	keyringStore, err := NewKeyringAPIKeyStore(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyring API key store: %w", err)
	}

	persistentStore := NewMigrationAPIKeyStore(keyringStore, apiKeyFileStore, logger)

	// Keep API keys in memory once read from the persistent store, so turning off
	// the lights on logout/shutdown does not depend on the session D-Bus (keyring)
	// still being available. A changed API key requires a restart.
	return NewMigrationAPIKeyStore(NewInMemoryAPIKeyStore(logger), persistentStore, logger), nil
}
