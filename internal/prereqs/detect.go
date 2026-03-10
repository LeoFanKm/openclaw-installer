package prereqs

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// DetectGit checks whether git is installed and returns its version.
func DetectGit() (installed bool, version string, err error) {
	path, err := exec.LookPath("git")
	if err != nil {
		return false, "", nil
	}
	_ = path

	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		return false, "", fmt.Errorf("git found but version check failed: %w", err)
	}

	version = strings.TrimSpace(string(out))
	// Extract version number from "git version 2.43.0" or similar.
	if parts := strings.Fields(version); len(parts) >= 3 {
		version = parts[2]
	}
	return true, version, nil
}

// DetectNode checks whether Node.js is installed, returns its version,
// and verifies the major version is >= 18.
func DetectNode() (installed bool, version string, err error) {
	_, err = exec.LookPath("node")
	if err != nil {
		return false, "", nil
	}

	out, err := exec.Command("node", "--version").Output()
	if err != nil {
		return false, "", fmt.Errorf("node found but version check failed: %w", err)
	}

	version = strings.TrimSpace(string(out))
	// version is like "v20.11.0"
	raw := strings.TrimPrefix(version, "v")

	major, parseErr := parseMajorVersion(raw)
	if parseErr != nil {
		return true, raw, fmt.Errorf("could not parse node version %q: %w", raw, parseErr)
	}
	if major < 18 {
		return true, raw, fmt.Errorf("node version %s is below minimum required v18", raw)
	}

	return true, raw, nil
}

// DetectNpm checks whether npm is installed and returns its version.
func DetectNpm() (installed bool, version string, err error) {
	_, err = exec.LookPath("npm")
	if err != nil {
		return false, "", nil
	}

	out, err := exec.Command("npm", "--version").Output()
	if err != nil {
		return false, "", fmt.Errorf("npm found but version check failed: %w", err)
	}

	version = strings.TrimSpace(string(out))
	return true, version, nil
}

var majorVersionRe = regexp.MustCompile(`^(\d+)`)

func parseMajorVersion(v string) (int, error) {
	m := majorVersionRe.FindStringSubmatch(v)
	if len(m) < 2 {
		return 0, fmt.Errorf("no major version found in %q", v)
	}
	return strconv.Atoi(m[1])
}
