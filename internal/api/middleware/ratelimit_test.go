package middleware

import (
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)
	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
	if rl.limit != 100 {
		t.Errorf("limit = %d; want 100", rl.limit)
	}
	if rl.window != time.Minute {
		t.Errorf("window = %v; want %v", rl.window, time.Minute)
	}
}

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, time.Second)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !rl.Allow("crew-001") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rl.Allow("crew-001") {
		t.Error("4th request should be denied")
	}
}

func TestRateLimiterAllowDifferentKeys(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	// Both crews should have independent limits
	if !rl.Allow("crew-001") {
		t.Error("First request for crew-001 should be allowed")
	}
	if !rl.Allow("crew-002") {
		t.Error("First request for crew-002 should be allowed")
	}
	if !rl.Allow("crew-001") {
		t.Error("Second request for crew-001 should be allowed")
	}
	if !rl.Allow("crew-002") {
		t.Error("Second request for crew-002 should be allowed")
	}

	// 3rd request for each should be denied
	if rl.Allow("crew-001") {
		t.Error("3rd request for crew-001 should be denied")
	}
	if rl.Allow("crew-002") {
		t.Error("3rd request for crew-002 should be denied")
	}
}

func TestRateLimiterRemaining(t *testing.T) {
	rl := NewRateLimiter(5, time.Second)

	// Initial remaining should be limit
	if remaining := rl.Remaining("crew-001"); remaining != 5 {
		t.Errorf("Remaining = %d; want 5", remaining)
	}

	// After 2 requests
	rl.Allow("crew-001")
	rl.Allow("crew-001")

	if remaining := rl.Remaining("crew-001"); remaining != 3 {
		t.Errorf("Remaining = %d; want 3", remaining)
	}
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	rl := NewRateLimiter(2, 100*time.Millisecond)

	// Use up the limit
	rl.Allow("crew-001")
	rl.Allow("crew-001")

	// Should be denied
	if rl.Allow("crew-001") {
		t.Error("3rd request should be denied before window expires")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	if !rl.Allow("crew-001") {
		t.Error("Request should be allowed after window expires")
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				rl.Allow("crew-001")
				rl.Remaining("crew-001")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent test timed out")
		}
	}
}
