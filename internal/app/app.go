package app

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
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
		return a.cmdReport(args[1:])
	case "budget":
		return a.cmdBudget(args[1:])
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
  %s list --month 2025-10
  %s import --file sample.csv --account default
  %s report categories --month 2025-11
  %s budget set --month 2025-12 --category groceries --limit 200
  %s budget status --month 2025-12
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

	file := fs.String("file", "", "CSV/OFX/QFX file path [required]")
	account := fs.String("account", "default", "Account name")
	source := fs.String("source", "", "Source label (default: csv/ofx based on extension)")

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

	ext := strings.ToLower(filepath.Ext(*file))
	src := *source
	if src == "" {
		if ext == ".csv" {
			src = "csv"
		} else {
			src = "ofx"
		}
	}

	var result ImportResult
	switch ext {
	case ".csv":
		result, err = ImportCSV(conn, *file, *account, src)
	case ".ofx", ".qfx":
		result, err = ImportOFX(conn, *file, *account, src)
	default:
		return fmt.Errorf("unsupported file type: %s (use .csv, .ofx, .qfx)", ext)
	}
	if err != nil {
		return err
	}

	fmt.Printf("Import complete: seen=%d inserted=%d ignored=%d\n", result.Seen, result.Inserted, result.Ignored)
	return nil
}

func (a *App) cmdReport(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Println(`Usage:
  pfm report <subcommand> [options]

Subcommands:
  month        Monthly summary
  categories   Category breakdown (expenses)

Examples:
  pfm report month --month 2026-01
  pfm report categories --month 2026-01
`)
		return nil
	}

	switch args[0] {
	case "month":
		return a.cmdReportMonth(args[1:])
	case "categories":
		return a.cmdReportCategories(args[1:])
	default:
		return fmt.Errorf("unknown report subcommand: %q (try: pfm report help)", args[0])
	}
}

func (a *App) cmdReportMonth(args []string) error {
	fs := flag.NewFlagSet("report month", flag.ContinueOnError)
	month := fs.String("month", "", "Month (YYYY-MM) [required]")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *month == "" {
		return errors.New("missing required flag: --month (YYYY-MM)")
	}

	conn, err := db.Open(a.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := db.Migrate(conn, a.SchemaPath); err != nil {
		return err
	}

	s, err := db.GetMonthSummary(conn, *month)
	if err != nil {
		return err
	}

	expenseAbs := -s.ExpenseBani

	fmt.Printf("Month: %s\n", s.Month)
	fmt.Printf("Transactions: %d\n", s.Count)
	fmt.Printf("Income:   %s\n", FormatRON(s.IncomeBani))
	fmt.Printf("Expenses: %s\n", FormatRON(expenseAbs))
	fmt.Printf("Net:      %s\n", FormatRON(s.NetBani))

	return nil
}

func (a *App) cmdReportCategories(args []string) error {
	fs := flag.NewFlagSet("report categories", flag.ContinueOnError)
	month := fs.String("month", "", "Month (YYYY-MM) [required]")
	all := fs.Bool("all", false, "Include income categories too (default: expenses only)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *month == "" {
		return errors.New("missing required flag: --month (YYYY-MM)")
	}

	conn, err := db.Open(a.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := db.Migrate(conn, a.SchemaPath); err != nil {
		return err
	}

	expensesOnly := !*all
	rows, grand, err := db.GetCategoryTotalsForMonth(conn, *month, expensesOnly)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Println("No matching transactions.")
		return nil
	}

	title := "Category totals (expenses)"
	if *all {
		title = "Category totals (all)"
	}
	fmt.Printf("%s for %s\n\n", title, *month)
	fmt.Printf("%-18s  %-8s  %s\n", "CATEGORY", "COUNT", "TOTAL")
	fmt.Printf("%s\n", "------------------  --------  ------------")

	for _, r := range rows {
		amt := r.TotalBani
		if expensesOnly {
			amt = -amt
		}
		fmt.Printf("%-18s  %-8d  %s\n", trunc(r.Category, 18), r.Count, FormatRON(amt))
	}

	if expensesOnly {
		grand = -grand
	}
	fmt.Printf("\nGrand total: %s\n", FormatRON(grand))
	return nil
}

func (a *App) cmdBudget(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Println(`Usage:
  pfm budget <subcommand> [options]

Subcommands:
  set      Set a budget for a month+category
  status   Show budgets vs spending for a month

Examples:
  pfm budget set --month 2026-01 --category groceries --limit 800
  pfm budget status --month 2026-01
`)
		return nil
	}

	switch args[0] {
	case "set":
		return a.cmdBudgetSet(args[1:])
	case "status":
		return a.cmdBudgetStatus(args[1:])
	default:
		return fmt.Errorf("unknown budget subcommand: %q (try: pfm budget help)", args[0])
	}
}

