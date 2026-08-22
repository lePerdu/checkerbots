package gameengine

import "testing"

func TestNewGameReturnsStandardStartingBoard(t *testing.T) {
	game := NewGame8x8()

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

	if len(game.LegalMoves) != 7 {
		t.Fatalf("expected 7 opening legal moves, got %d", len(game.LegalMoves))
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
	game := NewGame8x8()
	moves := game.LegalMoves

	expected := []Move{
		{{Row: 2, Col: 0}, {Row: 3, Col: 1}},
		{{Row: 2, Col: 2}, {Row: 3, Col: 1}},
		{{Row: 2, Col: 2}, {Row: 3, Col: 3}},
		{{Row: 2, Col: 4}, {Row: 3, Col: 3}},
		{{Row: 2, Col: 4}, {Row: 3, Col: 5}},
		{{Row: 2, Col: 6}, {Row: 3, Col: 5}},
		{{Row: 2, Col: 6}, {Row: 3, Col: 7}},
	}

	assertMovesEqual(t, moves, expected)
}

func TestGetLegalMovesReturnsSimpleMovesForRedTurn(t *testing.T) {
	game := NewGame8x8()
	game.Turn = PlayerSideRed
	computeLegalMoves(&game)

	expected := []Move{
		{{Row: 5, Col: 1}, {Row: 4, Col: 0}},
		{{Row: 5, Col: 1}, {Row: 4, Col: 2}},
		{{Row: 5, Col: 3}, {Row: 4, Col: 2}},
		{{Row: 5, Col: 3}, {Row: 4, Col: 4}},
		{{Row: 5, Col: 5}, {Row: 4, Col: 4}},
		{{Row: 5, Col: 5}, {Row: 4, Col: 6}},
		{{Row: 5, Col: 7}, {Row: 4, Col: 6}},
	}

	assertMovesEqual(t, game.LegalMoves, expected)
}

func TestGetLegalMovesSkipsBlockedAndCapturedPieces(t *testing.T) {
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-1", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 2, Col: 2}},
			{ID: "black-2", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 4, Col: 4}, Captured: true},
			{ID: "red-1", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 3, Col: 1}},
			{ID: "red-2", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 3, Col: 3}},
		},
		MoveHistory: []Move{},
	})
	moves := game.LegalMoves
	expected := []Move{
		{{Row: 2, Col: 2}, {Row: 4, Col: 0}},
		{{Row: 2, Col: 2}, {Row: 4, Col: 4}},
	}

	assertMovesEqual(t, moves, expected)
}

func TestGetLegalMovesIncludesBackwardMovesForKings(t *testing.T) {
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-king-1", Side: PlayerSideBlack, Kind: PieceKindKing, Position: Position{Row: 3, Col: 3}},
		},
		MoveHistory: []Move{},
	})
	moves := game.LegalMoves

	expected := []Move{
		{{Row: 3, Col: 3}, {Row: 2, Col: 2}},
		{{Row: 3, Col: 3}, {Row: 2, Col: 4}},
		{{Row: 3, Col: 3}, {Row: 4, Col: 2}},
		{{Row: 3, Col: 3}, {Row: 4, Col: 4}},
	}

	assertMovesEqual(t, moves, expected)
}

func TestApplyMoveMovesPieceAndAdvancesTurn(t *testing.T) {
	game := NewGame8x8()
	move := Move{{Row: 2, Col: 0}, {Row: 3, Col: 1}}

	err := ApplyMove(&game, move)
	if err != nil {
		t.Fatalf("expected move to be applied, got error: %v", err)
	}

	if game.Turn != PlayerSideRed {
		t.Fatalf("expected turn %q after move, got %q", PlayerSideRed, game.Turn)
	}

	if len(game.MoveHistory) != 1 {
		t.Fatalf("expected 1 move in history, got %d", len(game.MoveHistory))
	}

	if len(game.LegalMoves) == 0 {
		t.Fatalf("expected legal moves to be recomputed after move")
	}

	assertMoveEqual(t, game.MoveHistory[0], move)

	foundDestination := false
	foundOrigin := false
	for _, piece := range game.Pieces {
		if piece.Captured {
			continue
		}
		if piece.Position == move[1] && piece.Side == PlayerSideBlack {
			foundDestination = true
		}
		if piece.Position == move[0] {
			foundOrigin = true
		}
	}

	if !foundDestination {
		t.Fatalf("expected moved black piece at %+v", move[1])
	}

	if foundOrigin {
		t.Fatalf("expected no active piece to remain at %+v after move", move[0])
	}
}

