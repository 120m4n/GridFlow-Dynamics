package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/120m4n/GridFlow-Dynamics/internal/api/middleware"
	"github.com/120m4n/GridFlow-Dynamics/internal/domain"
)

func TestUUIDRegex(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"123e4567-e89b-12d3-a456-426614174000", true},
		{"invalid-uuid", false},
		{"550e8400e29b41d4a716446655440000", false},
		{"", false},
		{"550e8400-e29b-41d4-a716-44665544000G", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := UUIDRegex.MatchString(tt.input)
			if result != tt.expected {
				t.Errorf("UUIDRegex.MatchString(%s) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTrackingHandlerMethodNotAllowed(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	// Only POST is allowed
	app.Post("/api/v1/tracking", handler.Handle)

	methods := []string{"GET", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/tracking", nil)
			resp, _ := app.Test(req)

			// Fiber returns 405 for methods not allowed when route is not registered
			if resp.StatusCode != fiber.StatusMethodNotAllowed {
				t.Errorf("Status code = %d; want %d", resp.StatusCode, fiber.StatusMethodNotAllowed)
			}
		})
	}
}

func TestTrackingHandlerMissingSignature(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	app.Post("/api/v1/tracking", handler.Handle)

	body := `{"crewId":"550e8400-e29b-41d4-a716-446655440000"}`
	req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Status code = %d; want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestTrackingHandlerInvalidJSON(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	app.Post("/api/v1/tracking", handler.Handle)

	body := `{invalid json}`
	sig := hmacValidator.ComputeSignature([]byte(body))

	req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.SignatureHeader, sig)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Status code = %d; want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestTrackingHandlerValidation(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	app.Post("/api/v1/tracking", handler.Handle)

	tests := []struct {
		name        string
		payload     domain.TrackingPayload
		expectError bool
	}{
		{
			name: "invalid UUID",
			payload: domain.TrackingPayload{
				CrewID:             "not-a-uuid",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
				Status:             domain.TrackingStatusWorking,
				ProgressPercentage: 50,
				BatteryLevel:       80,
			},
			expectError: true,
		},
		{
			name: "invalid latitude",
			payload: domain.TrackingPayload{
				CrewID:             "550e8400-e29b-41d4-a716-446655440000",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 100.0, Longitude: -74.0},
				Status:             domain.TrackingStatusWorking,
				ProgressPercentage: 50,
				BatteryLevel:       80,
			},
			expectError: true,
		},
		{
			name: "invalid longitude",
			payload: domain.TrackingPayload{
				CrewID:             "550e8400-e29b-41d4-a716-446655440000",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -200.0},
				Status:             domain.TrackingStatusWorking,
				ProgressPercentage: 50,
				BatteryLevel:       80,
			},
			expectError: true,
		},
		{
			name: "invalid status",
			payload: domain.TrackingPayload{
				CrewID:             "550e8400-e29b-41d4-a716-446655440000",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
				Status:             "invalid_status",
				ProgressPercentage: 50,
				BatteryLevel:       80,
			},
			expectError: true,
		},
		{
			name: "invalid progress",
			payload: domain.TrackingPayload{
				CrewID:             "550e8400-e29b-41d4-a716-446655440000",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
				Status:             domain.TrackingStatusWorking,
				ProgressPercentage: 150,
				BatteryLevel:       80,
			},
			expectError: true,
		},
		{
			name: "invalid battery level",
			payload: domain.TrackingPayload{
				CrewID:             "550e8400-e29b-41d4-a716-446655440000",
				Timestamp:          time.Now(),
				GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
				Status:             domain.TrackingStatusWorking,
				ProgressPercentage: 50,
				BatteryLevel:       -10,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			sig := hmacValidator.ComputeSignature(body)

			req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(middleware.SignatureHeader, sig)
			resp, _ := app.Test(req)

			if tt.expectError && resp.StatusCode != fiber.StatusBadRequest {
				t.Errorf("Status code = %d; want %d for %s", resp.StatusCode, fiber.StatusBadRequest, tt.name)
			}
		})
	}
}

