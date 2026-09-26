// Package testengine is a drop-in replacement for game-engine intended for
// testing movement code without full checkers rules:
//
//   - It is always Black's turn.
//   - A new game starts with a single black piece at (0,0).
//   - Black may move to any unoccupied board square each turn.
//   - Pieces are never promoted to king and opponents are never captured.
//
// Callers can swap this package for game-engine by changing the import path;
// all exported types are aliases of the shared game-types package.
package testengine

import gametypes "checkerbots/apps/server/game-types"

// Re-export shared types as aliases so callers can reference them via the
// testengine package, matching the same API as game-engine.
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

// NewGame8x8 returns an 8x8 test game with a single black piece at (0,0).
func NewGame8x8() Game {
	return GameFromStored(StoredGame{
		Turn:      gametypes.PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{{
			ID:       "black-1",
			Side:     gametypes.PlayerSideBlack,
			Kind:     gametypes.PieceKindMan,
			Position: Position{Row: 0, Col: 0},
		}},
		MoveHistory: []Move{},
	})
}

// GameFromStored reconstructs a test game from stored state and recomputes legal moves.
func GameFromStored(stored StoredGame) Game {
	game := Game{StoredGame: stored}
	computeLegalMoves(&game)
	return game
}

// computeLegalMoves populates game.LegalMoves with every move from a black
// piece to any unoccupied square on the board.
func computeLegalMoves(game *Game) {
	occupied := make(map[Position]bool, len(game.Pieces))
	for _, piece := range game.Pieces {
		if !piece.Captured {
			occupied[piece.Position] = true
		}
	}

	var moves []Move
	for _, piece := range game.Pieces {
		if piece.Captured || piece.Side != game.Turn {
			continue
		}
		for row := 0; row < game.BoardSize; row++ {
			for col := 0; col < game.BoardSize; col++ {
				dest := Position{Row: row, Col: col}
				if dest == piece.Position || occupied[dest] {
					continue
				}
				moves = append(moves, Move{piece.Position, dest})
			}
		}
	}
	game.LegalMoves = moves
}

// ApplyMove moves a black piece to any unoccupied board square.
// The turn never changes; king promotion and captures are not handled.
func ApplyMove(game *Game, move Move) *ApplyMoveError {
	if game == nil {
		return &ApplyMoveError{Reason: "game is nil"}
	}
	if game.GameOver != nil {
		return &ApplyMoveError{Reason: "game is over"}
	}
	if len(move) < 2 {
		return &ApplyMoveError{Reason: "move must contain at least 2 positions"}
	}
	if len(move) > 2 {
		return &ApplyMoveError{Reason: "multi-step moves are not supported"}
	}

	from, to := move[0], move[1]

	pieceIndex := -1
	for i, p := range game.Pieces {
		if !p.Captured && p.Position == from {
			pieceIndex = i
			break
		}
	}
	if pieceIndex == -1 {
		return &ApplyMoveError{Reason: "no active piece at move origin"}
	}
	if game.Pieces[pieceIndex].Side != game.Turn {
		return &ApplyMoveError{Reason: "piece does not belong to current player"}
	}
	if !isInsideBoard(to, game.BoardSize) {
		return &ApplyMoveError{Reason: "destination is outside the board"}
	}
	for _, p := range game.Pieces {
		if !p.Captured && p.Position == to {
			return &ApplyMoveError{Reason: "destination is occupied"}
		}
	}

	updatedPieces := append([]Piece(nil), game.Pieces...)
	updatedPieces[pieceIndex].Position = to
	game.Pieces = updatedPieces
	game.MoveHistory = append(game.MoveHistory, append(Move(nil), move...))
	computeLegalMoves(game)

	return nil
}

func isInsideBoard(pos Position, boardSize int) bool {
	return pos.Row >= 0 && pos.Row < boardSize && pos.Col >= 0 && pos.Col < boardSize
}
