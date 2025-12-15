package tui

import (
	"context"
	"fmt"

	"github.com/bc/porkbun-tui/internal/api"
	"github.com/bc/porkbun-tui/internal/cache"
	"github.com/bc/porkbun-tui/internal/keys"
	"github.com/bc/porkbun-tui/internal/styles"
	"github.com/bc/porkbun-tui/internal/tui/views"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type View int

const (
	ViewDomains View = iota
	ViewDetail
	ViewDNS
	ViewNameservers
	ViewAvailability
	ViewTLD
	ViewCalendar
	ViewHelp
)

type App struct {
	client *api.Client
	cache  *cache.Cache
	width  int
	height int

	// Current view
	view     View
	prevView View

	// Views
	domainsView      *views.DomainsView
	detailView       *views.DetailView
	dnsView          *views.DNSView
	nameserversView  *views.NameserversView
	availabilityView *views.AvailabilityView
	tldView          *views.TLDView
	calendarView     *views.CalendarView
	helpView         *views.HelpView

	// Data
	pricing map[string]api.TLDPricing

	// State
	loading    bool
	refreshing bool // True while background refresh in progress
	demoMode   bool // True when running without API credentials
	err        error
	spinner    spinner.Model
}

// Messages
type domainsLoadedMsg struct {
	domains []api.Domain
}

type dnsLoadedMsg struct {
	records []api.DNSRecord
}

type nsLoadedMsg struct {
	nameservers []string
}

type nsSavedMsg struct{}

type availabilityResultMsg struct {
	result *api.AvailabilityResult
}

type pricingLoadedMsg struct {
	pricing map[string]api.TLDPricing
}

type errMsg struct {
	err error
}

func NewApp(client *api.Client, appCache *cache.Cache, cachedDomains []api.Domain, cachedPricing map[string]api.TLDPricing, demoMode bool) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	domainsView := views.NewDomainsView()

	// If we have cached domains, show them immediately
	hasCachedDomains := len(cachedDomains) > 0
	if hasCachedDomains {
		domainsView.SetDomains(cachedDomains)
	}

	tldView := views.NewTLDView()
	calendarView := views.NewCalendarView()

	// If we have cached data, populate the views
	if hasCachedDomains {
		tldView.SetData(cachedDomains, cachedPricing)
		calendarView.SetDomains(cachedDomains)
	}

	return &App{
		client:           client,
		cache:            appCache,
		view:             ViewDomains,
		domainsView:      domainsView,
		detailView:       views.NewDetailView(),
		dnsView:          views.NewDNSView(),
		nameserversView:  views.NewNameserversView(),
		availabilityView: views.NewAvailabilityView(),
		tldView:          tldView,
		calendarView:     calendarView,
		helpView:         views.NewHelpView(),
		pricing:          cachedPricing,
		spinner:          s,
		loading:          !hasCachedDomains && !demoMode, // Only show loading if no cached data and not demo
		refreshing:       !demoMode,                      // Don't refresh in demo mode
		demoMode:         demoMode,
	}
}

func (a *App) Init() tea.Cmd {
	if a.demoMode {
		return nil // No API calls in demo mode
	}
	return tea.Batch(
		a.spinner.Tick,
		a.loadDomains(),
		a.loadPricing(),
	)
}