func (a *App) cmdBudgetSet(args []string) error {
	fs := flag.NewFlagSet("budget set", flag.ContinueOnError)

	month := fs.String("month", "", "Month (YYYY-MM) [required]")
	category := fs.String("category", "", "Category [required]")
	limitStr := fs.String("limit", "", "Limit in RON (e.g. 800 or 800.50) [required]")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *month == "" || *category == "" || *limitStr == "" {
		return errors.New("missing required flags: --month, --category, --limit")
	}

	limitBani, err := ParseRON(*limitStr)
	if err != nil {
		return fmt.Errorf("invalid --limit: %w", err)
	}
	if limitBani < 0 {
		return errors.New("--limit must be positive")
	}

	conn, err := db.Open(a.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := db.Migrate(conn, a.SchemaPath); err != nil {
		return err
	}

	if err := db.UpsertBudget(conn, *month, *category, limitBani); err != nil {
		return err
	}

	fmt.Printf("Budget set: %s %s = %s\n", *month, *category, FormatRON(limitBani))
	return nil
}

func (a *App) cmdBudgetStatus(args []string) error {
	fs := flag.NewFlagSet("budget status", flag.ContinueOnError)

	month := fs.String("month", "", "Month (YYYY-MM) [required]")
	warnPct := fs.Int("warn", 80, "Warn threshold percent (default 80)")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *month == "" {
		return errors.New("missing required flag: --month (YYYY-MM)")
	}
	if *warnPct <= 0 || *warnPct > 1000 {
		return errors.New("--warn must be a reasonable percent (1..1000)")
	}

	conn, err := db.Open(a.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := db.Migrate(conn, a.SchemaPath); err != nil {
		return err
	}

	budgets, err := db.ListBudgetsForMonth(conn, *month)
	if err != nil {
		return err
	}
	if len(budgets) == 0 {
		fmt.Println("No budgets set for that month.")
		return nil
	}

	fmt.Printf("Budget status for %s\n\n", *month)
	fmt.Printf("%-18s  %-12s  %-12s  %-8s  %s\n", "CATEGORY", "LIMIT", "SPENT", "USED", "STATUS")
	fmt.Printf("%s\n", "------------------  ------------  ------------  --------  ------")

	for _, b := range budgets {
		spentNeg, err := db.GetSpentForMonthCategory(conn, b.Month, b.Category)
		if err != nil {
			return err
		}
		spentAbs := -spentNeg

		usedPct := 0
		if b.LimitBani > 0 {
			usedPct = int((spentAbs * 100) / b.LimitBani)
		}

		status := "OK"
		if spentAbs > b.LimitBani {
			status = "OVER"
		} else if usedPct >= *warnPct {
			status = "WARN"
		}

		fmt.Printf("%-18s  %-12s  %-12s  %-7d%%  %s\n",
			trunc(b.Category, 18),
			FormatRON(b.LimitBani),
			FormatRON(spentAbs),
			usedPct,
			status,
		)
	}

	return nil
}
