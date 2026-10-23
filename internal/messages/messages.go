// Package messages contains the commands and events exchanged between the
// application and its components (e.g. the tray icon) via channels.
package messages

import "com.github.yveskaufmann/hue-lighter/internal/config"

// Command is sent to the application, e.g. by the tray icon or the settings
// page. The application handles it by triggering the actual action on the
// responsible service.
type Command interface {
	isCommand()
}

// ToggleLightsCommand switches the configured lights on or off regardless of
// the sunrise/sunset times.
type ToggleLightsCommand struct{}

// StopCommand stops the application as if it received a stop signal.
type StopCommand struct{}

// OpenSettingsCommand opens the settings page.
type OpenSettingsCommand struct{}

// ApplySettingsCommand saves the given config and applies it to the running
// services. The result is sent to Reply, which must be buffered.
type ApplySettingsCommand struct {
	Config *config.Config
	Reply  chan<- error
}

func (ToggleLightsCommand) isCommand()  {}
func (StopCommand) isCommand()          {}
func (OpenSettingsCommand) isCommand()  {}
func (ApplySettingsCommand) isCommand() {}

// LightStateChangedEvent is published whenever the configured lights are
// assumed to have been switched on or off.
type LightStateChangedEvent struct {
	On bool
}
