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

	// Placeholders for upcoming stages:
	case "import":
		return errors.New("import: not implemented yet")
	case "add":
		return a.cmdAdd(args[1:])
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
  report          Generate reports (later)
  budget          Set/check budgets (later)
  search          Search/filter transactions (later)

Data:
  Database file defaults to: %s

Examples:
  %s init
  %s version
  %s add --date 2025-10-22 --payee "Lidl" --amount -23.45 --category groceries --memo "weekly"
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

	id, err := db.InsertTransaction(conn, db.AddTxParams{
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

	fmt.Printf("Added transaction #%d: %s | %s | %s | %s\n",
		id,
		postedAt.Format("2006-01-02"),
		*payee,
		FormatRON(amountBani),
		*category,
	)
	return nil
}
