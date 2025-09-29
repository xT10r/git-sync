# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Breaking

- Removed deprecated `internal/http` package in favor of new modular server architecture
- Refactored HTTP server implementation to use new `server` and `router` packages
- Changed main application startup sequence to use server package instead of direct HTTP server initialization
- Removed direct HTTP server startup from `cmd/main.go` in favor of server package
- Modified configuration system to support multi-repository setups with individual configurations

### Added

- Multi-repository synchronization support with individual configuration per repository
- New `server` package for HTTP server management and lifecycle control
- New `router` package for centralized routing and middleware management
- Webhook support with rate limiting capabilities
- Enhanced configuration system with repository-specific settings
- Improved error handling and validation for repository configurations
- Detailed debug logging for multi-repository operations
- Comprehensive documentation for multi-repository setup
- Webhook configuration options (enabled, rate limit, timeout)
- Enhanced validation for repository URLs, paths, and sync intervals
- Improved signal handling for graceful shutdown
- Enhanced HTTP server with better middleware support
- Route documentation and discovery endpoint (/)

### Changed

- Refactored main application architecture to support multiple repositories
- Updated configuration system to handle multiple repository definitions
- Enhanced Git synchronization logic for concurrent repository operations
- Improved error messages with specific repository context
- Updated documentation to reflect multi-repository capabilities
- Enhanced logging with repository-specific information
- Refactored HTTP server implementation for better modularity and testability
- Improved configuration validation with detailed error messages
- Updated metrics collection for multi-repository monitoring
- Enhanced command-line flag handling for repository-specific options

### Fixed

- Configuration validation for repository URLs and local paths
- Error handling in multi-repository synchronization
- Graceful shutdown handling for multiple repositories
- Resource cleanup for concurrent repository operations

### Security

- Enhanced credential handling for multiple repositories
- Improved validation of repository URLs to prevent SSRF
- Added webhook rate limiting to prevent abuse
- Enhanced configuration validation to prevent misconfigurations

## [v1.0.0] - 2024-07-01

### Added

- Initial release of `git-sync` service.
- Basic functionality for synchronizing a local directory with a remote Git repository.
- Support for synchronization intervals.
- HTTP server with basic authentication and custom token authentication.
- Multi-platform build support for Linux, Windows, and macOS.
- Improved logging messages and translated them to English.

\[Unreleased\]: https://github.com/xT10r/git-sync/compare/v2.0.0...HEAD
\[v2.0.0\]: https://github.com/xT10r/git-sync/compare/v1.0.0...v2.0.0

\[Unreleased\]: https://github.com/xT10r/git-sync/compare/v2.0.0...HEAD
\[v2.0.0\]: https://github.com/xT10r/git-sync/compare/v1.0.0...v2.0.0
