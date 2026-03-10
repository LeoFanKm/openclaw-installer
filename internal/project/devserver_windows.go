//go:build windows

package project

import (
	"fmt"
	"os/exec"
	"strconv"
)

// setProcessGroup is a no-op on Windows; process groups are handled
// differently via taskkill.
func setProcessGroup(cmd *exec.Cmd) {
	// No SysProcAttr needed on Windows for this use case.
}

// killProcessTree kills the process and all its children on Windows
// using taskkill /T /F.
func killProcessTree(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	pid := strconv.Itoa(cmd.Process.Pid)
	kill := exec.Command("taskkill", "/t", "/f", "/pid", pid)
	if output, err := kill.CombinedOutput(); err != nil {
		return fmt.Errorf("taskkill failed: %w\noutput: %s", err, string(output))
	}
	return nil
}
