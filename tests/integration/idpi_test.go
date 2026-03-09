//go:build integration

package integration

// IDPI integration tests — indirect prompt injection defence on /find and /pdf.
//
// Each test group spins up an isolated pinchtab server configured with specific
// IDPI settings via testutil.NewTestServer. Headed mode (Headless: false) is
// intentional: the browser is visible so the effect of each defence can be
// observed during development and code review.
//
// Test naming follows the pattern:
//
//	TestIDPI_<Endpoint>_<Mode>[_<Scenario>]
//
// Scenarios:
//   - CleanPage_NoSignal  — benign page; expect no IDPI headers or blocks
//   - InjectedPage_Warns  — page with injection phrase; warn mode emits header
//   - InjectedPage_Blocked — page with injection phrase; strict mode returns 403
//   - CustomPattern_Warns — user-defined custom pattern triggers warning
//   - Disabled_Passthrough — IDPI disabled; injection page passes transparently

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	appconfig "github.com/pinchtab/pinchtab/internal/config"
	"github.com/pinchtab/pinchtab/tests/testutil"
)

// idpiCfg returns an IDPIConfig preset. Callers can further override fields.
func idpiCfg(enabled, scan, strict bool, custom ...string) appconfig.IDPIConfig {
	return appconfig.IDPIConfig{
		Enabled:        enabled,
		ScanContent:    scan,
		StrictMode:     strict,
		CustomPatterns: custom,
	}
}

// newIDPIServer builds a ServerConfig for IDPI integration tests and starts a
// server using testutil.NewTestServer. Cleanup is registered automatically via
// t.Cleanup. Headed mode is always active because the user wants to monitor
// test runs in a visible browser.
func newIDPIServer(t *testing.T, cfg appconfig.IDPIConfig) *testutil.Client {
	t.Helper()
	port, err := testutil.GetFreePort()
	if err != nil {
		t.Fatalf("idpi: get free port: %v", err)
	}
	sc := testutil.DefaultConfig()
	sc.Port = port
	sc.Headless = false // headed mode: visible Chrome for development monitoring
	sc.IDPI = cfg
	srv := testutil.NewTestServer(t, sc)
	return testutil.NewClient(srv.URL)
}

// idpiNavigate navigates the given client to a URL and returns the resolved tab ID.
// The tab is closed automatically via t.Cleanup.
func idpiNavigate(t *testing.T, c *testutil.Client, url string) string {
	t.Helper()
	code, body := c.PostWithRetry(t, "/navigate", map[string]any{"url": url}, 2)
	if code != 200 {
		t.Fatalf("navigate to %s: status %d: %s", url, code, body)
	}
	tabID := testutil.JSONField(t, body, "tabId")
	if tabID == "" {
		t.Fatal("navigate response is missing tabId")
	}
	t.Cleanup(func() {
		c.Post(t, "/tab", map[string]any{"tabId": tabID, "action": "close"}) //nolint:errcheck
	})
	return tabID
}

// idpiFind calls POST /find for a specific tab and query, returning status,
// body, and full response headers for IDPI signal inspection.
func idpiFind(t *testing.T, c *testutil.Client, tabID, query string) (int, []byte, http.Header) {
	t.Helper()
	return c.PostWithHeaders(t, fmt.Sprintf("/tabs/%s/find", tabID), map[string]any{
		"query":     query,
		"threshold": 0.1,
		"topK":      5,
	})
}

// idpiPDF calls GET /tabs/{id}/pdf and returns status, body, and headers.
func idpiPDF(t *testing.T, c *testutil.Client, tabID string) (int, []byte, http.Header) {
	t.Helper()
	return c.GetWithHeaders(t, fmt.Sprintf("/tabs/%s/pdf", tabID))
}

// --- /find tests ---

