package main

import (
	"context"
	"embed"
	"encoding/gob"
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

type boardStateResponse struct {
	Turn              string                        `json:"turn"`
	GameOver          *gameOverResponse             `json:"gameOver"`
	BoardSize         int                           `json:"boardSize"`
	BoardCells        []boardCell                   `json:"boardCells"`
	LegalMovesByPiece map[string][]legalMoveSummary `json:"legalMovesByPiece"`
}

type gameOverResponse struct {
	Winner string `json:"winner"`
	Reason string `json:"reason"`
}

// legalMoveSummary describes one full legal move option, as the ordered list
// of squares visited. A simple step or single jump has 2 positions; a
// multi-jump sequence has one entry per square landed on.
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

type boardCell struct {
	Row         int         `json:"row"`
	Col         int         `json:"col"`
	IsDark      bool        `json:"isDark"`
	Piece       *boardPiece `json:"piece,omitempty"`
	SquareLabel string      `json:"squareLabel"`
}

type boardPiece struct {
	ID      string `json:"id"`
	Side    string `json:"side"`
	Kind    string `json:"kind"`
	Classes string `json:"classes"`
}

type WorldPos struct {
	XMM, YMM float32
}

type boardConfig struct {
	CellSizeMM float32
}

type robot struct {
	ID  string
	Pos WorldPos
}

type serverState struct {
	Game gameengine.Game
}

type storedState struct {
	Game gameengine.StoredGame
}

func saveState(filePath string, state *serverState) error {
	log.Printf("saving state to: %s", filePath)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	storage := storedState{
		Game: gameengine.GameToStored(state.Game),
	}
	if err := gob.NewEncoder(file).Encode(storage); err != nil {
		return err
	}
	return nil
}

func loadState(filePath string) (*serverState, error) {
	log.Printf("loading state from: %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var storage storedState
	if err := gob.NewDecoder(file).Decode(&storage); err != nil {
		return nil, err
	}
	return &serverState{
		Game: gameengine.GameFromStored(storage.Game),
	}, nil
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

	var state *serverState
	if loaded, err := loadState(statePath); err == nil {
		state = loaded
	} else {
		log.Printf("failed to load state: %v; using default state", err)
		state = &serverState{
			Game: gameengine.NewGame8x8(),
		}
	}

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
	mux.HandleFunc("GET /api/board", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(buildBoardStateResponse(&state.Game)); err != nil {
			log.Printf("encode board response: %v", err)
		}
	})
	mux.HandleFunc("POST /api/moves", func(w http.ResponseWriter, r *http.Request) {
		var request applyMoveRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if len(request.Path) < 2 {
			http.Error(w, "move path must contain at least 2 positions", http.StatusBadRequest)
			return
		}
		move := make(gameengine.Move, len(request.Path))
		for i, position := range request.Path {
			move[i] = gameengine.Position{Row: position.Row, Col: position.Col}
		}
		if err := gameengine.ApplyMove(&state.Game, move); err != nil {
			http.Error(w, err.Reason, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(buildBoardStateResponse(&state.Game)); err != nil {
			log.Printf("encode apply move response: %v", err)
		}
	})
	mux.HandleFunc("POST /games/new", func(w http.ResponseWriter, r *http.Request) {
		state.Game = gameengine.NewGame8x8()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(buildBoardStateResponse(&state.Game)); err != nil {
			log.Printf("encode new game response: %v", err)
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
		if err := saveState(statePath, state); err != nil {
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

func buildBoardStateResponse(game *gameengine.Game) boardStateResponse {
	legalMoves := game.LegalMoves
	piecesByPosition := make(map[gameengine.Position]gameengine.Piece, len(game.Pieces))
	for _, piece := range game.Pieces {
		if piece.Captured {
			continue
		}
		piecesByPosition[piece.Position] = piece
	}

	cells := make([]boardCell, 0, game.BoardSize*game.BoardSize)
	for row := game.BoardSize - 1; row >= 0; row-- {
		for col := 0; col < game.BoardSize; col++ {
			position := gameengine.Position{Row: row, Col: col}
			cell := boardCell{
				Row:         row,
				Col:         col,
				IsDark:      (row+col)%2 == 0,
				SquareLabel: squareLabel(position),
			}
			if piece, ok := piecesByPosition[position]; ok {
				cell.Piece = &boardPiece{
					ID:      piece.ID,
					Side:    string(piece.Side),
					Kind:    string(piece.Kind),
					Classes: pieceClasses(piece),
				}
			}
			cells = append(cells, cell)
		}
	}

	legalMovesByPiece := make(map[string][]legalMoveSummary)
	for _, move := range legalMoves {
		if len(move) < 2 {
			continue
		}

		piece, ok := piecesByPosition[move[0]]
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

	return boardStateResponse{
		Turn:              titleCaseTurn(game.Turn),
		BoardSize:         game.BoardSize,
		BoardCells:        cells,
		LegalMovesByPiece: legalMovesByPiece,
		GameOver:          gameOver,
	}
}

func squareLabel(position gameengine.Position) string {
	return string(rune('a'+position.Col)) + string(rune('1'+position.Row))
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
