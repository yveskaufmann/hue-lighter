# Template for Casks/hue-lighter.rb in the yveskaufmann/homebrew-tap repository.
# The version and sha256 placeholders are filled in by .github/workflows/release.yaml.
cask "hue-lighter" do
  version "@VERSION@"
  sha256 "@SHA256@"

  url "https://github.com/yveskaufmann/hue-lighter/releases/download/v#{version}/hue-lighter_#{version}_macos_universal.zip"
  name "Hue Lighter"
  desc "Turns Philips Hue lights on at sunset and off at sunrise, logout and shutdown"
  homepage "https://github.com/yveskaufmann/hue-lighter"

  livecheck do
    url :url
    strategy :github_latest
  end

  depends_on macos: :big_sur

  app "Hue Lighter.app"

  # The app is only ad-hoc signed (no Apple Developer ID), so Gatekeeper would
  # block it. Trusting this tap means trusting the app, so lift the quarantine.
  postflight do
    system_command "/usr/bin/xattr",
                   args: ["-dr", "com.apple.quarantine", "#{appdir}/Hue Lighter.app"]
  end

  # Stops the app gracefully (SIGTERM turns the lights off) and removes the
  # LaunchAgent the app registers for starting on login.
  uninstall launchctl: "com.github.yveskaufmann.hue-lighter",
            signal:    [["TERM", "com.github.yveskaufmann.hue-lighter"]]

  zap trash: [
    "~/Library/Application Support/hue-lighter",
    "~/Library/Logs/hue-lighter.log",
  ]

  caveats <<~EOS
    Start Hue Lighter once from /Applications, it then starts on every login.
    Config: ~/Library/Application Support/hue-lighter/config.yaml
  EOS
end
