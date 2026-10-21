package light_automation

import (
	"time"

	"com.github.yveskaufmann/hue-lighter/internal/config"
	"com.github.yveskaufmann/hue-lighter/internal/sunset"

	hueclient "com.github.yveskaufmann/hue-lighter/internal/hue_client"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	logger                *log.Entry
	client                *hueclient.Client
	config                *config.Config
	ticker                *time.Ticker
	tickerStop            chan struct{}
	lightStates           map[string]bool
	lastLightStateRefresh time.Time
}

func NewService(client *hueclient.Client, config *config.Config, logger *log.Entry) *Service {
	return &Service{
		logger:      logger.WithField("component", "LightAutomationService"),
		client:      client,
		config:      config,
		ticker:      nil,
		tickerStop:  make(chan struct{}),
		lightStates: make(map[string]bool),
	}
}

func (s *Service) Start() error {

	if s.ticker != nil {
		s.logger.Warn("Light Automation Service is already running")
		return nil
	}

	s.logger.Info("Starting Light Automation Service")
	s.ticker = time.NewTicker(1 * time.Second)
	go s.runAutomationTickerLoop()
	return nil

}

func (s *Service) runAutomationTickerLoop() {
	s.logger.Info("Started automation ticker loop")

	defer s.Stop()

	s.refreshLightStates()
	// Initial automation run to avoid delay during the first tick at the service start
	s.runAutomation()

	for {
		select {
		case <-s.ticker.C:
			s.runAutomation()
		case <-s.tickerStop:
			s.logger.Info("Stopping periodic tasks.")
			return
		}
	}

	// Example: Turn off all lights at midnight
}

func (s *Service) runAutomation() {
	tickTime := time.Now()

	s.logger.Debugf("Tick at %v", tickTime)

	if time.Since(s.lastLightStateRefresh) > 15*time.Second  {
		s.refreshLightStates()
	}

	sunriseTime, sunsetTime := sunset.CalculateSunriseSunset(s.config.Location.Latitude, s.config.Location.Longitude)

	s.logger.Debugf("Sunrise at %v, Sunset at %v", sunriseTime, sunsetTime)
	lightState := s.getLightState()
	isNight := tickTime.Before(sunriseTime) || tickTime.After(sunsetTime)
	// Only attempt to enable lights when both conditions are met:
	//  - tickTime is at night between sunset and next day's sunrise

	if isNight && !lightState {
		s.logger.Infof("Nighttime detected: turning lights ON (sunset: %v)", sunsetTime)
		s.setLightsState(true)

	} else if !isNight && lightState {
		s.logger.Infof("Daytime detected: turning lights OFF (sunrise: %v)", sunriseTime)
		s.setLightsState(false)
	}
}

// Determines if the lights were last known to be turned on
func (s *Service) getLightState() bool {
	for _, state := range s.lightStates {
		if !state {
			return false
		}
	}
	return true
}

func (s *Service) setLightsState(turnOn bool) {

	s.logger.Debugf("Setting lights state to %v", turnOn)
	s.logger.Debugf("Current light config: %v", s.config.Lights)

	for _, lightCfg := range s.config.Lights {
		if turnOn {

			if s.lightStates[*lightCfg.ID] {
				s.logger.Debugf("Light ID: %s is already on, skipping", *lightCfg.ID)
				continue
			}

			err := s.client.TurnOnLightById(*lightCfg.ID)
			if err != nil {
				s.logger.Errorf("Failed to turn on light ID: %s, error: %v", *lightCfg.ID, err)
			}

			s.lightStates[*lightCfg.ID] = true

			s.logger.Infof("Light ID: %s has been turned on", *lightCfg.ID)

		} else {

			if !s.lightStates[*lightCfg.ID] {
				s.logger.Debugf("Light ID: %s is already off, skipping", *lightCfg.ID)
				continue
			}

			err := s.client.TurnOffLightById(*lightCfg.ID)
			if err != nil {
				s.logger.Errorf("Failed to turn off light ID: %s, error: %v", *lightCfg.ID, err)
			}
			s.lightStates[*lightCfg.ID] = false

			s.logger.Infof("Light ID: %s has been turned off", *lightCfg.ID)
		}
	}
}

func (s *Service) refreshLightStates() {
	for _, lightCfg := range s.config.Lights {
		state, err := s.client.GetOneLightById(*lightCfg.ID)
		if err == nil {
			s.lightStates[*lightCfg.ID] = state.On.On
		} else {
			s.logger.Warnf("Could not refresh state for light %s: %v", *lightCfg.ID, err)
		}
	}

	s.lastLightStateRefresh = time.Now()
}

func (s *Service) StopAndTurnOffLights() error {
	s.setLightsState(false)
	s.Stop()
	return nil
}

func (s *Service) Stop() {
	if s.ticker == nil {
		s.logger.Warn("Light Automation Service is not running")
		return
	}

	s.logger.Info("Stopping Light Automation Service")

	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}
	close(s.tickerStop)
}
