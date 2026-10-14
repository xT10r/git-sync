package snapshot

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/object"
	stdHTTP "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"git-sync/internal/testutil"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func provider(t *testing.T, r *testutil.Remote, branch string) (*Provider, Config) {
	t.Helper()
	cfg := Config{URL: r.Path, Branch: branch, Root: filepath.Join(t.TempDir(), "root")}
	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { p.Close() })
	return p, cfg
}
func read(t *testing.T, path string) string {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}
func TestSnapshotLifecycle(t *testing.T) {
	remote := testutil.NewRemote(t)
	remote.Commit("dir/config", "main")
	remote.Branch("release")
	first := remote.Commit("dir/config", "release1")
	remote.Branch("master") // default branch must not decide the target
	p, cfg := provider(t, remote, "release")
	r, err := p.Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if r.SHA != first.String() || !r.Published {
		t.Fatalf("initial result: %+v", r)
	}
	if got := read(t, filepath.Join(cfg.Root, "current", "dir/config")); got != "release1" {
		t.Fatal(got)
	}
	pinned := r.Path
	noOp, err := p.Sync(context.Background())
	if err != nil || noOp.Published || noOp.SHA != r.SHA {
		t.Fatalf("no-op: %+v %v", noOp, err)
	}
	remote.Branch("release")
	second := remote.Commit("dir/config", "release2")
	r, err = p.Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if r.SHA != second.String() || !r.Published {
		t.Fatalf("new revision: %+v", r)
	}
	if got := read(t, filepath.Join(pinned, "dir/config")); got != "release1" {
		t.Fatal("old snapshot mutated:", got)
	}
	empty := remote.Commit("", "")
	r, err = p.Sync(context.Background())
	if err != nil || r.SHA != empty.String() || !r.Published {
		t.Fatalf("same tree new commit: %+v %v", r, err)
	}
	if err = p.Close(); err != nil {
		t.Fatal(err)
	}
	p2, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer p2.Close()
	if p2.Current().SHA != empty.String() {
		t.Fatal("published SHA not recovered")
	}
	r, err = p2.Sync(context.Background())
	if err != nil || r.Published {
		t.Fatalf("restart no-op: %+v %v", r, err)
	}
}
func TestForcePushAndMissingBranch(t *testing.T) {
	remote := testutil.NewRemote(t)
	first := remote.Commit("file", "one")
	p, cfg := provider(t, remote, "master")
	if _, err := p.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	remote.Commit("file", "two")
	if _, err := p.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	wt, _ := remote.Repo.Worktree()
	if err := wt.Reset(&git.ResetOptions{Commit: first, Mode: git.HardReset}); err != nil {
		t.Fatal(err)
	}
	rewritten := remote.Commit("file", "rewritten")
	result, err := p.Sync(context.Background())
	if err != nil || result.SHA != rewritten.String() {
		t.Fatalf("force-push: %+v %v", result, err)
	}
	if err = remote.Repo.Storer.RemoveReference(plumbing.NewBranchReferenceName("master")); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Sync(context.Background()); err == nil {
		t.Fatal("deleted branch used stale remote ref")
	}
	if got := read(t, filepath.Join(cfg.Root, "current", "file")); got != "rewritten" {
		t.Fatal("last good lost:", got)
	}
}
func TestFailurePreservesPublication(t *testing.T) {
	remote := testutil.NewRemote(t)
	first := remote.Commit("file", "one")
	p, cfg := provider(t, remote, "master")
	if _, err := p.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	remote.Commit("file", "two")
	failure := errors.New("injected rename failure")
	p.rename = func(string, string) error { return failure }
	if _, err := p.Sync(context.Background()); !errors.Is(err, failure) {
		t.Fatalf("publish error lost: %v", err)
	}
	if p.Current().SHA != first.String() || read(t, filepath.Join(cfg.Root, "current", "file")) != "one" {
		t.Fatal("last good changed")
	}
	p.rename = os.Rename
	if _, err := p.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.Sync(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := os.Rename(remote.Path, remote.Path+"-offline"); err != nil {
		t.Fatal(err)
	}
	defer os.Rename(remote.Path+"-offline", remote.Path)
	if _, err := p.Sync(context.Background()); err == nil {
		t.Fatal("expected fetch failure")
	}
	if got := read(t, filepath.Join(cfg.Root, "current", "file")); got != "two" {
		t.Fatal(got)
	}
}
func TestOwnedRootAndLock(t *testing.T) {
	remote := testutil.NewRemote(t)
	remote.Commit("file", "one")
	p, cfg := provider(t, remote, "master")
	if _, err := New(cfg); err == nil {
		t.Fatal("second writer acquired root")
	}
	p.Close()
	cfg.Branch = "other"
	if _, err := New(cfg); err == nil {
		t.Fatal("different branch reused root")
	}
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "unrelated"), []byte("keep"), 0600)
	cfg.Root = root
	if _, err := New(cfg); err == nil {
		t.Fatal("unowned root accepted")
	}
	if read(t, filepath.Join(root, "unrelated")) != "keep" {
		t.Fatal("foreign data changed")
	}
	cfg.Root = filepath.Join(t.TempDir(), "linked")
	if err := os.Symlink(root, cfg.Root); err != nil {
		t.Fatal(err)
	}
	if _, err := New(cfg); err == nil {
		t.Fatal("symlink root accepted")
	}
}
func TestRecoveryAndUnsupportedFiles(t *testing.T) {
	remote := testutil.NewRemote(t)
	first := remote.Commit("file", "one")
	p, cfg := provider(t, remote, "master")
	if _, err := p.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	remote.Commit("pointer", "version https://git-lfs.github.com/spec/v1\noid sha256:abc\nsize 10\n")
	if _, err := p.Sync(context.Background()); err == nil || !strings.Contains(err.Error(), "LFS") {
		t.Fatalf("expected LFS rejection: %v", err)
	}
	if p.Current().SHA != first.String() {
		t.Fatal("bad snapshot published")
	}
	p.Close()
	partial := filepath.Join(cfg.Root, "snapshots", ".prepare-interrupted")
	os.MkdirAll(partial, 0700)
	os.WriteFile(filepath.Join(partial, "partial"), []byte("partial"), 0600)
	os.Symlink(filepath.Join("snapshots", first.String()), filepath.Join(cfg.Root, ".current-next"))
	p2, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer p2.Close()
	if _, err = os.Stat(partial); !os.IsNotExist(err) {
		t.Fatalf("partial not cleaned: %v", err)
	}
	if p2.Current().SHA != first.String() {
		t.Fatal("recovery lost current")
	}
}
func TestAuth(t *testing.T) {
	p := &Provider{cfg: Config{Token: "secret", Username: "user", AuthMode: "basic"}}
	a, ok := p.auth().(*http.BasicAuth)
	if !ok || a.Username != "user" || a.Password != "secret" {
		t.Fatal("bad Basic config")
	}
	p.cfg.AuthMode = "bearer"
	if _, ok = p.auth().(*http.TokenAuth); !ok {
		t.Fatal("bad Bearer config")
	}
	p.cfg.Token = ""
	if p.auth() != nil {
		t.Fatal("anonymous auth must be nil")
	}
}
func TestConcurrentReads(t *testing.T) {
	remote := testutil.NewRemote(t)
	remote.Commit("file", "one")
	p, _ := provider(t, remote, "master")
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				p.Current()
			}
		}()
	}
	if _, err := p.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}

