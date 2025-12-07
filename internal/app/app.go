package app

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"example.com/pfm/internal/db"
)

const Version = "0.1.0"

type App struct {
	DBPath     string
	SchemaPath string
}

func New() *App {
	return &App{
		DBPath:     "pfm.db",
		SchemaPath: filepath.Join("internal", "db", "schema.sql"),
	}
}

func trunc(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	return s[:n]
}


func (a *App) Run(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		a.printHelp()
		return nil
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Printf("pfm %s (%s)\n", Version, runtime.Version())
		return nil

	case "init":
		return a.cmdInit(args[1:])
	case "import":
		return a.cmdImport(args[1:])
	case "add":
		return a.cmdAdd(args[1:])
	case "list":
		return a.cmdList(args[1:])
	case "report":
		return errors.New("report: not implemented yet")
	case "budget":
		return errors.New("budget: not implemented yet")
	case "search":
		return errors.New("search: not implemented yet")

	default:
		return fmt.Errorf("unknown command: %q (try: pfm help)", args[0])
	}
}

func (a *App) cmdInit(args []string) error {
	conn, err := db.Open(a.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := db.Migrate(conn, a.SchemaPath); err != nil {
		return err
	}

	fmt.Printf("Initialized database: %s\n", filepath.Clean(a.DBPath))
	return nil
}

func (a *App) printHelp() {
	exe := "pfm"
	fmt.Printf(`%s - Personal Finance CLI Manager

Usage:
  %s <command> [options]

Commands:
  help            Show this help
  version         Show version
  init            Create database + tables
  import          Import transactions from CSV/OFX (next)
  add             Add a transaction manually (later)
  list			  List transactions
  report          Generate reports (later)
  budget          Set/check budgets (later)
  search          Search/filter transactions (later)

Data:
  Database file defaults to: %s

Examples:
  %s init
  %s version
  %s add --date 2025-10-22 --payee "Lidl" --amount -23.45 --category groceries --memo "weekly"
  %s list --month 2026-01
  %s import --file sample.csv --account default
`, exe, exe, filepath.Clean(a.DBPath), exe, exe)
}

func (a *App) cmdAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)

	dateStr := fs.String("date", "", "Transaction date (YYYY-MM-DD) [required]")
	payee := fs.String("payee", "", "Payee/merchant [required]")
	amountStr := fs.String("amount", "", "Amount in RON (e.g. -12.34) [required]")
	category := fs.String("category", "uncategorized", "Category")
	memo := fs.String("memo", "", "Memo/notes")
	account := fs.String("account", "default", "Account name")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *dateStr == "" || *payee == "" || *amountStr == "" {
		return errors.New("missing required flags: --date, --payee, --amount")
	}

	postedAt, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		return fmt.Errorf("invalid --date (expected YYYY-MM-DD): %w", err)
	}

	amountBani, err := ParseRON(*amountStr)
	if err != nil {
		return fmt.Errorf("invalid --amount: %w", err)
	}

	conn, err := db.Open(a.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := db.Migrate(conn, a.SchemaPath); err != nil {
		return err
	}

	id, inserted, err := db.InsertTransaction(conn, db.AddTxParams{
		PostedAt:   postedAt,
		Payee:      *payee,
		Memo:       *memo,
		AmountBani: amountBani,
		Category:   *category,
		Account:    *account,
		Source:     "manual",
		ExternalID: nil,
	})
	if err != nil {
		return err
	}
	if !inserted {
		fmt.Println("Transaction was ignored (duplicate).")
		return nil
	}

	fmt.Printf("Added transaction #%d: %s | %s | %s | %s\n",
		id,
		postedAt.Format("2006-01-02"),
		*payee,
		FormatRON(amountBani),
		*category,
	)
	return nil
}

func (a *App) cmdList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)

	month := fs.String("month", "", "Filter by month (YYYY-MM)")
	fromStr := fs.String("from", "", "Start date (YYYY-MM-DD)")
	toStr := fs.String("to", "", "End date (YYYY-MM-DD)")
	category := fs.String("category", "", "Filter by category")
	text := fs.String("text", "", "Search text in payee/memo (case-insensitive)")
	limit := fs.Int("limit", 200, "Max rows to show")

	if err := fs.Parse(args); err != nil {
		return err
	}

	var from *time.Time
	var to *time.Time

	if *fromStr != "" {
		t, err := time.Parse("2006-01-02", *fromStr)
		if err != nil {
			return fmt.Errorf("invalid --from (expected YYYY-MM-DD): %w", err)
		}
		from = &t
	}

	if *toStr != "" {
		t, err := time.Parse("2006-01-02", *toStr)
		if err != nil {
			return fmt.Errorf("invalid --to (expected YYYY-MM-DD): %w", err)
		}
		to = &t
	}

	conn, err := db.Open(a.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := db.Migrate(conn, a.SchemaPath); err != nil {
		return err
	}

	rows, err := db.ListTransactions(conn, db.ListFilter{
		Month:    *month,
		From:     from,
		To:       to,
		Category: *category,
		Text:     *text,
		Limit:    *limit,
	})
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		fmt.Println("No transactions found.")
		return nil
	}

	fmt.Printf("%-5s  %-10s  %-10s  %-18s  %-12s  %s\n", "ID", "DATE", "ACCOUNT", "PAYEE", "AMOUNT", "CATEGORY")
	fmt.Printf("%s\n", "-----  ----------  ----------  ------------------  ------------  --------")

	var total int64
	for _, r := range rows {
		total += r.AmountBani
		payee := r.Payee
		if len(payee) > 18 {
			payee = payee[:18]
		}
		fmt.Printf("%-5d  %-10s  %-10s  %-18s  %-12s  %s\n",
			r.ID,
			r.PostedAt.Format("2006-01-02"),
			trunc(r.Account, 10),
			trunc(payee, 18),
			FormatRON(r.AmountBani),
			r.Category,
		)
	}

	fmt.Printf("\nShown: %d   Net total: %s\n", len(rows), FormatRON(total))
	return nil
}

func (a *App) cmdImport(args []string) error {
	fs := flag.NewFlagSet("import", flag.ContinueOnError)

	file := fs.String("file", "", "CSV file path [required]")
	account := fs.String("account", "default", "Account name")
	source := fs.String("source", "csv", "Source label (default: csv)")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *file == "" {
		return errors.New("missing required flag: --file")
	}

	conn, err := db.Open(a.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := db.Migrate(conn, a.SchemaPath); err != nil {
		return err
	}

	result, err := ImportCSV(conn, *file, *account, *source)
	if err != nil {
		return err
	}

	fmt.Printf("Import complete: seen=%d inserted=%d ignored=%d\n", result.Seen, result.Inserted, result.Ignored)
	return nil
}
