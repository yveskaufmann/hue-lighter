package app

import (
	"com.github.yveskaufmann/hue-lighter/internal/config"
	hueclient "com.github.yveskaufmann/hue-lighter/internal/hue_client"
	"com.github.yveskaufmann/hue-lighter/internal/logging"
	"com.github.yveskaufmann/hue-lighter/internal/messages"
	"com.github.yveskaufmann/hue-lighter/internal/services/device_registration"
	"com.github.yveskaufmann/hue-lighter/internal/services/light_automation"
	"com.github.yveskaufmann/hue-lighter/internal/tray"
)

// Bootstrap creates the application. withTray enables the tray icon.
func Bootstrap(withTray bool) *App {
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
	caPEM, caSource, err := hueclient.LoadCABundle()
	if err != nil {
		logger.Fatalf("CA bundle check failed: %v", err)
	}
	logger.Infof("Using CA bundle: %s", caSource)

	discoveryService := hueclient.NewBridgeDiscoveryService(logger)
	bridge, err := discoveryService.DiscoverFirstBridge(logger)
	if err != nil {
		logger.Fatalf("Failed to discover Hue Bridge: %v", err)
	}
	logger.Infof("Discovered Hue Bridge at IP: %s", bridge.IP)

	client, err := hueclient.NewClient(config.Meta.Name, bridge.ID, bridge.IP, store, caPEM, logger)
	if err != nil {
		logger.Fatalf("Failed to create Hue client: %v", err)
	}

	registerService := device_registration.NewService(client, store, logger)
	var trayIcon *tray.Tray
	var commands chan messages.Command
	var lightEvents chan messages.LightStateChangedEvent
	if withTray {
		commands = make(chan messages.Command, 1)
		lightEvents = make(chan messages.LightStateChangedEvent, 1)
		trayIcon = tray.New(logger, commands, lightEvents)
	}

	lightService := light_automation.NewService(client, config, logger, lightEvents)

	return &App{
		logger:          logger,
		registerService: registerService,
		client:          client,
		lightService:    lightService,
		tray:            trayIcon,
		commands:        commands,
		config:          config,
	}
}
