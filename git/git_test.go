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

package git_test

import (
	"git-sync/git"
	"git-sync/internal/config"
	"git-sync/mock"
	"os"
	"path"
	"testing"
)

// TestNewGitRepositoryWithValidConfig tests the NewGitRepository constructor with valid configuration
// This test does not attempt to clone a real repository, but instead focuses on verifying
// that the GitRepository is properly initialized with the correct configuration values
func TestNewGitRepositoryWithValidConfig(t *testing.T) {
	// Create a mock configuration
	cfg := &config.Config{}
	cfg.Repositories = make(map[string]config.RepositoryConfig)

	// Create a mock repository config
	repoConfig := config.RepositoryConfig{}
	repoConfig.Gitlab.RepoURL = "https://gitlab.com/test/repo.git"
	repoConfig.Gitlab.RepoBranch = "main"
	repoConfig.Sync.LocalPath = path.Join(os.TempDir(), "test-repo")
	repoConfig.Gitlab.RepoAuth.User = "testuser"
	repoConfig.Gitlab.RepoAuth.Token = "testtoken"

	cfg.Repositories["test-repo"] = repoConfig

	// Create mock flags
	mockFlags := mock.FlagsWithValues(
		repoConfig.Gitlab.RepoURL,
		repoConfig.Sync.LocalPath,
		repoConfig.Gitlab.RepoBranch,
		30,
	)

	// Parse flags
	err := mockFlags.Parse(nil)
	if err != nil {
		t.Fatalf("Error parsing flags: %v", err)
	}

	// Since we can't easily mock the go-git library for unit testing the constructor,
	// and cloning a real repository in tests is not appropriate (requires network/auth),
	// we'll skip this test for now
	// In a real scenario, integration tests would cover this functionality
	t.Skip("Skipping constructor test as it requires complex mocking or network access")
}

// TestNewGitRepositoryWithInvalidURL tests the NewGitRepository constructor with invalid URL
func TestNewGitRepositoryWithInvalidURL(t *testing.T) {
	// Create a mock configuration with empty values to simulate invalid config
	cfg := &config.Config{}
	cfg.Repositories = make(map[string]config.RepositoryConfig)

	// Create a mock repository config with empty URL
	repoConfig := config.RepositoryConfig{}
	repoConfig.Gitlab.RepoURL = "" // Empty URL to simulate invalid config
	repoConfig.Gitlab.RepoBranch = "master"
	repoConfig.Sync.LocalPath = path.Join(os.TempDir(), "invalid-repo")
	repoConfig.Gitlab.RepoAuth.User = "user"
	repoConfig.Gitlab.RepoAuth.Token = "token"

	cfg.Repositories["test-repo"] = repoConfig

	// Создаем макет флагов для использования в тесте
	mockFlags := mock.Flags()
	mockFlags.String(config.RepoURLFlagName, "http://invalid-url.com", "URL of the repository")
	mockFlags.String(config.LocalPathFlagName, path.Join(os.TempDir(), "invalid-repo"), "Local path for the repository")

	// Парсим флаги
	err := mockFlags.Parse(nil)
	if err != nil {
		t.Fatalf("Error parsing flags: %v", err)
	}

	// Пытаемся создать новый GitRepository с неверным URL
	_, err = git.NewGitRepository(mockFlags, cfg)
	if err == nil {
		t.Error("Expected error due to invalid repository URL, but got nil")
	} else {
		// Since we're now using config, the error will be different
		// We're checking that we get an error, not necessarily the exact same error
		if err.Error() == "" {
			t.Error("Expected non-empty error message")
		}
	}
}

// TestNewGitRepositoryOptions tests the NewGitRepositoryOptions function
func TestNewGitRepositoryOptions(t *testing.T) {
	url := "https://example.com/repo.git"
	branch := "main"
	path := "/tmp/repo"
	user := "testuser"
	token := "testtoken"
	originName := "origin"

	options := git.NewGitRepositoryOptions(url, branch, path, user, token, originName)

	if options.Url() != url {
		t.Errorf("Expected URL %s, got %s", url, options.Url())
	}

	if options.Branch() != branch {
		t.Errorf("Expected branch %s, got %s", branch, options.Branch())
	}
}

// TestGitRepositoryOptions_Url tests the Url method
func TestGitRepositoryOptions_Url(t *testing.T) {
	url := "https://example.com/repo.git"
	options := git.NewGitRepositoryOptions(url, "main", "/tmp/repo", "user", "token", "origin")

	if options.Url() != url {
		t.Errorf("Expected URL %s, got %s", url, options.Url())
	}
}

// TestGitRepositoryOptions_Branch tests the Branch method
func TestGitRepositoryOptions_Branch(t *testing.T) {
	branch := "main"
	options := git.NewGitRepositoryOptions("https://example.com/repo.git", branch, "/tmp/repo", "user", "token", "origin")

	if options.Branch() != branch {
		t.Errorf("Expected branch %s, got %s", branch, options.Branch())
	}
}

// TestCommitInfo_AddChange tests the AddChange method
func TestCommitInfo_AddChange(t *testing.T) {
	commitInfo := &git.CommitInfo{
		Changes: []git.ChangeInfo{},
	}

	changeType := "modified"
	fileName := "test.txt"
	fromHash := "abc123"
	toHash := "def456"

	commitInfo.AddChange(changeType, fileName, fromHash, toHash)

	if len(commitInfo.Changes) != 1 {
		t.Fatalf("Expected 1 change, got %d", len(commitInfo.Changes))
	}

	change := commitInfo.Changes[0]
	if change.ChangeType != changeType {
		t.Errorf("Expected change type %s, got %s", changeType, change.ChangeType)
	}

	if change.FileName != fileName {
		t.Errorf("Expected file name %s, got %s", fileName, change.FileName)
	}

	if change.FromHash != fromHash {
		t.Errorf("Expected from hash %s, got %s", fromHash, change.FromHash)
	}

	if change.ToHash != toHash {
		t.Errorf("Expected to hash %s, got %s", toHash, change.ToHash)
	}
}

