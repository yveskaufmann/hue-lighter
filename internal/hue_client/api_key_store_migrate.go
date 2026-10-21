package hueclient

import (
	log "github.com/sirupsen/logrus"
)

type MigrationAPIKeyStore struct {
	primaryStore APIKeyStore
	legacyStore APIKeyStore
	logger *log.Entry
}

// Migrates from given legacy api key store to the primary store.
// If an API key is not found in the primary store but exists in the legacy store,
// it will be automatically migrated to the primary store upon retrieval.
func NewMigrationAPIKeyStore(primaryStore APIKeyStore, legacyStore APIKeyStore, logger *log.Entry) *MigrationAPIKeyStore {

	logger = logger.WithField("component", "MigrationAPIKeyStore")

	return &MigrationAPIKeyStore{
		primaryStore: primaryStore,
		legacyStore:  legacyStore,
		logger:       logger,
	}
}

func (s *MigrationAPIKeyStore) Get(bridgeID string) (string, error) {

	apiKey, err := s.primaryStore.Get(bridgeID)
	if err == nil {
		return apiKey, nil
	}

	s.logger.Warnf("API key for bridge %s not found in primary store, trying legacy store", bridgeID)
	apiKey, err = s.legacyStore.Get(bridgeID)
	if err != nil {
		return "", err
	}

	// Migrate the API key to the primary store
	if err := s.primaryStore.Set(bridgeID, apiKey); err != nil {
		s.logger.Warnf("Failed to migrate API key for bridge %s to primary store: %v", bridgeID, err)
	}

	s.logger.Infof("Successfully migrated API key for bridge %s to primary store", bridgeID)

	return apiKey, nil
}

func (s *MigrationAPIKeyStore) Set(bridgeID string, apiKey string) error {
	if err := s.primaryStore.Set(bridgeID, apiKey); err != nil {
		s.logger.Warnf("Failed to set API key for bridge %s in primary store: %v", bridgeID, err)
		return err
	}

	// TODO: Unnecessary to set in legacy store if already set in primary store, consider removing this step
	if err := s.legacyStore.Set(bridgeID, apiKey); err != nil {
		s.logger.Warnf("Failed to set API key for bridge %s in legacy store: %v", bridgeID, err)
		return err
	}

	s.logger.Infof("Successfully set API key for bridge %s in both stores", bridgeID)
	return nil
}

func (s *MigrationAPIKeyStore) Remove(bridgeID string) error {
	if err := s.primaryStore.Remove(bridgeID); err != nil {
		s.logger.Warnf("Failed to remove API key for bridge %s from primary store: %v", bridgeID, err)
		return err
	}

	// TODO: Unnecessary to remove from legacy store if already removed from primary store, consider removing this step
	// only relevant for rollback situations
	if err := s.legacyStore.Remove(bridgeID); err != nil {
		s.logger.Warnf("Failed to remove API key for bridge %s from legacy store: %v", bridgeID, err)
		return err
	}
	s.logger.Infof("Successfully removed API key for bridge %s from both stores", bridgeID)
	return nil
}

