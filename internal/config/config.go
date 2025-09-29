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

package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git-sync/logger"

	"github.com/spf13/viper"
)

// Configuration key constants
const (
	GitlabRepoAuthKey         = "gitlab.repoAuth.tokenFile"
	HTTPServerAuthKey         = "http_server.auth.tokenFile"
	SyncIntervalKey           = "sync.interval"
	HTTPServerAddrKey         = "http_server.addr"
	SyncLocalPathKey          = "sync.local_path"
	GitlabRepoURLKey          = "gitlab.repoUrl"
	GitlabRepoBranchKey       = "gitlab.repoBranch"
	GitlabRepoAuthUserKey     = "gitlab.repoAuth.user"
	GitlabRepoAuthTokenKey    = "gitlab.repoAuth.token"
	HTTPServerAuthUsernameKey = "http_server.auth.username"
	HTTPServerAuthPasswordKey = "http_server.auth.password"
	ConfigFileKey             = "config.file"
	DebugKey                  = "debug"
	WebhookEnabledKey         = "webhook.enabled"
	WebhookRateLimitKey       = "webhook.rate_limit"
	WebhookTimeoutKey         = "webhook.timeout"
)

// Flag name constants
const (
	SyncIntervalFlagName      = "sync-interval"
	RepoURLFlagName           = "repo-url"
	RepoBranchFlagName        = "repo-branch"
	RepoUserFlagName          = "repo-user"
	RepoTokenFlagName         = "repo-token"
	RepoTokenFileFlagName     = "repo-token-file"
	HTTPAuthTokenFileFlagName = "http-auth-token-file"
	LocalPathFlagName         = "local-path"
	HTTPServerAddrFlagName    = "http-server-addr"
	HTTPAuthUsernameFlagName  = "http-auth-username"
	HTTPAuthPasswordFlagName  = "http-auth-password"
	HTTPAuthTokenFlagName     = "http-auth-token"
	ConfigFileFlagName        = "config-file"
	GenConfigFlagName         = "gen-config"
	ConfigHelpFlagName        = "config-help"
	DebugFlagName             = "debug"
	WebhookEnabledFlagName    = "webhook-enabled"
	WebhookRateLimitFlagName  = "webhook-rate-limit"
	WebhookTimeoutFlagName    = "webhook-timeout"
)

// Default value constants
const (
	DefaultHTTPServerAddr   = "0.0.0.0:8080"
	DefaultWebhookEnabled   = true
	DefaultWebhookRateLimit = 10
	DefaultWebhookTimeout   = 30
)

// Constants for flag types to avoid goconst warnings
const (
	FlagTypeDuration = "duration"
	FlagTypeBool     = "bool"
	FlagTypeInt      = "int"
	FlagTypeString   = "string"
)

// Config represents the configuration structure
type Config struct {
	Debug        bool                        `mapstructure:"debug"`
	Defaults     DefaultsConfig              `mapstructure:"defaults"`
	HttpServer   HttpServerConfig            `mapstructure:"http_server"`
	Repositories map[string]RepositoryConfig `mapstructure:"repositories"`
	Webhook      WebhookConfig               `mapstructure:"webhook"`
}

// DefaultsConfig represents default configuration values
type DefaultsConfig struct {
	Gitlab GitlabConfig `mapstructure:"gitlab"`
	Sync   SyncConfig   `mapstructure:"sync"`
}

// RepositoryConfig represents configuration for a single repository
type RepositoryConfig struct {
	Gitlab GitlabConfig `mapstructure:"gitlab"`
	Sync   SyncConfig   `mapstructure:"sync"`
}

// GitlabConfig represents GitLab configuration
type GitlabConfig struct {
	RepoURL    string         `mapstructure:"repourl"`
	RepoBranch string         `mapstructure:"repobranch"`
	RepoAuth   RepoAuthConfig `mapstructure:"repoauth"`
}

// RepoAuthConfig represents repository authentication configuration
type RepoAuthConfig struct {
	User      string `mapstructure:"user"`
	Token     string `mapstructure:"token"`
	TokenFile string `mapstructure:"token_file"`
}

