package db

import (
	"database/sql"
	"fmt"
)

type RuleRow struct {
	ID       int64
	Name     string
	Pattern  string
	Category string
	Priority int64
}

func AddRule(conn *sql.DB, name, pattern, category string, priority int64) (int64, error) {
	res, err := conn.Exec(`
		INSERT INTO category_rules (name, pattern, category, priority)
		VALUES (?, ?, ?, ?)
	`, name, pattern, category, priority)
	if err != nil {
		return 0, fmt.Errorf("add rule: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("rule id: %w", err)
	}
	return id, nil
}

func ListRules(conn *sql.DB) ([]RuleRow, error) {
	rows, err := conn.Query(`
		SELECT id, name, pattern, category, priority
		FROM category_rules
		ORDER BY priority ASC, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}
	defer rows.Close()

	var out []RuleRow
	for rows.Next() {
		var r RuleRow
		if err := rows.Scan(&r.ID, &r.Name, &r.Pattern, &r.Category, &r.Priority); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type TxForCategorize struct {
	ID       int64
	Payee    string
	Memo     string
	Category string
	PostedAt string // YYYY-MM-DD
}

func ListTxForCategorize(conn *sql.DB, month string, all bool) ([]TxForCategorize, error) {
	where := "category = 'uncategorized'"
	args := []any{}

	if month != "" {
		where += " AND posted_at LIKE ?"
		args = append(args, month+"-%")
	}
	if all {
	}

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT id, payee, memo, category, posted_at
		FROM transactions
		WHERE %s
		ORDER BY posted_at DESC, id DESC
	`, where), args...)
	if err != nil {
		return nil, fmt.Errorf("list tx for categorize: %w", err)
	}
	defer rows.Close()

	var out []TxForCategorize
	for rows.Next() {
		var t TxForCategorize
		if err := rows.Scan(&t.ID, &t.Payee, &t.Memo, &t.Category, &t.PostedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func UpdateTxCategory(conn *sql.DB, id int64, newCategory string) error {
	_, err := conn.Exec(`UPDATE transactions SET category = ? WHERE id = ?`, newCategory, id)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	return nil
}
