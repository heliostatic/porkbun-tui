package views

import (
	"fmt"
	"strings"

	"github.com/bc/porkbun-tui/internal/api"
	"github.com/bc/porkbun-tui/internal/keys"
	"github.com/bc/porkbun-tui/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DNSView struct {
	domain  string
	records []api.DNSRecord
	cursor  int
	offset  int
	height  int
	width   int
	loading bool
	err     error
}

func NewDNSView() *DNSView {
	return &DNSView{}
}

func (v *DNSView) SetDomain(domain string) {
	v.domain = domain
	v.records = nil
	v.cursor = 0
	v.offset = 0
	v.loading = true
	v.err = nil
}

func (v *DNSView) SetRecords(records []api.DNSRecord) {
	v.records = records
	v.loading = false
}

func (v *DNSView) SetError(err error) {
	v.err = err
	v.loading = false
}

func (v *DNSView) SetSize(width, height int) {
	v.width = width
	v.height = height - 6
	if v.height < 1 {
		v.height = 1
	}
}

func (v *DNSView) Update(msg tea.Msg) (*DNSView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Up):
			if v.cursor > 0 {
				v.cursor--
				if v.cursor < v.offset {
					v.offset = v.cursor
				}
			}
		case key.Matches(msg, keys.Keys.Down):
			if v.cursor < len(v.records)-1 {
				v.cursor++
				if v.cursor >= v.offset+v.height {
					v.offset = v.cursor - v.height + 1
				}
			}
		}
	}
	return v, nil
}

func (v *DNSView) View() string {
	var b strings.Builder

	// Title
	title := styles.TitleStyle.Render(fmt.Sprintf(" DNS Records: %s ", v.domain))
	b.WriteString(title)
	b.WriteString("\n\n")

	if v.loading {
		b.WriteString("  Loading DNS records...")
		return b.String()
	}

	if v.err != nil {
		b.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("  Error: %v", v.err)))
		if isAPIAccessError(v.err) {
			b.WriteString("\n\n")
			b.WriteString(styles.HelpStyle.Render("  This domain needs API access enabled.\n"))
			b.WriteString(styles.HelpStyle.Render("  Go to porkbun.com → Domain Management → " + v.domain + " → API Access → ON"))
		}
		return b.String()
	}

	if len(v.records) == 0 {
		b.WriteString("  No DNS records found.")
		return b.String()
	}

	// Column widths
	typeWidth := 8
	nameWidth := 30
	contentWidth := 40
	ttlWidth := 8

	// Header
	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s",
		typeWidth, "Type",
		nameWidth, "Name",
		contentWidth, "Content",
		ttlWidth, "TTL",
	)
	b.WriteString(styles.TableHeaderStyle.Render(header))
	b.WriteString("\n")

	// Calculate visible range
	visibleEnd := min(v.offset+v.height, len(v.records))

	// Rows
	for i := v.offset; i < visibleEnd; i++ {
		r := v.records[i]

		recType := r.Type
		name := truncate(r.Name, nameWidth)
		content := truncate(r.Content, contentWidth)
		ttl := r.TTL

		row := fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s",
			typeWidth, recType,
			nameWidth, name,
			contentWidth, content,
			ttlWidth, ttl,
		)

		if i == v.cursor {
			row = styles.TableSelectedStyle.Render(row)
		}

		b.WriteString(row)
		b.WriteString("\n")
	}

	// Selected record details
	if v.cursor < len(v.records) {
		r := v.records[v.cursor]
		b.WriteString("\n")
		b.WriteString(styles.BoxStyle.Render(v.recordDetail(r)))
	}

	return b.String()
}

func (v *DNSView) recordDetail(r api.DNSRecord) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s %s\n", styles.LabelStyle.Render("ID:"), r.ID))
	b.WriteString(fmt.Sprintf("%s %s\n", styles.LabelStyle.Render("Type:"), r.Type))
	b.WriteString(fmt.Sprintf("%s %s\n", styles.LabelStyle.Render("Name:"), r.Name))
	b.WriteString(fmt.Sprintf("%s %s\n", styles.LabelStyle.Render("Content:"), r.Content))
	b.WriteString(fmt.Sprintf("%s %s\n", styles.LabelStyle.Render("TTL:"), r.TTL))
	if r.Priority != "" && r.Priority != "0" {
		b.WriteString(fmt.Sprintf("%s %s\n", styles.LabelStyle.Render("Priority:"), r.Priority))
	}
	if r.Notes != "" {
		b.WriteString(fmt.Sprintf("%s %s", styles.LabelStyle.Render("Notes:"), r.Notes))
	}

	return b.String()
}

func (v *DNSView) HelpText() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		styles.HelpStyle.Render("j/k"),
		" navigate  ",
		styles.HelpStyle.Render("esc"),
		" back  ",
		styles.HelpStyle.Render("q"),
		" quit",
	)
}

func (v *DNSView) StatusText() string {
	return fmt.Sprintf("%d records", len(v.records))
}

func isAPIAccessError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not opted in") || strings.Contains(errStr, "api access")
}
