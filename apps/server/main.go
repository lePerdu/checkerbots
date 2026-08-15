package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

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

type legalMoveSummary struct {
	From squarePosition `json:"from"`
	To   squarePosition `json:"to"`
}

type squarePosition struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type applyMoveRequest struct {
	From squarePosition `json:"from"`
	To   squarePosition `json:"to"`
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
	Label   string `json:"label"`
	Classes string `json:"classes"`
}

func main() {
	tmpl := template.Must(template.ParseFS(assets, "templates/index.html"))

	staticFS, err := fs.Sub(assets, "static")
	if err != nil {
		log.Fatalf("load static assets: %v", err)
	}

	game := gameengine.NewGame()

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
		if err := json.NewEncoder(w).Encode(buildBoardStateResponse(&game)); err != nil {
			log.Printf("encode board response: %v", err)
		}
	})
	mux.HandleFunc("POST /api/moves", func(w http.ResponseWriter, r *http.Request) {
		var request applyMoveRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		move := gameengine.Move{
			{Row: request.From.Row, Col: request.From.Col},
			{Row: request.To.Row, Col: request.To.Col},
		}
		if err := gameengine.ApplyMove(&game, move); err != nil {
			http.Error(w, err.Reason, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(buildBoardStateResponse(&game)); err != nil {
			log.Printf("encode apply move response: %v", err)
		}
	})
	mux.HandleFunc("POST /games/new", func(w http.ResponseWriter, r *http.Request) {
		game = gameengine.NewGame()

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(buildBoardStateResponse(&game)); err != nil {
			log.Printf("encode new game response: %v", err)
		}
	})

	addr := ":8080"
	if value := os.Getenv("PORT"); value != "" {
		addr = ":" + value
	}

	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
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
					Label:   pieceGlyph(piece),
					Classes: "piece piece--" + string(piece.Side),
				}
			}
			cells = append(cells, cell)
		}
	}

	legalMovesByPiece := make(map[string][]legalMoveSummary)
	for _, move := range legalMoves {
		if len(move) != 2 {
			continue
		}

		piece, ok := piecesByPosition[move[0]]
		if !ok {
			continue
		}

		legalMovesByPiece[piece.ID] = append(legalMovesByPiece[piece.ID], legalMoveSummary{
			From: squarePosition{Row: move[0].Row, Col: move[0].Col},
			To:   squarePosition{Row: move[1].Row, Col: move[1].Col},
		})
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

func pieceGlyph(piece gameengine.Piece) string {
	if piece.Kind == gameengine.PieceKindKing {
		return "K"
	}
	return "●"
}

func titleCaseTurn(side gameengine.PlayerSide) string {
	if side == gameengine.PlayerSideBlack {
		return "Black"
	}
	return "Red"
}
