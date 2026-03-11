package project

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// DevServer manages the lifecycle of an npm dev server process.
type DevServer struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	logCh    chan string
	running  bool
	ready    bool
	readyURL string
}

// NewDevServer creates a new DevServer instance.
func NewDevServer() *DevServer {
	return &DevServer{
		logCh: make(chan string, 256),
	}
}

// Start launches the appropriate local UI/dev command in the given project directory.
// It monitors output for Vite's "Local:" line to detect readiness.
func (d *DevServer) Start(projectDir string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return fmt.Errorf("dev server is already running")
	}

	var (
		cmdName string
		args    []string
	)

	if isOfficialOpenClawRepo(projectDir) {
		pnpmCmd, pnpmPrefix, err := resolvePnpmCommand()
		if err != nil {
			return err
		}
		cmdName = pnpmCmd
		args = append(pnpmPrefix, "ui:dev")
	} else {
		cmdName = resolveNpmCommand()
		args = []string{"run", "dev"}
	}

	cmd := exec.Command(cmdName, args...)
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
	d.logCh = make(chan string, 256)
	d.running = true
	d.ready = false
	d.readyURL = ""

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
		logCh := d.logCh
		d.mu.Unlock()
		select {
		case logCh <- "[dev server exited]":
		default:
		}
		close(logCh)
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
