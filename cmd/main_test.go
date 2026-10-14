package main

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git-sync/internal/testutil"
)

func emptyEnv(string) string { return "" }
func TestExitCodes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want int
	}{
		{[]string{"--help"}, 0}, {nil, 2}, {[]string{"--unknown"}, 2},
		{[]string{"--repo-url=https://example.invalid/repo", "--local-path=" + t.TempDir()}, 0},
	} {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		var output bytes.Buffer
		code := run(ctx, tc.args, emptyEnv, &output)
		if code != tc.want {
			t.Fatalf("args %v code=%d, want %d: %s", tc.args, code, tc.want, output.String())
		}
	}
}
func TestRunPublishesAndStops(t *testing.T) {
	remote := testutil.NewRemote(t)
	remote.Commit("config", "v1")
	root := filepath.Join(t.TempDir(), "root")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan int, 1)
	go func() {
		done <- run(ctx, []string{"--repo-url=" + remote.Path, "--repo-branch=master", "--local-path=" + root, "--sync-interval=1h"}, emptyEnv, io.Discard)
	}()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	for {
		if data, err := os.ReadFile(filepath.Join(root, "current", "config")); err == nil && string(data) == "v1" {
			break
		}
		select {
		case code := <-done:
			t.Fatalf("early exit %d", code)
		case <-timer.C:
			t.Fatal("snapshot not published")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatal(code)
		}
	case <-time.After(time.Second):
		t.Fatal("process did not stop")
	}
}
func TestBindError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	remote := testutil.NewRemote(t)
	remote.Commit("config", "v1")
	code := run(context.Background(), []string{"--repo-url=" + remote.Path, "--repo-branch=master", "--local-path=" + t.TempDir(), "--http-server-addr=" + listener.Addr().String()}, emptyEnv, io.Discard)
	if code != 1 {
		t.Fatal("bind failure exit:", code)
	}
}
