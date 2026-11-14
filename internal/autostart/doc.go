// Package autostart registers hue-lighter to be started on login.
//
//   - Linux: inside Flatpak this is done via the XDG Background portal, which
//     asks the user once and then creates the autostart entry itself. Outside
//     of Flatpak it does nothing.
//   - macOS: when running from the Hue Lighter.app bundle, a per-user
//     LaunchAgent (~/Library/LaunchAgents) is written that starts the app on
//     login. Outside of an app bundle (e.g. `go run`) it does nothing.
//   - Other platforms: does nothing.
package autostart
