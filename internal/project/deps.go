package project

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

// InstallDeps runs "npm install" in the given project directory and
// streams stdout and stderr to progressCh.
func InstallDeps(projectDir string, progressCh chan<- string) error {
	progressCh <- "Running npm install..."

	npmCmd := "npm"
	if runtime.GOOS == "windows" {
		npmCmd = "npm.cmd"
	}

	cmd := exec.Command(npmCmd, "install")
	cmd.Dir = projectDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("pipe stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start npm install: %w", err)
	}

	// Merge stdout and stderr into progressCh.
	done := make(chan struct{}, 2)
	streamLines := func(r io.Reader) {
		defer func() { done <- struct{}{} }()
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 8192), 8192)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				progressCh <- line
			}
		}
	}

	go streamLines(stdout)
	go streamLines(stderr)

	// Wait for both readers to finish.
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	progressCh <- "npm install complete"
	return nil
}
