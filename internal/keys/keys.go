package keys

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Search   key.Binding
	Refresh  key.Binding
	Help     key.Binding
	Quit     key.Binding
	DNS      key.Binding
	NS       key.Binding
	Avail    key.Binding
	TLD      key.Binding
	Calendar key.Binding
	SortName key.Binding
	SortExp  key.Binding
	Tab      key.Binding
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k", "ctrl+k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j", "ctrl+j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	DNS: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "dns records"),
	),
	NS: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "nameservers"),
	),
	Avail: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "check availability"),
	),
	TLD: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "TLD breakdown"),
	),
	Calendar: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "calendar view"),
	),
	SortName: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "sort by name"),
	),
	SortExp: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "sort by expiration"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Search, k.Help, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Back},
		{k.Search, k.Refresh, k.SortName, k.SortExp},
		{k.DNS, k.NS, k.Avail, k.TLD, k.Calendar},
		{k.Help, k.Quit},
	}
}
