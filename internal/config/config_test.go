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
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDebugFlag tests that the debug flag is properly defined
func TestDebugFlag(t *testing.T) {
	// Test that the debug flag is properly defined
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	// Register all flags
	for _, flagInfo := range AllFlags {
		switch flagInfo.Type {
		case FlagTypeDuration:
			fs.Duration(flagInfo.Name, 0, flagInfo.Description)
		case FlagTypeBool:
			fs.Bool(flagInfo.Name, false, flagInfo.Description)
		default:
			fs.String(flagInfo.Name, "", flagInfo.Description)
		}
	}

	// Check if debug flag exists
	debugFlag := fs.Lookup(DebugFlagName)
	if debugFlag == nil {
		t.Errorf("Debug flag '%s' not found", DebugFlagName)
		return // Early return to avoid nil pointer dereference
	}

	// Check if debug flag has correct default value
	if debugFlag.DefValue != "false" {
		t.Errorf("Debug flag default value is '%s', expected 'false'", debugFlag.DefValue)
	}

	// Check if debug flag has correct description
	expectedDesc := "Enable debug logging"
	if debugFlag.Usage != expectedDesc {
		t.Errorf("Debug flag description is '%s', expected '%s'", debugFlag.Usage, expectedDesc)
	}
}

// TestAllFlags tests that all flags are properly defined
func TestAllFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	// Register all flags
	for _, flagInfo := range AllFlags {
		switch flagInfo.Type {
		case FlagTypeDuration:
			fs.Duration(flagInfo.Name, 0, flagInfo.Description)
		case FlagTypeBool:
			fs.Bool(flagInfo.Name, false, flagInfo.Description)
		default:
			fs.String(flagInfo.Name, "", flagInfo.Description)
		}
	}

	// Check that all flags are registered
	for _, flagInfo := range AllFlags {
		f := fs.Lookup(flagInfo.Name)
		if f == nil {
			t.Errorf("Flag '%s' not found", flagInfo.Name)
		}
	}
}

// TestSpecialCommandFlags tests that special command flags are properly defined
func TestSpecialCommandFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	// Register special command flags
	for _, flagInfo := range SpecialCommandFlags {
		switch flagInfo.Type {
		case FlagTypeDuration:
			fs.Duration(flagInfo.Name, 0, flagInfo.Description)
		case FlagTypeBool:
			fs.Bool(flagInfo.Name, false, flagInfo.Description)
		default:
			fs.String(flagInfo.Name, "", flagInfo.Description)
		}
	}

	// Check that all special command flags are registered
	for _, flagInfo := range SpecialCommandFlags {
		f := fs.Lookup(flagInfo.Name)
		if f == nil {
			t.Errorf("Special command flag '%s' not found", flagInfo.Name)
		}
	}
}

// TestConstants tests that all constants are properly defined
func TestConstants(t *testing.T) {
	// Test that all flag name constants are non-empty
	flagNames := []string{
		SyncIntervalFlagName,
		RepoURLFlagName,
		RepoBranchFlagName,
		RepoUserFlagName,
		RepoTokenFlagName,
		LocalPathFlagName,
		HTTPServerAddrFlagName,
		HTTPAuthUsernameFlagName,
		HTTPAuthPasswordFlagName,
		HTTPAuthTokenFlagName,
		ConfigFileFlagName,
		GenConfigFlagName,
		ConfigHelpFlagName,
		DebugFlagName,
	}

	for _, flagName := range flagNames {
		if flagName == "" {
			t.Errorf("Flag name constant is empty")
		}
	}

	// Test that all configuration key constants are non-empty
	configKeys := []string{
		GitlabRepoAuthKey,
		HTTPServerAuthKey,
		SyncIntervalKey,
		HTTPServerAddrKey,
		SyncLocalPathKey,
		GitlabRepoURLKey,
		GitlabRepoBranchKey,
		GitlabRepoAuthUserKey,
		GitlabRepoAuthTokenKey,
		HTTPServerAuthUsernameKey,
		HTTPServerAuthPasswordKey,
		ConfigFileKey,
		DebugKey,
	}

	for _, configKey := range configKeys {
		if configKey == "" {
			t.Errorf("Configuration key constant is empty")
		}
	}

	// Test default value constants
	if DefaultHTTPServerAddr == "" {
		t.Errorf("DefaultHTTPServerAddr constant is empty")
	}
}