func (a *App) loadDomains() tea.Cmd {
	return func() tea.Msg {
		domains, err := a.client.ListDomains(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return domainsLoadedMsg{domains}
	}
}

func (a *App) loadPricing() tea.Cmd {
	return func() tea.Msg {
		pricing, err := a.client.GetPricing(context.Background())
		if err != nil {
			// Pricing errors are non-fatal, just log and continue
			return nil
		}
		return pricingLoadedMsg{pricing}
	}
}

func (a *App) loadDNS(domain string) tea.Cmd {
	return func() tea.Msg {
		records, err := a.client.GetDNSRecords(context.Background(), domain)
		if err != nil {
			return errMsg{err}
		}
		return dnsLoadedMsg{records}
	}
}

func (a *App) loadNameservers(domain string) tea.Cmd {
	return func() tea.Msg {
		ns, err := a.client.GetNameservers(context.Background(), domain)
		if err != nil {
			return errMsg{err}
		}
		return nsLoadedMsg{ns}
	}
}

func (a *App) saveNameservers(domain string, ns []string) tea.Cmd {
	return func() tea.Msg {
		err := a.client.UpdateNameservers(context.Background(), domain, ns)
		if err != nil {
			return errMsg{err}
		}
		return nsSavedMsg{}
	}
}

func (a *App) checkAvailability(domain string) tea.Cmd {
	return func() tea.Msg {
		result, err := a.client.CheckAvailability(context.Background(), domain)
		if err != nil {
			return errMsg{err}
		}
		return availabilityResultMsg{result}
	}
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.domainsView.SetSize(msg.Width, msg.Height)
		a.detailView.SetSize(msg.Width, msg.Height)
		a.dnsView.SetSize(msg.Width, msg.Height)
		a.nameserversView.SetSize(msg.Width, msg.Height)
		a.availabilityView.SetSize(msg.Width, msg.Height)
		a.tldView.SetSize(msg.Width, msg.Height)
		a.calendarView.SetSize(msg.Width, msg.Height)
		a.helpView.SetSize(msg.Width, msg.Height)

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case domainsLoadedMsg:
		a.loading = false
		a.refreshing = false
		a.domainsView.SetDomains(msg.domains)
		a.tldView.SetData(msg.domains, a.pricing)
		a.calendarView.SetDomains(msg.domains)
		// Save to cache
		if a.cache != nil {
			_ = a.cache.SaveDomains(msg.domains)
		}

	case dnsLoadedMsg:
		a.dnsView.SetRecords(msg.records)

	case nsLoadedMsg:
		a.nameserversView.SetNameservers(msg.nameservers)

	case nsSavedMsg:
		a.nameserversView.SetSuccess("Nameservers updated successfully!")
		// Reload nameservers to confirm
		if d := a.domainsView.SelectedDomain(); d != nil {
			cmds = append(cmds, a.loadNameservers(d.Name))
		}

	case availabilityResultMsg:
		a.availabilityView.SetResult(msg.result)

	case pricingLoadedMsg:
		a.pricing = msg.pricing
		// Update TLD view with new pricing
		domains := a.domainsView.GetDomains()
		if len(domains) > 0 {
			a.tldView.SetData(domains, a.pricing)
		}
		// Save to cache
		if a.cache != nil {
			_ = a.cache.SavePricing(msg.pricing)
		}

	case errMsg:
		a.err = msg.err
		a.loading = false
		a.refreshing = false
		switch a.view {
		case ViewDNS:
			a.dnsView.SetError(msg.err)
		case ViewNameservers:
			a.nameserversView.SetError(msg.err)
		case ViewAvailability:
			a.availabilityView.SetError(msg.err)
		}

	case tea.KeyMsg:
		// Skip global keys when searching in domains view
		isSearching := a.view == ViewDomains && a.domainsView.IsSearching()

		// Global keys
		switch {
		case key.Matches(msg, keys.Keys.Quit):
			if isSearching {
				break // Let search handle it
			}
			if a.view == ViewHelp {
				a.view = a.prevView
				return a, nil
			}
			return a, tea.Quit

		case key.Matches(msg, keys.Keys.Help):
			if isSearching {
				break // Let search handle it
			}
			if a.view == ViewHelp {
				a.view = a.prevView
			} else {
				a.prevView = a.view
				a.view = ViewHelp
			}
			return a, nil
		}

		// View-specific handling
		switch a.view {
		case ViewDomains:
			return a.updateDomains(msg)
		case ViewDetail:
			return a.updateDetail(msg)
		case ViewDNS:
			return a.updateDNS(msg)
		case ViewNameservers:
			return a.updateNameservers(msg)
		case ViewAvailability:
			return a.updateAvailability(msg)
		case ViewTLD:
			return a.updateTLD(msg)
		case ViewCalendar:
			return a.updateCalendar(msg)
		case ViewHelp:
			if key.Matches(msg, keys.Keys.Back) {
				a.view = a.prevView
			}
			return a, nil
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *App) updateDomains(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Skip command keys when searching
	if !a.domainsView.IsSearching() {
		switch {
		case key.Matches(msg, keys.Keys.Enter):
			if d := a.domainsView.SelectedDomain(); d != nil {
				a.detailView.SetDomain(d)
				a.view = ViewDetail
			}
			return a, nil

		case key.Matches(msg, keys.Keys.DNS):
			if a.demoMode {
				return a, nil // No DNS view in demo mode
			}
			if d := a.domainsView.SelectedDomain(); d != nil {
				a.dnsView.SetDomain(d.Name)
				a.view = ViewDNS
				return a, a.loadDNS(d.Name)
			}
			return a, nil

		case key.Matches(msg, keys.Keys.NS):
			if a.demoMode {
				return a, nil // No NS view in demo mode
			}
			if d := a.domainsView.SelectedDomain(); d != nil {
				a.nameserversView.SetDomain(d.Name)
				a.view = ViewNameservers
				return a, a.loadNameservers(d.Name)
			}
			return a, nil

		case key.Matches(msg, keys.Keys.Avail):
			if a.demoMode {
				return a, nil // No availability check in demo mode
			}
			a.view = ViewAvailability
			return a, a.availabilityView.Focus()

		case key.Matches(msg, keys.Keys.TLD):
			a.view = ViewTLD
			return a, nil

		case key.Matches(msg, keys.Keys.Calendar):
			a.view = ViewCalendar
			return a, nil

		case key.Matches(msg, keys.Keys.Refresh):
			if a.demoMode {
				return a, nil // No refresh in demo mode
			}
			a.refreshing = true
			return a, tea.Batch(a.loadDomains(), a.loadPricing())
		}
	}

	var cmd tea.Cmd
	a.domainsView, cmd = a.domainsView.Update(msg)
	return a, cmd
}

func (a *App) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Keys.Back):
		a.view = ViewDomains
		return a, nil

	case key.Matches(msg, keys.Keys.DNS):
		if a.demoMode {
			return a, nil // No DNS view in demo mode
		}
		if d := a.domainsView.SelectedDomain(); d != nil {
			a.dnsView.SetDomain(d.Name)
			a.view = ViewDNS
			return a, a.loadDNS(d.Name)
		}
		return a, nil

	case key.Matches(msg, keys.Keys.NS):
		if a.demoMode {
			return a, nil // No NS view in demo mode
		}
		if d := a.domainsView.SelectedDomain(); d != nil {
			a.nameserversView.SetDomain(d.Name)
			a.view = ViewNameservers
			return a, a.loadNameservers(d.Name)
		}
		return a, nil

	case key.Matches(msg, keys.Keys.Up), key.Matches(msg, keys.Keys.Down):
		// Navigate to prev/next domain while staying in detail view
		a.domainsView, _ = a.domainsView.Update(msg)
		if d := a.domainsView.SelectedDomain(); d != nil {
			a.detailView.SetDomain(d)
		}
		return a, nil
	}

	return a, nil
}

