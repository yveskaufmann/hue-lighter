# Architecture

This document contains a compact container diagram for the running system and a sequence diagram describing bridge discovery.

## Container Diagram

```mermaid
flowchart TB
  subgraph LocalNetwork
    BRIDGE[Philips Hue Bridge]
  end

  subgraph Host [Desktop Session (Flatpak sandbox)]
    HUEL[Hue Lighter]
    TRAY[Tray Icon]
    CONFIG[Configuration]
    KEYRING[Keyring / Secret Service]
  end

  HUEL -->|reads| CONFIG
  HUEL -->|stores API key| KEYRING
  HUEL <-->|commands / light state events| TRAY
  HUEL -->|discovers & controls| BRIDGE

  classDef external fill:#f3f4f6,stroke:#bbb;
  class BRIDGE external
```

Notes:
- `Hue Lighter` runs in the user's desktop session (on Linux as Flatpak, started on login via the Background portal) and is responsible for discovery, registration, and automation. It stops and turns off the lights on `SIGTERM`, `SIGINT` or `SIGHUP` (logout, shutdown).
- The `Tray Icon` runs in its own goroutine; it sends a toggle command to the light automation and receives light state changed events, both via Go channels.
- `Configuration` defines which lights to manage; it lives in the XDG config directory (`~/.config/hue-lighter`, inside Flatpak `~/.var/app/io.github.yveskaufmann.HueLighter/config/hue-lighter`).
- The Philips Hue CA bundle, used to validate TLS when talking to the bridge, is embedded into the binary at build time (`HUE_CA_CERTS_PATH` overrides it).

## Bridge discovery — Sequence Diagram

```mermaid
sequenceDiagram
  participant System as Desktop session
  participant App as Hue Lighter
  participant Config as Config
  participant MDNS as mDNS/DNSSD
  participant Bridge as Philips Hue Bridge
  participant User as User (link button)

  System->>App: start on login (autostart)
  App->>Config: load config
  alt bridge address present
    App->>Bridge: connect to configured address
  else no configured bridge
    App->>MDNS: query for Hue bridges
    MDNS-->>App: return bridge address(es)
    App->>Bridge: connect to discovered bridge
  end

  alt no API key
    App->>Bridge: attempt registration (create user)
    Note right of User: Press link button on bridge
    User->>Bridge: press link button
    Bridge-->>App: return API key
    App->>Config: store API key (keyring)
  end

  App->>Bridge: authenticated API calls (toggle lights, query state)
  Bridge-->>App: respond

```

Notes:
- Discovery uses mDNS/DNSSD to find local Philips Hue Bridges when no address is configured.
- Registration requires the physical link button on the bridge to be pressed; the app will prompt and retry as needed.
- Once registered, the app stores the API key and proceeds to control configured lights.

