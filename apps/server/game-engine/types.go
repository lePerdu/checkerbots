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

// Move represents a single move from one square to another.
//
// The shape is intentionally small for now. Later tasks can extend it with
// captures, multi-jump paths, promotion metadata, or validation details.
type Move struct {
	From Position
	To   Position
}

// Game holds the rules-engine-owned game state.
//
// Board is included as an explicit future home for square occupancy, while Pieces
// keeps the initial API simple and easy to evolve in subsequent tasks.
type Game struct {
	Turn        PlayerSide
	BoardSize   int
	Pieces      []Piece
	MoveHistory []Move
	// `nil` if game is in-progress
	GameOver *GameOver
}

// GameOver summarizes terminal-game evaluation without committing future tasks to
// a richer result model yet.
type GameOver struct {
	Winner PlayerSide
	Reason string
}
