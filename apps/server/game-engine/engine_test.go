package gameengine

import (
	"reflect"
	"testing"
)

func TestNewGameReturnsStandardStartingBoard(t *testing.T) {
	game := NewGame()

	if game.GameOver != nil {
		t.Fatalf("expected nil GameOver, got %#v", game.GameOver)
	}

	if game.Turn != PlayerSideBlack {
		t.Fatalf("expected turn %q, got %q", PlayerSideBlack, game.Turn)
	}

	if game.BoardSize != 8 {
		t.Fatalf("expected board size %d, got %d", 8, game.BoardSize)
	}

	if len(game.MoveHistory) != 0 {
		t.Fatalf("expected empty move history, got %d entries", len(game.MoveHistory))
	}

	if len(game.Pieces) != 24 {
		t.Fatalf("expected 24 pieces, got %d", len(game.Pieces))
	}

	expectedBlack := map[Position]bool{
		{Row: 0, Col: 0}: true,
		{Row: 0, Col: 2}: true,
		{Row: 0, Col: 4}: true,
		{Row: 0, Col: 6}: true,
		{Row: 1, Col: 1}: true,
		{Row: 1, Col: 3}: true,
		{Row: 1, Col: 5}: true,
		{Row: 1, Col: 7}: true,
		{Row: 2, Col: 0}: true,
		{Row: 2, Col: 2}: true,
		{Row: 2, Col: 4}: true,
		{Row: 2, Col: 6}: true,
	}

	expectedRed := map[Position]bool{
		{Row: 5, Col: 1}: true,
		{Row: 5, Col: 3}: true,
		{Row: 5, Col: 5}: true,
		{Row: 5, Col: 7}: true,
		{Row: 6, Col: 0}: true,
		{Row: 6, Col: 2}: true,
		{Row: 6, Col: 4}: true,
		{Row: 6, Col: 6}: true,
		{Row: 7, Col: 1}: true,
		{Row: 7, Col: 3}: true,
		{Row: 7, Col: 5}: true,
		{Row: 7, Col: 7}: true,
	}

	seenIDs := map[string]bool{}
	blackCount := 0
	redCount := 0

	for _, piece := range game.Pieces {
		if seenIDs[piece.ID] {
			t.Fatalf("duplicate piece ID found: %q", piece.ID)
		}
		seenIDs[piece.ID] = true

		if piece.Captured {
			t.Fatalf("expected piece %q to start uncaptured", piece.ID)
		}

		if piece.Kind != PieceKindMan {
			t.Fatalf("expected piece %q to start as a man, got %q", piece.ID, piece.Kind)
		}

		if (piece.Position.Row+piece.Position.Col)%2 != 0 {
			t.Fatalf("expected piece %q to be on a playable dark square, got %+v", piece.ID, piece.Position)
		}

		switch piece.Side {
		case PlayerSideBlack:
			blackCount++
			if !expectedBlack[piece.Position] {
				t.Fatalf("unexpected black piece position: %+v", piece.Position)
			}
			delete(expectedBlack, piece.Position)
		case PlayerSideRed:
			redCount++
			if !expectedRed[piece.Position] {
				t.Fatalf("unexpected red piece position: %+v", piece.Position)
			}
			delete(expectedRed, piece.Position)
		default:
			t.Fatalf("unexpected player side: %q", piece.Side)
		}
	}

	if blackCount != 12 {
		t.Fatalf("expected 12 black pieces, got %d", blackCount)
	}

	if redCount != 12 {
		t.Fatalf("expected 12 red pieces, got %d", redCount)
	}

	if len(expectedBlack) != 0 {
		t.Fatalf("missing black pieces at positions: %+v", expectedBlack)
	}

	if len(expectedRed) != 0 {
		t.Fatalf("missing red pieces at positions: %+v", expectedRed)
	}
}

