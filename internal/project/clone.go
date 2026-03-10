package project

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

const DefaultRepoURL = "https://github.com/nicekid1/OpenClaw.git"

// CloneRepo clones the given git repository into targetDir.
// Progress messages from git are streamed to progressCh.
func CloneRepo(repoURL, targetDir string, progressCh chan<- string) error {
	if repoURL == "" {
		repoURL = DefaultRepoURL
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	progressCh <- fmt.Sprintf("Cloning %s into %s...", repoURL, targetDir)

	cmd := exec.Command("git", "clone", "--progress", repoURL, targetDir)
	// Git writes progress to stderr.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("pipe stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start git clone: %w", err)
	}

	scanner := bufio.NewScanner(stderr)
	// Git progress lines can be long and use \r for in-place updates.
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
