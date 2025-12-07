package db

import (
	"database/sql"
	"fmt"
	"time"
)

type AddTxParams struct {
	PostedAt    time.Time
	Payee       string
	Memo        string
	AmountBani  int64
	Category    string
	Account     string
	Source      string
	ExternalID  *string
}

func InsertTransaction(conn *sql.DB, p AddTxParams) (int64, bool, error) {
	var (
		res sql.Result
		err error
	)

	if p.ExternalID != nil {
		res, err = conn.Exec(`
			INSERT OR IGNORE INTO transactions
			(posted_at, payee, memo, amount_bani, category, account, source, external_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			p.PostedAt.Format("2006-01-02"),
			p.Payee,
			p.Memo,
			p.AmountBani,
			p.Category,
			p.Account,
			p.Source,
			p.ExternalID,
		)
	} else {
		res, err = conn.Exec(`
			INSERT INTO transactions
			(posted_at, payee, memo, amount_bani, category, account, source, external_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			p.PostedAt.Format("2006-01-02"),
			p.Payee,
			p.Memo,
			p.AmountBani,
			p.Category,
			p.Account,
			p.Source,
			p.ExternalID,
		)
	}

	if err != nil {
		return 0, false, fmt.Errorf("insert transaction: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, false, fmt.Errorf("last insert id: %w", err)
	}

	inserted := id != 0
	return id, inserted, nil
}
