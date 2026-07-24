package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bc/porkbun-tui/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestApp(demoMode bool) *App {
	return NewApp(nil, nil, nil, nil, demoMode)
}

func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// update sends a message and returns the App (Update returns tea.Model).
func update(t *testing.T, a *App, msg tea.Msg) (*App, tea.Cmd) {
	t.Helper()
	model, cmd := a.Update(msg)
	app, ok := model.(*App)
	if !ok {
		t.Fatalf("Update returned %T, want *App", model)
	}
	return app, cmd
}

func TestWindowSizeMsgSetsDimensions(t *testing.T) {
	a := newTestApp(false)
	a, _ = update(t, a, tea.WindowSizeMsg{Width: 120, Height: 40})

	if a.width != 120 || a.height != 40 {
		t.Errorf("size = %dx%d, want 120x40", a.width, a.height)
	}
}

func TestDomainsLoadedMsgPopulatesViewsAndClearsLoading(t *testing.T) {
	a := newTestApp(false)
	domains := []api.Domain{
		{Name: "example.com", TLD: "com", ExpireDate: time.Now().AddDate(1, 0, 0)},
		{Name: "example.org", TLD: "org", ExpireDate: time.Now().AddDate(0, 6, 0)},
	}

	a, _ = update(t, a, domainsLoadedMsg{domains})

	if a.loading {
		t.Error("loading = true after domainsLoadedMsg, want false")
	}
	if a.refreshing {
		t.Error("refreshing = true after domainsLoadedMsg, want false")
	}
	if got := len(a.domainsView.GetDomains()); got != 2 {
		t.Errorf("domainsView has %d domains, want 2", got)
	}
}

func TestPricingLoadedMsgStoresPricing(t *testing.T) {
	a := newTestApp(false)
	pricing := map[string]api.TLDPricing{
		"com": {TLD: "com", Registration: "11.06", Renewal: "11.06"},
	}

	a, _ = update(t, a, pricingLoadedMsg{pricing})

	if a.pricing["com"].Renewal != "11.06" {
		t.Errorf("pricing not stored: %v", a.pricing)
	}
}

func TestAvailabilityResultMsgReachesView(t *testing.T) {
	a := newTestApp(false)
	a, _ = update(t, a, availabilityResultMsg{&api.AvailabilityResult{Domain: "open.com", Available: true}})

	if !strings.Contains(a.availabilityView.View(), "open.com") {
		t.Error("availability view does not show the checked domain")
	}
}

func TestErrMsgRoutesToActiveView(t *testing.T) {
	a := newTestApp(false)
	a.view = ViewDNS

	a, _ = update(t, a, errMsg{errors.New("dns lookup failed")})

	if a.err == nil {
		t.Error("app err not set")
	}
	if a.loading || a.refreshing {
		t.Error("loading/refreshing not cleared on error")
	}
	if !strings.Contains(a.dnsView.View(), "dns lookup failed") {
		t.Error("error not routed to DNS view")
	}
}

func TestAvailabilityErrorClearsLoadingAfterLeavingView(t *testing.T) {
	a := newTestApp(false)
	a.view = ViewAvailability
	a.availabilityView.SetLoading(true)

	// User navigates away while the check is in flight, then it fails.
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyEsc})
	if a.view != ViewDomains {
		t.Fatalf("view = %v after esc, want ViewDomains", a.view)
	}
	a, _ = update(t, a, availabilityErrMsg{errors.New("rate limited")})

	if a.availabilityView.IsLoading() {
		t.Error("availability view still loading after error arrived off-view; checker is soft-locked")
	}
	if !strings.Contains(a.availabilityView.View(), "rate limited") {
		t.Error("availability error not shown when user returns to the view")
	}
}

func TestUnrelatedErrorDoesNotReachAvailabilityView(t *testing.T) {
	a := newTestApp(false)
	a.view = ViewAvailability
	a.availabilityView.SetLoading(true)

	// A background domains-refresh failure lands while the availability view
	// is open; it must not be misattributed to the in-flight check.
	a, _ = update(t, a, errMsg{errors.New("refresh blew up")})

	if !a.availabilityView.IsLoading() {
		t.Error("unrelated error cleared the availability in-flight guard")
	}
	if strings.Contains(a.availabilityView.View(), "refresh blew up") {
		t.Error("unrelated error rendered in the availability view")
	}
}

