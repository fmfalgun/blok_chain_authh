package common

import (
	"fmt"
	"sync"
	"time"
)

const (
	DefaultRequestsPerMinute = 60
	DefaultBanDurationMinutes = 5
	DefaultCleanupInterval = 10 * time.Minute
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	// requestCounts maps deviceID to request count
	requestCounts map[string]*RequestCounter
	// bannedDevices maps deviceID to ban expiration time
	bannedDevices map[string]time.Time
	// Configuration
	requestsPerMinute int
	banDurationMinutes int
	// Mutex for thread-safe operations
	mu sync.RWMutex
	// Cleanup ticker
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// RequestCounter tracks request counts for a device
type RequestCounter struct {
	Count        int
	WindowStart  time.Time
	ViolationCount int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int, banDurationMinutes int) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = DefaultRequestsPerMinute
	}
	if banDurationMinutes <= 0 {
		banDurationMinutes = DefaultBanDurationMinutes
	}

	rl := &RateLimiter{
		requestCounts:      make(map[string]*RequestCounter),
		bannedDevices:      make(map[string]time.Time),
		requestsPerMinute:  requestsPerMinute,
		banDurationMinutes: banDurationMinutes,
		stopCleanup:        make(chan bool),
	}

	// Start cleanup goroutine
	rl.startCleanup()

	return rl
}

// AllowRequest checks if a request from a device should be allowed
func (rl *RateLimiter) AllowRequest(deviceID string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check if device is banned
	if banExpiry, isBanned := rl.bannedDevices[deviceID]; isBanned {
		if time.Now().Before(banExpiry) {
			remainingTime := time.Until(banExpiry).Round(time.Second)
			return false, fmt.Errorf("device is temporarily banned (remaining: %v)", remainingTime)
		}
		// Ban expired, remove from banned list
		delete(rl.bannedDevices, deviceID)
	}

	// Get or create request counter for device
	counter, exists := rl.requestCounts[deviceID]
	if !exists {
		counter = &RequestCounter{
			Count:       0,
			WindowStart: time.Now(),
			ViolationCount: 0,
		}
		rl.requestCounts[deviceID] = counter
	}

	// Check if we need to reset the window
	if time.Since(counter.WindowStart) > time.Minute {
		counter.Count = 0
		counter.WindowStart = time.Now()
	}

	// Check rate limit
	if counter.Count >= rl.requestsPerMinute {
		counter.ViolationCount++

		// Ban device after 3 violations
		if counter.ViolationCount >= 3 {
			banExpiry := time.Now().Add(time.Duration(rl.banDurationMinutes) * time.Minute)
			rl.bannedDevices[deviceID] = banExpiry
			return false, fmt.Errorf("device banned for %d minutes due to repeated rate limit violations", rl.banDurationMinutes)
		}

		return false, fmt.Errorf("rate limit exceeded (%d/%d requests per minute)", counter.Count, rl.requestsPerMinute)
	}

	// Increment counter and allow request
	counter.Count++
	return true, nil
}

// GetDeviceStats returns statistics for a device
func (rl *RateLimiter) GetDeviceStats(deviceID string) map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := make(map[string]interface{})

	// Check ban status
	if banExpiry, isBanned := rl.bannedDevices[deviceID]; isBanned {
		stats["banned"] = true
		stats["banExpiry"] = banExpiry
		stats["remainingBanTime"] = time.Until(banExpiry).String()
	} else {
		stats["banned"] = false
	}

	// Get request counter stats
	if counter, exists := rl.requestCounts[deviceID]; exists {
		stats["requestCount"] = counter.Count
		stats["windowStart"] = counter.WindowStart
		stats["violationCount"] = counter.ViolationCount
		stats["windowAge"] = time.Since(counter.WindowStart).String()
	}

	stats["limit"] = rl.requestsPerMinute
	stats["limitPeriod"] = "1 minute"

	return stats
}

// UnbanDevice manually removes a device from the banned list
func (rl *RateLimiter) UnbanDevice(deviceID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.bannedDevices, deviceID)

	// Reset violation count
	if counter, exists := rl.requestCounts[deviceID]; exists {
		counter.ViolationCount = 0
	}
}

// ResetDevice resets all counters for a device
func (rl *RateLimiter) ResetDevice(deviceID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.requestCounts, deviceID)
	delete(rl.bannedDevices, deviceID)
}

// startCleanup starts the cleanup goroutine
func (rl *RateLimiter) startCleanup() {
	rl.cleanupTicker = time.NewTicker(DefaultCleanupInterval)

	go func() {
		for {
			select {
			case <-rl.cleanupTicker.C:
				rl.cleanup()
			case <-rl.stopCleanup:
				return
			}
		}
	}()
}

// cleanup removes expired entries
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Remove expired bans
	for deviceID, banExpiry := range rl.bannedDevices {
		if now.After(banExpiry) {
			delete(rl.bannedDevices, deviceID)
		}
	}

	// Remove old request counters (older than 1 hour)
	for deviceID, counter := range rl.requestCounts {
		if now.Sub(counter.WindowStart) > time.Hour {
			delete(rl.requestCounts, deviceID)
		}
	}
}

// Stop stops the rate limiter and cleanup goroutine
func (rl *RateLimiter) Stop() {
	if rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
	}
	if rl.stopCleanup != nil {
		close(rl.stopCleanup)
	}
}

// GetStats returns overall statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"totalDevices":      len(rl.requestCounts),
		"bannedDevices":     len(rl.bannedDevices),
		"requestsPerMinute": rl.requestsPerMinute,
		"banDurationMinutes": rl.banDurationMinutes,
	}
}
