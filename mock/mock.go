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

package mock

import (
	"flag"
	"git-sync/git"
	"git-sync/internal/config"
	"time"
)

// Flags creates a new flag set with default values for testing
func Flags() *flag.FlagSet {
	mockFlags := flag.NewFlagSet("test", flag.ContinueOnError)
	mockFlags.Duration(config.SyncIntervalFlagName, 30*time.Second, "Interval for synchronization")
	mockFlags.String(config.RepoBranchFlagName, "master", "Branch of the repository")
	mockFlags.String(config.RepoUserFlagName, "user", "Repository authentication user")
	mockFlags.String(config.RepoTokenFlagName, "token", "Repository authentication token")
	return mockFlags
}

// FlagsWithValues creates a new flag set with specific values for testing
func FlagsWithValues(repoURL, localPath, repoBranch string, syncInterval time.Duration) *flag.FlagSet {
	mockFlags := flag.NewFlagSet("test", flag.ContinueOnError)
	mockFlags.String(config.RepoURLFlagName, repoURL, "URL of the repository")
	mockFlags.String(config.LocalPathFlagName, localPath, "Local path for the repository")
	mockFlags.String(config.RepoBranchFlagName, repoBranch, "Branch of the repository")
	mockFlags.Duration(config.SyncIntervalFlagName, syncInterval, "Interval for synchronization")
	mockFlags.String(config.RepoUserFlagName, "user", "Repository authentication user")
	mockFlags.String(config.RepoTokenFlagName, "token", "Repository authentication token")
	return mockFlags
}

// Gitter is a mock implementation of the Gitter interface
type Gitter struct {
	hasChanges bool
	syncError  error
	options    *git.GitRepositoryOptions
}

// NewGitter creates a new mock Gitter with default values
func NewGitter() *Gitter {
	return &Gitter{
		hasChanges: false,
		syncError:  nil,
		options:    git.NewGitRepositoryOptions("http://example.com", "master", "/path/to/local/repo", "user", "token", "origin"),
	}
}

// NewGitterWithValues creates a new mock Gitter with specific values
func NewGitterWithValues(hasChanges bool, syncError error, options *git.GitRepositoryOptions) *Gitter {
	return &Gitter{
		hasChanges: hasChanges,
		syncError:  syncError,
		options:    options,
	}
}

// SetHasChanges sets the hasChanges flag for the mock
func (m *Gitter) SetHasChanges(hasChanges bool) {
	m.hasChanges = hasChanges
}

// SetSyncError sets the error to be returned by Sync method
func (m *Gitter) SetSyncError(err error) {
	m.syncError = err
}

// SetOptions sets the repository options for the mock
func (m *Gitter) SetOptions(options *git.GitRepositoryOptions) {
	m.options = options
}

func (m *Gitter) Sync() error {
	return m.syncError
}

func (m *Gitter) Options() *git.GitRepositoryOptions {
	return m.options
}

func (m *Gitter) HasChanges() bool {
	return m.hasChanges
}

func (m *Gitter) Commit() (*git.CommitInfo, error) {
	return &git.CommitInfo{Hash: "mockhash", Date: time.Now()}, nil
}

func (m *Gitter) CommitHash() string {
	return "mockhash"
}

// MockConfig creates a mock configuration for testing
func MockConfig() *config.Config {
	cfg := &config.Config{}

	cfg.Gitlab.RepoURL = "https://gitlab.com/test/repo.git"
	cfg.Gitlab.RepoBranch = "main"
	cfg.Gitlab.RepoAuth.User = "testuser"
	cfg.Gitlab.RepoAuth.Token = "testtoken"
	cfg.Sync.LocalPath = "/tmp/testrepo"
	cfg.Sync.Interval = 30

	return cfg
}

// MockConfigWithValues creates a mock configuration with specific values for testing
func MockConfigWithValues(repoURL, repoBranch, user, token, localPath string, interval int) *config.Config {
	cfg := &config.Config{}

	cfg.Gitlab.RepoURL = repoURL
	cfg.Gitlab.RepoBranch = repoBranch
	cfg.Gitlab.RepoAuth.User = user
	cfg.Gitlab.RepoAuth.Token = token
	cfg.Sync.LocalPath = localPath
	cfg.Sync.Interval = interval

	return cfg
}
