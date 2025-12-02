package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
)

const Version = "0.1.0"

type App struct {
	DBPath string
}

func New() *App {
	return &App{
		DBPath: "pfm.db",
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

func (a *App) printHelp() {
	exe := "pfm"
	fmt.Printf(`%s - Personal Finance CLI Manager

Usage:
  %s <command> [options]

Commands:
  help            Show this help
  version         Show version
  import          Import transactions from CSV/OFX (next)
  add             Add a transaction manually (later)
  report          Generate reports (later)
  budget          Set/check budgets (later)
  search          Search/filter transactions (later)

Data:
  Database file defaults to: %s

Examples:
  %s version
  %s help
`, exe, exe, filepath.Clean("pfm.db"), exe, exe)
}
