// main.go — MCP server entry point with transport selection
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyberspacesec/go-crt.sh/pkg/crtsh"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	transport := flag.String("transport", "stdio", "Transport mode: stdio, sse, http")
	addr := flag.String("addr", ":8080", "Listen address for SSE and HTTP modes (e.g. :8080)")
	baseURL := flag.String("base-url", "", "Base URL for SSE mode (e.g. https://my-server.com)")
	flag.Parse()

	// Create MCP server
	s := server.NewMCPServer(
		"go-crt.sh",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
		server.WithInstructions(
			"A certificate transparency search tool powered by crt.sh. "+
				"Use search_certificates to find certificates by domain, SHA-256 hash, serial number, etc. "+
				"Use get_certificate to retrieve a specific certificate by its crt.sh ID.",
		),
	)

	// Create shared crt.sh client
	client := crtsh.NewClient()

	// Register tools
	registerTools(s, client)

	// Start selected transport
	switch *transport {
	case "stdio":
		log.Println("Starting MCP server in stdio mode")
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("stdio server error: %v", err)
		}

	case "sse":
		opts := []server.SSEOption{
			server.WithSSEEndpoint("/sse"),
			server.WithMessageEndpoint("/message"),
		}
		if *baseURL != "" {
			opts = append(opts, server.WithBaseURL(*baseURL))
		}
		sseServer := server.NewSSEServer(s, opts...)

		// Graceful shutdown
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		go func() {
			<-ctx.Done()
			log.Println("Shutting down SSE server...")
			sseServer.Shutdown(context.Background())
		}()

		log.Printf("SSE server starting on %s", *addr)
		if err := sseServer.Start(*addr); err != nil {
			log.Fatalf("SSE server error: %v", err)
		}

	case "http":
		httpServer := server.NewStreamableHTTPServer(s,
			server.WithEndpointPath("/mcp"),
		)

		// Graceful shutdown
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		go func() {
			<-ctx.Done()
			log.Println("Shutting down HTTP server...")
			httpServer.Shutdown(context.Background())
		}()

		log.Printf("StreamableHTTP server starting on %s", *addr)
		if err := httpServer.Start(*addr); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown transport: %s (use stdio, sse, or http)\n", *transport)
		os.Exit(1)
	}
}
