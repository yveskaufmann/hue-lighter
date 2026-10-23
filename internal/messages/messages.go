// Package messages contains the commands and events exchanged between the
// application and its components (e.g. the tray icon) via channels.
package messages

// Command is sent to the application, e.g. by the tray icon. The application
// handles it by triggering the actual action on the responsible service.
type Command int

const (
	// ToggleLightsCommand switches the configured lights on or off regardless
	// of the sunrise/sunset times.
	ToggleLightsCommand Command = iota
	// StopCommand stops the application as if it received a stop signal.
	StopCommand
)

// LightStateChangedEvent is published whenever the configured lights are
// assumed to have been switched on or off.
type LightStateChangedEvent struct {
	On bool
}
