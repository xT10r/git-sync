// Copyright 2025 Aleksey Dobshikov
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"sync"
	"time"

	"git-sync/internal/config"
	"git-sync/internal/metrics"
	"git-sync/logger"
)

// RateLimiterRequest represents a request to check or record a rate limit
type RateLimiterRequest struct {
	IP       string
	Response chan bool // Response channel for Allow requests
}

// RateLimitEntry represents the rate limit data for a single IP
type RateLimitEntry struct {
	requests []time.Time
}

// RateLimiter keeps track of request counts per IP with background cleanup
type RateLimiter struct {
	limit        int
	window       time.Duration
	quit         chan struct{}
	requestsChan chan RateLimiterRequest
	mu           sync.RWMutex
	limits       map[string]*RateLimitEntry
}

// NewRateLimiter creates a new rate limiter with background cleanup
func NewRateLimiter(cfg *config.Config) *RateLimiter {
	// Default values
	limit := 10
	window := time.Minute

	// Use config values if available
	if cfg != nil && cfg.Webhook.RateLimit > 0 {
		limit = cfg.Webhook.RateLimit
	}

	rl := &RateLimiter{
		limit:        limit,
		window:       window,
		quit:         make(chan struct{}),
		requestsChan: make(chan RateLimiterRequest, 100), // Buffered channel
		limits:       make(map[string]*RateLimitEntry),
	}

	// Start background goroutine
	go rl.run()

	return rl
}

// Start begins the background cleanup process (already started in NewRateLimiter)
func (rl *RateLimiter) Start() {
	// Cleanup is already started in NewRateLimiter
}

// Stop terminates the background cleanup process
func (rl *RateLimiter) Stop() {
	close(rl.quit)
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	responseChan := make(chan bool, 1)
	req := RateLimiterRequest{
		IP:       ip,
		Response: responseChan,
	}

	// Send request to the rate limiter goroutine
	select {
	case rl.requestsChan <- req:
		// Wait for response
		return <-responseChan
	case <-time.After(5 * time.Second):
		// Timeout - fail open for safety
		logger.Debug("Rate limiter timeout for IP %s, allowing request", ip)
		return true
	}
}

// run is the main goroutine that manages the rate limits
func (rl *RateLimiter) run() {
	cleanupTicker := time.NewTicker(30 * time.Second) // Cleanup every 30 seconds
	defer cleanupTicker.Stop()

	for {
		select {
		case req := <-rl.requestsChan:
			// Handle rate limit check request
			allowed := rl.checkAndRecordRequest(req.IP)
			// Send response back
			select {
			case req.Response <- allowed:
			default:
				// Channel was closed or buffer full, ignore
			}
		case <-cleanupTicker.C:
			// Perform periodic cleanup
			rl.periodicCleanup()
		case <-rl.quit:
			logger.Debug("Rate limiter stopped")
			return
		}
	}
}

// checkAndRecordRequest checks if a request should be allowed and records it
func (rl *RateLimiter) checkAndRecordRequest(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get or create entry for this IP
	entry, exists := rl.limits[ip]
	if !exists {
		entry = &RateLimitEntry{requests: make([]time.Time, 0)}
		rl.limits[ip] = entry
	}

	// Clean up old requests outside the window
	var validRequests []time.Time
	oldCount := len(entry.requests)

	for _, reqTime := range entry.requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Log cleanup information
	if oldCount > len(validRequests) {
		removedCount := oldCount - len(validRequests)
		logger.Debug("Rate limit cleanup for IP %s: removed %d expired requests (older than %v)", ip, removedCount, windowStart.Format("15:04:05"))
	}

	// Update the entry with cleaned requests
	entry.requests = validRequests

	// Check if we're under the limit
	requestCount := len(validRequests)

	logger.Debug("Rate limiting check for IP %s: %d requests in current window (limit: %d)", ip, requestCount, rl.limit)

	// Update metrics with current request count for this IP
	metrics.SetWebhookRequestsPerIP(ip, float64(requestCount))

	if requestCount >= rl.limit {
		logger.Debug("Rate limit exceeded for IP %s: %d/%d requests (next reset at %v)", ip, requestCount, rl.limit, now.Add(rl.window).Format("15:04:05"))
		// Set this IP as rate limited in metrics
		metrics.SetWebhookRateLimitedIP(ip, true)
		return false
	}

	// Add current request
	entry.requests = append(entry.requests, now)

	remaining := rl.limit - requestCount - 1 // -1 because we're adding a new request
	logger.Debug("Rate limit OK for IP %s: %d/%d requests, %d remaining until limit", ip, requestCount, rl.limit, remaining)

	// If this is the first request after cleanup, explicitly log that the rate limit period has reset
	if requestCount == 0 && oldCount > 0 {
		logger.Debug("Rate limit period reset for IP %s: all previous requests expired", ip)
		// Remove rate limited status since limit has been reset
		metrics.SetWebhookRateLimitedIP(ip, false)
	} else if requestCount == 0 {
		// Ensure IP is not marked as rate limited if it has no requests
		metrics.SetWebhookRateLimitedIP(ip, false)
	}

	return true
}

// periodicCleanup performs cleanup of expired requests across all IPs
func (rl *RateLimiter) periodicCleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	totalRemoved := 0
	ipsWithCleanup := 0

	for ip, entry := range rl.limits {
		var validRequests []time.Time
		oldCount := len(entry.requests)

		for _, reqTime := range entry.requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}

		// Update requests for this IP
		entry.requests = validRequests

		// Log cleanup information if any requests were removed
		if oldCount > len(validRequests) {
			removedCount := oldCount - len(validRequests)
			totalRemoved += removedCount
			ipsWithCleanup++
			logger.Debug("Rate limiter cleanup for IP %s: removed %d expired requests", ip, removedCount)
		}

		// Update metrics with current request count
		requestCount := len(validRequests)
		metrics.SetWebhookRequestsPerIP(ip, float64(requestCount))

		// If IP is no longer rate limited, update metrics
		if requestCount < rl.limit {
			metrics.SetWebhookRateLimitedIP(ip, false)
		}

		// Remove IP entry entirely if no valid requests remain
		if len(validRequests) == 0 {
			delete(rl.limits, ip)
			logger.Debug("Periodic cleanup: removed IP %s from rate limiter (no active requests)", ip)
			// Remove metrics for this IP since it has no active requests
			metrics.RemoveWebhookRateLimitedIP(ip)
		}
	}

	if totalRemoved > 0 {
		logger.Debug("Rate limiter periodic cleanup completed: removed %d expired requests from %d IPs", totalRemoved, ipsWithCleanup)
	}
}

// GetStats returns statistics about the current state of the rate limiter
func (rl *RateLimiter) GetStats() map[string]int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := make(map[string]int)
	for ip, entry := range rl.limits {
		stats[ip] = len(entry.requests)
	}
	return stats
}
