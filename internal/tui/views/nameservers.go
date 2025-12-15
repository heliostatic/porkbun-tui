package views

import (
	"fmt"
	"strings"

	"github.com/bc/porkbun-tui/internal/keys"
	"github.com/bc/porkbun-tui/internal/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NSViewMode int

const (
	NSViewModeView NSViewMode = iota
	NSViewModeEdit
	NSViewModePreset
)

type NSPreset struct {
	Name string
	NS   []string
}

var NSPresets = []NSPreset{
	{
		Name: "Porkbun Default",
		NS:   []string{"maceio.ns.porkbun.com", "curitiba.ns.porkbun.com", "salvador.ns.porkbun.com", "fortaleza.ns.porkbun.com"},
	},
	{
		Name: "Cloudflare",
		NS:   []string{"ns1.cloudflare.com", "ns2.cloudflare.com"},
	},
	{
		Name: "Google Cloud DNS",
		NS:   []string{"ns-cloud-a1.googledomains.com", "ns-cloud-a2.googledomains.com", "ns-cloud-a3.googledomains.com", "ns-cloud-a4.googledomains.com"},
	},
}

type NameserversView struct {
	domain      string
	nameservers []string
	inputs      []textinput.Model
	cursor      int
	mode        NSViewMode
	presetIdx   int
	loading     bool
	saving      bool
	err         error
	success     string
	width       int
	height      int
}

func NewNameserversView() *NameserversView {
	return &NameserversView{
		mode: NSViewModeView,
	}
}

func (v *NameserversView) SetDomain(domain string) {
	v.domain = domain
	v.nameservers = nil
	v.loading = true
	v.err = nil
	v.success = ""
	v.mode = NSViewModeView
}

func (v *NameserversView) SetNameservers(ns []string) {
	v.nameservers = ns
	v.loading = false
	v.initInputs()
}

func (v *NameserversView) SetError(err error) {
	v.err = err
	v.loading = false
	v.saving = false
}

func (v *NameserversView) SetSuccess(msg string) {
	v.success = msg
	v.saving = false
}

func (v *NameserversView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *NameserversView) initInputs() {
	v.inputs = make([]textinput.Model, 4)
	for i := 0; i < 4; i++ {
		ti := textinput.New()
		ti.Placeholder = fmt.Sprintf("ns%d.example.com", i+1)
		ti.CharLimit = 100
		if i < len(v.nameservers) {
			ti.SetValue(v.nameservers[i])
		}
		v.inputs[i] = ti
	}
	v.cursor = 0
}

func (v *NameserversView) GetNameservers() []string {
	var ns []string
	for _, input := range v.inputs {
		if val := strings.TrimSpace(input.Value()); val != "" {
			ns = append(ns, val)
		}
	}
	return ns
}

func (v *NameserversView) IsSaving() bool {
	return v.saving
}

func (v *NameserversView) StartSaving() {
	v.saving = true
	v.err = nil
	v.success = ""
}

func (v *NameserversView) Update(msg tea.Msg) (*NameserversView, tea.Cmd) {
	switch v.mode {
	case NSViewModeEdit:
		return v.updateEdit(msg)
	case NSViewModePreset:
		return v.updatePreset(msg)
	default:
		return v.updateView(msg)
	}
}

func (v *NameserversView) updateView(msg tea.Msg) (*NameserversView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			v.mode = NSViewModeEdit
			v.inputs[0].Focus()
			return v, textinput.Blink
		case "p":
			v.mode = NSViewModePreset
			v.presetIdx = 0
		}
	}
	return v, nil
}

func (v *NameserversView) updateEdit(msg tea.Msg) (*NameserversView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Back):
			v.mode = NSViewModeView
			v.inputs[v.cursor].Blur()
			return v, nil
		case key.Matches(msg, keys.Keys.Tab), key.Matches(msg, keys.Keys.Down):
			v.inputs[v.cursor].Blur()
			v.cursor = (v.cursor + 1) % len(v.inputs)
			v.inputs[v.cursor].Focus()
			return v, textinput.Blink
		case key.Matches(msg, keys.Keys.Up):
			v.inputs[v.cursor].Blur()
			v.cursor = (v.cursor - 1 + len(v.inputs)) % len(v.inputs)
			v.inputs[v.cursor].Focus()
			return v, textinput.Blink
		case msg.String() == "ctrl+s":
			v.StartSaving()
			return v, nil // App will handle the actual save
		}
	}

	// Update current input
	v.inputs[v.cursor], _ = v.inputs[v.cursor].Update(msg)
	return v, nil
}

