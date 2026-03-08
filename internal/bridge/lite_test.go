package bridge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gost-dom/browser"
	"github.com/gost-dom/browser/dom"
)

func TestLiteBridge_Navigate(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
	<h1>Hello World</h1>
	<a href="/about">About</a>
	<button id="btn">Click Me</button>
	<input type="text" placeholder="Enter name">
</body>
</html>`))
	}))
	defer ts.Close()

	// Create lite bridge with handler
	b := browser.New(browser.WithHandler(ts.Config.Handler))
	lite := &LiteBridge{
		browser: b,
		tabID:   "lite-0",
		refMap:  make(map[string]dom.Element),
	}
	defer lite.Close()

	ctx := context.Background()

	// Navigate
	result, err := lite.Navigate(ctx, ts.URL)
	if err != nil {
		t.Fatalf("Navigate failed: %v", err)
	}

	if result.Title != "Test Page" {
		t.Errorf("expected title 'Test Page', got %q", result.Title)
	}

	// Text
	text, err := lite.Text(ctx)
	if err != nil {
		t.Fatalf("Text failed: %v", err)
	}
	if !strings.Contains(text, "Hello World") {
		t.Errorf("expected text to contain 'Hello World', got %q", text)
	}

	// Snapshot
	snap, err := lite.Snapshot(ctx, "")
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	if len(snap.Nodes) == 0 {
		t.Error("expected nodes in snapshot")
	}

	// Check for interactive elements
	snap, err = lite.Snapshot(ctx, "interactive")
	if err != nil {
		t.Fatalf("Snapshot interactive failed: %v", err)
	}

	foundLink := false
	foundButton := false
	foundInput := false
	for _, n := range snap.Nodes {
		if n.Role == "link" && n.Name == "About" {
			foundLink = true
		}
		if n.Role == "button" && n.Name == "Click Me" {
			foundButton = true
		}
		if n.Role == "textbox" {
			foundInput = true
		}
	}

	if !foundLink {
		t.Error("expected to find link in snapshot")
	}
	if !foundButton {
		t.Error("expected to find button in snapshot")
	}
	if !foundInput {
		t.Error("expected to find input in snapshot")
	}
}

func TestLiteBridge_Click(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<body>
	<button id="btn">Click Me</button>
</body>
</html>`))
	}))
	defer ts.Close()

	b := browser.New(browser.WithHandler(ts.Config.Handler))
	lite := &LiteBridge{
		browser: b,
		tabID:   "lite-0",
		refMap:  make(map[string]dom.Element),
	}
	defer lite.Close()

	ctx := context.Background()

	_, err := lite.Navigate(ctx, ts.URL)
	if err != nil {
		t.Fatalf("Navigate failed: %v", err)
	}

	// Snapshot to get refs
	snap, err := lite.Snapshot(ctx, "interactive")
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// Find button ref
	var buttonRef string
	for _, n := range snap.Nodes {
		if n.Role == "button" {
			buttonRef = n.Ref
			break
		}
	}

	if buttonRef == "" {
		t.Fatal("button ref not found")
	}

	// Click
	err = lite.Click(ctx, buttonRef)
	if err != nil {
		t.Fatalf("Click failed: %v", err)
	}
}

func TestLiteBridge_ScreenshotNotSupported(t *testing.T) {
	lite := NewLiteBridge()
	defer lite.Close()

	_, err := lite.Screenshot(context.Background())
	if err != ErrLiteNotSupported {
		t.Errorf("expected ErrLiteNotSupported, got %v", err)
	}
}

func TestLiteBridge_PDFNotSupported(t *testing.T) {
	lite := NewLiteBridge()
	defer lite.Close()

	_, err := lite.PDF(context.Background())
	if err != ErrLiteNotSupported {
		t.Errorf("expected ErrLiteNotSupported, got %v", err)
	}
}
