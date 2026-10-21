package hueclient

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func NewAPIKeyStore(logger *log.Entry) (APIKeyStore, error) {

	// TODO: Support to use different API key stores implementations based on configuration

	apiKeyFileStorePath := os.Getenv("HUE_API_KEY_STORE_PATH")
	if apiKeyFileStorePath == "" {
		apiKeyFileStorePath = "/var/lib/hue-lighter/api-keys.json"
	}

	apiKeyFileStore, err := NewFileAPIKeyStore(apiKeyFileStorePath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create file API key store: %w", err)
	}

	keyringStore, err := NewKeyringAPIKeyStore(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyring API key store: %w", err)
	}

	return NewMigrationAPIKeyStore(keyringStore, apiKeyFileStore, logger), nil
}
