package gitsync_test

import (
	"context"
	"git-sync/git"
	"git-sync/internal/constants"
	"git-sync/internal/gitsync"
	"git-sync/internal/handlers"
	"git-sync/mock"
	"testing"
	"time"
)

type observedRepo struct {
	mock.Gitter
	calls chan struct{}
}

func (r *observedRepo) Sync() error { r.calls <- struct{}{}; return nil }
func (r *observedRepo) Commit() (*git.CommitInfo, error) {
	return &git.CommitInfo{Hash: "fixture", Date: time.Unix(1700000000, 0)}, nil
}
func TestStart(t *testing.T) {
	for _, mode := range []string{"webhook", "periodic"} {
		t.Run(mode, func(t *testing.T) {
			handlers.WebhookCh = make(chan string, 1)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			fs := mock.Flags()
			if mode == "periodic" {
				if err := fs.Set(constants.FlagSyncInterval, "5ms"); err != nil {
					t.Fatal(err)
				}
			}
			service, err := gitsync.NewGitSync(fs, ctx)
			if err != nil {
				t.Fatal(err)
			}
			repo := &observedRepo{calls: make(chan struct{}, 100)}
			done := make(chan struct{})
			go func() { defer close(done); service.Start(repo) }()
			t.Cleanup(func() {
				cancel()
				select {
				case <-done:
				case <-time.After(2 * time.Second):
					t.Error("worker did not stop")
				}
			})
			if mode == "webhook" {
				handlers.WebhookCh <- "127.0.0.1"
			}
			select {
			case <-repo.calls:
			case <-time.After(2 * time.Second):
				t.Fatal("sync not invoked")
			}
			cancel()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("worker did not stop")
			}
		})
	}
}