// SyncConfig represents synchronization configuration
type SyncConfig struct {
	LocalPath string `mapstructure:"local_path"`
	Interval  int    `mapstructure:"interval"`
}

// HttpServerConfig represents HTTP server configuration
type HttpServerConfig struct {
	Addr string     `mapstructure:"addr"`
	Auth AuthConfig `mapstructure:"auth"`
}

// AuthConfig represents authentication configuration for HTTP server
type AuthConfig struct {
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Token     string `mapstructure:"token"`
	TokenFile string `mapstructure:"token_file"`
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	Enabled   bool `mapstructure:"enabled"`
	RateLimit int  `mapstructure:"rate_limit"`
	Timeout   int  `mapstructure:"timeout"`
}

// FlagInfo contains information about a command-line flag
type FlagInfo struct {
	Name         string
	ViperKey     string
	EnvVar       string
	Description  string
	DefaultValue interface{}
	Type         string
}

// AllFlags contains information about all flags
var AllFlags = []FlagInfo{
	{
		Name:         RepoURLFlagName,
		ViperKey:     GitlabRepoURLKey,
		EnvVar:       "GITSYNC_REPOSITORY_URL",
		Description:  "URL of the remote repository",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Name:         RepoBranchFlagName,
		ViperKey:     GitlabRepoBranchKey,
		EnvVar:       "GITSYNC_REPOSITORY_BRANCH",
		Description:  "Branch of the remote repository",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Name:         RepoUserFlagName,
		ViperKey:     GitlabRepoAuthUserKey,
		EnvVar:       "GITSYNC_REPOSITORY_USER",
		Description:  "Username for repository authentication",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Name:         RepoTokenFlagName,
		ViperKey:     GitlabRepoAuthTokenKey,
		EnvVar:       "GITSYNC_REPOSITORY_TOKEN",
		Description:  "Token for repository authentication",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Name:         RepoTokenFileFlagName,
		ViperKey:     GitlabRepoAuthKey,
		EnvVar:       "GITSYNC_REPOSITORY_TOKEN_FILE",
		Description:  "Path to file containing repository authentication token",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Name:         LocalPathFlagName,
		ViperKey:     SyncLocalPathKey,
		EnvVar:       "GITSYNC_LOCAL_PATH",
		Description:  "Path to the local repository",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Name:         SyncIntervalFlagName,
		ViperKey:     SyncIntervalKey,
		EnvVar:       "GITSYNC_INTERVAL",
		Description:  "Interval for repository synchronization (in seconds)",
		DefaultValue: 30,
		Type:         "duration",
	},
	{
		Name:         HTTPServerAddrFlagName,
		ViperKey:     HTTPServerAddrKey,
		EnvVar:       "GITSYNC_HTTP_SERVER_ADDR",
		Description:  "Address and port of the HTTP server",
		DefaultValue: DefaultHTTPServerAddr,
		Type:         "string",
	},
	{
		Name:         HTTPAuthUsernameFlagName,
		ViperKey:     HTTPServerAuthUsernameKey,
		EnvVar:       "GITSYNC_HTTP_AUTH_USERNAME",
		Description:  "Username for HTTP server authentication",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Name:         HTTPAuthPasswordFlagName,
		ViperKey:     HTTPServerAuthPasswordKey,
		EnvVar:       "GITSYNC_HTTP_AUTH_PASSWORD",
		Description:  "Password for HTTP server authentication",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Name:         HTTPAuthTokenFlagName,
		ViperKey:     "http_server.auth.token",
		EnvVar:       "GITSYNC_HTTP_AUTH_TOKEN",
		Description:  "Token for HTTP server authentication",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Name:         HTTPAuthTokenFileFlagName,
		ViperKey:     HTTPServerAuthKey,
		EnvVar:       "GITSYNC_HTTP_AUTH_TOKEN_FILE",
		Description:  "Path to file containing HTTP server authentication token",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Name:         ConfigFileFlagName,
		ViperKey:     ConfigFileKey,
		EnvVar:       "GITSYNC_CONFIG_FILE",
		Description:  "Path to the configuration file",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Name:         DebugFlagName,
		ViperKey:     DebugKey,
		EnvVar:       "GITSYNC_DEBUG",
		Description:  "Enable debug logging",
		DefaultValue: false,
		Type:         "bool",
	},
	{
		Name:         WebhookEnabledFlagName,
		ViperKey:     WebhookEnabledKey,
		EnvVar:       "GITSYNC_WEBHOOK_ENABLED",
		Description:  "Enable/disable webhook endpoint",
		DefaultValue: DefaultWebhookEnabled,
		Type:         "bool",
	},
	{
		Name:         WebhookRateLimitFlagName,
		ViperKey:     WebhookRateLimitKey,
		EnvVar:       "GITSYNC_WEBHOOK_RATE_LIMIT",
		Description:  "Max requests per minute per IP",
		DefaultValue: DefaultWebhookRateLimit,
		Type:         "int",
	},
	{
		Name:         WebhookTimeoutFlagName,
		ViperKey:     WebhookTimeoutKey,
		EnvVar:       "GITSYNC_WEBHOOK_TIMEOUT",
		Description:  "Timeout for sync operation (secs)",
		DefaultValue: DefaultWebhookTimeout,
		Type:         "int",
	},
}

