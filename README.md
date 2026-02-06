# Hue Lighter

Hue Lighter is a small automation tool for managing Philips Hue lights, specifically designed to automate the backlighting of a desktop PC. It intelligently turns lights on at sunset, off at sunrise, and ensures they are turned off gracefully when the machine shuts down.

## Features

-   **Sunset/Sunrise Automation**: Automatically turns configured lights on at sunset and off at sunrise based on your geographic location.
-   **Graceful Shutdown**: Turns off all configured lights when the machine is shut down, ensuring you don't leave them on by accident.
-   **System Service**: Runs as a background service on Linux systems using `systemd`.
-   **Automatic Registration**: On first run, it guides you through the simple process of registering the app with your Philips Hue Bridge by pressing the link button.

## Prerequisites

-   Go (version 1.24 or newer)
-   A Philips Hue Bridge connected to the same local network.
-   A Linux-based operating system with `systemd` and `systemctl`.
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

`hue-lighter` requires the Philips Hue CA certificate bundle to validate TLS connections to the Hue Bridge. This file is not stored in version control — you must obtain it yourself and place it on the host where the service runs.

**Legal Note**: The CA bundle cannot be redistributed with this project as it would violate the [Philips Hue Terms of Use and Conditions](https://developers.meethue.com/terms-of-use-and-conditions/), which restrict the redistribution of Philips Hue materials without explicit permission. Each user must download the bundle directly from Philips.

- Default path: `/etc/hue-lighter/cacert_bundle.pem`
- Override path: set the environment variable `HUE_CA_CERTS_PATH` to point to the bundle file.

Where to get the bundle:

Follow these steps to install the CA bundle:

1. Download the CA bundle from the [Philips Hue Developer Portal](https://developers.meethue.com/develop/application-design-guidance/using-https/).
2. Create the directory if it doesn't exist:
   ```sh
   mkdir -p configs/certs
   ```
3. Save the downloaded bundle to `configs/certs/cacert_bundle.pem`.

If the CA bundle is missing when the service starts, the application will terminate with an explanatory error indicating the missing bundle and how to install it. See the `HUE_CA_CERTS_PATH` environment variable if you need a non-default location.

### 3. Install the Service

The `Makefile` provides a simple way to install the application and service.

```sh
make install
```

This command will:
1.  Build the `hue-lighter` binary.
2.  Copy the binary to `/usr/bin/hue-lighter`.
3.  Copy the systemd service file to `/etc/systemd/system/hue-lighter.service`.
4.  Copy your configuration to `/etc/hue-lighter/config.yaml`.
5.  Reload the `systemd` daemon and enable the service to start on boot.

## Usage

### First-Time Use: Registering with the Hue Bridge

On the very first run, `hue-lighter` needs to create a user (API key) on your Philips Hue Bridge.

1.  **Start the service for the first time:**
    ```sh
    sudo systemctl start hue-lighter
    ```

2.  **Check the logs:**
    ```sh
    journalctl -u hue-lighter -f
    ```
    You will see a log message prompting you to **press the link button** on your Hue Bridge.

3.  **Press the button** on your bridge. The application will automatically detect it, create a user, and store the API key for future use. The service will then start its normal operation.

### Managing the Service

-   **Start the service:**
    ```sh
    sudo systemctl start hue-lighter
    ```
-   **Stop the service:**
    ```sh
    sudo systemctl stop hue-lighter
    ```
-   **Check the status:**
    ```sh-
    systemctl status hue-lighter
    ```
-   **View logs:**
    ```sh
    journalctl -u hue-lighter -f
    ```

### Machine Shutdown

The `hue-lighter.service` is configured with an `ExecStop` command that sends a shutdown signal to the application. When your machine shuts down, `systemd` will trigger this command, and the application will turn off all configured lights before exiting.

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

- **Linux with `systemd`:** The project installs a systemd unit and uses systemd lifecycle hooks (start/stop) for graceful shutdown and service management. The default paths are configured for Linux (`/etc/hue-lighter/`, `/usr/bin/hue-lighter`).
- **Docker:** The application can be containerized. If running in a container, mount your `configs/config.yaml` and provide the CA bundle via `HUE_CA_CERTS_PATH` or a bind mount. Be aware that systemd-specific features (ExecStop, unit files) will not behave the same inside containers.

Example containerization files are provided under `examples/containerized/` and a `docker-compose.yml` in `examples/`. See [docs/docker.md](docs/docker.md) for build and run instructions.

## License

This project is released under the MIT License — see the [LICENSE](LICENSE) file for details.

