package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/bc/porkbun-tui/internal/api"
	"github.com/bc/porkbun-tui/internal/styles"
	"github.com/charmbracelet/lipgloss"
)

type DetailView struct {
	domain *api.Domain
	width  int
	height int
}

func NewDetailView() *DetailView {
	return &DetailView{}
}

func (v *DetailView) SetDomain(d *api.Domain) {
	v.domain = d
}

func (v *DetailView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *DetailView) View() string {
	if v.domain == nil {
		return "No domain selected"
	}

	d := v.domain
	var b strings.Builder

	// Title
	title := styles.TitleStyle.Render(fmt.Sprintf(" %s ", d.Name))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Domain info
	daysUntil := int(time.Until(d.ExpireDate).Hours() / 24)

	rows := []struct {
		label string
		value string
		style lipgloss.Style
	}{
		{"TLD", d.TLD, styles.ValueStyle},
		{"Status", d.Status, styles.ValueStyle},
		{"Created", d.CreateDate.Format("2006-01-02"), styles.ValueStyle},
		{"Expires", d.ExpireDate.Format("2006-01-02"), styles.ExpirationStyle(daysUntil)},
		{"Days Left", fmt.Sprintf("%d", daysUntil), styles.ExpirationStyle(daysUntil)},
		{"Auto-Renew", boolToYesNo(d.AutoRenew), boolStyle(d.AutoRenew)},
		{"Security Lock", boolToYesNo(d.SecurityLock), boolStyle(d.SecurityLock)},
		{"WHOIS Privacy", boolToYesNo(d.WhoisPrivacy), boolStyle(d.WhoisPrivacy)},
	}

	labelWidth := 16 // Wide enough for "WHOIS Privacy:"
	for _, row := range rows {
		labelText := fmt.Sprintf("%-*s", labelWidth, row.label+":")
		label := styles.LabelStyle.Render(labelText)
		value := row.style.Render(row.value)
		b.WriteString(fmt.Sprintf("  %s %s\n", label, value))
	}

	// Labels if any
	if len(d.Labels) > 0 {
		labelText := fmt.Sprintf("%-*s", labelWidth, "Labels:")
		label := styles.LabelStyle.Render(labelText)
		value := styles.ValueStyle.Render(strings.Join(d.Labels, ", "))
		b.WriteString(fmt.Sprintf("  %s %s\n", label, value))
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("  j/k: prev/next domain  d: DNS  n: nameservers  esc: back"))

	return b.String()
}

func (v *DetailView) HelpText() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		styles.HelpStyle.Render("j/k"),
		" prev/next  ",
		styles.HelpStyle.Render("d"),
		" dns  ",
		styles.HelpStyle.Render("n"),
		" nameservers  ",
		styles.HelpStyle.Render("esc"),
		" back  ",
		styles.HelpStyle.Render("q"),
		" quit",
	)
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func boolStyle(b bool) lipgloss.Style {
	if b {
		return styles.AutoRenewOnStyle
	}
	return styles.AutoRenewOffStyle
}
