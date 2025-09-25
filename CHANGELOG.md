# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Debug logging support with `--debug` flag, `GITSYNC_DEBUG` environment variable, and `debug` config file option
- Global logger functions (`logger.Info()`, `logger.Debug()`, `logger.Warning()`, `logger.Error()`, `logger.Fatal()`)
- Debug mode control functions (`logger.SetDebug()`, `logger.IsDebug()`)
- Enhanced documentation for debug logging feature
- Configuration file support using Viper
- Command to generate sample configuration files (`gen-config`)
- Command to display configuration help (`config-help`)
- Enhanced documentation for configuration options
- Unified configuration system that integrates flags, environment variables, and config files
- Comprehensive test suite with significant improvements to code coverage:
  - internal/flags package: 37.6% → 95.0% coverage
  - internal/config package: 0% → 83.7% coverage
  - internal/metrics package: 0% → 100% coverage
  - git package: 25.6% → 35.1% coverage

### Changed
- Improved configuration precedence (flags > environment variables > config file > defaults)
- Removed internal/config/config.go from .gitignore as it's now actively used

## [v1.0.0] - 2024-07-01
### Added
- Initial release of `git-sync` service.
- Basic functionality for synchronizing a local directory with a remote Git repository.
- Support for synchronization intervals.
- HTTP server with basic authentication and custom token authentication.
- Multi-platform build support for Linux, Windows, and macOS.
- Improved logging messages and translated them to English.