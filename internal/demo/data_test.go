package demo

import "testing"

func TestCheckAvailabilityIsDeterministic(t *testing.T) {
	a := CheckAvailability("coolstartup.dev")
	b := CheckAvailability("coolstartup.dev")
	if *a != *b {
		t.Errorf("same input gave different results: %+v vs %+v", a, b)
	}
}

func TestCheckAvailabilityFamousNamesAreTaken(t *testing.T) {
	for _, domain := range []string{"google.com", "porkbun.com", "github.io"} {
		r := CheckAvailability(domain)
		if r.Available {
			t.Errorf("CheckAvailability(%q).Available = true, want taken", domain)
		}
		if r.Price != "" {
			t.Errorf("taken domain %q has price %q, want empty", domain, r.Price)
		}
	}
}

func TestCheckAvailabilityLongNamesAreAvailableWithTLDPricing(t *testing.T) {
	r := CheckAvailability("coolstartup.dev")
	if !r.Available {
		t.Fatal("coolstartup.dev should be available in demo data")
	}
	if want := Pricing()["dev"].Registration; r.Price != want {
		t.Errorf("price = %q, want %q from the demo pricing table", r.Price, want)
	}
}

func TestCheckAvailabilityUnknownTLDGetsFallbackPrice(t *testing.T) {
	r := CheckAvailability("somethinglong.pizza")
	if !r.Available {
		t.Fatal("somethinglong.pizza should be available in demo data")
	}
	if r.Price == "" {
		t.Error("available domain has no price")
	}
}