func (v *NameserversView) updatePreset(msg tea.Msg) (*NameserversView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Back):
			v.mode = NSViewModeView
		case key.Matches(msg, keys.Keys.Up):
			if v.presetIdx > 0 {
				v.presetIdx--
			}
		case key.Matches(msg, keys.Keys.Down):
			if v.presetIdx < len(NSPresets)-1 {
				v.presetIdx++
			}
		case key.Matches(msg, keys.Keys.Enter):
			// Apply preset
			preset := NSPresets[v.presetIdx]
			for i := range v.inputs {
				if i < len(preset.NS) {
					v.inputs[i].SetValue(preset.NS[i])
				} else {
					v.inputs[i].SetValue("")
				}
			}
			v.mode = NSViewModeEdit
			v.inputs[0].Focus()
			return v, textinput.Blink
		}
	}
	return v, nil
}

func (v *NameserversView) View() string {
	var b strings.Builder

	// Title
	title := styles.TitleStyle.Render(fmt.Sprintf(" Nameservers: %s ", v.domain))
	b.WriteString(title)
	b.WriteString("\n\n")

	if v.loading {
		b.WriteString("  Loading nameservers...")
		return b.String()
	}

	if v.err != nil {
		b.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("  Error: %v", v.err)))
		if isNSAPIAccessError(v.err) {
			b.WriteString("\n")
			b.WriteString(styles.HelpStyle.Render("  This domain needs API access enabled.\n"))
			b.WriteString(styles.HelpStyle.Render("  Go to porkbun.com → Domain Management → " + v.domain + " → API Access → ON"))
		}
		b.WriteString("\n\n")
	}

	if v.success != "" {
		b.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("  %s\n\n", v.success)))
	}

	switch v.mode {
	case NSViewModePreset:
		b.WriteString("  Select a preset:\n\n")
		for i, preset := range NSPresets {
			cursor := "  "
			if i == v.presetIdx {
				cursor = "> "
			}
			row := fmt.Sprintf("%s%s", cursor, preset.Name)
			if i == v.presetIdx {
				row = styles.TableSelectedStyle.Render(row)
			}
			b.WriteString(row)
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(styles.HelpStyle.Render("  enter to apply, esc to cancel"))

	case NSViewModeEdit:
		b.WriteString("  Edit nameservers:\n\n")
		for i, input := range v.inputs {
			label := fmt.Sprintf("  NS%d: ", i+1)
			b.WriteString(styles.LabelStyle.Render(label))
			if i == v.cursor {
				b.WriteString(styles.SearchStyle.Render(input.View()))
			} else {
				b.WriteString(input.View())
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		if v.saving {
			b.WriteString(styles.SpinnerStyle.Render("  Saving..."))
		} else {
			b.WriteString(styles.HelpStyle.Render("  tab/j/k to navigate, ctrl+s to save, esc to cancel"))
		}

	default:
		b.WriteString("  Current nameservers:\n\n")
		if len(v.nameservers) == 0 {
			b.WriteString("  No nameservers configured.\n")
		} else {
			for i, ns := range v.nameservers {
				b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, styles.ValueStyle.Render(ns)))
			}
		}
		b.WriteString("\n")
		b.WriteString(styles.HelpStyle.Render("  e to edit, p for presets, esc to go back"))
	}

	return b.String()
}

func (v *NameserversView) HelpText() string {
	switch v.mode {
	case NSViewModeEdit:
		return lipgloss.JoinHorizontal(lipgloss.Top,
			styles.HelpStyle.Render("tab/j/k"),
			" navigate  ",
			styles.HelpStyle.Render("ctrl+s"),
			" save  ",
			styles.HelpStyle.Render("esc"),
			" cancel",
		)
	case NSViewModePreset:
		return lipgloss.JoinHorizontal(lipgloss.Top,
			styles.HelpStyle.Render("j/k"),
			" navigate  ",
			styles.HelpStyle.Render("enter"),
			" apply  ",
			styles.HelpStyle.Render("esc"),
			" cancel",
		)
	default:
		return lipgloss.JoinHorizontal(lipgloss.Top,
			styles.HelpStyle.Render("e"),
			" edit  ",
			styles.HelpStyle.Render("p"),
			" presets  ",
			styles.HelpStyle.Render("esc"),
			" back  ",
			styles.HelpStyle.Render("q"),
			" quit",
		)
	}
}

func (v *NameserversView) StatusText() string {
	return fmt.Sprintf("%d nameservers", len(v.nameservers))
}

func isNSAPIAccessError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not opted in") || strings.Contains(errStr, "api access")
}