func (a *App) updateDNS(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Keys.Back) {
		a.view = ViewDomains
		return a, nil
	}

	var cmd tea.Cmd
	a.dnsView, cmd = a.dnsView.Update(msg)
	return a, cmd
}

func (a *App) updateNameservers(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Keys.Back) && !a.nameserversView.IsSaving() {
		a.view = ViewDomains
		return a, nil
	}

	// Check for save command
	if msg.String() == "ctrl+s" && a.nameserversView.IsSaving() {
		if d := a.domainsView.SelectedDomain(); d != nil {
			ns := a.nameserversView.GetNameservers()
			return a, a.saveNameservers(d.Name, ns)
		}
	}

	var cmd tea.Cmd
	a.nameserversView, cmd = a.nameserversView.Update(msg)

	// Check if save was triggered
	if a.nameserversView.IsSaving() {
		if d := a.domainsView.SelectedDomain(); d != nil {
			ns := a.nameserversView.GetNameservers()
			return a, a.saveNameservers(d.Name, ns)
		}
	}

	return a, cmd
}

func (a *App) updateAvailability(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Keys.Back) {
		a.view = ViewDomains
		return a, nil
	}

	if key.Matches(msg, keys.Keys.Enter) && !a.availabilityView.IsLoading() {
		domain := a.availabilityView.GetDomain()
		if domain != "" {
			a.availabilityView.SetLoading(true)
			a.availabilityView.ClearInput()
			return a, a.checkAvailability(domain)
		}
		return a, nil
	}

	var cmd tea.Cmd
	a.availabilityView, cmd = a.availabilityView.Update(msg)
	return a, cmd
}

