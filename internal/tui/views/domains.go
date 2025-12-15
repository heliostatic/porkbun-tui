package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bc/porkbun-tui/internal/api"
	"github.com/bc/porkbun-tui/internal/keys"
	"github.com/bc/porkbun-tui/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SortField int

const (
	SortByName SortField = iota
	SortByExpiration
)

type DomainsView struct {
	domains       []api.Domain
	filtered      []api.Domain
	cursor        int
	offset        int
	height        int
	width         int
	searchInput   textinput.Model
	searching     bool
	sortField     SortField
	sortAscending bool
}

func NewDomainsView() *DomainsView {
	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.Prompt = ""
	ti.CharLimit = 50
	ti.Width = 30

	return &DomainsView{
		searchInput:   ti,
		sortField:     SortByExpiration,
		sortAscending: true,
	}
}

func (v *DomainsView) SetDomains(domains []api.Domain) {
	v.domains = domains
	v.applyFilter()
	v.sortDomains()
}

func (v *DomainsView) SetSize(width, height int) {
	v.width = width
	v.height = height - 6 // Account for header, status bar, etc.
	if v.height < 1 {
		v.height = 1
	}
}

func (v *DomainsView) applyFilter() {
	query := strings.ToLower(v.searchInput.Value())
	if query == "" {
		v.filtered = v.domains
		return
	}

	v.filtered = nil
	for _, d := range v.domains {
		if strings.Contains(strings.ToLower(d.Name), query) {
			v.filtered = append(v.filtered, d)
		}
	}

	// Reset cursor if out of bounds
	if v.cursor >= len(v.filtered) {
		v.cursor = max(0, len(v.filtered)-1)
	}
}

func (v *DomainsView) sortDomains() {
	sort.Slice(v.filtered, func(i, j int) bool {
		var less bool
		switch v.sortField {
		case SortByName:
			less = v.filtered[i].Name < v.filtered[j].Name
		case SortByExpiration:
			less = v.filtered[i].ExpireDate.Before(v.filtered[j].ExpireDate)
		}
		if !v.sortAscending {
			return !less
		}
		return less
	})
}

func (v *DomainsView) SelectedDomain() *api.Domain {
	if len(v.filtered) == 0 || v.cursor >= len(v.filtered) {
		return nil
	}
	return &v.filtered[v.cursor]
}

func (v *DomainsView) GetDomains() []api.Domain {
	return v.domains
}

func (v *DomainsView) IsSearching() bool {
	return v.searching
}

func (v *DomainsView) Update(msg tea.Msg) (*DomainsView, tea.Cmd) {
	var cmd tea.Cmd

	if v.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", "esc":
				v.searching = false
				v.searchInput.Blur()
				return v, nil
			case "ctrl+k", "up":
				if v.cursor > 0 {
					v.cursor--
					if v.cursor < v.offset {
						v.offset = v.cursor
					}
				}
				return v, nil
			case "ctrl+j", "down":
				if v.cursor < len(v.filtered)-1 {
					v.cursor++
					if v.cursor >= v.offset+v.height {
						v.offset = v.cursor - v.height + 1
					}
				}
				return v, nil
			}
		}
		v.searchInput, cmd = v.searchInput.Update(msg)
		v.applyFilter()
		v.sortDomains()
		return v, cmd
	}

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
			if v.cursor < len(v.filtered)-1 {
				v.cursor++
				if v.cursor >= v.offset+v.height {
					v.offset = v.cursor - v.height + 1
				}
			}
		case key.Matches(msg, keys.Keys.Search):
			v.searching = true
			v.searchInput.Focus()
			return v, textinput.Blink
		case key.Matches(msg, keys.Keys.Back):
			// Clear search filter if active
			if v.searchInput.Value() != "" {
				v.searchInput.SetValue("")
				v.applyFilter()
				v.sortDomains()
			}
			return v, nil
		case key.Matches(msg, keys.Keys.SortName):
			if v.sortField == SortByName {
				v.sortAscending = !v.sortAscending
			} else {
				v.sortField = SortByName
				v.sortAscending = true
			}
			v.sortDomains()
		case key.Matches(msg, keys.Keys.SortExp):
			if v.sortField == SortByExpiration {
				v.sortAscending = !v.sortAscending
			} else {
				v.sortField = SortByExpiration
				v.sortAscending = true
			}
			v.sortDomains()
		}
	}

	return v, nil
}

