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

package router

import (
	"context"
	"fmt"
	"git-sync/internal/config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewRouter(t *testing.T) {
	cfg := &config.Config{}
	router := New(cfg)

	if router == nil {
		t.Error("New() should not return nil")
		return // Early return to avoid nil pointer dereference
	}

	if router.routes == nil {
		t.Error("Router routes map should not be nil")
		return // Early return to avoid nil pointer dereference
	}

	if router.mux == nil {
		t.Error("Router mux should not be nil")
		return // Early return to avoid nil pointer dereference
	}

	if len(router.routes) != 0 {
		t.Error("New router should have no routes registered")
	}
}

func TestRegisterRoute(t *testing.T) {
	cfg := &config.Config{}
	router := New(cfg)

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "test handler"); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}

	router.RegisterRoute("/test", "GET", "Test route", nil, handlerFunc)

	if len(router.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(router.routes))
	}

	route, exists := router.routes["/test"]
	if !exists {
		t.Error("Route /test should exist")
	}

	if route.Path != "/test" {
		t.Errorf("Expected path /test, got %s", route.Path)
	}

	if route.Method != "GET" {
		t.Errorf("Expected method GET, got %s", route.Method)
	}

	if route.Description != "Test route" {
		t.Errorf("Expected description 'Test route', got %s", route.Description)
	}
}

func TestGetRoutes(t *testing.T) {
	cfg := &config.Config{}
	router := New(cfg)

	// Register multiple routes
	handlerFunc1 := func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "test handler 1"); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}

	handlerFunc2 := func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "test handler 2"); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}

	router.RegisterRoute("/test1", "GET", "Test route 1", nil, handlerFunc1)
	router.RegisterRoute("/test2", "POST", "Test route 2", nil, handlerFunc2)

	routes := router.GetRoutes()

	if len(routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(routes))
	}

	// Routes should be sorted by path
	if routes[0].Path != "/test1" {
		t.Error("Routes should be sorted by path")
	}

	if routes[1].Path != "/test2" {
		t.Error("Routes should be sorted by path")
	}
}

func TestGetMux(t *testing.T) {
	cfg := &config.Config{}
	router := New(cfg)

	mux := router.GetMux()

	if mux == nil {
		t.Error("GetMux() should not return nil")
	}

	if mux != router.mux {
		t.Error("GetMux() should return the router's mux")
	}
}

func TestSetupMiddleware(t *testing.T) {
	// Test with no authentication
	cfg := &config.Config{}
	cfg.HttpServer.Auth.Username = ""
	cfg.HttpServer.Auth.Password = ""
	cfg.HttpServer.Auth.Token = ""

	router := New(cfg)
	router.SetupMiddleware()

	// Test with basic auth
	cfg.HttpServer.Auth.Username = "testuser"
	cfg.HttpServer.Auth.Password = "testpass"

	router2 := New(cfg)
	router2.SetupMiddleware()

	// Test with bearer token
	cfg.HttpServer.Auth.Username = ""
	cfg.HttpServer.Auth.Password = ""
	cfg.HttpServer.Auth.Token = "testtoken"

	router3 := New(cfg)
	router3.SetupMiddleware()
}

func TestApplyMiddleware(t *testing.T) {
	cfg := &config.Config{}
	router := New(cfg)

	// Test applying middleware to a handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "test"); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Apply middleware (should work even with empty middleware chain)
	wrappedHandler := router.ApplyMiddleware(handler)

	if wrappedHandler == nil {
		t.Error("ApplyMiddleware should not return nil")
	}
}

func TestRootHandler(t *testing.T) {
	cfg := &config.Config{}
	router := New(cfg)

	// Register a test route
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "test handler"); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}

	router.RegisterRoute("/test", "GET", "Test route", nil, handlerFunc)

	// Create a test request to the root handler
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Call the root handler directly
	router.rootHandler(w, req)

	// Check the response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check that response body is not empty (without printing it to console)
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}

	// The response should contain the registered route
	// We check for the presence of the route without printing to console
	if !strings.Contains(body, "/test") {
		t.Error("Response body should contain registered route")
	}
}

func TestBasicAuthMiddleware(t *testing.T) {
	cfg := &config.Config{}
	cfg.HttpServer.Auth.Username = "testuser"
	cfg.HttpServer.Auth.Password = "testpass"

	router := New(cfg)

	// Apply basic auth middleware
	protectedHandler := router.basicAuthMiddleware("testuser", "testpass")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "protected content"); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	// Test with correct credentials
	req := httptest.NewRequest("GET", "/protected", nil)
	req.SetBasicAuth("testuser", "testpass")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with correct credentials, got %d", resp.StatusCode)
	}

	// Test with incorrect credentials
	req = httptest.NewRequest("GET", "/protected", nil)
	req.SetBasicAuth("wronguser", "wrongpass")
	w = httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with incorrect credentials, got %d", resp.StatusCode)
	}
}

func TestBearerAuthMiddleware(t *testing.T) {
	cfg := &config.Config{}
	router := New(cfg)

	// Apply bearer auth middleware
	protectedHandler := router.bearerAuthMiddleware("testtoken")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "protected content"); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	// Test with correct token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with correct token, got %d", resp.StatusCode)
	}

	// Test with incorrect token
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer wrongtoken")
	w = httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with incorrect token, got %d", resp.StatusCode)
	}

	// Test with missing Authorization header
	req = httptest.NewRequest("GET", "/protected", nil)
	w = httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with missing Authorization header, got %d", resp.StatusCode)
	}
}

func TestStartServer(t *testing.T) {
	// Create a test config with a random available port
	cfg := &config.Config{}
	cfg.HttpServer.Addr = ":0" // Use port 0 to get an available port
	cfg.HttpServer.Auth.Username = ""
	cfg.HttpServer.Auth.Password = ""
	cfg.HttpServer.Auth.Token = ""

	router := New(cfg)

	// Create a context with timeout for the test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the server in a goroutine
	go router.StartServer(ctx)

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// The test passes if we reach this point without panicking
	// In a more comprehensive test, we would make HTTP requests to verify the server is running
}
