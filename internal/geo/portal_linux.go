//go:build linux

package geo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	portalBusName     = "org.freedesktop.portal.Desktop"
	portalObjectPath  = "/org/freedesktop/portal/desktop"
	locationInterface = "org.freedesktop.portal.Location"
	requestInterface  = "org.freedesktop.portal.Request"
	sessionInterface  = "org.freedesktop.portal.Session"

	sessionToken = "hue_lighter_location"
	requestToken = "hue_lighter_location_start"
	// accuracyCity is enough for the sunrise/sunset calculation and the least
	// privacy invasive accuracy which still makes sense.
	accuracyCity = uint32(2)
	// The user may have to confirm a dialog, so give them time.
	locationTimeout = 60 * time.Second
)

// locationServiceLocation asks the XDG Location portal (GeoClue) for the
// current location. It works inside Flatpak and on the host; the user or a
// disabled location service may deny it.
func locationServiceLocation(ctx context.Context) (Location, error) {
	ctx, cancel := context.WithTimeout(ctx, locationTimeout)
	defer cancel()

	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return Location{}, fmt.Errorf("failed to connect to session bus: %w", err)
	}
	defer conn.Close()

	// The request object path is derived from our unique name and token;
	// subscribe before calling to not miss the signals.
	sender := strings.ReplaceAll(strings.TrimPrefix(conn.Names()[0], ":"), ".", "_")
	requestPath := dbus.ObjectPath(fmt.Sprintf("%s/request/%s/%s", portalObjectPath, sender, requestToken))

	for _, match := range [][]dbus.MatchOption{
		{dbus.WithMatchInterface(locationInterface), dbus.WithMatchMember("LocationUpdated")},
		{dbus.WithMatchObjectPath(requestPath), dbus.WithMatchInterface(requestInterface), dbus.WithMatchMember("Response")},
	} {
		if err := conn.AddMatchSignal(match...); err != nil {
			return Location{}, fmt.Errorf("failed to subscribe to location portal signals: %w", err)
		}
	}
	signals := make(chan *dbus.Signal, 4)
	conn.Signal(signals)

	portal := conn.Object(portalBusName, portalObjectPath)

	var session dbus.ObjectPath
	if err := portal.CallWithContext(ctx, locationInterface+".CreateSession", 0, map[string]dbus.Variant{
		"session_handle_token": dbus.MakeVariant(sessionToken),
		"accuracy":             dbus.MakeVariant(accuracyCity),
	}).Store(&session); err != nil {
		return Location{}, fmt.Errorf("CreateSession failed: %w", err)
	}
	defer conn.Object(portalBusName, session).Call(sessionInterface+".Close", 0)

	var request dbus.ObjectPath
	if err := portal.CallWithContext(ctx, locationInterface+".Start", 0, session, "", map[string]dbus.Variant{
		"handle_token": dbus.MakeVariant(requestToken),
	}).Store(&request); err != nil {
		return Location{}, fmt.Errorf("Start failed: %w", err)
	}

	for {
		select {
		case signal := <-signals:
			switch {
			case signal.Name == requestInterface+".Response" && signal.Path == request:
				if response, _ := signal.Body[0].(uint32); response != 0 {
					return Location{}, fmt.Errorf("location access was denied or cancelled (response = %d)", response)
				}
			case signal.Name == locationInterface+".LocationUpdated" && len(signal.Body) >= 2:
				if path, _ := signal.Body[0].(dbus.ObjectPath); path != session {
					continue
				}
				return parseLocationUpdate(signal.Body[1])
			}
		case <-ctx.Done():
			return Location{}, fmt.Errorf("no location from location portal: %w", ctx.Err())
		}
	}
}

func parseLocationUpdate(body interface{}) (Location, error) {
	properties, ok := body.(map[string]dbus.Variant)
	if !ok {
		return Location{}, fmt.Errorf("unexpected LocationUpdated payload %T", body)
	}
	latitude, latOK := properties["Latitude"].Value().(float64)
	longitude, lonOK := properties["Longitude"].Value().(float64)
	if !latOK || !lonOK {
		return Location{}, fmt.Errorf("LocationUpdated without latitude/longitude")
	}
	return Location{Latitude: latitude, Longitude: longitude, Source: SourceLocationService}, nil
}
