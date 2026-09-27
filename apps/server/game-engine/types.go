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

type PieceID string

// Piece represents one logical checkers piece.
type Piece struct {
	ID       PieceID
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
//
// TODO: Define a move to just be a pair of positions and just don't complete
// (switch sides) until a move sequence is finished?
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

// GameResult summarizes terminal-game evaluation without committing future tasks to
// a richer result model yet.
type GameResult struct {
	Winner PlayerSide
	Reason string
}

// GameConfig holds the parameters for creating a new game. It is internal to
// game-engine; test-engine has its own fixed starting configuration.
type GameConfig struct {
	BoardSize   int
	InitialRows int
	// CaptureColumns is the number of extra columns reserved on each side of
	// the board for captured pieces. See StoredGame.CaptureColumns.
	CaptureColumns int
}
