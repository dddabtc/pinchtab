package actions

import (
	"github.com/pinchtab/pinchtab/internal/cli/apiclient"
	"github.com/spf13/cobra"
	"net/http"
	"net/url"
)

func Text(client *http.Client, base, token string, args []string) {
	params := url.Values{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--raw":
			params.Set("mode", "raw")
		case "--tab":
			if i+1 < len(args) {
				i++
				params.Set("tabId", args[i])
			}
		}
	}
	apiclient.DoGet(client, base, token, "/text", params)
}

func TextWithFlags(client *http.Client, base, token string, cmd *cobra.Command) {
	params := url.Values{}
	if v, _ := cmd.Flags().GetBool("raw"); v {
		params.Set("mode", "raw")
	}
	if v, _ := cmd.Flags().GetString("tab"); v != "" {
		params.Set("tabId", v)
	}
	apiclient.DoGet(client, base, token, "/text", params)
}
