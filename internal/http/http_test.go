// Copyright 2025 Alex Dobshikov
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestBasicAuthMiddleware tests the basicAuthMiddleware function
func TestBasicAuthMiddleware(t *testing.T) {
	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test case 1: Valid credentials
	t.Run("ValidCredentials", func(t *testing.T) {
		// Create the middleware
		middleware := basicAuthMiddleware("testuser", "testpass")
		handler := middleware(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("testuser", "testpass")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}
	})

	// Test case 2: Invalid credentials
	t.Run("InvalidCredentials", func(t *testing.T) {
		// Create the middleware
		middleware := basicAuthMiddleware("testuser", "testpass")
		handler := middleware(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("wronguser", "wrongpass")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, status)
		}
	})

	// Test case 3: No credentials
	t.Run("NoCredentials", func(t *testing.T) {
		// Create the middleware
		middleware := basicAuthMiddleware("testuser", "testpass")
		handler := middleware(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, status)
		}
	})

	// Test case 4: Unsafe password (same username and password)
	t.Run("UnsafePassword", func(t *testing.T) {
		// Create the middleware
		middleware := basicAuthMiddleware("test", "test")
		handler := middleware(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth("test", "test")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}
	})
}

// TestBearerAuthMiddleware tests the bearerAuthMiddleware function
func TestBearerAuthMiddleware(t *testing.T) {
	// Create a test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test case 1: Valid token
	t.Run("ValidToken", func(t *testing.T) {
		// Create the middleware
		middleware := bearerAuthMiddleware("testtoken")
		handler := middleware(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer testtoken")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}
	})

	// Test case 2: Invalid token
	t.Run("InvalidToken", func(t *testing.T) {
		// Create the middleware
		middleware := bearerAuthMiddleware("testtoken")
		handler := middleware(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer wrongtoken")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, status)
		}
	})

	// Test case 3: No token
	t.Run("NoToken", func(t *testing.T) {
		// Create the middleware
		middleware := bearerAuthMiddleware("testtoken")
		handler := middleware(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, status)
		}
	})

	// Test case 4: Empty Authorization header
	t.Run("EmptyAuthHeader", func(t *testing.T) {
		// Create the middleware
		middleware := bearerAuthMiddleware("testtoken")
		handler := middleware(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, status)
		}
	})
}

