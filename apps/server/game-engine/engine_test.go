package gameengine

import (
	"reflect"
	"testing"
)

func TestNewGameReturnsInitialScaffold(t *testing.T) {
	game := NewGame()

	if game.GameOver != nil {
		t.Fatalf("expected nil GameOver, got %q", game.GameOver)
	}

	if game.Turn != PlayerSideBlack {
		t.Fatalf("expected turn %q, got %q", PlayerSideBlack, game.Turn)
	}

	if game.BoardSize != 8 {
		t.Fatalf("expected board size %d, got %d", 8, game.BoardSize)
	}

	if len(game.Pieces) != 0 {
		t.Fatalf("expected no pieces in placeholder scaffold, got %d", len(game.Pieces))
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
