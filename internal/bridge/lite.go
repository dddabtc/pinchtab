// Package bridge provides browser control implementations.
// This file implements the "lite" engine using Gost-DOM for fast,
// Chrome-free DOM capture and interaction.
package bridge

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gost-dom/browser"
	"github.com/gost-dom/browser/dom"
	"github.com/gost-dom/browser/html"
)

var (
	ErrLiteNotSupported = errors.New("operation not supported in lite mode")
)

// LiteBridge implements Bridge using Gost-DOM for lightweight DOM operations.
// It provides fast page capture without requiring Chrome.
type LiteBridge struct {
	browser *browser.Browser
	window  html.Window
	mu      sync.RWMutex
	tabID   string
	url     string
	refMap  map[string]dom.Element // cached refs from last snapshot
}

// NewLiteBridge creates a new Gost-DOM based bridge.
func NewLiteBridge() *LiteBridge {
	b := browser.New()
	return &LiteBridge{
		browser: b,
		tabID:   "lite-0",
		refMap:  make(map[string]dom.Element),
	}
}

// Navigate opens a URL in the lite browser.
func (l *LiteBridge) Navigate(ctx context.Context, url string) (*NavigateResult, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Close previous window if any
	if l.window != nil {
		l.window.Close()
	}

	win, err := l.browser.Open(url)
	if err != nil {
		return nil, fmt.Errorf("lite navigate: %w", err)
	}

	l.window = win
	l.url = url
	l.refMap = make(map[string]dom.Element) // clear refs

	// Process any pending JS events
	if l.browser.Clock != nil {
		_ = l.browser.Clock.ProcessEvents(ctx)
	}

	title := l.getTitle(win)

	return &NavigateResult{
		TabID: l.tabID,
		URL:   url,
		Title: title,
	}, nil
}

// Snapshot returns the DOM tree as accessible nodes.
func (l *LiteBridge) Snapshot(ctx context.Context, filter string) (*SnapshotResult, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.window == nil {
		return nil, errors.New("no page loaded")
	}

	doc := l.window.Document()
	if doc == nil {
		return nil, errors.New("no document")
	}

	body := doc.Body()
	if body == nil {
		return nil, errors.New("no body")
	}

	// Clear and rebuild ref map
	l.refMap = make(map[string]dom.Element)

	nodes := l.walkDOM(body, filter, 0)

	return &SnapshotResult{
		TabID: l.tabID,
		URL:   l.url,
		Title: l.getTitle(l.window),
		Nodes: nodes,
	}, nil
}

// walkDOM traverses the DOM tree and builds snapshot nodes.
func (l *LiteBridge) walkDOM(node dom.Node, filter string, depth int) []SnapshotNode {
	var nodes []SnapshotNode

	el, isElement := node.(dom.Element)
	if !isElement {
		return nodes
	}

	role := l.getRole(el)
	name := l.getAccessibleName(el)
	interactive := l.isInteractive(el)

	// Apply filter
	if filter == "interactive" && !interactive {
		// Still walk children
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			nodes = append(nodes, l.walkDOM(child, filter, depth)...)
		}
		return nodes
	}

	// Generate ref
	ref := fmt.Sprintf("e%d", len(l.refMap))
	l.refMap[ref] = el

	sn := SnapshotNode{
		Ref:         ref,
		Role:        role,
		Name:        name,
		Tag:         strings.ToLower(el.TagName()),
		Interactive: interactive,
		Depth:       depth,
	}

	// Get value for inputs
	if input, ok := el.(html.HTMLInputElement); ok {
		sn.Value = input.Value()
	}

	nodes = append(nodes, sn)

	// Walk children
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		nodes = append(nodes, l.walkDOM(child, filter, depth+1)...)
	}

	return nodes
}

// getRole returns the ARIA role for an element.
func (l *LiteBridge) getRole(el dom.Element) string {
	// Check explicit role
	if role, ok := el.GetAttribute("role"); ok {
		return role
	}

	// Implicit roles
	tag := strings.ToLower(el.TagName())
	switch tag {
	case "a":
		if _, hasHref := el.GetAttribute("href"); hasHref {
			return "link"
		}
	case "button":
		return "button"
	case "input":
		inputType, _ := el.GetAttribute("type")
		switch inputType {
		case "submit", "button":
			return "button"
		case "checkbox":
			return "checkbox"
		case "radio":
			return "radio"
		case "text", "":
			return "textbox"
		default:
			return "textbox"
		}
	case "textarea":
		return "textbox"
	case "select":
		return "combobox"
	case "img":
		return "img"
	case "nav":
		return "navigation"
	case "main":
		return "main"
	case "header":
		return "banner"
	case "footer":
		return "contentinfo"
	case "aside":
		return "complementary"
	case "form":
		return "form"
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return "heading"
	case "ul", "ol":
		return "list"
	case "li":
		return "listitem"
	case "table":
		return "table"
	case "tr":
		return "row"
	case "td":
		return "cell"
	case "th":
		return "columnheader"
	}

	return "generic"
}

