# Hue Lighter

Hue Lighter is a small automation tool for managing Philips Hue lights, specifically designed to automate the backlighting of a desktop PC. It intelligently turns lights on at sunset, off at sunrise, and ensures they are turned off gracefully when you log out or the machine shuts down.

## Motivation

While Philips Hue already offers sunset and sunrise automation, I found it frustrating that my backlights would turn on even when I wasn't at my desk. I also wanted the lights to turn off automatically when I shut down my PC. This project was a fun side quest to solve those problems and experiment with automation beyond the built-in Hue features.

## Features

-   **Sunset/Sunrise Automation**: Turns configured lights on at sunset and off at sunrise based on your location.
-   **Tray Icon**: Switch the lights on or off manually. A manual switch is kept until the lights are back in the state the automation would choose.
-   **Graceful Shutdown**: Turns off all configured lights when you log out or shut down.
-   **Desktop App**: Runs in your desktop session and starts on login (Flatpak on Linux, menu bar app on macOS).

## Installation

### macOS (Homebrew)

```sh
brew tap yveskaufmann/tap
brew trust yveskaufmann/tap
brew install --cask hue-lighter
```

> [!NOTE]
> The app isn't signed with an Apple Developer ID. The cask removes the quarantine flag, so macOS doesn't block it.

### Linux

Download the binary for your architecture from the [latest release](https://github.com/yveskaufmann/hue-lighter/releases/latest).

### Docker

See [docs/docker.md](docs/docker.md).

### From source

Requires Go 1.25+, `make` and the Philips Hue CA bundle. The Flatpak additionally needs `flatpak-builder` with `org.freedesktop.Sdk//25.08` and `org.freedesktop.Sdk.Extension.golang//25.08` from Flathub.

1. Clone the repository:
   ```sh
   git clone https://github.com/yveskaufmann/hue-lighter.git
   cd hue-lighter
   ```
2. Download the CA bundle from the [Philips Hue Developer Portal](https://developers.meethue.com/develop/application-design-guidance/using-https/) and save it as `configs/certs/cacert_bundle.pem`. It gets embedded into the binary.
3. Build and install:

   | Target          | Command            |
   | --------------- | ------------------ |
   | Linux (Flatpak) | `make flatpak`     |
   | macOS (app)     | `make mac_install` |
   | Binary only     | `make build`       |

> [!IMPORTANT]
> The [Philips Hue Terms of Use](https://developers.meethue.com/terms-of-use-and-conditions/) don't allow redistributing the CA bundle, so it isn't part of this repository.

## Usage

### First start

1. Start **Hue Lighter**.
2. Press the link button on your Hue Bridge to pair the app. The API key is stored in your keyring (macOS: Keychain).
3. Set your location and lights in the settings page.

> [!NOTE]
> The app requests to run in the background, so it can start on every login.

### Configuration

| Installation    | Config file                                                                   |
| --------------- | ----------------------------------------------------------------------------- |
| macOS           | `~/Library/Application Support/hue-lighter/config.yaml`                       |
| Linux (Flatpak) | `~/.var/app/io.github.yveskaufmann.HueLighter/config/hue-lighter/config.yaml` |
| Linux (binary)  | `~/.config/hue-lighter/config.yaml`                                           |

See [configs/config.example.yaml](configs/config.example.yaml) for the format. Light IDs are shown in the Philips Hue app under Settings → Lights → (light) → Info.

| Variable            | Purpose                                             |
| ------------------- | --------------------------------------------------- |
| `CONFIG_PATH`       | Use a different config file.                        |
| `HUE_CA_CERTS_PATH` | Use a CA bundle file instead of the embedded one.   |
| `LOG_LEVEL`         | `trace`, `debug`, `info` (default), `warn`, `error` |
| `LOG_FORMAT`        | `text` (default) or `json`                          |

Pass `--no-tray` to run without tray icon, e.g. headless or in Docker.

### Logs

-   **macOS:** `~/Library/Logs/hue-lighter.log`
-   **Linux:** stderr, e.g. `flatpak run io.github.yveskaufmann.HueLighter` in a terminal

### Stopping

The lights turn off when the app quits, including on logout and shutdown (`SIGTERM`, `SIGINT`, `SIGHUP`).

> [!WARNING]
> `flatpak kill` terminates the app immediately, without turning the lights off. Use `kill -TERM $(pgrep -x hue-lighter)` instead.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This project is released under the MIT License — see the [LICENSE](LICENSE) file for details.
