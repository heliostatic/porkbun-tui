package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	Highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	Special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	// Expiration colors
	ColorGreen  = lipgloss.Color("#73F59F")
	ColorYellow = lipgloss.Color("#F5D573")
	ColorOrange = lipgloss.Color("#F5A573")
	ColorRed    = lipgloss.Color("#F57373")
	ColorGray   = lipgloss.Color("#888888")

	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Title bar
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(Highlight).
			Padding(0, 1)

	// Status bar at bottom
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(Subtle).
			Padding(0, 1)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#5A5A5A")).
				Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(Highlight).
				Bold(true)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Domain status styles
	AutoRenewOnStyle = lipgloss.NewStyle().
				Foreground(ColorGreen)

	AutoRenewOffStyle = lipgloss.NewStyle().
				Foreground(ColorRed)

	// Search input
	SearchStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Highlight).
			Padding(0, 1)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	// Spinner style
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Highlight)

	// Detail view styles
	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Width(16)

	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	// Box style for panels
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Subtle).
			Padding(1)
)

// ExpirationStyle returns the appropriate style based on days until expiration
func ExpirationStyle(daysUntil int) lipgloss.Style {
	base := lipgloss.NewStyle()
	switch {
	case daysUntil < 0:
		return base.Foreground(ColorRed).Bold(true) // Expired
	case daysUntil < 7:
		return base.Foreground(ColorRed)
	case daysUntil < 30:
		return base.Foreground(ColorOrange)
	case daysUntil < 90:
		return base.Foreground(ColorYellow)
	default:
		return base.Foreground(ColorGreen)
	}
}