func TestApplyMoveCapturesOpponentPiece(t *testing.T) {
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-1", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 2, Col: 0}},
			{ID: "red-1", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 3, Col: 1}},
		},
		MoveHistory: []Move{},
	})
	move := Move{{Row: 2, Col: 0}, {Row: 4, Col: 2}}

	err := ApplyMove(&game, move)
	if err != nil {
		t.Fatalf("expected capture move to be applied, got error: %v", err)
	}

	if game.Turn != PlayerSideRed {
		t.Fatalf("expected turn %q after capture, got %q", PlayerSideRed, game.Turn)
	}

	if len(game.MoveHistory) != 1 {
		t.Fatalf("expected capture move to be appended to history, got %+v", game.MoveHistory)
	}
	assertMoveEqual(t, game.MoveHistory[0], move)

	for _, piece := range game.Pieces {
		switch piece.ID {
		case "black-1":
			if piece.Position != move[1] {
				t.Fatalf("expected capturing piece at %+v, got %+v", move[1], piece.Position)
			}
		case "red-1":
			if !piece.Captured {
				t.Fatalf("expected jumped piece to be marked captured")
			}
		}
	}
}

func TestGetLegalMovesPrefersJumpsOverSimpleMoves(t *testing.T) {
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-jumper", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 2, Col: 2}},
			{ID: "black-simple", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 2, Col: 6}},
			{ID: "red-1", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 3, Col: 3}},
		},
		MoveHistory: []Move{},
	})
	expected := []Move{
		{{Row: 2, Col: 2}, {Row: 4, Col: 4}},
	}

	assertMovesEqual(t, game.LegalMoves, expected)
}

func TestApplyMoveRejectsJumpWithoutPieceToCapture(t *testing.T) {
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-1", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 2, Col: 0}},
		},
		MoveHistory: []Move{},
	})
	move := Move{{Row: 2, Col: 0}, {Row: 4, Col: 2}}

	err := ApplyMove(&game, move)
	if err == nil {
		t.Fatalf("expected jump without captured piece to be rejected")
	}
	if err.Reason != "jump requires a piece to capture" {
		t.Fatalf("expected reason %q, got %q", "jump requires a piece to capture", err.Reason)
	}
}

func TestApplyMoveRejectsJumpOverOwnPiece(t *testing.T) {
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-1", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 2, Col: 0}},
			{ID: "black-2", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 3, Col: 1}},
		},
		MoveHistory: []Move{},
	})
	move := Move{{Row: 2, Col: 0}, {Row: 4, Col: 2}}

	err := ApplyMove(&game, move)
	if err == nil {
		t.Fatalf("expected jump over own piece to be rejected")
	}
	if err.Reason != "cannot capture your own piece" {
		t.Fatalf("expected reason %q, got %q", "cannot capture your own piece", err.Reason)
	}
}

func TestApplyMoveRejectsInvalidMove(t *testing.T) {
	game := NewGame8x8()
	move := Move{{Row: 0, Col: 0}, {Row: 1, Col: 1}}

	err := ApplyMove(&game, move)
	if err == nil {
		t.Fatalf("expected invalid move to be rejected")
	}
	if err.Reason != "destination is occupied" {
		t.Fatalf("expected reason %q, got %q", "destination is occupied", err.Reason)
	}

	if len(game.MoveHistory) != 0 {
		t.Fatalf("expected invalid move to leave history unchanged, got %+v", game.MoveHistory)
	}

	if game.Turn != PlayerSideBlack {
		t.Fatalf("expected invalid move to leave turn unchanged, got %q", game.Turn)
	}

	for _, piece := range game.Pieces {
		if piece.ID == "black-1" && piece.Position != (Position{Row: 0, Col: 0}) {
			t.Fatalf("expected invalid move to leave piece in place, got %+v", piece.Position)
		}
	}
}

func TestGetLegalMovesFindsMultiStepJumpSequence(t *testing.T) {
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-1", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 0, Col: 2}},
			{ID: "red-1", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 1, Col: 3}},
			{ID: "red-2", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 3, Col: 3}},
		},
		MoveHistory: []Move{},
	})
	expected := []Move{
		{{Row: 0, Col: 2}, {Row: 2, Col: 4}, {Row: 4, Col: 2}},
	}

	assertMovesEqual(t, game.LegalMoves, expected)
}

func TestGetLegalMovesKingCannotJumpSamePieceTwice(t *testing.T) {
	// The king could otherwise hop back and forth over the single red piece
	// forever; it must only be allowed to capture it once.
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-king", Side: PlayerSideBlack, Kind: PieceKindKing, Position: Position{Row: 2, Col: 2}},
			{ID: "red-1", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 3, Col: 3}},
		},
		MoveHistory: []Move{},
	})
	expected := []Move{
		{{Row: 2, Col: 2}, {Row: 4, Col: 4}},
	}

	assertMovesEqual(t, game.LegalMoves, expected)
}

