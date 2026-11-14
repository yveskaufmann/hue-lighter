
# Hue Lighter

Hue Lighter is a small automation tool for managing Philips Hue lights, specifically designed to automate the backlighting of a desktop PC. It intelligently turns lights on at sunset, off at sunrise, and ensures they are turned off gracefully when you log out or the machine shuts down.

## Motivation

While Philips Hue already offers sunset and sunrise automation, I found it frustrating that my backlights would turn on even when I wasn't at my desk. I also wanted the lights to turn off automatically when I shut down my PC. This project was a fun side quest to solve those problems and experiment with automation beyond the built-in Hue features.

## Features

-   **Sunset/Sunrise Automation**: Automatically turns configured lights on at sunset and off at sunrise based on your geographic location.
-   **Tray Icon**: Switch the lights on or off manually regardless of the time of day. A manual switch is kept until the lights are back in the state the automation would choose.
-   **Graceful Shutdown**: Turns off all configured lights when you log out or shut down the machine, ensuring you don't leave them on by accident.
-   **Desktop App**: Runs in your desktop session and starts automatically on login, as a Flatpak on Linux and as a menu bar app on macOS.
-   **Automatic Registration**: On first run, it guides you through the simple process of registering the app with your Philips Hue Bridge by pressing the link button.

## Installation

### macOS (Homebrew)

