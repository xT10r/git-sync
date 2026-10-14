// Package snapshot prepares immutable Git snapshots and atomically publishes current.
package snapshot

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHTTP "github.com/go-git/go-git/v5/plumbing/transport/http"
)

type Config struct {
	URL      string
	Branch   string
	Root     string
	Username string
	Token    string
	AuthMode string
}

type Result struct {
	SHA         string    `json:"sha"`
	Path        string    `json:"path"`
	Published   bool      `json:"published"`
	PublishedAt time.Time `json:"published_at"`
}

// OperationError preserves the cause while exposing a safe, bounded operation label.
type OperationError struct {
	Operation string
	Cause     error
}

func (e *OperationError) Error() string { return e.Operation + ": " + e.Cause.Error() }
func (e *OperationError) Unwrap() error { return e.Cause }

type identity struct {
	Version int    `json:"version"`
	URL     string `json:"url"`
	Branch  string `json:"branch"`
}

// Provider owns root until Close. Consumers must never write inside root.
// A single mutex serializes publication, recovery, and Close; no network IO runs in New.
type Provider struct {
	mu     sync.Mutex
	cfg    Config
	lock   *os.File
	repo   *git.Repository
	last   Result
	closed bool
	// rename is per-instance to allow deterministic publication-failure tests.
	rename func(string, string) error
}

func New(cfg Config) (*Provider, error) {
	if cfg.URL == "" || cfg.Root == "" || cfg.Branch == "" {
		return nil, errors.New("repository URL, branch and snapshot root are required")
	}
	if err := plumbing.NewBranchReferenceName(cfg.Branch).Validate(); err != nil {
		return nil, fmt.Errorf("invalid branch: %w", err)
	}
	if u, err := url.Parse(cfg.URL); err == nil && u.User != nil {
		return nil, errors.New("credentials in repository URL are forbidden; use repo-user/repo-token")
	}
	if cfg.AuthMode == "" {
		cfg.AuthMode = "basic"
	}
	if cfg.AuthMode != "basic" && cfg.AuthMode != "bearer" {
		return nil, errors.New("repo-auth-mode must be basic or bearer")
	}
	root, err := filepath.Abs(cfg.Root)
	if err != nil {
		return nil, err
	}
	cfg.Root = root
	if root == filepath.VolumeName(root)+string(os.PathSeparator) {
		return nil, errors.New("filesystem root cannot be a snapshot root")
	}
	if st, err := os.Lstat(root); err == nil && (!st.IsDir() || st.Mode()&os.ModeSymlink != 0) {
		return nil, errors.New("snapshot root must be a real directory")
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err = os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	// Refuse legacy checkouts and unrelated content before creating service data.
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	marker := filepath.Join(root, ".git-sync.json")
	markerInfo, markerErr := os.Lstat(marker)
	if markerErr != nil && !os.IsNotExist(markerErr) {
		return nil, markerErr
	}
	if markerErr == nil && (!markerInfo.Mode().IsRegular()) {
		return nil, errors.New("invalid ownership marker")
	}
	if os.IsNotExist(markerErr) {
		for _, e := range entries {
			if e.Name() != ".git-sync.lock" {
				return nil, errors.New("snapshot root is not empty or owned by git-sync; use a new root")
			}
		}
	}
	lockPath := filepath.Join(root, ".git-sync.lock")
	if st, e := os.Lstat(lockPath); e == nil && !st.Mode().IsRegular() {
		return nil, errors.New("invalid root lock")
	} else if e != nil && !os.IsNotExist(e) {
		return nil, e
	}
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = lockFile(lock); err != nil {
		lock.Close()
		return nil, fmt.Errorf("snapshot root is locked or locking unsupported: %w", err)
	}
	p := &Provider{cfg: cfg, lock: lock, rename: os.Rename}
	ok := false
	defer func() {
		if !ok {
			p.Close()
		}
	}()
	want := identity{Version: 1, URL: cfg.URL, Branch: cfg.Branch}
	if data, e := os.ReadFile(marker); e == nil {
		var got identity
		if e = json.Unmarshal(data, &got); e != nil || got != want {
			return nil, errors.New("snapshot root belongs to a different repository/branch or format")
		}
	} else if os.IsNotExist(e) {
		data, _ := json.Marshal(want)
		if err = os.WriteFile(marker, data, 0600); err != nil {
			return nil, err
		}
	} else {
		return nil, e
	}
	for _, name := range []string{"cache", "snapshots"} {
		path := filepath.Join(root, name)
		if st, e := os.Lstat(path); e == nil && (!st.IsDir() || st.Mode()&os.ModeSymlink != 0) {
			return nil, fmt.Errorf("invalid managed directory %s", name)
		} else if e != nil && !os.IsNotExist(e) {
			return nil, e
		}
	}
	if err = os.MkdirAll(filepath.Join(root, "snapshots"), 0755); err != nil {
		return nil, err
	}
	cache := filepath.Join(root, "cache")
	p.repo, err = git.PlainOpen(cache)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		p.repo, err = git.PlainInit(cache, true)
		if err == nil {
			_, err = p.repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{cfg.URL}})
		}
	}
	if err != nil {
		return nil, fmt.Errorf("open Git cache: %w", err)
	}
	remote, err := p.repo.Remote("origin")
	if errors.Is(err, git.ErrRemoteNotFound) {
		// Recover an initialization interrupted between PlainInit and CreateRemote.
		remote, err = p.repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{cfg.URL}})
	}
	if err != nil {
		return nil, err
	}
	if len(remote.Config().URLs) != 1 || remote.Config().URLs[0] != cfg.URL {
		return nil, errors.New("Git cache origin differs from configured repository")
	}
	if err = p.recover(); err != nil {
		return nil, err
	}
	ok = true
	return p, nil
}

