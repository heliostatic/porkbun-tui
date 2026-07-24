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

	// Buy flow: confirming is the result awaiting a y/n answer,
	// pendingCents its price converted for the create endpoint.
	confirming   *api.AvailabilityResult
	pendingCents int
	purchasing   bool
	purchased    string
}

func NewAvailabilityView() *AvailabilityView {
	ti := textinput.New()
	ti.Placeholder = "example.com"
	ti.CharLimit = 100
	ti.Width = 30 // bubbles renders only Width+1 placeholder runes
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
	v.purchased = ""
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

// StartBuyConfirmation arms the y/n purchase prompt for the most recent
// check. It refuses quietly when there is nothing buyable (no results,
// latest is taken, or a purchase is already in flight) and surfaces an
// error when the price cannot be converted to cents.
func (v *AvailabilityView) StartBuyConfirmation() {
	if v.purchasing || v.confirming != nil || len(v.results) == 0 {
		return
	}
	latest := v.results[0]
	if !latest.Available {
		return
	}
	cents, err := api.DollarsToCents(latest.Price)
	if err != nil {
		v.err = fmt.Errorf("cannot buy %s: unparsable price %q", latest.Domain, latest.Price)
		return
	}
	v.confirming = &latest
	v.pendingCents = cents
	v.err = nil
	v.purchased = ""
}

func (v *AvailabilityView) IsConfirming() bool {
	return v.confirming != nil
}

func (v *AvailabilityView) PendingPurchase() (string, int) {
	if v.confirming == nil {
		return "", 0
	}
	return v.confirming.Domain, v.pendingCents
}

func (v *AvailabilityView) CancelBuyConfirmation() {
	v.confirming = nil
}

func (v *AvailabilityView) SetPurchasing() {
	v.confirming = nil
	v.purchasing = true
}

func (v *AvailabilityView) IsPurchasing() bool {
	return v.purchasing
}

func (v *AvailabilityView) SetPurchaseResult(r *api.RegistrationResult) {
	v.purchasing = false
	v.err = nil
	if r != nil {
		v.purchased = fmt.Sprintf("Registered %s — order #%d, balance %s",
			r.Domain, r.OrderID, centsToDollars(r.BalanceCents))
	}
}

func (v *AvailabilityView) SetPurchaseError(err error) {
	v.purchasing = false
	v.err = err
}

func centsToDollars(c int) string {
	return fmt.Sprintf("$%d.%02d", c/100, c%100)
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

	// Input; rendered bare — the SearchStyle border produces visual
	// artifacts (same defect previously fixed in the nameserver view).
	b.WriteString("  Enter domain to check:\n\n")
	b.WriteString("  ")
	b.WriteString(v.input.View())
	b.WriteString("\n\n")

	if v.loading {
		b.WriteString(styles.SpinnerStyle.Render("  Checking availability..."))
		b.WriteString("\n")
	}

	if v.confirming != nil {
		prompt := fmt.Sprintf("  Buy %s for %s? This will charge your Porkbun account balance.",
			v.confirming.Domain, centsToDollars(v.pendingCents))
		b.WriteString(styles.PremiumStyle.Render(prompt))
		b.WriteString("\n")
		b.WriteString(styles.HelpStyle.Render("  y confirm · n cancel"))
		b.WriteString("\n\n")
	}

	if v.purchasing {
		b.WriteString(styles.SpinnerStyle.Render("  Purchasing..."))
		b.WriteString("\n\n")
	}

	if v.purchased != "" {
		b.WriteString(styles.SuccessStyle.Render("  " + v.purchased))
		b.WriteString("\n\n")
	}

	if v.err != nil {
		b.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("  Error: %v", v.err)))
		b.WriteString("\n\n")
	}

	// Results
	if len(v.results) > 0 {
		const domainWidth = 34

		b.WriteString("  Recent checks:\n\n")
		header := fmt.Sprintf("  %-*s  %-9s  %9s", domainWidth, "Domain", "Status", "Price/yr")
		b.WriteString(styles.TableHeaderStyle.Render(header))
		b.WriteString("\n")

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

			price := ""
			if r.Available && r.Price != "" {
				price = fmt.Sprintf("$%8s", r.Price)
			}

			row := fmt.Sprintf("  %-*s  %s  %s",
				domainWidth, truncate(r.Domain, domainWidth),
				style.Render(fmt.Sprintf("%-9s", status)),
				price,
			)
			if r.Premium {
				row += styles.PremiumStyle.Render("  premium")
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
		styles.HelpStyle.Render("ctrl+b"),
		" buy  ",
		styles.HelpStyle.Render("esc"),
		" back  ",
		styles.HelpStyle.Render("q"),
		" quit",
	)
}

func (v *AvailabilityView) StatusText() string {
	return "Domain availability checker"
}
