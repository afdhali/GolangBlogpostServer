package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/afdhali/GolangBlogpostServer/internal/di"
)

func main() {
	// test update init
	// Initialize application with dependency injection
	app, err := di.InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Cleanup()

	// Get configuration from router (already injected)
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// Setup HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        app.Router.Setup(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", port)
		log.Printf("🌍 Environment: %s", os.Getenv("APP_ENV"))
		log.Printf("🔗 API: http://localhost:%s/api/v1", port)
		log.Printf("❤️  Health: http://localhost:%s/health", port)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️  Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}