package project

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const officialOpenClawGatewayURL = "http://127.0.0.1:18789/"

// DevServer manages the lifecycle of the local OpenClaw runtime.
type DevServer struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	logCh    chan string
	external bool
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

// Start launches the appropriate local runtime in the given project directory.
func (d *DevServer) Start(projectDir string) error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return fmt.Errorf("dev server is already running")
	}

	d.logCh = make(chan string, 256)
	d.ready = false
	d.readyURL = ""
	logCh := d.logCh
	d.mu.Unlock()

	official := isOfficialOpenClawRepo(projectDir)

	var (
		cmdName string
		args    []string
		urlHint string
	)

	if official {
		if err := d.runOfficialSetup(projectDir, logCh); err != nil {
			close(logCh)
			return err
		}
		resolvedURL, err := resolveOfficialDashboardURL(projectDir)
		if err == nil && resolvedURL != "" {
			urlHint = resolvedURL
		} else {
			urlHint = officialOpenClawGatewayURL
		}
		if isTCPPortListening("127.0.0.1:18789") {
			d.mu.Lock()
			d.cmd = nil
			d.external = true
			d.running = true
			d.ready = true
			d.readyURL = urlHint
			d.mu.Unlock()
			sendLogLine(logCh, "Detected an existing OpenClaw gateway on 127.0.0.1:18789")
			sendLogLine(logCh, fmt.Sprintf("Dashboard URL: %s", urlHint))
			sendLogLine(logCh, "Using existing gateway; Stop will only disconnect the installer UI.")
			return nil
		}
		cmdName = resolveNodeCommand()
		args = []string{"openclaw.mjs", "gateway", "--port", "18789", "--verbose"}
	} else {
		cmdName = resolveNpmCommand()
		args = []string{"run", "dev"}
	}

	cmd := exec.Command(cmdName, args...)
	cmd.Dir = projectDir
	setProcessGroup(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		close(logCh)
		return fmt.Errorf("pipe stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		close(logCh)
		return fmt.Errorf("pipe stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		close(logCh)
		return fmt.Errorf("start dev server: %w", err)
	}

	d.mu.Lock()
	d.cmd = cmd
	d.external = false
	d.running = true
	d.ready = false
	d.readyURL = ""
	d.mu.Unlock()

	var readyOnce sync.Once
	markReady := func(url string) {
		readyOnce.Do(func() {
			d.mu.Lock()
			d.ready = true
			d.readyURL = url
			d.mu.Unlock()
			if url != "" {
				sendLogLine(logCh, fmt.Sprintf("Dashboard URL: %s", url))
			}
		})
	}

	streamDone := make(chan struct{}, 2)
	streamFn := func(r io.Reader) {
		defer func() { streamDone <- struct{}{} }()
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 8192), 8192)
		for scanner.Scan() {
			line := scanner.Text()
			if official {
				if strings.Contains(line, "listening on ws://127.0.0.1:18789") {
					markReady(urlHint)
				}
			} else {
				if strings.Contains(line, "Local:") && strings.Contains(line, "http://localhost:") {
					markReady(extractFirstURL(line))
				}
			}
			sendLogLine(logCh, line)
		}
	}

	go streamFn(stdout)
	go streamFn(stderr)

	go func() {
		_ = cmd.Wait()
		<-streamDone
		<-streamDone
		d.mu.Lock()
		d.running = false
		d.ready = false
		d.cmd = nil
		d.mu.Unlock()
		sendLogLine(logCh, "[dev server exited]")
		close(logCh)
	}()

	return nil
}

func (d *DevServer) runOfficialSetup(projectDir string, logCh chan string) error {
	sendLogLine(logCh, "Running openclaw setup...")
	if err := runCommandAndStream(logCh, projectDir, resolveNodeCommand(), "openclaw.mjs", "setup"); err != nil {
		return fmt.Errorf("openclaw setup failed: %w", err)
	}
	sendLogLine(logCh, "openclaw setup complete")
	return nil
}

func resolveOfficialDashboardURL(projectDir string) (string, error) {
	cmd := exec.Command(resolveNodeCommand(), "openclaw.mjs", "dashboard", "--no-open")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolve dashboard url: %w", err)
	}

	for _, line := range strings.Split(string(output), "\n") {
		if idx := strings.Index(line, "Dashboard URL:"); idx >= 0 {
			return strings.TrimSpace(line[idx+len("Dashboard URL:"):]), nil
		}
	}

	return "", fmt.Errorf("dashboard url not found")
}

func runCommandAndStream(logCh chan string, dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("pipe stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", strings.Join(append([]string{name}, args...), " "), err)
	}

	done := make(chan struct{}, 2)
	stream := func(r io.Reader) {
		defer func() { done <- struct{}{} }()
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 8192), 8192)
		for scanner.Scan() {
			sendLogLine(logCh, scanner.Text())
		}
	}

	go stream(stdout)
	go stream(stderr)

	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func extractFirstURL(line string) string {
	for _, field := range strings.Fields(line) {
		if strings.HasPrefix(field, "http://") || strings.HasPrefix(field, "https://") {
			return strings.TrimSpace(field)
		}
	}
	return ""
}

func sendLogLine(logCh chan string, line string) {
	if line == "" {
		return
	}
	select {
	case logCh <- line:
	default:
	}
}

func isTCPPortListening(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// Stop kills the dev server process and its child processes.
func (d *DevServer) Stop() error {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return nil
	}

	if d.external || d.cmd == nil || d.cmd.Process == nil {
		logCh := d.logCh
		d.cmd = nil
		d.external = false
		d.running = false
		d.ready = false
		d.readyURL = ""
		d.mu.Unlock()
		sendLogLine(logCh, "Disconnected from the existing OpenClaw gateway.")
		close(logCh)
		return nil
	}

	err := killProcessTree(d.cmd)
	d.cmd = nil
	d.external = false
	d.running = false
	d.ready = false
	d.readyURL = ""
	d.mu.Unlock()
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

// IsReady reports whether the local runtime is accepting connections.
func (d *DevServer) IsReady() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.ready
}

// ReadyURL returns the local URL reported by the runtime, or "" if not yet ready.
func (d *DevServer) ReadyURL() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.readyURL
}
