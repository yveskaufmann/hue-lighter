package app

import (
	"com.github.yveskaufmann/hue-lighter/internal/config"
	hueclient "com.github.yveskaufmann/hue-lighter/internal/hue_client"
	"com.github.yveskaufmann/hue-lighter/internal/logging"
	"com.github.yveskaufmann/hue-lighter/internal/services/device_registration"
	"com.github.yveskaufmann/hue-lighter/internal/services/events"
	"com.github.yveskaufmann/hue-lighter/internal/services/light_automation"
)

func Bootstrap() *App {
	logger := logging.NewLogger().WithField("component", "app")

	config, err := config.LoadConfigFromDefaultPath()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	store, err := hueclient.NewAPIKeyStore(logger)
	if err != nil {
		logger.Fatalf("Failed to create API key store: %v", err)
	}

	// Verify CA bundle is present before attempting discovery or creating clients.
	certPath, err := hueclient.ResolveCABundlePath()
	if err != nil {
		logger.Fatalf("CA bundle check failed: %v", err)
	}
	logger.Infof("Using CA bundle: %s", certPath)

	discoveryService := hueclient.NewBridgeDiscoveryService(logger)
	bridge, err := discoveryService.DiscoverFirstBridge(logger)
	if err != nil {
		logger.Fatalf("Failed to discover Hue Bridge: %v", err)
	}
	logger.Infof("Discovered Hue Bridge at IP: %s", bridge.IP)

	stopChn := make(chan struct{})

	client, err := hueclient.NewClient(config.Meta.Name, bridge.ID, bridge.IP, store, certPath, logger)
	if err != nil {
		logger.Fatalf("Failed to create Hue client: %v", err)
	}

	registerService := device_registration.NewService(client, store, logger)
	lightService := light_automation.NewService(client, config, logger)
	eventService := events.NewExternalEventService(lightService, logger, stopChn)

	return &App{
		logger:          logger,
		registerService: registerService,
		client:          client,
		eventService:    eventService,
		lightService:    lightService,
		config:          config,
		StopChn:         stopChn,
	}
}
