package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type TxRow struct {
	ID         int64
	PostedAt   time.Time
	Payee      string
	Memo       string
	AmountBani int64
	Category   string
	Account    string
	Source     string
}

type ListFilter struct {
	Month    string // "YYYY-MM"
	From     *time.Time
	To       *time.Time
	Category string
	Text     string
	Limit    int
}

func ListTransactions(conn *sql.DB, f ListFilter) ([]TxRow, error) {
	where := make([]string, 0, 6)
	args := make([]any, 0, 6)

	// Month filter
	if f.Month != "" {
		where = append(where, "posted_at LIKE ?")
		args = append(args, f.Month+"-%")
	}

	if f.From != nil {
		where = append(where, "posted_at >= ?")
		args = append(args, f.From.Format("2006-01-02"))
	}
	if f.To != nil {
		where = append(where, "posted_at <= ?")
		args = append(args, f.To.Format("2006-01-02"))
	}

	if f.Category != "" {
		where = append(where, "category = ?")
		args = append(args, f.Category)
	}

	if f.Text != "" {
		// Case-insensitive match
		where = append(where, "(LOWER(payee) LIKE ? OR LOWER(memo) LIKE ?)")
		p := "%" + strings.ToLower(f.Text) + "%"
		args = append(args, p, p)
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 200
	}

	query := `
		SELECT id, posted_at, payee, memo, amount_bani, category, account, source
		FROM transactions
	`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY posted_at DESC, id DESC"
	query += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	out := []TxRow{}
	for rows.Next() {
		var (
			id         int64
			postedAtS  string
			payee      string
			memo       string
			amountBani int64
			category   string
			account    string
			source     string
		)
		if err := rows.Scan(&id, &postedAtS, &payee, &memo, &amountBani, &category, &account, &source); err != nil {
			return nil, err
		}
		postedAt, err := time.Parse("2006-01-02", postedAtS)
		if err != nil {
			return nil, fmt.Errorf("bad posted_at in db: %q: %w", postedAtS, err)
		}
		out = append(out, TxRow{
			ID:         id,
			PostedAt:   postedAt,
			Payee:      payee,
			Memo:       memo,
			AmountBani: amountBani,
			Category:   category,
			Account:    account,
			Source:     source,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
