package gameengine

import "strconv"

// NewGame creates a standard starting game state for checkers.
func NewGame() Game {
	pieces := make([]Piece, 0, 24)

	pieceIndex := 0
	for row := 0; row < 3; row++ {
		for col := 0; col < 8; col++ {
			if (row+col)%2 != 0 {
				continue
			}

			pieces = append(pieces, Piece{
				ID:       pieceID(PlayerSideBlack, pieceIndex),
				Side:     PlayerSideBlack,
				Kind:     PieceKindMan,
				Position: Position{Row: row, Col: col},
			})
			pieceIndex++
		}
	}

	pieceIndex = 0
	for row := 5; row < 8; row++ {
		for col := 0; col < 8; col++ {
			if (row+col)%2 != 0 {
				continue
			}

			pieces = append(pieces, Piece{
				ID:       pieceID(PlayerSideRed, pieceIndex),
				Side:     PlayerSideRed,
				Kind:     PieceKindMan,
				Position: Position{Row: row, Col: col},
			})
			pieceIndex++
		}
	}

	return Game{
		Turn:        PlayerSideBlack,
		BoardSize:   8,
		Pieces:      pieces,
		MoveHistory: []Move{},
	}
}

func pieceID(side PlayerSide, index int) string {
	return string(side) + "-" + strconv.Itoa(index+1)
}

// GetLegalMoves returns the legal moves for the current player.
//
// This placeholder keeps the package importable while the rules implementation is
// built out in later tasks.
func GetLegalMoves(game Game) []Move {
	_ = game
	return []Move{}
}

// ApplyMove applies a move and returns the updated game state.
//
// This placeholder returns the game unchanged until move validation and board
// mutation are implemented.
func ApplyMove(game Game, move Move) Game {
	_ = move
	return game
}
