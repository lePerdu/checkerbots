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

func TestGetLegalMovesReturnsEmptySliceForPlaceholder(t *testing.T) {
	moves := GetLegalMoves(NewGame())
	if len(moves) != 0 {
		t.Fatalf("expected no legal moves in placeholder implementation, got %d", len(moves))
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