// SpecialCommandFlags contains information about special command flags
var SpecialCommandFlags = []FlagInfo{
	{
		Name:         GenConfigFlagName,
		Description:  "Generate configuration file",
		DefaultValue: false,
		Type:         "bool",
	},
	{
		Name:         ConfigHelpFlagName,
		Description:  "Show help for configuration",
		DefaultValue: false,
		Type:         "bool",
	},
}

// ReadConfig reads configuration from file, environment variables and flags
func ReadConfig(fs *flag.FlagSet, configPath string) (*Config, error) {
	logger.Debug("Reading configuration")

	// Set default values
	viper.SetDefault(SyncIntervalKey, 30)
	viper.SetDefault(HTTPServerAddrKey, DefaultHTTPServerAddr)
	viper.SetDefault(DebugKey, false) // Set default for debug mode
	viper.SetDefault(WebhookEnabledKey, DefaultWebhookEnabled)
	viper.SetDefault(WebhookRateLimitKey, DefaultWebhookRateLimit)
	viper.SetDefault(WebhookTimeoutKey, DefaultWebhookTimeout)

	// Set config file name and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Check if config file path is specified via flag or parameter
	configFilePath := configPath
	if configFilePath == "" && fs != nil {
		// Try to get config file path from flag
		if f := fs.Lookup(ConfigFileFlagName); f != nil {
			if flagValue := f.Value.String(); flagValue != "" {
				configFilePath = flagValue
			}
		}
	}

	// Set config path
	if configFilePath != "" {
		// Extract directory from file path
		dir := filepath.Dir(configFilePath)
		viper.AddConfigPath(dir)
		// Also set the config file name directly
		viper.SetConfigFile(configFilePath)
		logger.Debug("Using config file: %s", configFilePath)
	}
	// Add default config paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.gitsync")

	// Bind environment variables
	bindEnvs()

	// Set values from flags if provided
	if fs != nil {
		setValuesFromFlags(fs)
	}

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Debug("Config file not found, using defaults and environment variables")
		} else {
			logger.Debug("Error reading config file: %v", err)
		}
	} else {
		logger.Debug("Using config file: %s", viper.ConfigFileUsed())
	}

	// Read values into config struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %v", err)
	}

	logger.Debug("Configuration read successfully, found %d repositories", len(config.Repositories))

	// Apply defaults to each repository
	for repoName, repoConfig := range config.Repositories {
		logger.Debug("Applying defaults for repository: %s", repoName)

		// Apply GitLab defaults
		if repoConfig.Gitlab.RepoBranch == "" && config.Defaults.Gitlab.RepoBranch != "" {
			repoConfig.Gitlab.RepoBranch = config.Defaults.Gitlab.RepoBranch
			logger.Debug("Applied default branch %s for repository %s", config.Defaults.Gitlab.RepoBranch, repoName)
		}

		// Apply GitLab authentication defaults
		// If repository doesn't have user set, use default user
		if repoConfig.Gitlab.RepoAuth.User == "" && config.Defaults.Gitlab.RepoAuth.User != "" {
			repoConfig.Gitlab.RepoAuth.User = config.Defaults.Gitlab.RepoAuth.User
			logger.Debug("Applied default user for repository %s", repoName)
		}

		// If repository doesn't have token set, use default token
		if repoConfig.Gitlab.RepoAuth.Token == "" && config.Defaults.Gitlab.RepoAuth.Token != "" {
			repoConfig.Gitlab.RepoAuth.Token = config.Defaults.Gitlab.RepoAuth.Token
			logger.Debug("Applied default token for repository %s", repoName)
		}

		// If repository doesn't have token_file set, use default token_file
		if repoConfig.Gitlab.RepoAuth.TokenFile == "" && config.Defaults.Gitlab.RepoAuth.TokenFile != "" {
			repoConfig.Gitlab.RepoAuth.TokenFile = config.Defaults.Gitlab.RepoAuth.TokenFile
			logger.Debug("Applied default token file for repository %s", repoName)
		}

		// Apply Sync defaults
		if repoConfig.Sync.Interval == 0 && config.Defaults.Sync.Interval != 0 {
			repoConfig.Sync.Interval = config.Defaults.Sync.Interval
			logger.Debug("Applied default sync interval %d for repository %s", config.Defaults.Sync.Interval, repoName)
		}

		// Update the repository config
		config.Repositories[repoName] = repoConfig
	}

	// Read token files if specified for each repository
	for repoName, repoConfig := range config.Repositories {
		if repoConfig.Gitlab.RepoAuth.TokenFile != "" {
			logger.Debug("Reading token file for repository %s: %s", repoName, repoConfig.Gitlab.RepoAuth.TokenFile)
			if token, err := readTokenFromFile(repoConfig.Gitlab.RepoAuth.TokenFile); err == nil {
				// Update the token in the repository config
				config.Repositories[repoName] = RepositoryConfig{
					Gitlab: GitlabConfig{
						RepoURL:    repoConfig.Gitlab.RepoURL,
						RepoBranch: repoConfig.Gitlab.RepoBranch,
						RepoAuth: RepoAuthConfig{
							User:      repoConfig.Gitlab.RepoAuth.User,
							Token:     token,
							TokenFile: repoConfig.Gitlab.RepoAuth.TokenFile,
						},
					},
					Sync: repoConfig.Sync,
				}
				logger.Debug("Successfully read token from file for repository %s", repoName)
			} else {
				logger.Debug("Error reading repository token file for %s: %v", repoName, err)
			}
		}
	}

	if config.HttpServer.Auth.TokenFile != "" {
		logger.Debug("Reading HTTP auth token file: %s", config.HttpServer.Auth.TokenFile)
		if token, err := readTokenFromFile(config.HttpServer.Auth.TokenFile); err == nil {
			// Update the token in the HTTP server config
			config.HttpServer = HttpServerConfig{
				Addr: config.HttpServer.Addr,
				Auth: AuthConfig{
					Username:  config.HttpServer.Auth.Username,
					Password:  config.HttpServer.Auth.Password,
					Token:     token,
					TokenFile: config.HttpServer.Auth.TokenFile,
				},
			}
			logger.Debug("Successfully read HTTP auth token from file")
		} else {
			logger.Debug("Error reading HTTP auth token file: %v", err)
		}
	}

	return &config, nil
}

