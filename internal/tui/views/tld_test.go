package views

import (
	"testing"

	"github.com/bc/porkbun-tui/internal/api"
)

func TestTLDView_SetData_GroupsByTLD(t *testing.T) {
	v := NewTLDView()

	domains := []api.Domain{
		{Name: "a.com", TLD: "com"},
		{Name: "b.com", TLD: "com"},
		{Name: "c.io", TLD: "io"},
		{Name: "d.dev", TLD: "dev"},
	}

	pricing := map[string]api.TLDPricing{
		"com": {Renewal: "10.00"},
		"io":  {Renewal: "30.00"},
		"dev": {Renewal: "15.00"},
	}

	v.SetData(domains, pricing)

	// Should have 3 TLD groups
	if len(v.groups) != 3 {
		t.Fatalf("expected 3 TLD groups, got %d", len(v.groups))
	}
}

func TestTLDView_SetData_CalculatesTotalCost(t *testing.T) {
	v := NewTLDView()

	domains := []api.Domain{
		{Name: "a.com", TLD: "com"},
		{Name: "b.com", TLD: "com"},
		{Name: "c.com", TLD: "com"},
	}

	pricing := map[string]api.TLDPricing{
		"com": {Renewal: "10.00"},
	}

	v.SetData(domains, pricing)

	if len(v.groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(v.groups))
	}

	group := v.groups[0]
	if group.RenewalPrice != 10.00 {
		t.Errorf("expected renewal price 10.00, got %.2f", group.RenewalPrice)
	}
	if group.TotalCost != 30.00 {
		t.Errorf("expected total cost 30.00 (3 * 10), got %.2f", group.TotalCost)
	}
}

func TestTLDView_SetData_SortedByTotalCost(t *testing.T) {
	v := NewTLDView()

	domains := []api.Domain{
		{Name: "a.com", TLD: "com"},     // 1 domain * $10 = $10
		{Name: "b.io", TLD: "io"},       // 1 domain * $30 = $30
		{Name: "c.io", TLD: "io"},       // total io = $60
		{Name: "d.dev", TLD: "dev"},     // 1 domain * $15 = $15
	}

	pricing := map[string]api.TLDPricing{
		"com": {Renewal: "10.00"},
		"io":  {Renewal: "30.00"},
		"dev": {Renewal: "15.00"},
	}

	v.SetData(domains, pricing)

	// Should be sorted by total cost descending
	// io ($60), dev ($15), com ($10)
	if v.groups[0].TLD != "io" {
		t.Errorf("first group should be 'io' (highest cost), got '%s'", v.groups[0].TLD)
	}
	if v.groups[1].TLD != "dev" {
		t.Errorf("second group should be 'dev', got '%s'", v.groups[1].TLD)
	}
	if v.groups[2].TLD != "com" {
		t.Errorf("third group should be 'com' (lowest cost), got '%s'", v.groups[2].TLD)
	}
}

func TestTLDView_SetData_MissingPricing(t *testing.T) {
	v := NewTLDView()

	domains := []api.Domain{
		{Name: "a.com", TLD: "com"},
		{Name: "b.xyz", TLD: "xyz"}, // No pricing for xyz
	}

	pricing := map[string]api.TLDPricing{
		"com": {Renewal: "10.00"},
		// xyz is missing
	}

	v.SetData(domains, pricing)

	if len(v.groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(v.groups))
	}

	// Find the xyz group
	var xyzGroup *TLDGroup
	for i := range v.groups {
		if v.groups[i].TLD == "xyz" {
			xyzGroup = &v.groups[i]
			break
		}
	}

	if xyzGroup == nil {
		t.Fatal("xyz group not found")
	}

	// Should have zero pricing
	if xyzGroup.RenewalPrice != 0 {
		t.Errorf("expected xyz renewal price 0, got %.2f", xyzGroup.RenewalPrice)
	}
	if xyzGroup.TotalCost != 0 {
		t.Errorf("expected xyz total cost 0, got %.2f", xyzGroup.TotalCost)
	}
}

func TestTLDView_SetData_NilPricing(t *testing.T) {
	v := NewTLDView()

	domains := []api.Domain{
		{Name: "a.com", TLD: "com"},
	}

	// Nil pricing map
	v.SetData(domains, nil)

	if len(v.groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(v.groups))
	}

	if v.groups[0].RenewalPrice != 0 {
		t.Errorf("expected renewal price 0 with nil pricing, got %.2f", v.groups[0].RenewalPrice)
	}
}

func TestTLDView_SetData_Empty(t *testing.T) {
	v := NewTLDView()

	v.SetData([]api.Domain{}, nil)

	if len(v.groups) != 0 {
		t.Errorf("expected 0 groups for empty domains, got %d", len(v.groups))
	}
}

func TestTLDView_GetLineForCursor(t *testing.T) {
	v := NewTLDView()

	domains := []api.Domain{
		{Name: "a.com", TLD: "com"},
		{Name: "b.com", TLD: "com"},
		{Name: "c.io", TLD: "io"},
	}

	v.SetData(domains, nil)

	// Expand first group
	v.expanded["com"] = true

	// Cursor at 0 (com header) -> line 0
	v.cursor = 0
	if line := v.getLineForCursor(); line != 0 {
		t.Errorf("cursor 0 should be at line 0, got %d", line)
	}

	// Cursor at 1 (io header) -> line 3 (com header + 2 domains)
	v.cursor = 1
	if line := v.getLineForCursor(); line != 3 {
		t.Errorf("cursor 1 should be at line 3, got %d", line)
	}
}

func TestTLDView_GetTotalLines(t *testing.T) {
	v := NewTLDView()

	domains := []api.Domain{
		{Name: "a.com", TLD: "com"},
		{Name: "b.com", TLD: "com"},
		{Name: "c.io", TLD: "io"},
	}

	v.SetData(domains, nil)

	// Both collapsed: 2 lines (one per TLD header)
	totalLines := v.getTotalLines()
	if totalLines != 2 {
		t.Errorf("expected 2 total lines (collapsed), got %d", totalLines)
	}

	// Expand com
	v.expanded["com"] = true
	totalLines = v.getTotalLines()
	// 1 (com header) + 2 (com domains) + 1 (io header) = 4
	if totalLines != 4 {
		t.Errorf("expected 4 total lines with com expanded, got %d", totalLines)
	}
}

func TestTLDView_StatusText(t *testing.T) {
	v := NewTLDView()

	domains := []api.Domain{
		{Name: "a.com", TLD: "com"},
		{Name: "b.io", TLD: "io"},
		{Name: "c.io", TLD: "io"},
	}

	v.SetData(domains, nil)

	status := v.StatusText()
	if status != "2 TLDs" {
		t.Errorf("expected '2 TLDs', got '%s'", status)
	}
}

func TestTLDView_ExpandedMap(t *testing.T) {
	v := NewTLDView()

	domains := []api.Domain{
		{Name: "a.com", TLD: "com"},
		{Name: "b.io", TLD: "io"},
	}

	v.SetData(domains, nil)

	// Nothing should be expanded initially
	if v.expanded["com"] {
		t.Error("com should not be expanded initially")
	}
	if v.expanded["io"] {
		t.Error("io should not be expanded initially")
	}

	// Expand com
	v.expanded["com"] = true
	if !v.expanded["com"] {
		t.Error("com should be expanded after setting")
	}
}
