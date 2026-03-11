package project

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

// InstallDeps installs project dependencies and
// streams stdout and stderr to progressCh.
func InstallDeps(projectDir string, progressCh chan<- string) error {
	var (
		cmdName string
		args    []string
		label   string
	)

	if isOfficialOpenClawRepo(projectDir) {
		pnpmCmd, pnpmPrefix, err := resolvePnpmCommand()
		if err != nil {
			return err
		}
		cmdName = pnpmCmd
		args = append(pnpmPrefix, "install")
		label = "pnpm install"
	} else {
		cmdName = resolveNpmCommand()
		args = []string{"install"}
		label = "npm install"
	}

	progressCh <- fmt.Sprintf("Running %s...", label)

	cmd := exec.Command(cmdName, args...)
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
		return fmt.Errorf("start %s: %w", label, err)
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
		return fmt.Errorf("%s failed: %w", label, err)
	}

	progressCh <- fmt.Sprintf("%s complete", label)
	return nil
}
