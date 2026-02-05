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
	s.logger.Info("Running automation ticker loop")

	defer s.Stop()

	s.refreshLightStates()

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

	s.logger.Infof("Tick at %v", tickTime)

	if time.Since(s.lastLightStateRefresh) > 5*time.Minute {
		s.refreshLightStates()
	}

	sunriseTime, sunsetTime := sunset.CalculateSunriseSunset(s.config.Location.Latitude, s.config.Location.Longitude)

	s.logger.Infof("Sunrise at %v, Sunset at %v", sunriseTime, sunsetTime)
	isNight := tickTime.Before(sunriseTime) || tickTime.After(sunsetTime)
	// Only attempt to enable lights when both conditions are met:
	//  - tickTime is at night between sunset and next day's sunrise
	if isNight {
		s.setLightsState(true)

	} else {
		s.setLightsState(false)
	}
}

func (s *Service) setLightsState(turnOn bool) {
	for _, lightCfg := range s.config.Lights {
		if turnOn {
			s.logger.Info("It's nighttime and we've reached lights on time, turning on lights")

			if s.lightStates[*lightCfg.ID] {
				s.logger.Infof("Light ID: %s is already on, skipping", *lightCfg.ID)
				continue
			}

			err := s.client.TurnOnLightById(*lightCfg.ID)
			if err != nil {
				s.logger.Errorf("Failed to turn on light ID: %s, error: %v", *lightCfg.ID, err)
			}

			s.lightStates[*lightCfg.ID] = true
		} else {
			s.logger.Info("It's daytime, lights should remain off")

			if !s.lightStates[*lightCfg.ID] {
				s.logger.Infof("Light ID: %s is already off, skipping", *lightCfg.ID)
				continue
			}

			err := s.client.TurnOffLightById(*lightCfg.ID)
			if err != nil {
				s.logger.Errorf("Failed to turn off light ID: %s, error: %v", *lightCfg.ID, err)
			}
			s.lightStates[*lightCfg.ID] = false
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
	s.Stop()
	s.setLightsState(false)
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
