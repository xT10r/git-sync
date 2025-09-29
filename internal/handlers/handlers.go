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

package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"git-sync/internal/config"
	"git-sync/internal/metrics"
	"git-sync/internal/router"
	"git-sync/internal/webhook"
	"git-sync/logger"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// WebhookTriggeredMessage is the message returned when webhook is triggered
const WebhookTriggeredMessage = "Synchronization triggered by webhook"

// WebhookCh is the channel for webhook signals
var WebhookCh = make(chan struct {
	ClientIP  string
	Timestamp time.Time
	UserAgent string
}, 1)

// webhookMu is a mutex to prevent concurrent webhook requests
var webhookMu sync.Mutex

// rateLimiter is the global rate limiter for webhook requests
var rateLimiter *webhook.RateLimiter

// SetConfig sets the configuration for the handlers and initializes the rate limiter
func SetConfig(cfg *config.Config) {
	// Initialize the rate limiter with the provided configuration
	if rateLimiter == nil {
		rateLimiter = webhook.NewRateLimiter(cfg)
	}
}

// RegisterRoutes registers all handler routes with the router
func RegisterRoutes(r *router.Router) {
	// Register metrics handler
	metricsHandler := promhttp.Handler()
	r.RegisterRoute("/metrics", "GET", "Prometheus metrics", metricsHandler, nil)

	// Register webhook handler
	r.RegisterRoute("/webhook", "POST", "Trigger manual synchronization", nil, WebhookHandlerFunc)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// WebhookHandlerFunc handles webhook requests
func WebhookHandlerFunc(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Set content type
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Validate request method
	if r.Method != http.MethodPost {
		logger.Debug("Webhook request rejected: invalid method %s", r.Method)
		metrics.IncrementWebhookRequest("bad_request")
		sendErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Only POST method is allowed")
		metrics.ObserveWebhookRequestDuration(time.Since(startTime).Seconds())
		return
	}

	// Validate request content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && contentType != "application/json" {
		logger.Debug("Webhook request rejected: invalid content type %s", contentType)
		metrics.IncrementWebhookRequest("bad_request")
		sendErrorResponse(w, http.StatusBadRequest, "Bad Request", "Invalid content type")
		metrics.ObserveWebhookRequestDuration(time.Since(startTime).Seconds())
		return
	}

	// Get client IP
	clientIP := getClientIP(r)

	// Get user agent
	userAgent := r.UserAgent()

	// Log request details
	logger.Debug("Webhook triggered from IP: %s, User-Agent: %s", clientIP, userAgent)

	// Check rate limit
	logger.Debug("Checking rate limit for IP: %s", clientIP)
	if !rateLimiter.Allow(clientIP) {
		logger.Debug("Webhook request from %s rejected: rate limit exceeded", clientIP)
		metrics.IncrementWebhookRequest("rate_limited")
		sendErrorResponse(w, http.StatusTooManyRequests, "Too Many Requests", "Rate limit exceeded")
		metrics.ObserveWebhookRequestDuration(time.Since(startTime).Seconds())
		return
	}

	// Try to acquire lock to prevent concurrent webhook requests
	if !webhookMu.TryLock() {
		logger.Debug("Webhook request from %s rejected: service busy", clientIP)
		metrics.IncrementWebhookRequest("service_unavailable")
		sendErrorResponse(w, http.StatusServiceUnavailable, "Service Unavailable", "Synchronization already in progress")
		metrics.ObserveWebhookRequestDuration(time.Since(startTime).Seconds())
		return
	}

	// Release lock when function exits
	defer webhookMu.Unlock()

	// Send signal to webhook channel
	WebhookCh <- struct {
		ClientIP  string
		Timestamp time.Time
		UserAgent string
	}{
		ClientIP:  clientIP,
		Timestamp: time.Now(),
		UserAgent: userAgent,
	}

	// Send success response
	response := WebhookResponse{
		Message:    WebhookTriggeredMessage,
		Time:       time.Now(),
		Repository: "main-repo", // TODO: Get actual repository name
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Debug("Error encoding webhook response: %v", err)
		// Even if we can't encode JSON, we've already sent the status code
		return
	}

	metrics.IncrementWebhookRequest("success")
	logger.Debug("Webhook response sent successfully to %s", clientIP)
	logger.Debug("Webhook request processed in %v", time.Since(startTime))
	metrics.ObserveWebhookRequestDuration(time.Since(startTime).Seconds())
}

// sendErrorResponse sends a standardized error response
func sendErrorResponse(w http.ResponseWriter, statusCode int, errorType, message string) {
	response := WebhookErrorResponse{
		Error:   errorType,
		Message: message,
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Debug("Error encoding error response: %v", err)
		// Fallback to plain text error
		http.Error(w, message, statusCode)
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple are provided
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// WebhookResponse represents the response structure for successful webhook requests
type WebhookResponse struct {
	Message    string    `json:"message"`
	Time       time.Time `json:"time"`
	Repository string    `json:"repository"`
}

// WebhookErrorResponse represents the response structure for error webhook requests
type WebhookErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
