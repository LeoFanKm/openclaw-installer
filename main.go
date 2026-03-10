package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openclaw/openclaw-installer/internal/platform"
	"github.com/openclaw/openclaw-installer/internal/server"
)

// Version is set at build time via -ldflags "-X main.Version=..."
var Version = "dev"

//go:embed frontend/*
var frontendFS embed.FS

func main() {
	port := findAvailablePort()
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := fmt.Sprintf("http://%s", addr)

	srv := server.New(frontendFS)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv.Handler(),
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Printf("OpenClaw Installer running at %s\n", url)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Give the server a moment to start before opening the browser.
	time.Sleep(200 * time.Millisecond)
	if err := platform.OpenBrowser(url); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not open browser: %v\n", err)
	}

	<-ctx.Done()
	fmt.Println("\nShutting down...")

	// Stop dev server if running.
	srv.Cleanup()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	fmt.Println("Goodbye.")
}

// findAvailablePort tries the default port 17834, then falls back to a random port.
func findAvailablePort() int {
	const defaultPort = 17834
	if ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", defaultPort)); err == nil {
		ln.Close()
		return defaultPort
	}
	// Fallback: let the OS pick a free port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("cannot find available port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}
