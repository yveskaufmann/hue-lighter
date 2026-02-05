package device_registration

import (
	"fmt"
	"time"

	hueclient "com.github.yveskaufmann/hue-lighter/internal/hue_client"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	client      *hueclient.Client
	apiKeyStore hueclient.APIKeyStore
	logger      *log.Entry
}

func NewService(client *hueclient.Client, apiKeyStore hueclient.APIKeyStore, logger *log.Entry) *Service {
	return &Service{
		client:      client,
		apiKeyStore: apiKeyStore,
		logger:      logger.WithField("component", "RegisterService"),
	}
}

func (s *Service) RegisterDevice(deviceName string) error {

	logger := s.logger.WithFields(log.Fields{
		"device": deviceName,
		"bridge": s.client.BridgeID(),
	})

	apiKeyIdentifier := fmt.Sprintf("%s#%s", s.client.BridgeID(), deviceName)
	if key, _ := s.apiKeyStore.Get(apiKeyIdentifier); key != "" {
		s.logger.Info("Device is already registered, skipping registration")
		return nil
	}

	// TODO: Check if device is already registered

	logger.Info("Registering device...")
	logger.Info("Press the link button on your Philips Hue bridge within the next 15 seconds!")

	<-time.After(15 * time.Second)
	// TODO: The username is the API key
	registerResponse, err := s.client.RegisterDevice(deviceName)
	if err != nil {
		logger.WithError(err).Error("Failed to invoke device registration API call")
		return err
	}

	if registerResponse.HasError() {
		logger.WithError(registerResponse.ToError()).Error("Device registration failed")
		if registerResponse.Error.Type == hueclient.HueErrorTypeLinkButtonNotPressed {
			logger.Error("Link button was not pressed on the Hue Bridge, please try again.")
		}
		return registerResponse.ToError()
	}

	logger.WithFields(log.Fields{"ClientKey": registerResponse.Success.ClientKey}).Info("Device registered successfully")

	err = s.apiKeyStore.Set(fmt.Sprintf("%s#%s", s.client.BridgeID(), s.client.DeviceName()), registerResponse.Success.Username)
	if err != nil {
		logger.WithError(err).Error("Failed to store API key")
		return err
	}

	logger.Info("Successfully registered device")

	return nil
}
