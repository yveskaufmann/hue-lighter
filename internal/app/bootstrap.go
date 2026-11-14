package app

import (
	"com.github.yveskaufmann/hue-lighter/internal/config"
	"com.github.yveskaufmann/hue-lighter/internal/geo"
	hueclient "com.github.yveskaufmann/hue-lighter/internal/hue_client"
	"com.github.yveskaufmann/hue-lighter/internal/logging"
	"com.github.yveskaufmann/hue-lighter/internal/messages"
	"com.github.yveskaufmann/hue-lighter/internal/services/device_registration"
	"com.github.yveskaufmann/hue-lighter/internal/services/light_automation"
	"com.github.yveskaufmann/hue-lighter/internal/settings"
	"com.github.yveskaufmann/hue-lighter/internal/tray"
)

// Bootstrap creates the application. withTray enables the tray icon.
func Bootstrap(withTray bool) *App {
	logger := logging.NewLogger().WithField("component", "app")

	// Desktop sessions (with tray) start with a default config that is completed
	// on the settings page; headless runs (e.g. containers) require a config.
	var cfg *config.Config
	var configCreated bool
	var err error
	if withTray {
		cfg, configCreated, err = config.LoadOrCreateDefault()
	} else {
		cfg, err = config.LoadConfigFromDefaultPath()
	}
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}
	if configCreated {
		logger.Info("No config found, created a default config, opening the settings page")
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

	client, err := hueclient.NewClient(cfg.Meta.Name, bridge.ID, bridge.IP, store, caPEM, logger)
	if err != nil {
		logger.Fatalf("Failed to create Hue client: %v", err)
	}

	registerService := device_registration.NewService(client, store, logger)
	var trayIcon *tray.Tray
	var settingsServer *settings.Server
	var commands chan messages.Command
	var lightEvents chan messages.LightStateChangedEvent
	if withTray {
		commands = make(chan messages.Command, 1)
		lightEvents = make(chan messages.LightStateChangedEvent, 1)
		trayIcon = tray.New(logger, commands, lightEvents)
		settingsServer = settings.NewServer(client, geo.Default, commands, logger)
		settingsServer.SetConfig(cfg)
		if configCreated {
			// Handled once the app runs (buffered channel, nothing sent before).
			commands <- messages.OpenSettingsCommand{}
		}
	}

	lightService := light_automation.NewService(client, cfg, logger, lightEvents)

	return &App{
		logger:          logger,
		registerService: registerService,
		client:          client,
		lightService:    lightService,
		tray:            trayIcon,
		settings:        settingsServer,
		commands:        commands,
		config:          cfg,
		stopChn:         make(chan struct{}),
	}
}
