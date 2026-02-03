// Package main is the entry point for the Service Worker PostgreSQL.
// This service subscribes to NATS messages and persists them to PostgreSQL.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/120m4n/GridFlow-Dynamics/service-worker-ps/internal/config"
	"github.com/120m4n/GridFlow-Dynamics/service-worker-ps/internal/repository"
	"github.com/120m4n/GridFlow-Dynamics/service-worker-ps/internal/subscriber"
	"github.com/120m4n/GridFlow-Dynamics/service-worker-ps/internal/worker"
)

func main() {
	log.Println("Starting Service Worker PostgreSQL...")

	// Load configuration
	cfg := config.Load()
	log.Printf("Configuration loaded:")
	log.Printf("  - NATS URL: %s", cfg.NATS.URL)
	log.Printf("  - NATS Subject: %s", cfg.NATS.Subject)
	log.Printf("  - Repository Type: %s", cfg.Repository)
	log.Printf("  - Number of Workers: %d", cfg.Worker.NumWorkers)
	log.Printf("  - Worker Buffer Size: %d", cfg.Worker.BufferSize)

	// Create repository using factory pattern
	factory := repository.NewFactory()
	repo, err := factory.CreateRepository(cfg.Repository, cfg.Database.ConnectionString)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing repository: %v", err)
		}
	}()

	// Test repository connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := repo.HealthCheck(ctx); err != nil {
		cancel()
		log.Fatalf("Repository health check failed: %v", err)
	}
	cancel()
	log.Println("Repository connection verified successfully")

	// Create worker pool
	shutdownTimeout := time.Duration(cfg.Worker.ShutdownTimeout) * time.Second
	pool := worker.NewPool(cfg.Worker.NumWorkers, cfg.Worker.BufferSize, repo, shutdownTimeout)
	pool.Start()
	defer func() {
		if err := pool.Shutdown(); err != nil {
			log.Printf("Error shutting down worker pool: %v", err)
		}
	}()

	// Create message handler that submits to worker pool
	messageHandler := func(ctx context.Context, msg *subscriber.Message) error {
		pool.Submit(msg)
		return nil
	}

	// Create and start NATS subscriber
	sub, err := subscriber.NewSubscriber(cfg.NATS.URL, messageHandler)
	if err != nil {
		log.Fatalf("Failed to create NATS subscriber: %v", err)
	}
	defer func() {
		if err := sub.Close(); err != nil {
			log.Printf("Error closing NATS subscriber: %v", err)
		}
	}()

	// Subscribe to NATS subject
	if err := sub.Subscribe(cfg.NATS.Subject); err != nil {
		log.Fatalf("Failed to subscribe to NATS subject: %v", err)
	}

	log.Println("Service Worker PostgreSQL is running")
	log.Printf("Listening for messages on subject: %s", cfg.NATS.Subject)
	log.Println("Press Ctrl+C to shutdown gracefully")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutdown signal received, initiating graceful shutdown...")
	log.Println("Service Worker PostgreSQL stopped")
}
