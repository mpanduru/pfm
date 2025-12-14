package db

import (
	"database/sql"
	"fmt"
)

type BudgetRow struct {
	Month     string
	Category  string
	LimitBani int64
}

func UpsertBudget(conn *sql.DB, month, category string, limitBani int64) error {
	_, err := conn.Exec(`
		INSERT INTO budgets (month, category, limit_bani)
		VALUES (?, ?, ?)
		ON CONFLICT(month, category) DO UPDATE SET limit_bani = excluded.limit_bani
	`, month, category, limitBani)
	if err != nil {
		return fmt.Errorf("upsert budget: %w", err)
	}
	return nil
}

func ListBudgetsForMonth(conn *sql.DB, month string) ([]BudgetRow, error) {
	rows, err := conn.Query(`
		SELECT month, category, limit_bani
		FROM budgets
		WHERE month = ?
		ORDER BY category ASC
	`, month)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	defer rows.Close()

	var out []BudgetRow
	for rows.Next() {
		var b BudgetRow
		if err := rows.Scan(&b.Month, &b.Category, &b.LimitBani); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func GetSpentForMonthCategory(conn *sql.DB, month, category string) (int64, error) {
	row := conn.QueryRow(`
		SELECT COALESCE(SUM(amount_bani), 0)
		FROM transactions
		WHERE posted_at LIKE ?
		  AND category = ?
		  AND amount_bani < 0
	`, month+"-%", category)

	var spent int64
	if err := row.Scan(&spent); err != nil {
		return 0, fmt.Errorf("spent query: %w", err)
	}
	return spent, nil
}
