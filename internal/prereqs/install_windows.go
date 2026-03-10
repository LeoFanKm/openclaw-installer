//go:build windows

package prereqs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	gitWindowsURL = "https://github.com/git-for-windows/git/releases/download/v2.47.1.windows.1/Git-2.47.1-64-bit.exe"
	nodeLTSURL    = "https://nodejs.org/dist/v22.13.1/node-v22.13.1-x64.msi"
)

// InstallGit downloads and silently installs Git for Windows.
func InstallGit(progressCh chan<- string) error {
	progressCh <- "Downloading Git for Windows..."

	tmpDir, err := os.MkdirTemp("", "openclaw-git-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	installerPath := filepath.Join(tmpDir, "git-installer.exe")
	if err := downloadFile(gitWindowsURL, installerPath, progressCh); err != nil {
		return fmt.Errorf("download git: %w", err)
	}

	progressCh <- "Installing Git (this may take a minute)..."
	cmd := exec.Command(installerPath, "/VERYSILENT", "/NORESTART", "/NOCANCEL", "/SP-",
		"/CLOSEAPPLICATIONS", "/RESTARTAPPLICATIONS",
		"/COMPONENTS=icons,ext\\reg\\shellhere,assoc,assoc_sh")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git install failed: %w\noutput: %s", err, string(output))
	}

	progressCh <- "Git installed successfully"
	return nil
}

// InstallNode downloads and silently installs Node.js LTS on Windows.
func InstallNode(progressCh chan<- string) error {
	progressCh <- "Downloading Node.js LTS..."

	tmpDir, err := os.MkdirTemp("", "openclaw-node-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	msiPath := filepath.Join(tmpDir, "node-lts.msi")
	if err := downloadFile(nodeLTSURL, msiPath, progressCh); err != nil {
		return fmt.Errorf("download node: %w", err)
	}

	progressCh <- "Installing Node.js (this may take a minute)..."
	cmd := exec.Command("msiexec", "/i", msiPath, "/qn", "/norestart")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("node install failed: %w\noutput: %s", err, string(output))
	}

	progressCh <- "Node.js installed successfully"
	return nil
}

// RefreshPath reads the current PATH from the Windows registry using the
// reg query command (no external dependencies) and updates the process
// environment so newly installed tools are immediately available.
func RefreshPath() error {
	var parts []string

	// System PATH
	if sysPath := queryRegistryPath(
		`HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		"Path",
	); sysPath != "" {
		parts = append(parts, splitPath(sysPath)...)
	}

	// User PATH
	if userPath := queryRegistryPath(
		`HKCU\Environment`,
		"Path",
	); userPath != "" {
		parts = append(parts, splitPath(userPath)...)
	}

	if len(parts) > 0 {
		newPath := strings.Join(parts, string(os.PathListSeparator))
		os.Setenv("PATH", newPath)
	}
	return nil
}

// queryRegistryPath uses "reg query" to read a string value from the registry.
// Returns empty string on any error.
func queryRegistryPath(keyPath, valueName string) string {
	out, err := exec.Command("reg", "query", keyPath, "/v", valueName).Output()
	if err != nil {
		return ""
	}
	// Output format:
	//   HKEY_...\Environment
	//       Path    REG_SZ    C:\Windows;C:\...
	// or    Path    REG_EXPAND_SZ    %SystemRoot%\...
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for lines containing REG_SZ or REG_EXPAND_SZ
		for _, regType := range []string{"REG_EXPAND_SZ", "REG_SZ"} {
			idx := strings.Index(line, regType)
			if idx >= 0 {
				value := strings.TrimSpace(line[idx+len(regType):])
				// Expand environment variables in REG_EXPAND_SZ values.
				return os.ExpandEnv(value)
			}
		}
	}
	return ""
}

func splitPath(p string) []string {
	var result []string
	for _, s := range strings.Split(p, string(os.PathListSeparator)) {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
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
