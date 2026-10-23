package light_automation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"com.github.yveskaufmann/hue-lighter/internal/config"
	hueclient "com.github.yveskaufmann/hue-lighter/internal/hue_client"
	"com.github.yveskaufmann/hue-lighter/internal/messages"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBridge serves the light endpoints of the Hue CLIP v2 API from an in-memory light state.
type fakeBridge struct {
	mu     sync.Mutex
	lights map[string]bool
}

func (b *fakeBridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := strings.TrimPrefix(r.URL.Path, "/clip/v2/resource/light/")
	on, ok := b.lights[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(hueclient.LightList{Data: []hueclient.LightListItem{{ID: id, On: hueclient.LightOnState{On: on}}}})
	case http.MethodPut:
		var update hueclient.LightBodyUpdate
		json.NewDecoder(r.Body).Decode(&update)
		b.lights[id] = update.On.On
		json.NewEncoder(w).Encode(hueclient.LightUpdateResponse{Data: []hueclient.ResourceIdentifier{{}}})
	}
}

func newTestService(t *testing.T, bridge *fakeBridge, events chan messages.LightStateChangedEvent, lightIDs ...string) *Service {
	server := httptest.NewServer(bridge)
	t.Cleanup(server.Close)

	logger := logrus.New().WithField("test", t.Name())
	store := hueclient.NewInMemoryAPIKeyStore(logger)
	require.NoError(t, store.Set("bridge-1#test-device", "api-key"))
	client := hueclient.NewClientWithHTTPClient("test-device", "bridge-1", server.URL, store, server.Client(), logger)

	return NewService(client, testConfig(lightIDs...), logger, events)
}

func testConfig(lightIDs ...string) *config.Config {
	cfg := &config.Config{Location: config.Location{Latitude: 52.52, Longitude: 13.40}}
	for _, id := range lightIDs {
		cfg.Lights = append(cfg.Lights, config.Light{ID: &id, Name: &id})
	}
	return cfg
}

func TestService_ApplyConfig(t *testing.T) {
	bridge := &fakeBridge{lights: map[string]bool{"light-a": true, "light-b": true, "light-c": false}}
	events := make(chan messages.LightStateChangedEvent, 1)
	service := newTestService(t, bridge, events, "light-a", "light-b")

	service.refreshLightStates()
	assert.Equal(t, map[string]bool{"light-a": true, "light-b": true}, service.lightStates)
	assert.Equal(t, messages.LightStateChangedEvent{On: true}, <-events)

	service.ApplyConfig(testConfig("light-b", "light-c"))

	// light-a is no longer configured, light-c was fetched from the bridge.
	assert.Equal(t, map[string]bool{"light-b": true, "light-c": false}, service.lightStates)
	assert.Equal(t, messages.LightStateChangedEvent{On: false}, <-events)

	// Toggling only touches the newly configured lights.
	service.ToggleLights()
	assert.Equal(t, map[string]bool{"light-a": true, "light-b": true, "light-c": true}, bridge.lights)
}
