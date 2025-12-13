package db

import (
	"database/sql"
	"fmt"
)

type MonthSummary struct {
	Month        string
	Count        int64
	IncomeBani   int64
	ExpenseBani  int64
	NetBani      int64
}

func GetMonthSummary(conn *sql.DB, month string) (MonthSummary, error) {
	// month: YYYY-MM
	var s MonthSummary
	s.Month = month

	row := conn.QueryRow(`
		SELECT
			COUNT(*) AS cnt,
			COALESCE(SUM(CASE WHEN amount_bani > 0 THEN amount_bani ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN amount_bani < 0 THEN amount_bani ELSE 0 END), 0) AS expense,
			COALESCE(SUM(amount_bani), 0) AS net
		FROM transactions
		WHERE posted_at LIKE ?
	`, month+"-%")

	if err := row.Scan(&s.Count, &s.IncomeBani, &s.ExpenseBani, &s.NetBani); err != nil {
		return MonthSummary{}, fmt.Errorf("month summary: %w", err)
	}
	return s, nil
}

type CategoryTotal struct {
	Category   string
	TotalBani  int64
	Count      int64
}

func GetCategoryTotalsForMonth(conn *sql.DB, month string, expensesOnly bool) ([]CategoryTotal, int64, error) {
	where := "posted_at LIKE ?"
	if expensesOnly {
		where += " AND amount_bani < 0"
	}

	rows, err := conn.Query(fmt.Sprintf(`
		SELECT category, COUNT(*) AS cnt, COALESCE(SUM(amount_bani), 0) AS total
		FROM transactions
		WHERE %s
		GROUP BY category
		ORDER BY total ASC, category ASC
	`, where), month+"-%")
	if err != nil {
		return nil, 0, fmt.Errorf("category totals: %w", err)
	}
	defer rows.Close()

	var out []CategoryTotal
	var grand int64

	for rows.Next() {
		var c CategoryTotal
		if err := rows.Scan(&c.Category, &c.Count, &c.TotalBani); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
		grand += c.TotalBani
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, grand, nil
}
