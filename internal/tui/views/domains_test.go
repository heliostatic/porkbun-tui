package views

import (
	"testing"
	"time"

	"github.com/bc/porkbun-tui/internal/api"
)

func TestDomainsView_SetDomains(t *testing.T) {
	v := NewDomainsView()

	domains := []api.Domain{
		{Name: "zebra.com", ExpireDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "apple.com", ExpireDate: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "mango.com", ExpireDate: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)},
	}

	v.SetDomains(domains)

	// Default sort is by expiration ascending
	if v.filtered[0].Name != "apple.com" {
		t.Errorf("expected first domain 'apple.com' (earliest expiry), got '%s'", v.filtered[0].Name)
	}
	if v.filtered[2].Name != "mango.com" {
		t.Errorf("expected last domain 'mango.com' (latest expiry), got '%s'", v.filtered[2].Name)
	}
}

func TestDomainsView_SortByName(t *testing.T) {
	v := NewDomainsView()

	domains := []api.Domain{
		{Name: "zebra.com", ExpireDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "apple.com", ExpireDate: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "mango.com", ExpireDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
	}

	v.SetDomains(domains)

	// Change sort to name
	v.sortField = SortByName
	v.sortAscending = true
	v.sortDomains()

	if v.filtered[0].Name != "apple.com" {
		t.Errorf("expected first domain 'apple.com', got '%s'", v.filtered[0].Name)
	}
	if v.filtered[1].Name != "mango.com" {
		t.Errorf("expected second domain 'mango.com', got '%s'", v.filtered[1].Name)
	}
	if v.filtered[2].Name != "zebra.com" {
		t.Errorf("expected third domain 'zebra.com', got '%s'", v.filtered[2].Name)
	}
}

func TestDomainsView_SortDescending(t *testing.T) {
	v := NewDomainsView()

	domains := []api.Domain{
		{Name: "zebra.com", ExpireDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "apple.com", ExpireDate: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "mango.com", ExpireDate: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)},
	}

	v.SetDomains(domains)

	// Sort by expiration descending
	v.sortAscending = false
	v.sortDomains()

	if v.filtered[0].Name != "mango.com" {
		t.Errorf("expected first domain 'mango.com' (latest expiry), got '%s'", v.filtered[0].Name)
	}
	if v.filtered[2].Name != "apple.com" {
		t.Errorf("expected last domain 'apple.com' (earliest expiry), got '%s'", v.filtered[2].Name)
	}
}

func TestDomainsView_Filter(t *testing.T) {
	v := NewDomainsView()

	domains := []api.Domain{
		{Name: "example.com", ExpireDate: time.Now()},
		{Name: "test.io", ExpireDate: time.Now()},
		{Name: "example.org", ExpireDate: time.Now()},
		{Name: "mysite.com", ExpireDate: time.Now()},
	}

	v.SetDomains(domains)

	// Filter for "example"
	v.searchInput.SetValue("example")
	v.applyFilter()

	if len(v.filtered) != 2 {
		t.Fatalf("expected 2 filtered domains, got %d", len(v.filtered))
	}

	// Both should contain "example"
	for _, d := range v.filtered {
		if d.Name != "example.com" && d.Name != "example.org" {
			t.Errorf("unexpected domain in filter: %s", d.Name)
		}
	}
}

func TestDomainsView_FilterCaseInsensitive(t *testing.T) {
	v := NewDomainsView()

	domains := []api.Domain{
		{Name: "EXAMPLE.COM", ExpireDate: time.Now()},
		{Name: "test.io", ExpireDate: time.Now()},
	}

	v.SetDomains(domains)

	// Filter with lowercase
	v.searchInput.SetValue("example")
	v.applyFilter()

	if len(v.filtered) != 1 {
		t.Fatalf("expected 1 filtered domain, got %d", len(v.filtered))
	}
	if v.filtered[0].Name != "EXAMPLE.COM" {
		t.Errorf("expected 'EXAMPLE.COM', got '%s'", v.filtered[0].Name)
	}
}

func TestDomainsView_FilterNoMatch(t *testing.T) {
	v := NewDomainsView()

	domains := []api.Domain{
		{Name: "example.com", ExpireDate: time.Now()},
		{Name: "test.io", ExpireDate: time.Now()},
	}

	v.SetDomains(domains)

	// Filter for something that doesn't exist
	v.searchInput.SetValue("xyz123")
	v.applyFilter()

	if len(v.filtered) != 0 {
		t.Errorf("expected 0 filtered domains, got %d", len(v.filtered))
	}
}

func TestDomainsView_SelectedDomain(t *testing.T) {
	v := NewDomainsView()

	domains := []api.Domain{
		{Name: "first.com", ExpireDate: time.Now()},
		{Name: "second.com", ExpireDate: time.Now()},
	}

	v.SetDomains(domains)
	v.cursor = 1

	selected := v.SelectedDomain()
	if selected == nil {
		t.Fatal("expected selected domain, got nil")
	}
	if selected.Name != "second.com" {
		t.Errorf("expected 'second.com', got '%s'", selected.Name)
	}
}

func TestDomainsView_SelectedDomain_Empty(t *testing.T) {
	v := NewDomainsView()

	selected := v.SelectedDomain()
	if selected != nil {
		t.Errorf("expected nil for empty list, got %v", selected)
	}
}

func TestDomainsView_StatusText(t *testing.T) {
	v := NewDomainsView()

	domains := []api.Domain{
		{Name: "a.com", ExpireDate: time.Now()},
		{Name: "b.com", ExpireDate: time.Now()},
		{Name: "c.com", ExpireDate: time.Now()},
	}

	v.SetDomains(domains)

	status := v.StatusText()
	if status != "3 domains" {
		t.Errorf("expected '3 domains', got '%s'", status)
	}

	// With filter
	v.searchInput.SetValue("a")
	v.applyFilter()

	status = v.StatusText()
	if status != "1/3 domains" {
		t.Errorf("expected '1/3 domains', got '%s'", status)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, expected %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}
