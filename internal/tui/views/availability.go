package views

import (
	"fmt"
	"strings"

	"github.com/bc/porkbun-tui/internal/api"
	"github.com/bc/porkbun-tui/internal/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AvailabilityView struct {
	input   textinput.Model
	results []api.AvailabilityResult
	loading bool
	err     error
	width   int
	height  int
}

func NewAvailabilityView() *AvailabilityView {
	ti := textinput.New()
	ti.Placeholder = "example.com"
	ti.CharLimit = 100
	ti.Focus()

	return &AvailabilityView{
		input: ti,
	}
}

func (v *AvailabilityView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *AvailabilityView) SetLoading(loading bool) {
	v.loading = loading
}

func (v *AvailabilityView) SetResult(result *api.AvailabilityResult) {
	v.loading = false
	v.err = nil
	if result != nil {
		v.results = append([]api.AvailabilityResult{*result}, v.results...)
		// Keep only last 10 results
		if len(v.results) > 10 {
			v.results = v.results[:10]
		}
	}
}

func (v *AvailabilityView) SetError(err error) {
	v.loading = false
	v.err = err
}

func (v *AvailabilityView) GetDomain() string {
	return strings.TrimSpace(v.input.Value())
}

func (v *AvailabilityView) ClearInput() {
	v.input.SetValue("")
}

func (v *AvailabilityView) IsLoading() bool {
	return v.loading
}

func (v *AvailabilityView) Focus() tea.Cmd {
	v.input.Focus()
	return textinput.Blink
}

func (v *AvailabilityView) Update(msg tea.Msg) (*AvailabilityView, tea.Cmd) {
	var cmd tea.Cmd
	v.input, cmd = v.input.Update(msg)
	return v, cmd
}

func (v *AvailabilityView) View() string {
	var b strings.Builder

	// Title
	title := styles.TitleStyle.Render(" Domain Availability Checker ")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Input
	b.WriteString("  Enter domain to check:\n\n")
	b.WriteString("  ")
	b.WriteString(styles.SearchStyle.Render(v.input.View()))
	b.WriteString("\n\n")

	if v.loading {
		b.WriteString(styles.SpinnerStyle.Render("  Checking availability..."))
		b.WriteString("\n")
	}

	if v.err != nil {
		b.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("  Error: %v", v.err)))
		b.WriteString("\n\n")
	}

	// Results
	if len(v.results) > 0 {
		b.WriteString("  Recent checks:\n\n")

		for _, r := range v.results {
			var status string
			var style lipgloss.Style

			if r.Available {
				status = "AVAILABLE"
				style = styles.SuccessStyle
			} else {
				status = "TAKEN"
				style = styles.ErrorStyle
			}

			row := fmt.Sprintf("  %-30s  %s", r.Domain, style.Render(status))
			if r.Available && r.Price != "" {
				priceStr := r.Price
				if r.Premium {
					priceStr += " (premium)"
				}
				row += fmt.Sprintf("  %s", styles.HelpStyle.Render(priceStr))
			}
			b.WriteString(row)
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (v *AvailabilityView) HelpText() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		styles.HelpStyle.Render("enter"),
		" check  ",
		styles.HelpStyle.Render("esc"),
		" back  ",
		styles.HelpStyle.Render("q"),
		" quit",
	)
}

func (v *AvailabilityView) StatusText() string {
	return "Domain availability checker"
}
