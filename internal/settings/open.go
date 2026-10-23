package settings

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/godbus/dbus/v5"
)

const flatpakInfoPath = "/.flatpak-info"

// OpenURL opens url in the user's default browser. Inside Flatpak this goes
// through the OpenURI portal, as host programs like xdg-open are not available.
func OpenURL(url string) error {
	if _, err := os.Stat(flatpakInfoPath); err == nil {
		return openViaPortal(url)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open %s: %w", cmd.Path, err)
	}
	// Reap the process without blocking the caller.
	go cmd.Wait()
	return nil
}

func openViaPortal(url string) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}
	defer conn.Close()

	portal := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	call := portal.Call("org.freedesktop.portal.OpenURI.OpenURI", 0, "", url, map[string]dbus.Variant{})
	if call.Err != nil {
		return fmt.Errorf("OpenURI portal call failed: %w", call.Err)
	}
	return nil
}
