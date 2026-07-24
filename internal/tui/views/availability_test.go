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

	out := v.View()
	for _, want := range []string{"AVAILABLE", "TAKEN", "11.06", "(premium)", "taken.com", "open.com", "fancy.com"} {
		if !strings.Contains(out, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestAvailabilityViewRendersError(t *testing.T) {
	v := NewAvailabilityView()
	v.SetError(errors.New("rate limit exceeded"))

	if out := v.View(); !strings.Contains(out, "rate limit exceeded") {
		t.Error("View() does not render the error message")
	}
}

func TestAvailabilityGetDomainTrimsWhitespace(t *testing.T) {
	v := NewAvailabilityView()
	v.input.SetValue("  example.com  ")

	if got := v.GetDomain(); got != "example.com" {
		t.Errorf("GetDomain() = %q, want example.com", got)
	}
}
