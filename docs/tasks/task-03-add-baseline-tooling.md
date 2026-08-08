# Task: Add baseline developer tooling

## Context

The repo should have a minimal baseline for formatting, linting, and ignore files before major implementation begins.

Relevant docs:
- `docs/epics-and-issues.md`
- `docs/stack-decisions.md`

## Goal

Set up a lightweight baseline tooling configuration appropriate for the chosen primary stack(s).

## In scope

- Add Dev Container config compatible with the Zed editor
- Add formatter configuration
- Add lint configuration
- Add relevant ignore files
- Add minimal scripts or commands if a root project manifest is introduced

## Out of scope

- Full CI setup
- Extensive pre-commit automation
- Complex multi-language tooling unless already justified by stack decisions

## Expected files or areas

Depends on chosen stack, but likely:
- `package.json`
- `.gitignore`
- `.editorconfig`
- formatter/linter config files

## Implementation notes

- Keep the setup simple and consistent with the stack decision document.
- Prefer commonly used defaults over custom rules.
- Avoid over-engineering the root tooling.

## Acceptance criteria

- Formatting configuration exists.
- Lint configuration exists.
- Ignore files exist.
- There is a documented way to run the tooling locally.

## Validation

- Run the formatter/linter command if possible.
- If the project is not fully scaffolded yet, validate config file structure and document the intended commands.

## Deliverables

- Baseline project tooling configuration

## Notes for the final response

Summarize the tooling added and how it should be run locally.
