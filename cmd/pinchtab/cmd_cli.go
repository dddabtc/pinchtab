package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	browseractions "github.com/pinchtab/pinchtab/internal/cli/actions"
	"github.com/pinchtab/pinchtab/internal/config"
	"github.com/spf13/cobra"
)

var quickCmd = &cobra.Command{
	Use:   "quick <url>",
	Short: "Navigate + analyze page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Quick(client, base, token, args)
		})
	},
}

var navCmd = &cobra.Command{
	Use:   "nav <url>",
	Short: "Navigate to URL",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Navigate(client, base, token, args)
		})
	},
}

var snapCmd = &cobra.Command{
	Use:   "snap",
	Short: "Snapshot accessibility tree",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Snapshot(client, base, token, args)
		})
	},
}

var clickCmd = &cobra.Command{
	Use:   "click <ref>",
	Short: "Click element",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Action(client, base, token, "click", args)
		})
	},
}

var typeCmd = &cobra.Command{
	Use:   "type <ref> <text>",
	Short: "Type into element",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Action(client, base, token, "type", args)
		})
	},
}

var screenshotCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Take a screenshot",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Screenshot(client, base, token, args)
		})
	},
}

var tabsCmd = &cobra.Command{
	Use:   "tabs",
	Short: "List or manage tabs",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Tabs(client, base, token, args)
		})
	},
}

var instancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "List or manage instances",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Instances(client, base, token)
		})
	},
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check server health",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Health(client, base, token)
		})
	},
}

var pressCmd = &cobra.Command{
	Use:   "press <key>",
	Short: "Press key (Enter, Tab, Escape...)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Action(client, base, token, "press", args)
		})
	},
}

var fillCmd = &cobra.Command{
	Use:   "fill <ref|selector> <text>",
	Short: "Fill input directly",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Action(client, base, token, "fill", args)
		})
	},
}

var hoverCmd = &cobra.Command{
	Use:   "hover <ref>",
	Short: "Hover element",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Action(client, base, token, "hover", args)
		})
	},
}

var scrollCmd = &cobra.Command{
	Use:   "scroll <ref|pixels>",
	Short: "Scroll to element or by pixels",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Action(client, base, token, "scroll", args)
		})
	},
}

var evalCmd = &cobra.Command{
	Use:   "eval <expression>",
	Short: "Evaluate JavaScript",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Evaluate(client, base, token, args)
		})
	},
}

var pdfCmd = &cobra.Command{
	Use:   "pdf",
	Short: "Export the current page as PDF",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.PDF(client, base, token, args)
		})
	},
}

var textCmd = &cobra.Command{
	Use:   "text",
	Short: "Extract page text",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Text(client, base, token, args)
		})
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download <url>",
	Short: "Download a file via browser session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Download(client, base, token, args, output)
		})
	},
}

var uploadCmd = &cobra.Command{
	Use:   "upload <file-path>",
	Short: "Upload a file to a file input element",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		selector, _ := cmd.Flags().GetString("selector")
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Upload(client, base, token, args, selector)
		})
	},
}

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List browser profiles",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Profiles(client, base, token)
		})
	},
}

var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Manage browser instances",
}

var findCmd = &cobra.Command{
	Use:   "find <query>",
	Short: "Find elements by natural language query",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Find(client, base, token, args)
		})
	},
}

var selectCmd = &cobra.Command{
	Use:   "select <ref> <value>",
	Short: "Select option in dropdown",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		runCLIWith(cfg, func(client *http.Client, base, token string) {
			browseractions.Action(client, base, token, "select", args)
		})
	},
}

