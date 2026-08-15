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
		for col := row % 2; col < 8; col += 2 {
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
		for col := row % 2; col < 8; col += 2 {
			pieces = append(pieces, Piece{
				ID:       pieceID(PlayerSideRed, pieceIndex),
				Side:     PlayerSideRed,
				Kind:     PieceKindMan,
				Position: Position{Row: row, Col: col},
			})
			pieceIndex++
		}
	}

	game := Game{
		Turn:        PlayerSideBlack,
		BoardSize:   8,
		Pieces:      pieces,
		MoveHistory: []Move{},
	}
	computeLegalMoves(&game)
	return game
}

func pieceID(side PlayerSide, index int) string {
	return string(side) + "-" + strconv.Itoa(index+1)
}

// computeLegalMoves returns the legal moves for the current player.
//
// Jump and forced-capture rules are intentionally deferred to a later task.
func computeLegalMoves(game *Game) {
	game.LegalMoves = getJumpMoves(*game)
	if len(game.LegalMoves) > 0 {
		return
	}

	game.LegalMoves = getNonJumpMoves(*game)
	if len(game.LegalMoves) == 0 {
		game.GameOver = &GameOver{
			Winner: otherSide(game.Turn),
			Reason: "No more moves!",
		}
	}
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

			moves = append(moves, Move{piece.Position, destination})
		}
	}

	return moves
}

func getBoardCache(game Game) BoardCache {
	return getBoardCacheFromPieces(game.Pieces)
}

func getBoardCacheFromPieces(pieces []Piece) BoardCache {
	board := BoardCache{}
	for i := range pieces {
		piece := &pieces[i]
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
			moves = append(moves, Move{piece.Position, dest})
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

// ApplyMove applies a legal move in place.
//
// It returns nil when the move was applied. Invalid moves return an
// ApplyMoveError describing the reason. This implementation supports simple
// diagonal moves and single-jump captures. Multi-jump sequences, forced-capture
// enforcement, promotion, and terminal-state evaluation are deferred.
func ApplyMove(game *Game, move Move) *ApplyMoveError {
	if game == nil {
		return &ApplyMoveError{Reason: "game is nil"}
	}
	if game.GameOver != nil {
		return &ApplyMoveError{Reason: "game is over"}
	}
	if len(move) != 2 {
		return &ApplyMoveError{Reason: "move must contain exactly 2 positions"}
	}

	from := move[0]
	to := move[1]

	updatedPieces := append([]Piece(nil), game.Pieces...)
	pieceIndex := findActivePieceAt(updatedPieces, from)
	if pieceIndex == -1 {
		return &ApplyMoveError{Reason: "no active piece at move origin"}
	}

	piece := updatedPieces[pieceIndex]
	if piece.Side != game.Turn {
		return &ApplyMoveError{Reason: "piece does not belong to current player"}
	}

	if !isInsideBoard(from, game.BoardSize) || !isInsideBoard(to, game.BoardSize) {
		return &ApplyMoveError{Reason: "move is outside the board"}
	}

	board := getBoardCacheFromPieces(updatedPieces)
	if _, occupied := board[to]; occupied {
		return &ApplyMoveError{Reason: "destination is occupied"}
	}

	deltaRow := to.Row - from.Row
	deltaCol := to.Col - from.Col
	absRow := abs(deltaRow)
	absCol := abs(deltaCol)

	if absRow != absCol || (absRow != 1 && absRow != 2) {
		return &ApplyMoveError{Reason: "move must be a diagonal step or jump"}
	}

	if !canPieceMoveBy(piece, deltaRow, deltaCol) {
		return &ApplyMoveError{Reason: "piece cannot move in that direction"}
	}

	if absRow == 2 {
		middle := Position{
			Row: from.Row + deltaRow/2,
			Col: from.Col + deltaCol/2,
		}
		capturedIndex := findActivePieceAt(updatedPieces, middle)
		if capturedIndex == -1 {
			return &ApplyMoveError{Reason: "jump requires a piece to capture"}
		}
		if updatedPieces[capturedIndex].Side == piece.Side {
			return &ApplyMoveError{Reason: "cannot capture your own piece"}
		}
		updatedPieces[capturedIndex].Captured = true
	}

	if to.Row == kingRow(piece.Side) {
		updatedPieces[pieceIndex].Kind = PieceKindKing
	}

	updatedPieces[pieceIndex].Position = to
	game.Pieces = updatedPieces
	game.Turn = otherSide(game.Turn)
	game.MoveHistory = append(game.MoveHistory, append(Move(nil), move...))
	computeLegalMoves(game)

	return nil
}

func kingRow(side PlayerSide) int {
	if side == PlayerSideBlack {
		return 7
	}
	return 0
}

func findActivePieceAt(pieces []Piece, position Position) int {
	for i, piece := range pieces {
		if piece.Captured {
			continue
		}
		if piece.Position == position {
			return i
		}
	}
	return -1
}

func canPieceMoveBy(piece Piece, deltaRow int, deltaCol int) bool {
	if abs(deltaRow) != abs(deltaCol) {
		return false
	}

	if piece.Kind == PieceKindKing {
		return abs(deltaRow) == 1 || abs(deltaRow) == 2
	}

	switch piece.Side {
	case PlayerSideBlack:
		return deltaRow == 1 || deltaRow == 2
	case PlayerSideRed:
		return deltaRow == -1 || deltaRow == -2
	default:
		return false
	}
}

func otherSide(side PlayerSide) PlayerSide {
	if side == PlayerSideBlack {
		return PlayerSideRed
	}
	return PlayerSideBlack
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
