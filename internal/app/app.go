package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

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
		return errors.New("add: not implemented yet")
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
`, exe, exe, filepath.Clean(a.DBPath), exe, exe)
}
