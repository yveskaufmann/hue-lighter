# Architecture

This document contains a compact container diagram for the running system and a sequence diagram describing bridge discovery.

## Container Diagram

```mermaid
flowchart TB
  subgraph LocalNetwork
    BRIDGE[Philips Hue Bridge]
  end

  subgraph Host [Linux Host]
    HUEL[Hue Lighter Service]
    CONFIG[Configuration]
    CA[CA bundle]
  end

  HUEL -->|reads| CONFIG
  HUEL -->|uses TLS bundle| CA
  HUEL -->|discovers & controls| BRIDGE

  classDef external fill:#f3f4f6,stroke:#bbb;
  class BRIDGE external
```

Notes:
- `Hue Lighter Service` runs on the host (systemd) and is responsible for discovery, registration, and automation.
- `Configuration` defines which lights to manage and where the bridge is located (IP/hostname or discovery fallback).
- The service uses the CA bundle to validate TLS when talking to the bridge.

## Bridge discovery — Sequence Diagram

```mermaid
sequenceDiagram
  participant System as systemd
  participant App as Hue Lighter
  participant Config as Config
  participant MDNS as mDNS/DNSSD
  participant Bridge as Philips Hue Bridge
  participant User as User (link button)

  System->>App: start service
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
    App->>Config: store API key (local storage)
  end

  App->>Bridge: authenticated API calls (toggle lights, query state)
  Bridge-->>App: respond

```

Notes:
- Discovery uses mDNS/DNSSD to find local Philips Hue Bridges when no address is configured.
- Registration requires the physical link button on the bridge to be pressed; the app will prompt and retry as needed.
- Once registered, the app stores the API key and proceeds to control configured lights.

