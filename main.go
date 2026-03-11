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
	// Use --lan flag to bind to all interfaces for LAN testing.
	lanMode := len(os.Args) > 1 && os.Args[1] == "--lan"

	port := findAvailablePort(lanMode)
	bindHost := "127.0.0.1"
	if lanMode {
		bindHost = "0.0.0.0"
	}
	addr := fmt.Sprintf("%s:%d", bindHost, port)
	localURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	srv := server.New(frontendFS)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv.Handler(),
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Printf("OpenClaw Installer running at %s\n", localURL)
		if lanMode {
			if lanIP := getLanIP(); lanIP != "" {
				fmt.Printf("LAN access: http://%s:%d\n", lanIP, port)
			}
		}
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Give the server a moment to start before opening the browser.
	time.Sleep(200 * time.Millisecond)
	if err := platform.OpenBrowser(localURL); err != nil {
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
func findAvailablePort(lan bool) int {
	host := "127.0.0.1"
	if lan {
		host = "0.0.0.0"
	}
	const defaultPort = 17834
	if ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, defaultPort)); err == nil {
		ln.Close()
		return defaultPort
	}
	// Fallback: let the OS pick a free port.
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:0", host))
	if err != nil {
		log.Fatalf("cannot find available port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

// getLanIP returns the best non-loopback IPv4 address, preferring
// real LAN addresses (192.168.x.x, 10.x.x.x, 172.16-31.x.x) over
// link-local (169.254.x.x) or virtual interfaces.
func getLanIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	var fallback string
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue
		}
		ip := ipNet.IP.To4()
		// Skip link-local (169.254.x.x) — these aren't routable on LAN.
		if ip[0] == 169 && ip[1] == 254 {
			if fallback == "" {
				fallback = ip.String()
			}
			continue
		}
		// Prefer common private ranges.
		if ip[0] == 192 || ip[0] == 10 || (ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31) {
			return ip.String()
		}
		if fallback == "" {
			fallback = ip.String()
		}
	}
	return fallback
}
