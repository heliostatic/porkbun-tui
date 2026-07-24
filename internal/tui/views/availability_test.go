package views

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/bc/porkbun-tui/internal/api"
)

func TestAvailabilitySetResultPrependsAndCaps(t *testing.T) {
	v := NewAvailabilityView()

	for i := 0; i < 12; i++ {
		v.SetResult(&api.AvailabilityResult{Domain: fmt.Sprintf("domain%d.com", i)})
	}

	if len(v.results) != 10 {
		t.Fatalf("len(results) = %d, want 10 (capped)", len(v.results))
	}
	if v.results[0].Domain != "domain11.com" {
		t.Errorf("results[0].Domain = %q, want domain11.com (newest first)", v.results[0].Domain)
	}
}

func TestAvailabilitySetResultClearsLoadingAndError(t *testing.T) {
	v := NewAvailabilityView()
	v.SetError(errors.New("boom"))
	v.SetLoading(true)

	v.SetResult(&api.AvailabilityResult{Domain: "example.com"})

	if v.IsLoading() {
		t.Error("IsLoading() = true after SetResult, want false")
	}
	if v.err != nil {
		t.Errorf("err = %v after SetResult, want nil", v.err)
	}
}

func TestAvailabilityViewRendersStatuses(t *testing.T) {
	v := NewAvailabilityView()
	v.SetResult(&api.AvailabilityResult{Domain: "taken.com", Available: false})
	v.SetResult(&api.AvailabilityResult{Domain: "open.com", Available: true, Price: "11.06"})
	v.SetResult(&api.AvailabilityResult{Domain: "fancy.com", Available: true, Price: "1200.00", Premium: true})

	// Assert per row so a status/domain mix-up ("taken.com AVAILABLE") fails.
	rows := map[string][]string{
		"taken.com": {"TAKEN"},
		"open.com":  {"AVAILABLE", "11.06"},
		"fancy.com": {"AVAILABLE", "1200.00", "premium"},
	}
	for _, line := range strings.Split(v.View(), "\n") {
		for domain, wants := range rows {
			if !strings.Contains(line, domain) {
				continue
			}
			for _, want := range wants {
				if !strings.Contains(line, want) {
					t.Errorf("row for %s missing %q: %q", domain, want, line)
				}
			}
			delete(rows, domain)
		}
	}
	for domain := range rows {
		t.Errorf("View() has no row for %s", domain)
	}
}

func TestAvailabilityViewInputHasNoBorderArtifacts(t *testing.T) {
	v := NewAvailabilityView()

	// The SearchStyle rounded border renders with artifacts around the
	// textinput (same defect previously fixed in the nameserver view).
	if strings.Contains(v.View(), "╭") {
		t.Error("View() wraps the input in a border, which renders broken")
	}
}

func TestAvailabilityViewShowsFullPlaceholder(t *testing.T) {
	v := NewAvailabilityView()

	// bubbles renders only Width+1 placeholder runes; with Width unset the
	// "example.com" placeholder collapses to "> e".
	if !strings.Contains(v.View(), "example.com") {
		t.Error("View() does not show the full placeholder")
	}
}

func TestAvailabilityViewRendersPriceColumn(t *testing.T) {
	v := NewAvailabilityView()
	v.SetResult(&api.AvailabilityResult{Domain: "open.com", Available: true, Price: "2.04"})

	out := v.View()
	if !strings.Contains(out, "Price/yr") {
		t.Error("View() missing the Price/yr column header")
	}
	if !strings.Contains(out, "$") {
		t.Error("View() renders the price without a currency symbol")
	}
}

func TestAvailabilityViewRendersError(t *testing.T) {
	v := NewAvailabilityView()
	v.SetError(errors.New("rate limit exceeded"))

	if out := v.View(); !strings.Contains(out, "rate limit exceeded") {
		t.Error("View() does not render the error message")
	}
}

func TestAvailabilityBuyConfirmationTargetsLatestAvailable(t *testing.T) {
	v := NewAvailabilityView()
	v.SetResult(&api.AvailabilityResult{Domain: "old-taken.com", Available: false})
	v.SetResult(&api.AvailabilityResult{Domain: "fresh.xyz", Available: true, Price: "2.04"})

	v.StartBuyConfirmation()

	if !v.IsConfirming() {
		t.Fatal("not confirming after StartBuyConfirmation on an available latest result")
	}
	domain, cents := v.PendingPurchase()
	if domain != "fresh.xyz" {
		t.Errorf("pending domain = %q, want fresh.xyz", domain)
	}
	if cents != 204 {
		t.Errorf("pending cents = %d, want 204", cents)
	}
}

func TestAvailabilityBuyConfirmationRefusedWhenLatestTaken(t *testing.T) {
	v := NewAvailabilityView()
	v.SetResult(&api.AvailabilityResult{Domain: "open.com", Available: true, Price: "9.00"})
	// The taken result carries a parsable price so this test pins the
	// availability guard itself, not the price-parse guard behind it.
	v.SetResult(&api.AvailabilityResult{Domain: "taken.com", Available: false, Price: "9.99"})

	v.StartBuyConfirmation()

	if v.IsConfirming() {
		t.Error("confirming a purchase when the latest check is TAKEN")
	}
}

