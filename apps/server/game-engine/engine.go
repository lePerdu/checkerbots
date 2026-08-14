package gameengine

import (
	"strconv"
)

// TODO: Use 2d array? Track this in `Game`?
type BoardCache map[Position]*Piece

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

// GetLegalMoves returns the legal non-capturing moves for the current player.
//
// Jump and forced-capture rules are intentionally deferred to a later task.
func GetLegalMoves(game Game) []Move {
	// jumps := getJumpMoves(game)
	// if len(jumps) > 0 {
	// 	return jumps
	// }

	return getNonJumpMoves(game)
}

func getNonJumpMoves(game Game) []Move {
	occupied := map[Position]bool{}
	for _, piece := range game.Pieces {
		if piece.Captured {
			continue
		}
		occupied[piece.Position] = true
	}

	moves := []Move{}
	for _, piece := range game.Pieces {
		if piece.Captured || piece.Side != game.Turn {
			continue
		}

		for _, delta := range moveDeltasForPiece(piece) {
			destination := Position{
				Row: piece.Position.Row + delta.Row,
				Col: piece.Position.Col + delta.Col,
			}
			if !isInsideBoard(destination, game.BoardSize) {
				continue
			}
			if occupied[destination] {
				continue
			}

			moves = append(moves, Move{
				From: piece.Position,
				To:   destination,
			})
		}
	}

	return moves
}

func getBoardCache(game Game) BoardCache {
	board := BoardCache{}
	for i := range game.Pieces {
		piece := &game.Pieces[i]
		if piece.Captured {
			continue
		}
		board[piece.Position] = piece
	}
	return board
}

func getJumpStepsForPiece(game Game, board BoardCache, piece Piece) []Position {
	// TODO: Pre-allocate capacity of 2 (4 for king) since that's the most a piece can every have?
	destinations := []Position{}
	for _, delta := range moveDeltasForPiece(piece) {
		over := Position{
			Row: piece.Position.Row + delta.Row,
			Col: piece.Position.Col + delta.Col,
		}
		destination := Position{
			Row: piece.Position.Row + 2*delta.Row,
			Col: piece.Position.Col + 2*delta.Col,
		}
		if !isInsideBoard(destination, game.BoardSize) {
			continue
		}
		if other, exists := board[over]; !exists || other.Side == piece.Side {
			continue
		}
		if _, exists := board[destination]; exists {
			continue
		}

		destinations = append(destinations, destination)
	}
	return destinations
}

func getJumpMoves(game Game) []Move {
	board := getBoardCache(game)
	moves := []Move{}
	for _, piece := range game.Pieces {
		if piece.Captured || piece.Side != game.Turn {
			continue
		}

		for _, dest := range getJumpStepsForPiece(game, board, piece) {
			moves = append(moves, Move{
				From: piece.Position,
				To:   dest,
			})
		}
	}

	return moves
}

func moveDeltasForPiece(piece Piece) []Position {
	if piece.Kind == PieceKindKing {
		return []Position{
			{Row: -1, Col: -1},
			{Row: -1, Col: 1},
			{Row: 1, Col: -1},
			{Row: 1, Col: 1},
		}
	}

	if piece.Side == PlayerSideBlack {
		return []Position{
			{Row: 1, Col: -1},
			{Row: 1, Col: 1},
		}
	}

	return []Position{
		{Row: -1, Col: -1},
		{Row: -1, Col: 1},
	}
}

func isInsideBoard(position Position, boardSize int) bool {
	return position.Row >= 0 && position.Row < boardSize && position.Col >= 0 && position.Col < boardSize
}

// ApplyMove applies a move and returns the updated game state.
//
// This placeholder returns the game unchanged until move validation and board
// mutation are implemented.
func ApplyMove(game Game, move Move) Game {
	_ = move
	return game
}
