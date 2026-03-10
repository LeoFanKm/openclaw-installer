package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"sync"

	"github.com/openclaw/openclaw-installer/internal/config"
	"github.com/openclaw/openclaw-installer/internal/platform"
	"github.com/openclaw/openclaw-installer/internal/prereqs"
	"github.com/openclaw/openclaw-installer/internal/project"
)

// AppState holds the shared mutable state for the installer.
type AppState struct {
	mu sync.Mutex

	// Target directory for the OpenClaw project.
	ProjectDir string `json:"projectDir"`

	// Whether the repo has been cloned.
	Cloned bool `json:"cloned"`

	// Whether .dev.vars has been saved.
	Configured bool `json:"configured"`

	// Channels for SSE streaming.
	prereqProgressCh chan string
	projectProgressCh chan string

	// DevServer instance.
	devServer *project.DevServer
}

// Server wraps the HTTP mux and application state.
type Server struct {
	mux       *http.ServeMux
	state     *AppState
	frontend  embed.FS
}

// New creates a new Server with the given embedded frontend filesystem.
func New(frontendFS embed.FS) *Server {
	s := &Server{
		mux:      http.NewServeMux(),
		frontend: frontendFS,
		state: &AppState{
			ProjectDir:        platform.DefaultProjectDir(),
			prereqProgressCh:  make(chan string, 128),
			projectProgressCh: make(chan string, 128),
			devServer:         project.NewDevServer(),
		},
	}
	s.registerRoutes()
	return s
}

// Handler returns the http.Handler for this server.
func (s *Server) Handler() http.Handler {
	return s.mux
}

// Cleanup stops the dev server if running.
func (s *Server) Cleanup() {
	_ = s.state.devServer.Stop()
}

func (s *Server) registerRoutes() {
	// API routes.
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("POST /api/prereqs/install", s.handlePrereqsInstall)
	s.mux.HandleFunc("GET /api/prereqs/progress", s.handlePrereqsProgress)
	s.mux.HandleFunc("POST /api/project/clone", s.handleProjectClone)
	s.mux.HandleFunc("GET /api/project/progress", s.handleProjectProgress)
	s.mux.HandleFunc("POST /api/config/save", s.handleConfigSave)
	s.mux.HandleFunc("POST /api/config/test", s.handleConfigTest)
	s.mux.HandleFunc("POST /api/dev/start", s.handleDevStart)
	s.mux.HandleFunc("POST /api/dev/stop", s.handleDevStop)
	s.mux.HandleFunc("GET /api/dev/logs", s.handleDevLogs)

	// Serve embedded frontend files.
	frontendSub, err := fs.Sub(s.frontend, "frontend")
	if err != nil {
		panic(fmt.Sprintf("cannot create sub-filesystem for frontend: %v", err))
	}
	fileServer := http.FileServer(http.FS(frontendSub))
	s.mux.Handle("/", fileServer)
}

