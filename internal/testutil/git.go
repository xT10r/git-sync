// Package testutil provides local Git fixtures; tests never contact external servers.
package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Remote struct {
	T    *testing.T
	Path string
	Repo *git.Repository
}

func NewRemote(t *testing.T) *Remote {
	t.Helper()
	path := t.TempDir()
	repo, err := git.PlainInit(path, false)
	if err != nil {
		t.Fatal(err)
	}
	return &Remote{t, path, repo}
}
func (r *Remote) Commit(name, contents string) plumbing.Hash {
	r.T.Helper()
	if name != "" {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(r.Path, name)), 0755); err != nil {
			r.T.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(r.Path, name), []byte(contents), 0644); err != nil {
			r.T.Fatal(err)
		}
	}
	wt, err := r.Repo.Worktree()
	if err != nil {
		r.T.Fatal(err)
	}
	if name != "" {
		if _, err = wt.Add(name); err != nil {
			r.T.Fatal(err)
		}
	}
	hash, err := wt.Commit("fixture commit", &git.CommitOptions{Author: &object.Signature{Name: "Fixture", Email: "fixture@example.invalid", When: time.Unix(1700000000, 0)}, AllowEmptyCommits: true})
	if err != nil {
		r.T.Fatal(err)
	}
	return hash
}
func (r *Remote) Branch(name string) {
	r.T.Helper()
	wt, err := r.Repo.Worktree()
	if err != nil {
		r.T.Fatal(err)
	}
	ref := plumbing.NewBranchReferenceName(name)
	_, err = r.Repo.Reference(ref, false)
	create := err == plumbing.ErrReferenceNotFound
	if err = wt.Checkout(&git.CheckoutOptions{Branch: ref, Create: create, Force: true}); err != nil {
		r.T.Fatal(err)
	}
}