func TestGetLegalMovesJumpSequenceEndsWhenPieceIsPromoted(t *testing.T) {
	// black-1 can jump to row 7 (its promotion row) and would have another
	// jump available from there, but becoming a king must end the sequence.
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-1", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 5, Col: 3}},
			{ID: "red-1", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 6, Col: 4}},
			{ID: "red-2", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 6, Col: 6}},
		},
		MoveHistory: []Move{},
	})
	expected := []Move{
		{{Row: 5, Col: 3}, {Row: 7, Col: 5}},
	}

	assertMovesEqual(t, game.LegalMoves, expected)
}

func TestGetLegalMovesKingContinuesJumpingPastPromotionRow(t *testing.T) {
	// black-king is already a king, so landing on row 7 (black's promotion
	// row) doesn't stop it from continuing the jump sequence.
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-king", Side: PlayerSideBlack, Kind: PieceKindKing, Position: Position{Row: 5, Col: 3}},
			{ID: "red-1", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 6, Col: 4}},
			{ID: "red-2", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 6, Col: 6}},
		},
		MoveHistory: []Move{},
	})
	expected := []Move{
		{{Row: 5, Col: 3}, {Row: 7, Col: 5}, {Row: 5, Col: 7}},
	}

	assertMovesEqual(t, game.LegalMoves, expected)
}

func TestGetLegalMovesAllowsDifferentLengthJumpSequencesSimultaneously(t *testing.T) {
	// black-1 has only a single jump available, while black-2 has a
	// two-step jump sequence available; both should be legal moves.
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-1", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 0, Col: 0}},
			{ID: "red-1", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 1, Col: 1}},

			{ID: "black-2", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 0, Col: 4}},
			{ID: "red-2", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 1, Col: 5}},
			{ID: "red-3", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 3, Col: 5}},
		},
		MoveHistory: []Move{},
	})
	expected := []Move{
		{{Row: 0, Col: 0}, {Row: 2, Col: 2}},
		{{Row: 0, Col: 4}, {Row: 2, Col: 6}, {Row: 4, Col: 4}},
	}

	assertMovesEqual(t, game.LegalMoves, expected)
}

func TestApplyMoveAppliesMultiStepJumpSequence(t *testing.T) {
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-1", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 0, Col: 2}},
			{ID: "red-1", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 1, Col: 3}},
			{ID: "red-2", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 3, Col: 3}},
		},
		MoveHistory: []Move{},
	})
	move := Move{{Row: 0, Col: 2}, {Row: 2, Col: 4}, {Row: 4, Col: 2}}

	err := ApplyMove(&game, move)
	if err != nil {
		t.Fatalf("expected multi-step jump to be applied, got error: %v", err)
	}

	if game.Turn != PlayerSideRed {
		t.Fatalf("expected turn %q after multi-step jump, got %q", PlayerSideRed, game.Turn)
	}

	assertMoveEqual(t, game.MoveHistory[0], move)

	for _, piece := range game.Pieces {
		switch piece.ID {
		case "black-1":
			if piece.Position != (Position{Row: 4, Col: 2}) {
				t.Fatalf("expected jumping piece at %+v, got %+v", Position{Row: 4, Col: 2}, piece.Position)
			}
		case "red-1", "red-2":
			if !piece.Captured {
				t.Fatalf("expected piece %q to be captured", piece.ID)
			}
		}
	}
}

func TestApplyMoveRejectsMultiStepMoveNotInLegalMoves(t *testing.T) {
	game := GameFromStored(StoredGame{
		Turn:      PlayerSideBlack,
		BoardSize: 8,
		Pieces: []Piece{
			{ID: "black-1", Side: PlayerSideBlack, Kind: PieceKindMan, Position: Position{Row: 0, Col: 2}},
			{ID: "red-1", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 1, Col: 3}},
			{ID: "red-2", Side: PlayerSideRed, Kind: PieceKindMan, Position: Position{Row: 3, Col: 3}},
		},
		MoveHistory: []Move{},
	})
	// Only one destination is reachable after landing at (2,4); (4,6) is not.
	move := Move{{Row: 0, Col: 2}, {Row: 2, Col: 4}, {Row: 4, Col: 6}}

	err := ApplyMove(&game, move)
	if err == nil {
		t.Fatalf("expected illegal multi-step move to be rejected")
	}
	if err.Reason != "move is not a legal move" {
		t.Fatalf("expected reason %q, got %q", "move is not a legal move", err.Reason)
	}
}

func assertMovesEqual(t *testing.T, got []Move, want []Move) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d moves, got %d: %+v", len(want), len(got), got)
	}

	used := make([]bool, len(got))
	for _, expectedMove := range want {
		found := false
		for i, actualMove := range got {
			if used[i] {
				continue
			}
			if movesEqual(actualMove, expectedMove) {
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected move not returned: %+v; got %+v", expectedMove, got)
		}
	}
}

func assertMoveEqual(t *testing.T, got Move, want Move) {
	t.Helper()
	if !movesEqual(got, want) {
		t.Fatalf("expected move %+v, got %+v", want, got)
	}
}

func movesEqual(a Move, b Move) bool {
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
