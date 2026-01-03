# Architecture

## Overview

The application is a **single-binary Go CLI** built using:
- Go standard library
- SQLite for local persistence
- Minimal external dependencies (OFX parsing, TUI)

The architecture follows a **layered, dependency-directional design**:

```bash
CLI / TUI
↓
Application Logic (internal/app)
↓
Database Access (internal/db)
↓
SQLite
```

Higher layers depend on lower layers, never the reverse.

## Directory Structure

```bash
├── cmd
│   └── pfm # Entry point
│       └── main.go
├── go.mod
├── go.sum
├── internal
│   ├── app
│   │   ├── app.go # Command routing
│   │   ├── categorize.go
│   │   ├── import_csv.go
│   │   ├── import_ofx.go
│   │   ├── money.go # RON to bani parsing/formatting
│   │   └── tui.go
│   └── db
│       ├── budgets.go
│       ├── db.go # SQLite connection
│       ├── list.go
│       ├── migrate.go
│       ├── reports.go
│       ├── rules.go
│       ├── schema.sql # Database schema
│       ├── search.go
│       └── transactions.go
...
```

## Responsibilities

### `cmd/pfm`
- Argument handoff to the application
- Process exit handling
- No business logic

### `internal/app`
- CLI command parsing (`flag.FlagSet`)
- Validation and orchestration
- Converting user input into DB operations
- Formatting output for terminal
- TUI state management

### `internal/db`
- Raw SQL only
- No business rules
- Returns Go structs, not domain logic

---

## Error Handling Strategy

- Errors propagate upward
- CLI prints errors once, at the top level
- No panics except unrecoverable startup failures
