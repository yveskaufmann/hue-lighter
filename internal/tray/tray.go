package tray

import (
	_ "embed"

	"com.github.yveskaufmann/hue-lighter/internal/messages"
	"github.com/gogpu/systray"
	log "github.com/sirupsen/logrus"
)

// IconSVG is the tray icon as SVG, kept for platforms/themes that accept vector icons.
//
//go:embed assets/icon.svg
var IconSVG []byte

// IconPNG is the tray icon as PNG, which is the format systray expects.
//
//go:embed assets/icon.png
var IconPNG []byte

// PowerIconSVG is the icon of the menu item toggling the lights, as SVG.
//
//go:embed assets/power.svg
var PowerIconSVG []byte

// PowerIconPNG is the icon of the menu item toggling the lights, as PNG.
//
//go:embed assets/power.png
var PowerIconPNG []byte

// Menu labels describe what clicking the lights menu item does.
const (
	labelTurnOn  = "Turn Lights On"
	labelTurnOff = "Turn Lights Off"
	labelQuit    = "Quit"
)

// Tray shows a system tray icon which allows toggling the configured lights
// regardless of the sunrise/sunset times. It sends commands to the application
// and only learns about the light state from the light automation's events.
type Tray struct {
	logger     *log.Entry
	commands   chan<- messages.Command
	events     <-chan messages.LightStateChangedEvent
	tray       *systray.SystemTray
	toggleItem *systray.MenuItem
	stopChn    chan struct{}
}

func New(logger *log.Entry, commands chan<- messages.Command, events <-chan messages.LightStateChangedEvent) *Tray {
	return &Tray{
		logger:   logger.WithField("component", "Tray"),
		commands: commands,
		events:   events,
		stopChn:  make(chan struct{}),
	}
}

// Run shows the tray icon and blocks until Stop is called.
func (t *Tray) Run() error {
	t.tray = systray.New()

	menu := systray.NewMenu()
	// The label is only updated from LightStateChangedEvents, so it always
	// matches the state reported by the light automation service. It stays
	// disabled until the service reported the light state for the first time.
	t.toggleItem = menu.AddWithIcon(labelTurnOn, PowerIconPNG, t.toggleLights)
	t.toggleItem.SetDisabled(true)
	menu.AddSeparator()
	menu.Add(labelQuit, t.quit)

	t.tray.SetIcon(IconPNG).
		SetAppName("hue-lighter").
		SetTooltip("hue-lighter").
		SetMenu(menu)
	t.tray.Show()

	go t.handleEvents()

	t.logger.Info("Tray icon started")
	return t.tray.Run()
}

// Stop removes the tray icon and unblocks Run.
func (t *Tray) Stop() {
	t.logger.Info("Stopping tray icon")
	close(t.stopChn)
	if t.tray != nil {
		t.tray.Remove()
	}
}

func (t *Tray) quit() {
	t.logger.Info("Quit requested via tray icon")
	// Unlike toggling, stopping must not be dropped; the app already stopping
	// (and no longer reading commands) is signaled via stopChn.
	select {
	case t.commands <- messages.StopCommand:
	case <-t.stopChn:
	}
}

func (t *Tray) toggleLights() {
	select {
	case t.commands <- messages.ToggleLightsCommand:
	default:
		t.logger.Debug("Toggle already pending, ignoring click")
	}
}

func (t *Tray) handleEvents() {
	for {
		select {
		case event := <-t.events:
			t.setLightState(event.On)
		case <-t.stopChn:
			return
		}
	}
}

func (t *Tray) setLightState(on bool) {
	t.toggleItem.SetDisabled(false)
	if on {
		t.toggleItem.SetLabel(labelTurnOff)
		t.tray.SetTooltip("hue-lighter: lights on")
	} else {
		t.toggleItem.SetLabel(labelTurnOn)
		t.tray.SetTooltip("hue-lighter: lights off")
	}
}
