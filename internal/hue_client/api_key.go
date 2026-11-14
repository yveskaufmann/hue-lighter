package hueclient

import (
	"fmt"
)

var ErrMissingAPIKey = fmt.Errorf("missing API key for Hue bridge")

type APIKeyStore interface {
	Get(bridgeID string) (string, error)
	Set(bridgeID string, apiKey string) error
	Remove(bridgeID string) error
}
