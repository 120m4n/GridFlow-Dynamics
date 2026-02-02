// Package handlers provides HTTP handlers for the API.
package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/120m4n/GridFlow-Dynamics/internal/api/middleware"
	"github.com/120m4n/GridFlow-Dynamics/internal/domain"
	"github.com/120m4n/GridFlow-Dynamics/internal/messaging"
)

// InventarioHandler maneja las solicitudes de inventario de cuadrilla.
type InventarioHandler struct {
	publisher     *messaging.Publisher
	rateLimiter   *middleware.RateLimiter
	hmacValidator *middleware.HMACValidator
}

// NewInventarioHandler crea un nuevo handler de inventario.
func NewInventarioHandler(publisher *messaging.Publisher, rateLimiter *middleware.RateLimiter, hmacValidator *middleware.HMACValidator) *InventarioHandler {
	return &InventarioHandler{
		publisher:     publisher,
		rateLimiter:   rateLimiter,
		hmacValidator: hmacValidator,
	}
}

// RespuestaAPI representa la respuesta de la API.
type RespuestaAPI struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Handle maneja las solicitudes POST al endpoint de inventario de cuadrilla usando Fiber.
func (h *InventarioHandler) Handle(c *fiber.Ctx) error {
	// Validar firma HMAC
	body := c.Body()
	signature := c.Get(middleware.SignatureHeader)
	if !h.hmacValidator.ValidateSignature(body, signature) {
		return h.sendError(c, fiber.StatusUnauthorized, "Firma HMAC-SHA256 inválida o faltante")
	}

	// Parsear el payload
	var mensaje domain.MensajeInventarioCuadrilla
	if err := c.BodyParser(&mensaje); err != nil {
		return h.sendError(c, fiber.StatusBadRequest, fmt.Sprintf("Payload JSON inválido: %v", err))
	}

	// Validar el payload
	if err := mensaje.Validar(); err != nil {
		return h.sendError(c, fiber.StatusBadRequest, err.Error())
	}

	// Verificar límite de tasa
	if !h.rateLimiter.Allow(mensaje.GrupoTrabajo) {
		remaining := h.rateLimiter.Remaining(mensaje.GrupoTrabajo)
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		return h.sendError(c, fiber.StatusTooManyRequests, "Rate limit excedido (100 req/min)")
	}

	// Configurar headers de límite de tasa
	remaining := h.rateLimiter.Remaining(mensaje.GrupoTrabajo)
	c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	c.Set("X-RateLimit-Limit", "100")

	// Convertir a evento
	evento := h.mensajeAEvento(&mensaje)

	// Publicar a NATS (si el publisher está disponible)
	if h.publisher != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.publisher.Publish(ctx, messaging.SubjectInventarioCuadrilla, evento); err != nil {
			log.Printf("Fallo al publicar evento de inventario: %v", err)
			return h.sendError(c, fiber.StatusInternalServerError, "Fallo al procesar mensaje de inventario")
		}
	}

	log.Printf("Mensaje de inventario recibido de cuadrilla %s: empleado=%s, estado=%s, progreso=%d%%, ODT=%s",
		mensaje.GrupoTrabajo, mensaje.NombreEmpleado, mensaje.Estado, mensaje.PorcentajeProgreso, mensaje.CodigoODT)

	// Enviar respuesta exitosa
	return h.sendSuccess(c, "Mensaje de inventario de cuadrilla recibido correctamente.")
}

func (h *InventarioHandler) mensajeAEvento(m *domain.MensajeInventarioCuadrilla) *domain.EventoInventarioCuadrilla {
	return &domain.EventoInventarioCuadrilla{
		GrupoTrabajo:       m.GrupoTrabajo,
		NombreEmpleado:     m.NombreEmpleado,
		Timestamp:          m.Timestamp,
		Coordenadas:        m.Coordenadas,
		CodigoODT:          m.CodigoODT,
		Estado:             m.Estado,
		PorcentajeProgreso: m.PorcentajeProgreso,
		NivelBateria:       m.NivelBateria,
		RecibidoEn:         time.Now(),
	}
}

func (h *InventarioHandler) sendError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(RespuestaAPI{
		Status: "error",
		Error:  message,
	})
}

func (h *InventarioHandler) sendSuccess(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusOK).JSON(RespuestaAPI{
		Status:  "success",
		Message: message,
	})
}
