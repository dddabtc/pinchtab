package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pinchtab/pinchtab/internal/activity"
	"github.com/pinchtab/pinchtab/internal/bridge"
	"github.com/pinchtab/pinchtab/internal/httpx"
)

// HandleSolveCloudflare attempts to solve a Cloudflare Turnstile challenge on the current page.
//
// @Endpoint POST /solve-cloudflare
// @Description Detect and solve Cloudflare Turnstile/Interstitial challenges
//
// @Param tabId string body Tab ID (optional — uses default tab)
// @Param maxAttempts int body Max solve attempts (optional, default: 3)
// @Param timeout float64 body Timeout in ms (optional, default: 30000)
//
// @Response 200 application/json Returns {tabId, solved, challengeType, attempts, title}
// @Response 400 application/json Invalid request body or parameters
// @Response 423 application/json Tab is locked by another owner
// @Response 500 application/json Chrome/CDP error
//
// @Example curl:
//
//	curl -X POST http://localhost:9867/solve-cloudflare \
//	  -H "Content-Type: application/json" \
//	  -d '{"maxAttempts": 3, "timeout": 30000}'
//
// @Example cli:
//
//	pinchtab solve-cf
func (h *Handlers) HandleSolveCloudflare(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TabID       string  `json:"tabId"`
		MaxAttempts int     `json:"maxAttempts"`
		Timeout     float64 `json:"timeout"`
	}

	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&req); err != nil {
		httpx.Error(w, 400, fmt.Errorf("decode: %w", err))
		return
	}

	ctx, resolvedTabID, err := h.tabContext(r, req.TabID)
	if err != nil {
		httpx.Error(w, 404, err)
		return
	}

	owner := resolveOwner(r, "")
	if err := h.enforceTabLease(resolvedTabID, owner); err != nil {
		httpx.ErrorCode(w, 423, "tab_locked", err.Error(), false, nil)
		return
	}

	if _, ok := h.enforceCurrentTabDomainPolicy(w, r, ctx, resolvedTabID); !ok {
		return
	}

	h.recordActivity(r, activity.Update{Action: "solve-cloudflare", TabID: resolvedTabID})

	timeout := 30 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Millisecond
	}

	tCtx, tCancel := context.WithTimeout(ctx, timeout)
	defer tCancel()
	go httpx.CancelOnClientDone(r.Context(), tCancel)

	result, err := bridge.SolveCloudflare(tCtx, req.MaxAttempts)
	if err != nil {
		httpx.Error(w, 500, fmt.Errorf("solve cloudflare: %w", err))
		return
	}

	httpx.JSON(w, 200, map[string]any{
		"tabId":         resolvedTabID,
		"solved":        result.Solved,
		"challengeType": result.ChallengeType,
		"attempts":      result.Attempts,
		"title":         result.Title,
	})
}

// HandleTabSolveCloudflare handles POST /tabs/{id}/solve-cloudflare.
//
// @Endpoint POST /tabs/{id}/solve-cloudflare
// @Description Solve Cloudflare challenge on a specific tab
func (h *Handlers) HandleTabSolveCloudflare(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("id")
	if tabID == "" {
		httpx.Error(w, 400, fmt.Errorf("tab id required"))
		return
	}

	body := map[string]any{}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize))
	if err := dec.Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		httpx.Error(w, 400, fmt.Errorf("decode: %w", err))
		return
	}

	body["tabId"] = tabID
	payload, err := json.Marshal(body)
	if err != nil {
		httpx.Error(w, 500, fmt.Errorf("encode: %w", err))
		return
	}

	cloned := r.Clone(r.Context())
	cloned.Body = io.NopCloser(bytes.NewReader(payload))
	cloned.ContentLength = int64(len(payload))
	cloned.Header = r.Header.Clone()
	cloned.Header.Set("Content-Type", "application/json")
	h.HandleSolveCloudflare(w, cloned)
}
