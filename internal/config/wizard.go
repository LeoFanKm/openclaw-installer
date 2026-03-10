package config

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SaveConfig writes configuration key-value pairs to a .dev.vars file
// in the given project directory.
func SaveConfig(projectDir string, cfg map[string]string) error {
	filePath := filepath.Join(projectDir, ".dev.vars")

	// Sort keys for deterministic output.
	keys := make([]string, 0, len(cfg))
	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(cfg[k])
		sb.WriteString("\n")
	}

	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return fmt.Errorf("create project dir: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(sb.String()), 0o600); err != nil {
		return fmt.Errorf("write .dev.vars: %w", err)
	}

	return nil
}

// LoadExistingConfig reads an existing .dev.vars file and returns
// its key-value pairs. Returns an empty map (not an error) if the
// file does not exist.
func LoadExistingConfig(projectDir string) (map[string]string, error) {
	filePath := filepath.Join(projectDir, ".dev.vars")

	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("open .dev.vars: %w", err)
	}
	defer f.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if key != "" {
			result[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read .dev.vars: %w", err)
	}
	return result, nil
}

// TestClerkConnection verifies a Clerk secret key by making a request
// to the Clerk API. Returns true if the key is valid.
func TestClerkConnection(secretKey string) (bool, error) {
	if secretKey == "" {
		return false, fmt.Errorf("secret key is empty")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.clerk.com/v1/clients", nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+secretKey)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("clerk API request failed: %w", err)
	}
	defer resp.Body.Close()

	// 200 or 401 both confirm the API is reachable;
	// only 200 means the key is valid.
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return false, nil
	}

	return false, fmt.Errorf("unexpected clerk API status: %d", resp.StatusCode)
}

// GenerateJWTSecret generates a cryptographically random 32-byte secret
// and returns it as a hex-encoded string.
func GenerateJWTSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// This should never fail on a modern OS, but just in case
		// return a deterministic fallback is worse than panicking.
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b)
}