// TestGenerateSampleConfig tests the GenerateSampleConfig function
func TestGenerateSampleConfig(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Test generating a sample config
	err := GenerateSampleConfig(configFile)
	if err != nil {
		t.Fatalf("GenerateSampleConfig failed: %v", err)
	}

	// Check that the file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

// TestGetConfigDescription tests the GetConfigDescription function
func TestGetConfigDescription(t *testing.T) {
	descriptions := GetConfigDescription()

	// Check that we get descriptions for all configuration keys
	expectedKeys := []string{
		GitlabRepoURLKey,
		GitlabRepoBranchKey,
		GitlabRepoAuthUserKey,
		GitlabRepoAuthTokenKey,
		SyncLocalPathKey,
		SyncIntervalKey,
		HTTPServerAddrKey,
		HTTPServerAuthUsernameKey,
		HTTPServerAuthPasswordKey,
		ConfigFileKey,
		DebugKey,
		GitlabRepoAuthKey,
		HTTPServerAuthKey,
	}

	for _, key := range expectedKeys {
		if _, exists := descriptions[key]; !exists {
			t.Errorf("Description for key '%s' not found", key)
		}
	}
}

// TestFlagInfoStructure tests the FlagInfo structure
func TestFlagInfoStructure(t *testing.T) {
	// Test that AllFlags contains the expected number of flags
	expectedFlagCount := 17 // Updated to include repo-token-file, http-auth-token-file, and webhook flags
	if len(AllFlags) != expectedFlagCount {
		t.Errorf("Expected %d flags, got %d", expectedFlagCount, len(AllFlags))
	}

	// Test that SpecialCommandFlags contains the expected number of flags
	expectedSpecialFlagCount := 2 // gen-config and config-help
	if len(SpecialCommandFlags) != expectedSpecialFlagCount {
		t.Errorf("Expected %d special flags, got %d", expectedSpecialFlagCount, len(SpecialCommandFlags))
	}

	// Test that each FlagInfo has required fields
	for _, flagInfo := range AllFlags {
		if flagInfo.Name == "" {
			t.Error("FlagInfo.Name is empty")
		}
		if flagInfo.Description == "" {
			t.Errorf("FlagInfo.Description is empty for flag '%s'", flagInfo.Name)
		}
		if flagInfo.EnvVar == "" {
			t.Errorf("FlagInfo.EnvVar is empty for flag '%s'", flagInfo.Name)
		}
	}

	// Test that each SpecialCommandFlag has required fields
	for _, flagInfo := range SpecialCommandFlags {
		if flagInfo.Name == "" {
			t.Error("SpecialCommandFlag.Name is empty")
		}
		if flagInfo.Description == "" {
			t.Errorf("SpecialCommandFlag.Description is empty for flag '%s'", flagInfo.Name)
		}
	}
}

// TestGetEnvVarForConfigKey tests the getEnvVarForConfigKey function
func TestGetEnvVarForConfigKey(t *testing.T) {
	// Test with a regular config key
	envVar := getEnvVarForConfigKey(GitlabRepoURLKey)
	if envVar != "GITSYNC_REPOSITORY_URL" {
		t.Errorf("Expected 'GITSYNC_REPOSITORY_URL', got '%s'", envVar)
	}

	// Test with a special token file key
	envVar = getEnvVarForConfigKey(GitlabRepoAuthKey)
	if envVar != "GITSYNC_REPOSITORY_TOKEN_FILE" {
		t.Errorf("Expected 'GITSYNC_REPOSITORY_TOKEN_FILE', got '%s'", envVar)
	}

	// Test with HTTP server token file key
	envVar = getEnvVarForConfigKey(HTTPServerAuthKey)
	if envVar != "GITSYNC_HTTP_AUTH_TOKEN_FILE" {
		t.Errorf("Expected 'GITSYNC_HTTP_AUTH_TOKEN_FILE', got '%s'", envVar)
	}
}

// TestGetFlagNameForConfigKey tests the getFlagNameForConfigKey function
func TestGetFlagNameForConfigKey(t *testing.T) {
	// Test with a regular config key
	flagName := getFlagNameForConfigKey(GitlabRepoURLKey)
	if flagName != RepoURLFlagName {
		t.Errorf("Expected '%s', got '%s'", RepoURLFlagName, flagName)
	}

	// Test with a special token file key
	flagName = getFlagNameForConfigKey(GitlabRepoAuthKey)
	if flagName != "repo-token-file" {
		t.Errorf("Expected 'repo-token-file', got '%s'", flagName)
	}

	// Test with HTTP server token file key
	flagName = getFlagNameForConfigKey(HTTPServerAuthKey)
	if flagName != "http-auth-token-file" {
		t.Errorf("Expected 'http-auth-token-file', got '%s'", flagName)
	}
}

// TestBindEnvs tests the bindEnvs function
func TestBindEnvs(t *testing.T) {
	// This function is hard to test directly since it interacts with viper
	// We'll just test that it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("bindEnvs panicked: %v", r)
		}
	}()
	bindEnvs()
}