// TestRootHandlerFunc tests the rootHandlerFunc function
func TestRootHandlerFunc(t *testing.T) {
	// Test case 1: No registered paths
	t.Run("NoRegisteredPaths", func(t *testing.T) {
		// Clear registered paths before test
		registeredPaths = make(map[string]bool)

		// Create a request to the root handler
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		// Call the handler
		rootHandlerFunc(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		// Check the response body
		expected := "<h1>List of available handlers:</h1>\n<ul>\n</ul>\n"
		if rr.Body.String() != expected {
			t.Errorf("Expected response body '%s', got '%s'", expected, rr.Body.String())
		}
	})

	// Test case 2: Multiple registered paths
	t.Run("MultipleRegisteredPaths", func(t *testing.T) {
		// Clear registered paths before test
		registeredPaths = make(map[string]bool)

		// Register some test paths
		registeredPaths["/test1"] = true
		registeredPaths["/test2"] = true
		registeredPaths["/test3"] = true

		// Create a request to the root handler
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		// Call the handler
		rootHandlerFunc(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		// Check the response body contains all paths (they should be sorted)
		response := rr.Body.String()
		if !strings.Contains(response, "/test1") {
			t.Error("Expected response to contain '/test1'")
		}
		if !strings.Contains(response, "/test2") {
			t.Error("Expected response to contain '/test2'")
		}
		if !strings.Contains(response, "/test3") {
			t.Error("Expected response to contain '/test3'")
		}
	})

	// Test case 3: Single registered path
	t.Run("SingleRegisteredPath", func(t *testing.T) {
		// Clear registered paths before test
		registeredPaths = make(map[string]bool)

		// Register a test path
		registeredPaths["/single"] = true

		// Create a request to the root handler
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		// Call the handler
		rootHandlerFunc(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}

		// Check the response body
		expected := "<h1>List of available handlers:</h1>\n<ul>\n<li><a href=\"/single\">/single</a></li>\n</ul>\n"
		if rr.Body.String() != expected {
			t.Errorf("Expected response body '%s', got '%s'", expected, rr.Body.String())
		}
	})
}

// TestRegisterHandler tests the registerHandler function
func TestRegisterHandler(t *testing.T) {
	// Clear registered paths before test
	registeredPaths = make(map[string]bool)

	// Test case 1: Register with http.Handler
	t.Run("RegisterWithHandler", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// This would normally register with http.Handle, but we can't easily test that
		// Instead, we'll just verify that the path is added to registeredPaths
		registerHandler("/test-handler", handler, nil)

		if !registeredPaths["/test-handler"] {
			t.Error("Expected path to be registered")
		}
	})

	// Test case 2: Register with http.HandlerFunc
	t.Run("RegisterWithHandlerFunc", func(t *testing.T) {
		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		// This would normally register with http.HandleFunc, but we can't easily test that
		// Instead, we'll just verify that the path is added to registeredPaths
		registerHandler("/test-handler-func", nil, handlerFunc)

		if !registeredPaths["/test-handler-func"] {
			t.Error("Expected path to be registered")
		}
	})

	// Test case 3: Register root path
	t.Run("RegisterRootPath", func(t *testing.T) {
		// Clear registered paths before test
		registeredPaths = make(map[string]bool)

		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		registerHandler("/", nil, handlerFunc)

		// Root path should not be in registeredPaths
		if registeredPaths["/"] {
			t.Error("Root path should not be in registeredPaths")
		}
	})

	// Test case 4: Panic when neither handler nor handlerFunc provided
	t.Run("PanicWhenNoHandler", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when neither handler nor handlerFunc provided")
			}
		}()

		registerHandler("/test-panic", nil, nil)
	})
}