// --- API Handlers ---

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	gitInstalled, gitVersion, _ := prereqs.DetectGit()
	nodeInstalled, nodeVersion, _ := prereqs.DetectNode()
	npmInstalled, npmVersion, _ := prereqs.DetectNpm()

	s.state.mu.Lock()
	projectDir := s.state.ProjectDir
	cloned := s.state.Cloned
	configured := s.state.Configured
	s.state.mu.Unlock()

	// Check if project directory exists with a package.json (already cloned).
	if !cloned {
		if _, err := os.Stat(projectDir + "/package.json"); err == nil {
			s.state.mu.Lock()
			s.state.Cloned = true
			cloned = true
			s.state.mu.Unlock()
		}
	}

	// Check if .dev.vars exists (already configured).
	if !configured {
		if _, err := os.Stat(projectDir + "/.dev.vars"); err == nil {
			s.state.mu.Lock()
			s.state.Configured = true
			configured = true
			s.state.mu.Unlock()
		}
	}

	resp := map[string]any{
		"git": map[string]any{
			"installed": gitInstalled,
			"version":   gitVersion,
		},
		"node": map[string]any{
			"installed": nodeInstalled,
			"version":   nodeVersion,
		},
		"npm": map[string]any{
			"installed": npmInstalled,
			"version":   npmVersion,
		},
		"projectDir":       projectDir,
		"cloned":           cloned,
		"configured":       configured,
		"devServerRunning": s.state.devServer.IsRunning(),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handlePrereqsInstall(w http.ResponseWriter, r *http.Request) {
	// Drain any leftover messages in the channel.
	drainChannel(s.state.prereqProgressCh)

	go func() {
		ch := s.state.prereqProgressCh

		gitInstalled, _, _ := prereqs.DetectGit()
		if !gitInstalled {
			ch <- "Git not found, installing..."
			if err := prereqs.InstallGit(ch); err != nil {
				ch <- fmt.Sprintf("ERROR: Git install failed: %v", err)
				ch <- "DONE"
				return
			}
		} else {
			ch <- "Git already installed, skipping"
		}

		nodeInstalled, _, _ := prereqs.DetectNode()
		if !nodeInstalled {
			ch <- "Node.js not found, installing..."
			if err := prereqs.InstallNode(ch); err != nil {
				ch <- fmt.Sprintf("ERROR: Node.js install failed: %v", err)
				ch <- "DONE"
				return
			}
		} else {
			ch <- "Node.js already installed, skipping"
		}

		// Refresh PATH so newly installed tools are found.
		if err := prereqs.RefreshPath(); err != nil {
			ch <- fmt.Sprintf("WARNING: Could not refresh PATH: %v", err)
		}

		ch <- "All prerequisites installed"
		ch <- "DONE"
	}()

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handlePrereqsProgress(w http.ResponseWriter, r *http.Request) {
	s.streamSSE(w, r, s.state.prereqProgressCh)
}

func (s *Server) handleProjectClone(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TargetDir string `json:"targetDir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}

	if body.TargetDir != "" {
		s.state.mu.Lock()
		s.state.ProjectDir = body.TargetDir
		s.state.mu.Unlock()
	}

	s.state.mu.Lock()
	targetDir := s.state.ProjectDir
	s.state.mu.Unlock()

	// Drain any leftover messages.
	drainChannel(s.state.projectProgressCh)

	go func() {
		ch := s.state.projectProgressCh

		// Clone.
		if err := project.CloneRepo("", targetDir, ch); err != nil {
			ch <- fmt.Sprintf("ERROR: Clone failed: %v", err)
			ch <- "DONE"
			return
		}

		s.state.mu.Lock()
		s.state.Cloned = true
		s.state.mu.Unlock()

		// Install npm dependencies.
		ch <- "Starting npm install..."
		if err := project.InstallDeps(targetDir, ch); err != nil {
			ch <- fmt.Sprintf("ERROR: npm install failed: %v", err)
			ch <- "DONE"
			return
		}

		ch <- "Project setup complete"
		ch <- "DONE"
	}()

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleProjectProgress(w http.ResponseWriter, r *http.Request) {
	s.streamSSE(w, r, s.state.projectProgressCh)
}

func (s *Server) handleConfigSave(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}

	s.state.mu.Lock()
	projectDir := s.state.ProjectDir
	s.state.mu.Unlock()

	if err := config.SaveConfig(projectDir, body); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	s.state.mu.Lock()
	s.state.Configured = true
	s.state.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleConfigTest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ClerkSecretKey string `json:"clerkSecretKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}

	valid, err := config.TestClerkConnection(body.ClerkSecretKey)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"valid": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"valid": valid})
}

func (s *Server) handleDevStart(w http.ResponseWriter, r *http.Request) {
	s.state.mu.Lock()
	projectDir := s.state.ProjectDir
	s.state.mu.Unlock()

	if err := s.state.devServer.Start(projectDir); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDevStop(w http.ResponseWriter, r *http.Request) {
	if err := s.state.devServer.Stop(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDevLogs(w http.ResponseWriter, r *http.Request) {
	s.streamSSE(w, r, s.state.devServer.Logs())
}

// --- SSE streaming ---

func (s *Server) streamSSE(w http.ResponseWriter, r *http.Request, ch <-chan string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "streaming unsupported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				// Channel closed.
				fmt.Fprintf(w, "data: DONE\n\n")
				flusher.Flush()
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
			if msg == "DONE" {
				return
			}
		}
	}
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func drainChannel(ch chan string) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
