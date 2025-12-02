# Overview
PFM (personal finance manager) is a command-line tool for tracking personal income and expenses. Import transactions from bank statements, categorize them automatically, set budgets, and generate insightful reports—all from your terminal.

## Why This Project?
CLI tools force you to think about user experience in a constrained environment. You'll work with file formats, data persistence, and creating a pleasant terminal interface while building something you can actually use daily.

## User Stories
• As a user, I can import transactions from CSV/OFX files
• As a user, I can manually add income and expenses
• As a user, I can categorize transactions automatically
• As a user, I can set budgets per category and get alerts
• As a user, I can generate reports (monthly spending, category breakdown)
• As a user, I can search and filter transactions

## Technical Details
• CLI with subcommands (import, add, report, budget, search)
• Parse CSV and OFX formats
• SQLite database for local storage
• Transaction categorization with rules (regex matching)
• Budget tracking with alerts
• Report generation with charts in terminal
• Interactive TUI (Terminal UI) for browsing

# Setup
## Clone the repository
```bash
git clone https://github.com/mpanduru/pfm.git
cd pfm
```

## Build and run the application
```bash
go build -o pfm ./cmd
./pfm help
```