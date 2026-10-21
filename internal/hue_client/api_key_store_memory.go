package hueclient

import (
	log "github.com/sirupsen/logrus"
)

type InMemoryAPIKeyStore struct {
	store  map[string]string
	logger *log.Entry
}

func NewInMemoryAPIKeyStore(logger *log.Entry) *InMemoryAPIKeyStore {
	return &InMemoryAPIKeyStore{
		store:  make(map[string]string),
		logger: logger.WithField("component", "InMemoryAPIKeyStore"),
	}
}

func (s *InMemoryAPIKeyStore) Get(bridgeID string) (string, error) {
	apiKey, exists := s.store[bridgeID]
	if !exists {
		s.logger.Warnf("API key for bridge %s not found", bridgeID)
		return "", ErrMissingAPIKey
	}
	return apiKey, nil
}

func (s *InMemoryAPIKeyStore) Set(bridgeID string, apiKey string) error {
	s.store[bridgeID] = apiKey
	s.logger.Infof("Stored API key for bridge %s (redacted)", bridgeID)
	return nil
}

func (s *InMemoryAPIKeyStore) Remove(bridgeID string) error {
	delete(s.store, bridgeID)
	s.logger.Infof("Removed API key for bridge %s", bridgeID)

	return nil
}

