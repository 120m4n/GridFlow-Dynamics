// Package main is the entry point for the GridFlow-Dynamics platform.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	// Initialize services if connected to RabbitMQ
	if conn.IsConnected() {
		if err := initializeServices(ctx, conn); err != nil {
			log.Fatalf("Failed to initialize services: %v", err)
		}
	}

	log.Println("GridFlow-Dynamics Platform is running")
	log.Printf("Configured to support 200 simultaneous crews")
	log.Printf("Server port: %s", cfg.Server.Port)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down GridFlow-Dynamics Platform...")
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