// bindEnvs binds environment variables to viper keys
func bindEnvs() {
	for _, flagInfo := range AllFlags {
		if flagInfo.ViperKey != "" {
			if err := viper.BindEnv(flagInfo.ViperKey, flagInfo.EnvVar); err != nil {
				logger.Debug("Error binding environment variable %s to key %s: %v", flagInfo.EnvVar, flagInfo.ViperKey, err)
			}
		}
	}
}

// setValuesFromFlags sets Viper values from command-line flags
func setValuesFromFlags(fs *flag.FlagSet) {
	// Create a map of explicitly set flags
	explicitlySetFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		explicitlySetFlags[f.Name] = true
	})

	// Now process all flags
	fs.VisitAll(func(f *flag.Flag) {
		for _, flagInfo := range AllFlags {
			if f.Name == flagInfo.Name && flagInfo.ViperKey != "" {
				// Only set the value if the flag was explicitly provided, not just using default value
				if explicitlySetFlags[f.Name] {
					// Special handling for duration flags
					if f.Name == SyncIntervalFlagName {
						// Convert duration to seconds
						if dur, err := time.ParseDuration(f.Value.String()); err == nil {
							viper.Set(flagInfo.ViperKey, int(dur.Seconds()))
						}
					} else {
						viper.Set(flagInfo.ViperKey, f.Value.String())
					}
				}
				break // Found the flag, no need to continue
			}
		}
	})
}