func init() {
	quickCmd.GroupID = "browser"
	navCmd.GroupID = "browser"
	snapCmd.GroupID = "browser"
	clickCmd.GroupID = "browser"
	typeCmd.GroupID = "browser"
	screenshotCmd.GroupID = "browser"
	tabsCmd.GroupID = "browser"
	instancesCmd.GroupID = "management"
	healthCmd.GroupID = "management"
	pressCmd.GroupID = "browser"
	fillCmd.GroupID = "browser"
	hoverCmd.GroupID = "browser"
	scrollCmd.GroupID = "browser"
	evalCmd.GroupID = "browser"
	pdfCmd.GroupID = "browser"
	textCmd.GroupID = "browser"
	profilesCmd.GroupID = "management"
	downloadCmd.GroupID = "browser"
	uploadCmd.GroupID = "browser"
	findCmd.GroupID = "browser"
	selectCmd.GroupID = "browser"

	// Disable cobra flag parsing for commands that parse flags manually.
	// FParseErrWhitelist silently swallows unknown flags (dropping them from args),
	// so we must disable cobra's parser entirely and let the action code handle flags.
	snapCmd.DisableFlagParsing = true
	screenshotCmd.DisableFlagParsing = true
	pdfCmd.DisableFlagParsing = true
	findCmd.DisableFlagParsing = true
	navCmd.DisableFlagParsing = true
	clickCmd.DisableFlagParsing = true
	hoverCmd.DisableFlagParsing = true
	textCmd.DisableFlagParsing = true
	tabsCmd.DisableFlagParsing = true

	uploadCmd.Flags().StringP("selector", "s", "", "CSS selector for file input")
	downloadCmd.Flags().StringP("output", "o", "", "Save downloaded file to path")

	rootCmd.AddCommand(quickCmd)
	rootCmd.AddCommand(navCmd)
	rootCmd.AddCommand(snapCmd)
	rootCmd.AddCommand(clickCmd)
	rootCmd.AddCommand(typeCmd)
	rootCmd.AddCommand(screenshotCmd)
	rootCmd.AddCommand(tabsCmd)
	rootCmd.AddCommand(instancesCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(pressCmd)
	rootCmd.AddCommand(fillCmd)
	rootCmd.AddCommand(hoverCmd)
	rootCmd.AddCommand(scrollCmd)
	rootCmd.AddCommand(evalCmd)
	rootCmd.AddCommand(pdfCmd)
	rootCmd.AddCommand(textCmd)
	rootCmd.AddCommand(profilesCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(findCmd)
	rootCmd.AddCommand(selectCmd)

	instanceCmd.GroupID = "management"
	instanceCmd.AddCommand(&cobra.Command{
		Use:   "start <name>",
		Short: "Start a browser instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			runCLIWith(cfg, func(client *http.Client, base, token string) {
				browseractions.InstanceStart(client, base, token, args)
			})
		},
	})
	instanceCmd.AddCommand(&cobra.Command{
		Use:   "navigate <id> <url>",
		Short: "Navigate an instance to a URL",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			runCLIWith(cfg, func(client *http.Client, base, token string) {
				browseractions.InstanceNavigate(client, base, token, args)
			})
		},
	})
	instanceCmd.AddCommand(&cobra.Command{
		Use:   "stop <id>",
		Short: "Stop a browser instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			runCLIWith(cfg, func(client *http.Client, base, token string) {
				browseractions.InstanceStop(client, base, token, args)
			})
		},
	})
	instanceCmd.AddCommand(&cobra.Command{
		Use:   "logs <id>",
		Short: "Get instance logs",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()
			runCLIWith(cfg, func(client *http.Client, base, token string) {
				browseractions.InstanceLogs(client, base, token, args)
			})
		},
	})
	rootCmd.AddCommand(instanceCmd)
}

func runCLIWith(cfg *config.RuntimeConfig, fn func(client *http.Client, base, token string)) {
	client := &http.Client{Timeout: 60 * time.Second}

	bind := cfg.Bind
	if bind == "" {
		bind = "127.0.0.1"
	}
	port := cfg.Port
	if port == "" {
		port = "9867"
	}
	base := fmt.Sprintf("http://%s:%s", bind, port)

	if envURL := os.Getenv("PINCHTAB_URL"); envURL != "" {
		base = strings.TrimRight(envURL, "/")
	}

	token := cfg.Token
	if envToken := os.Getenv("PINCHTAB_TOKEN"); envToken != "" {
		token = envToken
	}

	fn(client, base, token)
}
