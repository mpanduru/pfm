# Data Model

## Storage

The application uses a **local SQLite database** (`pfm.db` by default).

There is no server component and no background daemon.

---

## Money Representation (RON)

All monetary values are stored as **integer bani**:

- 1 RON = 100 bani
- Stored as `INTEGER`
- No floats are stored in the database

### Conversion
- Input: `"12.34"` → `1234`
- Output: `1234` → `"12.34 RON"`

Implemented in `internal/app/money.go`.

---

## Tables

### `transactions`

```sql
transactions (
  id            INTEGER PRIMARY KEY,
  posted_at     TEXT,       -- YYYY-MM-DD
  payee         TEXT,
  memo          TEXT,
  amount_bani   INTEGER,
  category      TEXT,
  account       TEXT,
  source        TEXT,
  external_id   TEXT,
  created_at    TEXT
)
```

Notes:
- Expenses are negative
- Income is positive
- external_id is used for import deduplication

### `category_rules`

```sql
category_rules (
  id        INTEGER PRIMARY KEY,
  name      TEXT,
  pattern   TEXT,     -- regex
  category  TEXT,
  priority  INTEGER
)
```

Notes:
- Rules are evaluated in ascending priority order.

### `budgets`

```sql
budgets (
  id          INTEGER PRIMARY KEY,
  month       TEXT,   -- YYYY-MM
  category    TEXT,
  limit_bani  INTEGER
)
```

Notes:
- Budgets are unique per (month, category).