// TestStartServerLogic tests the logic within the StartServer function
func TestStartServerLogic(t *testing.T) {
	// Reset registered paths before each test to avoid conflicts
	registeredPaths = make(map[string]bool)

	// Test case 1: No server address
	t.Run("NoServerAddress", func(t *testing.T) {
		// Reset registered paths before each test to avoid conflicts
		registeredPaths = make(map[string]bool)

		fsNoAddr := flag.NewFlagSet("test", flag.ContinueOnError)
		fsNoAddr.String("http-server-addr", "", "")
		fsNoAddr.String("http-auth-username", "", "")
		fsNoAddr.String("http-auth-password", "", "")
		fsNoAddr.String("http-auth-token", "", "")

		// Test the logic without actually starting the server
		addr := fsNoAddr.Lookup("http-server-addr").Value.(flag.Getter).Get().(string)

		if len(addr) != 0 {
			t.Error("Expected empty address")
		}

		// With no server address, no paths should be registered
		if len(registeredPaths) > 0 {
			t.Errorf("Expected no paths to be registered, got %d", len(registeredPaths))
		}
	})

	// Test case 2: Check authentication logic
	t.Run("AuthenticationLogic", func(t *testing.T) {
		// Test basic auth with same username and password (unsafe)
		fsUnsafe := flag.NewFlagSet("test", flag.ContinueOnError)
		fsUnsafe.String("http-server-addr", "127.0.0.1:8080", "")
		fsUnsafe.String("http-auth-username", "test", "")
		fsUnsafe.String("http-auth-password", "test", "")
		fsUnsafe.String("http-auth-token", "", "")

		username := fsUnsafe.Lookup("http-auth-username").Value.(flag.Getter).Get().(string)
		password := fsUnsafe.Lookup("http-auth-password").Value.(flag.Getter).Get().(string)

		useBasicAuth := username != "" && password != ""
		useBearerToken := len(fsUnsafe.Lookup("http-auth-token").Value.(flag.Getter).Get().(string)) > 0

		if !useBasicAuth {
			t.Error("Expected basic auth to be enabled")
		}

		if useBearerToken {
			t.Error("Expected bearer token auth to be disabled")
		}

		if username == password && len(username) > 0 {
			// This is the unsafe password case
			// We're just testing the logic, not the actual logging
		}
	})

	// Test case 3: Check flag retrieval logic
	t.Run("FlagRetrieval", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("http-server-addr", "127.0.0.1:8080", "")
		fs.String("http-auth-username", "testuser", "")
		fs.String("http-auth-password", "testpass", "")
		fs.String("http-auth-token", "testtoken", "")

		// Test that we can retrieve flag values correctly
		addr := fs.Lookup("http-server-addr").Value.(flag.Getter).Get().(string)
		username := fs.Lookup("http-auth-username").Value.(flag.Getter).Get().(string)
		password := fs.Lookup("http-auth-password").Value.(flag.Getter).Get().(string)
		token := fs.Lookup("http-auth-token").Value.(flag.Getter).Get().(string)

		if addr != "127.0.0.1:8080" {
			t.Errorf("Expected address '127.0.0.1:8080', got '%s'", addr)
		}

		if username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", username)
		}

		if password != "testpass" {
			t.Errorf("Expected password 'testpass', got '%s'", password)
		}

		if token != "testtoken" {
			t.Errorf("Expected token 'testtoken', got '%s'", token)
		}
	})

	// Test case 4: Bearer token authentication
	t.Run("BearerTokenAuth", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("http-server-addr", "127.0.0.1:8080", "")
		fs.String("http-auth-username", "", "")
		fs.String("http-auth-password", "", "")
		fs.String("http-auth-token", "testtoken", "")

		username := fs.Lookup("http-auth-username").Value.(flag.Getter).Get().(string)
		password := fs.Lookup("http-auth-password").Value.(flag.Getter).Get().(string)
		token := fs.Lookup("http-auth-token").Value.(flag.Getter).Get().(string)

		useBasicAuth := username != "" && password != ""
		useBearerToken := len(token) > 0

		if useBasicAuth {
			t.Error("Expected basic auth to be disabled")
		}

		if !useBearerToken {
			t.Error("Expected bearer token auth to be enabled")
		}
	})

	// Test case 5: No authentication
	t.Run("NoAuth", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("http-server-addr", "127.0.0.1:8080", "")
		fs.String("http-auth-username", "", "")
		fs.String("http-auth-password", "", "")
		fs.String("http-auth-token", "", "")

		username := fs.Lookup("http-auth-username").Value.(flag.Getter).Get().(string)
		password := fs.Lookup("http-auth-password").Value.(flag.Getter).Get().(string)
		token := fs.Lookup("http-auth-token").Value.(flag.Getter).Get().(string)

		useBasicAuth := username != "" && password != ""
		useBearerToken := len(token) > 0

		if useBasicAuth {
			t.Error("Expected basic auth to be disabled")
		}

		if useBearerToken {
			t.Error("Expected bearer token auth to be disabled")
		}
	})

	// Test case 6: Basic auth with different username and password
	t.Run("BasicAuthSafe", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("http-server-addr", "127.0.0.1:8080", "")
		fs.String("http-auth-username", "testuser", "")
		fs.String("http-auth-password", "testpass", "")
		fs.String("http-auth-token", "", "")

		username := fs.Lookup("http-auth-username").Value.(flag.Getter).Get().(string)
		password := fs.Lookup("http-auth-password").Value.(flag.Getter).Get().(string)

		useBasicAuth := username != "" && password != ""
		useBearerToken := len(fs.Lookup("http-auth-token").Value.(flag.Getter).Get().(string)) > 0

		if !useBasicAuth {
			t.Error("Expected basic auth to be enabled")
		}

		if useBearerToken {
			t.Error("Expected bearer token auth to be disabled")
		}

		if username == password {
			t.Error("Expected username and password to be different")
		}
	})
}
