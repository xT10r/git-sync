//go:build linux || darwin

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"git-sync/internal/testutil"
)

func TestProcessHelper(t *testing.T) {
	if os.Getenv("GITSYNC_PROCESS_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"git-sync"}, os.Args[i+1:]...)
			main()
			return
		}
	}
	os.Exit(99)
}
func TestProcessPublishesAndHandlesSIGTERM(t *testing.T) {
	remote := testutil.NewRemote(t)
	remote.Commit("config", "one")
	root := filepath.Join(t.TempDir(), "root")
	var output bytes.Buffer
	cmd := exec.Command(os.Args[0], "-test.run=^TestProcessHelper$", "--", "--repo-url="+remote.Path, "--repo-branch=master", "--local-path="+root, "--sync-interval=20ms")
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "GITSYNC_") {
			cmd.Env = append(cmd.Env, value)
		}
	}
	cmd.Env = append(cmd.Env, "GITSYNC_PROCESS_HELPER=1")
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	stopped := false
	t.Cleanup(func() {
		if !stopped {
			cmd.Process.Kill()
			<-done
		}
	})
	waitContent := func(expected string) {
		t.Helper()
		deadline := time.NewTimer(5 * time.Second)
		defer deadline.Stop()
		for {
			data, err := os.ReadFile(filepath.Join(root, "current", "config"))
			if err == nil && string(data) == expected {
				return
			}
			select {
			case err := <-done:
				stopped = true
				t.Fatalf("early process exit: %v: %s", err, output.String())
			case <-deadline.C:
				t.Fatal("process did not publish " + expected)
			default:
				time.Sleep(time.Millisecond)
			}
		}
	}
	waitContent("one")
	pinned, err := os.Readlink(filepath.Join(root, "current"))
	if err != nil {
		t.Fatal(err)
	}
	remote.Commit("config", "two")
	waitContent("two")
	if got, err := os.ReadFile(filepath.Join(root, pinned, "config")); err != nil || string(got) != "one" {
		t.Fatalf("old snapshot changed: %s %v", got, err)
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		stopped = true
		if err != nil {
			t.Fatalf("SIGTERM exit: %v: %s", err, output.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("SIGTERM did not stop process")
	}
}
