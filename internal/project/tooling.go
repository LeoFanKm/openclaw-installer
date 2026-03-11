package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func isOfficialOpenClawRepo(projectDir string) bool {
	if _, err := os.Stat(filepath.Join(projectDir, "ui", "package.json")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(projectDir, "scripts", "ui.js")); err == nil {
		return true
	}
	return false
}

func resolveNpmCommand() string {
	if runtime.GOOS == "windows" {
		return "npm.cmd"
	}
	return "npm"
}

func resolveNodeCommand() string {
	if runtime.GOOS == "windows" {
		return "node.exe"
	}
	return "node"
}

func resolvePnpmCommand() (string, []string, error) {
	candidates := []struct {
		cmd  string
		args []string
	}{
		{cmd: "pnpm"},
		{cmd: "corepack", args: []string{"pnpm"}},
	}

	if runtime.GOOS == "windows" {
		candidates = []struct {
			cmd  string
			args []string
		}{
			{cmd: "pnpm.cmd"},
			{cmd: "pnpm"},
			{cmd: "corepack.cmd", args: []string{"pnpm"}},
			{cmd: "corepack", args: []string{"pnpm"}},
		}
	}

	for _, candidate := range candidates {
		if _, err := exec.LookPath(candidate.cmd); err == nil {
			return candidate.cmd, candidate.args, nil
		}
	}

	return "", nil, fmt.Errorf("pnpm not found; install pnpm or enable corepack first")
}