func (a *App) updateTLD(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Keys.Back) {
		a.view = ViewDomains
		return a, nil
	}

	var cmd tea.Cmd
	a.tldView, cmd = a.tldView.Update(msg)
	return a, cmd
}

func (a *App) updateCalendar(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Keys.Back) {
		a.view = ViewDomains
		return a, nil
	}

	var cmd tea.Cmd
	a.calendarView, cmd = a.calendarView.Update(msg)
	return a, cmd
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	var content string

	if a.loading && a.view == ViewDomains {
		content = fmt.Sprintf("\n  %s Loading domains...", a.spinner.View())
	} else {
		switch a.view {
		case ViewDomains:
			content = a.domainsView.View()
		case ViewDetail:
			content = a.detailView.View()
		case ViewDNS:
			content = a.dnsView.View()
		case ViewNameservers:
			content = a.nameserversView.View()
		case ViewAvailability:
			content = a.availabilityView.View()
		case ViewTLD:
			content = a.tldView.View()
		case ViewCalendar:
			content = a.calendarView.View()
		case ViewHelp:
			content = a.helpView.View()
		}
	}

	// Build layout
	titleBar := styles.TitleStyle.Render(" Porkbun TUI ")
	statusBar := a.statusBar()
	helpBar := a.helpBar()

	// Calculate content height
	contentHeight := a.height - 4 // title + status + help + padding

	// Render
	return lipgloss.JoinVertical(lipgloss.Left,
		titleBar,
		lipgloss.NewStyle().Height(contentHeight).Render(content),
		statusBar,
		helpBar,
	)
}

func (a *App) statusBar() string {
	var status string

	switch a.view {
	case ViewDomains:
		status = a.domainsView.StatusText()
	case ViewDetail:
		if d := a.domainsView.SelectedDomain(); d != nil {
			status = d.Name
		}
	case ViewDNS:
		status = a.dnsView.StatusText()
	case ViewNameservers:
		status = a.nameserversView.StatusText()
	case ViewAvailability:
		status = a.availabilityView.StatusText()
	case ViewTLD:
		status = a.tldView.StatusText()
	case ViewCalendar:
		status = a.calendarView.StatusText()
	case ViewHelp:
		status = "Help"
	}

	// Show refresh indicator
	if a.refreshing {
		status = "↻ " + status
	}

	if a.err != nil {
		status = styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", a.err))
	}

	return styles.StatusBarStyle.Width(a.width).Render(status)
}

func (a *App) helpBar() string {
	var help string

	switch a.view {
	case ViewDomains:
		help = a.domainsView.HelpText()
	case ViewDetail:
		help = a.detailView.HelpText()
	case ViewDNS:
		help = a.dnsView.HelpText()
	case ViewNameservers:
		help = a.nameserversView.HelpText()
	case ViewAvailability:
		help = a.availabilityView.HelpText()
	case ViewTLD:
		help = a.tldView.HelpText()
	case ViewCalendar:
		help = a.calendarView.HelpText()
	case ViewHelp:
		help = styles.HelpStyle.Render("? or esc to close")
	}

	return help
}