func (v *DomainsView) View() string {
	if len(v.filtered) == 0 {
		if v.searchInput.Value() != "" {
			return "\n  No domains match your search."
		}
		return "\n  No domains found."
	}

	var b strings.Builder
	b.WriteString("\n")

	// Search bar if searching or filter active
	if v.searching {
		b.WriteString("  Search: ")
		b.WriteString(v.searchInput.View())
		b.WriteString("\n")
	} else if v.searchInput.Value() != "" {
		b.WriteString(styles.HelpStyle.Render(fmt.Sprintf("  Filter: \"%s\" (/ to edit, esc to clear)", v.searchInput.Value())))
		b.WriteString("\n")
	}

	// Column widths
	nameWidth := 35
	expWidth := 12
	daysWidth := 8
	autoWidth := 10
	statusWidth := 10

	// Header
	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s  %-*s",
		nameWidth, v.sortIndicator("Domain", SortByName),
		expWidth, v.sortIndicator("Expires", SortByExpiration),
		daysWidth, "Days",
		autoWidth, "AutoRenew",
		statusWidth, "Status",
	)
	b.WriteString(styles.TableHeaderStyle.Render(header))
	b.WriteString("\n")

	// Calculate visible range
	visibleEnd := min(v.offset+v.height, len(v.filtered))

	// Rows
	for i := v.offset; i < visibleEnd; i++ {
		d := v.filtered[i]
		daysUntil := int(time.Until(d.ExpireDate).Hours() / 24)

		// Format fields
		name := truncate(d.Name, nameWidth)
		expDate := d.ExpireDate.Format("2006-01-02")
		daysStr := fmt.Sprintf("%d", daysUntil)
		if daysUntil < 0 {
			daysStr = "EXPIRED"
		}

		autoRenew := "No"
		autoStyle := styles.AutoRenewOffStyle
		if d.AutoRenew {
			autoRenew = "Yes"
			autoStyle = styles.AutoRenewOnStyle
		}

		status := d.Status
		if status == "" {
			status = "Active"
		}

		// Build row with fixed-width columns (style applied after padding)
		namePad := fmt.Sprintf("%-*s", nameWidth, name)
		expPad := fmt.Sprintf("%-*s", expWidth, expDate)
		daysPad := fmt.Sprintf("%-*s", daysWidth, daysStr)
		autoPad := fmt.Sprintf("%-*s", autoWidth, autoRenew)
		statusPad := fmt.Sprintf("%-*s", statusWidth, status)

		// Apply colors to individual columns
		daysStyled := styles.ExpirationStyle(daysUntil).Render(daysPad)
		autoStyled := autoStyle.Render(autoPad)

		row := fmt.Sprintf("  %s  %s  %s  %s  %s",
			namePad,
			expPad,
			daysStyled,
			autoStyled,
			statusPad,
		)

		if i == v.cursor {
			row = styles.TableSelectedStyle.Render(row)
		}

		b.WriteString(row)
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(v.filtered) > v.height {
		scrollInfo := fmt.Sprintf(" %d-%d of %d ", v.offset+1, visibleEnd, len(v.filtered))
		b.WriteString(styles.HelpStyle.Render(scrollInfo))
	}

	return b.String()
}

func (v *DomainsView) sortIndicator(label string, field SortField) string {
	if v.sortField != field {
		return label
	}
	arrow := "▲"
	if !v.sortAscending {
		arrow = "▼"
	}
	return label + " " + arrow
}

func (v *DomainsView) StatusText() string {
	total := len(v.domains)
	filtered := len(v.filtered)

	if v.searchInput.Value() != "" {
		return fmt.Sprintf("%d/%d domains", filtered, total)
	}
	return fmt.Sprintf("%d domains", total)
}

func (v *DomainsView) HelpText() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		styles.HelpStyle.Render("j/k"),
		" navigate  ",
		styles.HelpStyle.Render("enter"),
		" details  ",
		styles.HelpStyle.Render("/"),
		" search  ",
		styles.HelpStyle.Render("d"),
		" dns  ",
		styles.HelpStyle.Render("n"),
		" ns  ",
		styles.HelpStyle.Render("t"),
		" tld  ",
		styles.HelpStyle.Render("c"),
		" cal  ",
		styles.HelpStyle.Render("?"),
		" help  ",
		styles.HelpStyle.Render("q"),
		" quit",
	)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
