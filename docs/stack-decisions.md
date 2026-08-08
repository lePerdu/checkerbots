# CheckerBots Stack Decisions

This document records the initial implementation stack choices for the CheckerBots MVP.

## Decision summary

### Coordination server
- **Language:** Go
- **Deployment model:** native compiled binary
- **Primary responsibilities:**
  - manage game state
  - track robot state and latest known poses
  - validate moves
  - dispatch robot commands
  - expose HTTP and WebSocket interfaces
  - serve the web UI assets and templates

### Web UI
- **Rendering model:** server-rendered HTML
- **Styling:** plain CSS
- **Interactivity:** vanilla JavaScript as needed
- **Asset location:** under the server app, not a separate frontend runtime

### Simulated robot service
- **Language:** Go
- **Reasoning:** keeps the simulator aligned with the server stack and preserves the native-binary deployment model

### Camera tracking
- **Decision:** deferred
- **Reasoning:** camera tracking is optional for the first MVP, and computer vision requirements should not force an early stack choice for the rest of the system

## Why Go

Go is the best fit for the current project constraints:

- compiles to a native binary, making deployment simple
- avoids heavy language runtimes like Node.js and Python for the main system
- has strong built-in support for HTTP servers and static asset serving
- has a straightforward concurrency model for handling realtime robot connections and move execution
- supports a fast MVP iteration loop without introducing a large framework

Go is a strong match for the coordination server because the project needs a practical mix of:
- network services
- realtime updates
- state management
- background coordination logic

## Why server-rendered HTML + CSS + vanilla JS

The web UI does not currently need a large client-side framework.

For the MVP, the UI primarily needs to:
- show the current game board
- display robot status and positions
- submit moves
- support setup, calibration, and recovery workflows

These needs can be handled well with:
- HTML templates rendered by the server
- plain CSS for styling
- small amounts of JavaScript for dynamic behavior such as:
  - WebSocket updates
  - board refreshes or animations
  - interactive controls where needed

This approach keeps the system lightweight and avoids the deployment and tooling overhead of a separate SPA frontend.

## Server asset layout decision

The web UI assets should live under the server app rather than in a separate frontend application runtime.

Expected direction:
- `apps/server/`
  - server code
  - HTML templates
  - static CSS
  - static JavaScript

This keeps the MVP deployment model simple:
- one primary server binary
- one place for templates and static assets
- no separate frontend runtime in production

## Tradeoffs and caveats

### Benefits
- simple deployment
- low runtime overhead
- small operational footprint
- fewer moving parts in the MVP
- no requirement for Node.js or Python in the main app path

### Tradeoffs
- richer browser interactivity will require hand-written JavaScript instead of relying on a large UI framework
- if the UI becomes significantly more complex later, the team may want to revisit whether additional client-side structure is helpful
- camera tracking may eventually justify a separate implementation stack depending on computer vision needs

## Deferred decisions

### Camera tracking stack
This decision is intentionally deferred.

If camera tracking becomes necessary, it should be evaluated independently based on:
- computer vision library support
- calibration and pose-estimation needs
- ease of deployment on the target hardware

The rest of the system should treat camera tracking as an optional external service so that a future implementation can use whatever stack is most practical.

## MVP guidance based on this decision

The next implementation steps should assume:
- Go for the coordination server
- Go for the simulated robot service
- server-rendered templates for the UI
- static assets and templates stored under `apps/server/`
- camera tracking remains out of scope for early scaffolding unless interface hooks are easy to leave in place
