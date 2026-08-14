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

type pageData struct {
	Turn       string
	BoardSize  int
	BoardCells []boardCell
}

type boardCell struct {
	Row         int
	Col         int
	IsDark      bool
	Piece       *boardPiece
	SquareLabel string
}

type boardPiece struct {
	Side    string
	Kind    string
	Label   string
	Classes string
}

func main() {
	tmpl := template.Must(template.ParseFS(assets, "templates/index.html"))

	staticFS, err := fs.Sub(assets, "static")
	if err != nil {
		log.Fatalf("load static assets: %v", err)
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

		game := gameengine.NewGame()
		data := buildPageData(game)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("render index: %v", err)
		}
	})
	mux.HandleFunc("POST /games/new", func(w http.ResponseWriter, r *http.Request) {
		game := gameengine.NewGame()
		response := struct {
			Turn string `json:"turn"`
		}{Turn: string(game.Turn)}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(response); err != nil {
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

func buildPageData(game gameengine.Game) pageData {
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
					Side:    string(piece.Side),
					Kind:    string(piece.Kind),
					Label:   pieceGlyph(piece),
					Classes: "piece piece--" + string(piece.Side),
				}
			}
			cells = append(cells, cell)
		}
	}

	return pageData{
		Turn:       titleCaseTurn(game.Turn),
		BoardSize:  game.BoardSize,
		BoardCells: cells,
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