// readTokenFromFile reads a token from a file and trims whitespace
func readTokenFromFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// Trim whitespace and newlines
	return strings.TrimSpace(string(content)), nil
}

// GenerateSampleConfig generates a sample configuration file
func GenerateSampleConfig(filePath string) error {
	logger.Debug("Generating sample configuration file at: %s", filePath)

	// Set sample values for defaults
	viper.Set("debug", false)

	// Set defaults
	viper.Set("defaults.gitlab.repoauth.user", "gitlab-user")
	viper.Set("defaults.gitlab.repoauth.token", "your-default-gitlab-token")
	viper.Set("defaults.gitlab.repoauth.token_file", "") // Default token file path (optional)
	viper.Set("defaults.gitlab.repobranch", "main")
	viper.Set("defaults.sync.interval", 30)

	// Set HTTP server config
	viper.Set("http_server.addr", DefaultHTTPServerAddr)
	viper.Set("http_server.auth.username", "admin")
	viper.Set("http_server.auth.password", "password")

	// Set webhook config
	viper.Set("webhook.enabled", DefaultWebhookEnabled)
	viper.Set("webhook.rate_limit", DefaultWebhookRateLimit)
	viper.Set("webhook.timeout", DefaultWebhookTimeout)

	// Set repositories
	// Repository that uses default credentials
	viper.Set("repositories.main-repo.gitlab.repourl", "https://gitlab.example.com/username/repository1.git")
	// Note: No repoauth specified - will use defaults
	viper.Set("repositories.main-repo.sync.local_path", "./repo1")

	// Repository with its own credentials
	viper.Set("repositories.secondary-repo.gitlab.repoauth.token", "your-gitlab-token-2")
	viper.Set("repositories.secondary-repo.gitlab.repoauth.user", "secondary-user")
	viper.Set("repositories.secondary-repo.gitlab.repobranch", "develop")
	viper.Set("repositories.secondary-repo.gitlab.repourl", "https://gitlab.example.com/username/repository2.git")
	viper.Set("repositories.secondary-repo.sync.interval", 60)
	viper.Set("repositories.secondary-repo.sync.local_path", "./repo2")

	// Repository that inherits token from defaults but uses its own user
	viper.Set("repositories.external-repo.gitlab.repoauth.user", "different-user")
	viper.Set("repositories.external-repo.gitlab.repourl", "https://gitlab.example.com/anotheruser/repository.git")
	// Note: No token specified - will use default token
	viper.Set("repositories.external-repo.sync.local_path", "./external-repo")

	// Repository with all its own credentials
	viper.Set("repositories.custom-repo.gitlab.repoauth.token", "your-gitlab-token-4")
	viper.Set("repositories.custom-repo.gitlab.repoauth.user", "custom-user")
	viper.Set("repositories.custom-repo.gitlab.repobranch", "feature-branch")
	viper.Set("repositories.custom-repo.gitlab.repourl", "https://gitlab.example.com/custom/repo.git")
	viper.Set("repositories.custom-repo.sync.interval", 300)
	viper.Set("repositories.custom-repo.sync.local_path", "./custom-repo")

	// Write config to file
	viper.SetConfigFile(filePath)
	err := viper.WriteConfig()
	if err != nil {
		logger.Debug("Error writing sample configuration file: %v", err)
		return err
	}

	logger.Debug("Sample configuration file generated successfully")
	return nil
}

