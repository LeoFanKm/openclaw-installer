package project

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// DevServer manages the lifecycle of an npm dev server process.
type DevServer struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	logCh   chan string
	running bool
	ready   bool
	readyURL string
}

// NewDevServer creates a new DevServer instance.
func NewDevServer() *DevServer {
	return &DevServer{
		logCh: make(chan string, 256),
	}
}

// Start launches "npm run dev" in the given project directory.
// It monitors output for Vite's "Local:" line to detect readiness.
func (d *DevServer) Start(projectDir string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return fmt.Errorf("dev server is already running")
	}

	npmCmd := "npm"
	if runtime.GOOS == "windows" {
		npmCmd = "npm.cmd"
	}

	cmd := exec.Command(npmCmd, "run", "dev")
	cmd.Dir = projectDir
	setProcessGroup(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("pipe stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start dev server: %w", err)
	}

	d.cmd = cmd
	d.running = true
	d.ready = false

	// Stream output in background goroutines.
	streamFn := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 8192), 8192)
		for scanner.Scan() {
			line := scanner.Text()
			// Detect Vite readiness.
			if strings.Contains(line, "Local:") && strings.Contains(line, "http://localhost:") {
				d.mu.Lock()
				d.ready = true
				// Extract URL.
				if idx := strings.Index(line, "http://localhost:"); idx >= 0 {
					d.readyURL = strings.TrimSpace(line[idx:])
				}
				d.mu.Unlock()
			}
			// Non-blocking send to log channel.
			select {
			case d.logCh <- line:
			default:
				// Drop if channel is full (consumer too slow).
			}
		}
	}

	go streamFn(stdout)
	go streamFn(stderr)

	// Monitor process exit.
	go func() {
		_ = cmd.Wait()
		d.mu.Lock()
		d.running = false
		d.ready = false
		d.mu.Unlock()
		select {
		case d.logCh <- "[dev server exited]":
		default:
		}
	}()

	return nil
}

// Stop kills the dev server process and its child processes.
func (d *DevServer) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running || d.cmd == nil || d.cmd.Process == nil {
		return nil
	}

	err := killProcessTree(d.cmd)
	d.running = false
	d.ready = false
	return err
}

// Logs returns a read-only channel that receives dev server log lines.
func (d *DevServer) Logs() <-chan string {
	return d.logCh
}

// IsRunning reports whether the dev server process is currently running.
func (d *DevServer) IsRunning() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.running
}

// IsReady reports whether Vite has printed its "Local:" URL,
// indicating the dev server is accepting connections.
func (d *DevServer) IsReady() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.ready
}

// ReadyURL returns the local URL reported by Vite, or "" if not yet ready.
func (d *DevServer) ReadyURL() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.readyURL
}
