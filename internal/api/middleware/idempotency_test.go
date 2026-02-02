package middleware

import (
	"sync"
	"testing"
	"time"
)

func TestNewInMemoryIdempotencyStore(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute, 30*time.Second)
	if store == nil {
		t.Fatal("NewInMemoryIdempotencyStore returned nil")
	}
	defer store.Close()

	if store.ttl != time.Minute {
		t.Errorf("TTL = %v; want %v", store.ttl, time.Minute)
	}
}

func TestInMemoryIdempotencyStoreSetAndGet(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute, 30*time.Second)
	defer store.Close()

	result := &IdempotencyResult{
		StatusCode:  200,
		Body:        []byte(`{"success":true}`),
		ContentType: "application/json",
		CreatedAt:   time.Now(),
	}

	// Set the result
	if err := store.Set("test-key", result); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get the result
	got, exists, err := store.Get("test-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !exists {
		t.Fatal("Key should exist")
	}
	if got.StatusCode != result.StatusCode {
		t.Errorf("StatusCode = %d; want %d", got.StatusCode, result.StatusCode)
	}
	if string(got.Body) != string(result.Body) {
		t.Errorf("Body = %s; want %s", got.Body, result.Body)
	}
	if got.ContentType != result.ContentType {
		t.Errorf("ContentType = %s; want %s", got.ContentType, result.ContentType)
	}
}

func TestInMemoryIdempotencyStoreGetNonExistent(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute, 30*time.Second)
	defer store.Close()

	got, exists, err := store.Get("non-existent-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if exists {
		t.Error("Key should not exist")
	}
	if got != nil {
		t.Error("Result should be nil for non-existent key")
	}
}

func TestInMemoryIdempotencyStoreExists(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute, 30*time.Second)
	defer store.Close()

	// Check non-existent key
	exists, err := store.Exists("test-key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Key should not exist initially")
	}

	// Add key
	result := &IdempotencyResult{StatusCode: 200, Body: []byte(`{}`)}
	store.Set("test-key", result)

	// Check existing key
	exists, err = store.Exists("test-key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Key should exist after Set")
	}
}

func TestInMemoryIdempotencyStoreDelete(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute, 30*time.Second)
	defer store.Close()

	// Add and delete key
	result := &IdempotencyResult{StatusCode: 200, Body: []byte(`{}`)}
	store.Set("test-key", result)
	store.Delete("test-key")

	// Check it's gone
	exists, _ := store.Exists("test-key")
	if exists {
		t.Error("Key should not exist after Delete")
	}
}

