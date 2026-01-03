# CLI Commands

## Global Behavior

- All commands use standard flags (`flag` package)
- Database schema is auto-migrated on startup
- Errors are printed to stderr

---

## Core Commands

### `pfm init`
Initialize the database.

### `pfm add`
Add a transaction manually.

Flags:
- `--date YYYY-MM-DD`
- `--payee TEXT`
- `--amount RON`
- `--category TEXT`
- `--memo TEXT`
- `--account TEXT`

---

### `pfm import`
Import CSV or OFX/QFX.

Flags:
- `--file PATH`
- `--account TEXT`
- `--source TEXT`

File type is detected by extension.

---

### `pfm list`
List transactions.

Filters:
- `--month`
- `--category`
- `--text`
- `--from / --to`

---

### `pfm search`
Advanced filtering.

Additional filters:
- `--min`
- `--max`
- `--account`

---

### `pfm report`

Subcommands:
- `month`
- `categories`

---

### `pfm budget`

Subcommands:
- `set`
- `status`

---

### `pfm rule`

Subcommands:
- `add`
- `list`

---

### `pfm categorize`
Apply regex rules to uncategorized transactions.

Flags:
- `--month`
- `--all`
- `--dry-run`