// TestIDPI_Find_WarnMode verifies that /find emits X-IDPI-Warning headers and
// populates idpiWarning in the JSON body when injection content is detected in
// the AX tree, without blocking the request (StrictMode: false).
func TestIDPI_Find_WarnMode(t *testing.T) {
	c := newIDPIServer(t, idpiCfg(true, true, false))

	t.Run("CleanPage_NoSignal", func(t *testing.T) {
		tabID := idpiNavigate(t, c, idpiCleanPageURL(t))
		code, body, hdr := idpiFind(t, c, tabID, "safe action button")
		if code != 200 {
			t.Fatalf("expected 200, got %d: %s", code, body)
		}
		if w := hdr.Get("X-IDPI-Warning"); w != "" {
			t.Errorf("expected no X-IDPI-Warning on clean page, got: %s", w)
		}
		if w := jsonFieldRaw(t, body, "idpiWarning"); w != "" && w != "null" {
			t.Errorf("expected no idpiWarning in response, got: %s", w)
		}
	})

	t.Run("InjectedPage_Warns", func(t *testing.T) {
		tabID := idpiNavigate(t, c, idpiInjectPageURL(t))
		code, body, hdr := idpiFind(t, c, tabID, "continue button")
		if code != 200 {
			t.Fatalf("expected 200 in warn mode, got %d: %s", code, body)
		}
		warning := hdr.Get("X-IDPI-Warning")
		if warning == "" {
			t.Error("expected X-IDPI-Warning header on injection page, got none")
		}
		if !strings.Contains(strings.ToLower(warning), "injection") {
			t.Errorf("expected warning to mention injection, got: %s", warning)
		}
		// idpiWarning field should match the header value
		if jw := testutil.JSONField(t, body, "idpiWarning"); jw != warning {
			t.Errorf("idpiWarning field %q does not match X-IDPI-Warning header %q", jw, warning)
		}
		// X-IDPI-Pattern should identify the matched phrase
		if p := hdr.Get("X-IDPI-Pattern"); p == "" {
			t.Error("expected X-IDPI-Pattern header, got none")
		}
	})

	t.Run("InjectedPage_TabEndpoint_Warns", func(t *testing.T) {
		// Ensure /tabs/{id}/find and POST /find are both covered.
		tabID := idpiNavigate(t, c, idpiInjectPageURL(t))
		code, body, hdr := c.PostWithHeaders(t, "/find", map[string]any{
			"query":     "malicious paragraph",
			"tabId":     tabID,
			"threshold": 0.1,
		})
		if code != 200 {
			t.Fatalf("POST /find: expected 200 in warn mode, got %d: %s", code, body)
		}
		if w := hdr.Get("X-IDPI-Warning"); w == "" {
			t.Error("POST /find: expected X-IDPI-Warning on injection page, got none")
		}
	})

	t.Run("CustomPattern_Warns", func(t *testing.T) {
		// Custom pattern "reveal your system prompt" is not a built-in; the
		// fixture page contains it so this verifies the custom-pattern path.
		cCustom := newIDPIServer(t, appconfig.IDPIConfig{
			Enabled:        true,
			ScanContent:    true,
			StrictMode:     false,
			CustomPatterns: []string{"reveal your system prompt"},
		})
		tabID := idpiNavigate(t, cCustom, idpiInjectPageURL(t))
		_, _, hdr := idpiFind(t, cCustom, tabID, "action button")
		if w := hdr.Get("X-IDPI-Warning"); w == "" {
			t.Error("expected X-IDPI-Warning from custom pattern, got none")
		}
	})

	t.Run("Disabled_Passthrough", func(t *testing.T) {
		cOff := newIDPIServer(t, idpiCfg(false, false, false))
		tabID := idpiNavigate(t, cOff, idpiInjectPageURL(t))
		code, body, hdr := idpiFind(t, cOff, tabID, "continue button")
		if code != 200 {
			t.Fatalf("expected 200 with IDPI disabled, got %d: %s", code, body)
		}
		if w := hdr.Get("X-IDPI-Warning"); w != "" {
			t.Errorf("expected no X-IDPI-Warning when IDPI disabled, got: %s", w)
		}
	})
}

// TestIDPI_Find_StrictMode verifies that /find returns HTTP 403 when injection
// content is detected and StrictMode is enabled.
func TestIDPI_Find_StrictMode(t *testing.T) {
	c := newIDPIServer(t, idpiCfg(true, true, true))

	t.Run("CleanPage_Unaffected", func(t *testing.T) {
		tabID := idpiNavigate(t, c, idpiCleanPageURL(t))
		code, body, _ := idpiFind(t, c, tabID, "safe action button")
		if code != 200 {
			t.Fatalf("strict mode: expected 200 for clean page, got %d: %s", code, body)
		}
	})

	t.Run("InjectedPage_Blocked", func(t *testing.T) {
		tabID := idpiNavigate(t, c, idpiInjectPageURL(t))
		code, body, _ := idpiFind(t, c, tabID, "continue button")
		if code != http.StatusForbidden {
			t.Fatalf("strict mode: expected 403 for injection page, got %d: %s", code, body)
		}
		if !strings.Contains(strings.ToLower(string(body)), "idpi") {
			t.Errorf("expected idpi mention in 403 body, got: %s", body)
		}
	})

	t.Run("ScanDisabled_NoBlock", func(t *testing.T) {
		// StrictMode=true but ScanContent=false: no content scan → no block.
		cNoScan := newIDPIServer(t, appconfig.IDPIConfig{
			Enabled:     true,
			ScanContent: false,
			StrictMode:  true,
		})
		tabID := idpiNavigate(t, cNoScan, idpiInjectPageURL(t))
		code, _, _ := idpiFind(t, cNoScan, tabID, "continue button")
		if code == http.StatusForbidden {
			t.Error("expected no block when ScanContent=false even in strict mode")
		}
	})
}

