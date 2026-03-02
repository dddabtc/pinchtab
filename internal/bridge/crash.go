package bridge

import (
	"log/slog"
	"time"

	"github.com/chromedp/cdproto/inspector"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// CrashEvent contains information about a crash
type CrashEvent struct {
	Time      time.Time
	TargetID  string
	TabID     string
	URL       string
	Reason    string
	LastError string
}

// CrashHandler is called when a crash is detected
type CrashHandler func(event CrashEvent)

// MonitorCrashes listens for browser and tab crashes
func (b *Bridge) MonitorCrashes(handler CrashHandler) {
	if b.BrowserCtx == nil {
		slog.Warn("cannot monitor crashes: no browser context")
		return
	}

	// Listen for target crashes on browser context
	chromedp.ListenTarget(b.BrowserCtx, func(ev interface{}) {
		switch e := ev.(type) {
		case *inspector.EventTargetCrashed:
			slog.Error("🔥 TARGET CRASHED",
				"event", "inspector.targetCrashed",
			)
			if handler != nil {
				handler(CrashEvent{
					Time:   time.Now(),
					Reason: "inspector.targetCrashed",
				})
			}

		case *target.EventTargetCrashed:
			slog.Error("🔥 TARGET CRASHED",
				"targetId", e.TargetID,
				"status", e.Status,
				"errorCode", e.ErrorCode,
			)
			if handler != nil {
				handler(CrashEvent{
					Time:     time.Now(),
					TargetID: string(e.TargetID),
					Reason:   e.Status,
				})
			}

		case *target.EventTargetDestroyed:
			slog.Debug("target destroyed", "targetId", e.TargetID)
		}
	})

	// Monitor browser context cancellation
	go func() {
		<-b.BrowserCtx.Done()
		err := b.BrowserCtx.Err()
		slog.Warn("🔥 BROWSER CONTEXT ENDED",
			"error", err,
		)
		if handler != nil {
			reason := "context cancelled"
			if err != nil {
				reason = err.Error()
			}
			handler(CrashEvent{
				Time:   time.Now(),
				Reason: reason,
			})
		}
	}()

	slog.Info("crash monitoring enabled")
}

// GetCrashLogs returns recent crash information from Chrome's preferences
func (b *Bridge) GetCrashLogs() []string {
	if b.Config == nil || b.Config.ProfileDir == "" {
		return nil
	}

	var logs []string

	// Check if last exit was unclean
	if WasUncleanExit(b.Config.ProfileDir) {
		logs = append(logs, "Previous session ended with unclean exit (crash detected)")
	}

	return logs
}
