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

// buyReadyApp returns an app in the availability view with a fresh
// AVAILABLE result, ready for the buy flow.
func buyReadyApp(t *testing.T) *App {
	t.Helper()
	a := newTestApp(false)
	a.view = ViewAvailability
	a, _ = update(t, a, availabilityResultMsg{&api.AvailabilityResult{
		Domain: "fresh.xyz", Available: true, Price: "2.04",
	}})
	return a
}

func TestAvailabilityCtrlBEntersConfirmation(t *testing.T) {
	a := buyReadyApp(t)

	a, cmd := update(t, a, tea.KeyMsg{Type: tea.KeyCtrlB})

	if !a.availabilityView.IsConfirming() {
		t.Error("ctrl+b did not enter buy confirmation")
	}
	if cmd != nil {
		t.Error("ctrl+b queued a command; nothing should fire before y")
	}
}

func TestAvailabilityConfirmYStartsPurchase(t *testing.T) {
	a := buyReadyApp(t)
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyCtrlB})

	a, cmd := update(t, a, keyMsg("y"))

	if !a.availabilityView.IsPurchasing() {
		t.Error("view not purchasing after y")
	}
	if a.availabilityView.IsConfirming() {
		t.Error("still confirming after y")
	}
	if cmd == nil {
		t.Error("no purchase command queued after y")
	}
}

func TestAvailabilityConfirmNCancels(t *testing.T) {
	a := buyReadyApp(t)
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyCtrlB})

	a, cmd := update(t, a, keyMsg("n"))

	if a.availabilityView.IsConfirming() {
		t.Error("still confirming after n")
	}
	if a.availabilityView.IsPurchasing() {
		t.Error("purchasing after cancel")
	}
	if cmd != nil {
		t.Error("cancel queued a command")
	}
}

func TestAvailabilityConfirmEscCancelsWithoutLeavingView(t *testing.T) {
	a := buyReadyApp(t)
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyCtrlB})

	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyEsc})

	if a.view != ViewAvailability {
		t.Error("esc during confirmation left the availability view")
	}
	if a.availabilityView.IsConfirming() {
		t.Error("still confirming after esc")
	}
}

func TestAvailabilityKeysSwallowedWhileConfirming(t *testing.T) {
	a := buyReadyApp(t)
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyCtrlB})

	a, cmd := update(t, a, keyMsg("x"))

	if got := a.availabilityView.GetDomain(); got != "" {
		t.Errorf("typed rune reached the input during confirmation: %q", got)
	}
	if cmd != nil {
		t.Error("swallowed key queued a command")
	}
	if !a.availabilityView.IsConfirming() {
		t.Error("stray key ended the confirmation")
	}
}

func TestAvailabilityEnterDoesNotCheckWhileConfirming(t *testing.T) {
	a := buyReadyApp(t)
	for _, r := range "x.com" {
		a, _ = update(t, a, keyMsg(string(r)))
	}
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyCtrlB})

	a, cmd := update(t, a, tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("enter fired a command while confirming a purchase")
	}
	if a.availabilityView.IsLoading() {
		t.Error("a check started while confirming a purchase")
	}
}

func TestPurchaseResultMsgRefreshesDomains(t *testing.T) {
	a := buyReadyApp(t)
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyCtrlB})
	a, _ = update(t, a, keyMsg("y"))

	a, cmd := update(t, a, purchaseResultMsg{&api.RegistrationResult{
		Domain: "fresh.xyz", OrderID: 123456, CostCents: 204, BalanceCents: 5000,
	}})

	if a.availabilityView.IsPurchasing() {
		t.Error("still purchasing after result")
	}
	if !strings.Contains(a.availabilityView.View(), "Registered") {
		t.Error("success not shown in availability view")
	}
	if !a.refreshing {
		t.Error("domains not refreshing after a purchase")
	}
	if cmd == nil {
		t.Error("no refresh command queued after a purchase")
	}
}

