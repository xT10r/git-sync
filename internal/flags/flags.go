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

package flags

import (
	"flag"
	"fmt"
	"git-sync/internal/config"
	"git-sync/logger"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// ConsoleFlags represents a set of command-line flags.
type ConsoleFlags struct {
	Gitsync *flag.FlagSet
}

// NewConsoleFlags creates a new set of command-line flags.
func NewConsoleFlags() *ConsoleFlags {
	fs := ParseFlags()
	return &ConsoleFlags{
		Gitsync: fs,
	}
}

// ParseFlags initializes flags using a flag set.
func ParseFlags() *flag.FlagSet {
	fs := flag.NewFlagSet("git-sync", flag.ContinueOnError)

	// Register all standard flags from the unified configuration
	for _, flagInfo := range config.AllFlags {
		switch flagInfo.Type {
		case "duration":
			fs.Duration(flagInfo.Name, getEnvDuration(flagInfo.EnvVar, 30*time.Second), fmt.Sprintf("%s (%s)", flagInfo.Description, flagInfo.EnvVar))
		case "bool":
			fs.Bool(flagInfo.Name, getEnvBool(flagInfo.EnvVar, false), fmt.Sprintf("%s (%s)", flagInfo.Description, flagInfo.EnvVar))
		default:
			fs.String(flagInfo.Name, getEnv(flagInfo.EnvVar, ""), fmt.Sprintf("%s (%s)", flagInfo.Description, flagInfo.EnvVar))
		}
	}

	// Register special command flags
	for _, flagInfo := range config.SpecialCommandFlags {
		fs.Bool(flagInfo.Name, false, flagInfo.Description)
	}

	// Parse flags
	err := fs.Parse(os.Args[1:])
	if err != nil {
		// If there's a parsing error, exit with a clear message
		if err != flag.ErrHelp {
			fmt.Fprintf(os.Stderr, "[ERROR] Invalid flag usage: %v\n", err)
			fmt.Fprintf(os.Stderr, "Use --help or --config-help for more information\n")
			os.Exit(1)
		}
		// For help, we return the flag set and let main handle it
	}

	return fs
}

// CheckRequiredFlags checks if all required flags are set
func (cf *ConsoleFlags) CheckRequiredFlags() error {
	requiredFlags := []string{
		config.RepoURLFlagName,
		config.RepoBranchFlagName,
		config.LocalPathFlagName,
	}

	var missingFlags []string
	for _, flagName := range requiredFlags {
		flag := cf.Gitsync.Lookup(flagName)
		if flag == nil || flag.Value.String() == "" {
			missingFlags = append(missingFlags, flagName)
		}
	}

	if len(missingFlags) > 0 {
		return fmt.Errorf("required flags are missing: %v. Use --help or --config-help for more information", strings.Join(missingFlags, ", "))
	}
	return nil
}

func (consoleFlags *ConsoleFlags) ValidateFlags() error {
	if err := validateFlags(consoleFlags.Gitsync); err != nil {
		return err
	}
	return nil
}

func validateFlags(fs *flag.FlagSet) error {
	// Repo URL
	if err := validateFlagURL(fs, config.RepoURLFlagName, "Repository URL"); err != nil {
		return err
	}

	// Local path
	if err := validateFlagLocalPath(fs, config.LocalPathFlagName, "Local Path"); err != nil {
		return err
	}

	// Repo Branch
	validateFlagOptional(fs, config.RepoBranchFlagName, "Repository Branch")

	// Repo user
	validateFlagOptional(fs, config.RepoUserFlagName, "Repository User")

	// Repo token
	validateFlagOptional(fs, config.RepoTokenFlagName, "Repository Token")

	// Repo token file
	validateFlagOptional(fs, config.RepoTokenFileFlagName, "Repository Token File")

	// Sync interval
	if err := validateFlagSyncInterval(fs, config.SyncIntervalFlagName, "Sync Interval"); err != nil {
		return err
	}

	// HTTP Server Addr
	if err := validateFlagsHttpServer(fs); err != nil {
		return err
	}

	// HTTP Auth Token File
	validateFlagOptional(fs, config.HTTPAuthTokenFileFlagName, "HTTP Auth Token File")

	return nil
}

func getFlagValue(fs *flag.FlagSet, flagName string) (string, bool) {
	if f := fs.Lookup(flagName); f != nil {
		value := f.Value.String()
		return value, true
	}
	return "", false
}

// getEnv returns the value of an environment variable or a default value if the variable is not set.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvDuration returns the value of an environment variable in time.Duration format or a default value if the variable is not set or has an incorrect format.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

// getEnvBool returns the value of an environment variable as a boolean or a default value if the variable is not set.
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// Convert string to boolean
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

func validateFlagURL(fs *flag.FlagSet, fn string, desc string) error {
	repoUrl, isExists := getFlagValue(fs, fn)

	// Check if flag exists
	if !isExists {
		return fmt.Errorf("%s is not set", desc)
	}

	// Check URL validity
	_, err := url.ParseRequestURI(repoUrl)
	if err != nil {
		return err
	}
	return nil
}

func validateFlagLocalPath(fs *flag.FlagSet, fn string, desc string) error {
	localPath, isExists := getFlagValue(fs, fn)

	if !isExists {
		return fmt.Errorf("%s is not set", desc)
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("specified local path does not exist: %s", localPath)
	}
	return nil
}

func validateFlagOptional(fs *flag.FlagSet, fn string, desc string) {
	if fv, isExists := getFlagValue(fs, fn); !isExists {
		logger.Warning("%s is not set\n", desc)
	} else if fv == "" {
		logger.Warning("%s is empty\n", desc)
	}
}

func validateFlagSyncInterval(fs *flag.FlagSet, fn string, desc string) error {
	fv, isExists := getFlagValue(fs, fn)
	if !isExists {
		return fmt.Errorf("%s is not set", desc)
	}

	if duration, err := time.ParseDuration(fv); err != nil {
		return fmt.Errorf("failed to convert sync interval string to duration")
	} else {
		if duration <= 0 {
			return fmt.Errorf("sync interval must be positive")
		}
	}
	return nil
}

func validateFlagsHttpServer(fs *flag.FlagSet) error {
	// HTTP Server Addr
	httpServerAddr, _ := getFlagValue(fs, config.HTTPServerAddrFlagName)

	if len(httpServerAddr) == 0 {
		return nil
	}

	// Split address into IP and port
	parts := strings.Split(httpServerAddr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("address must be in the format IP:PORT")
	}

	// Validate IP address
	ip := net.ParseIP(parts[0])
	if ip == nil {
		return fmt.Errorf("invalid HTTP server IP-address")
	}

	// Validate port
	port, err := strconv.Atoi(parts[1])
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port. Valid port range is [1-65535]")
	}

	// HTTP Server Auth username
	username, _ := getFlagValue(fs, config.HTTPAuthUsernameFlagName)
	password, _ := getFlagValue(fs, config.HTTPAuthPasswordFlagName)

	// HTTP Server Auth Bearer Token
	token, _ := getFlagValue(fs, config.HTTPAuthTokenFlagName)

	// HTTP Server Auth Token File
	tokenFile, _ := getFlagValue(fs, config.HTTPAuthTokenFileFlagName)

	if len(username) == 0 && len(password) == 0 && len(token) == 0 && len(tokenFile) == 0 {
		logger.Warning("HTTP server: authentication is not enabled")
	}

	return nil
}
