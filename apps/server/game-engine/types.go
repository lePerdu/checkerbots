package gameengine

import "time"

// GameStatus summarizes lifecycle state for the single active game MVP.
type GameStatus string

const (
	GameStatusSetup    GameStatus = "setup"
	GameStatusActive   GameStatus = "active"
	GameStatusFinished GameStatus = "finished"
	GameStatusAborted  GameStatus = "aborted"
)

// PlayerSide identifies which player or team a piece belongs to.
type PlayerSide string

const (
	PlayerSideRed   PlayerSide = "red"
	PlayerSideBlack PlayerSide = "black"
)

// PieceKind models the current MVP need for standard checkers pieces and kings.
type PieceKind string

const (
	PieceKindMan  PieceKind = "man"
	PieceKindKing PieceKind = "king"
)

// Piece represents one logical game piece on the board.
type Piece struct {
	ID       string     `json:"id"`
	Side     PlayerSide `json:"side"`
	Kind     PieceKind  `json:"kind"`
	Row      int        `json:"row"`
	Col      int        `json:"col"`
	Captured bool       `json:"captured"`
}

// Game is the shared server/UI view of a single checkers game.
//
// The shape is intentionally single-game friendly for MVP use, while Kind and
// Status leave room for future extension if the project later supports more
// than one game type or richer lifecycle handling.
type Game struct {
	ID        string      `json:"id"`
	Status    GameStatus  `json:"status"`
	Turn      PlayerSide  `json:"turn"`
	Pieces    []Piece     `json:"pieces"`
	MoveCount int         `json:"move_count"`
	Winner    *PlayerSide `json:"winner,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