Hue Lighter is distributed as `Hue Lighter.app` through the Homebrew tap [yveskaufmann/homebrew-tap](https://github.com/yveskaufmann/homebrew-tap). It requires macOS 11 (Big Sur) or newer and runs on Apple silicon and Intel Macs.

1. Register the tap. Homebrew only installs casks from taps you trust, so trust it as well:
   ```sh
   brew tap yveskaufmann/tap
   brew trust yveskaufmann/tap
   ```
2. Install the app:
   ```sh
   brew install --cask yveskaufmann/tap/hue-lighter
   ```
3. Start **Hue Lighter** once from `/Applications`. It shows up in the menu bar (there is no Dock icon), asks for access to the local network to find your Hue Bridge, and registers itself to start on every login.

The app is not signed with an Apple Developer ID. The cask removes the quarantine flag, so Gatekeeper doesn't block it.

-   **Update:** `brew upgrade --cask hue-lighter`
-   **Uninstall:** `brew uninstall --cask hue-lighter` stops the app (turning the lights off) and removes the login item. Add `--zap` to also remove the config and the log file.
-   **Logs:** `~/Library/Logs/hue-lighter.log`

If you installed an earlier version with the macOS installer package, remove it first with `scripts/uninstall.sh Darwin --legacy-only`. Otherwise Hue Lighter runs twice.

### Linux (release binary)

Every [release](https://github.com/yveskaufmann/hue-lighter/releases) has static binaries for `amd64` and `arm64`. Replace `<version>` with the release version (without the `v`):

```sh
curl -LO https://github.com/yveskaufmann/hue-lighter/releases/download/v<version>/hue-lighter_<version>_linux_amd64.tar.gz
sha256sum --check --ignore-missing <(curl -sL https://github.com/yveskaufmann/hue-lighter/releases/download/v<version>/SHA256SUMS)
tar -xzf hue-lighter_<version>_linux_amd64.tar.gz
install -Dm755 hue-lighter ~/.local/bin/hue-lighter
```

Then start `hue-lighter`. The release binary doesn't register itself to start on login, unlike the Flatpak. Add it to your desktop's autostart settings, or [build the Flatpak](#linux-flatpak) instead.

The tray icon needs a desktop with StatusNotifierItem support (on GNOME the AppIndicator extension, which Ubuntu ships by default).

### From source

#### Prerequisites

-   Go (version 1.25 or newer)
-   `make`
-   The Philips Hue CA bundle (see [CA bundle](#ca-bundle-required))
-   Linux: `flatpak` and `flatpak-builder` with the `org.freedesktop.Sdk//25.08` and `org.freedesktop.Sdk.Extension.golang//25.08` from Flathub.
-   macOS: the Xcode Command Line Tools (`xcode-select --install`).

#### Clone the repository

```sh
git clone https://github.com/yveskaufmann/hue-lighter.git
cd hue-lighter
```

#### CA bundle (required)

`hue-lighter` requires the Philips Hue CA certificate bundle to validate TLS connections to the Hue Bridge. This file is not stored in version control — you must obtain it yourself. All build targets embed it into the binary.

**Legal Note**: The CA bundle cannot be redistributed with this project as it would violate the [Philips Hue Terms of Use and Conditions](https://developers.meethue.com/terms-of-use-and-conditions/), which restrict the redistribution of Philips Hue materials without explicit permission. Each user must download the bundle directly from Philips. Because the bundle is embedded, do not share binaries or Flatpaks you built with it.

Follow these steps to provide the CA bundle:

1. Download the CA bundle from the [Philips Hue Developer Portal](https://developers.meethue.com/develop/application-design-guidance/using-https/).
2. Create the directory if it doesn't exist:
   ```sh
   mkdir -p configs/certs
   ```
3. Save the downloaded bundle to `configs/certs/cacert_bundle.pem`.

To use a bundle file instead of the embedded one (e.g. in a container), set the environment variable `HUE_CA_CERTS_PATH` to its path. If neither is available, the application terminates with an explanatory error.

#### Linux (Flatpak)

```sh
make flatpak
```

This command will:
1.  Copy the CA bundle so it gets embedded into the binary.
2.  Vendor the Go modules (the Flatpak build runs offline).
3.  Build `io.github.yveskaufmann.HueLighter` with `flatpak-builder` and install it for the current user.

Start it with `flatpak run io.github.yveskaufmann.HueLighter` (or `make flatpak-run`) or from your app launcher.

#### macOS (app bundle)

```sh
make mac_install
```

This builds a universal `Hue Lighter.app`, installs it to `/Applications` and starts it. `make mac_uninstall` removes it again; your config is kept.

To only build the app, run `make mac_app`. It creates `bin/macos/Hue Lighter.app` and the release zip in `dist/`.

#### Binary only

```sh
make build && ./bin/hue-lighter
```

## Configuration

When the desktop app starts for the first time, it creates a default config without any lights and opens the settings page, where you can set your location and pick your lights. You can also edit the file directly:

| Installation | Config file |
| --- | --- |
| macOS | `~/Library/Application Support/hue-lighter/config.yaml` |
| Linux (Flatpak) | `~/.var/app/io.github.yveskaufmann.HueLighter/config/hue-lighter/config.yaml` |
| Linux (binary) | `~/.config/hue-lighter/config.yaml` |

Set `CONFIG_PATH` to use a different file. With `--no-tray` (e.g. in Docker), no default is created and the config file must exist.

The config looks like this (see [configs/config.example.yaml](configs/config.example.yaml)):

```yaml
meta:
  version: 1
  name: "Hue Lighter Automation"
  description: "Configuration for Hue Lighter Automation"
location:
  # Your geographic location for sunset/sunrise calculation.
  latitude: 52.5208271
  longitude: 13.4093387
lights:
  # Add each light you want to automate here.
  # You can find the ID of your lights in the Philips Hue app
  # under Settings > Lights > (Select a light) > Info.
  - id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    name: "Office Hue Play Left"
  - id: "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy"
    name: "Office Hue Play Right"
```

-   **`location`**: Update `latitude` and `longitude` to your coordinates. You can use an online tool like google maps to find them.
-   **`lights`**: Add the `id` and `name` for each light you want to control.

**Important**: If you keep a config in the repository as `configs/config.yaml`, it is ignored by git to protect your sensitive data (coordinates and light IDs). Never commit this file to version control.

## Usage

### First-Time Use: Registering with the Hue Bridge

On the very first run, `hue-lighter` needs to create a user (API key) on your Philips Hue Bridge.

1.  **Start the app for the first time.** It logs a message prompting you to **press the link button** on your Hue Bridge. The log is in `~/Library/Logs/hue-lighter.log` on macOS; on Linux, start the app from a terminal to see it:
    ```sh
    flatpak run io.github.yveskaufmann.HueLighter
    ```

2.  **Press the button** on your bridge. The application will automatically detect it, create a user, and store the API key in your keyring (macOS: Keychain) for future use. It will then start its normal operation.

3.  **Allow running in the background** when asked (Flatpak). `hue-lighter` registers itself via the Background portal to start automatically on login (`~/.config/autostart/io.github.yveskaufmann.HueLighter.desktop`). On macOS, the app writes a LaunchAgent (`~/Library/LaunchAgents/com.github.yveskaufmann.hue-lighter.plist`) instead.

### Managing the App (Linux Flatpak)

-   **Start:** `flatpak run io.github.yveskaufmann.HueLighter` (or `make flatpak-run`), or start "Hue Lighter" from your app launcher.
-   **Stop:** `kill -TERM $(pgrep -x hue-lighter)` — this turns the configured lights off. Note that `flatpak kill` terminates the app immediately without turning the lights off.
-   **Disable autostart:** turn off "Hue Lighter" in your desktop's background apps / autostart settings.
-   **Uninstall:** `flatpak uninstall io.github.yveskaufmann.HueLighter`
-   **Headless (no tray icon):** pass `--no-tray`.

### Logout and Shutdown

When `hue-lighter` receives `SIGTERM`, `SIGINT` or `SIGHUP`, it turns off all configured lights before exiting. When you log out or shut down, your session sends `SIGTERM` to the app, so the lights are turned off automatically.

## Development

-   **Build the binary:**
    ```sh
    make build
    ```
-   **Run tests:**
    ```sh
    make test
    ```
-   **Format code:**
    ```sh
    make fmt
    ```

The project includes a `.vscode/launch.json` file for easy debugging in Visual Studio Code.

## Architecture

See the architecture diagrams and sequence flows in [docs/architecture.md](docs/architecture.md).

## Supported Platforms

- **macOS 11 or newer:** `Hue Lighter.app` runs in the menu bar and starts on login via a LaunchAgent. It is installed via Homebrew or `make mac_install`. Config and the fallback API key file live in `~/Library/Application Support/hue-lighter/`, logs in `~/Library/Logs/hue-lighter.log`.
- **Linux (Flatpak):** The app runs in the desktop session and is installed via `make flatpak`. Config and the fallback API key file live in `~/.var/app/io.github.yveskaufmann.HueLighter/config/hue-lighter/`, outside of Flatpak in `~/.config/hue-lighter/`.
- **Docker:** The application can be containerized. Run it with `--no-tray`, mount your `configs/config.yaml` (`CONFIG_PATH`) and provide the CA bundle via `HUE_CA_CERTS_PATH`. The lights are turned off when the container receives `SIGTERM` (e.g. `docker stop`).

Example containerization files are provided under `examples/containerized/` and a `docker-compose.yml` in `examples/`. See [docs/docker.md](docs/docker.md) for build and run instructions.

## License

This project is released under the MIT License — see the [LICENSE](LICENSE) file for details.

