package actions

import (
	"fmt"
	"github.com/pinchtab/pinchtab/internal/cli"
	"github.com/pinchtab/pinchtab/internal/cli/apiclient"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func Tabs(client *http.Client, base, token string, args []string) {
	if len(args) == 0 {
		// List all tabs
		apiclient.DoGet(client, base, token, "/tabs", nil)
		return
	}

	cmd := args[0]
	subArgs := args[1:]

	// Check if this is a tab operation (navigate, snapshot, click, etc.)
	// Pattern: pinchtab tab <operation> <tabId> [args...]
	if isTabOperation(cmd) {
		TabOperation(client, base, token, cmd, subArgs)
		return
	}

	// Legacy: pinchtab tab new/close
	switch cmd {
	case "new":
		url := ""
		if len(subArgs) > 0 {
			url = subArgs[0]
		}

		// Check if any instances are running
		instances := getInstances(client, base, token)
		if len(instances) == 0 {
			fmt.Fprintln(os.Stderr, cli.StyleStderr(cli.WarningStyle, "No instances running, launching default..."))
			launchInstance(client, base, token, "default")
			fmt.Fprintln(os.Stderr, cli.StyleStderr(cli.SuccessStyle, "Instance launched"))
		}

		body := map[string]any{"action": "new"}
		if url != "" {
			body["url"] = url
		}
		apiclient.DoPost(client, base, token, "/tab", body)

	case "close":
		if len(subArgs) < 1 {
			cli.Fatal("Usage: pinchtab tab close <tabId>")
		}
		apiclient.DoPost(client, base, token, "/tab", map[string]any{
			"action": "close",
			"tabId":  subArgs[0],
		})

	default:
		TabOperation(client, base, token, cmd, subArgs)
	}
}

func isTabOperation(op string) bool {
	ops := map[string]bool{
		"navigate": true, "snapshot": true, "screenshot": true,
		"click": true, "type": true, "press": true, "fill": true,
		"hover": true, "scroll": true, "select": true, "focus": true,
		"text": true, "eval": true, "evaluate": true, "pdf": true,
		"cookies": true, "lock": true, "unlock": true, "locks": true,
		"fingerprint": true, "info": true,
	}
	return ops[op]
}

func TabOperation(client *http.Client, base, token string, op string, args []string) {
	if len(args) < 1 {
		cli.Fatal("Usage: pinchtab tab %s <tabId> [args...]", op)
	}

	tabID := args[0]
	restArgs := args[1:]

	switch op {
	case "navigate":
		if len(restArgs) < 1 {
			cli.Fatal("Usage: pinchtab tab navigate <tabId> <url> [--timeout N] [--block-images]")
		}
		body := map[string]any{"url": restArgs[0]}
		for i := 1; i < len(restArgs); i++ {
			switch restArgs[i] {
			case "--timeout":
				if i+1 < len(restArgs) {
					body["timeout"] = restArgs[i+1]
					i++
				}
			case "--block-images":
				body["blockImages"] = true
			case "--block-ads":
				body["blockAds"] = true
			}
		}
		apiclient.DoPost(client, base, token, fmt.Sprintf("/tabs/%s/navigate", tabID), body)

	case "snapshot":
		params := url.Values{}
		for _, arg := range restArgs {
			switch arg {
			case "-i", "--interactive":
				params.Set("interactive", "true")
			case "-c", "--compact":
				params.Set("compact", "true")
			case "-d", "--diff":
				params.Set("diff", "true")
			}
		}
		apiclient.DoGet(client, base, token, fmt.Sprintf("/tabs/%s/snapshot", tabID), params)

	case "screenshot", "ss":
		params := url.Values{}
		outFile := ""
		for i := 0; i < len(restArgs); i++ {
			switch restArgs[i] {
			case "-o", "--output":
				if i+1 < len(restArgs) {
					outFile = restArgs[i+1]
					i++
				}
			case "-q", "--quality":
				if i+1 < len(restArgs) {
					params.Set("quality", restArgs[i+1])
					i++
				}
			}
		}
		params.Set("raw", "true")
		data := apiclient.DoGetRaw(client, base, token, fmt.Sprintf("/tabs/%s/screenshot", tabID), params)
		if outFile == "" {
			outFile = fmt.Sprintf("screenshot-%s.png", time.Now().Format("20060102-150405"))
		}
		if data != nil {
			if err := os.WriteFile(outFile, data, 0600); err != nil {
				cli.Fatal("Write failed: %v", err)
			}
			fmt.Println(cli.StyleStdout(cli.SuccessStyle, fmt.Sprintf("Saved %s (%d bytes)", outFile, len(data))))
		}

	case "click", "hover", "focus":
		if len(restArgs) < 1 {
			cli.Fatal("Usage: pinchtab tab %s <tabId> <ref>", op)
		}
		body := map[string]any{"kind": op, "ref": restArgs[0]}
		apiclient.DoPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "type":
		if len(restArgs) < 2 {
			cli.Fatal("Usage: pinchtab tab type <tabId> <ref> <text>")
		}
		body := map[string]any{"kind": "type", "ref": restArgs[0], "text": strings.Join(restArgs[1:], " ")}
		apiclient.DoPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "fill":
		if len(restArgs) < 2 {
			cli.Fatal("Usage: pinchtab tab fill <tabId> <ref> <text>")
		}
		body := map[string]any{"kind": "fill", "ref": restArgs[0], "text": strings.Join(restArgs[1:], " ")}
		apiclient.DoPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "press":
		if len(restArgs) < 1 {
			cli.Fatal("Usage: pinchtab tab press <tabId> <key>")
		}
		body := map[string]any{"kind": "press", "key": restArgs[0]}
		apiclient.DoPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "scroll":
		if len(restArgs) < 1 {
			cli.Fatal("Usage: pinchtab tab scroll <tabId> <direction|pixels>")
		}
		body := map[string]any{}
		if v, err := strconv.Atoi(restArgs[0]); err == nil {
			body["kind"] = "scroll"
			body["scrollY"] = v
		} else {
			body["kind"] = "scroll"
			body["direction"] = restArgs[0]
		}
		apiclient.DoPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "select":
		if len(restArgs) < 2 {
			cli.Fatal("Usage: pinchtab tab select <tabId> <ref> <value>")
		}
		body := map[string]any{"kind": "select", "ref": restArgs[0], "value": restArgs[1]}
		apiclient.DoPost(client, base, token, fmt.Sprintf("/tabs/%s/action", tabID), body)

	case "text":
		params := url.Values{}
		for _, arg := range restArgs {
			if arg == "--raw" {
				params.Set("raw", "true")
			}
		}
		apiclient.DoGet(client, base, token, fmt.Sprintf("/tabs/%s/text", tabID), params)

	case "eval", "evaluate":
		if len(restArgs) < 1 {
			cli.Fatal("Usage: pinchtab tab eval <tabId> <expression>")
		}
		body := map[string]any{"expression": strings.Join(restArgs, " ")}
		apiclient.DoPost(client, base, token, fmt.Sprintf("/tabs/%s/evaluate", tabID), body)

	case "pdf":
		params := url.Values{}
		outFile := ""
		for i := 0; i < len(restArgs); i++ {
			switch restArgs[i] {
			case "-o", "--output":
				if i+1 < len(restArgs) {
					outFile = restArgs[i+1]
					i++
				}
			case "--landscape":
				params.Set("landscape", "true")
			case "--scale":
				if i+1 < len(restArgs) {
					params.Set("scale", restArgs[i+1])
					i++
				}
			}
		}
		params.Set("raw", "true")
		data := apiclient.DoGetRaw(client, base, token, fmt.Sprintf("/tabs/%s/pdf", tabID), params)
		if outFile == "" {
			outFile = fmt.Sprintf("page-%s.pdf", time.Now().Format("20060102-150405"))
		}
		if data != nil {
			if err := os.WriteFile(outFile, data, 0600); err != nil {
				cli.Fatal("Write failed: %v", err)
			}
			fmt.Println(cli.StyleStdout(cli.SuccessStyle, fmt.Sprintf("Saved %s (%d bytes)", outFile, len(data))))
		}

	case "cookies":
		apiclient.DoGet(client, base, token, fmt.Sprintf("/tabs/%s/cookies", tabID), nil)

	case "lock":
		body := map[string]any{}
		for i := 0; i < len(restArgs); i++ {
			switch restArgs[i] {
			case "--owner":
				if i+1 < len(restArgs) {
					body["owner"] = restArgs[i+1]
					i++
				}
			case "--ttl":
				if i+1 < len(restArgs) {
					if ttl, err := strconv.Atoi(restArgs[i+1]); err == nil {
						body["ttl"] = ttl
					}
					i++
				}
			}
		}
		apiclient.DoPost(client, base, token, fmt.Sprintf("/tabs/%s/lock", tabID), body)

	case "unlock":
		body := map[string]any{}
		for i := 0; i < len(restArgs); i++ {
			switch restArgs[i] {
			case "--owner":
				if i+1 < len(restArgs) {
					body["owner"] = restArgs[i+1]
					i++
				}
			}
		}
		apiclient.DoPost(client, base, token, fmt.Sprintf("/tabs/%s/unlock", tabID), body)

	case "locks":
		apiclient.DoGet(client, base, token, fmt.Sprintf("/tabs/%s/locks", tabID), nil)

	case "info":
		apiclient.DoGet(client, base, token, fmt.Sprintf("/tabs/%s", tabID), nil)

	default:
		cli.Fatal("Unknown tab operation: %s", op)
	}
}

// TabNew opens a new tab (exported for cobra subcommand).
func TabNew(client *http.Client, base, token string, body map[string]any) {
	// Check if any instances are running
	instances := getInstances(client, base, token)
	if len(instances) == 0 {
		fmt.Fprintln(os.Stderr, cli.StyleStderr(cli.WarningStyle, "No instances running, launching default..."))
		launchInstance(client, base, token, "default")
		fmt.Fprintln(os.Stderr, cli.StyleStderr(cli.SuccessStyle, "Instance launched"))
	}
	apiclient.DoPost(client, base, token, "/tab", body)
}

// TabClose closes a tab by ID (exported for cobra subcommand).
func TabClose(client *http.Client, base, token string, tabID string) {
	apiclient.DoPost(client, base, token, "/tab", map[string]any{
		"action": "close",
		"tabId":  tabID,
	})
}
