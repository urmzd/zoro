package subprocess

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

const surrealVersion = "v3.0.4"

// SurrealProcess manages a SurrealDB subprocess.
type SurrealProcess struct {
	cmd  *exec.Cmd
	port int
}

// URL returns the WebSocket URL for the managed SurrealDB instance.
func (p *SurrealProcess) URL() string {
	return fmt.Sprintf("ws://127.0.0.1:%d", p.port)
}

// Stop terminates the SurrealDB process.
func (p *SurrealProcess) Stop() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	log.Println("[surrealdb] stopping")
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

// StartSurreal downloads the SurrealDB binary if needed and starts it.
func StartSurreal(ctx context.Context, dataDir string, port int) (*SurrealProcess, error) {
	binDir := filepath.Join(dataDir, "bin")
	binPath, err := ensureSurrealBinary(binDir)
	if err != nil {
		return nil, fmt.Errorf("ensure surreal binary: %w", err)
	}

	dbDir := filepath.Join(dataDir, "data")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dbDir, "zoro.db")
	bind := fmt.Sprintf("127.0.0.1:%d", port)

	cmd := exec.CommandContext(ctx, binPath,
		"start",
		"--user", "root",
		"--pass", "root",
		"--bind", bind,
		"surrealkv://"+dbPath,
	)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	log.Printf("[surrealdb] starting on port %d (data: %s)", port, dbPath)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start surreal: %w", err)
	}

	if err := waitForHealth(fmt.Sprintf("http://127.0.0.1:%d/health", port), 30*time.Second); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("surreal not ready: %w", err)
	}

	log.Println("[surrealdb] ready")
	return &SurrealProcess{cmd: cmd, port: port}, nil
}

func ensureSurrealBinary(binDir string) (string, error) {
	name := "surreal"
	if runtime.GOOS == "windows" {
		name = "surreal.exe"
	}
	binPath := filepath.Join(binDir, name)

	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	log.Println("[surrealdb] downloading binary...")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return "", err
	}

	url := surrealDownloadURL()
	log.Printf("[surrealdb] downloading from %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned %d", resp.StatusCode)
	}

	if runtime.GOOS == "windows" {
		// Windows: direct .exe download
		f, err := os.OpenFile(binPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return "", err
		}
		defer f.Close()
		if _, err := io.Copy(f, resp.Body); err != nil {
			return "", fmt.Errorf("write binary: %w", err)
		}
	} else {
		// Unix: .tgz archive containing the binary
		if err := extractTarGz(resp.Body, binDir, "surreal"); err != nil {
			return "", fmt.Errorf("extract: %w", err)
		}
		os.Chmod(binPath, 0o755)
	}

	log.Println("[surrealdb] binary installed")
	return binPath, nil
}

func surrealDownloadURL() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch names to SurrealDB release names
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	}

	if osName == "windows" {
		return fmt.Sprintf("https://download.surrealdb.com/%s/surreal-%s.windows-%s.exe", surrealVersion, surrealVersion, arch)
	}
	return fmt.Sprintf("https://download.surrealdb.com/%s/surreal-%s.%s-%s.tgz", surrealVersion, surrealVersion, osName, arch)
}

func extractTarGz(r io.Reader, destDir string, targetName string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if filepath.Base(hdr.Name) == targetName && hdr.Typeflag == tar.TypeReg {
			outPath := filepath.Join(destDir, targetName)
			f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
			return nil
		}
	}
	return fmt.Errorf("binary %q not found in archive", targetName)
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
			log.Printf("[surrealdb] waiting for health (attempt %d)...", attempt)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout after %s (%d attempts, url: %s)", timeout, attempt, strconv.Quote(url))
}
