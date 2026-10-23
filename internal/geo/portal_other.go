//go:build !linux

package geo

import (
	"context"
	"errors"
)

// locationServiceLocation is only implemented on Linux (XDG Location portal);
// macOS' CoreLocation is not reachable without CGO.
func locationServiceLocation(ctx context.Context) (Location, error) {
	return Location{}, errors.New("location service is not supported on this platform")
}