// PrintUnifiedHelp prints unified help information
func PrintUnifiedHelp() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nCommand-line Flags:\n")
	fmt.Fprintf(os.Stderr, "==================\n")

	// Print standard flags
	for _, flagInfo := range AllFlags {
		switch flagInfo.Type {
		case FlagTypeDuration:
			fmt.Fprintf(os.Stderr, "  -%s duration\n", flagInfo.Name)
		case FlagTypeBool:
			fmt.Fprintf(os.Stderr, "  -%s\n", flagInfo.Name)
		case FlagTypeInt:
			fmt.Fprintf(os.Stderr, "  -%s int\n", flagInfo.Name)
		default:
			fmt.Fprintf(os.Stderr, "  -%s string\n", flagInfo.Name)
		}
		fmt.Fprintf(os.Stderr, "        %s (%s)\n", flagInfo.Description, flagInfo.EnvVar)
	}

	// Add special command flags
	fmt.Fprintf(os.Stderr, "\nSpecial Commands:\n")
	fmt.Fprintf(os.Stderr, "================\n")
	for _, flagInfo := range SpecialCommandFlags {
		fmt.Fprintf(os.Stderr, "  -%s\n", flagInfo.Name)
		fmt.Fprintf(os.Stderr, "        %s\n", flagInfo.Description)
	}

	// Add detailed configuration help
	fmt.Fprintf(os.Stderr, "\nConfiguration Options (via config file, environment variables, or flags):\n")
	fmt.Fprintf(os.Stderr, "=========================================================================\n")
	PrintConfigHelpInFormat()
}

// GetConfigDescription returns English descriptions for configuration parameters
func GetConfigDescription() map[string]string {
	descriptions := make(map[string]string)

	// Add descriptions for all configuration options
	for _, flagInfo := range AllFlags {
		if flagInfo.ViperKey != "" {
			descriptions[flagInfo.ViperKey] = flagInfo.Description
		}
	}

	// Add descriptions for tokenFile which are not in flags but shown in help
	descriptions[GitlabRepoAuthKey] = "Path to file containing repository authentication token"
	descriptions[HTTPServerAuthKey] = "Path to file containing HTTP server authentication token"

	return descriptions
}

// PrintConfigHelp prints help information for all configuration options
func PrintConfigHelp() {
	descriptions := GetConfigDescription()
	fmt.Fprintf(os.Stderr, "Configuration Options:\n")
	fmt.Fprintf(os.Stderr, "=====================\n")

	// Group by section
	var gitlabOptions []string
	var syncOptions []string
	var httpOptions []string
	var webhookOptions []string

	// Add options from AllFlags
	for _, flagInfo := range AllFlags {
		if flagInfo.ViperKey != "" {
			switch {
			case strings.HasPrefix(flagInfo.ViperKey, "gitlab."):
				gitlabOptions = append(gitlabOptions, flagInfo.ViperKey)
			case strings.HasPrefix(flagInfo.ViperKey, "sync."):
				syncOptions = append(syncOptions, flagInfo.ViperKey)
			case strings.HasPrefix(flagInfo.ViperKey, "http_server."):
				httpOptions = append(httpOptions, flagInfo.ViperKey)
			case strings.HasPrefix(flagInfo.ViperKey, "webhook."):
				webhookOptions = append(webhookOptions, flagInfo.ViperKey)
			}
		}
	}

	// Add tokenFile options which are not in flags but shown in help
	gitlabOptions = append(gitlabOptions, GitlabRepoAuthKey)
	httpOptions = append(httpOptions, HTTPServerAuthKey)

	fmt.Fprintf(os.Stderr, "\nGitLab Settings:\n")
	for _, opt := range gitlabOptions {
		if desc, exists := descriptions[opt]; exists {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", opt, desc)
			showEnvAndFlag(opt)
		}
	}

	fmt.Fprintf(os.Stderr, "\nSync Settings:\n")
	for _, opt := range syncOptions {
		if desc, exists := descriptions[opt]; exists {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", opt, desc)
			showEnvAndFlag(opt)
		}
	}

	fmt.Fprintf(os.Stderr, "\nHTTP Server Settings:\n")
	for _, opt := range httpOptions {
		if desc, exists := descriptions[opt]; exists {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", opt, desc)
			showEnvAndFlag(opt)
		}
	}

	fmt.Fprintf(os.Stderr, "\nWebhook Settings:\n")
	for _, opt := range webhookOptions {
		if desc, exists := descriptions[opt]; exists {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", opt, desc)
			showEnvAndFlag(opt)
		}
	}
}

