package settings

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"

	"com.github.yveskaufmann/hue-lighter/internal/config"
	"com.github.yveskaufmann/hue-lighter/internal/geo"
	hueclient "com.github.yveskaufmann/hue-lighter/internal/hue_client"
	"com.github.yveskaufmann/hue-lighter/internal/messages"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCatalog struct {
	lights []hueclient.LightListItem
	err    error
}

func (c *fakeCatalog) GetAllLights() (*hueclient.LightList, error) {
	if c.err != nil {
		return nil, c.err
	}
	return &hueclient.LightList{Data: c.lights}, nil
}

func bridgeLight(id, name string) hueclient.LightListItem {
	return hueclient.LightListItem{ID: id, Meta: hueclient.LightMeta{Name: name}}
}

func ptr(s string) *string { return &s }

type testServer struct {
	server   *Server
	commands chan messages.Command
	url      string
	base     string
	client   *http.Client
}

func newTestServer(t *testing.T, catalog LightCatalog) *testServer {
	commands := make(chan messages.Command, 1)
	locate := func(ctx context.Context, logger *logrus.Entry) (geo.Location, error) {
		return geo.Location{Latitude: 52.5, Longitude: 13.3667, Source: geo.SourceTimeZone}, nil
	}
	server := NewServer(catalog, locate, commands, logrus.New().WithField("test", t.Name()))

	cfg := &config.Config{
		Location: config.Location{Latitude: 52.56, Longitude: 13.51},
		Lights: []config.Light{
			{ID: ptr("light-l"), Name: ptr("Old name L")},
			{ID: ptr("light-gone"), Name: ptr("Removed light")},
		},
	}
	cfg.Meta.Name = "Hue Lighter Automation"
	server.SetConfig(cfg)

	openURL, err := server.Start()
	require.NoError(t, err)
	t.Cleanup(func() { server.Stop(context.Background()) })

	jar, _ := cookiejar.New(nil)
	parsed, _ := url.Parse(openURL)
	return &testServer{
		server:   server,
		commands: commands,
		url:      openURL,
		base:     "http://" + parsed.Host,
		client:   &http.Client{Jar: jar},
	}
}

// login opens the URL with the token, as the browser does, to get the session cookie.
func (ts *testServer) login(t *testing.T) {
	response, err := ts.client.Get(ts.url)
	require.NoError(t, err)
	response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "/", response.Request.URL.RequestURI(), "token is dropped from the address")
}

func (ts *testServer) do(t *testing.T, method, path, contentType, body string) (*http.Response, map[string]interface{}) {
	request, err := http.NewRequest(method, ts.base+path, strings.NewReader(body))
	require.NoError(t, err)
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response, err := ts.client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	var decoded map[string]interface{}
	json.NewDecoder(response.Body).Decode(&decoded)
	return response, decoded
}

func TestServer_RequiresToken(t *testing.T) {
	ts := newTestServer(t, &fakeCatalog{})

	response, _ := ts.do(t, http.MethodGet, "/api/settings", "", "")
	assert.Equal(t, http.StatusForbidden, response.StatusCode, "no session cookie")

	response, _ = ts.do(t, http.MethodGet, "/?token=wrong", "", "")
	assert.Equal(t, http.StatusForbidden, response.StatusCode, "wrong token")

	ts.login(t)
	response, _ = ts.do(t, http.MethodGet, "/api/settings", "", "")
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestServer_RejectsForeignHost(t *testing.T) {
	ts := newTestServer(t, &fakeCatalog{})
	ts.login(t)

	request, err := http.NewRequest(http.MethodGet, ts.base+"/api/settings", nil)
	require.NoError(t, err)
	request.Host = "evil.example:80"
	response, err := ts.client.Do(request)
	require.NoError(t, err)
	response.Body.Close()

	assert.Equal(t, http.StatusForbidden, response.StatusCode)
}

func TestServer_ServesPage(t *testing.T) {
	ts := newTestServer(t, &fakeCatalog{})
	ts.login(t)

	for _, path := range []string{"/", "/static/app.js", "/static/app.css"} {
		response, err := ts.client.Get(ts.base + path)
		require.NoError(t, err)
		response.Body.Close()
		assert.Equal(t, http.StatusOK, response.StatusCode, path)
	}
}

func TestServer_GetSettings(t *testing.T) {
	ts := newTestServer(t, &fakeCatalog{lights: []hueclient.LightListItem{
		bridgeLight("light-l", "Arbeitszimmer Hue Play L"),
		bridgeLight("light-r", "Arbeitszimmer Hue Play R"),
	}})
	ts.login(t)

	response, body := ts.do(t, http.MethodGet, "/api/settings", "", "")

	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, map[string]interface{}{"latitude": 52.56, "longitude": 13.51}, body["location"])
	assert.Equal(t, []interface{}{
		// Renamed on the bridge: the current name is shown.
		map[string]interface{}{"id": "light-l", "name": "Arbeitszimmer Hue Play L", "available": true},
		map[string]interface{}{"id": "light-gone", "name": "Removed light", "available": false},
	}, body["lights"])
}

