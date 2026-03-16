package bridge

import (
	"context"
	"strings"
	"testing"

	"github.com/pinchtab/pinchtab/internal/config"
)

func TestClickAction_UsesCoordinatePathIncludingZeroZero(t *testing.T) {
	b := New(context.TODO(), nil, &config.RuntimeConfig{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := b.Actions[ActionClick](ctx, ActionRequest{HasXY: true, X: 0, Y: 0})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if strings.Contains(err.Error(), "need selector") {
		t.Fatalf("expected coordinate path, got selector/ref validation error: %v", err)
	}
}

func TestDoubleClickAction_UsesCoordinatePathIncludingZeroZero(t *testing.T) {
	b := New(context.TODO(), nil, &config.RuntimeConfig{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := b.Actions[ActionDoubleClick](ctx, ActionRequest{HasXY: true, X: 0, Y: 0})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if strings.Contains(err.Error(), "need selector") {
		t.Fatalf("expected coordinate path, got selector/ref validation error: %v", err)
	}
}

func TestHoverAction_UsesCoordinatePath(t *testing.T) {
	b := New(context.TODO(), nil, &config.RuntimeConfig{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := b.Actions[ActionHover](ctx, ActionRequest{HasXY: true, X: 12.5, Y: 34.5})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if strings.Contains(err.Error(), "need selector") {
		t.Fatalf("expected coordinate path, got selector/ref validation error: %v", err)
	}
}

func TestCheckAction_Registered(t *testing.T) {
	b := New(context.TODO(), nil, &config.RuntimeConfig{})
	if _, ok := b.Actions[ActionCheck]; !ok {
		t.Fatal("ActionCheck not registered in action registry")
	}
}

func TestUncheckAction_Registered(t *testing.T) {
	b := New(context.TODO(), nil, &config.RuntimeConfig{})
	if _, ok := b.Actions[ActionUncheck]; !ok {
		t.Fatal("ActionUncheck not registered in action registry")
	}
}

func TestCheckAction_RequiresSelectorOrRef(t *testing.T) {
	b := New(context.TODO(), nil, &config.RuntimeConfig{})
	_, err := b.Actions[ActionCheck](context.Background(), ActionRequest{})
	if err == nil {
		t.Fatal("expected error when no selector/ref/nodeId provided")
	}
	if !strings.Contains(err.Error(), "need selector") {
		t.Fatalf("expected 'need selector' error, got: %v", err)
	}
}

func TestUncheckAction_RequiresSelectorOrRef(t *testing.T) {
	b := New(context.TODO(), nil, &config.RuntimeConfig{})
	_, err := b.Actions[ActionUncheck](context.Background(), ActionRequest{})
	if err == nil {
		t.Fatal("expected error when no selector/ref/nodeId provided")
	}
	if !strings.Contains(err.Error(), "need selector") {
		t.Fatalf("expected 'need selector' error, got: %v", err)
	}
}

func TestCheckAction_WithNodeID_UsesResolveNode(t *testing.T) {
	b := New(context.TODO(), nil, &config.RuntimeConfig{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := b.Actions[ActionCheck](ctx, ActionRequest{NodeID: 42})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	// Should NOT be a validation error — it should attempt the CDP path
	if strings.Contains(err.Error(), "need selector") {
		t.Fatalf("expected CDP path, got validation error: %v", err)
	}
}

func TestUncheckAction_WithSelector_UsesCSSPath(t *testing.T) {
	b := New(context.TODO(), nil, &config.RuntimeConfig{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := b.Actions[ActionUncheck](ctx, ActionRequest{Selector: "#my-checkbox"})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if strings.Contains(err.Error(), "need selector") {
		t.Fatalf("expected CSS path, got validation error: %v", err)
	}
}

func TestXpathString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "'hello'"},
		{"it's", "concat('it', \"'\", 's')"},
		{"say \"hi\"", "'say \"hi\"'"},
		{"it's a \"test\"", "concat('it', \"'\", 's a \"test\"')"},
		{"", "''"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := xpathString(tt.input)
			if result != tt.expected {
				t.Errorf("xpathString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
