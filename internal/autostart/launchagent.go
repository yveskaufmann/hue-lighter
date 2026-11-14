package autostart

import (
	"bytes"
	"encoding/xml"
	"strings"
)

// LaunchAgentLabel is the launchd label of the macOS LaunchAgent. The Homebrew
// cask and scripts/uninstall.sh refer to it as well.
const LaunchAgentLabel = "com.github.yveskaufmann.hue-lighter"

// appBundleExecutableMarker identifies executables inside a macOS app bundle.
const appBundleExecutableMarker = ".app/Contents/MacOS/"

// isAppBundleExecutable reports whether executable lives inside an app bundle.
func isAppBundleExecutable(executable string) bool {
	return strings.Contains(executable, appBundleExecutableMarker)
}

// launchAgentPlist renders the LaunchAgent starting executable on login.
//
// The executable is started directly (not via `open`) so launchd supervises
// the app itself: it is restarted after a crash, but not after a regular quit
// via the tray menu (exit code 0). Logs are written to logFile.
func launchAgentPlist(executable string, logFile string) []byte {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>` + LaunchAgentLabel + `</string>
	<key>ProgramArguments</key>
	<array>
		<string>`)
	xmlEscape(&b, executable)
	b.WriteString(`</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>LimitLoadToSessionType</key>
	<string>Aqua</string>
	<key>KeepAlive</key>
	<dict>
		<key>SuccessfulExit</key>
		<false/>
	</dict>
	<key>ProcessType</key>
	<string>Interactive</string>
	<key>StandardOutPath</key>
	<string>`)
	xmlEscape(&b, logFile)
	b.WriteString(`</string>
	<key>StandardErrorPath</key>
	<string>`)
	xmlEscape(&b, logFile)
	b.WriteString(`</string>
</dict>
</plist>
`)
	return b.Bytes()
}

func xmlEscape(b *bytes.Buffer, s string) {
	// xml.EscapeText only fails if the writer fails, which bytes.Buffer never does.
	_ = xml.EscapeText(b, []byte(s))
}
