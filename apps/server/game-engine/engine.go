package gameengine

import (
	"strconv"
)

// TODO: Use 2d array? Track this in `Game` for easy lookups later?
type boardCache map[Position]*Piece

func NewGame(config GameConfig) Game {
	if config.BoardSize <= 0 {
		panic("BoardSize must be positive")
	}
	if config.BoardSize%2 != 0 {
		panic("BoardSize must be even")
	}
	if config.InitialRows <= 0 {
		panic("InitialRows must be positive")
	}

	pieces := make([]Piece, 0, config.InitialRows*config.BoardSize/2)

	pieceIndex := 0
	for row := 0; row < config.InitialRows; row++ {
		for col := row % 2; col < config.BoardSize; col += 2 {
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
	for row := config.BoardSize - 1; row >= config.BoardSize-config.InitialRows; row-- {
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
		BoardSize:   config.BoardSize,
		Pieces:      pieces,
		MoveHistory: []Move{},
	}
	computeLegalMoves(&game)
	return game
}

func NewGame8x8() Game {
	return NewGame(GameConfig{
		BoardSize:   8,
		InitialRows: 3,
	})
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

func getBoardCache(game Game) boardCache {
	return getBoardCacheFromPieces(game.Pieces)
}

func getBoardCacheFromPieces(pieces []Piece) boardCache {
	board := boardCache{}
	for i := range pieces {
		piece := &pieces[i]
		if piece.Captured {
			continue
		}
		board[piece.Position] = piece
	}
	return board
}

// jumpStep describes a single jump: the position of the captured piece and
// the landing square.
type jumpStep struct {
	Over Position
	Dest Position
}

func getJumpStepsFrom(game Game, board boardCache, side PlayerSide, kind PieceKind, position Position, captured map[Position]bool) []jumpStep {
	// TODO: Pre-allocate capacity of 2 (4 for king) since that's the most a piece can every have?
	steps := []jumpStep{}
	for _, delta := range moveDeltasFor(side, kind) {
		over := Position{
			Row: position.Row + delta.Row,
			Col: position.Col + delta.Col,
		}
		destination := Position{
			Row: position.Row + 2*delta.Row,
			Col: position.Col + 2*delta.Col,
		}
		if !isInsideBoard(destination, game.BoardSize) {
			continue
		}
		// A king can't jump the same piece twice within a single sequence.
		if captured[over] {
			continue
		}
		if other, exists := board[over]; !exists || other.Side == side {
			continue
		}
		if _, exists := board[destination]; exists {
			continue
		}

		steps = append(steps, jumpStep{Over: over, Dest: destination})
	}
	return steps
}

// getJumpSequences recursively explores all legal multi-jump paths available
// to a single piece, starting from its current position.
//
// board reflects the state of the board with the moving piece's starting
// square cleared out (since it's no longer there once it starts jumping),
// but with all other pieces - including ones already captured earlier in
// this same sequence - still present, since real captures aren't resolved
// until the whole move is applied. captured tracks the positions already
// jumped-over in this sequence so they can't be captured again.
//
// A sequence stops as soon as a man reaches its promotion row - it can't
// keep jumping in the same turn. A piece that starts (or already became) a
// king isn't affected by the promotion row and keeps jumping normally.
func getJumpSequences(game Game, board boardCache, side PlayerSide, kind PieceKind, position Position, captured map[Position]bool, path Move) []Move {
	steps := getJumpStepsFrom(game, board, side, kind, position, captured)
	if len(steps) == 0 {
		if len(path) > 1 {
			return []Move{append(Move{}, path...)}
		}
		return nil
	}

	moves := []Move{}
	for _, step := range steps {
		newCaptured := make(map[Position]bool, len(captured)+1)
		for pos := range captured {
			newCaptured[pos] = true
		}
		newCaptured[step.Over] = true

		newPath := append(append(Move{}, path...), step.Dest)

		if kind == PieceKindMan && step.Dest.Row == kingRow(side) {
			// Promotion ends the jump sequence immediately.
			moves = append(moves, newPath)
			continue
		}

		moves = append(moves, getJumpSequences(game, board, side, kind, step.Dest, newCaptured, newPath)...)
	}
	return moves
}

func getJumpMoves(game Game) []Move {
	board := getBoardCache(game)
	moves := []Move{}
	for _, piece := range game.Pieces {
		if piece.Captured || piece.Side != game.Turn {
			continue
		}

		movingBoard := copyBoardCache(board)
		delete(movingBoard, piece.Position)

		sequences := getJumpSequences(game, movingBoard, piece.Side, piece.Kind, piece.Position, map[Position]bool{}, Move{piece.Position})
		moves = append(moves, sequences...)
	}

	return moves
}

func copyBoardCache(board boardCache) boardCache {
	copied := make(boardCache, len(board))
	for pos, piece := range board {
		copied[pos] = piece
	}
	return copied
}

func moveDeltasForPiece(piece Piece) []Position {
	return moveDeltasFor(piece.Side, piece.Kind)
}

func moveDeltasFor(side PlayerSide, kind PieceKind) []Position {
	if kind == PieceKindKing {
		return []Position{
			{Row: -1, Col: -1},
			{Row: -1, Col: 1},
			{Row: 1, Col: -1},
			{Row: 1, Col: 1},
		}
	}

	if side == PlayerSideBlack {
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
// ApplyMoveError describing the reason. Simple diagonal moves and single-jump
// captures are validated step-by-step below. Multi-jump sequences (moves with
// more than 2 positions) are instead validated by matching them exactly
// against game.LegalMoves, since that list already encodes all multi-jump
// rules (no double-capturing a piece, promotion ending a sequence, etc.) -
// re-deriving those rules here would just duplicate computeLegalMoves.
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
		return applyMultiStepMove(game, move)
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

// applyMultiStepMove applies a move with more than 2 positions, i.e. a
// multi-jump sequence. It's only valid if it exactly matches one of the
// current player's precomputed legal moves.
func applyMultiStepMove(game *Game, move Move) *ApplyMoveError {
	var matched Move
	for _, legalMove := range game.LegalMoves {
		if movePathsEqual(legalMove, move) {
			matched = legalMove
			break
		}
	}
	if matched == nil {
		return &ApplyMoveError{Reason: "move is not a legal move"}
	}

	updatedPieces := append([]Piece(nil), game.Pieces...)
	pieceIndex := findActivePieceAt(updatedPieces, matched[0])
	if pieceIndex == -1 {
		return &ApplyMoveError{Reason: "no active piece at move origin"}
	}
	piece := updatedPieces[pieceIndex]

	for i := 0; i < len(matched)-1; i++ {
		from := matched[i]
		to := matched[i+1]
		if abs(to.Row-from.Row) != 2 {
			continue
		}
		middle := Position{
			Row: from.Row + (to.Row-from.Row)/2,
			Col: from.Col + (to.Col-from.Col)/2,
		}
		capturedIndex := findActivePieceAt(updatedPieces, middle)
		if capturedIndex == -1 {
			return &ApplyMoveError{Reason: "jump requires a piece to capture"}
		}
		updatedPieces[capturedIndex].Captured = true
	}

	destination := matched[len(matched)-1]
	if destination.Row == kingRow(piece.Side) {
		updatedPieces[pieceIndex].Kind = PieceKindKing
	}
	updatedPieces[pieceIndex].Position = destination

	game.Pieces = updatedPieces
	game.Turn = otherSide(game.Turn)
	game.MoveHistory = append(game.MoveHistory, append(Move(nil), matched...))
	computeLegalMoves(game)

	return nil
}

func movePathsEqual(a Move, b Move) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
