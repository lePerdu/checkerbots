# Networking Plan

CheckerBots should use a local Wi-Fi network with a dedicated router or access point for robot communication.

## Decision summary

- **Primary network:** 2.4 GHz Wi-Fi
- **Topology:** centralized star topology
- **Infrastructure:** dedicated portable router/AP near the board
- **Connection direction:** robots initiate outbound connections to the server
- **Robot transport:** WebSocket connection to the coordination server
- **Discovery:** mDNS advertisement for the server address
- **BLE role:** optional provisioning/debug only, not primary control
- **Mesh role:** avoid for MVP unless Wi-Fi range/reliability proves insufficient

## Recommended topology

```text
Web UI / operator
      |
      v
Coordination server
      |
      v
Dedicated Wi-Fi router/AP
      |
      +-- robot-01
      +-- robot-02
      +-- ...
      +-- robot-24
```

The coordination server remains the source of truth for game state and robot state. Robots do not coordinate directly with each other.

This matches the project model where the server validates moves, dispatches commands, tracks robot state, and handles recovery workflows.

## Why Wi-Fi with a dedicated router/AP

Wi-Fi is the best initial fit for 24 robots because it is widely supported, easy to debug, and maps cleanly to the existing Go server architecture.

A dedicated router/AP should be used instead of venue Wi-Fi or a laptop hotspot.

Benefits:

- supports 24 low-bandwidth robot clients comfortably
- keeps the system portable between locations
- avoids dependency on venue network policies or internet access
- gives a stable SSID/password for robot configuration
- provides normal IP networking for logs, diagnostics, WebSockets, and future OTA updates
- is easier to observe and troubleshoot than Bluetooth or mesh protocols

Preferred physical setup:

- router/AP placed near the board
- server connected to router by Ethernet when practical
- robots connected over 2.4 GHz Wi-Fi
- no internet dependency required for normal operation

## Server discovery with mDNS

Robots should discover the coordination server using mDNS instead of hardcoding a server IP address.

Recommended service name:

```text
_checkerbots._tcp.local
```

Example advertised instance:

```text
checkerbots-server._checkerbots._tcp.local
```

The server should advertise:

- service type: `_checkerbots._tcp`
- hostname, for example `checkerbots.local`
- port for robot WebSocket connections
- optional TXT metadata such as protocol version

Example TXT records:

```text
proto=v0
transport=websocket
path=/robots/ws
```

Robots should resolve the mDNS service during startup and reconnect attempts. If mDNS resolution fails, firmware may optionally fall back to a configured static hostname or IP address.

mDNS keeps the setup portable: the router can assign a different IP address at a new location without requiring every robot to be reflashed or reconfigured.

## Robot connection model

Robots should initiate the connection to the server.

Recommended flow:

1. Robot boots.
2. Robot connects to the dedicated Wi-Fi SSID.
3. Robot discovers the server using mDNS.
4. Robot opens a WebSocket connection to the server.
5. Robot sends `hello` with identity and capabilities.
6. Robot sends periodic `heartbeat`, telemetry, and command results.
7. Server sends commands over the same WebSocket.

This avoids inbound connections to robots and works well with DHCP.

## Server-hosted WebSocket endpoint

The Go coordination server should host a robot WebSocket endpoint, for example:

```text
/robots/ws
```

Expected robot-to-server messages include:

- `hello`
- `heartbeat`
- `telemetry`
- `pose_update`
- `command_result`
- `error`

Expected server-to-robot commands include:

- `identify`
- `stop`
- `move_to_square`
- `move_to_pose`
- `set_pose`
- `ping`

Each command should include a command ID so command results can be correlated with the original request.

The WebSocket layer should be treated as a transport. Game rules, robot state, command lifecycle, and recovery behavior should remain server-side concerns.

## Potential microcontrollers

### ESP32

Best default choice for physical prototypes.

Pros:

- inexpensive
- strong Wi-Fi support
- BLE available for optional provisioning/debug
- more RAM and CPU headroom than ESP8266
- widely supported by Arduino and ESP-IDF ecosystems
- enough capability for WebSocket client logic, reconnect handling, and serial control

Cons:

- higher power use than BLE-only controllers
- firmware still needs careful reconnect and watchdog behavior

### ESP8266 / ESP-01

Good low-cost option if the firmware remains simple.

Pros:

- very cheap
- small
- Wi-Fi support
- already noted as a possible Roomba bridge option

Cons:

- less RAM and CPU headroom
- no BLE provisioning option
- tighter limits for robust TLS/WebSocket/diagnostic features
- ESP-01 has limited GPIO and can be less convenient during development

### Arduino Nano 33 IoT

Viable but likely more expensive than necessary.

Pros:

- I already have one :)
- Wi-Fi capable
- existing boards are available for experimentation
- familiar Arduino workflow
- extra sensors may help prototypes