func TestAvailabilityBuyConfirmationRefusedWithNoResults(t *testing.T) {
	v := NewAvailabilityView()

	v.StartBuyConfirmation()

	if v.IsConfirming() {
		t.Error("confirming a purchase with no results")
	}
}

func TestAvailabilityBuyConfirmationRejectsUnparsablePrice(t *testing.T) {
	v := NewAvailabilityView()
	v.SetResult(&api.AvailabilityResult{Domain: "weird.com", Available: true, Price: "N/A"})

	v.StartBuyConfirmation()

	if v.IsConfirming() {
		t.Error("confirming a purchase with an unparsable price")
	}
	if !strings.Contains(v.View(), "N/A") {
		t.Error("View() does not surface the unparsable price problem")
	}
}

func TestAvailabilityBuyConfirmationRefusedWhilePurchasing(t *testing.T) {
	v := NewAvailabilityView()
	v.SetResult(&api.AvailabilityResult{Domain: "fresh.xyz", Available: true, Price: "2.04"})
	v.StartBuyConfirmation()
	v.SetPurchasing()

	v.StartBuyConfirmation()

	if v.IsConfirming() {
		t.Error("confirmation restarted while a purchase is in flight")
	}
}

func TestAvailabilityViewRendersConfirmationPrompt(t *testing.T) {
	v := NewAvailabilityView()
	v.SetResult(&api.AvailabilityResult{Domain: "fresh.xyz", Available: true, Price: "2.04"})
	v.StartBuyConfirmation()

	out := v.View()
	for _, want := range []string{"fresh.xyz", "$2.04", "balance", "y", "n"} {
		if !strings.Contains(out, want) {
			t.Errorf("confirmation prompt missing %q", want)
		}
	}
}

func TestAvailabilityCancelBuyConfirmation(t *testing.T) {
	v := NewAvailabilityView()
	v.SetResult(&api.AvailabilityResult{Domain: "fresh.xyz", Available: true, Price: "2.04"})
	v.StartBuyConfirmation()

	v.CancelBuyConfirmation()

	if v.IsConfirming() {
		t.Error("still confirming after cancel")
	}
	if v.IsPurchasing() {
		t.Error("purchasing after cancel")
	}
}

func TestAvailabilityPurchaseResultRendersSuccess(t *testing.T) {
	v := NewAvailabilityView()
	v.SetResult(&api.AvailabilityResult{Domain: "fresh.xyz", Available: true, Price: "2.04"})
	v.StartBuyConfirmation()
	v.SetPurchasing()

	v.SetPurchaseResult(&api.RegistrationResult{
		Domain: "fresh.xyz", OrderID: 123456, CostCents: 204, BalanceCents: 5000,
	})

	if v.IsPurchasing() {
		t.Error("still purchasing after result")
	}
	out := v.View()
	for _, want := range []string{"Registered", "fresh.xyz", "123456", "$50.00"} {
		if !strings.Contains(out, want) {
			t.Errorf("success message missing %q", want)
		}
	}
}

func TestAvailabilityPurchaseErrorClearsPurchasing(t *testing.T) {
	v := NewAvailabilityView()
	v.SetResult(&api.AvailabilityResult{Domain: "fresh.xyz", Available: true, Price: "2.04"})
	v.StartBuyConfirmation()
	v.SetPurchasing()

	v.SetPurchaseError(errors.New("insufficient funds"))

	if v.IsPurchasing() {
		t.Error("still purchasing after error; buy flow is soft-locked")
	}
	if !strings.Contains(v.View(), "insufficient funds") {
		t.Error("purchase error not rendered")
	}
}

func TestAvailabilityNewCheckClearsPurchaseMessage(t *testing.T) {
	v := NewAvailabilityView()
	v.SetResult(&api.AvailabilityResult{Domain: "fresh.xyz", Available: true, Price: "2.04"})
	v.StartBuyConfirmation()
	v.SetPurchasing()
	v.SetPurchaseResult(&api.RegistrationResult{Domain: "fresh.xyz", OrderID: 1, CostCents: 204, BalanceCents: 100})

	v.SetResult(&api.AvailabilityResult{Domain: "next.com", Available: false})

	if strings.Contains(v.View(), "Registered") {
		t.Error("stale purchase success message shown after a new check")
	}
}

func TestAvailabilityHelpTextMentionsBuy(t *testing.T) {
	v := NewAvailabilityView()

	if !strings.Contains(v.HelpText(), "buy") {
		t.Error("help text does not document the buy key")
	}
}

func TestAvailabilityGetDomainTrimsWhitespace(t *testing.T) {
	v := NewAvailabilityView()
	v.input.SetValue("  example.com  ")

	if got := v.GetDomain(); got != "example.com" {
		t.Errorf("GetDomain() = %q, want example.com", got)
	}
}
