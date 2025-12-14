PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS transactions (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  posted_at     TEXT NOT NULL,
  payee         TEXT NOT NULL,
  memo          TEXT NOT NULL DEFAULT '',
  amount_bani   INTEGER NOT NULL,
  category      TEXT NOT NULL DEFAULT 'uncategorized',
  account       TEXT NOT NULL DEFAULT 'default',
  source        TEXT NOT NULL DEFAULT 'manual',
  external_id   TEXT,
  created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_transactions_dedupe
ON transactions(account, source, external_id)
WHERE external_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_transactions_posted_at ON transactions(posted_at);
CREATE INDEX IF NOT EXISTS ix_transactions_category ON transactions(category);

CREATE TABLE IF NOT EXISTS category_rules (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  name      TEXT NOT NULL,
  pattern   TEXT NOT NULL,
  category  TEXT NOT NULL,
  priority  INTEGER NOT NULL DEFAULT 100
);

CREATE TABLE IF NOT EXISTS budgets (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  month           TEXT NOT NULL,
  category        TEXT NOT NULL,
  limit_bani      INTEGER NOT NULL,
  created_at      TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(month, category)
);
