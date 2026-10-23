package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"sort"
	"strings"

	"com.github.yveskaufmann/hue-lighter/internal/config"
	"com.github.yveskaufmann/hue-lighter/internal/geo"
	"com.github.yveskaufmann/hue-lighter/internal/messages"
)

type locationDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type lightDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Available is false when the light is no longer known to the bridge and
	// nil when the bridge could not be asked.
	Available *bool `json:"available,omitempty"`
}

type settingsDTO struct {
	Location locationDTO `json:"location"`
	Lights   []lightDTO  `json:"lights"`
}

type errorDTO struct {
	Error string `json:"error"`
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	cfg := s.currentConfig()

	// Mark lights the bridge does not know (anymore) and show their current name.
	bridgeLights, err := s.bridgeLights()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to load lights from bridge")
	}
	bridgeNames := make(map[string]string, len(bridgeLights))
	for _, light := range bridgeLights {
		bridgeNames[light.ID] = light.Name
	}

	settings := settingsDTO{
		Location: locationDTO{Latitude: cfg.Location.Latitude, Longitude: cfg.Location.Longitude},
		Lights:   []lightDTO{},
	}
	for _, light := range cfg.Lights {
		dto := lightDTO{ID: deref(light.ID), Name: deref(light.Name)}
		if err == nil {
			name, available := bridgeNames[dto.ID]
			dto.Available = &available
			if available {
				dto.Name = name
			}
		}
		settings.Lights = append(settings.Lights, dto)
	}

	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handleGetBridgeLights(w http.ResponseWriter, r *http.Request) {
	lights, err := s.bridgeLights()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to load lights from bridge")
		writeJSON(w, http.StatusBadGateway, errorDTO{Error: "The lights could not be loaded from the Hue bridge."})
		return
	}
	writeJSON(w, http.StatusOK, lights)
}

func (s *Server) handleGetLocationDefault(w http.ResponseWriter, r *http.Request) {
	location, err := s.locate(r.Context(), s.logger)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorDTO{Error: "Your location could not be determined, please enter it manually."})
		return
	}
	writeJSON(w, http.StatusOK, location)
}

func (s *Server) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	// Requiring JSON also protects against cross-site form posts.
	if mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mediaType != "application/json" {
		writeJSON(w, http.StatusUnsupportedMediaType, errorDTO{Error: "expected application/json"})
		return
	}

	var settings settingsDTO
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&settings); err != nil {
		writeJSON(w, http.StatusBadRequest, errorDTO{Error: "invalid request: " + err.Error()})
		return
	}

	cfg, err := s.applyTo(s.currentConfig(), settings)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorDTO{Error: err.Error()})
		return
	}

	reply := make(chan error, 1)
	select {
	case s.commands <- messages.ApplySettingsCommand{Config: cfg, Reply: reply}:
	case <-r.Context().Done():
		return
	}

	select {
	case err := <-reply:
		if err != nil {
			s.logger.WithError(err).Error("Failed to apply settings")
			writeJSON(w, http.StatusInternalServerError, errorDTO{Error: "The settings could not be saved: " + err.Error()})
			return
		}
	case <-r.Context().Done():
		return
	}

	s.SetConfig(cfg)
	s.handleGetSettings(w, r)
}

// applyTo validates the settings and returns a copy of cfg with them applied.
func (s *Server) applyTo(cfg *config.Config, settings settingsDTO) (*config.Config, error) {
	location := settings.Location
	if location.Latitude < -90 || location.Latitude > 90 {
		return nil, errors.New("The latitude must be between -90 and 90.")
	}
	if location.Longitude < -180 || location.Longitude > 180 {
		return nil, errors.New("The longitude must be between -180 and 180.")
	}
	if len(settings.Lights) == 0 {
		return nil, errors.New("Please select at least one light.")
	}

	seen := make(map[string]bool, len(settings.Lights))
	lights := make([]config.Light, 0, len(settings.Lights))
	for _, light := range settings.Lights {
		id, name := strings.TrimSpace(light.ID), strings.TrimSpace(light.Name)
		if id == "" {
			return nil, fmt.Errorf("The light %q has no ID.", name)
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		lights = append(lights, config.Light{ID: &id, Name: &name})
	}

	cfg.Location = config.Location{Latitude: location.Latitude, Longitude: location.Longitude}
	cfg.Lights = lights
	return cfg, nil
}

// bridgeLights returns the lights known to the bridge, sorted by name.
func (s *Server) bridgeLights() ([]lightDTO, error) {
	list, err := s.lights.GetAllLights()
	if err != nil {
		return nil, err
	}
	if len(list.Errors) > 0 {
		return nil, fmt.Errorf("bridge returned error: %s", list.Errors[0].Description)
	}

	lights := make([]lightDTO, 0, len(list.Data))
	for _, light := range list.Data {
		lights = append(lights, lightDTO{ID: light.ID, Name: light.Meta.Name})
	}
	sort.Slice(lights, func(i, j int) bool {
		return strings.ToLower(lights[i].Name) < strings.ToLower(lights[j].Name)
	})
	return lights, nil
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Ensure geo.Default satisfies LocationProvider.
var _ LocationProvider = geo.Default
