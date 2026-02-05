package hueclient

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
)

var ErrMissingAPIKey = fmt.Errorf("missing API key for Hue bridge")

type APIKeyStore interface {
	Get(bridgeID string) (string, error)
	Set(bridgeID string, apiKey string) error
	Remove(bridgeID string) error
}

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

type FileAPIKeyStore struct {
	store             InMemoryAPIKeyStore
	filePath          string
	lastLoadTimestamp time.Time
	refreshInterval   time.Duration
	logger            *log.Entry
}

func NewFileAPIKeyStore(filePath string, logger *log.Entry) (*FileAPIKeyStore, error) {
	logger = logger.WithField("component", "FileAPIKeyStore")

	memoryStore := InMemoryAPIKeyStore{
		store:  make(map[string]string),
		logger: logger,
	}

	store := &FileAPIKeyStore{
		store:             memoryStore,
		filePath:          filePath,
		lastLoadTimestamp: time.Time{},
		refreshInterval:   5 * time.Second,
		logger:            logger,
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

// Load all keys from the file into a memory store
func (s *FileAPIKeyStore) load() error {

	if time.Since(s.lastLoadTimestamp) < s.refreshInterval {
		s.logger.WithFields(log.Fields{
			"lastLoadTime":    s.lastLoadTimestamp,
			"refreshInterval": s.refreshInterval,
		}).Debug("Skipping load from file because refresh interval not reached")
		return nil
	}

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&s.store.store); err != nil {
		return err
	}

	s.lastLoadTimestamp = time.Now()
	s.logger.WithFields(log.Fields{"storePath": s.filePath}).Info("Loaded API keys from file store")
	return nil

}

func (s *FileAPIKeyStore) save() error {

	baseDir := path.Dir(s.filePath)
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return err
	}

	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err = encoder.Encode(s.store.store); err != nil {
		return err
	}

	s.logger.WithFields(log.Fields{"storePath": s.filePath}).Info("Stored API keys to file store")

	return nil
}

func (s *FileAPIKeyStore) Get(bridgeID string) (string, error) {
	if err := s.load(); err != nil {
		return "", err
	}

	return s.store.Get(bridgeID)
}

func (s *FileAPIKeyStore) Set(bridgeID string, apiKey string) error {
	if err := s.load(); err != nil {
		return err
	}

	if err := s.store.Set(bridgeID, apiKey); err != nil {
		return err
	}
	return s.save()
}

func (s *FileAPIKeyStore) Remove(bridgeID string) error {
	if err := s.load(); err != nil {
		return err
	}

	if err := s.store.Remove(bridgeID); err != nil {
		return err
	}
	return s.save()
}
