package hueclient

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "com.yv35.hue-lighter"
)

type keyringApiStore struct{
	logger *log.Entry
}

func NewKeyringAPIKeyStore(logger *log.Entry) (APIKeyStore, error) {
	logger = logger.WithField("component", "KeyringAPIKeyStore")
	return &keyringApiStore{
		logger: logger,
	}, nil
}

func (s *keyringApiStore) Get(bridgeID string) (string, error) {
	apiKey, err := keyring.Get(serviceName, fmt.Sprintf("bridge-%s-api-key", bridgeID))
	if err != nil {
		s.logger.Warnf("Failed to get API key for bridge %s: %v", bridgeID, err)
		return "", err
	}

	return apiKey, nil
}

func (s *keyringApiStore) Set(bridgeID string, apiKey string) error {
	err := keyring.Set(serviceName, fmt.Sprintf("bridge-%s-api-key", bridgeID), apiKey)
	if err != nil {
		s.logger.Warnf("Failed to store API key for bridge %s: %v", bridgeID, err)
		return err
	}
	s.logger.Infof("Stored API key for bridge %s (redacted)", bridgeID)
	return nil
}

func (s *keyringApiStore) Remove(bridgeID string) error {
	err := keyring.Delete(serviceName, fmt.Sprintf("bridge-%s-api-key", bridgeID))
	if err != nil {
		s.logger.Warnf("Failed to remove API key for bridge %s: %v", bridgeID, err)
		return err
	}
	s.logger.Infof("Removed API key for bridge %s", bridgeID)

	return nil
}
