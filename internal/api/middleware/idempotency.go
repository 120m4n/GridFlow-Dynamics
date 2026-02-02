// Package middleware provides HTTP middleware for the API.
package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// IdempotencyResult represents the outcome of an idempotent operation.
type IdempotencyResult struct {
	// StatusCode is the HTTP status code from the original response.
	StatusCode int `json:"status_code"`
	// Body is the response body from the original response.
	Body []byte `json:"body"`
	// ContentType is the Content-Type header from the original response.
	ContentType string `json:"content_type"`
	// CreatedAt is the timestamp when the result was stored.
	CreatedAt time.Time `json:"created_at"`
}

// IdempotencyStore defines the interface for storing idempotency records.
// This interface is designed to be easily implemented with different backends
// such as Redis, PostgreSQL, or in-memory storage.
//
// Future Redis Integration:
// To implement this interface with Redis, create a new struct that uses
// a Redis client and implements all methods. The key would be stored as-is,
// and the IdempotencyResult would be serialized to JSON for storage.
// Use Redis SETEX or SET with EX option for TTL support.
//
// Example Redis implementation pseudocode:
//
//	type RedisIdempotencyStore struct {
//	    client *redis.Client
//	    ttl    time.Duration
//	}
//
//	func (s *RedisIdempotencyStore) Set(key string, result *IdempotencyResult) error {
//	    data, _ := json.Marshal(result)
//	    return s.client.SetEx(ctx, "idempotency:"+key, data, s.ttl).Err()
//	}
//
//	func (s *RedisIdempotencyStore) Get(key string) (*IdempotencyResult, bool, error) {
//	    data, err := s.client.Get(ctx, "idempotency:"+key).Bytes()
//	    if err == redis.Nil { return nil, false, nil }
//	    if err != nil { return nil, false, err }
//	    var result IdempotencyResult
//	    json.Unmarshal(data, &result)
//	    return &result, true, nil
//	}
type IdempotencyStore interface {
	// Set stores an idempotency result for the given key.
	// The key should be a unique identifier for the request (e.g., HMAC hash).
	// Returns an error if the storage operation fails.
	Set(key string, result *IdempotencyResult) error

	// Get retrieves an idempotency result for the given key.
	// Returns the result, a boolean indicating if the key exists, and any error.
	Get(key string) (*IdempotencyResult, bool, error)

	// Exists checks if a key exists in the store without retrieving the full result.
	// This can be more efficient than Get for simple duplicate detection.
	Exists(key string) (bool, error)

	// Delete removes an idempotency record for the given key.
	// This can be used for cache invalidation or manual cleanup.
	Delete(key string) error

	// Close gracefully shuts down the store and releases resources.
	Close() error
}

// idempotencyEntry represents an entry in the in-memory store with expiration.
type idempotencyEntry struct {
	result    *IdempotencyResult
	expiresAt time.Time
}

// InMemoryIdempotencyStore implements IdempotencyStore using an in-memory map.
// It provides TTL-based expiration and efficient cleanup using a background goroutine.
//
// This implementation is suitable for single-instance deployments or development.
// For distributed systems, consider using the Redis implementation.
type InMemoryIdempotencyStore struct {
	entries     map[string]*idempotencyEntry
	ttl         time.Duration
	mu          sync.RWMutex
	stopCleanup chan struct{}
}

// NewInMemoryIdempotencyStore creates a new in-memory idempotency store.
// ttl specifies how long entries should be kept before expiring.
// cleanupInterval determines how often the cleanup goroutine runs.
func NewInMemoryIdempotencyStore(ttl time.Duration, cleanupInterval time.Duration) *InMemoryIdempotencyStore {
	store := &InMemoryIdempotencyStore{
		entries:     make(map[string]*idempotencyEntry),
		ttl:         ttl,
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup
	go store.cleanupLoop(cleanupInterval)

	return store
}

// Set stores an idempotency result with the configured TTL.
func (s *InMemoryIdempotencyStore) Set(key string, result *IdempotencyResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[key] = &idempotencyEntry{
		result:    result,
		expiresAt: time.Now().Add(s.ttl),
	}
	return nil
}

// Get retrieves an idempotency result if it exists and hasn't expired.
func (s *InMemoryIdempotencyStore) Get(key string) (*IdempotencyResult, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.entries[key]
	if !exists {
		return nil, false, nil
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		return nil, false, nil
	}

	return entry.result, true, nil
}

// Exists checks if a non-expired entry exists for the given key.
func (s *InMemoryIdempotencyStore) Exists(key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.entries[key]
	if !exists {
		return false, nil
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		return false, nil
	}

	return true, nil
}

// Delete removes an entry from the store.
func (s *InMemoryIdempotencyStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, key)
	return nil
}

