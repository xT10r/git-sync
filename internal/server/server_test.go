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

package server

import (
	"context"
	"git-sync/internal/config"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{}
	cfg.HttpServer.Addr = ":8080"

	server := NewServer(cfg)

	if server == nil {
		t.Error("NewServer() should not return nil")
		return // Early return to avoid nil pointer dereference
	}

	if server.config != cfg {
		t.Error("Server config not set correctly")
	}

	if server.router == nil {
		t.Error("Server router should not be nil")
	}
}

func TestServerStart(t *testing.T) {
	// Create a test config with a random available port
	cfg := &config.Config{}
	cfg.HttpServer.Addr = ":0" // Use port 0 to get an available port
	cfg.HttpServer.Auth.Username = ""
	cfg.HttpServer.Auth.Password = ""
	cfg.HttpServer.Auth.Token = ""

	server := NewServer(cfg)

	// Create a context with timeout for the test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the server in a goroutine
	go server.Start(ctx)

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// The test passes if we reach this point without panicking
	// In a more comprehensive test, we would make HTTP requests to verify the server is running
}
