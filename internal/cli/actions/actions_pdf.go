package actions

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pinchtab/pinchtab/internal/cli"
	"github.com/pinchtab/pinchtab/internal/cli/apiclient"
	"github.com/spf13/cobra"
)

func PDF(client *http.Client, base, token string, args []string) {
	params := url.Values{}
	params.Set("raw", "true")
	outFile := ""
	tabID := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 < len(args) {
				i++
				outFile = args[i]
			}
		case "--landscape":
			params.Set("landscape", "true")
		case "--scale":
			if i+1 < len(args) {
				i++
				params.Set("scale", args[i])
			}
		case "--tab":
			if i+1 < len(args) {
				i++
				tabID = args[i]
			}
		// Paper dimensions
		case "--paper-width":
			if i+1 < len(args) {
				i++
				params.Set("paperWidth", args[i])
			}
		case "--paper-height":
			if i+1 < len(args) {
				i++
				params.Set("paperHeight", args[i])
			}
		// Margins
		case "--margin-top":
			if i+1 < len(args) {
				i++
				params.Set("marginTop", args[i])
			}
		case "--margin-bottom":
			if i+1 < len(args) {
				i++
				params.Set("marginBottom", args[i])
			}
		case "--margin-left":
			if i+1 < len(args) {
				i++
				params.Set("marginLeft", args[i])
			}
		case "--margin-right":
			if i+1 < len(args) {
				i++
				params.Set("marginRight", args[i])
			}
		// Content options
		case "--page-ranges":
			if i+1 < len(args) {
				i++
				params.Set("pageRanges", args[i])
			}
		case "--prefer-css-page-size":
			params.Set("preferCSSPageSize", "true")
		// Header/Footer
		case "--display-header-footer":
			params.Set("displayHeaderFooter", "true")
		case "--header-template":
			if i+1 < len(args) {
				i++
				params.Set("headerTemplate", args[i])
			}
		case "--footer-template":
			if i+1 < len(args) {
				i++
				params.Set("footerTemplate", args[i])
			}
		// Accessibility
		case "--generate-tagged-pdf":
			params.Set("generateTaggedPDF", "true")
		case "--generate-document-outline":
			params.Set("generateDocumentOutline", "true")
		// Output options
		case "--file-output":
			params.Del("raw")
			params.Set("output", "file")
		case "--path":
			if i+1 < len(args) {
				i++
				params.Set("path", args[i])
			}
		case "--raw":
			params.Set("raw", "true")
		}
	}

	if outFile == "" {
		outFile = fmt.Sprintf("page-%s.pdf", time.Now().Format("20060102-150405"))
	}

	var data []byte
	if tabID != "" {
		data = apiclient.DoGetRaw(client, base, token, fmt.Sprintf("/tabs/%s/pdf", tabID), params)
	} else {
		data = apiclient.DoGetRaw(client, base, token, "/pdf", params)
	}
	if data == nil {
		return
	}
	if err := os.WriteFile(outFile, data, 0600); err != nil {
		cli.Fatal("Write failed: %v", err)
	}
	fmt.Println(cli.StyleStdout(cli.SuccessStyle, fmt.Sprintf("Saved %s (%d bytes)", outFile, len(data))))
}

func PDFWithFlags(client *http.Client, base, token string, cmd *cobra.Command) {
	params := url.Values{}
	params.Set("raw", "true")

	outFile, _ := cmd.Flags().GetString("output")
	tabID, _ := cmd.Flags().GetString("tab")

	if v, _ := cmd.Flags().GetBool("landscape"); v {
		params.Set("landscape", "true")
	}
	if v, _ := cmd.Flags().GetString("scale"); v != "" {
		params.Set("scale", v)
	}
	if v, _ := cmd.Flags().GetString("paper-width"); v != "" {
		params.Set("paperWidth", v)
	}
	if v, _ := cmd.Flags().GetString("paper-height"); v != "" {
		params.Set("paperHeight", v)
	}
	if v, _ := cmd.Flags().GetString("margin-top"); v != "" {
		params.Set("marginTop", v)
	}
	if v, _ := cmd.Flags().GetString("margin-bottom"); v != "" {
		params.Set("marginBottom", v)
	}
	if v, _ := cmd.Flags().GetString("margin-left"); v != "" {
		params.Set("marginLeft", v)
	}
	if v, _ := cmd.Flags().GetString("margin-right"); v != "" {
		params.Set("marginRight", v)
	}
	if v, _ := cmd.Flags().GetString("page-ranges"); v != "" {
		params.Set("pageRanges", v)
	}
	if v, _ := cmd.Flags().GetBool("prefer-css-page-size"); v {
		params.Set("preferCSSPageSize", "true")
	}
	if v, _ := cmd.Flags().GetBool("display-header-footer"); v {
		params.Set("displayHeaderFooter", "true")
	}
	if v, _ := cmd.Flags().GetString("header-template"); v != "" {
		params.Set("headerTemplate", v)
	}
	if v, _ := cmd.Flags().GetString("footer-template"); v != "" {
		params.Set("footerTemplate", v)
	}
	if v, _ := cmd.Flags().GetBool("generate-tagged-pdf"); v {
		params.Set("generateTaggedPDF", "true")
	}
	if v, _ := cmd.Flags().GetBool("generate-document-outline"); v {
		params.Set("generateDocumentOutline", "true")
	}
	if v, _ := cmd.Flags().GetBool("file-output"); v {
		params.Del("raw")
		params.Set("output", "file")
	}
	if v, _ := cmd.Flags().GetString("path"); v != "" {
		params.Set("path", v)
	}

	if outFile == "" {
		outFile = fmt.Sprintf("page-%s.pdf", time.Now().Format("20060102-150405"))
	}

	var data []byte
	if tabID != "" {
		data = apiclient.DoGetRaw(client, base, token, fmt.Sprintf("/tabs/%s/pdf", tabID), params)
	} else {
		data = apiclient.DoGetRaw(client, base, token, "/pdf", params)
	}
	if data == nil {
		return
	}
	if err := os.WriteFile(outFile, data, 0600); err != nil {
		cli.Fatal("Write failed: %v", err)
	}
	fmt.Println(cli.StyleStdout(cli.SuccessStyle, fmt.Sprintf("Saved %s (%d bytes)", outFile, len(data))))
}