// --- /pdf tests ---

// TestIDPI_PDF_WarnMode verifies that /pdf emits X-IDPI-Warning header when
// injection content is detected in page text, without blocking PDF generation.
func TestIDPI_PDF_WarnMode(t *testing.T) {
	c := newIDPIServer(t, idpiCfg(true, true, false))

	t.Run("CleanPage_NoSignal", func(t *testing.T) {
		tabID := idpiNavigate(t, c, idpiCleanPageURL(t))
		code, body, hdr := idpiPDF(t, c, tabID)
		if code != 200 {
			t.Fatalf("expected 200, got %d: %s", code, body)
		}
		if w := hdr.Get("X-IDPI-Warning"); w != "" {
			t.Errorf("expected no X-IDPI-Warning on clean page, got: %s", w)
		}
		assertPDFBody(t, body)
	})

	t.Run("InjectedPage_Warns", func(t *testing.T) {
		tabID := idpiNavigate(t, c, idpiInjectPageURL(t))
		code, body, hdr := idpiPDF(t, c, tabID)
		if code != 200 {
			t.Fatalf("expected 200 in warn mode, got %d: %s", code, body)
		}
		if w := hdr.Get("X-IDPI-Warning"); w == "" {
			t.Error("expected X-IDPI-Warning header on injection page PDF, got none")
		}
		// PDF should still be returned (not blocked in warn mode)
		assertPDFBody(t, body)
	})

	t.Run("InjectedPage_TabPDFEndpoint_Warns", func(t *testing.T) {
		// POST /tabs/{id}/pdf should also trigger the IDPI scan.
		tabID := idpiNavigate(t, c, idpiInjectPageURL(t))
		code, body, hdr := c.PostWithHeaders(t, fmt.Sprintf("/tabs/%s/pdf", tabID), map[string]any{})
		if code != 200 {
			t.Fatalf("POST /tabs/{id}/pdf: expected 200 in warn mode, got %d: %s", code, body)
		}
		if w := hdr.Get("X-IDPI-Warning"); w == "" {
			t.Error("POST /tabs/{id}/pdf: expected X-IDPI-Warning on injection page, got none")
		}
	})

	t.Run("Disabled_Passthrough", func(t *testing.T) {
		cOff := newIDPIServer(t, idpiCfg(false, false, false))
		tabID := idpiNavigate(t, cOff, idpiInjectPageURL(t))
		code, body, hdr := idpiPDF(t, cOff, tabID)
		if code != 200 {
			t.Fatalf("expected 200 with IDPI disabled, got %d: %s", code, body)
		}
		if w := hdr.Get("X-IDPI-Warning"); w != "" {
			t.Errorf("expected no X-IDPI-Warning when IDPI disabled, got: %s", w)
		}
		assertPDFBody(t, body)
	})
}