func TestPurchaseErrMsgClearsPurchasingOffView(t *testing.T) {
	a := buyReadyApp(t)
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyCtrlB})
	a, _ = update(t, a, keyMsg("y"))

	// Navigate away before the purchase fails; the error must still clear
	// the in-flight state (same defect class as the availability soft-lock).
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyEsc})
	a, _ = update(t, a, purchaseErrMsg{errors.New("insufficient funds")})

	if a.availabilityView.IsPurchasing() {
		t.Error("purchase error off-view left the buy flow soft-locked")
	}
	if !strings.Contains(a.availabilityView.View(), "insufficient funds") {
		t.Error("purchase error not shown on return to the view")
	}
}

func TestAvailabilityEnterDoesNotCheckWhilePurchasing(t *testing.T) {
	a := buyReadyApp(t)
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyCtrlB})
	a, _ = update(t, a, keyMsg("y")) // purchase in flight

	for _, r := range "bar.com" {
		a, _ = update(t, a, keyMsg(string(r)))
	}
	a, cmd := update(t, a, tea.KeyMsg{Type: tea.KeyEnter})

	if a.availabilityView.IsLoading() {
		t.Error("a check started while a purchase is in flight; its result would wipe the receipt")
	}
	if cmd != nil {
		t.Error("enter queued a command while purchasing")
	}
}

func TestTypingQInAvailabilityInputDoesNotQuit(t *testing.T) {
	a := newTestApp(false)
	a.view = ViewAvailability

	a, cmd := update(t, a, keyMsg("q"))

	if cmd != nil {
		if _, quit := cmd().(tea.QuitMsg); quit {
			t.Fatal("typing q into the availability input quit the app")
		}
	}
	if got := a.availabilityView.GetDomain(); got != "q" {
		t.Errorf("input = %q after typing q, want %q", got, "q")
	}
}

func TestTypingHelpKeyInAvailabilityInputDoesNotOpenHelp(t *testing.T) {
	a := newTestApp(false)
	a.view = ViewAvailability

	a, _ = update(t, a, keyMsg("?"))

	if a.view != ViewAvailability {
		t.Errorf("view = %v after typing ?, want ViewAvailability", a.view)
	}
}

func TestQDuringBuyConfirmationDoesNotQuit(t *testing.T) {
	a := buyReadyApp(t)
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyCtrlB})

	a, cmd := update(t, a, keyMsg("q"))

	if cmd != nil {
		if _, quit := cmd().(tea.QuitMsg); quit {
			t.Fatal("q during the purchase confirmation quit the app")
		}
	}
	if !a.availabilityView.IsConfirming() {
		t.Error("q ended the confirmation; it should be swallowed")
	}
}

func TestHelpKeyDuringBuyConfirmationIsSwallowed(t *testing.T) {
	a := buyReadyApp(t)
	a, _ = update(t, a, tea.KeyMsg{Type: tea.KeyCtrlB})

	a, _ = update(t, a, keyMsg("?"))

	if a.view != ViewAvailability {
		t.Errorf("? during confirmation switched view to %v; an armed y/n prompt must not survive a view excursion", a.view)
	}
	if !a.availabilityView.IsConfirming() {
		t.Error("? cancelled the confirmation")
	}
}

func TestCtrlCAlwaysQuits(t *testing.T) {
	a := newTestApp(false)
	a.view = ViewAvailability

	_, cmd := update(t, a, tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Fatal("ctrl+c returned no command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("ctrl+c returned %T, want tea.QuitMsg", cmd())
	}
}

func TestTypingQInNameserverEditDoesNotQuit(t *testing.T) {
	a := newTestApp(false)
	a.view = ViewNameservers
	a.nameserversView.SetNameservers([]string{"ns1.example.com"})
	a, _ = update(t, a, keyMsg("e")) // enter edit mode

	a, cmd := update(t, a, keyMsg("q"))

	if cmd != nil {
		if _, quit := cmd().(tea.QuitMsg); quit {
			t.Fatal("typing q into a nameserver input quit the app")
		}
	}
	if a.view != ViewNameservers {
		t.Errorf("view = %v, want ViewNameservers", a.view)
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