func TestGetLegalMovesReturnsSimpleOpeningMovesForBlack(t *testing.T) {
	moves := GetLegalMoves(NewGame())

	expected := map[Move]bool{
		{From: Position{Row: 2, Col: 0}, To: Position{Row: 3, Col: 1}}: true,
		{From: Position{Row: 2, Col: 2}, To: Position{Row: 3, Col: 1}}: true,
		{From: Position{Row: 2, Col: 2}, To: Position{Row: 3, Col: 3}}: true,
		{From: Position{Row: 2, Col: 4}, To: Position{Row: 3, Col: 3}}: true,
		{From: Position{Row: 2, Col: 4}, To: Position{Row: 3, Col: 5}}: true,
		{From: Position{Row: 2, Col: 6}, To: Position{Row: 3, Col: 5}}: true,
		{From: Position{Row: 2, Col: 6}, To: Position{Row: 3, Col: 7}}: true,
	}

	if len(moves) != len(expected) {
		t.Fatalf("expected %d moves, got %d: %+v", len(expected), len(moves), moves)
	}

	for _, move := range moves {
		if !expected[move] {
			t.Fatalf("unexpected move returned: %+v", move)
		}
		delete(expected, move)
	}

	if len(expected) != 0 {
		t.Fatalf("expected moves not returned: %+v", expected)
	}
}

func TestGetLegalMovesReturnsSimpleMovesForRedTurn(t *testing.T) {
	game := NewGame()
	game.Turn = PlayerSideRed

	moves := GetLegalMoves(game)

	expected := map[Move]bool{
		{From: Position{Row: 5, Col: 1}, To: Position{Row: 4, Col: 0}}: true,
		{From: Position{Row: 5, Col: 1}, To: Position{Row: 4, Col: 2}}: true,
		{From: Position{Row: 5, Col: 3}, To: Position{Row: 4, Col: 2}}: true,
		{From: Position{Row: 5, Col: 3}, To: Position{Row: 4, Col: 4}}: true,
		{From: Position{Row: 5, Col: 5}, To: Position{Row: 4, Col: 4}}: true,
		{From: Position{Row: 5, Col: 5}, To: Position{Row: 4, Col: 6}}: true,
		{From: Position{Row: 5, Col: 7}, To: Position{Row: 4, Col: 6}}: true,
	}

	if len(moves) != len(expected) {
		t.Fatalf("expected %d moves, got %d: %+v", len(expected), len(moves), moves)
	}

	for _, move := range moves {
		if !expected[move] {
			t.Fatalf("unexpected move returned: %+v", move)
		}
		delete(expected, move)
	}

	if len(expected) != 0 {
		t.Fatalf("expected moves not returned: %+v", expected)
	}
}

func TestGetLegalMovesSkipsBlockedAndCapturedPieces(t *testing.T) {
	game := Game{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-1", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 2, Col: 2}},
			{ID: "black-2", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 4, Col: 4}, Captured: true},
			{ID: "red-1", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 3, Col: 1}},
			{ID: "red-2", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 3, Col: 3}},
		},
		MoveHistory: []Move{},
	}

	moves := GetLegalMoves(game)
	if len(moves) != 0 {
		t.Fatalf("expected no simple moves for a blocked piece, got %+v", moves)
	}
}

func TestGetLegalMovesIncludesBackwardMovesForKings(t *testing.T) {
	game := Game{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-king-1", Side: PlayerSideBlack, Kind: PieceKindKing, Position: Position{Row: 3, Col: 3}},
		},
		MoveHistory: []Move{},
	}

	moves := GetLegalMoves(game)

	expected := map[Move]bool{
		{From: Position{Row: 3, Col: 3}, To: Position{Row: 2, Col: 2}}: true,
		{From: Position{Row: 3, Col: 3}, To: Position{Row: 2, Col: 4}}: true,
		{From: Position{Row: 3, Col: 3}, To: Position{Row: 4, Col: 2}}: true,
		{From: Position{Row: 3, Col: 3}, To: Position{Row: 4, Col: 4}}: true,
	}

	if len(moves) != len(expected) {
		t.Fatalf("expected %d moves, got %d: %+v", len(expected), len(moves), moves)
	}

	for _, move := range moves {
		if !expected[move] {
			t.Fatalf("unexpected move returned: %+v", move)
		}
		delete(expected, move)
	}

	if len(expected) != 0 {
		t.Fatalf("expected moves not returned: %+v", expected)
	}
}

func TestApplyMoveReturnsGameUnchangedForPlaceholder(t *testing.T) {
	game := NewGame()
	move := Move{From: Position{Row: 2, Col: 1}, To: Position{Row: 3, Col: 0}}

	updated := ApplyMove(game, move)

	if !reflect.DeepEqual(updated, game) {
		t.Fatalf("expected placeholder ApplyMove to return the input game unchanged")
	}
}
