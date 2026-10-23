
# Hue Lighter

Hue Lighter is a small automation tool for managing Philips Hue lights, specifically designed to automate the backlighting of a desktop PC. It intelligently turns lights on at sunset, off at sunrise, and ensures they are turned off gracefully when you log out or the machine shuts down.

## Motivation

While Philips Hue already offers sunset and sunrise automation, I found it frustrating that my backlights would turn on even when I wasn't at my desk. I also wanted the lights to turn off automatically when I shut down my PC. This project was a fun side quest to solve those problems and experiment with automation beyond the built-in Hue features.

## Features

-   **Sunset/Sunrise Automation**: Automatically turns configured lights on at sunset and off at sunrise based on your geographic location.
-   **Tray Icon**: Switch the lights on or off manually regardless of the time of day. A manual switch is kept until the lights are back in the state the automation would choose.
-   **Graceful Shutdown**: Turns off all configured lights when you log out or shut down the machine, ensuring you don't leave them on by accident.
-   **Desktop App**: Runs in your desktop session, distributed as Flatpak on Linux and started automatically on login.
-   **Automatic Registration**: On first run, it guides you through the simple process of registering the app with your Philips Hue Bridge by pressing the link button.

## Prerequisites

-   Go (version 1.25 or newer)
-   A Philips Hue Bridge connected to the same local network.
-   Linux: `flatpak` and `flatpak-builder` with the `org.freedesktop.Sdk//25.08` and `org.freedesktop.Sdk.Extension.golang//25.08` from Flathub.
-   A desktop with tray icon support (StatusNotifierItem; on GNOME the AppIndicator extension, which Ubuntu ships by default).
-   `make` for easy installation.

## Setup and Installation

### 1. Clone the Repository

```sh
git clone https://github.com/yveskaufmann/hue-lighter.git
cd hue-lighter
```

### 2. Configure Your Lights and Location

Before installing, you need to create your personal configuration file:

```sh
cp configs/config.example.yaml configs/config.yaml
```

Then edit `configs/config.yaml` to match your setup:

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

**Important**: Your personal `configs/config.yaml` file is ignored by git to protect your sensitive data (coordinates and light IDs). Never commit this file to version control.

### CA bundle (required)

`hue-lighter` requires the Philips Hue CA certificate bundle to validate TLS connections to the Hue Bridge. This file is not stored in version control — you must obtain it yourself. `make build` and `make flatpak` embed it into the binary.

**Legal Note**: The CA bundle cannot be redistributed with this project as it would violate the [Philips Hue Terms of Use and Conditions](https://developers.meethue.com/terms-of-use-and-conditions/), which restrict the redistribution of Philips Hue materials without explicit permission. Each user must download the bundle directly from Philips. Because the bundle is embedded, do not share binaries or Flatpaks you built with it.

Follow these steps to provide the CA bundle:

1. Download the CA bundle from the [Philips Hue Developer Portal](https://developers.meethue.com/develop/application-design-guidance/using-https/).
2. Create the directory if it doesn't exist:
   ```sh
   mkdir -p configs/certs
   ```
3. Save the downloaded bundle to `configs/certs/cacert_bundle.pem`.

To use a bundle file instead of the embedded one (e.g. in a container), set the environment variable `HUE_CA_CERTS_PATH` to its path. If neither is available, the application terminates with an explanatory error.

### 3. Build and Install the Flatpak (Linux)

```sh
make flatpak
```

This command will:
1.  Copy the CA bundle so it gets embedded into the binary.
2.  Vendor the Go modules (the Flatpak build runs offline).
3.  Build `io.github.yveskaufmann.HueLighter` with `flatpak-builder` and install it for the current user.

Then copy your configuration into the app's config directory:

```sh
mkdir -p ~/.var/app/io.github.yveskaufmann.HueLighter/config/hue-lighter
cp configs/config.yaml ~/.var/app/io.github.yveskaufmann.HueLighter/config/hue-lighter/config.yaml
```

When running `hue-lighter` outside of Flatpak (e.g. `make build && ./bin/hue-lighter`), the config is read from `~/.config/hue-lighter/config.yaml` instead. Set `CONFIG_PATH` to use a different file.

## Usage

### First-Time Use: Registering with the Hue Bridge

On the very first run, `hue-lighter` needs to create a user (API key) on your Philips Hue Bridge.

1.  **Start the app for the first time** from a terminal to see its logs:
    ```sh
    flatpak run io.github.yveskaufmann.HueLighter
    ```
    You will see a log message prompting you to **press the link button** on your Hue Bridge.

2.  **Press the button** on your bridge. The application will automatically detect it, create a user, and store the API key in your keyring for future use. It will then start its normal operation.

3.  **Allow running in the background** when asked. `hue-lighter` registers itself via the Background portal to start automatically on login (`~/.config/autostart/io.github.yveskaufmann.HueLighter.desktop`).

### Managing the App

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

- **Linux (Flatpak):** The app runs in the desktop session and is installed via `make flatpak`. Config and the fallback API key file live in `~/.var/app/io.github.yveskaufmann.HueLighter/config/hue-lighter/`, outside of Flatpak in `~/.config/hue-lighter/`.
- **Docker:** The application can be containerized. Run it with `--no-tray`, mount your `configs/config.yaml` (`CONFIG_PATH`) and provide the CA bundle via `HUE_CA_CERTS_PATH`. The lights are turned off when the container receives `SIGTERM` (e.g. `docker stop`).

Example containerization files are provided under `examples/containerized/` and a `docker-compose.yml` in `examples/`. See [docs/docker.md](docs/docker.md) for build and run instructions.

## License

This project is released under the MIT License — see the [LICENSE](LICENSE) file for details.

