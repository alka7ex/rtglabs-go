package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"rtglabs-go/internal/server" // Assuming this returns *echo.Echo or *http.Server
)

// gracefulShutdown remains the same
func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done() // Block until interrupt signal is received

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Re-arm context for second Ctrl+C to force exit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")
	done <- true // Signal main goroutine that shutdown is complete
}

func main() {
	// Initialize your Ent client (database connection) here BEFORE starting the server.
	// This is important because if the DB connection fails, you want to exit early.
	// For example:

	srv := server.NewServer() // Assuming this returns *http.Server or something equivalent that has ListenAndServe

	// Create a channel for graceful shutdown notification
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine.
	// This goroutine will wait for OS signals and then call srv.Shutdown().
	go gracefulShutdown(srv, done)

	// Start the server in a separate goroutine.
	// This allows the main goroutine to wait for the `done` channel,
	// keeping the application alive until gracefulShutdown signals completion.
	go func() {
		log.Printf("Starting HTTP server on %s", srv.Addr) // Log the address server is listening on
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed to start or crashed: %v", err) // Use log.Fatalf to ensure exit on fatal error
		}
		// If ListenAndServe returns http.ErrServerClosed, it means graceful shutdown was initiated.
		log.Println("HTTP server stopped listening.")
	}()

	// The main goroutine now blocks, waiting for the graceful shutdown to complete.
	// It will only unblock when the `done` channel receives a value,
	// which happens after gracefulShutdown finishes.
	<-done
	log.Println("Graceful shutdown complete. Application exiting.")
}
