package subprocess

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// SearXNGProcess manages a SearXNG subprocess running from a local Python venv.
type SearXNGProcess struct {
	cmd  *exec.Cmd
	port int
}

// URL returns the HTTP URL for the managed SearXNG instance.
func (p *SearXNGProcess) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", p.port)
}

// Stop terminates the SearXNG process.
func (p *SearXNGProcess) Stop() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	log.Println("[searxng] stopping")
	_ = p.cmd.Process.Signal(os.Interrupt)
	done := make(chan error, 1)
	go func() { done <- p.cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		p.cmd.Process.Kill()
	}
	return nil
}

// StartSearXNG ensures a Python venv with SearXNG is set up, then starts it.
func StartSearXNG(ctx context.Context, dataDir string, port int, settingsPath string) (*SearXNGProcess, error) {
	venvDir := filepath.Join(dataDir, "searxng-venv")

	if err := ensureSearXNGVenv(venvDir); err != nil {
		return nil, fmt.Errorf("setup searxng: %w", err)
	}

	pythonBin := filepath.Join(venvDir, "bin", "python")

	// Inline script that starts SearXNG via Flask
	script := fmt.Sprintf(`
import os, sys
os.environ["SEARXNG_SETTINGS_PATH"] = %q
from searx.webapp import app
app.run(host="127.0.0.1", port=%d)
`, settingsPath, port)

	cmd := exec.CommandContext(ctx, pythonBin, "-c", script)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	log.Printf("[searxng] starting on port %d", port)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start searxng: %w", err)
	}

	if err := waitForHealth(fmt.Sprintf("http://127.0.0.1:%d", port), 60*time.Second); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("searxng not ready: %w", err)
	}

	log.Println("[searxng] ready")
	return &SearXNGProcess{cmd: cmd, port: port}, nil
}

func ensureSearXNGVenv(venvDir string) error {
	pythonBin := filepath.Join(venvDir, "bin", "python")
	// Check if venv exists and has searx installed
	if _, err := os.Stat(pythonBin); err == nil {
		check := exec.Command(pythonBin, "-c", "import searx")
		if check.Run() == nil {
			return nil // already set up
		}
	}

	log.Println("[searxng] creating python venv and installing searxng (first run)...")

	// Find system python3
	python3, err := exec.LookPath("python3")
	if err != nil {
		return fmt.Errorf("python3 not found: %w", err)
	}

	// Create venv
	if err := os.MkdirAll(filepath.Dir(venvDir), 0o755); err != nil {
		return err
	}
	if out, err := exec.Command(python3, "-m", "venv", venvDir).CombinedOutput(); err != nil {
		return fmt.Errorf("create venv: %w\n%s", err, out)
	}

	pip := filepath.Join(venvDir, "bin", "pip")

	// Install build deps first
	log.Println("[searxng] installing build dependencies...")
	depCmd := exec.Command(pip, "install", "--quiet", "setuptools", "wheel", "msgspec", "typing_extensions")
	depCmd.Stdout = os.Stderr
	depCmd.Stderr = os.Stderr
	if err := depCmd.Run(); err != nil {
		os.RemoveAll(venvDir)
		return fmt.Errorf("pip install setuptools: %w", err)
	}

	// Install SearXNG
	log.Println("[searxng] pip installing searxng...")
	cmd := exec.Command(pip, "install", "--quiet", "--no-build-isolation",
		"git+https://github.com/searxng/searxng.git@master")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Clean up broken venv
		os.RemoveAll(venvDir)
		return fmt.Errorf("pip install searxng: %w", err)
	}

	log.Println("[searxng] installed")
	return nil
}

func waitForHealth(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		if attempt%10 == 0 {
			log.Printf("[health] waiting (attempt %d, url: %s)...", attempt, url)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout after %s (%d attempts, url: %s)", timeout, attempt, strconv.Quote(url))
}
