# Workflows

## Import → Categorize → Budget → Report

1. Import transactions (CSV/OFX)
2. Uncategorized transactions are stored
3. Categorization rules are applied
4. Budgets track category spending
5. Reports summarize results

---

## Deduplication

- CSV: optional `external_id`
- OFX: `FITID` is used automatically
- Duplicate imports are ignored silently

---

## Budget Alerts

- Budgets are evaluated dynamically
- Status:
  - OK
  - WARN (threshold configurable)
  - OVER
