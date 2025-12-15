package views

import (
	"testing"
	"time"

	"github.com/bc/porkbun-tui/internal/api"
)

func TestCalendarView_SetDomains_GroupsByMonth(t *testing.T) {
	v := NewCalendarView()

	domains := []api.Domain{
		{Name: "jan1.com", ExpireDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Name: "jan2.com", ExpireDate: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)},
		{Name: "mar1.com", ExpireDate: time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC)},
		{Name: "jun1.com", ExpireDate: time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)},
	}

	v.SetDomains(domains)

	// Should have 3 groups (Jan, Mar, Jun)
	if len(v.groups) != 3 {
		t.Fatalf("expected 3 month groups, got %d", len(v.groups))
	}

	// First group should be January with 2 domains
	if v.groups[0].Month != time.January {
		t.Errorf("expected first group to be January, got %s", v.groups[0].Month)
	}
	if len(v.groups[0].Domains) != 2 {
		t.Errorf("expected 2 domains in January, got %d", len(v.groups[0].Domains))
	}

	// Second group should be March
	if v.groups[1].Month != time.March {
		t.Errorf("expected second group to be March, got %s", v.groups[1].Month)
	}

	// Third group should be June
	if v.groups[2].Month != time.June {
		t.Errorf("expected third group to be June, got %s", v.groups[2].Month)
	}
}

func TestCalendarView_SetDomains_SortedByDate(t *testing.T) {
	v := NewCalendarView()

	// Domains out of order
	domains := []api.Domain{
		{Name: "dec.com", ExpireDate: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "jan.com", ExpireDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "jun.com", ExpireDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
	}

	v.SetDomains(domains)

	// Groups should be sorted chronologically
	if v.groups[0].Month != time.January {
		t.Errorf("first group should be January, got %s", v.groups[0].Month)
	}
	if v.groups[1].Month != time.June {
		t.Errorf("second group should be June, got %s", v.groups[1].Month)
	}
	if v.groups[2].Month != time.December {
		t.Errorf("third group should be December, got %s", v.groups[2].Month)
	}
}

func TestCalendarView_SetDomains_DomainsWithinGroupSorted(t *testing.T) {
	v := NewCalendarView()

	domains := []api.Domain{
		{Name: "late.com", ExpireDate: time.Date(2025, 1, 25, 0, 0, 0, 0, time.UTC)},
		{Name: "early.com", ExpireDate: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)},
		{Name: "mid.com", ExpireDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
	}

	v.SetDomains(domains)

	if len(v.groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(v.groups))
	}

	// Domains within the group should be sorted by date
	if v.groups[0].Domains[0].Name != "early.com" {
		t.Errorf("first domain should be 'early.com', got '%s'", v.groups[0].Domains[0].Name)
	}
	if v.groups[0].Domains[1].Name != "mid.com" {
		t.Errorf("second domain should be 'mid.com', got '%s'", v.groups[0].Domains[1].Name)
	}
	if v.groups[0].Domains[2].Name != "late.com" {
		t.Errorf("third domain should be 'late.com', got '%s'", v.groups[0].Domains[2].Name)
	}
}

func TestCalendarView_SetDomains_FirstGroupExpanded(t *testing.T) {
	v := NewCalendarView()

	domains := []api.Domain{
		{Name: "a.com", ExpireDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "b.com", ExpireDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
	}

	v.SetDomains(domains)

	// First group should be expanded by default
	if !v.groups[0].Expanded {
		t.Error("first group should be expanded by default")
	}
	// Other groups should not be expanded
	if v.groups[1].Expanded {
		t.Error("second group should not be expanded by default")
	}
}

func TestCalendarView_SetDomains_DifferentYears(t *testing.T) {
	v := NewCalendarView()

	domains := []api.Domain{
		{Name: "2025.com", ExpireDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "2026.com", ExpireDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
	}

	v.SetDomains(domains)

	// Should have 2 groups (same month, different years)
	if len(v.groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(v.groups))
	}

	// 2025 should come first
	if v.groups[0].Year != 2025 {
		t.Errorf("first group should be 2025, got %d", v.groups[0].Year)
	}
	if v.groups[1].Year != 2026 {
		t.Errorf("second group should be 2026, got %d", v.groups[1].Year)
	}
}

func TestCalendarView_SetDomains_Empty(t *testing.T) {
	v := NewCalendarView()

	v.SetDomains([]api.Domain{})

	if len(v.groups) != 0 {
		t.Errorf("expected 0 groups for empty domains, got %d", len(v.groups))
	}
}

func TestCalendarView_GetLineForCursor(t *testing.T) {
	v := NewCalendarView()

	domains := []api.Domain{
		{Name: "jan1.com", ExpireDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "jan2.com", ExpireDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Name: "feb1.com", ExpireDate: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	v.SetDomains(domains)
	// First group (January) is expanded by default with 2 domains

	// Cursor at 0 (January header) -> line 0
	v.cursor = 0
	if line := v.getLineForCursor(); line != 0 {
		t.Errorf("cursor 0 should be at line 0, got %d", line)
	}

	// Cursor at 1 (February header) -> line 3 (January header + 2 domains)
	v.cursor = 1
	if line := v.getLineForCursor(); line != 3 {
		t.Errorf("cursor 1 should be at line 3, got %d", line)
	}
}

func TestCalendarView_GetTotalLines(t *testing.T) {
	v := NewCalendarView()

	domains := []api.Domain{
		{Name: "jan1.com", ExpireDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "jan2.com", ExpireDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Name: "feb1.com", ExpireDate: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	v.SetDomains(domains)
	// January expanded (1 header + 2 domains), February collapsed (1 header)

	totalLines := v.getTotalLines()
	// 1 (Jan header) + 2 (Jan domains) + 1 (Feb header) = 4
	if totalLines != 4 {
		t.Errorf("expected 4 total lines, got %d", totalLines)
	}

	// Expand February
	v.groups[1].Expanded = true
	totalLines = v.getTotalLines()
	// 1 + 2 + 1 + 1 = 5
	if totalLines != 5 {
		t.Errorf("expected 5 total lines with Feb expanded, got %d", totalLines)
	}
}

func TestCalendarView_StatusText(t *testing.T) {
	v := NewCalendarView()

	domains := []api.Domain{
		{Name: "a.com", ExpireDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "b.com", ExpireDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Name: "c.com", ExpireDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
	}

	v.SetDomains(domains)

	status := v.StatusText()
	if status != "2 months, 3 domains" {
		t.Errorf("expected '2 months, 3 domains', got '%s'", status)
	}
}
