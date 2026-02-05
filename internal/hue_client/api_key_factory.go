package hueclient

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func NewAPIKeyStore(logger *log.Entry) (APIKeyStore, error) {

	// TODO: Support to use different API key stores implementations based on configuration

	apiStorePath := os.Getenv("HUE_API_KEY_STORE_PATH")
	if apiStorePath == "" {
		apiStorePath = "/var/lib/hue-lighter/api-keys.json"
	}

	apiKeyStore, err := NewFileAPIKeyStore(apiStorePath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create file API key store: %w", err)
	}

	return apiKeyStore, nil
}
