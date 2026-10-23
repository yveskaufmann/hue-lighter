package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"com.github.yveskaufmann/hue-lighter/internal/autostart"
	"com.github.yveskaufmann/hue-lighter/internal/config"
	hueclient "com.github.yveskaufmann/hue-lighter/internal/hue_client"
	"com.github.yveskaufmann/hue-lighter/internal/messages"
	"com.github.yveskaufmann/hue-lighter/internal/services/device_registration"
	"com.github.yveskaufmann/hue-lighter/internal/services/light_automation"
	"com.github.yveskaufmann/hue-lighter/internal/tray"
	log "github.com/sirupsen/logrus"
)

type App struct {
	logger          *log.Entry
	registerService *device_registration.Service
	lightService    *light_automation.Service
	tray            *tray.Tray
	commands        <-chan messages.Command
	client          *hueclient.Client
	config          *config.Config
}

func (a *App) Logger() *log.Entry {
	return a.logger
}

// Run starts the application and blocks until SIGTERM, SIGINT or SIGHUP is
// received or a StopCommand was sent (e.g. by the tray icon). The configured lights are
// turned off before Run returns.
func (a *App) Run() error {

	err := a.registerService.RegisterDevice(a.client.DeviceName())
	if err != nil {
		return fmt.Errorf("failed to register device: %w", err)
	}
	a.logger.Infof("Device registered successfully")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer signal.Stop(signalChan)

	if err := a.lightService.Start(); err != nil {
		return fmt.Errorf("failed to start light automation service: %w", err)
	}
	a.logger.Info("Light automation service started successfully")

	if a.tray != nil {
		go func() {
			if err := a.tray.Run(); err != nil {
				a.logger.WithError(err).Error("Tray icon stopped unexpectedly")
			}
		}()
	}

	// Only desktop sessions (with tray) are started on login, not headless runs.
	autostartCtx, cancelAutostart := context.WithCancel(context.Background())
	autostartDone := make(chan struct{})
	if a.tray != nil {
		go func() {
			defer close(autostartDone)
			autostart.Enable(autostartCtx, a.logger)
		}()
	} else {
		close(autostartDone)
	}

	a.handleCommandsUntilStopped(signalChan)

	// Stop waiting for the Background portal and close its D-Bus connection.
	cancelAutostart()
	<-autostartDone

	return a.Stop()
}

// handleCommandsUntilStopped handles commands (e.g. from the tray icon) by
// triggering the actual actions on the services, until a stop signal or a
// StopCommand is received. Without tray icon, commands is nil and only signals
// are received.
func (a *App) handleCommandsUntilStopped(signalChan <-chan os.Signal) {
	for {
		select {
		case sig := <-signalChan:
			a.logger.Infof("Received %v signal, shutting down...", sig)
			return
		case cmd := <-a.commands:
			switch cmd {
			case messages.StopCommand:
				a.logger.Info("Received stop command, shutting down...")
				return
			case messages.ToggleLightsCommand:
				a.lightService.ToggleLights()
			default:
				a.logger.Warnf("Received unknown command: %v", cmd)
			}
		}
	}
}

// Stop stops all services and turns off the configured lights.
func (a *App) Stop() error {
	a.logger.Info("Stopping application")

	if a.tray != nil {
		a.tray.Stop()
	}

	if err := a.lightService.StopAndTurnOffLights(); err != nil {
		return fmt.Errorf("failed to turn off lights: %w", err)
	}

	return nil
}
