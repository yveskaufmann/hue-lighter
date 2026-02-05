package hueclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	deviceName  string
	baseURL     string
	bridgeID    string
	apiKeyStore APIKeyStore
	client      *http.Client
	logger      *log.Entry
}

func NewClient(deviceName string, bridgeID string, bridgeIP string, apiKeyStore APIKeyStore, caBundlePath string, logger *log.Entry) (*Client, error) {

	logger = logger.WithField("component", "HueClient")

	tlsConfig, err := NewBridgeTLSConfig(bridgeID, caBundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %w", err)
	}

	return &Client{
		deviceName:  deviceName,
		baseURL:     fmt.Sprintf("https://%s", bridgeIP),
		apiKeyStore: apiKeyStore,
		client:      &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}},
		bridgeID:    bridgeID,
		logger:      logger,
	}, nil
}

func (c *Client) doRequest(path string, method string, reqBody interface{}, respResource interface{}) error {

	var reqBodyReader io.Reader
	if reqBody != nil {
		w := bytes.Buffer{}
		encoder := json.NewEncoder(&w)
		if err := encoder.Encode(reqBody); err != nil {
			return fmt.Errorf("failed to encode request body: %v", err)
		}
		reqBodyReader = &w

		c.logger.Debugf("Light Request Body: %s", w.String())
	}

	if after, ok := strings.CutPrefix(path, "/"); ok {
		path = after
	}
	url := fmt.Sprintf("%s/%s", c.baseURL, path)

	c.logger.Debugf("Making %s request to %s", method, url)

	req, err := http.NewRequest(method, url, reqBodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	skipApiKey := false
	if strings.HasPrefix(path, "api") || strings.HasPrefix(path, "/api") {
		skipApiKey = true
	}

	if !skipApiKey {
		apiKey, err := c.apiKeyStore.Get(fmt.Sprintf("%s#%s", c.bridgeID, c.deviceName))
		if err != nil {
			if errors.Is(err, ErrMissingAPIKey) {
				return fmt.Errorf("%w %q", ErrMissingAPIKey, c.bridgeID)
			}
			return fmt.Errorf("failed to load api key for hue bridge %q: %w", c.bridgeID, err)
		}
		req.Header.Set("hue-application-key", apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %v", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {

		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}

		return fmt.Errorf("request failed with status code: %d, response: %s", response.StatusCode, body)
	}

	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&respResource); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	return nil
}

func (c *Client) BridgeID() string {
	return c.bridgeID
}

func (c *Client) DeviceName() string {
	return c.deviceName
}