// getAccessibleName returns the accessible name for an element.
func (l *LiteBridge) getAccessibleName(el dom.Element) string {
	// aria-label takes precedence
	if label, ok := el.GetAttribute("aria-label"); ok {
		return label
	}

	// title attribute
	if title, ok := el.GetAttribute("title"); ok {
		return title
	}

	// alt for images
	if strings.ToLower(el.TagName()) == "img" {
		if alt, ok := el.GetAttribute("alt"); ok {
			return alt
		}
	}

	// placeholder for inputs
	if strings.ToLower(el.TagName()) == "input" || strings.ToLower(el.TagName()) == "textarea" {
		if placeholder, ok := el.GetAttribute("placeholder"); ok {
			return placeholder
		}
	}

	// Text content for interactive elements
	if l.isInteractive(el) {
		text := strings.TrimSpace(el.TextContent())
		if len(text) > 100 {
			text = text[:100] + "..."
		}
		return text
	}

	return ""
}

// getTitle extracts the page title from the <title> element.
func (l *LiteBridge) getTitle(win html.Window) string {
	if win == nil {
		return ""
	}
	doc := win.Document()
	if doc == nil {
		return ""
	}
	// Query for the title element
	titleEl, err := doc.QuerySelector("title")
	if err != nil || titleEl == nil {
		return ""
	}
	return strings.TrimSpace(titleEl.TextContent())
}

// isInteractive returns true if the element is interactive.
func (l *LiteBridge) isInteractive(el dom.Element) bool {
	tag := strings.ToLower(el.TagName())
	switch tag {
	case "a":
		_, hasHref := el.GetAttribute("href")
		return hasHref
	case "button", "input", "textarea", "select":
		return true
	}

	// Check for onclick or tabindex
	if _, ok := el.GetAttribute("onclick"); ok {
		return true
	}
	if tabindex, ok := el.GetAttribute("tabindex"); ok && tabindex != "-1" {
		return true
	}

	return false
}

// Text returns the text content of the page.
func (l *LiteBridge) Text(ctx context.Context) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.window == nil {
		return "", errors.New("no page loaded")
	}

	doc := l.window.Document()
	if doc == nil {
		return "", errors.New("no document")
	}

	body := doc.Body()
	if body == nil {
		return "", errors.New("no body")
	}

	return body.TextContent(), nil
}

// Click performs a click action on an element by ref.
func (l *LiteBridge) Click(ctx context.Context, ref string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	el, ok := l.refMap[ref]
	if !ok {
		return fmt.Errorf("ref %q not found (run snapshot first)", ref)
	}

	// Try to click if it's an HTMLElement
	if htmlEl, ok := el.(html.HTMLElement); ok {
		htmlEl.Click()

		// Process events after click
		if l.browser.Clock != nil {
			_ = l.browser.Clock.ProcessEvents(ctx)
		}

		return nil
	}

	return errors.New("element does not support click")
}

// Type enters text into an element by ref.
func (l *LiteBridge) Type(ctx context.Context, ref, text string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	el, ok := l.refMap[ref]
	if !ok {
		return fmt.Errorf("ref %q not found (run snapshot first)", ref)
	}

	// Set value on input elements
	if input, ok := el.(html.HTMLInputElement); ok {
		input.SetValue(text)
		return nil
	}

	// Try setting value attribute as fallback
	el.SetAttribute("value", text)
	return nil
}

// Screenshot is not supported in lite mode.
func (l *LiteBridge) Screenshot(ctx context.Context) ([]byte, error) {
	return nil, ErrLiteNotSupported
}

// PDF is not supported in lite mode.
func (l *LiteBridge) PDF(ctx context.Context) ([]byte, error) {
	return nil, ErrLiteNotSupported
}

// Tabs returns the single lite tab.
func (l *LiteBridge) Tabs(ctx context.Context) ([]TabInfo, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.window == nil {
		return []TabInfo{}, nil
	}

	return []TabInfo{
		{
			ID:    l.tabID,
			URL:   l.url,
			Title: l.getTitle(l.window),
			Type:  "page",
		},
	}, nil
}

// Close shuts down the lite browser.
func (l *LiteBridge) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.window != nil {
		l.window.Close()
		l.window = nil
	}
	if l.browser != nil {
		l.browser.Close()
		l.browser = nil
	}
	return nil
}

// Cookies returns cookies (limited support).
func (l *LiteBridge) Cookies(ctx context.Context) ([]http.Cookie, error) {
	// Gost-DOM has internal cookie handling but limited JS exposure
	return []http.Cookie{}, nil
}

// Evaluate runs JavaScript (if script engine is configured).
func (l *LiteBridge) Evaluate(ctx context.Context, script string) (any, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.window == nil {
		return nil, errors.New("no page loaded")
	}

	// Note: This requires a script engine to be configured on the browser
	// For now, return not supported
	return nil, errors.New("evaluate requires script engine configuration")
}

// NavigateResult holds the result of a navigation.
type NavigateResult struct {
	TabID string `json:"tabId"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

// SnapshotResult holds the result of a snapshot.
type SnapshotResult struct {
	TabID string         `json:"tabId"`
	URL   string         `json:"url"`
	Title string         `json:"title"`
	Nodes []SnapshotNode `json:"nodes"`
}

// SnapshotNode represents a DOM element in the snapshot.
type SnapshotNode struct {
	Ref         string `json:"ref"`
	Role        string `json:"role"`
	Name        string `json:"name,omitempty"`
	Tag         string `json:"tag"`
	Value       string `json:"value,omitempty"`
	Interactive bool   `json:"interactive,omitempty"`
	Depth       int    `json:"depth"`
}

// TabInfo holds information about a tab.
type TabInfo struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Title string `json:"title"`
	Type  string `json:"type"`
}
