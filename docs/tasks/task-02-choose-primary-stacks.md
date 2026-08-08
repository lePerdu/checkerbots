# Task: Choose primary server and UI stacks

## Context

The project needs a documented decision for the first implementation stack so subsequent setup tasks can proceed consistently.

Relevant docs:
- `docs/phased-implementation-backlog.md`
- `docs/epics-and-issues.md`

## Goal

Write a short decision document choosing the primary stack for the coordination server and web UI, with brief tradeoffs.

## In scope

- Choose the web UI stack
- Choose the coordination server stack
- Document rationale and tradeoffs
- State any assumptions about simulation and camera-tracking compatibility

## Out of scope

- Scaffolding the actual applications
- Installing dependencies
- Exhaustive research into every possible framework

## Expected files or areas

- `docs/`
- Suggested new file: `docs/stack-decisions.md`

## Implementation notes

- Optimize for fast iteration and MVP delivery.
- Keep the decision practical and lightweight.
- The document should be short but concrete.
- It is fine to note that camera tracking may later use a different language/runtime if needed.

## Acceptance criteria

- A document exists describing the chosen server stack.
- A document exists describing the chosen UI stack.
- The document briefly explains tradeoffs and why the choice fits the MVP.

## Validation

- Proofread the document for clarity and consistency with existing planning docs.

## Deliverables

- Stack decision document

## Notes for the final response

Summarize the chosen stacks and any major tradeoffs or future caveats.