func TestInMemoryIdempotencyStoreTTLExpiration(t *testing.T) {
	store := NewInMemoryIdempotencyStore(100*time.Millisecond, 50*time.Millisecond)
	defer store.Close()

	result := &IdempotencyResult{StatusCode: 200, Body: []byte(`{}`)}
	store.Set("test-key", result)

	// Should exist immediately
	exists, _ := store.Exists("test-key")
	if !exists {
		t.Error("Key should exist immediately after Set")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not exist after TTL
	exists, _ = store.Exists("test-key")
	if exists {
		t.Error("Key should not exist after TTL expiration")
	}

	// Get should also return false for expired
	_, found, _ := store.Get("test-key")
	if found {
		t.Error("Get should return false for expired key")
	}
}

func TestInMemoryIdempotencyStoreCleanup(t *testing.T) {
	store := NewInMemoryIdempotencyStore(50*time.Millisecond, 25*time.Millisecond)
	defer store.Close()

	// Add multiple entries
	for i := 0; i < 10; i++ {
		result := &IdempotencyResult{StatusCode: 200, Body: []byte(`{}`)}
		store.Set("test-key-"+string(rune('0'+i)), result)
	}

	// Verify entries exist
	if store.Size() != 10 {
		t.Errorf("Size = %d; want 10", store.Size())
	}

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Entries should be cleaned up
	if store.Size() != 0 {
		t.Errorf("Size = %d; want 0 after cleanup", store.Size())
	}
}

func TestInMemoryIdempotencyStoreConcurrent(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute, 30*time.Second)
	defer store.Close()

	var wg sync.WaitGroup
	numGoroutines := 50
	numOperations := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "key-" + string(rune('A'+id%26))
				result := &IdempotencyResult{StatusCode: 200, Body: []byte(`{}`)}
				store.Set(key, result)
				store.Get(key)
				store.Exists(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestIdempotencyKeyGenerator(t *testing.T) {
	gen := NewIdempotencyKeyGenerator("test-secret")
	if gen == nil {
		t.Fatal("NewIdempotencyKeyGenerator returned nil")
	}

	body := []byte(`{"crewId":"123"}`)

	// Same body should produce same key
	key1 := gen.GenerateKey(body)
	key2 := gen.GenerateKey(body)
	if key1 != key2 {
		t.Error("Same body should produce same key")
	}

	// Different body should produce different key
	key3 := gen.GenerateKey([]byte(`{"crewId":"456"}`))
	if key1 == key3 {
		t.Error("Different body should produce different key")
	}

	// Key should be hex-encoded (64 chars for SHA-256)
	if len(key1) != 64 {
		t.Errorf("Key length = %d; want 64 (hex-encoded SHA-256)", len(key1))
	}
}

func TestIdempotencyKeyGeneratorWithContext(t *testing.T) {
	gen := NewIdempotencyKeyGenerator("test-secret")

	body := []byte(`{"data":"same"}`)

	// Same body with different context should produce different keys
	key1 := gen.GenerateKeyWithContext(body, "crew-001")
	key2 := gen.GenerateKeyWithContext(body, "crew-002")
	if key1 == key2 {
		t.Error("Different context should produce different keys")
	}

	// Same body with same context should produce same key
	key3 := gen.GenerateKeyWithContext(body, "crew-001")
	if key1 != key3 {
		t.Error("Same context and body should produce same key")
	}
}

func TestIdempotencyKeyGeneratorDifferentSecrets(t *testing.T) {
	gen1 := NewIdempotencyKeyGenerator("secret-1")
	gen2 := NewIdempotencyKeyGenerator("secret-2")

	body := []byte(`{"crewId":"123"}`)

	key1 := gen1.GenerateKey(body)
	key2 := gen2.GenerateKey(body)

	if key1 == key2 {
		t.Error("Different secrets should produce different keys")
	}
}

func TestIdempotencyMiddleware(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute, 30*time.Second)
	defer store.Close()

	middleware := NewIdempotencyMiddleware(store, "test-secret")
	if middleware == nil {
		t.Fatal("NewIdempotencyMiddleware returned nil")
	}

	body := []byte(`{"crewId":"123","data":"test"}`)

	// Generate key
	key := middleware.GenerateKey(body)
	if key == "" {
		t.Error("GenerateKey should return non-empty key")
	}

	// Check duplicate - should not exist
	_, exists, err := middleware.CheckDuplicate(key)
	if err != nil {
		t.Fatalf("CheckDuplicate failed: %v", err)
	}
	if exists {
		t.Error("Key should not exist initially")
	}

	// Store result
	responseBody := []byte(`{"success":true}`)
	err = middleware.StoreResult(key, 200, responseBody, "application/json")
	if err != nil {
		t.Fatalf("StoreResult failed: %v", err)
	}

	// Check duplicate - should exist now
	result, exists, err := middleware.CheckDuplicate(key)
	if err != nil {
		t.Fatalf("CheckDuplicate failed: %v", err)
	}
	if !exists {
		t.Error("Key should exist after StoreResult")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d; want 200", result.StatusCode)
	}
	if string(result.Body) != string(responseBody) {
		t.Errorf("Body = %s; want %s", result.Body, responseBody)
	}
}

func TestIdempotencyMiddlewareWithContext(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute, 30*time.Second)
	defer store.Close()

	middleware := NewIdempotencyMiddleware(store, "test-secret")

	body := []byte(`{"data":"same-data"}`)

	// Generate keys with different contexts
	key1 := middleware.GenerateKeyWithContext(body, "crew-001")
	key2 := middleware.GenerateKeyWithContext(body, "crew-002")

	// Store result for key1
	middleware.StoreResult(key1, 200, []byte(`{"crew":"001"}`), "application/json")

	// key1 should exist
	_, exists1, _ := middleware.CheckDuplicate(key1)
	if !exists1 {
		t.Error("key1 should exist")
	}

	// key2 should not exist (different context)
	_, exists2, _ := middleware.CheckDuplicate(key2)
	if exists2 {
		t.Error("key2 should not exist (different context)")
	}
}

func TestIdempotencyResultFields(t *testing.T) {
	result := &IdempotencyResult{
		StatusCode:  201,
		Body:        []byte(`{"id":"abc-123"}`),
		ContentType: "application/json; charset=utf-8",
		CreatedAt:   time.Now(),
	}

	if result.StatusCode != 201 {
		t.Errorf("StatusCode = %d; want 201", result.StatusCode)
	}
	if string(result.Body) != `{"id":"abc-123"}` {
		t.Errorf("Body = %s; want %s", result.Body, `{"id":"abc-123"}`)
	}
	if result.ContentType != "application/json; charset=utf-8" {
		t.Errorf("ContentType = %s; want %s", result.ContentType, "application/json; charset=utf-8")
	}
	if result.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestInMemoryIdempotencyStoreSize(t *testing.T) {
	store := NewInMemoryIdempotencyStore(time.Minute, 30*time.Second)
	defer store.Close()

	// Initially empty
	if store.Size() != 0 {
		t.Errorf("Initial size = %d; want 0", store.Size())
	}

	// Add entries
	for i := 0; i < 5; i++ {
		result := &IdempotencyResult{StatusCode: 200, Body: []byte(`{}`)}
		store.Set("key-"+string(rune('0'+i)), result)
	}

	if store.Size() != 5 {
		t.Errorf("Size after adding 5 = %d; want 5", store.Size())
	}

	// Delete one
	store.Delete("key-0")

	if store.Size() != 4 {
		t.Errorf("Size after deleting 1 = %d; want 4", store.Size())
	}
}
