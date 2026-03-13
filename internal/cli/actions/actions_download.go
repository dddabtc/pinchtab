package actions

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/pinchtab/pinchtab/internal/cli"
	"github.com/pinchtab/pinchtab/internal/cli/apiclient"
)

func Download(client *http.Client, base, token string, args []string, output string) {
	if len(args) < 1 {
		cli.Fatal("Usage: pinchtab download <url> [-o <file>]")
	}

	targetURL := args[0]

	params := url.Values{}
	params.Set("url", targetURL)

	result := apiclient.DoGet(client, base, token, "/download", params)

	// If -o flag set, decode base64 and save to file
	if output != "" {
		b64, _ := result["data"].(string)
		if b64 == "" {
			cli.Fatal("No base64 data in response")
		}
		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			cli.Fatal("Failed to decode base64: %v", err)
		}
		if err := os.WriteFile(output, data, 0600); err != nil {
			cli.Fatal("Write failed: %v", err)
		}
		fmt.Println(cli.StyleStdout(cli.SuccessStyle, fmt.Sprintf("Saved %s (%d bytes)", output, len(data))))
	}
}
