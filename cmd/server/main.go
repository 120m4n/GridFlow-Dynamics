// Package main is the entry point for the GridFlow-Dynamics platform.
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

	"github.com/120m4n/GridFlow-Dynamics/internal/api/handlers"
	"github.com/120m4n/GridFlow-Dynamics/internal/api/middleware"
	"github.com/120m4n/GridFlow-Dynamics/internal/config"
	"github.com/120m4n/GridFlow-Dynamics/internal/messaging"
	"github.com/120m4n/GridFlow-Dynamics/internal/services/alert"
	"github.com/120m4n/GridFlow-Dynamics/internal/services/crew"
	"github.com/120m4n/GridFlow-Dynamics/internal/services/task"
)

func main() {
	log.Println("Starting GridFlow-Dynamics Platform...")

	// Load configuration
	cfg := config.Load()

	// Create RabbitMQ connection
	conn := messaging.NewConnection(cfg.RabbitMQ.URL)
	if err := conn.Connect(); err != nil {
		log.Printf("Warning: Could not connect to RabbitMQ: %v", err)
		log.Println("The platform will run in standalone mode without messaging")
	} else {
		log.Println("Connected to RabbitMQ successfully")
		defer conn.Close()
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create publisher for API handlers
	var publisher *messaging.Publisher
	if conn.IsConnected() {
		var err error
		publisher, err = messaging.NewPublisher(conn)
		if err != nil {
			log.Fatalf("Failed to create publisher: %v", err)
		}
		defer publisher.Close()
	}

	// Initialize services if connected to RabbitMQ
	if conn.IsConnected() {
		if err := initializeServices(ctx, conn); err != nil {
			log.Fatalf("Failed to initialize services: %v", err)
		}
	}

	// Setup HTTP server with tracking API
	mux := http.NewServeMux()

	// Create middleware
	rateLimiter := middleware.NewRateLimiter(cfg.API.RateLimitPerMin, time.Minute)
	hmacValidator := middleware.NewHMACValidator(cfg.API.HMACSecret)

	// Create tracking handler
	trackingHandler := handlers.NewTrackingHandler(publisher, rateLimiter, hmacValidator)
	mux.Handle("/api/v1/tracking", trackingHandler)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Start HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting HTTP server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	log.Println("GridFlow-Dynamics Platform is running")
	log.Printf("Configured to support 200 simultaneous crews")
	log.Printf("Tracking API endpoint: POST /api/v1/tracking")
	log.Printf("Rate limit: %d requests/minute per crew", cfg.API.RateLimitPerMin)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down GridFlow-Dynamics Platform...")

	// Graceful shutdown of HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}

func initializeServices(ctx context.Context, conn *messaging.Connection) error {
	// Create publisher
	publisher, err := messaging.NewPublisher(conn)
	if err != nil {
		return err
	}

	// Initialize Crew Tracking Service
	crewConsumer, err := messaging.NewConsumer(conn, "crew-tracking-queue")
	if err != nil {
		return err
	}
	crewService := crew.NewService(publisher, crewConsumer)
	if err := crewService.Start(); err != nil {
		return err
	}
	log.Println("Crew Tracking Service started")

	// Initialize Task Management Service
	taskConsumer, err := messaging.NewConsumer(conn, "task-management-queue")
	if err != nil {
		return err
	}
	taskService := task.NewService(publisher, taskConsumer)
	if err := taskService.Start(); err != nil {
		return err
	}
	log.Println("Task Management Service started")

	// Initialize Alert Management Service
	alertConsumer, err := messaging.NewConsumer(conn, "alert-management-queue")
	if err != nil {
		return err
	}
	alertService := alert.NewService(publisher, alertConsumer)
	if err := alertService.Start(); err != nil {
		return err
	}
	log.Println("Alert Management Service started")

	_ = ctx // context available for future use

	return nil
}
