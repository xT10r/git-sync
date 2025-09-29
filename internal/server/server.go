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
Package server provides HTTP server functionality for the git-sync service.
It uses the router package to manage all HTTP endpoints.
*/
package server

import (
	"context"

	"git-sync/api"
	"git-sync/internal/config"
	"git-sync/internal/handlers"
	"git-sync/internal/router"
	"git-sync/logger"
)

// Server wraps the router and provides HTTP server functionality
type Server struct {
	router *router.Router
	config *config.Config
}

// NewServer creates a new HTTP server instance
func NewServer(cfg *config.Config) *Server {
	return &Server{
		router: router.New(cfg),
		config: cfg,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) {
	// Setup middleware
	s.router.SetupMiddleware()

	// Set config for handlers
	handlers.SetConfig(s.config)

	// Register routes from different packages
	api.RegisterRoutes(s.router)
	handlers.RegisterRoutes(s.router)

	// Start the server
	s.router.StartServer(ctx)
	logger.Info("HTTP server started")
	logger.Debug("HTTP server started with config: %+v", s.config)
}
