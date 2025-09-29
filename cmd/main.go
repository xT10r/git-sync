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

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"git-sync/git"
	"git-sync/internal/config"
	"git-sync/internal/flags"
	"git-sync/internal/gitsync"
	"git-sync/internal/interfaces"
	"git-sync/internal/server"  // Use server package instead of http package
	"git-sync/internal/version" // Add version import
	"git-sync/logger"

	// Import packages that register routes (using blank identifiers to avoid "imported and not used" errors)
	_ "git-sync/api"
	_ "git-sync/internal/handlers"
)

func main() {
	// Parse flags first for early command detection
	tempFs := flag.NewFlagSet("git-sync-temp", flag.ContinueOnError)
	// Suppress default error output
	tempFs.SetOutput(io.Discard)
	genConfigFlag := tempFs.Bool(config.GenConfigFlagName, false, "Generate config file")
	configHelpFlag := tempFs.Bool(config.ConfigHelpFlagName, false, "Show config help")
	helpFlag := tempFs.Bool("help", false, "Show help")
	versionFlag := tempFs.Bool("version", false, "Show version information") // Add version flag

	// Also add the config-file flag to the temporary flag set so it can be parsed
	configFilePath := tempFs.String(config.ConfigFileFlagName, "", "Path to the configuration file")

	// Parse only the flags, not subcommands
	args := os.Args[1:]
	var flagArgs []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
		} else {
			// Stop at first non-flag argument (could be subcommand)
			break
		}
	}

	// Try to parse flags
	err := tempFs.Parse(flagArgs)
	if err != nil {
		// If there's a flag parsing error, show a more helpful message
		if err == flag.ErrHelp {
			// Use the unified help from config package
			config.PrintUnifiedHelp()
			return
		}
		// Check if the error is about the config-file flag needing an argument
		// In this case, we should continue with the main parsing and let it show a proper error
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, config.ConfigFileFlagName) && strings.Contains(errStr, "needs an argument"):
			// Continue with main parsing, the config-file flag will be handled there
		case strings.Contains(errStr, "flag provided but not defined") && strings.Contains(errStr, config.ConfigFileFlagName):
			// Continue with main parsing, the config-file flag will be handled there
		default:
			fmt.Fprintf(os.Stderr, "[ERROR] Invalid flag usage: %v\n", err)
			fmt.Fprintf(os.Stderr, "Use --help or --config-help for more information\n")
			os.Exit(1)
		}
	}

	// Check if we're generating a config file
	if *genConfigFlag {
		handleGenConfig(tempFs)
		return
	}

	// Check if we're showing config help
	if *configHelpFlag {
		// Use the unified help from config package
		config.PrintUnifiedHelp()
		return
	}

	// Check if we're showing help
	if *helpFlag {
		// Use the unified help from config package
		config.PrintUnifiedHelp()
		return
	}

	// Check if we're showing version
	if *versionFlag {
		// Show version information
		fmt.Println(version.String())
		return
	}

	// Create context and cancel function
	ctx, cancel := context.WithCancel(context.Background())
	// Note: We can't use defer cancel() here because of os.Exit calls below
	// We'll call cancel() explicitly before os.Exit or at the end of the function

	// Parse flags with the main flag set
	flagSet := flags.NewConsoleFlags()

	// Check if we need to show help (from the main parser)
	helpNeeded := false
	flagSet.Gitsync.VisitAll(func(f *flag.Flag) {
		if f.Name == "help" {
			if help, err := strconv.ParseBool(f.Value.String()); err == nil && help {
				helpNeeded = true
			}
		}
	})

	if helpNeeded {
		config.PrintUnifiedHelp()
		cancel() // Cancel context before returning
		return
	}

	// Read configuration using Viper (combining flags, env vars, and config file)
	cfg, err := config.ReadConfig(flagSet.Gitsync, *configFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading configuration: %v\n", err)
		cancel() // Cancel context before exiting
		os.Exit(1)
	}

	// Check if we have any repositories configured
	if len(cfg.Repositories) == 0 {
		fmt.Fprintf(os.Stderr, "No repositories configured. Please configure at least one repository in the config file.\n")
		cancel() // Cancel context before exiting
		os.Exit(1)
	}

	// Set debug mode for logger
	logger.SetDebug(cfg.Debug)

	// Enable detailed logging (with file and line numbers) when debug mode is enabled
	logger.SetDetailedLogging(cfg.Debug)

	// Log the configuration
	logger.Info("Starting git-sync with %d repositories\n", len(cfg.Repositories))

	// Create GitSync instances for each repository
	var gitSyncs []*gitsync.GitSync
	var gitRepos []interfaces.Gitter

	// Start http server using the server package
	serverInstance := server.NewServer(cfg)
	serverInstance.Start(ctx)

	// Create GitSync and GitRepository instances for each repository
	for repoName, repoConfig := range cfg.Repositories {
		logger.Info("Configuring repository: %s\n", repoName)
		logger.Debug("Processing repository configuration: %s", repoName)

		// Create a temporary flag set for this repository
		tempFs := flag.NewFlagSet("git-sync-"+repoName, flag.ContinueOnError)

		// Add flags with values from the repository config
		tempFs.String(config.RepoURLFlagName, repoConfig.Gitlab.RepoURL, "")
		tempFs.String(config.RepoBranchFlagName, repoConfig.Gitlab.RepoBranch, "")
		tempFs.String(config.LocalPathFlagName, repoConfig.Sync.LocalPath, "")
		tempFs.String(config.RepoUserFlagName, repoConfig.Gitlab.RepoAuth.User, "")
		tempFs.String(config.RepoTokenFlagName, repoConfig.Gitlab.RepoAuth.Token, "")
		tempFs.String(config.RepoTokenFileFlagName, repoConfig.Gitlab.RepoAuth.TokenFile, "")

		// Parse the flags
		if err := tempFs.Parse([]string{}); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing flags for repository %s: %v\n", repoName, err)
			os.Exit(1)
		}

		// Check if required configuration values are set
		if err := checkRequiredConfigValuesForRepo(repoName, &repoConfig); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}

		// Validate configuration values
		if err := validateConfigValuesForRepo(repoName, &repoConfig); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}

		// Create GitSync instance
		gitSync, err := gitsync.NewGitSync(tempFs, &config.Config{
			Debug:      cfg.Debug,
			Defaults:   cfg.Defaults,
			HttpServer: cfg.HttpServer,
			Repositories: map[string]config.RepositoryConfig{
				repoName: repoConfig,
			},
		}, ctx, repoName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating GitSync object for repository %s: %v\n", repoName, err)
			os.Exit(1)
		}
		gitSyncs = append(gitSyncs, gitSync)

		// Create GitRepository instance
		gitRepo, err := git.NewGitRepository(tempFs, &config.Config{
			Debug:      cfg.Debug,
			Defaults:   cfg.Defaults,
			HttpServer: cfg.HttpServer,
			Repositories: map[string]config.RepositoryConfig{
				repoName: repoConfig,
			},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating GitRepository object for repository %s: %v\n", repoName, err)
			os.Exit(1)
		}
		gitRepos = append(gitRepos, gitRepo)

		// Log the repository configuration
		logger.Info("Repository %s: GitLab repo %s, branch %s, local path %s, sync interval %ds\n",
			repoName, repoConfig.Gitlab.RepoURL, repoConfig.Gitlab.RepoBranch, repoConfig.Sync.LocalPath, repoConfig.Sync.Interval)
	}

	// Start periodic synchronization for each repository in separate goroutines
	for i, gitSync := range gitSyncs {
		go gitSync.Start(gitRepos[i])
	}

	// Wait for SIGINT or SIGTERM signals to terminate the program
	waitForSignals(cancel)

	// Cancel context and wait for goroutines to finish
	cancel()
}