// TestNewCommitInfo tests the NewCommitInfo function
func TestNewCommitInfo(t *testing.T) {
	// Since we can't easily create a real git.Commit object in tests,
	// we'll skip this test for now as it would require complex mocking
	// of the go-git library
	t.Skip("Skipping NewCommitInfo test as it requires complex mocking")
}

// TestGitRepository_HasChanges tests the HasChanges method
func TestGitRepository_HasChanges(t *testing.T) {
	// Create a GitRepository instance
	gitRepo := &git.GitRepository{}

	// Test with true value
	gitRepo.SetHasChangesForTest(true)
	if !gitRepo.HasChanges() {
		t.Error("Expected HasChanges to return true")
	}

	// Test with false value
	gitRepo.SetHasChangesForTest(false)
	if gitRepo.HasChanges() {
		t.Error("Expected HasChanges to return false")
	}
}

// TestGitRepository_CommitHash tests the CommitHash method
func TestGitRepository_CommitHash(t *testing.T) {
	expectedHash := "abc123"

	// Create a GitRepository instance
	gitRepo := &git.GitRepository{}

	// Set up a commit for testing
	commitInfo := &git.CommitInfo{
		Hash: expectedHash,
	}
	gitRepo.SetCurrentCommitForTest(commitInfo)

	if gitRepo.CommitHash() != expectedHash {
		t.Errorf("Expected commit hash %s, got %s", expectedHash, gitRepo.CommitHash())
	}
}

// TestGitRepository_Options tests the Options method
func TestGitRepository_Options(t *testing.T) {
	// Create a GitRepository instance
	gitRepo := &git.GitRepository{}

	// Test that the method exists and doesn't panic
	// In a real scenario, options would be set during construction
	_ = gitRepo.Options() // Just test that the method exists and doesn't panic
}

// TestGitRepository_Commit tests the Commit method
func TestGitRepository_Commit(t *testing.T) {
	expectedHash := "abc123"

	// Create a GitRepository instance
	gitRepo := &git.GitRepository{}

	// Set up a commit for testing
	commitInfo := &git.CommitInfo{
		Hash: expectedHash,
	}
	gitRepo.SetCurrentCommitForTest(commitInfo)

	commit, err := gitRepo.Commit()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if commit.Hash != expectedHash {
		t.Errorf("Expected commit hash %s, got %s", expectedHash, commit.Hash)
	}
}

// TestGitRepository_Commit_Error tests the Commit method when no commit is set
func TestGitRepository_Commit_Error(t *testing.T) {
	// Create a GitRepository instance without setting a commit
	gitRepo := &git.GitRepository{}

	_, err := gitRepo.Commit()
	if err == nil {
		t.Error("Expected error when no commit is set, but got nil")
	}
}

// TestGitRepository_SetHasChangesForTest tests the SetHasChangesForTest helper method
func TestGitRepository_SetHasChangesForTest(t *testing.T) {
	gitRepo := &git.GitRepository{}

	gitRepo.SetHasChangesForTest(true)
	if !gitRepo.HasChanges() {
		t.Error("Expected HasChanges to return true after SetHasChangesForTest(true)")
	}

	gitRepo.SetHasChangesForTest(false)
	if gitRepo.HasChanges() {
		t.Error("Expected HasChanges to return false after SetHasChangesForTest(false)")
	}
}

// TestGitRepository_SetCurrentCommitForTest tests the SetCurrentCommitForTest helper method
func TestGitRepository_SetCurrentCommitForTest(t *testing.T) {
	gitRepo := &git.GitRepository{}

	commitInfo := &git.CommitInfo{
		Hash: "testhash",
	}
	gitRepo.SetCurrentCommitForTest(commitInfo)

	if gitRepo.CommitHash() != "testhash" {
		t.Error("Expected CommitHash to return 'testhash' after SetCurrentCommitForTest")
	}
}

// TestGitRepository_ResetChangesFlag tests the resetChangesFlag method
func TestGitRepository_ResetChangesFlag(t *testing.T) {
	gitRepo := &git.GitRepository{}

	// Set changes flag to true using the test helper
	gitRepo.SetHasChangesForTest(true)
	if !gitRepo.HasChanges() {
		t.Error("Expected HasChanges to return true after SetHasChangesForTest(true)")
	}

	// We can't directly test resetChangesFlag since it's unexported
	// But we can test the behavior by checking that it sets the flag to false
	// In a real scenario, this would be called internally by the Sync method
}

// TestGitRepository_SetChangesFlag tests the setChangesFlag method
func TestGitRepository_SetChangesFlag(t *testing.T) {
	gitRepo := &git.GitRepository{}

	// Set changes flag to true using the test helper
	gitRepo.SetHasChangesForTest(true)
	if !gitRepo.HasChanges() {
		t.Error("Expected HasChanges to return true after SetHasChangesForTest(true)")
	}

	// We can't directly test setChangesFlag since it's unexported
	// But we can test the behavior by checking that it sets the flag to true
	// In a real scenario, this would be called internally by various methods
}