func TestTrackingHandlerRateLimiting(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(2, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	app.Post("/api/v1/tracking", handler.Handle)

	payload := domain.TrackingPayload{
		CrewID:             "550e8400-e29b-41d4-a716-446655440000",
		Timestamp:          time.Now(),
		GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
		Status:             domain.TrackingStatusWorking,
		ProgressPercentage: 50,
		BatteryLevel:       80,
	}
	body, _ := json.Marshal(payload)
	sig := hmacValidator.ComputeSignature(body)

	// First 2 requests should pass validation (may succeed since no publisher)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(middleware.SignatureHeader, sig)
		resp, _ := app.Test(req)
		// Should not be rate limited (will be 200 since no publisher)
		if resp.StatusCode == fiber.StatusTooManyRequests {
			t.Errorf("Request %d should not be rate limited", i+1)
		}
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.SignatureHeader, sig)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Errorf("Status code = %d; want %d", resp.StatusCode, fiber.StatusTooManyRequests)
	}
}

func TestTrackingHandlerSuccess(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	app.Post("/api/v1/tracking", handler.Handle)

	payload := domain.TrackingPayload{
		CrewID:             "550e8400-e29b-41d4-a716-446655440000",
		Timestamp:          time.Now(),
		GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
		Status:             domain.TrackingStatusWorking,
		ProgressPercentage: 50,
		BatteryLevel:       80,
	}
	body, _ := json.Marshal(payload)
	sig := hmacValidator.ComputeSignature(body)

	req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.SignatureHeader, sig)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Status code = %d; want %d", resp.StatusCode, fiber.StatusOK)
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	var response TrackingResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Response should be successful")
	}
	if response.RequestID != payload.CrewID {
		t.Errorf("RequestID = %s; want %s", response.RequestID, payload.CrewID)
	}
}

func TestTrackingResponse(t *testing.T) {
	resp := TrackingResponse{
		Success:   true,
		Message:   "Success",
		RequestID: "test-123",
	}

	if !resp.Success {
		t.Error("Success should be true")
	}

	if resp.Message != "Success" {
		t.Errorf("Message = %s; want Success", resp.Message)
	}

	if resp.RequestID != "test-123" {
		t.Errorf("RequestID = %s; want test-123", resp.RequestID)
	}
}

func TestTrackingHandlerWithIdempotency(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	
	// Create idempotency store and middleware
	idempotencyStore := middleware.NewInMemoryIdempotencyStore(time.Minute, 30*time.Second)
	defer idempotencyStore.Close()
	idempotencyMiddleware := middleware.NewIdempotencyMiddleware(idempotencyStore, "idempotency-secret")
	
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator).WithIdempotency(idempotencyMiddleware)

	app.Post("/api/v1/tracking", handler.Handle)

	payload := domain.TrackingPayload{
		CrewID:             "550e8400-e29b-41d4-a716-446655440000",
		Timestamp:          time.Now(),
		GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
		Status:             domain.TrackingStatusWorking,
		ProgressPercentage: 50,
		BatteryLevel:       80,
	}
	body, _ := json.Marshal(payload)
	sig := hmacValidator.ComputeSignature(body)

	// First request should succeed
	req1 := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body)))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set(middleware.SignatureHeader, sig)
	resp1, _ := app.Test(req1)

	if resp1.StatusCode != fiber.StatusOK {
		t.Errorf("First request status code = %d; want %d", resp1.StatusCode, fiber.StatusOK)
	}

	// Check idempotency key header is set
	idempotencyKey := resp1.Header.Get(IdempotencyKeyHeader)
	if idempotencyKey == "" {
		t.Error("Idempotency key header should be set")
	}

	// Read first response
	respBody1, _ := io.ReadAll(resp1.Body)
	var response1 TrackingResponse
	json.Unmarshal(respBody1, &response1)

	// Second request with same body should return cached response
	req2 := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set(middleware.SignatureHeader, sig)
	resp2, _ := app.Test(req2)

	if resp2.StatusCode != fiber.StatusOK {
		t.Errorf("Second request status code = %d; want %d", resp2.StatusCode, fiber.StatusOK)
	}

	// Read second response
	respBody2, _ := io.ReadAll(resp2.Body)
	var response2 TrackingResponse
	json.Unmarshal(respBody2, &response2)

	// Both responses should be the same
	if response1.RequestID != response2.RequestID {
		t.Errorf("Response RequestID mismatch: %s vs %s", response1.RequestID, response2.RequestID)
	}
	if response1.Message != response2.Message {
		t.Errorf("Response Message mismatch: %s vs %s", response1.Message, response2.Message)
	}
}

