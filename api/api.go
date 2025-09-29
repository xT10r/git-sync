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

/*
Package api contains request handlers for various HTTP REST API endpoints.
It defines data models for responses. Uses the built-in Go HTTP server.
Thus, it forms an API for monitoring and managing git-sync.
*/

package api

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"git-sync/internal/router"
	"git-sync/internal/version"
	"git-sync/logger"
)

// StatusResponse represents the response structure for status API
type StatusResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// HealthResponse represents the response structure for health check API
type HealthResponse struct {
	Status     string `json:"status"`
	Timestamp  string `json:"timestamp"`
	Version    string `json:"version"`
	GoVersion  string `json:"go_version"`
	Goroutines int    `json:"goroutines"`
}

// VersionInfo represents the response structure for version API
type VersionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	Dirty   string `json:"dirty"`
}

// RegisterRoutes registers all API routes with the router
func RegisterRoutes(r *router.Router) {
	r.RegisterRoute("/status", "GET", "Application status information", nil, handlerStatus)
	r.RegisterRoute("/health", "GET", "Detailed health information", nil, handlerHealth)
	r.RegisterRoute("/ready", "GET", "Readiness status for Kubernetes probes", nil, handlerReady)
	r.RegisterRoute("/version", "GET", "Version information", nil, handlerVersion)
}

// handlerStatus returns the basic status of the application
func handlerStatus(w http.ResponseWriter, r *http.Request) {
	logger.Debug("API /status endpoint requested from %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	resp := StatusResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   version.Version, // Use actual version from build info
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Debug("Error encoding /status response: %v", err)
	}
}

// handlerHealth returns detailed health information about the application
func handlerHealth(w http.ResponseWriter, r *http.Request) {
	logger.Debug("API /health endpoint requested from %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	resp := HealthResponse{
		Status:     "healthy",
		Timestamp:  time.Now().Format(time.RFC3339),
		Version:    version.Version, // Use actual version from build info
		GoVersion:  runtime.Version(),
		Goroutines: runtime.NumGoroutine(),
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Debug("Error encoding /health response: %v", err)
	}
}

// handlerReady returns readiness status for Kubernetes probes
func handlerReady(w http.ResponseWriter, r *http.Request) {
	logger.Debug("API /ready endpoint requested from %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// For now, we're always ready. In a more complex application,
	// this would check dependencies like database connections, etc.
	resp := StatusResponse{
		Status:    "ready",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   version.Version, // Use actual version from build info
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Debug("Error encoding /ready response: %v", err)
	}
}

// handlerVersion returns version information about the application
func handlerVersion(w http.ResponseWriter, r *http.Request) {
	logger.Debug("API /version endpoint requested from %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-App-Version", version.Version)

	resp := VersionInfo{
		Version: version.Version,
		Commit:  version.Commit,
		Date:    version.Date,
		Dirty:   version.Dirty,
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Debug("Error encoding /version response: %v", err)
	}
}
