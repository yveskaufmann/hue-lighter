package light_automation

// Command is sent to the light automation service, e.g. by the tray icon.
type Command int

const (
	// ToggleLightsCommand switches the configured lights on or off regardless
	// of the sunrise/sunset times.
	ToggleLightsCommand Command = iota
)

// LightStateChangedEvent is published whenever the configured lights are
// assumed to have been switched on or off.
type LightStateChangedEvent struct {
	On bool
}
