package app

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"example.com/pfm/internal/db"

	"github.com/aclindsa/ofxgo"
)

func ImportOFX(conn *sql.DB, path string, account string, source string) (ImportResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return ImportResult{}, err
	}
	defer f.Close()

	resp, err := ofxgo.ParseResponse(f)
	if err != nil {
		return ImportResult{}, fmt.Errorf("parse OFX: %w", err)
	}

	checkCurrency := func(cur ofxgo.CurrSymbol) error {
		s := strings.TrimSpace(cur.String())
		if s == "" {
			return nil
		}
		if strings.ToUpper(s) != "RON" {
			return fmt.Errorf("OFX currency is %q, expected RON", s)
		}
		return nil
	}

	var out ImportResult

	for _, msg := range resp.Bank {
		stmt, ok := msg.(*ofxgo.StatementResponse)
		if !ok {
			continue
		}
		if err := checkCurrency(stmt.CurDef); err != nil {
			return out, err
		}

		for _, trn := range stmt.BankTranList.Transactions {
			out.Seen++

			postedAt := trn.DtPosted.Time

			payee := strings.TrimSpace(trn.Name.String())
			memo := strings.TrimSpace(trn.Memo.String())

			fitID := strings.TrimSpace(trn.FiTID.String())
			var externalID *string
			if fitID != "" {
				externalID = &fitID
			}

			amountBani, err := ParseRON(trn.TrnAmt.String())
			if err != nil {
				return out, fmt.Errorf("row %d: invalid amount %q: %w", out.Seen+1, trn.TrnAmt.String(), err)
			}

			_, inserted, err := db.InsertTransaction(conn, db.AddTxParams{
				PostedAt:   postedAt,
				Payee:      payee,
				Memo:       memo,
				AmountBani: amountBani,
				Category:   "uncategorized",
				Account:    account,
				Source:     source,
				ExternalID: externalID,
			})
			if err != nil {
				return out, fmt.Errorf("row %d: insert: %w", out.Seen+1, err)
			}

			if inserted {
				out.Inserted++
			} else {
				out.Ignored++
			}
		}
	}

	for _, msg := range resp.CreditCard {
		stmt, ok := msg.(*ofxgo.CCStatementResponse)
		if !ok {
			continue
		}
		if err := checkCurrency(stmt.CurDef); err != nil {
			return out, err
		}

		for _, trn := range stmt.BankTranList.Transactions {
			out.Seen++

			postedAt := trn.DtPosted.Time

			payee := strings.TrimSpace(trn.Name.String())
			memo := strings.TrimSpace(trn.Memo.String())

			fitID := strings.TrimSpace(trn.FiTID.String())
			var externalID *string
			if fitID != "" {
				externalID = &fitID
			}

			amountBani, err := ParseRON(trn.TrnAmt.String())
			if err != nil {
				return out, fmt.Errorf("row %d: invalid amount %q: %w", out.Seen+1, trn.TrnAmt.String(), err)
			}

			_, inserted, err := db.InsertTransaction(conn, db.AddTxParams{
				PostedAt:   postedAt,
				Payee:      payee,
				Memo:       memo,
				AmountBani: amountBani,
				Category:   "uncategorized",
				Account:    account,
				Source:     source,
				ExternalID: externalID,
			})
			if err != nil {
				return out, fmt.Errorf("row %d: insert: %w", out.Seen+1, err)
			}

			if inserted {
				out.Inserted++
			} else {
				out.Ignored++
			}
		}
	}

	_ = time.UTC

	return out, nil
}
