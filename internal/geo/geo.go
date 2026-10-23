// Package geo determines a default location for the sunrise/sunset
// calculation: the precise location from the OS location service if the user
// allows it, otherwise the reference city of the system's time zone.
package geo

import (
	"context"
	"errors"
	"math"

	log "github.com/sirupsen/logrus"
)

// Source describes where a location was determined from.
type Source string

const (
	SourceLocationService Source = "location-service"
	SourceTimeZone        Source = "time-zone"
)

// Location is a geographic position in decimal degrees.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Source    Source  `json:"source"`
}

// ErrNoLocation is returned when no location could be determined.
var ErrNoLocation = errors.New("location could not be determined")

// Default returns the best available location: the OS location service
// first, then the reference city of the system time zone.
func Default(ctx context.Context, logger *log.Entry) (Location, error) {
	logger = logger.WithField("component", "Geo")

	if location, err := locationServiceLocation(ctx); err == nil {
		return rounded(location), nil
	} else {
		logger.WithError(err).Info("Location service unavailable, falling back to time zone")
	}

	location, err := timeZoneLocation()
	if err != nil {
		logger.WithError(err).Warn("Failed to determine location from time zone")
		return Location{}, ErrNoLocation
	}
	return rounded(location), nil
}

// rounded limits the precision to about 10 m, more is not needed for the
// sunrise/sunset calculation.
func rounded(location Location) Location {
	location.Latitude = math.Round(location.Latitude*1e4) / 1e4
	location.Longitude = math.Round(location.Longitude*1e4) / 1e4
	return location
}
