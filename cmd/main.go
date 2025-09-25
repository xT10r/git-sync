// Copyright 2024 Aleksey Dobshikov
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
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/viper"

	"git-sync/git"
	"git-sync/internal/config"
	"git-sync/internal/flags"
	"git-sync/internal/gitsync"
	"git-sync/internal/http"
	"git-sync/logger"
)

func main() {
	// Parse flags first for early command detection
	tempFs := flag.NewFlagSet("git-sync-temp", flag.ContinueOnError)
	genConfigFlag := tempFs.Bool(config.GenConfigFlagName, false, "Generate config file")
	configHelpFlag := tempFs.Bool(config.ConfigHelpFlagName, false, "Show config help")
	helpFlag := tempFs.Bool("help", false, "Show help")

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
	_ = tempFs.Parse(flagArgs)

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

	// Create context and cancel function
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flagSet := flags.NewConsoleFlags()

	// Check if required flags are set
	if err := flagSet.CheckRequiredFlags(); err != nil {
		logger.Error("%v", err)
		fmt.Fprintf(os.Stderr, "\n")
		// Show unified help when required flags are missing
		config.PrintUnifiedHelp()
		os.Exit(0)
	}

	// Validate flag values
	if err := flagSet.ValidateFlags(); err != nil {
		logger.Error("%v", err)
		os.Exit(0)
	}

	// Read configuration using Viper (combining flags, env vars, and config file)
	cfg, err := config.ReadConfig(flagSet.Gitsync, "")
	if err != nil {
		logger.Error("Error reading configuration: %v\n", err)
		return
	}

	// Set debug mode for logger
	logger.SetDebug(viper.GetBool(config.DebugKey))

	// Log the configuration source being used
	logger.Info("Using configuration: GitLab repo %s, branch %s, local path %s, sync interval %ds\n",
		cfg.Gitlab.RepoURL, cfg.Gitlab.RepoBranch, cfg.Sync.LocalPath, cfg.Sync.Interval)

	gitSync, err := gitsync.NewGitSync(flagSet.Gitsync, cfg, ctx)
	if err != nil {
		logger.Error("Error creating GitSync object: %v\n", err)
		return
	}

	gitRepo, err := git.NewGitRepository(flagSet.Gitsync, cfg)
	if err != nil {
		logger.Error("Error creating GitRepository object: %v\n", err)
		return
	}

	// Start http server
	http.StartServer(flagSet.Gitsync, ctx)

	// Start periodic synchronization in a separate goroutine
	go gitSync.Start(gitRepo)

	// Wait for SIGINT or SIGTERM signals to terminate the program
	waitForSignals(cancel)

	// Cancel context and wait for goroutines to finish
	cancel()
}

// handleGenConfig handles the gen-config command
func handleGenConfig(fs *flag.FlagSet) {
	configPath := "config.yaml"

	// Try to get output path from flag if available
	if outputFlag := fs.Lookup("output"); outputFlag != nil {
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
}
