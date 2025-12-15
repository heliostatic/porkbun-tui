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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MonthGroup struct {
	Year     int
	Month    time.Month
	Domains  []api.Domain
	Expanded bool
}

type CalendarView struct {
	groups []MonthGroup
	cursor int // Which month group is selected
	offset int // Line offset for scrolling
	height int
	width  int
}

func NewCalendarView() *CalendarView {
	return &CalendarView{}
}

func (v *CalendarView) SetDomains(domains []api.Domain) {
	// Group domains by expiration month/year
	groupMap := make(map[string]*MonthGroup)

	for _, d := range domains {
		key := fmt.Sprintf("%d-%02d", d.ExpireDate.Year(), d.ExpireDate.Month())
		if g, ok := groupMap[key]; ok {
			g.Domains = append(g.Domains, d)
		} else {
			groupMap[key] = &MonthGroup{
				Year:     d.ExpireDate.Year(),
				Month:    d.ExpireDate.Month(),
				Domains:  []api.Domain{d},
				Expanded: false,
			}
		}
	}

	// Convert to slice
	v.groups = make([]MonthGroup, 0, len(groupMap))
	for _, g := range groupMap {
		// Sort domains within group by expiration date
		sort.Slice(g.Domains, func(i, j int) bool {
			return g.Domains[i].ExpireDate.Before(g.Domains[j].ExpireDate)
		})
		v.groups = append(v.groups, *g)
	}

	// Sort groups by date (earliest first)
	sort.Slice(v.groups, func(i, j int) bool {
		if v.groups[i].Year != v.groups[j].Year {
			return v.groups[i].Year < v.groups[j].Year
		}
		return v.groups[i].Month < v.groups[j].Month
	})

	// Auto-expand first month (usually the most urgent)
	if len(v.groups) > 0 {
		v.groups[0].Expanded = true
	}

	v.cursor = 0
	v.offset = 0
}

func (v *CalendarView) SetSize(width, height int) {
	v.width = width
	v.height = height - 6
	if v.height < 1 {
		v.height = 1
	}
}

// getLineForCursor returns the line number where the cursor's month header is
func (v *CalendarView) getLineForCursor() int {
	line := 0
	for i := 0; i < v.cursor && i < len(v.groups); i++ {
		line++ // Month header line
		if v.groups[i].Expanded {
			line += len(v.groups[i].Domains)
		}
	}
	return line
}

// getTotalLines returns the total number of content lines
func (v *CalendarView) getTotalLines() int {
	lines := 0
	for _, g := range v.groups {
		lines++ // Month header
		if g.Expanded {
			lines += len(g.Domains)
		}
	}
	return lines
}

func (v *CalendarView) adjustOffset() {
	cursorLine := v.getLineForCursor()

	// If cursor is above visible area, scroll up
	if cursorLine < v.offset {
		v.offset = cursorLine
	}

	// If cursor is below visible area, scroll down
	if cursorLine >= v.offset+v.height {
		v.offset = cursorLine - v.height + 1
	}

	// Make sure we don't scroll past the end
	totalLines := v.getTotalLines()
	if v.offset > totalLines-v.height {
		v.offset = max(0, totalLines-v.height)
	}
}

func (v *CalendarView) Update(msg tea.Msg) (*CalendarView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Up):
			if v.cursor > 0 {
				v.cursor--
				v.adjustOffset()
			}
		case key.Matches(msg, keys.Keys.Down):
			if v.cursor < len(v.groups)-1 {
				v.cursor++
				v.adjustOffset()
			}
		case key.Matches(msg, keys.Keys.Enter):
			if v.cursor < len(v.groups) {
				v.groups[v.cursor].Expanded = !v.groups[v.cursor].Expanded
				v.adjustOffset()
			}
		}
	}
	return v, nil
}

func (v *CalendarView) View() string {
	var b strings.Builder

	// Title
	title := styles.TitleStyle.Render(" Expiration Calendar ")
	b.WriteString(title)
	b.WriteString("\n\n")

	if len(v.groups) == 0 {
		b.WriteString("  No domains to display.")
		return b.String()
	}

	// Build all lines first, then slice for viewport
	var lines []string

	for i, g := range v.groups {
		// Month header
		expandIndicator := "▶"
		if g.Expanded {
			expandIndicator = "▼"
		}

		header := fmt.Sprintf("%s %s %d (%d domains)",
			expandIndicator,
			g.Month.String(),
			g.Year,
			len(g.Domains),
		)

		// Color the header based on urgency
		daysUntilFirst := int(time.Until(g.Domains[0].ExpireDate).Hours() / 24)

		if i == v.cursor {
			header = styles.TableSelectedStyle.Render("  " + header)
		} else {
			header = "  " + styles.ExpirationStyle(daysUntilFirst).Render(header)
		}

		lines = append(lines, header)

		// Add expanded domains
		if g.Expanded {
			for _, d := range g.Domains {
				daysUntil := int(time.Until(d.ExpireDate).Hours() / 24)
				expDate := d.ExpireDate.Format("Jan 02")

				var daysStr string
				if daysUntil < 0 {
					daysStr = "EXPIRED"
				} else {
					daysStr = fmt.Sprintf("%d days", daysUntil)
				}

				domainLine := fmt.Sprintf("      %-30s  %s  %s",
					truncate(d.Name, 30),
					expDate,
					daysStr,
				)

				domainLine = styles.ExpirationStyle(daysUntil).Render(domainLine)
				lines = append(lines, domainLine)
			}
		}
	}

	// Render visible lines based on offset
	visibleEnd := min(v.offset+v.height, len(lines))
	for i := v.offset; i < visibleEnd; i++ {
		b.WriteString(lines[i])
		b.WriteString("\n")
	}

	// Scroll indicator if content exceeds viewport
	totalLines := len(lines)
	if totalLines > v.height {
		scrollInfo := fmt.Sprintf(" %d-%d of %d lines ", v.offset+1, visibleEnd, totalLines)
		b.WriteString(styles.HelpStyle.Render(scrollInfo))
	}

	return b.String()
}

func (v *CalendarView) HelpText() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		styles.HelpStyle.Render("j/k"),
		" navigate  ",
		styles.HelpStyle.Render("enter"),
		" expand  ",
		styles.HelpStyle.Render("esc"),
		" back  ",
		styles.HelpStyle.Render("q"),
		" quit",
	)
}

func (v *CalendarView) StatusText() string {
	totalDomains := 0
	for _, g := range v.groups {
		totalDomains += len(g.Domains)
	}
	return fmt.Sprintf("%d months, %d domains", len(v.groups), totalDomains)
}
