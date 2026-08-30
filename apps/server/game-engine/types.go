package gameengine

// PlayerSide identifies which player a piece belongs to.
type PlayerSide string

const (
	PlayerSideRed   PlayerSide = "red"
	PlayerSideBlack PlayerSide = "black"
)

// PieceKind models a standard piece or a king.
type PieceKind string

const (
	PieceKindMan  PieceKind = "man"
	PieceKindKing PieceKind = "king"
)

// Position identifies a square on the board using zero-based row and column indices.
// (0,0) is the bottom-left corner of the black side.
type Position struct {
	Row int
	Col int
}

// Piece represents one logical checkers piece.
type Piece struct {
	ID       string
	Side     PlayerSide
	Kind     PieceKind
	Position Position
	Captured bool
}

// Move represents a move path as an ordered list of visited positions.
//
// A simple step or single jump is represented as exactly two positions.
// Multi-jump sequences are represented as the full chain of positions
// visited, e.g. [start, afterJump1, afterJump2, ...].
type Move []Position

// ApplyMoveError describes why a move could not be applied.
type ApplyMoveError struct {
	Reason string
}

// Error implements the error interface.
func (e *ApplyMoveError) Error() string {
	if e == nil {
		return ""
	}
	return e.Reason
}

// Config for creating a new game.
type GameConfig struct {
	BoardSize   int
	InitialRows int
}

// Game holds the rules-engine-owned game state.
//
// Board is included as an explicit future home for square occupancy, while Pieces
// keeps the initial API simple and easy to evolve in subsequent tasks.
type Game struct {
	StoredGame
	LegalMoves []Move
}

// CapturedRedPieces returns references to captured red pieces, in the order
// they were captured.
func (g *Game) CapturedRedPieces() []*Piece {
	pieces := make([]*Piece, len(g.CapturedRedPieceIndices))
	for i, index := range g.CapturedRedPieceIndices {
		pieces[i] = &g.Pieces[index]
	}
	return pieces
}

// CapturedBlackPieces returns references to captured black pieces, in the
// order they were captured.
func (g *Game) CapturedBlackPieces() []*Piece {
	pieces := make([]*Piece, len(g.CapturedBlackPieceIndices))
	for i, index := range g.CapturedBlackPieceIndices {
		pieces[i] = &g.Pieces[index]
	}
	return pieces
}

// GameOver summarizes terminal-game evaluation without committing future tasks to
// a richer result model yet.
type GameOver struct {
	Winner PlayerSide
	Reason string
}

// Minimal state needed for storing/loading a game
type StoredGame struct {
	Turn        PlayerSide
	BoardSize   int
	Pieces      []Piece
	MoveHistory []Move
	// Indices into Pieces of captured pieces, in the order they were
	// captured.
	CapturedRedPieceIndices   []int
	CapturedBlackPieceIndices []int
	// `nil` if game is in-progress
	GameOver *GameOver
}