func TestServer_GetBridgeLights(t *testing.T) {
	t.Run("sorted by name", func(t *testing.T) {
		ts := newTestServer(t, &fakeCatalog{lights: []hueclient.LightListItem{
			bridgeLight("light-r", "Arbeitszimmer Hue Play R"),
			bridgeLight("light-k", "küche"),
			bridgeLight("light-l", "Arbeitszimmer Hue Play L"),
		}})
		ts.login(t)

		response, err := ts.client.Get(ts.base + "/api/bridge/lights")
		require.NoError(t, err)
		defer response.Body.Close()
		var lights []lightDTO
		require.NoError(t, json.NewDecoder(response.Body).Decode(&lights))

		assert.Equal(t, []lightDTO{
			{ID: "light-l", Name: "Arbeitszimmer Hue Play L"},
			{ID: "light-r", Name: "Arbeitszimmer Hue Play R"},
			{ID: "light-k", Name: "küche"},
		}, lights)
	})

	t.Run("bridge unavailable", func(t *testing.T) {
		ts := newTestServer(t, &fakeCatalog{err: errors.New("connection refused")})
		ts.login(t)

		response, body := ts.do(t, http.MethodGet, "/api/bridge/lights", "", "")

		assert.Equal(t, http.StatusBadGateway, response.StatusCode)
		assert.Contains(t, body["error"], "could not be loaded")
	})
}

func TestServer_GetLocationDefault(t *testing.T) {
	ts := newTestServer(t, &fakeCatalog{})
	ts.login(t)

	response, body := ts.do(t, http.MethodGet, "/api/location/default", "", "")

	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, map[string]interface{}{"latitude": 52.5, "longitude": 13.3667, "source": "time-zone"}, body)
}

func TestServer_PutSettings(t *testing.T) {
	ts := newTestServer(t, &fakeCatalog{lights: []hueclient.LightListItem{bridgeLight("light-r", "Hue Play R")}})
	ts.login(t)

	// Play the application: apply the settings sent as command.
	applied := make(chan *config.Config, 1)
	go func() {
		cmd := (<-ts.commands).(messages.ApplySettingsCommand)
		applied <- cmd.Config
		cmd.Reply <- nil
	}()

	response, body := ts.do(t, http.MethodPut, "/api/settings", "application/json",
		`{"location":{"latitude":48.1,"longitude":11.6},"lights":[{"id":"light-r","name":"Hue Play R"},{"id":"light-r","name":"Hue Play R"}]}`)

	require.Equal(t, http.StatusOK, response.StatusCode, body)
	cfg := <-applied
	assert.Equal(t, config.Location{Latitude: 48.1, Longitude: 11.6}, cfg.Location)
	assert.Equal(t, []config.Light{{ID: ptr("light-r"), Name: ptr("Hue Play R")}}, cfg.Lights, "duplicates are dropped")
	assert.Equal(t, "Hue Lighter Automation", cfg.Meta.Name, "meta is kept")

	// The response and later reads show the saved settings.
	assert.Equal(t, []interface{}{map[string]interface{}{"id": "light-r", "name": "Hue Play R", "available": true}}, body["lights"])
}

func TestServer_PutSettings_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		status      int
		errorText   string
	}{
		{name: "no lights", contentType: "application/json", status: http.StatusBadRequest,
			body:      `{"location":{"latitude":48.1,"longitude":11.6},"lights":[]}`,
			errorText: "at least one light"},
		{name: "latitude out of range", contentType: "application/json", status: http.StatusBadRequest,
			body:      `{"location":{"latitude":91,"longitude":11.6},"lights":[{"id":"a","name":"A"}]}`,
			errorText: "latitude"},
		{name: "light without id", contentType: "application/json", status: http.StatusBadRequest,
			body:      `{"location":{"latitude":48.1,"longitude":11.6},"lights":[{"id":" ","name":"A"}]}`,
			errorText: "no ID"},
		{name: "not json", contentType: "application/x-www-form-urlencoded", status: http.StatusUnsupportedMediaType,
			body: `lights=a`, errorText: "application/json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := newTestServer(t, &fakeCatalog{})
			ts.login(t)

			response, body := ts.do(t, http.MethodPut, "/api/settings", tt.contentType, tt.body)

			assert.Equal(t, tt.status, response.StatusCode)
			assert.Contains(t, body["error"], tt.errorText)
			assert.Empty(t, ts.commands, "nothing is sent to the application")
		})
	}
}

func TestServer_PutSettings_ApplyFails(t *testing.T) {
	ts := newTestServer(t, &fakeCatalog{})
	ts.login(t)

	go func() {
		cmd := (<-ts.commands).(messages.ApplySettingsCommand)
		cmd.Reply <- errors.New("disk full")
	}()

	response, body := ts.do(t, http.MethodPut, "/api/settings", "application/json",
		`{"location":{"latitude":48.1,"longitude":11.6},"lights":[{"id":"a","name":"A"}]}`)

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	assert.Contains(t, body["error"], "disk full")
}
