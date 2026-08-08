package gameengine

// NewGame creates an empty, active game state scaffold ready for later rules work.
//
// Later tasks will populate the starting board and piece layout.
func NewGame() Game {
	return Game{
		Turn:        PlayerSideBlack,
		BoardSize:   8,
		Pieces:      []Piece{},
		MoveHistory: []Move{},
	}
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
