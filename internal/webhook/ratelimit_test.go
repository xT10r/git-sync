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
	"testing"
	"time"

	"git-sync/internal/config"
)

func TestRateLimiterAllow(t *testing.T) {
	// Create a rate limiter with a limit of 2 requests per minute
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			RateLimit: 2,
		},
	}

	rl := NewRateLimiter(cfg)
	defer rl.Stop()

	ip := "192.168.1.1"

	// First request should be allowed
	if !rl.Allow(ip) {
		t.Error("First request should be allowed")
	}

	// Second request should be allowed
	if !rl.Allow(ip) {
		t.Error("Second request should be allowed")
	}

	// Third request should be denied
	if rl.Allow(ip) {
		t.Error("Third request should be denied")
	}
}

func TestRateLimiterExpiration(t *testing.T) {
	// Create a rate limiter with a limit of 1 request per second (for testing)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			RateLimit: 1,
		},
	}

	rl := NewRateLimiter(cfg)
	defer rl.Stop()

	ip := "192.168.1.2"

	// First request should be allowed
	if !rl.Allow(ip) {
		t.Error("First request should be allowed")
	}

	// Second request should be denied
	if rl.Allow(ip) {
		t.Error("Second request should be denied")
	}

	// Wait for requests to expire
	time.Sleep(61 * time.Second)

	// After expiration, request should be allowed again
	if !rl.Allow(ip) {
		t.Error("Request should be allowed after expiration")
	}
}

func TestRateLimiterStats(t *testing.T) {
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			RateLimit: 10,
		},
	}

	rl := NewRateLimiter(cfg)
	defer rl.Stop()

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Make some requests
	rl.Allow(ip1)
	rl.Allow(ip1)
	rl.Allow(ip2)

	// Check stats
	stats := rl.GetStats()
	if stats[ip1] != 2 {
		t.Errorf("Expected 2 requests for IP1, got %d", stats[ip1])
	}
	if stats[ip2] != 1 {
		t.Errorf("Expected 1 request for IP2, got %d", stats[ip2])
	}
}