func TestHelpKeyTogglesHelpView(t *testing.T) {
	a := newTestApp(false)

	a, _ = update(t, a, keyMsg("?"))
	if a.view != ViewHelp {
		t.Fatalf("view = %v after ?, want ViewHelp", a.view)
	}

	a, _ = update(t, a, keyMsg("?"))
	if a.view != ViewDomains {
		t.Errorf("view = %v after second ?, want ViewDomains", a.view)
	}
}

func TestViewSwitchingKeys(t *testing.T) {
	a := newTestApp(false)

	a, _ = update(t, a, keyMsg("t"))
	if a.view != ViewTLD {
		t.Fatalf("view = %v after t, want ViewTLD", a.view)
	}

	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyEsc})
	if a.view != ViewDomains {
		t.Fatalf("view = %v after esc, want ViewDomains", a.view)
	}

	a, _ = update(t, a, keyMsg("c"))
	if a.view != ViewCalendar {
		t.Fatalf("view = %v after c, want ViewCalendar", a.view)
	}

	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyEsc})
	a, _ = update(t, a, keyMsg("a"))
	if a.view != ViewAvailability {
		t.Errorf("view = %v after a, want ViewAvailability", a.view)
	}
}

func TestDemoModeBlocksAPIDependentViews(t *testing.T) {
	// Domains must be loaded so SelectedDomain() is non-nil: without one the
	// d/n handlers bail out on their own and the guard assertions are vacuous.
	domains := []api.Domain{{Name: "example.com", TLD: "com"}}

	for _, k := range []string{"a", "d", "n", "r"} {
		a := NewApp(nil, nil, domains, nil, true)
		a2, cmd := update(t, a, keyMsg(k))
		if a2.view != ViewDomains {
			t.Errorf("demo mode: view = %v after %q, want ViewDomains", a2.view, k)
		}
		// The guard must return before any command is created; a non-nil cmd
		// means an API call (against a nil client here) was queued.
		if cmd != nil {
			t.Errorf("demo mode: cmd != nil after %q; an API command was queued", k)
		}
	}
}

func TestAvailabilityEnterStartsCheck(t *testing.T) {
	a := newTestApp(false)
	a.view = ViewAvailability

	for _, r := range "x.com" {
		a, _ = update(t, a, keyMsg(string(r)))
	}
	a, cmd := update(t, a, tea.KeyMsg{Type: tea.KeyEnter})

	if !a.availabilityView.IsLoading() {
		t.Error("view not loading after Enter with a domain")
	}
	if got := a.availabilityView.GetDomain(); got != "" {
		t.Errorf("input = %q after Enter, want cleared", got)
	}
	if cmd == nil {
		t.Error("cmd = nil after Enter; no check command was queued")
	}
}

func TestAvailabilityEnterWhileLoadingDoesNotStartAnother(t *testing.T) {
	a := newTestApp(false)
	a.view = ViewAvailability

	for _, r := range "x.com" {
		a, _ = update(t, a, keyMsg(string(r)))
	}
	a.availabilityView.SetLoading(true)
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyEnter})

	// The check path clears the input; an intact input proves the in-flight
	// guard swallowed the Enter instead of firing a second rate-limited call.
	if got := a.availabilityView.GetDomain(); got != "x.com" {
		t.Errorf("input = %q after guarded Enter, want x.com untouched", got)
	}
}

func TestAvailabilityEnterWithEmptyInputDoesNothing(t *testing.T) {
	a := newTestApp(false)
	a.view = ViewAvailability

	a, cmd := update(t, a, tea.KeyMsg{Type: tea.KeyEnter})

	if a.availabilityView.IsLoading() {
		t.Error("view loading after Enter on empty input")
	}
	if cmd != nil {
		t.Error("cmd != nil after Enter on empty input; a check was queued")
	}
}

func TestQuitKeyReturnsQuit(t *testing.T) {
	a := newTestApp(false)
	_, cmd := update(t, a, keyMsg("q"))

	if cmd == nil {
		t.Fatal("cmd = nil after q, want tea.Quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("cmd() = %T, want tea.QuitMsg", cmd())
	}
}
