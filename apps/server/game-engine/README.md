# Game engine

This package is the standalone home for checkers rules logic.

Current scope:
- package-level game types owned by the rules engine
- minimal entry points for `NewGame`, `GetLegalMoves`, `ApplyMove`, and `IsGameOver`
- placeholder behavior so the package is importable and testable before full rules work begins

Future tasks can extend this package with:
- standard starting board initialization
- legal move generation
- capture and multi-jump handling
- promotion and terminal-state detection

The package is intentionally separate from HTTP, realtime transport, and robot orchestration code.
