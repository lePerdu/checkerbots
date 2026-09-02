package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	gameengine "checkerbots/apps/server/game-engine"
)

//go:embed static/* templates/*
var assets embed.FS

type pageData struct{}

// AppSnapshot is the full serializable state sent to clients on initial SSE connect
// and available via GET /api/state.
type AppSnapshot struct {
	Version int64           `json:"version"`
	Game    GameSnapshot    `json:"game"`
	Robots  []RobotSnapshot `json:"robots"`
	// CellSizeMM is the physical size of one board square in millimetres.
	CellSizeMM float64 `json:"cellSizeMM"`
	// RobotDiameterMM is the physical robot diameter in millimetres.
	RobotDiameterMM float64   `json:"robotDiameterMM"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// RobotSnapshot is the serializable form of a single robot's state.
type RobotSnapshot struct {
	ID        RobotID            `json:"id"`
	PieceID   gameengine.PieceID `json:"piece_id"`
	Pose      Pose               `json:"pose"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// GameSnapshot carries the board and game state delivered to the frontend.
// It is sent as a game.updated SSE event whenever the game changes.
//
// The frontend renders a single grid of BoardSize rows by
// BoardSize+2*CaptureColumns columns: columns [0, BoardSize) are the actual
// checkers board, columns [-CaptureColumns, 0) hold black's captured pieces,
// and columns [BoardSize, BoardSize+CaptureColumns) hold red's. Pieces
// (on-board or captured) all share the same Row/Col position fields, with
// their extended-grid position already resolved server-side.
type GameSnapshot struct {
	Turn              string                                    `json:"turn"`
	GameOver          *gameOverResponse                         `json:"gameOver"`
	BoardSize         int                                       `json:"boardSize"`
	CaptureColumns    int                                       `json:"captureColumns"`
	Pieces            []gamePiece                               `json:"pieces"`
	LegalMovesByPiece map[gameengine.PieceID][]legalMoveSummary `json:"legalMovesByPiece"`
}

type gameOverResponse struct {
	Winner string `json:"winner"`
	Reason string `json:"reason"`
}

// legalMoveSummary describes one full legal move option as the ordered list of
// squares visited. A simple step or single jump has 2 positions; a multi-jump
// sequence has one entry per square landed on.
type legalMoveSummary struct {
	Path []squarePosition `json:"path"`
}

type squarePosition struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type applyMoveRequest struct {
	Path []squarePosition `json:"path"`
}

// gamePiece is the serializable form of a single piece, whether it's
// currently on the board or set aside as captured - both share the same
// Row/Col position fields, using the extended grid described on GameSnapshot.
type gamePiece struct {
	ID      gameengine.PieceID `json:"id"`
	Side    string             `json:"side"`
	Kind    string             `json:"kind"`
	Classes string             `json:"classes"`
	Row     int                `json:"row"`
	Col     int                `json:"col"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Error: apiError{Code: code, Message: message},
	})
}

func main() {
	tmpl := template.Must(template.ParseFS(assets, "templates/index.html"))

	staticFS, err := fs.Sub(assets, "static")
	if err != nil {
		log.Fatalf("load static assets: %v", err)
	}

	var statePath string
	if tmpDir, set := os.LookupEnv("TMP_DIR"); set {
		statePath = path.Join(tmpDir, "checkerbots.gob")
	}

	initial, err := loadState(statePath)
	if err != nil {
		log.Printf("failed to load state: %v; using default state", err)
		initial = makeInitialStoredState()
	}

	h := newSseHub()
	go h.run()

	mgr := newStateManager()
	go mgr.run(initial, h.broadcast)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, pageData{}); err != nil {
			log.Printf("render index: %v", err)
		}
	})

	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(mgr.getSnapshot()); err != nil {
			log.Printf("encode state response: %v", err)
		}
	})

	mux.HandleFunc("GET /api/board", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(mgr.getSnapshot().Game); err != nil {
			log.Printf("encode board response: %v", err)
		}
	})

	mux.HandleFunc("POST /api/moves", func(w http.ResponseWriter, r *http.Request) {
		var req applyMoveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "bad_request", "invalid request body")
			return
		}
		if len(req.Path) < 2 {
			writeJSONError(w, http.StatusBadRequest, "bad_request", "move path must contain at least 2 positions")
			return
		}
		move := make(gameengine.Move, len(req.Path))
		for i, p := range req.Path {
			move[i] = gameengine.Position{Row: p.Row, Col: p.Col}
		}
		if err := mgr.applyMove(move); err != nil {
			writeJSONError(w, http.StatusConflict, "illegal_move", err.Reason)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("POST /api/new-game", func(w http.ResponseWriter, r *http.Request) {
		mgr.newGame()
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("/api/events: client connected")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Subscribe before getting the snapshot to avoid missing events between
		// the two steps. Any event received before the snapshot is a harmless
		// duplicate since game state is idempotent to re-apply.
		ch := h.subscribe()
		defer h.unsubscribe(ch)

		snap := mgr.getSnapshot()
		if msg, err := encodeSSEEvent(sseEvent{name: "state.snapshot", data: snap}); err != nil {
			log.Printf("/api/events: encode initial state.snapshot: %v", err)
		} else {
			w.Write(msg)
			flusher.Flush()
		}

		keepAliveInterval := 15 * time.Second
		keepAliveTimer := time.NewTicker(keepAliveInterval)
		defer keepAliveTimer.Stop()

		for {
			select {
			case <-keepAliveTimer.C:
				if _, err := w.Write([]byte(":\n")); err != nil {
					log.Printf("/api/events: keep-alive write error: %v", err)
				}
				flusher.Flush()
			case msg, ok := <-ch:
				if !ok {
					// Hub dropped this client due to a full buffer.
					return
				}
				if _, err := w.Write(msg); err != nil {
					log.Printf("/api/events: write error: %v", err)
				}
				flusher.Flush()
				keepAliveTimer.Reset(keepAliveInterval)
			case <-r.Context().Done():
				log.Printf("/api/events: client disconnected")
				return
			}
		}
	})

	addr := ":8080"
	if value := os.Getenv("PORT"); value != "" {
		addr = ":" + value
	}

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-interruptChan
		log.Printf("signal received (%s); saving state before exit", sig)
		if err := mgr.save(statePath); err != nil {
			log.Printf("save state: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server shutdown: %v", err)
		}
	}()

	log.Printf("server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func buildAppSnapshot(state *appState) AppSnapshot {
	robots := make([]RobotSnapshot, 0, len(state.Robots))
	for id := range state.Robots {
		robots = append(robots, buildRobotSnapshot(state, id))
	}
	return AppSnapshot{
		Version:         state.Version,
		Game:            buildGameSnapshot(&state.Game),
		Robots:          robots,
		CellSizeMM:      boardCellSizeMM,
		RobotDiameterMM: 340,
		UpdatedAt:       state.UpdatedAt,
	}
}

func buildRobotSnapshot(state *appState, id RobotID) RobotSnapshot {
	assignedPiece, _ := state.Assignments.GetPieceIDByRobotID(id)
	return RobotSnapshot{
		ID:        id,
		PieceID:   assignedPiece,
		Pose:      state.Robots[id].Pose,
		UpdatedAt: state.Robots[id].UpdatedAt,
	}
}

func buildGameSnapshot(game *gameengine.Game) GameSnapshot {
	// activeByPosition only holds non-captured pieces: captured pieces have no
	// legal moves, and gameengine already gives them an extended-grid Position
	// (in the capture columns) that's disjoint from any real board position.
	activeByPosition := make(map[gameengine.Position]gameengine.Piece, len(game.Pieces))
	pieces := make([]gamePiece, 0, len(game.Pieces))
	for _, piece := range game.Pieces {
		if !piece.Captured {
			activeByPosition[piece.Position] = piece
		}
		pieces = append(pieces, gamePiece{
			ID:      piece.ID,
			Side:    string(piece.Side),
			Kind:    string(piece.Kind),
			Classes: pieceClasses(piece),
			Row:     piece.Position.Row,
			Col:     piece.Position.Col,
		})
	}

	legalMovesByPiece := make(map[gameengine.PieceID][]legalMoveSummary)
	for _, move := range game.LegalMoves {
		if len(move) < 2 {
			continue
		}
		piece, ok := activeByPosition[move[0]]
		if !ok {
			continue
		}
		path := make([]squarePosition, len(move))
		for i, position := range move {
			path[i] = squarePosition{Row: position.Row, Col: position.Col}
		}
		legalMovesByPiece[piece.ID] = append(legalMovesByPiece[piece.ID], legalMoveSummary{Path: path})
	}

	var gameOver *gameOverResponse
	if game.GameOver != nil {
		gameOver = &gameOverResponse{
			Winner: titleCaseTurn(game.GameOver.Winner),
			Reason: game.GameOver.Reason,
		}
	}

	return GameSnapshot{
		Turn:              titleCaseTurn(game.Turn),
		BoardSize:         game.BoardSize,
		CaptureColumns:    game.CaptureColumns,
		Pieces:            pieces,
		LegalMovesByPiece: legalMovesByPiece,
		GameOver:          gameOver,
	}
}

func pieceClasses(piece gameengine.Piece) string {
	classes := "piece piece--" + string(piece.Side)
	if piece.Kind == gameengine.PieceKindKing {
		classes += " piece--king"
	}
	return classes
}

func titleCaseTurn(side gameengine.PlayerSide) string {
	if side == gameengine.PlayerSideBlack {
		return "Black"
	}
	return "Red"
}