// recover trusts only fully renamed SHA directories; interrupted prepares are unreferenced.
func (p *Provider) recover() error {
	current := filepath.Join(p.cfg.Root, "current")
	target, err := os.Readlink(current)
	if err == nil {
		// Targets are relative and restricted to a single SHA beneath snapshots.
		sha := filepath.Base(target)
		if !plumbing.IsHash(sha) || target != filepath.Join("snapshots", sha) {
			return errors.New("invalid current snapshot target")
		}
		info, e := os.Lstat(filepath.Join(p.cfg.Root, target))
		if e != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("current snapshot is missing or invalid")
		}
		st, e := os.Lstat(current)
		if e != nil {
			return e
		}
		p.last = Result{SHA: sha, Path: filepath.Join(p.cfg.Root, target), PublishedAt: st.ModTime()}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read current: %w", err)
	}
	entries, err := os.ReadDir(filepath.Join(p.cfg.Root, "snapshots"))
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".prepare-") {
			if err = os.RemoveAll(filepath.Join(p.cfg.Root, "snapshots", e.Name())); err != nil {
				return err
			}
		}
	}
	// A process may have stopped after creating the new pointer but before rename.
	if err = os.Remove(filepath.Join(p.cfg.Root, ".current-next")); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (p *Provider) auth() transport.AuthMethod {
	if p.cfg.Token == "" {
		return nil
	}
	if p.cfg.AuthMode == "bearer" {
		return &gitHTTP.TokenAuth{Token: p.cfg.Token}
	}
	user := p.cfg.Username
	if user == "" {
		user = "git"
	}
	return &gitHTTP.BasicAuth{Username: user, Password: p.cfg.Token}
}

func (p *Provider) Current() Result { p.mu.Lock(); defer p.mu.Unlock(); return p.last }

