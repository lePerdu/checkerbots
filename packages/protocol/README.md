# Protocol Package

This package defines the shared robot protocol messages used by the server, simulator, and future robot adapters.

This protocol shouldn't contain game and board-related concepts: it mainly concerns robot setup and motion.

Current contents:
- `protocol.go` - Go structs, constants, and package/type documentation for protocol v0

The protocol package is intentionally focused on wire-level message shapes first. Broader shared domain types can grow here as the rest of the system is implemented.
