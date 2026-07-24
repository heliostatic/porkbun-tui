# Fix Identified Issues Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the real Porkbun domain-availability check (currently a stub), add tests for the untested TUI update loop and availability view, and document config-file permissions.

**Architecture:** The availability check becomes a direct HTTP POST to Porkbun's `/domain/checkDomain/{domain}` endpoint (the `porkbun-go` v1.0.2 SDK does not expose it), implemented inside the existing `internal/api.Client` with an overridable base URL so tests can use `httptest`. TUI tests live in the same packages as the code (internal white-box tests), constructing Bubble Tea messages directly and asserting on model state and rendered output.

**Tech Stack:** Go 1.21+, Bubble Tea/Bubbles/Lipgloss, stdlib `net/http` + `net/http/httptest`. No new dependencies.

## Global Constraints

- No new module dependencies; use the Go standard library only.
- All code gofmt-formatted; `go build ./...`, `go vet ./...`, and `go test ./...` must pass after every task.
- Commit messages follow existing history style: imperative mood, no prefix (e.g. "Add tests for app update loop"), ending with the line `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`.
- Porkbun API contract (verified against Porkbun docs and multiple client implementations):
  - `POST https://api.porkbun.com/api/json/v3/domain/checkDomain/{domain}`
  - Request body: `{"apikey": "...", "secretapikey": "..."}`
  - Success response: `{"status":"SUCCESS","response":{"avail":"yes"|"no","price":"11.06","premium":"yes"|"no", ...}}`
  - Error response: `{"status":"ERROR","message":"..."}`
  - Rate limit: 1 check per 10 seconds per API key.

---

### Task 1: Real availability check in the API client

**Files:**
- Modify: `internal/api/client.go` (struct + `NewClient` + replace `CheckAvailability` stub at lines 142–152)
- Test: `internal/api/client_test.go` (create)

**Interfaces:**
- Consumes: existing `config.Config{APIKey, SecretKey}`, existing `AvailabilityResult` struct (unchanged).
- Produces: `func (c *Client) CheckAvailability(ctx context.Context, domain string) (*AvailabilityResult, error)` — same signature as today, so `internal/tui/app.go:193` needs no changes. `Client` gains unexported fields `apiKey`, `secretKey`, `baseURL`, `httpClient` (tests in package `api` set them directly).

- [ ] **Step 1: Write the failing tests**

Create `internal/api/client_test.go`:

```go
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(serverURL string) *Client {
	return &Client{
		apiKey:     "pk1_test",
		secretKey:  "sk1_test",
		baseURL:    serverURL,
		httpClient: &http.Client{},
	}
}

func TestCheckAvailabilityAvailable(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		w.Write([]byte(`{"status":"SUCCESS","response":{"avail":"yes","type":"registration","price":"11.06","premium":"no"}}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	result, err := c.CheckAvailability(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("CheckAvailability returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/domain/checkDomain/example.com" {
		t.Errorf("path = %q, want /domain/checkDomain/example.com", gotPath)
	}
	if gotBody["apikey"] != "pk1_test" || gotBody["secretapikey"] != "sk1_test" {
		t.Errorf("request body missing credentials: %v", gotBody)
	}
	if !result.Available {
		t.Error("result.Available = false, want true")
	}
	if result.Domain != "example.com" {
		t.Errorf("result.Domain = %q, want example.com", result.Domain)
	}
	if result.Price != "11.06" {
		t.Errorf("result.Price = %q, want 11.06", result.Price)
	}
	if result.Premium {
		t.Error("result.Premium = true, want false")
	}
}

func TestCheckAvailabilityTakenAndPremium(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"SUCCESS","response":{"avail":"no","price":"1200.00","premium":"yes"}}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	result, err := c.CheckAvailability(context.Background(), "taken.com")
	if err != nil {
		t.Fatalf("CheckAvailability returned error: %v", err)
	}
	if result.Available {
		t.Error("result.Available = true, want false")
	}
	if !result.Premium {
		t.Error("result.Premium = false, want true")
	}
}

func TestCheckAvailabilityAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"status":"ERROR","message":"Invalid API key. (002)"}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	_, err := c.CheckAvailability(context.Background(), "example.com")
	if err == nil {
		t.Fatal("CheckAvailability returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("error %q does not contain API message", err)
	}
}

func TestCheckAvailabilityMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	_, err := c.CheckAvailability(context.Background(), "example.com")
	if err == nil {
		t.Fatal("CheckAvailability returned nil error, want error")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/api/ -run TestCheckAvailability -v`
Expected: compile FAILURE — `Client` has no fields `apiKey`, `baseURL`, `httpClient`.

- [ ] **Step 3: Implement the client changes**

In `internal/api/client.go`, update imports:

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bc/porkbun-tui/internal/config"
	"github.com/tuzzmaniandevil/porkbun-go"
)
```

Replace the `Client` struct and `NewClient`:

```go
const defaultBaseURL = "https://api.porkbun.com/api/json/v3"

type Client struct {
	pb         *porkbun.Client
	apiKey     string
	secretKey  string
	baseURL    string
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	pb := porkbun.NewClient(&porkbun.Options{
		ApiKey:       cfg.APIKey,
		SecretApiKey: cfg.SecretKey,
	})
	return &Client{
		pb:         pb,
		apiKey:     cfg.APIKey,
		secretKey:  cfg.SecretKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}
```

Replace the `CheckAvailability` stub (and its placeholder comment) with:

```go
type checkDomainResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Response struct {
		Avail   string `json:"avail"`
		Price   string `json:"price"`
		Premium string `json:"premium"`
	} `json:"response"`
}

// CheckAvailability calls Porkbun's checkDomain endpoint directly because the
// porkbun-go SDK (v1.0.2) does not expose it. Porkbun rate-limits this
// endpoint to one check per 10 seconds.
func (c *Client) CheckAvailability(ctx context.Context, domain string) (*AvailabilityResult, error) {
	body, err := json.Marshal(map[string]string{
		"apikey":       c.apiKey,
		"secretapikey": c.secretKey,
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/domain/checkDomain/%s", c.baseURL, domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsed checkDomainResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decoding checkDomain response: %w", err)
	}
	if parsed.Status != "SUCCESS" {
		if parsed.Message != "" {
			return nil, fmt.Errorf("porkbun: %s", parsed.Message)
		}
		return nil, fmt.Errorf("porkbun: checkDomain failed (HTTP %d)", resp.StatusCode)
	}

	return &AvailabilityResult{
		Domain:    domain,
		Available: parsed.Response.Avail == "yes",
		Price:     parsed.Response.Price,
		Premium:   parsed.Response.Premium == "yes",
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/api/ -v && go build ./... && go vet ./...`
Expected: all 4 tests PASS, clean build and vet.

- [ ] **Step 5: Commit**

```bash
git add internal/api/client.go internal/api/client_test.go
git commit -m "Implement domain availability check via Porkbun checkDomain API"
```

---

### Task 2: Availability view tests

**Files:**
- Test: `internal/tui/views/availability_test.go` (create)

**Interfaces:**
- Consumes: existing `views.AvailabilityView` API (`NewAvailabilityView`, `SetResult`, `SetError`, `SetLoading`, `IsLoading`, `View`) and `api.AvailabilityResult`. No production code changes.
- Produces: nothing consumed by later tasks.

- [ ] **Step 1: Write the tests**

Create `internal/tui/views/availability_test.go`:

```go
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
```

- [ ] **Step 2: Run the tests**

Run: `go test ./internal/tui/views/ -run TestAvailability -v`
Expected: all 5 tests PASS (these test existing behavior; a failure means the test found a real bug — investigate before changing the test).

- [ ] **Step 3: Commit**

```bash
git add internal/tui/views/availability_test.go
git commit -m "Add availability view tests"
```

---

### Task 3: App update-loop tests

**Files:**
- Test: `internal/tui/app_test.go` (create)

**Interfaces:**
- Consumes: existing `tui.NewApp`, `App.Update`, message types (`domainsLoadedMsg`, `availabilityResultMsg`, `pricingLoadedMsg`, `errMsg`), `View` constants, `api.Domain`, `api.TLDPricing`. `NewApp(nil, nil, nil, nil, false)` is valid — client/cache are only dereferenced inside `tea.Cmd` closures and cache is nil-checked.
- Produces: nothing consumed by later tasks.

- [ ] **Step 1: Write the tests**

Create `internal/tui/app_test.go`:

```go
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
```

- [ ] **Step 2: Run the tests**

Run: `go test ./internal/tui/ -v`
Expected: all tests PASS. As in Task 2, these pin existing behavior — a failure means a real bug or a wrong assumption about the API; investigate rather than force-pass. (Note: `demo` mode with `r` should be a no-op per `app.go:381`; if any assertion fails on details like this, check the source line cited and fix the test to match actual intended behavior.)

- [ ] **Step 3: Verify coverage improved**

Run: `go test -cover ./internal/tui/`
Expected: coverage well above 0% (roughly 25–40%; the exact number doesn't matter, just that the update loop is now exercised).

- [ ] **Step 4: Commit**

```bash
git add internal/tui/app_test.go
git commit -m "Add app update-loop tests"
```

---

### Task 4: Documentation fixes

**Files:**
- Modify: `README.md` (Config File section after line 66; Features list availability bullet)

**Interfaces:** none — docs only.

- [ ] **Step 1: Add chmod recommendation to the Config File section**

In `README.md`, change the Config File section to:

```markdown
### Config File

Create `~/.config/porkbun-tui/config.yaml`:

```yaml
api_key: pk1_xxx
secret_key: sk1_xxx
```

Since this file contains your API credentials, restrict its permissions:

```bash
chmod 600 ~/.config/porkbun-tui/config.yaml
```
```

- [ ] **Step 2: Note the rate limit on the availability feature**

In the Features list, change:

```markdown
- **Domain Availability** - Check if a domain is available for registration
```

to:

```markdown
- **Domain Availability** - Check if a domain is available for registration, with pricing (Porkbun rate-limits checks to one per 10 seconds)
```

- [ ] **Step 3: Verify and commit**

Run: `go build ./... && go test ./...`
Expected: everything still passes (docs-only change; this is the final gate).

```bash
git add README.md
git commit -m "Document config file permissions and availability rate limit"
```
