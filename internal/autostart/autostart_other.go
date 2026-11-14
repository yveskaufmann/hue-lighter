//go:build !linux && !darwin

package autostart

import (
	"context"

	log "github.com/sirupsen/logrus"
)

// Enable does nothing on this platform.
func Enable(_ context.Context, logger *log.Entry) {
	logger.WithField("component", "Autostart").Debug("Autostart is not supported on this platform")
}
