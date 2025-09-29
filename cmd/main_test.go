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
	"flag"
	"git-sync/internal/config"
	"os"
	"testing"
)

// Test constants to avoid goconst warnings
const (
	testRepoURL   = "https://gitlab.com/test/repo.git"
	testBranch    = "main"
	testLocalPath = "/tmp/test-repo"
)

// TestCheckRequiredConfigValuesForRepo tests the checkRequiredConfigValuesForRepo function
func TestCheckRequiredConfigValuesForRepo(t *testing.T) {
	// Test with valid configuration
	repoConfig := &config.RepositoryConfig{}
	repoConfig.Gitlab.RepoURL = testRepoURL
	repoConfig.Gitlab.RepoBranch = testBranch
	repoConfig.Sync.LocalPath = testLocalPath

	err := checkRequiredConfigValuesForRepo("test-repo", repoConfig)
	if err != nil {
		t.Errorf("Expected no error for valid configuration, got: %v", err)
	}

	// Test with missing repo URL
	repoConfigMissingURL := &config.RepositoryConfig{}
	repoConfigMissingURL.Gitlab.RepoBranch = testBranch
	repoConfigMissingURL.Sync.LocalPath = testLocalPath

	err = checkRequiredConfigValuesForRepo("test-repo", repoConfigMissingURL)
	if err == nil {
		t.Error("Expected error for missing repo URL, got nil")
	}

	// Test with missing repo branch
	repoConfigMissingBranch := &config.RepositoryConfig{}
	repoConfigMissingBranch.Gitlab.RepoURL = testRepoURL
	repoConfigMissingBranch.Sync.LocalPath = testLocalPath

	err = checkRequiredConfigValuesForRepo("test-repo", repoConfigMissingBranch)
	if err == nil {
		t.Error("Expected error for missing repo branch, got nil")
	}

	// Test with missing local path
	repoConfigMissingPath := &config.RepositoryConfig{}
	repoConfigMissingPath.Gitlab.RepoURL = testRepoURL
	repoConfigMissingPath.Gitlab.RepoBranch = testBranch

	err = checkRequiredConfigValuesForRepo("test-repo", repoConfigMissingPath)
	if err == nil {
		t.Error("Expected error for missing local path, got nil")
	}
}

// TestValidateConfigValuesForRepo tests the validateConfigValuesForRepo function
func TestValidateConfigValuesForRepo(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gitsync-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with valid configuration
	repoConfig := &config.RepositoryConfig{}
	repoConfig.Gitlab.RepoURL = testRepoURL
	repoConfig.Gitlab.RepoBranch = testBranch
	repoConfig.Sync.LocalPath = tempDir
	repoConfig.Sync.Interval = 30

	err = validateConfigValuesForRepo("test-repo", repoConfig)
	if err != nil {
		t.Errorf("Expected no error for valid configuration, got: %v", err)
	}

	// Test with invalid URL
	repoConfigInvalidURL := &config.RepositoryConfig{}
	repoConfigInvalidURL.Gitlab.RepoURL = "invalid-url"
	repoConfigInvalidURL.Gitlab.RepoBranch = testBranch
	repoConfigInvalidURL.Sync.LocalPath = tempDir
	repoConfigInvalidURL.Sync.Interval = 30

	err = validateConfigValuesForRepo("test-repo", repoConfigInvalidURL)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	// Test with non-existent local path
	repoConfigInvalidPath := &config.RepositoryConfig{}
	repoConfigInvalidPath.Gitlab.RepoURL = testRepoURL
	repoConfigInvalidPath.Gitlab.RepoBranch = testBranch
	repoConfigInvalidPath.Sync.LocalPath = "/non/existent/path"
	repoConfigInvalidPath.Sync.Interval = 30

	err = validateConfigValuesForRepo("test-repo", repoConfigInvalidPath)
	if err == nil {
		t.Error("Expected error for non-existent local path, got nil")
	}

	// Test with invalid interval
	repoConfigInvalidInterval := &config.RepositoryConfig{}
	repoConfigInvalidInterval.Gitlab.RepoURL = testRepoURL
	repoConfigInvalidInterval.Gitlab.RepoBranch = testBranch
	repoConfigInvalidInterval.Sync.LocalPath = tempDir
	repoConfigInvalidInterval.Sync.Interval = -1

	err = validateConfigValuesForRepo("test-repo", repoConfigInvalidInterval)
	if err == nil {
		t.Error("Expected error for invalid interval, got nil")
	}

	// Test with zero interval
	repoConfigZeroInterval := &config.RepositoryConfig{}
	repoConfigZeroInterval.Gitlab.RepoURL = testRepoURL
	repoConfigZeroInterval.Gitlab.RepoBranch = testBranch
	repoConfigZeroInterval.Sync.LocalPath = tempDir
	repoConfigZeroInterval.Sync.Interval = 0

	err = validateConfigValuesForRepo("test-repo", repoConfigZeroInterval)
	if err == nil {
		t.Error("Expected error for zero interval, got nil")
	}
}

// TestHandleGenConfig tests the handleGenConfig function
func TestHandleGenConfig(t *testing.T) {
	// Create a temporary file path for testing
	tempFile, err := os.CreateTemp("", "gitsync-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Create a flag set with the config file flag
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String(config.ConfigFileFlagName, tempFile.Name(), "Path to the configuration file")

	// Parse the flags
	err = fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Test that handleGenConfig doesn't panic
	// We can't easily test the file creation without mocking, but we can at least
	// verify the function doesn't crash
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("handleGenConfig panicked: %v", r)
			}
		}()
		handleGenConfig(fs)
	}()
}
