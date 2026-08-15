# Server App

This app hosts the coordination server.

Expected responsibilities:
- manage game state
- track robot state and latest known poses
- validate moves
- dispatch robot commands
- expose HTTP and realtime APIs
- serve the web UI

## Run locally

From the repo root:

```sh
go run ./apps/server
```

Then open http://localhost:8080.

To use a different port:

```sh
PORT=3000 go run ./apps/server
```

## Live reload

This app includes an `air` config at `apps/server/.air.toml`.

Install `air` once:

```sh
go install github.com/air-verse/air@latest
```

Then start the server with live reload from the repo root:

```sh
make server-dev
```

`air` will rebuild and restart the server when files under `apps/server` change, including Go, HTML, CSS, and JavaScript files.

## Current placeholder UI

The current server-rendered placeholder UI includes:
- an 8x8 checkers board
- pieces in starting positions
- a "New game" button
- a turn indicator

Current routes:
- `GET /` renders the page
- `POST /games/new` returns placeholder new-game data for the UI
- `GET /static/...` serves CSS and JavaScript assets
