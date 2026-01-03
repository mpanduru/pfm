package app

import (
	"fmt"
	"strings"

	"example.com/pfm/internal/db"

	tea "github.com/charmbracelet/bubbletea"
)

type tuiModel struct {
	rows []db.TxRow

	cursor int
	filter string

	typingFilter bool
	filterInput   string

	showDetails bool
	width       int
	height      int
}

func newTUIModel(rows []db.TxRow) tuiModel {
	return tuiModel{rows: rows}
}

func (m tuiModel) Init() tea.Cmd { return nil }

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		k := msg.String()

		if m.typingFilter {
			switch k {
			case "enter":
				m.filter = strings.TrimSpace(m.filterInput)
				m.typingFilter = false
				m.filterInput = ""
				m.cursor = 0
				return m, nil
			case "esc":
				m.typingFilter = false
				m.filterInput = ""
				return m, nil
			case "backspace":
				if len(m.filterInput) > 0 {
					m.filterInput = m.filterInput[:len(m.filterInput)-1]
				}
				return m, nil
			default:
				if len(k) == 1 {
					m.filterInput += k
				}
				return m, nil
			}
		}

		switch k {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.filtered())-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			m.showDetails = !m.showDetails
			return m, nil
		case "/":
			m.typingFilter = true
			m.filterInput = ""
			return m, nil
		case "esc":
			m.filter = ""
			m.cursor = 0
			return m, nil
		}
	}

	return m, nil
}

func (m tuiModel) View() string {
	header := "pfm tui  |  ↑/↓ move  enter details  / filter  esc clear  q quit\n"
	if m.typingFilter {
		return header + "\nFilter: " + m.filterInput + "█\n"
	}

	rows := m.filtered()
	if len(rows) == 0 {
		return header + "\nNo matching transactions.\n"
	}

	if m.cursor >= len(rows) {
		m.cursor = len(rows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	var b strings.Builder
	b.WriteString(header)
	if m.filter != "" {
		b.WriteString(fmt.Sprintf("Filter: %q\n\n", m.filter))
	} else {
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("%-10s  %-18s  %-12s  %s\n", "DATE", "PAYEE", "AMOUNT", "CATEGORY"))
	b.WriteString("----------  ------------------  ------------  --------\n")

	maxLines := m.height - 8
	if maxLines < 5 {
		maxLines = 5
	}

	start := 0
	if m.cursor >= maxLines {
		start = m.cursor - maxLines + 1
	}
	end := start + maxLines
	if end > len(rows) {
		end = len(rows)
	}

	for i := start; i < end; i++ {
		r := rows[i]
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		payee := trunc(r.Payee, 18)
		b.WriteString(fmt.Sprintf("%s%-10s  %-18s  %-12s  %s\n",
			prefix,
			r.PostedAt.Format("2006-01-02"),
			payee,
			FormatRON(r.AmountBani),
			r.Category,
		))
	}

	if m.showDetails {
		r := rows[m.cursor]
		b.WriteString("\n--- details ---\n")
		b.WriteString(fmt.Sprintf("ID: %d\n", r.ID))
		b.WriteString(fmt.Sprintf("Date: %s\n", r.PostedAt.Format("2006-01-02")))
		b.WriteString(fmt.Sprintf("Payee: %s\n", r.Payee))
		if strings.TrimSpace(r.Memo) != "" {
			b.WriteString(fmt.Sprintf("Memo: %s\n", r.Memo))
		}
		b.WriteString(fmt.Sprintf("Amount: %s\n", FormatRON(r.AmountBani)))
		b.WriteString(fmt.Sprintf("Category: %s\n", r.Category))
		b.WriteString(fmt.Sprintf("Account: %s\n", r.Account))
		b.WriteString(fmt.Sprintf("Source: %s\n", r.Source))
	}

	return b.String()
}

func (m tuiModel) filtered() []db.TxRow {
	if strings.TrimSpace(m.filter) == "" {
		return m.rows
	}
	f := strings.ToLower(strings.TrimSpace(m.filter))
	out := make([]db.TxRow, 0, len(m.rows))
	for _, r := range m.rows {
		hay := strings.ToLower(r.Payee + " " + r.Memo + " " + r.Category + " " + r.Account)
		if strings.Contains(hay, f) {
			out = append(out, r)
		}
	}
	return out
}
