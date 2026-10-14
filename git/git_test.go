package git_test

import (
	"git-sync/git"
	"git-sync/internal/constants"
	"git-sync/internal/testutil"
	"git-sync/mock"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGitRepositoryWithValidURL(t *testing.T) {
	remote := testutil.NewRemote(t)
	expected := remote.Commit("config.txt", "v1")
	fs := mock.Flags()
	fs.String(constants.FlagRepoUrl, remote.Path, "")
	fs.String(constants.FlagLocalPath, filepath.Join(t.TempDir(), "checkout"), "")
	repo, err := git.NewGitRepository(fs)
	if err != nil {
		t.Fatal(err)
	}
	commit, err := repo.Commit()
	if err != nil {
		t.Fatal(err)
	}
	if commit.Hash != expected.String() {
		t.Fatalf("commit = %s, want %s", commit.Hash, expected)
	}
}

func TestNewGitRepositoryWithInvalidURL(t *testing.T) {
	fs := mock.Flags()
	fs.String(constants.FlagRepoUrl, filepath.Join(t.TempDir(), "missing-remote"), "")
	fs.String(constants.FlagLocalPath, filepath.Join(t.TempDir(), "checkout"), "")
	_, err := git.NewGitRepository(fs)
	if err == nil || !strings.Contains(err.Error(), "failed to clone repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}
