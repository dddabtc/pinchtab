package actions

import (
	"fmt"
	"net/http"

	"github.com/pinchtab/pinchtab/internal/cli"
	"github.com/pinchtab/pinchtab/internal/cli/apiclient"
	"github.com/spf13/cobra"
)

func Find(client *http.Client, base, token string, args []string) {
	if len(args) == 0 {
		cli.Fatal("Usage: pinchtab find <query> [--tab <id>] [--threshold <n>] [--explain] [--ref-only]")
	}

	query := args[0]
	tabID := ""
	threshold := ""
	explain := false
	refOnly := false

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--tab":
			if i+1 < len(args) {
				i++
				tabID = args[i]
			}
		case "--threshold":
			if i+1 < len(args) {
				i++
				threshold = args[i]
			}
		case "--explain":
			explain = true
		case "--ref-only":
			refOnly = true
		}
	}

	body := map[string]any{"query": query}
	if tabID != "" {
		body["tabId"] = tabID
	}
	if threshold != "" {
		body["threshold"] = threshold
	}
	if explain {
		body["explain"] = true
	}

	path := "/find"
	if tabID != "" {
		path = fmt.Sprintf("/tabs/%s/find", tabID)
		delete(body, "tabId")
	}

	result := apiclient.DoPost(client, base, token, path, body)

	if refOnly {
		if ref, ok := result["best_ref"].(string); ok && ref != "" {
			fmt.Println(ref)
			return
		}
		cli.Fatal("No element found")
	}
}

func FindWithFlags(client *http.Client, base, token string, query string, cmd *cobra.Command) {
	tabID, _ := cmd.Flags().GetString("tab")
	threshold, _ := cmd.Flags().GetString("threshold")
	explain, _ := cmd.Flags().GetBool("explain")
	refOnly, _ := cmd.Flags().GetBool("ref-only")

	body := map[string]any{"query": query}
	if threshold != "" {
		body["threshold"] = threshold
	}
	if explain {
		body["explain"] = true
	}

	path := "/find"
	if tabID != "" {
		path = fmt.Sprintf("/tabs/%s/find", tabID)
	}

	result := apiclient.DoPost(client, base, token, path, body)

	if refOnly {
		if ref, ok := result["best_ref"].(string); ok && ref != "" {
			fmt.Println(ref)
			return
		}
		cli.Fatal("No element found")
	}
}
