// Package testutil provides shared helpers for pinchtab integration tests.
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appconfig "github.com/pinchtab/pinchtab/internal/config"
)

const (
	HealthTimeout   = 30 * time.Second
	HealthTimeoutCI = 60 * time.Second
	ShutdownTimeout = 10 * time.Second
	InstanceTimeout = 30 * time.Second
)

// ServerConfig mirrors key fields from internal/config.RuntimeConfig for test setup.
// Only includes fields commonly overridden in tests.
type ServerConfig struct {
	// Server
	Port string // default: "19867"

	// Chrome
	Headless          bool   // default: true
	Stealth           string // default: "light"
	MaxTabs           int    // default: 0 (uses config default: 20)
	TabEvictionPolicy string // default: "" (uses config default: "reject")

	// Security - all default to true for tests (unlike production defaults)
	AllowEvaluate   bool // default: true (tests need this)
	AllowMacro      bool // default: false
	AllowScreencast bool // default: false
	AllowDownload   bool // default: false
	AllowUpload     bool // default: false

	// IDPI configuration applied to security.idpi in config.json.
	// Default zero value disables all IDPI checks.
	IDPI appconfig.IDPIConfig

	// Orchestrator
	Strategy         string // default: "" (uses config default: "simple")
	AllocationPolicy string // default: "" (uses config default: "fcfs")
}

func DefaultConfig() ServerConfig {
	port := os.Getenv("PINCHTAB_TEST_PORT")
	if port == "" {
		port = "19867"
	}
	return ServerConfig{
		Port:            port,
		Headless:        true,
		Stealth:         "light",
		AllowEvaluate:   true, // tests need evaluate for assertions
		AllowDownload:   true, // tests validate download error handling
		AllowUpload:     true, // tests validate upload error handling
		AllowScreencast: true, // orchestrator uses /screencast/tabs to fetch tabs
	}
}

type Server struct {
	URL        string
	Dir        string // root temp dir (binary, state, profiles)
	BinaryPath string
	StateDir   string
	ProfileDir string
	cmd        *exec.Cmd
}

// NewTestServer is the preferred way to create a test server. Cleanup runs
// automatically via t.Cleanup, even on panic. Use StartServer for TestMain
// where there's no *testing.T.
func NewTestServer(t *testing.T, cfg ServerConfig) *Server {
	t.Helper()
	srv, err := StartServer(cfg)
	if err != nil {
		t.Fatalf("start test server: %v", err)
	}
	t.Cleanup(srv.Stop)
	return srv
}

// StartServer builds, launches, and waits for health. Caller must call Stop().
func StartServer(cfg ServerConfig) (*Server, error) {
	testDir, err := os.MkdirTemp("", "pinchtab-test-*")
	if err != nil {
		return nil, fmt.Errorf("create test dir: %w", err)
	}
	fmt.Fprintf(os.Stderr, "testutil: test dir: %s\n", testDir)

	s := &Server{
		URL:        fmt.Sprintf("http://localhost:%s", cfg.Port),
		Dir:        testDir,
		BinaryPath: filepath.Join(testDir, "pinchtab"),
		StateDir:   filepath.Join(testDir, "state"),
		ProfileDir: filepath.Join(testDir, "profiles"),
	}

	for _, d := range []string{s.StateDir, s.ProfileDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			s.Cleanup()
			return nil, fmt.Errorf("create %s: %w", d, err)
		}
	}

	if err := writeServerConfig(filepath.Join(s.Dir, "config.json"), s, cfg); err != nil {
		s.Cleanup()
		return nil, fmt.Errorf("write config: %w", err)
	}

	build := exec.Command("go", "build", "-o", s.BinaryPath, "./cmd/pinchtab/") // #nosec G204 -- BinaryPath is from os.MkdirTemp, not user input
	build.Dir = FindRepoRoot()
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		s.Cleanup()
		return nil, fmt.Errorf("build pinchtab: %w", err)
	}

	// Strip existing BRIDGE_*/PINCHTAB_* to avoid test pollution from host config
	env := filterEnv(os.Environ(), "BRIDGE_", "PINCHTAB_")

	// Keep env limited to process wiring. Runtime behavior now comes from config.json.
	env = append(env,
		"PINCHTAB_PORT="+cfg.Port,
		"PINCHTAB_CONFIG="+filepath.Join(s.Dir, "config.json"), // Isolate from host config
	)

	// Chrome binary from host (for CI environments)
	if bin := os.Getenv("CHROME_BIN"); bin != "" {
		env = append(env, "CHROME_BIN="+bin)
	} else if legacyBin := os.Getenv("CHROME_BINARY"); legacyBin != "" {
		env = append(env, "CHROME_BIN="+legacyBin)
	}

	s.cmd = exec.Command(s.BinaryPath) // #nosec G204 -- BinaryPath is from os.MkdirTemp, not user input
	s.cmd.Env = env
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr
	s.cmd.SysProcAttr = processGroupAttr()

	if err := s.cmd.Start(); err != nil {
		s.Cleanup()
		return nil, fmt.Errorf("start pinchtab: %w", err)
	}

	timeout := HealthTimeout
	if os.Getenv("CI") == "true" {
		timeout = HealthTimeoutCI
	}
	if !WaitForHealth(s.URL, timeout) {
		s.Stop()
		return nil, fmt.Errorf("pinchtab did not become healthy within %v", timeout)
	}

	return s, nil
}