// TestIDPI_PDF_StrictMode verifies that /pdf returns HTTP 403 and no PDF bytes
// when injection content is detected and StrictMode is enabled.
func TestIDPI_PDF_StrictMode(t *testing.T) {
	c := newIDPIServer(t, idpiCfg(true, true, true))

	t.Run("CleanPage_Unaffected", func(t *testing.T) {
		tabID := idpiNavigate(t, c, idpiCleanPageURL(t))
		code, body, _ := idpiPDF(t, c, tabID)
		if code != 200 {
			t.Fatalf("strict mode: expected 200 for clean page, got %d: %s", code, body)
		}
		assertPDFBody(t, body)
	})

	t.Run("InjectedPage_Blocked", func(t *testing.T) {
		tabID := idpiNavigate(t, c, idpiInjectPageURL(t))
		code, body, _ := idpiPDF(t, c, tabID)
		if code != http.StatusForbidden {
			t.Fatalf("strict mode: expected 403 for injection page PDF, got %d: %s", code, body)
		}
		if !strings.Contains(strings.ToLower(string(body)), "idpi") {
			t.Errorf("expected idpi mention in 403 body, got: %s", body)
		}
	})

	t.Run("TabPDF_InjectedPage_Blocked", func(t *testing.T) {
		// POST /tabs/{id}/pdf delegates to HandlePDF, so the same scan applies.
		tabID := idpiNavigate(t, c, idpiInjectPageURL(t))
		code, _, _ := c.PostWithHeaders(t, fmt.Sprintf("/tabs/%s/pdf", tabID), map[string]any{})
		if code != http.StatusForbidden {
			t.Fatalf("POST /tabs/{id}/pdf strict mode: expected 403, got %d", code)
		}
	})
}

// --- edge-case / integration tests ---

// TestIDPI_Find_MultipleInjectionPhrases ensures only the first matching phrase
// is reported and subsequent matches don't cause duplicate headers (warn mode).
func TestIDPI_Find_MultipleInjectionPhrases(t *testing.T) {
	c := newIDPIServer(t, idpiCfg(true, true, false))
	tabID := idpiNavigate(t, c, idpiInjectPageURL(t))
	_, _, hdr := idpiFind(t, c, tabID, "continue button")
	// X-IDPI-Warning must be set exactly once (single value, not multi-valued).
	vals := hdr.Values("X-IDPI-Warning")
	if len(vals) == 0 {
		t.Fatal("expected at least one X-IDPI-Warning header")
	}
	if len(vals) > 1 {
		t.Errorf("expected exactly one X-IDPI-Warning header value, got %d: %v", len(vals), vals)
	}
}

// TestIDPI_PDF_RawMode_Blocked ensures that /pdf?raw=true is also blocked in
// strict mode — it delegates to the same HandlePDF function.
func TestIDPI_PDF_RawMode_Blocked(t *testing.T) {
	c := newIDPIServer(t, idpiCfg(true, true, true))
	tabID := idpiNavigate(t, c, idpiInjectPageURL(t))
	code, _, _ := c.GetWithHeaders(t, fmt.Sprintf("/tabs/%s/pdf?raw=true", tabID))
	if code != http.StatusForbidden {
		t.Fatalf("expected 403 for raw PDF in strict mode on injection page, got %d", code)
	}
}

// TestIDPI_Find_WrapContent_NotApplied verifies that WrapContent (which applies
// to /text responses) has no effect on /find responses.
func TestIDPI_Find_WrapContent_NotApplied(t *testing.T) {
	c := newIDPIServer(t, appconfig.IDPIConfig{
		Enabled:     true,
		ScanContent: true,
		WrapContent: true,
		StrictMode:  false,
	})
	tabID := idpiNavigate(t, c, idpiCleanPageURL(t))
	code, body, _ := idpiFind(t, c, tabID, "safe action button")
	if code != 200 {
		t.Fatalf("expected 200, got %d: %s", code, body)
	}
	// /find always returns JSON; WrapContent must not alter the response type.
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Errorf("expected JSON response from /find with WrapContent=true, got: %s", body)
	}
}

// --- helpers ---

// assertPDFBody verifies that the response body contains a valid PDF (base64 or raw).
func assertPDFBody(t *testing.T, body []byte) {
	t.Helper()
	s := strings.TrimSpace(string(body))
	// Raw PDF bytes begin with %PDF
	if strings.HasPrefix(s, "%PDF") {
		return
	}
	// JSON base64 response
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("response is neither raw PDF nor JSON: %s", body)
	}
	b64, _ := m["base64"].(string)
	if b64 == "" {
		t.Fatal("JSON PDF response missing 'base64' field")
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("base64 field is invalid: %v", err)
	}
	if len(data) < 4 || string(data[:4]) != "%PDF" {
		t.Errorf("decoded base64 data is not a PDF (prefix: %q)", string(data[:min(4, len(data))]))
	}
}

// jsonFieldRaw returns the raw JSON representation of a top-level key, or ""
// if the key is absent. Unlike testutil.JSONField it does not coerce to string,
// making it suitable for asserting null / missing optional fields.
func jsonFieldRaw(t *testing.T, data []byte, key string) string {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	return string(v)
}
