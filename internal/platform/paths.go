package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

// HomeDir returns the user's home directory.
func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback for edge cases.
		if runtime.GOOS == "windows" {
			return os.Getenv("USERPROFILE")
		}
		return os.Getenv("HOME")
	}
	return home
}

// DefaultProjectDir returns the default directory for OpenClaw projects.
// On macOS/Linux: ~/Documents/OpenClaw
// On Windows: %USERPROFILE%\Documents\OpenClaw
func DefaultProjectDir() string {
	return filepath.Join(HomeDir(), "Documents", "OpenClaw")
}
