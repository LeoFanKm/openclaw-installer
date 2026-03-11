package project

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const DefaultRepoURL = "https://github.com/openclaw/openclaw.git"

// GitHub zip archive URL for the default branch.
const DefaultZipURL = "https://github.com/openclaw/openclaw/archive/refs/heads/master.zip"

// CloneRepo downloads the project into targetDir.
// It first tries a fast zip download, falling back to git clone.
func CloneRepo(repoURL, targetDir string, progressCh chan<- string) error {
	if repoURL == "" || repoURL == DefaultRepoURL {
		// Use zip download for the default repo — much faster.
		return downloadZip(DefaultZipURL, targetDir, progressCh)
	}
	return gitClone(repoURL, targetDir, progressCh)
}

// downloadZip fetches a GitHub archive zip and extracts it to targetDir.
func downloadZip(zipURL, targetDir string, progressCh chan<- string) error {
	progressCh <- "Downloading project archive..."

	// Create a temp file for the zip.
	tmpFile, err := os.CreateTemp("", "openclaw-*.zip")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Download.
	resp, err := http.Get(zipURL)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		tmpFile.Close()
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)
	lastPercent := -1

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := tmpFile.Write(buf[:n]); wErr != nil {
				tmpFile.Close()
				return fmt.Errorf("write temp file: %w", wErr)
			}
			downloaded += int64(n)
			if totalSize > 0 {
				pct := int(downloaded * 100 / totalSize)
				if pct != lastPercent && pct%10 == 0 {
					progressCh <- fmt.Sprintf("Downloaded %d%%", pct)
					lastPercent = pct
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			tmpFile.Close()
			return fmt.Errorf("download read error: %w", readErr)
		}
	}
	tmpFile.Close()

	progressCh <- fmt.Sprintf("Download complete (%d MB), extracting...", downloaded/(1024*1024))

	// Extract zip.
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	r, err := zip.OpenReader(tmpPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	// GitHub zips have a top-level directory like "ClaudingBot-master/".
	// We strip it so files go directly into targetDir.
	var stripPrefix string
	if len(r.File) > 0 {
		first := r.File[0].Name
		if idx := strings.Index(first, "/"); idx >= 0 {
			stripPrefix = first[:idx+1]
		}
	}

	fileCount := 0
	for _, f := range r.File {
		name := f.Name
		if stripPrefix != "" {
			name = strings.TrimPrefix(name, stripPrefix)
		}
		if name == "" {
			continue
		}

		destPath := filepath.Join(targetDir, filepath.FromSlash(name))

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0o755)
			continue
		}

		// Ensure parent dir exists.
		os.MkdirAll(filepath.Dir(destPath), 0o755)

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("create file %s: %w", name, err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("open zip entry %s: %w", name, err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return fmt.Errorf("extract %s: %w", name, err)
		}
		fileCount++
	}

	progressCh <- fmt.Sprintf("Extracted %d files", fileCount)

	// Initialize as a git repo so future git operations work.
	progressCh <- "Initializing git repository..."
	cmd := exec.Command("git", "init")
	cmd.Dir = targetDir
	if out, err := cmd.CombinedOutput(); err != nil {
		progressCh <- fmt.Sprintf("WARNING: git init failed: %s", string(out))
	}

	progressCh <- "Clone complete"
	return nil
}

// gitClone uses git to clone a repository (fallback for custom URLs).
func gitClone(repoURL, targetDir string, progressCh chan<- string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	progressCh <- fmt.Sprintf("Cloning %s into %s...", repoURL, targetDir)

	cmd := exec.Command("git", "clone", "--depth", "1", "--progress", repoURL, targetDir)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("pipe stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start git clone: %w", err)
	}

	scanner := bufio.NewScanner(stderr)
	scanner.Buffer(make([]byte, 0, 4096), 4096)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			progressCh <- line
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	progressCh <- "Clone complete"
	return nil
}
