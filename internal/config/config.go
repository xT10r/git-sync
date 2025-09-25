// Copyright 2024 Aleksey Dobshikov
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

package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
	DebugKey                  = "debug" // New debug configuration key
)

// Flag name constants
const (
	SyncIntervalFlagName     = "sync-interval"
	RepoURLFlagName          = "repo-url"
	RepoBranchFlagName       = "repo-branch"
	RepoUserFlagName         = "repo-user"
	RepoTokenFlagName        = "repo-token"
	LocalPathFlagName        = "local-path"
	HTTPServerAddrFlagName   = "http-server-addr"
	HTTPAuthUsernameFlagName = "http-auth-username"
	HTTPAuthPasswordFlagName = "http-auth-password"
	HTTPAuthTokenFlagName    = "http-auth-token"
	ConfigFileFlagName       = "config-file"
	GenConfigFlagName        = "gen-config"
	ConfigHelpFlagName       = "config-help"
	DebugFlagName            = "debug" // New debug flag name
)

// Default value constants
const (
	DefaultHTTPServerAddr = "0.0.0.0:8080"
)

// Config represents the configuration structure
type Config struct {
	Gitlab struct {
		RepoURL    string `mapstructure:"repoUrl"`
		RepoBranch string `mapstructure:"repoBranch"`
		RepoAuth   struct {
			User      string `mapstructure:"user"`
			Token     string `mapstructure:"token"`
			TokenFile string `mapstructure:"token_file"`
		} `mapstructure:"repoAuth"`
	} `mapstructure:"gitlab"`
	Sync struct {
		LocalPath string `mapstructure:"local_path"`
		Interval  int    `mapstructure:"interval"`
	} `mapstructure:"sync"`
	HttpServer struct {
		Addr string `mapstructure:"addr"`
		Auth struct {
			Username  string `mapstructure:"username"`
			Password  string `mapstructure:"password"`
			Token     string `mapstructure:"token"`
			TokenFile string `mapstructure:"token_file"`
		} `mapstructure:"auth"`
	} `mapstructure:"http_server"`
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
	// Set default values
	viper.SetDefault(SyncIntervalKey, 30)
	viper.SetDefault(HTTPServerAddrKey, DefaultHTTPServerAddr)
	viper.SetDefault(DebugKey, false) // Set default for debug mode

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
		viper.AddConfigPath(configFilePath)
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
			log.Println("Config file not found, using defaults and environment variables")
		} else {
			log.Printf("Error reading config file: %v\n", err)
		}
	} else {
		log.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	// Read values into config struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %v", err)
	}

	return &config, nil
}

// bindEnvs binds environment variables to viper keys
func bindEnvs() {
	for _, flagInfo := range AllFlags {
		if flagInfo.ViperKey != "" {
			viper.BindEnv(flagInfo.ViperKey, flagInfo.EnvVar)
		}
	}
}

// setValuesFromFlags sets Viper values from command-line flags
func setValuesFromFlags(fs *flag.FlagSet) {
	fs.VisitAll(func(f *flag.Flag) {
		for _, flagInfo := range AllFlags {
			if f.Name == flagInfo.Name && flagInfo.ViperKey != "" {
				if f.Value.String() != "" {
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

// GenerateSampleConfig generates a sample configuration file
func GenerateSampleConfig(filePath string) error {
	// Set sample values using the unified flag information
	for _, flagInfo := range AllFlags {
		if flagInfo.ViperKey != "" {
			switch flagInfo.ViperKey {
			case GitlabRepoURLKey:
				viper.Set(flagInfo.ViperKey, "https://gitlab.example.com/username/repository.git")
			case GitlabRepoBranchKey:
				viper.Set(flagInfo.ViperKey, "main")
			case GitlabRepoAuthUserKey:
				viper.Set(flagInfo.ViperKey, "gitlab-user")
			case GitlabRepoAuthTokenKey:
				viper.Set(flagInfo.ViperKey, "your-gitlab-token")
			case SyncLocalPathKey:
				viper.Set(flagInfo.ViperKey, "./repo")
			case SyncIntervalKey:
				viper.Set(flagInfo.ViperKey, 30)
			case HTTPServerAddrKey:
				viper.Set(flagInfo.ViperKey, DefaultHTTPServerAddr)
			case HTTPServerAuthUsernameKey:
				viper.Set(flagInfo.ViperKey, "admin")
			case HTTPServerAuthPasswordKey:
				viper.Set(flagInfo.ViperKey, "password")
			case DebugKey:
				viper.Set(flagInfo.ViperKey, false)
			}
		}
	}

	// Write config to file
	viper.SetConfigFile(filePath)
	return viper.WriteConfig()
}

// PrintUnifiedHelp prints unified help information
func PrintUnifiedHelp() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nCommand-line Flags:\n")
	fmt.Fprintf(os.Stderr, "==================\n")

	// Print standard flags
	for _, flagInfo := range AllFlags {
		if flagInfo.Type == "duration" {
			fmt.Fprintf(os.Stderr, "  -%s duration\n", flagInfo.Name)
		} else if flagInfo.Type == "bool" {
			fmt.Fprintf(os.Stderr, "  -%s\n", flagInfo.Name)
		} else {
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

	// Add options from AllFlags
	for _, flagInfo := range AllFlags {
		if flagInfo.ViperKey != "" {
			if strings.HasPrefix(flagInfo.ViperKey, "gitlab.") {
				gitlabOptions = append(gitlabOptions, flagInfo.ViperKey)
			} else if strings.HasPrefix(flagInfo.ViperKey, "sync.") {
				syncOptions = append(syncOptions, flagInfo.ViperKey)
			} else if strings.HasPrefix(flagInfo.ViperKey, "http_server.") {
				httpOptions = append(httpOptions, flagInfo.ViperKey)
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
}

// PrintConfigHelpInFormat prints help information for all configuration options in a format suitable for command-line help
func PrintConfigHelpInFormat() {
	descriptions := GetConfigDescription()

	// Group by section
	var gitlabOptions []string
	var syncOptions []string
	var httpOptions []string

	// Add options from AllFlags
	for _, flagInfo := range AllFlags {
		if flagInfo.ViperKey != "" {
			if strings.HasPrefix(flagInfo.ViperKey, "gitlab.") {
				gitlabOptions = append(gitlabOptions, flagInfo.ViperKey)
			} else if strings.HasPrefix(flagInfo.ViperKey, "sync.") {
				syncOptions = append(syncOptions, flagInfo.ViperKey)
			} else if strings.HasPrefix(flagInfo.ViperKey, "http_server.") {
				httpOptions = append(httpOptions, flagInfo.ViperKey)
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
					flagName = "repo-token-file"
				case HTTPServerAuthKey:
					flagName = "http-auth-token-file"
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
					flagName = "repo-token-file"
				case HTTPServerAuthKey:
					flagName = "http-auth-token-file"
				}
			}

			if flagName != "" {
				fmt.Fprintf(os.Stderr, "  -%s string\n", flagName)
				fmt.Fprintf(os.Stderr, "        %s (env: %s)\n", desc, getEnvVarForConfigKey(opt))
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
		fmt.Fprintf(os.Stderr, "    Command-line flag: --%s\n", "repo-token-file")
	case HTTPServerAuthKey:
		fmt.Fprintf(os.Stderr, "    Environment variable: %s\n", "GITSYNC_HTTP_AUTH_TOKEN_FILE")
		fmt.Fprintf(os.Stderr, "    Command-line flag: --%s\n", "http-auth-token-file")
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
		return "repo-token-file"
	case HTTPServerAuthKey:
		return "http-auth-token-file"
	}

	return ""
}
