package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHMACValidator(t *testing.T) {
	v := NewHMACValidator("test-secret")
	if v == nil {
		t.Fatal("NewHMACValidator returned nil")
	}
}

func TestHMACValidatorComputeSignature(t *testing.T) {
	v := NewHMACValidator("test-secret")
	body := []byte(`{"crewId":"123"}`)

	sig1 := v.ComputeSignature(body)
	sig2 := v.ComputeSignature(body)

	// Same body should produce same signature
	if sig1 != sig2 {
		t.Error("Same body should produce same signature")
	}

	// Different body should produce different signature
	sig3 := v.ComputeSignature([]byte(`{"crewId":"456"}`))
	if sig1 == sig3 {
		t.Error("Different body should produce different signature")
	}
}

func TestHMACValidatorValidateSignature(t *testing.T) {
	v := NewHMACValidator("test-secret")
	body := []byte(`{"crewId":"123"}`)

	validSig := v.ComputeSignature(body)

	tests := []struct {
		name      string
		body      []byte
		signature string
		expected  bool
	}{
		{
			name:      "valid signature",
			body:      body,
			signature: validSig,
			expected:  true,
		},
		{
			name:      "invalid signature",
			body:      body,
			signature: "invalid-signature",
			expected:  false,
		},
		{
			name:      "empty signature",
			body:      body,
			signature: "",
			expected:  false,
		},
		{
			name:      "wrong body",
			body:      []byte(`{"crewId":"different"}`),
			signature: validSig,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateSignature(tt.body, tt.signature)
			if result != tt.expected {
				t.Errorf("ValidateSignature = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestHMACValidatorValidateRequest(t *testing.T) {
	v := NewHMACValidator("test-secret")
	body := []byte(`{"crewId":"123"}`)
	validSig := v.ComputeSignature(body)

	tests := []struct {
		name       string
		body       []byte
		signature  string
		expectOK   bool
		expectBody bool
	}{
		{
			name:       "valid request",
			body:       body,
			signature:  validSig,
			expectOK:   true,
			expectBody: true,
		},
		{
			name:       "missing signature",
			body:       body,
			signature:  "",
			expectOK:   false,
			expectBody: false,
		},
		{
			name:       "invalid signature",
			body:       body,
			signature:  "wrong",
			expectOK:   false,
			expectBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tracking", bytes.NewReader(tt.body))
			if tt.signature != "" {
				req.Header.Set(SignatureHeader, tt.signature)
			}

			resultBody, ok := v.ValidateRequest(req)

			if ok != tt.expectOK {
				t.Errorf("ValidateRequest ok = %v; want %v", ok, tt.expectOK)
			}

			if tt.expectBody && resultBody == nil {
				t.Error("Expected body to be returned")
			}
		})
	}
}

func TestDifferentSecrets(t *testing.T) {
	v1 := NewHMACValidator("secret-1")
	v2 := NewHMACValidator("secret-2")

	body := []byte(`{"crewId":"123"}`)

	sig1 := v1.ComputeSignature(body)
	sig2 := v2.ComputeSignature(body)

	if sig1 == sig2 {
		t.Error("Different secrets should produce different signatures")
	}

	// Validate with correct validator
	if !v1.ValidateSignature(body, sig1) {
		t.Error("Signature should be valid with matching validator")
	}

	// Validate with wrong validator
	if v2.ValidateSignature(body, sig1) {
		t.Error("Signature should be invalid with different validator")
	}
}
