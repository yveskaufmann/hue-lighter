# Docker / Containerization

This document shows a simple example to containerize `hue-lighter` for development or testing.

Files provided:

- `Dockerfile` — multi-stage builder image that produces a small runtime image (Debian-slim with `ca-certificates`).
- `examples/docker-compose.yml` — example Compose file that mounts the config and CA bundle and uses host networking for discovery.
 - `examples/containerized/Dockerfile` — multi-stage builder image that produces a small runtime image (Debian-slim with `ca-certificates`).
 - `examples/docker-compose.yml` — example Compose file that mounts the config and CA bundle and uses host networking for discovery.

## Build the image

docker build -t hue-lighter:local .
docker compose build
From the repository root you can build the example image directly or use the provided Compose example.

Build locally using the example Dockerfile (from repo root):

```sh
# build using the example Dockerfile folder
docker build -t hue-lighter:local -f examples/containerized/Dockerfile .
```

Or using the provided docker-compose example (from the `examples/` folder):

```sh
cd examples
docker compose build
```

## Run (single container)

```sh
# run with host networking (required for mDNS / bridge discovery)
docker run --rm \
  --network host \
  -v $(pwd)/configs/config.yaml:/etc/hue-lighter/config.yaml:ro \
  -v $(pwd)/configs/certs/cacert_bundle.pem:/etc/hue-lighter/cacert_bundle.pem:ro \
  -v hue-lighter-data:/var/lib/hue-lighter \
  -e CONFIG_PATH=/etc/hue-lighter/config.yaml \
  -e HUE_CA_CERTS_PATH=/etc/hue-lighter/cacert_bundle.pem \
  -e HUE_API_KEY_STORE_PATH=/var/lib/hue-lighter/api-keys.json \
  hue-lighter:local
```

!**Note:** On the first run, you need to press the link button on your Hue Bridge to register the application. See the "First-Time Use" section in the main README for details.

## Run with docker-compose (recommended for testing)

From `examples/`:

```sh
(cd examples && docker compose up --build)
```

## Notes and caveats

- Discovery: the Hue bridge discovery uses mDNS/DNSSD. To allow discovery from a container, `--network host` is the simplest approach. On macOS/Windows, host networking behaves differently; using host networking is best on Linux hosts.
- Shutdown: the app turns off the configured lights on `SIGTERM`, `SIGINT` or `SIGHUP`, so `docker stop` turns them off.
- Tray icon: containers have no desktop session, so the image runs `hue-lighter --no-tray`.
- CA bundle: the Philips Hue CA bundle must be provided by the user and mounted into the container at the expected path (or set `HUE_CA_CERTS_PATH` to another path inside the container). It is excluded via `.dockerignore`, so it is never embedded into the image.
- API keys: without a desktop keyring, API keys are kept in the file set via `HUE_API_KEY_STORE_PATH` (mounted volume).
- Security: do not publish images that embed your private `configs/config.yaml` or API keys.

If you'd like, I can:
- Add a small `Dockerfile` optimization (distroless image) — but this may require additional testing for TLS/certificates.
- Create a GitHub Actions workflow to build and publish images on tags.

*** End of file
