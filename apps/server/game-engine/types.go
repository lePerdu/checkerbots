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
	// CaptureColumns is the number of extra columns reserved on each side of
	// the board for captured pieces. See Game.CaptureColumns.
	CaptureColumns int
}

// Game holds the rules-engine-owned game state.
//
// Board is included as an explicit future home for square occupancy, while Pieces
// keeps the initial API simple and easy to evolve in subsequent tasks.
type Game struct {
	StoredGame
	LegalMoves []Move
}

// GameOver summarizes terminal-game evaluation without committing future tasks to
// a richer result model yet.
type GameOver struct {
	Winner PlayerSide
	Reason string
}

// Minimal state needed for storing/loading a game
type StoredGame struct {
	Turn      PlayerSide
	BoardSize int
	// CaptureColumns is the number of extra columns reserved on each side of
	// the board for captured pieces: red's captured pieces sit in columns
	// [BoardSize, BoardSize+CaptureColumns), black's sit in columns
	// [-CaptureColumns, 0). See capturedPiecePosition.
	CaptureColumns int
	Pieces         []Piece
	MoveHistory    []Move
	// `nil` if game is in-progress
	GameOver *GameOver
}
