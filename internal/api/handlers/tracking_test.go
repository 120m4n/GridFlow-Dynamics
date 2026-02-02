package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/120m4n/GridFlow-Dynamics/internal/api/middleware"
	"github.com/120m4n/GridFlow-Dynamics/internal/domain"
)

func TestInventarioHandlerValidarHMAC(t *testing.T) {
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")

	handler := NewInventarioHandler(nil, rateLimiter, hmacValidator)

	app := fiber.New()
	app.Post("/test", handler.Handle)

	body := []byte(`{"grupoTrabajo":"G0/TEST"}`)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Error en test: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("StatusCode = %d; esperado %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestInventarioHandlerPayloadInvalido(t *testing.T) {
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")

	handler := NewInventarioHandler(nil, rateLimiter, hmacValidator)

	app := fiber.New()
	app.Post("/test", handler.Handle)

	body := []byte(`invalid json`)
	signature := hmacValidator.Sign(body)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.SignatureHeader, signature)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Error en test: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("StatusCode = %d; esperado %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestInventarioHandlerValidaciones(t *testing.T) {
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")

	handler := NewInventarioHandler(nil, rateLimiter, hmacValidator)

	app := fiber.New()
	app.Post("/test", handler.Handle)

	tests := []struct {
		nombre      string
		mensaje     domain.MensajeInventarioCuadrilla
		esperaError bool
	}{
		{
			nombre: "GrupoTrabajo vacío",
			mensaje: domain.MensajeInventarioCuadrilla{
				GrupoTrabajo:       "",
				NombreEmpleado:     "Juan Perez",
				Timestamp:          time.Now(),
				Coordenadas:        domain.Coordenadas{Latitud: 40.0, Longitud: -74.0},
				CodigoODT:          "ODT-001",
				Estado:             "trabajando",
				PorcentajeProgreso: 75,
				NivelBateria:       85,
			},
			esperaError: true,
		},
		{
			nombre: "Latitud inválida",
			mensaje: domain.MensajeInventarioCuadrilla{
				GrupoTrabajo:       "G0/TEST",
				NombreEmpleado:     "Juan Perez",
				Timestamp:          time.Now(),
				Coordenadas:        domain.Coordenadas{Latitud: 100.0, Longitud: -74.0},
				CodigoODT:          "ODT-001",
				Estado:             "trabajando",
				PorcentajeProgreso: 75,
				NivelBateria:       85,
			},
			esperaError: true,
		},
		{
			nombre: "Estado inválido",
			mensaje: domain.MensajeInventarioCuadrilla{
				GrupoTrabajo:       "G0/TEST",
				NombreEmpleado:     "Juan Perez",
				Timestamp:          time.Now(),
				Coordenadas:        domain.Coordenadas{Latitud: 40.0, Longitud: -74.0},
				CodigoODT:          "ODT-001",
				Estado:             "invalido",
				PorcentajeProgreso: 75,
				NivelBateria:       85,
			},
			esperaError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.nombre, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.mensaje)
			signature := hmacValidator.Sign(bodyBytes)

			req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(middleware.SignatureHeader, signature)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Error en test: %v", err)
			}

			if tt.esperaError && resp.StatusCode != fiber.StatusBadRequest {
				t.Errorf("Se esperaba error 400, obtuvo %d", resp.StatusCode)
			}
		})
	}
}

func TestInventarioHandlerRateLimit(t *testing.T) {
	rateLimiter := middleware.NewRateLimiter(2, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")

	handler := NewInventarioHandler(nil, rateLimiter, hmacValidator)

	app := fiber.New()
	app.Post("/test", handler.Handle)

	mensaje := domain.MensajeInventarioCuadrilla{
		GrupoTrabajo:       "G0/TEST",
		NombreEmpleado:     "Juan Perez",
		Timestamp:          time.Now(),
		Coordenadas:        domain.Coordenadas{Latitud: 40.0, Longitud: -74.0},
		CodigoODT:          "ODT-001",
		Estado:             "trabajando",
		PorcentajeProgreso: 75,
		NivelBateria:       85,
	}

	bodyBytes, _ := json.Marshal(mensaje)
	signature := hmacValidator.Sign(bodyBytes)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(middleware.SignatureHeader, signature)

		resp, _ := app.Test(req, -1)

		if i < 2 {
			if resp.StatusCode == fiber.StatusTooManyRequests {
				t.Errorf("Request %d: no debería estar limitado aún", i+1)
			}
		} else {
			if resp.StatusCode != fiber.StatusTooManyRequests {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Request %d: debería estar limitado, obtuvo status %d, body: %s", i+1, resp.StatusCode, string(body))
			}
		}
	}
}