Cons:

- higher cost per robot
- may not provide enough benefit for a 24-robot fleet compared with ESP32

### Raspberry Pi Zero W or similar

Useful for advanced experiments but not recommended as the default robot controller.

Pros:

- full Linux environment
- easy debugging and rich language support
- can run higher-level software locally

Cons:

- higher cost and power use
- slower boot
- SD card and OS maintenance burden
- overkill for serial bridge/control duties

## Reconnect behavior

Reconnect behavior is part of the core design, not an afterthought.

Robots should handle:

- router reboot
- server restart
- temporary Wi-Fi drop
- mDNS lookup failure
- duplicate robot ID connection
- command in progress during disconnect
- stale command after reconnect

Recommended behavior:

- use exponential backoff with jitter for reconnect attempts
- send a fresh `hello` after every reconnect
- include firmware version, robot ID, current state, and last command ID in `hello`
- treat commands as invalid after a disconnect unless the server explicitly resumes or reconciles them
- server should mark robots offline after missed heartbeats
- server should reject or replace duplicate connections for the same robot ID deterministically

## Safety behavior

Every robot should have local safety behavior independent of the server.

Minimum safety rules:

- stop motors if the WebSocket disconnects while moving
- stop motors if no valid command/heartbeat window is observed for a configured timeout
- stop motors on malformed or unsupported movement commands
- ignore stale commands with old command IDs or incompatible protocol versions
- expose a manual physical power/stop procedure for demos

The server should provide a stop-all recovery action, but robot firmware should not rely on the server for basic failsafe behavior.

Suggested defaults for early prototypes:

- heartbeat interval: 1-2 seconds
- server offline threshold: 5-10 seconds
- robot command watchdog while moving: 1-3 seconds without expected control traffic

Exact values should be adjusted after physical testing.

## Security concerns

The first physical versions can run on a private local network, but security should still be explicit.

Recommended MVP security posture:

- dedicated WPA2/WPA3-protected SSID
- no dependence on internet access
- do not expose the robot WebSocket endpoint on public networks
- avoid hardcoded secrets in source code
- use per-robot IDs that are not treated as secrets
- optionally add per-robot tokens before public demos or multi-user environments

Risks to account for:

- anyone on the Wi-Fi network may be able to send robot commands if there is no application-level authentication
- plaintext WebSockets expose commands to local network observers
- mDNS advertises the server presence to devices on the same LAN
- lost/stolen robots may contain Wi-Fi credentials in firmware or flash storage

Possible later improvements:

- per-robot enrollment tokens
- signed or authenticated robot messages
- TLS WebSocket connections if feasible on the selected microcontroller
- separate operator and robot VLANs or SSIDs
- allowlist robot IDs and reject unknown devices
- firmware update signing

For MVP development, network isolation plus local-only operation is acceptable. Before public demos, add at least per-robot authentication tokens and an operator checklist for router/server configuration.

## Implementation phases

### Phase 1: single-robot Wi-Fi prototype

Goals:

- validate Wi-Fi connection
- validate mDNS discovery
- validate WebSocket connection to the Go server
- validate basic command/result flow
- validate local stop-on-disconnect behavior

Deliverables:

- one microcontroller connects to the dedicated SSID
- robot discovers `_checkerbots._tcp.local`
- robot sends `hello` and `heartbeat`
- server can send `identify`, `ping`, and `stop`
- robot stops safely when server or Wi-Fi disappears

### Phase 2: movement-capable robot

Goals:

- connect the microcontroller to the Roomba control interface
- execute simple motion commands from the server
- report command results and errors

Deliverables:

- `move_to_pose` or equivalent low-level movement test
- command IDs included in requests and results
- movement watchdog implemented in firmware
- server marks command success, failure, timeout, or interrupted

### Phase 3: small fleet test

Goals:

- prove that the architecture works with multiple robots
- exercise reconnect and duplicate identity behavior
- refine telemetry and heartbeat rates

Deliverables:

- 3-4 robots connected simultaneously
- all robots visible in the server registry
- one moving robot at a time
- router reboot and server restart tests documented in code comments or test notes
- stop-all behavior verified

### Phase 4: full 24-robot connectivity test

Goals:

- validate router/AP capacity
- validate idle telemetry load
- validate full-fleet reconnect behavior

Deliverables:

- 24 robot controllers connected to the dedicated AP
- stable heartbeat reporting
- server can identify each robot
- server handles sequential commands across different robots
- reconnect storm behavior tested after router/server restart

### Phase 5: demo hardening

Goals:

- make the system portable and safe for repeated physical demonstrations

Deliverables:

- fixed router/AP configuration
- documented SSID and local hostname assumptions in code comments/config
- per-robot labels matching robot IDs
- optional per-robot auth tokens
- operator emergency-stop workflow
- startup checklist for router, server, robots, and UI
