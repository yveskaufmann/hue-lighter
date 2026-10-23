// Package autostart registers hue-lighter to be started on login.
//
// Inside Flatpak this is done via the XDG Background portal, which asks the
// user once and then creates the autostart entry itself. Outside of Flatpak
// it does nothing.
package autostart

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	log "github.com/sirupsen/logrus"
)

const (
	portalBusName       = "org.freedesktop.portal.Desktop"
	portalObjectPath    = "/org/freedesktop/portal/desktop"
	backgroundInterface = "org.freedesktop.portal.Background"
	requestInterface    = "org.freedesktop.portal.Request"
	flatpakInfoPath     = "/.flatpak-info"

	handleToken = "hue_lighter_autostart"
	// The user may have to confirm a dialog, so give them time.
	responseTimeout = 5 * time.Minute
)

// Enable requests the Background portal to start hue-lighter on login. It
// blocks until the portal answered or ctx is cancelled, so callers usually run
// it in a goroutine and cancel ctx when the application stops.
func Enable(ctx context.Context, logger *log.Entry) {
	logger = logger.WithField("component", "Autostart")

	if _, err := os.Stat(flatpakInfoPath); err != nil {
		logger.Debug("Not running inside Flatpak, skipping autostart registration")
		return
	}

	if err := requestBackground(ctx, logger); err != nil {
		logger.WithError(err).Warn("Failed to register autostart")
	}
}

func requestBackground(ctx context.Context, logger *log.Entry) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}
	defer conn.Close()

	// Subscribe to the Response signal before sending the request to not miss
	// it; the request object path is derived from our unique name and token.
	sender := strings.ReplaceAll(strings.TrimPrefix(conn.Names()[0], ":"), ".", "_")
	requestPath := dbus.ObjectPath(fmt.Sprintf("%s/request/%s/%s", portalObjectPath, sender, handleToken))

	if err := conn.AddMatchSignal(
		dbus.WithMatchObjectPath(requestPath),
		dbus.WithMatchInterface(requestInterface),
		dbus.WithMatchMember("Response"),
	); err != nil {
		return fmt.Errorf("failed to subscribe to portal response: %w", err)
	}

	signals := make(chan *dbus.Signal, 1)
	conn.Signal(signals)

	options := map[string]dbus.Variant{
		"handle_token":     dbus.MakeVariant(handleToken),
		"reason":           dbus.MakeVariant("Automatically control your Hue lights based on sunrise and sunset."),
		"autostart":        dbus.MakeVariant(true),
		"commandline":      dbus.MakeVariant([]string{"hue-lighter"}),
		"dbus-activatable": dbus.MakeVariant(false),
	}

	var handle dbus.ObjectPath
	portal := conn.Object(portalBusName, portalObjectPath)
	if err := portal.CallWithContext(ctx, backgroundInterface+".RequestBackground", 0, "", options).Store(&handle); err != nil {
		if ctx.Err() != nil {
			logger.Debug("Application is stopping, cancelled autostart registration")
			return nil
		}
		return fmt.Errorf("RequestBackground call failed: %w", err)
	}

	timeout := time.After(responseTimeout)
	for {
		select {
		case signal := <-signals:
			if signal.Path != handle || len(signal.Body) < 2 {
				continue
			}
			return handleResponse(logger, signal.Body)
		case <-timeout:
			return fmt.Errorf("no response from Background portal within %v", responseTimeout)
		case <-ctx.Done():
			logger.Debug("Application is stopping, stopped waiting for the Background portal")
			return nil
		}
	}
}

func handleResponse(logger *log.Entry, body []interface{}) error {
	response, _ := body[0].(uint32)
	results, _ := body[1].(map[string]dbus.Variant)

	if response != 0 {
		return fmt.Errorf("background request was denied or cancelled (response = %d)", response)
	}

	autostart, _ := results["autostart"].Value().(bool)
	if !autostart {
		logger.Warn("Autostart was not granted, hue-lighter will not start on login")
		return nil
	}

	logger.Info("Autostart on login enabled")
	return nil
}
