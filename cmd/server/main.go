// Package main is the entry point for the GridFlow-Dynamics platform.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/120m4n/GridFlow-Dynamics/internal/api/handlers"
	"github.com/120m4n/GridFlow-Dynamics/internal/api/middleware"
	"github.com/120m4n/GridFlow-Dynamics/internal/config"
	"github.com/120m4n/GridFlow-Dynamics/internal/messaging"
)

func main() {
	log.Println("Iniciando GridFlow-Dynamics Platform...")

	// Cargar configuración
	cfg := config.Load()

	// Crear conexión NATS
	conn := messaging.NewConnection(cfg.NATS.URL)
	if err := conn.Connect(); err != nil {
		log.Printf("Advertencia: No se pudo conectar a NATS: %v", err)
		log.Println("La plataforma funcionará en modo standalone sin mensajería")
	} else {
		log.Println("Conectado a NATS exitosamente")
		defer conn.Close()
	}

	// Crear publisher para handlers de API
	var publisher *messaging.Publisher
	if conn.IsConnected() {
		var err error
		publisher, err = messaging.NewPublisher(conn)
		if err != nil {
			log.Fatalf("Fallo al crear publisher: %v", err)
		}
		defer publisher.Close()
	}

	// Configurar aplicación Fiber
	app := fiber.New(fiber.Config{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	// Crear middleware
	rateLimiter := middleware.NewRateLimiter(cfg.API.RateLimitPerMin, time.Minute)
	hmacValidator := middleware.NewHMACValidator(cfg.API.HMACSecret)

	// Crear handler de inventario
	inventarioHandler := handlers.NewInventarioHandler(publisher, rateLimiter, hmacValidator)
	app.Post("/api/v1/mensaje_inventario/cuadrilla", inventarioHandler.Handle)

	// Endpoint de salud
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})

	// Iniciar servidor HTTP en una goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		log.Printf("Iniciando servidor HTTP en puerto %s", cfg.Server.Port)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Servidor HTTP falló: %v", err)
		}
	}()

	log.Println("GridFlow-Dynamics Platform está corriendo")
	log.Printf("Configurado para soportar 200 cuadrillas simultáneas")
	log.Printf("Endpoint de inventario: POST /api/v1/mensaje_inventario/cuadrilla")
	log.Printf("Rate limit: %d requests/minuto por cuadrilla", cfg.API.RateLimitPerMin)

	// Esperar señal de apagado
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Apagando GridFlow-Dynamics Platform...")

	// Apagado graceful del servidor HTTP
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("Error al apagar servidor HTTP: %v", err)
	}
}
