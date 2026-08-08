# Development Tooling

This document describes the baseline local development tooling for CheckerBots.

## Stack assumptions

The current baseline assumes:
- Go for the coordination server
- Go for the simulated robot service
- server-rendered HTML, CSS, and JavaScript served from the server app

## Included tooling

### Formatting
- `gofmt` for Go source formatting
- `.editorconfig` for basic cross-editor formatting conventions

### Linting
- `golangci-lint` with a small baseline configuration in `.golangci.yml`

### Convenience commands
A root `Makefile` provides common commands:

- `make fmt` - format Go source files under `apps/` and `packages/`
- `make lint` - run `golangci-lint`
- `make test` - run Go tests
- `make check` - run lint and tests

## Dev container

A dev container configuration is provided under `.devcontainer/`.

Current intent:
- provide a consistent Go development environment
- work cleanly in Zed and other editor environments that support dev containers
- keep setup lightweight while the repo is still being scaffolded

## Notes

- The repo does not yet include a root Go module or runnable apps.
- As implementation begins, the tooling may need small adjustments to match the final repo/module layout.
- Additional frontend-specific tooling is intentionally deferred because the web UI will be server-rendered and use minimal JavaScript.