// Close stops the cleanup goroutine and releases resources.
func (s *InMemoryIdempotencyStore) Close() error {
	close(s.stopCleanup)
	return nil
}

// cleanupLoop periodically removes expired entries from the store.
func (s *InMemoryIdempotencyStore) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCleanup:
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup removes all expired entries from the store.
func (s *InMemoryIdempotencyStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, entry := range s.entries {
		if now.After(entry.expiresAt) {
			delete(s.entries, key)
		}
	}
}

// Size returns the current number of entries in the store.
// Useful for monitoring and testing.
func (s *InMemoryIdempotencyStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

// IdempotencyKeyGenerator generates idempotency keys using HMAC-SHA256.
// The key is computed from the request body and a secret key, ensuring
// that identical requests produce the same key while different requests
// produce different keys.
type IdempotencyKeyGenerator struct {
	secretKey []byte
}

// NewIdempotencyKeyGenerator creates a new key generator with the given secret.
func NewIdempotencyKeyGenerator(secretKey string) *IdempotencyKeyGenerator {
	return &IdempotencyKeyGenerator{
		secretKey: []byte(secretKey),
	}
}

// GenerateKey computes an HMAC-SHA256 hash of the request body.
// This hash serves as a unique identifier for the request content.
// Optionally, additional context (like crew ID or timestamp) can be included
// to make the key more specific to the request context.
func (g *IdempotencyKeyGenerator) GenerateKey(body []byte) string {
	mac := hmac.New(sha256.New, g.secretKey)
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// GenerateKeyWithContext computes an HMAC-SHA256 hash including additional context.
// This is useful when you want to scope idempotency to a specific resource or user.
// For example, passing the crew ID as context ensures that identical payloads
// from different crews are treated as different requests.
func (g *IdempotencyKeyGenerator) GenerateKeyWithContext(body []byte, context string) string {
	mac := hmac.New(sha256.New, g.secretKey)
	mac.Write([]byte(context))
	mac.Write([]byte(":"))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// IdempotencyMiddleware provides request idempotency for API endpoints.
// It uses HMAC-based key generation and a configurable store backend.
type IdempotencyMiddleware struct {
	store        IdempotencyStore
	keyGenerator *IdempotencyKeyGenerator
}

// NewIdempotencyMiddleware creates a new idempotency middleware.
func NewIdempotencyMiddleware(store IdempotencyStore, secretKey string) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{
		store:        store,
		keyGenerator: NewIdempotencyKeyGenerator(secretKey),
	}
}

// GenerateKey generates an idempotency key for the given request body.
func (m *IdempotencyMiddleware) GenerateKey(body []byte) string {
	return m.keyGenerator.GenerateKey(body)
}

// GenerateKeyWithContext generates an idempotency key with additional context.
func (m *IdempotencyMiddleware) GenerateKeyWithContext(body []byte, context string) string {
	return m.keyGenerator.GenerateKeyWithContext(body, context)
}

// CheckDuplicate checks if a request with the given key has been processed.
// Returns the cached result if found, nil otherwise.
func (m *IdempotencyMiddleware) CheckDuplicate(key string) (*IdempotencyResult, bool, error) {
	return m.store.Get(key)
}

// StoreResult stores the result of a processed request for future duplicate detection.
func (m *IdempotencyMiddleware) StoreResult(key string, statusCode int, body []byte, contentType string) error {
	result := &IdempotencyResult{
		StatusCode:  statusCode,
		Body:        body,
		ContentType: contentType,
		CreatedAt:   time.Now(),
	}
	return m.store.Set(key, result)
}
