package views

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bc/porkbun-tui/internal/api"
	"github.com/bc/porkbun-tui/internal/keys"
	"github.com/bc/porkbun-tui/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TLDGroup struct {
	TLD          string
	Domains      []api.Domain
	RenewalPrice float64
	TotalCost    float64
}

type TLDView struct {
	groups   []TLDGroup
	cursor   int // Which TLD group is selected
	offset   int // Line offset for scrolling
	height   int
	width    int
	expanded map[string]bool
}

func NewTLDView() *TLDView {
	return &TLDView{
		expanded: make(map[string]bool),
	}
}

func (v *TLDView) SetData(domains []api.Domain, pricing map[string]api.TLDPricing) {
	// Group domains by TLD
	groupMap := make(map[string][]api.Domain)
	for _, d := range domains {
		groupMap[d.TLD] = append(groupMap[d.TLD], d)
	}

	// Convert to slice and calculate costs
	v.groups = make([]TLDGroup, 0, len(groupMap))
	for tld, doms := range groupMap {
		var renewalPrice float64
		if p, ok := pricing[tld]; ok {
			renewalPrice, _ = strconv.ParseFloat(p.Renewal, 64)
		}

		v.groups = append(v.groups, TLDGroup{
			TLD:          tld,
			Domains:      doms,
			RenewalPrice: renewalPrice,
			TotalCost:    renewalPrice * float64(len(doms)),
		})
	}

	// Sort by total cost (highest first)
	sort.Slice(v.groups, func(i, j int) bool {
		return v.groups[i].TotalCost > v.groups[j].TotalCost
	})

	v.cursor = 0
	v.offset = 0
}

func (v *TLDView) SetSize(width, height int) {
	v.width = width
	v.height = height - 8
	if v.height < 1 {
		v.height = 1
	}
}

// getLineForCursor returns the line number where the cursor's TLD header is
func (v *TLDView) getLineForCursor() int {
	line := 0
	for i := 0; i < v.cursor && i < len(v.groups); i++ {
		line++ // TLD header line
		if v.expanded[v.groups[i].TLD] {
			line += len(v.groups[i].Domains)
		}
	}
	return line
}

// getTotalLines returns the total number of content lines
func (v *TLDView) getTotalLines() int {
	lines := 0
	for _, g := range v.groups {
		lines++ // TLD header
		if v.expanded[g.TLD] {
			lines += len(g.Domains)
		}
	}
	return lines
}

func (v *TLDView) adjustOffset() {
	cursorLine := v.getLineForCursor()

	// If cursor is above visible area, scroll up
	if cursorLine < v.offset {
		v.offset = cursorLine
	}

	// If cursor is below visible area, scroll down
	// Account for the TLD header line itself
	if cursorLine >= v.offset+v.height {
		v.offset = cursorLine - v.height + 1
	}

	// Make sure we don't scroll past the end
	totalLines := v.getTotalLines()
	if v.offset > totalLines-v.height {
		v.offset = max(0, totalLines-v.height)
	}
}

func (v *TLDView) Update(msg tea.Msg) (*TLDView, tea.Cmd) {
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
				tld := v.groups[v.cursor].TLD
				v.expanded[tld] = !v.expanded[tld]
				v.adjustOffset()
			}
		}
	}
	return v, nil
}

func (v *TLDView) View() string {
	var b strings.Builder

	// Title
	title := styles.TitleStyle.Render(" TLD Breakdown ")
	b.WriteString(title)
	b.WriteString("\n\n")

	if len(v.groups) == 0 {
		b.WriteString("  No domains to display.")
		return b.String()
	}

	// Column widths
	tldWidth := 12
	countWidth := 8
	renewalWidth := 12
	totalWidth := 14

	// Header
	header := fmt.Sprintf("  %-*s  %*s  %*s  %*s",
		tldWidth, "TLD",
		countWidth, "Count",
		renewalWidth, "Renewal",
		totalWidth, "Total/Year",
	)
	b.WriteString(styles.TableHeaderStyle.Render(header))
	b.WriteString("\n")

	// Calculate grand totals
	var grandTotalDomains int
	var grandTotalCost float64
	for _, g := range v.groups {
		grandTotalDomains += len(g.Domains)
		grandTotalCost += g.TotalCost
	}

	// Build all lines first, then slice for viewport
	var lines []string

	for i, g := range v.groups {
		// Format fields
		tldStr := fmt.Sprintf("%-*s", tldWidth, g.TLD)
		countStr := fmt.Sprintf("%*d", countWidth, len(g.Domains))

		var renewalStr, totalStr string
		if g.RenewalPrice > 0 {
			renewalStr = fmt.Sprintf("%*.2f", renewalWidth-1, g.RenewalPrice)
			totalStr = fmt.Sprintf("%*.2f", totalWidth-1, g.TotalCost)
		} else {
			renewalStr = fmt.Sprintf("%*s", renewalWidth, "N/A")
			totalStr = fmt.Sprintf("%*s", totalWidth, "N/A")
		}

		// Expand indicator
		expandIndicator := "▶"
		if v.expanded[g.TLD] {
			expandIndicator = "▼"
		}

		row := fmt.Sprintf("%s %-*s  %s  $%s  $%s",
			expandIndicator,
			tldWidth, tldStr,
			countStr,
			renewalStr,
			totalStr,
		)

		if i == v.cursor {
			row = styles.TableSelectedStyle.Render(row)
		}

		lines = append(lines, row)

		// Add expanded domains
		if v.expanded[g.TLD] {
			for _, d := range g.Domains {
				domainRow := styles.HelpStyle.Render(fmt.Sprintf("    %s", d.Name))
				lines = append(lines, domainRow)
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
		b.WriteString("\n")
	}

	// Grand total
	b.WriteString("\n")
	totalLine := fmt.Sprintf("  Total: %d domains, $%.2f/year", grandTotalDomains, grandTotalCost)
	b.WriteString(styles.ValueStyle.Render(totalLine))

	return b.String()
}

func (v *TLDView) HelpText() string {
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

func (v *TLDView) StatusText() string {
	return fmt.Sprintf("%d TLDs", len(v.groups))
}