func writeServerConfig(path string, srv *Server, cfg ServerConfig) error {
	fc := appconfig.DefaultFileConfig()
	fc.Server.Port = cfg.Port
	fc.Server.StateDir = srv.StateDir
	fc.Profiles.BaseDir = srv.ProfileDir
	fc.Profiles.DefaultProfile = "default"

	fc.InstanceDefaults.Mode = modeString(cfg.Headless)
	noRestore := true
	fc.InstanceDefaults.NoRestore = &noRestore
	fc.InstanceDefaults.StealthLevel = cfg.Stealth
	if cfg.MaxTabs > 0 {
		maxTabs := cfg.MaxTabs
		fc.InstanceDefaults.MaxTabs = &maxTabs
	}
	if cfg.TabEvictionPolicy != "" {
		fc.InstanceDefaults.TabEvictionPolicy = cfg.TabEvictionPolicy
	}

	allowEvaluate := cfg.AllowEvaluate
	allowMacro := cfg.AllowMacro
	allowScreencast := cfg.AllowScreencast
	allowDownload := cfg.AllowDownload
	allowUpload := cfg.AllowUpload
	fc.Security.AllowEvaluate = &allowEvaluate
	fc.Security.AllowMacro = &allowMacro
	fc.Security.AllowScreencast = &allowScreencast
	fc.Security.AllowDownload = &allowDownload
	fc.Security.AllowUpload = &allowUpload
	fc.Security.IDPI = cfg.IDPI

	if cfg.Strategy != "" {
		fc.MultiInstance.Strategy = cfg.Strategy
	}
	if cfg.AllocationPolicy != "" {
		fc.MultiInstance.AllocationPolicy = cfg.AllocationPolicy
	}

	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func modeString(headless bool) string {
	if headless {
		return "headless"
	}
	return "headed"
}

func (s *Server) Stop() {
	if s.cmd != nil && s.cmd.Process != nil {
		TerminateProcessGroup(s.cmd, ShutdownTimeout)
	}
	s.Cleanup()
}

// Cleanup removes the test directory unless PINCHTAB_TEST_KEEP_DIR is set.
func (s *Server) Cleanup() {
	if os.Getenv("PINCHTAB_TEST_KEEP_DIR") != "" {
		fmt.Fprintf(os.Stderr, "testutil: keeping test dir (PINCHTAB_TEST_KEEP_DIR set): %s\n", s.Dir)
		return
	}
	_ = os.RemoveAll(s.Dir)
}

func WaitForHealth(base string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "GET", base+"/health", nil)
		if err != nil {
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			healthy := resp.StatusCode == 200
			_ = resp.Body.Close()
			if healthy {
				return true
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// FindRepoRoot walks up from cwd looking for go.mod.
func FindRepoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Join("..", "..")
}

// LaunchInstance creates an instance and polls until it can navigate.
func LaunchInstance(base string) (string, error) {
	resp, err := http.Post(
		base+"/instances/launch",
		"application/json",
		strings.NewReader(`{"mode":"headless"}`),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("launch failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse launch response: %w", err)
	}

	id, ok := result["id"].(string)
	if !ok {
		return "", fmt.Errorf("no instance id in launch response: %v", result)
	}

	fmt.Fprintf(os.Stderr, "testutil: launched instance %s\n", id)

	ctx, cancel := context.WithTimeout(context.Background(), InstanceTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("instance %s did not become ready within %v", id, InstanceTimeout)
		default:
		}

		openResp, err := http.Post(
			base+"/instances/"+id+"/tabs/open",
			"application/json",
			strings.NewReader(`{"url":"about:blank"}`),
		)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var tabID string
		if openResp.StatusCode == 200 {
			openBody, _ := io.ReadAll(openResp.Body)
			var open map[string]any
			if err := json.Unmarshal(openBody, &open); err == nil {
				if v, ok := open["tabId"].(string); ok {
					tabID = v
				}
			}
		}
		_ = openResp.Body.Close()

		if tabID == "" {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		navResp, err := http.Post(
			base+"/tabs/"+tabID+"/navigate",
			"application/json",
			strings.NewReader(`{"url":"about:blank"}`),
		)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		status := navResp.StatusCode
		_ = navResp.Body.Close()

		if status == 200 {
			fmt.Fprintf(os.Stderr, "testutil: instance %s is ready\n", id)
			return id, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func filterEnv(env []string, prefixes ...string) []string {
	out := make([]string, 0, len(env))
	for _, e := range env {
		skip := false
		for _, p := range prefixes {
			if strings.HasPrefix(e, p) {
				skip = true
				break
			}
		}
		if !skip {
			out = append(out, e)
		}
	}
	return out
}
