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
	a.view = ViewAvailability

	a, _ = update(t, a, errMsg{errors.New("check failed")})

	if a.err == nil {
		t.Error("app err not set")
	}
	if a.loading || a.refreshing {
		t.Error("loading/refreshing not cleared on error")
	}
	if !strings.Contains(a.availabilityView.View(), "check failed") {
		t.Error("error not routed to availability view")
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
	for _, k := range []string{"a", "d", "n", "r"} {
		a := newTestApp(true)
		a, _ = update(t, a, keyMsg(k))
		if a.view != ViewDomains {
			t.Errorf("demo mode: view = %v after %q, want ViewDomains", a.view, k)
		}
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
