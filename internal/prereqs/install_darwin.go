//go:build darwin

package prereqs

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	nodeLTSMacURL = "https://nodejs.org/dist/v22.13.1/node-v22.13.1.pkg"
)

// InstallGit triggers Xcode Command Line Tools installation on macOS,
// which provides git.
func InstallGit(progressCh chan<- string) error {
	progressCh <- "Installing Xcode Command Line Tools (includes Git)..."

	cmd := exec.Command("xcode-select", "--install")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Error code 1 means already installed.
		if strings.Contains(string(output), "already installed") {
			progressCh <- "Xcode Command Line Tools already installed"
			return nil
		}
		return fmt.Errorf("xcode-select --install failed: %w\noutput: %s", err, string(output))
	}

	progressCh <- "Xcode Command Line Tools installation triggered. A system dialog may appear — please follow the prompts."
	return nil
}

// InstallNode downloads and installs Node.js LTS from the official .pkg installer.
func InstallNode(progressCh chan<- string) error {
	progressCh <- "Downloading Node.js LTS..."

	tmpDir, err := os.MkdirTemp("", "openclaw-node-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pkgPath := filepath.Join(tmpDir, "node-lts.pkg")
	if err := downloadFile(nodeLTSMacURL, pkgPath, progressCh); err != nil {
		return fmt.Errorf("download node: %w", err)
	}

	progressCh <- "Installing Node.js (may require admin password)..."
	cmd := exec.Command("sudo", "installer", "-pkg", pkgPath, "-target", "/")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("node install failed: %w\noutput: %s", err, string(output))
	}

	progressCh <- "Node.js installed successfully"
	return nil
}

// RefreshPath re-reads PATH from /etc/paths and common shell profile files
// so that newly installed tools are immediately available.
func RefreshPath() error {
	var parts []string

	// Read /etc/paths
	if f, err := os.Open("/etc/paths"); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				parts = append(parts, line)
			}
		}
		f.Close()
	}

	// Read /etc/paths.d/*
	if entries, err := os.ReadDir("/etc/paths.d"); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join("/etc/paths.d", entry.Name()))
			if err != nil {
				continue
			}
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					parts = append(parts, line)
				}
			}
		}
	}

	// Common additional paths for Homebrew, nvm, etc.
	home, _ := os.UserHomeDir()
	extras := []string{
		"/usr/local/bin",
		"/opt/homebrew/bin",
		filepath.Join(home, ".nvm/versions/node"),
	}
	for _, p := range extras {
		if _, err := os.Stat(p); err == nil {
			parts = append(parts, p)
		}
	}

	if len(parts) > 0 {
		existing := os.Getenv("PATH")
		combined := strings.Join(parts, string(os.PathListSeparator))
		if existing != "" {
			combined = combined + string(os.PathListSeparator) + existing
		}
		os.Setenv("PATH", combined)
	}
	return nil
}

func downloadFile(url, dest string, progressCh chan<- string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	totalBytes := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)
	lastPercent := -1

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := f.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("write file: %w", writeErr)
			}
			downloaded += int64(n)
			if totalBytes > 0 {
				pct := int(downloaded * 100 / totalBytes)
				if pct != lastPercent && pct%5 == 0 {
					lastPercent = pct
					progressCh <- fmt.Sprintf("Downloading... %d%%", pct)
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("read response: %w", readErr)
		}
	}

	progressCh <- "Download complete"
	return nil
}
