package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

const (
	// SignatureHeader is the HTTP header containing the HMAC signature.
	SignatureHeader = "X-Signature-256"
)

// HMACValidator validates HMAC-SHA256 signatures on requests.
type HMACValidator struct {
	secretKey []byte
}

// NewHMACValidator creates a new HMAC validator with the given secret key.
func NewHMACValidator(secretKey string) *HMACValidator {
	return &HMACValidator{
		secretKey: []byte(secretKey),
	}
}

// ValidateSignature validates the HMAC-SHA256 signature of the request body.
func (v *HMACValidator) ValidateSignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// Compute expected signature
	mac := hmac.New(sha256.New, v.secretKey)
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// ComputeSignature computes the HMAC-SHA256 signature for the given body.
func (v *HMACValidator) ComputeSignature(body []byte) string {
	mac := hmac.New(sha256.New, v.secretKey)
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateRequest validates the signature from an HTTP request.
// It reads the body and returns it for further processing.
func (v *HMACValidator) ValidateRequest(r *http.Request) ([]byte, bool) {
	signature := r.Header.Get(SignatureHeader)
	if signature == "" {
		return nil, false
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, false
	}

	return body, v.ValidateSignature(body, signature)
}