// checkRequiredConfigValuesForRepo checks if all required configuration values are set for a specific repository
func checkRequiredConfigValuesForRepo(repoName string, repoConfig *config.RepositoryConfig) error {
	var missingFields []string

	if repoConfig.Gitlab.RepoURL == "" {
		missingFields = append(missingFields, "repo-url")
	}

	if repoConfig.Gitlab.RepoBranch == "" {
		missingFields = append(missingFields, "repo-branch")
	}

	if repoConfig.Sync.LocalPath == "" {
		missingFields = append(missingFields, "local-path")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("required configuration values for repository %s are missing: %v. Use --help or --config-help for more information", repoName, strings.Join(missingFields, ", "))
	}

	return nil
}

// validateConfigValuesForRepo validates configuration values for a specific repository
func validateConfigValuesForRepo(repoName string, repoConfig *config.RepositoryConfig) error {
	// Validate repo URL
	if _, err := url.ParseRequestURI(repoConfig.Gitlab.RepoURL); err != nil {
		return fmt.Errorf("invalid repository URL for repository %s: %v", repoName, err)
	}

	// Validate local path
	if _, err := os.Stat(repoConfig.Sync.LocalPath); os.IsNotExist(err) {
		return fmt.Errorf("specified local path for repository %s does not exist: %s", repoName, repoConfig.Sync.LocalPath)
	}

	// Validate sync interval
	if repoConfig.Sync.Interval <= 0 {
		return fmt.Errorf("sync interval for repository %s must be positive", repoName)
	}

	// Validate HTTP server address if set
	// Note: HTTP server validation is done once for the entire config, not per repository

	return nil
}

// handleGenConfig handles the gen-config command
func handleGenConfig(fs *flag.FlagSet) {
	configPath := "config.yaml"

	// Try to get output path from flag if available
	if outputFlag := fs.Lookup(config.ConfigFileFlagName); outputFlag != nil {
		if outputPath := outputFlag.Value.String(); outputPath != "" {
			configPath = outputPath
		}
	}

	if err := config.GenerateSampleConfig(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Sample config file generated at: %s\n", configPath)
	fmt.Fprintf(os.Stderr, "You can view all configuration options with: git-sync --config-help\n")
}

// waitForSignals waits for SIGINT or SIGTERM signals and calls the cancel function to terminate the program.
func waitForSignals(cancel context.CancelFunc) {
	// Call the cancel function to terminate the program
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signals
	sig := <-signalChan
	logger.Info("Received signal %s. Shutting down...\n", sig)
	logger.Debug("Shutdown initiated by signal: %s", sig)
}
