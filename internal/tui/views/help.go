package views

import (
	"fmt"
	"strings"

	"github.com/bc/porkbun-tui/internal/styles"
)

type HelpView struct {
	width  int
	height int
}

func NewHelpView() *HelpView {
	return &HelpView{}
}

func (v *HelpView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *HelpView) View() string {
	var b strings.Builder

	title := styles.TitleStyle.Render(" Keyboard Shortcuts ")
	b.WriteString(title)
	b.WriteString("\n\n")

	sections := []struct {
		title string
		items []struct {
			key  string
			desc string
		}
	}{
		{
			title: "Navigation",
			items: []struct {
				key  string
				desc string
			}{
				{"j / k / Up / Down", "Move up/down in lists"},
				{"Enter", "Select / open details"},
				{"Esc", "Go back / cancel"},
				{"Tab", "Next field (in forms)"},
			},
		},
		{
			title: "Domain List",
			items: []struct {
				key  string
				desc string
			}{
				{"/", "Search/filter domains"},
				{"1", "Sort by name"},
				{"2", "Sort by expiration date"},
				{"r", "Refresh domain list"},
			},
		},
		{
			title: "Views",
			items: []struct {
				key  string
				desc string
			}{
				{"d", "View DNS records"},
				{"n", "View/edit nameservers"},
				{"a", "Domain availability checker"},
				{"t", "TLD breakdown (costs by TLD)"},
				{"c", "Calendar view (by expiration)"},
			},
		},
		{
			title: "Nameserver Edit",
			items: []struct {
				key  string
				desc string
			}{
				{"e", "Edit nameservers"},
				{"p", "Apply preset (Cloudflare, etc.)"},
				{"Ctrl+S", "Save changes"},
			},
		},
		{
			title: "General",
			items: []struct {
				key  string
				desc string
			}{
				{"?", "Toggle this help"},
				{"q / Ctrl+C", "Quit"},
			},
		},
	}

	for _, section := range sections {
		b.WriteString(styles.TableHeaderStyle.Render(fmt.Sprintf(" %s ", section.title)))
		b.WriteString("\n")
		for _, item := range section.items {
			key := styles.LabelStyle.Width(20).Render(item.key)
			desc := styles.ValueStyle.Render(item.desc)
			b.WriteString(fmt.Sprintf("  %s %s\n", key, desc))
		}
		b.WriteString("\n")
	}

	b.WriteString(styles.HelpStyle.Render("  Press ? or Esc to close"))

	return b.String()
}