// PrintConfigHelpInFormat prints help information for all configuration options in a format suitable for command-line help
func PrintConfigHelpInFormat() {
	descriptions := GetConfigDescription()

	// Group by section
	var gitlabOptions []string
	var syncOptions []string
	var httpOptions []string
	var webhookOptions []string

	// Add options from AllFlags
	for _, flagInfo := range AllFlags {
		if flagInfo.ViperKey != "" {
			switch {
			case strings.HasPrefix(flagInfo.ViperKey, "gitlab."):
				gitlabOptions = append(gitlabOptions, flagInfo.ViperKey)
			case strings.HasPrefix(flagInfo.ViperKey, "sync."):
				syncOptions = append(syncOptions, flagInfo.ViperKey)
			case strings.HasPrefix(flagInfo.ViperKey, "http_server."):
				httpOptions = append(httpOptions, flagInfo.ViperKey)
			case strings.HasPrefix(flagInfo.ViperKey, "webhook."):
				webhookOptions = append(webhookOptions, flagInfo.ViperKey)
			}
		}
	}

	// Add tokenFile options which are not in flags but shown in help
	gitlabOptions = append(gitlabOptions, GitlabRepoAuthKey)
	httpOptions = append(httpOptions, HTTPServerAuthKey)

	fmt.Fprintf(os.Stderr, "\nGitLab Settings:\n")
	for _, opt := range gitlabOptions {
		if desc, exists := descriptions[opt]; exists {
			// Find the corresponding flag name
			flagName := ""
			for _, flagInfo := range AllFlags {
				if flagInfo.ViperKey == opt {
					flagName = flagInfo.Name
					break
				}
			}

			// Handle special cases for tokenFile
			if flagName == "" {
				switch opt {
				case GitlabRepoAuthKey:
					flagName = RepoTokenFileFlagName
				case HTTPServerAuthKey:
					flagName = HTTPAuthTokenFileFlagName
				}
			}

			if flagName != "" {
				fmt.Fprintf(os.Stderr, "  -%s string\n", flagName)
				fmt.Fprintf(os.Stderr, "        %s (env: %s)\n", desc, getEnvVarForConfigKey(opt))
			}
		}
	}

	fmt.Fprintf(os.Stderr, "\nSync Settings:\n")
	for _, opt := range syncOptions {
		if desc, exists := descriptions[opt]; exists {
			// Find the corresponding flag name
			flagName := ""
			for _, flagInfo := range AllFlags {
				if flagInfo.ViperKey == opt {
					flagName = flagInfo.Name
					break
				}
			}

			if flagName != "" {
				if flagName == SyncIntervalFlagName {
					fmt.Fprintf(os.Stderr, "  -%s duration\n", flagName)
					fmt.Fprintf(os.Stderr, "        %s (env: %s)\n", desc, getEnvVarForConfigKey(opt))
				} else {
					fmt.Fprintf(os.Stderr, "  -%s string\n", flagName)
					fmt.Fprintf(os.Stderr, "        %s (env: %s)\n", desc, getEnvVarForConfigKey(opt))
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "\nHTTP Server Settings:\n")
	for _, opt := range httpOptions {
		if desc, exists := descriptions[opt]; exists {
			// Find the corresponding flag name
			flagName := ""
			for _, flagInfo := range AllFlags {
				if flagInfo.ViperKey == opt {
					flagName = flagInfo.Name
					break
				}
			}

			// Handle special cases for tokenFile
			if flagName == "" {
				switch opt {
				case GitlabRepoAuthKey:
					flagName = RepoTokenFileFlagName
				case HTTPServerAuthKey:
					flagName = HTTPAuthTokenFileFlagName
				}
			}

			if flagName != "" {
				fmt.Fprintf(os.Stderr, "  -%s string\n", flagName)
				fmt.Fprintf(os.Stderr, "        %s (env: %s)\n", desc, getEnvVarForConfigKey(opt))
			}
		}
	}

	fmt.Fprintf(os.Stderr, "\nWebhook Settings:\n")
	for _, opt := range webhookOptions {
		if desc, exists := descriptions[opt]; exists {
			// Find the corresponding flag name
			flagName := ""
			for _, flagInfoItem := range AllFlags {
				if flagInfoItem.ViperKey == opt {
					flagName = flagInfoItem.Name
					break
				}
			}

			if flagName != "" {
				// Find the flag info to get its type
				var flagInfo *FlagInfo
				for _, fi := range AllFlags {
					if fi.ViperKey == opt {
						flagInfo = &fi
						break
					}
				}

				if flagInfo != nil {
					switch flagInfo.Type {
					case "int":
						fmt.Fprintf(os.Stderr, "  -%s int\n", flagName)
						fmt.Fprintf(os.Stderr, "        %s (env: %s)\n", desc, getEnvVarForConfigKey(opt))
					case "bool":
						fmt.Fprintf(os.Stderr, "  -%s\n", flagName)
						fmt.Fprintf(os.Stderr, "        %s (env: %s)\n", desc, getEnvVarForConfigKey(opt))
					default:
						fmt.Fprintf(os.Stderr, "  -%s string\n", flagName)
						fmt.Fprintf(os.Stderr, "        %s (env: %s)\n", desc, getEnvVarForConfigKey(opt))
					}
				}
			}
		}
	}
}

func showEnvAndFlag(configKey string) {
	// Find the flag info for this config key
	for _, flagInfo := range AllFlags {
		if flagInfo.ViperKey == configKey {
			fmt.Fprintf(os.Stderr, "    Environment variable: %s\n", flagInfo.EnvVar)
			fmt.Fprintf(os.Stderr, "    Command-line flag: --%s\n", flagInfo.Name)
			return
		}
	}

	// Handle special cases for tokenFile which are not in flags but shown in help
	switch configKey {
	case GitlabRepoAuthKey:
		fmt.Fprintf(os.Stderr, "    Environment variable: %s\n", "GITSYNC_REPOSITORY_TOKEN_FILE")
		fmt.Fprintf(os.Stderr, "    Command-line flag: --%s\n", RepoTokenFileFlagName)
	case HTTPServerAuthKey:
		fmt.Fprintf(os.Stderr, "    Environment variable: %s\n", "GITSYNC_HTTP_AUTH_TOKEN_FILE")
		fmt.Fprintf(os.Stderr, "    Command-line flag: --%s\n", HTTPAuthTokenFileFlagName)
	}
}

func getEnvVarForConfigKey(configKey string) string {
	// Find the flag info for this config key
	for _, flagInfo := range AllFlags {
		if flagInfo.ViperKey == configKey {
			return flagInfo.EnvVar
		}
	}

	// Handle special cases for tokenFile which are not in flags but shown in help
	switch configKey {
	case GitlabRepoAuthKey:
		return "GITSYNC_REPOSITORY_TOKEN_FILE"
	case HTTPServerAuthKey:
		return "GITSYNC_HTTP_AUTH_TOKEN_FILE"
	}

	return ""
}

func getFlagNameForConfigKey(configKey string) string {
	// Find the flag info for this config key
	for _, flagInfo := range AllFlags {
		if flagInfo.ViperKey == configKey {
			return flagInfo.Name
		}
	}

	// Handle special cases for tokenFile which are not in flags but shown in help
	switch configKey {
	case GitlabRepoAuthKey:
		return RepoTokenFileFlagName
	case HTTPServerAuthKey:
		return HTTPAuthTokenFileFlagName
	}

	return ""
}
