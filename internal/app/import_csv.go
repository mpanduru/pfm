package app

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"example.com/pfm/internal/db"
)

type ImportResult struct {
	Seen     int
	Inserted int
	Ignored  int
}

func ImportCSV(conn *sql.DB, path string, account string, source string) (ImportResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return ImportResult{}, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return ImportResult{}, fmt.Errorf("read header: %w", err)
	}

	col := map[string]int{}
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}

	need := []string{"date", "payee", "amount"}
	for _, n := range need {
		if _, ok := col[n]; !ok {
			return ImportResult{}, fmt.Errorf("missing required column %q", n)
		}
	}

	get := func(row []string, name string) string {
		i, ok := col[name]
		if !ok || i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	var res ImportResult

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return res, fmt.Errorf("read row: %w", err)
		}
		res.Seen++

		dateStr := get(row, "date")
		payee := get(row, "payee")
		amountStr := get(row, "amount")
		category := get(row, "category")
		memo := get(row, "memo")
		external := get(row, "external_id")

		if category == "" {
			category = "uncategorized"
		}
		if memo == "" {
			memo = ""
		}

		postedAt, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return res, fmt.Errorf("row %d: invalid date %q: %w", res.Seen+1, dateStr, err)
		}

		amountBani, err := ParseRON(amountStr)
		if err != nil {
			return res, fmt.Errorf("row %d: invalid amount %q: %w", res.Seen+1, amountStr, err)
		}

		var externalID *string
		if external != "" {
			externalID = &external
		}

		_, inserted, err := db.InsertTransaction(conn, db.AddTxParams{
			PostedAt:   postedAt,
			Payee:      payee,
			Memo:       memo,
			AmountBani: amountBani,
			Category:   category,
			Account:    account,
			Source:     source,
			ExternalID: externalID,
		})
		if err != nil {
			return res, fmt.Errorf("row %d: insert: %w", res.Seen+1, err)
		}

		if inserted {
			res.Inserted++
		} else {
			res.Ignored++
		}
	}

	return res, nil
}
