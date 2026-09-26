package gameengine

import gametypes "checkerbots/apps/server/game-types"

// Shared types are defined in game-types and re-exported here as aliases so
// callers can reference them via either package interchangeably.
type (
	PlayerSide     = gametypes.PlayerSide
	PieceKind      = gametypes.PieceKind
	Position       = gametypes.Position
	PieceID        = gametypes.PieceID
	Piece          = gametypes.Piece
	Move           = gametypes.Move
	ApplyMoveError = gametypes.ApplyMoveError
	Game           = gametypes.Game
	GameOver       = gametypes.GameOver
	StoredGame     = gametypes.StoredGame
)

const (
	PlayerSideRed   = gametypes.PlayerSideRed
	PlayerSideBlack = gametypes.PlayerSideBlack
	PieceKindMan    = gametypes.PieceKindMan
	PieceKindKing   = gametypes.PieceKindKing
)

// GameConfig holds the parameters for creating a new game. It is internal to
// game-engine; test-engine has its own fixed starting configuration.
type GameConfig struct {
	BoardSize   int
	InitialRows int
	// CaptureColumns is the number of extra columns reserved on each side of
	// the board for captured pieces. See StoredGame.CaptureColumns.
	CaptureColumns int
}
