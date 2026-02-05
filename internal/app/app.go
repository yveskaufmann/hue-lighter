package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"com.github.yveskaufmann/hue-lighter/internal/config"
	hueclient "com.github.yveskaufmann/hue-lighter/internal/hue_client"
	"com.github.yveskaufmann/hue-lighter/internal/services/device_registration"
	"com.github.yveskaufmann/hue-lighter/internal/services/events"
	"com.github.yveskaufmann/hue-lighter/internal/services/light_automation"
	log "github.com/sirupsen/logrus"
)

type App struct {
	logger          *log.Entry
	registerService *device_registration.Service
	lightService    *light_automation.Service
	eventService    *events.ExternalEventService
	client          *hueclient.Client
	config          *config.Config
	StopChn         chan struct{}
}

func (a *App) Logger() *log.Entry {
	return a.logger
}

func (a *App) EventService() *events.ExternalEventService {
	return a.eventService
}

func (a *App) Run() error {
	a.logger.Info("Starting application")

	err := a.registerService.RegisterDevice(a.client.DeviceName())
	if err != nil {
		return fmt.Errorf("failed to register device: %w", err)
	}

	if err := a.lightService.Start(); err != nil {
		return fmt.Errorf("failed to start light automation service: %w", err)
	}

	if err := a.eventService.Start(); err != nil {
		return fmt.Errorf("failed to start event service: %w", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

eventLoop:
	for {
		select {
		case <-signalChan:
			a.logger.Info("Received interrupt signal, shutting down...")
			break eventLoop
		case <-a.StopChn:
			a.logger.Info("Received stop signal, shutting down...")
			break eventLoop
		}
	}

	close(signalChan)
	close(a.StopChn)

	a.Stop()

	return nil
}

func (a *App) Stop() error {
	a.logger.Info("Stopping application")

	a.lightService.Stop()
	a.eventService.Stop()

	return nil
}

func (a *App) SendShutdownEvent() error {

	a.logger.Info("Starting application")
	err := a.registerService.RegisterDevice(a.client.DeviceName())
	if err != nil {
		return fmt.Errorf("failed to register device: %w", err)
	}

	defer a.logger.Info("Shutdown event sent, exiting application")

	return a.eventService.StopAndTurnOffLights()
}