func TestRejectSymlinkAndGitlink(t *testing.T) {
	for _, kind := range []string{"symlink", "submodule"} {
		t.Run(kind, func(t *testing.T) {
			remote := testutil.NewRemote(t)
			first := remote.Commit("file", "good")
			p, _ := provider(t, remote, "master")
			if _, err := p.Sync(context.Background()); err != nil {
				t.Fatal(err)
			}
			wt, _ := remote.Repo.Worktree()
			if kind == "symlink" {
				if err := os.Symlink("file", filepath.Join(remote.Path, "link")); err != nil {
					t.Fatal(err)
				}
				if _, err := wt.Add("link"); err != nil {
					t.Fatal(err)
				}
			} else {
				idx, err := remote.Repo.Storer.Index()
				if err != nil {
					t.Fatal(err)
				}
				idx.Entries = append(idx.Entries, &index.Entry{Name: "module", Hash: first, Mode: filemode.Submodule})
				if err = remote.Repo.Storer.SetIndex(idx); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := wt.Commit("unsupported fixture", &git.CommitOptions{Author: fixtureAuthor()}); err != nil {
				t.Fatal(err)
			}
			if _, err := p.Sync(context.Background()); err == nil {
				t.Fatal("unsupported entry published")
			}
			if p.Current().SHA != first.String() {
				t.Fatal("last-good changed")
			}
		})
	}
}
func TestSnapshotReadConsistency(t *testing.T) {
	remote := testutil.NewRemote(t)
	remote.Commit("a", "0")
	remote.Commit("b", "0")
	p, cfg := provider(t, remote, "master")
	if _, err := p.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		for ctx.Err() == nil {
			target, err := os.Readlink(filepath.Join(cfg.Root, "current"))
			if err != nil {
				done <- err
				return
			}
			a, err := os.ReadFile(filepath.Join(cfg.Root, target, "a"))
			if err != nil {
				done <- err
				return
			}
			b, err := os.ReadFile(filepath.Join(cfg.Root, target, "b"))
			if err != nil {
				done <- err
				return
			}
			if string(a) != string(b) {
				done <- errors.New("mixed snapshot observed")
				return
			}
		}
		done <- nil
	}()
	for i := 1; i <= 10; i++ {
		value := strconv.Itoa(i)
		remote.Commit("a", value)
		remote.Commit("b", value)
		if _, err := p.Sync(context.Background()); err != nil {
			cancel()
			<-done
			t.Fatal(err)
		}
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
func TestHTTPSCredentialsReachTransport(t *testing.T) {
	for _, mode := range []string{"anonymous", "basic", "bearer"} {
		t.Run(mode, func(t *testing.T) {
			headers := make(chan string, 10)
			server := httptest.NewServer(stdHTTP.HandlerFunc(func(w stdHTTP.ResponseWriter, r *stdHTTP.Request) {
				select {
				case headers <- r.Header.Get("Authorization"):
				default:
				}
				w.WriteHeader(stdHTTP.StatusUnauthorized)
			}))
			defer server.Close()
			cfg := Config{URL: server.URL + "/repo", Branch: "main", Root: t.TempDir()}
			expected := ""
			if mode != "anonymous" {
				cfg.AuthMode = mode
				cfg.Username = "user"
				cfg.Token = "secret"
				if mode == "basic" {
					expected = "Basic " + base64.StdEncoding.EncodeToString([]byte("user:secret"))
				} else {
					expected = "Bearer secret"
				}
			}
			p, err := New(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer p.Close()
			if _, err = p.Sync(context.Background()); err == nil {
				t.Fatal("auth rejection must fail")
			}
			select {
			case got := <-headers:
				if got != expected {
					t.Fatalf("Authorization=%q, want %q", got, expected)
				}
			default:
				t.Fatal("no transport request")
			}
		})
	}
}

func fixtureAuthor() *object.Signature {
	return &object.Signature{Name: "Fixture", Email: "fixture@example.invalid", When: time.Unix(1700000000, 0)}
}

func TestRecoverInterruptedCacheInitialization(t *testing.T) {
	remote := testutil.NewRemote(t)
	remote.Commit("file", "one")
	p, cfg := provider(t, remote, "master")
	if err := p.repo.DeleteRemote("origin"); err != nil {
		t.Fatal(err)
	}
	p.Close()
	recovered, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Close()
	if _, err = recovered.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
}