func (p *Provider) Sync(ctx context.Context) (Result, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return Result{}, errors.New("snapshot provider is closed")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	ref := plumbing.NewBranchReferenceName(p.cfg.Branch)
	remoteRef := plumbing.ReferenceName("refs/remotes/origin/" + p.cfg.Branch)
	err := p.repo.FetchContext(ctx, &git.FetchOptions{RemoteName: "origin", RefSpecs: []config.RefSpec{config.RefSpec("+" + ref.String() + ":" + remoteRef.String())}, Auth: p.auth(), Tags: git.NoTags, Force: true})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return Result{}, &OperationError{Operation: "fetch", Cause: err}
	}
	target, err := p.repo.Reference(remoteRef, true)
	if err != nil {
		return Result{}, fmt.Errorf("resolve branch: %w", err)
	}
	sha := target.Hash().String()
	if p.last.SHA == sha {
		r := p.last
		r.Published = false
		return r, nil
	}
	final := filepath.Join(p.cfg.Root, "snapshots", sha)
	st, err := os.Lstat(final)
	if os.IsNotExist(err) {
		temp, e := os.MkdirTemp(filepath.Join(p.cfg.Root, "snapshots"), ".prepare-")
		if e != nil {
			return Result{}, e
		}
		defer os.RemoveAll(temp)
		commit, e := p.repo.CommitObject(target.Hash())
		if e != nil {
			return Result{}, e
		}
		tree, e := commit.Tree()
		if e != nil {
			return Result{}, e
		}
		if e = exportTree(ctx, tree, temp); e != nil {
			return Result{}, &OperationError{Operation: "prepare", Cause: e}
		}
		// MkdirTemp is private while preparing; readers can traverse after completion.
		if e = os.Chmod(temp, 0755); e != nil {
			return Result{}, e
		}
		if e = os.Rename(temp, final); e != nil {
			return Result{}, fmt.Errorf("finish snapshot: %w", e)
		}
	} else if err != nil {
		return Result{}, err
	} else if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return Result{}, errors.New("invalid snapshot directory")
	}
	if err = ctx.Err(); err != nil {
		return Result{}, err
	}
	next := filepath.Join(p.cfg.Root, ".current-next")
	if err = os.Symlink(filepath.Join("snapshots", sha), next); err != nil {
		return Result{}, &OperationError{Operation: "publish", Cause: err}
	}
	defer os.Remove(next)
	if err = p.rename(next, filepath.Join(p.cfg.Root, "current")); err != nil {
		return Result{}, &OperationError{Operation: "publish", Cause: err}
	}
	p.last = Result{SHA: sha, Path: final, Published: true, PublishedAt: time.Now().UTC()}
	return p.last, nil
}

func exportTree(ctx context.Context, tree *object.Tree, root string) error {
	walker := object.NewTreeWalker(tree, true, nil)
	defer walker.Close()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		name, entry, err := walker.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		clean := filepath.FromSlash(name)
		if strings.Contains(name, "\\") || filepath.IsAbs(clean) || filepath.Clean(clean) != clean || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
			return errors.New("unsafe Git path")
		}
		if entry.Mode == filemode.Dir {
			continue
		}
		if entry.Mode != filemode.Regular && entry.Mode != filemode.Executable {
			return fmt.Errorf("unsupported Git file mode for %s: %s (symlinks and submodules are not supported)", name, entry.Mode)
		}
		file, err := tree.TreeEntryFile(&entry)
		if err != nil {
			return err
		}
		reader, err := file.Reader()
		if err != nil {
			return err
		}
		err = writeFile(ctx, reader, filepath.Join(root, clean), entry.Mode)
		reader.Close()
		if err != nil {
			return fmt.Errorf("prepare %s: %w", name, err)
		}
	}
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(b []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(b)
}
func writeFile(ctx context.Context, r io.Reader, path string, mode filemode.FileMode) error {
	reader := bufio.NewReader(r)
	prefix, _ := reader.Peek(128)
	if strings.HasPrefix(string(prefix), "version https://git-lfs.github.com/spec/v1") {
		return errors.New("Git LFS objects are not supported")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	perm := os.FileMode(0644)
	if mode == filemode.Executable {
		perm = 0755
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, contextReader{ctx, reader})
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	return p.lock.Close()
}