// TestSetValuesFromFlags tests the setValuesFromFlags function
func TestSetValuesFromFlags(t *testing.T) {
	// Create a flag set with some values
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String(RepoURLFlagName, "https://example.com/repo.git", "Repository URL")
	fs.String(RepoBranchFlagName, "main", "Repository branch")
	fs.Duration(SyncIntervalFlagName, 30*time.Second, "Sync interval")

	// Parse the flags
	err := fs.Parse([]string{
		"--" + RepoURLFlagName + "=https://example.com/repo.git",
		"--" + RepoBranchFlagName + "=main",
		"--" + SyncIntervalFlagName + "=30s",
	})
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// This function is hard to test directly since it interacts with viper
	// We'll just test that it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("setValuesFromFlags panicked: %v", r)
		}
	}()
	setValuesFromFlags(fs)
}

// TestPrintConfigHelp tests the PrintConfigHelp function
func TestPrintConfigHelp(t *testing.T) {
	// We'll just test that it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintConfigHelp panicked: %v", r)
		}
	}()
	PrintConfigHelp()
}

// TestPrintConfigHelpInFormat tests the PrintConfigHelpInFormat function
func TestPrintConfigHelpInFormat(t *testing.T) {
	// We'll just test that it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintConfigHelpInFormat panicked: %v", r)
		}
	}()
	PrintConfigHelpInFormat()
}

// TestPrintUnifiedHelp tests the PrintUnifiedHelp function
func TestPrintUnifiedHelp(t *testing.T) {
	// We'll just test that it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintUnifiedHelp panicked: %v", r)
		}
	}()
	PrintUnifiedHelp()
}

// TestShowEnvAndFlag tests the showEnvAndFlag function
func TestShowEnvAndFlag(t *testing.T) {
	// Test with a regular config key
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("showEnvAndFlag panicked with regular key: %v", r)
		}
	}()
	showEnvAndFlag(GitlabRepoURLKey)

	// Test with a special token file key
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("showEnvAndFlag panicked with token file key: %v", r)
		}
	}()
	showEnvAndFlag(GitlabRepoAuthKey)

	// Test with HTTP server token file key
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("showEnvAndFlag panicked with HTTP token file key: %v", r)
		}
	}()
	showEnvAndFlag(HTTPServerAuthKey)
}

// TestConfigStructure tests the Config struct structure
func TestConfigStructure(t *testing.T) {
	// Create a config instance
	config := &Config{}

	// Initialize the repositories map
	config.Repositories = make(map[string]RepositoryConfig)

	// Test that we can access all fields without panicking
	// Since we now have multiple repositories, we'll test with a sample repository
	repoConfig := RepositoryConfig{}
	repoConfig.Gitlab.RepoURL = "test-url"
	repoConfig.Gitlab.RepoBranch = "test-branch"
	repoConfig.Gitlab.RepoAuth.User = "test-user"
	repoConfig.Gitlab.RepoAuth.Token = "test-token"
	repoConfig.Gitlab.RepoAuth.TokenFile = "test-token-file"
	repoConfig.Sync.LocalPath = "test-path"
	repoConfig.Sync.Interval = 30

	config.Repositories["test-repo"] = repoConfig

	_ = config.Debug
	_ = config.Defaults.Gitlab.RepoAuth.User
	_ = config.Defaults.Gitlab.RepoBranch
	_ = config.Defaults.Sync.Interval
	_ = config.HttpServer.Addr
	_ = config.HttpServer.Auth.Username
	_ = config.HttpServer.Auth.Password
	_ = config.HttpServer.Auth.Token
	_ = config.HttpServer.Auth.TokenFile

	// Test accessing the repository config
	testRepo := config.Repositories["test-repo"]
	_ = testRepo.Gitlab.RepoURL
	_ = testRepo.Gitlab.RepoBranch
	_ = testRepo.Gitlab.RepoAuth.User
	_ = testRepo.Gitlab.RepoAuth.Token
	_ = testRepo.Gitlab.RepoAuth.TokenFile
	_ = testRepo.Sync.LocalPath
	_ = testRepo.Sync.Interval
}