func TestTrackingHandlerIdempotencyDifferentPayloads(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	
	// Create idempotency store and middleware
	idempotencyStore := middleware.NewInMemoryIdempotencyStore(time.Minute, 30*time.Second)
	defer idempotencyStore.Close()
	idempotencyMiddleware := middleware.NewIdempotencyMiddleware(idempotencyStore, "idempotency-secret")
	
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator).WithIdempotency(idempotencyMiddleware)

	app.Post("/api/v1/tracking", handler.Handle)

	// First payload
	payload1 := domain.TrackingPayload{
		CrewID:             "550e8400-e29b-41d4-a716-446655440000",
		Timestamp:          time.Now(),
		GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
		Status:             domain.TrackingStatusWorking,
		ProgressPercentage: 50,
		BatteryLevel:       80,
	}
	body1, _ := json.Marshal(payload1)
	sig1 := hmacValidator.ComputeSignature(body1)

	// Second payload (different progress)
	payload2 := domain.TrackingPayload{
		CrewID:             "550e8400-e29b-41d4-a716-446655440000",
		Timestamp:          time.Now(),
		GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
		Status:             domain.TrackingStatusWorking,
		ProgressPercentage: 75, // Different progress
		BatteryLevel:       80,
	}
	body2, _ := json.Marshal(payload2)
	sig2 := hmacValidator.ComputeSignature(body2)

	// First request
	req1 := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body1)))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set(middleware.SignatureHeader, sig1)
	resp1, _ := app.Test(req1)
	idempotencyKey1 := resp1.Header.Get(IdempotencyKeyHeader)

	// Second request with different payload should get different idempotency key
	req2 := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body2)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set(middleware.SignatureHeader, sig2)
	resp2, _ := app.Test(req2)
	idempotencyKey2 := resp2.Header.Get(IdempotencyKeyHeader)

	if idempotencyKey1 == idempotencyKey2 {
		t.Error("Different payloads should have different idempotency keys")
	}

	// Both requests should succeed (not be blocked as duplicates)
	if resp1.StatusCode != fiber.StatusOK {
		t.Errorf("First request status code = %d; want %d", resp1.StatusCode, fiber.StatusOK)
	}
	if resp2.StatusCode != fiber.StatusOK {
		t.Errorf("Second request status code = %d; want %d", resp2.StatusCode, fiber.StatusOK)
	}
}

func TestTrackingHandlerIdempotencyDisabled(t *testing.T) {
	app := fiber.New()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	hmacValidator := middleware.NewHMACValidator("test-secret")
	
	// Handler without idempotency
	handler := NewTrackingHandler(nil, rateLimiter, hmacValidator)

	app.Post("/api/v1/tracking", handler.Handle)

	payload := domain.TrackingPayload{
		CrewID:             "550e8400-e29b-41d4-a716-446655440000",
		Timestamp:          time.Now(),
		GPSCoordinates:     domain.GPSCoordinates{Latitude: 40.0, Longitude: -74.0},
		Status:             domain.TrackingStatusWorking,
		ProgressPercentage: 50,
		BatteryLevel:       80,
	}
	body, _ := json.Marshal(payload)
	sig := hmacValidator.ComputeSignature(body)

	req := httptest.NewRequest("POST", "/api/v1/tracking", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.SignatureHeader, sig)
	resp, _ := app.Test(req)

	// Should work without idempotency
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Request status code = %d; want %d", resp.StatusCode, fiber.StatusOK)
	}

	// Should not have idempotency key header when disabled
	idempotencyKey := resp.Header.Get(IdempotencyKeyHeader)
	if idempotencyKey != "" {
		t.Error("Idempotency key header should not be set when idempotency is disabled")
	}
}
