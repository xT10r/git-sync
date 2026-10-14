package config

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	base := []string{"--repo-url=https://example.invalid/repo.git", "--local-path=/tmp/snapshots"}
	for _, tc := range []struct {
		name      string
		args      []string
		env       map[string]string
		wantError bool
	}{
		{name: "valid"},
		{name: "IPv6", args: []string{"--http-server-addr=[::1]:8080"}},
		{name: "bad env duration", env: map[string]string{"GITSYNC_INTERVAL": "oops"}, wantError: true},
		{name: "CLI overrides bad env duration", args: []string{"--sync-interval=5s"}, env: map[string]string{"GITSYNC_INTERVAL": "oops"}},
		{name: "zero timeout", args: []string{"--sync-timeout=0s"}, wantError: true},
		{name: "partial Basic", args: []string{"--http-auth-username=user"}, wantError: true},
		{name: "conflicting auth", args: []string{"--http-auth-username=user", "--http-auth-password=pass", "--http-auth-token=secret"}, wantError: true},
		{name: "unsupported SSH", args: []string{"--repo-url=ssh://example.invalid/repo"}, wantError: true},
		{name: "URL secrets", args: []string{"--repo-url=https://user:secret@example.invalid/repo"}, wantError: true},
		{name: "positional arg", args: []string{"extra"}, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			cfg, err := Parse(append(append([]string{}, base...), tc.args...), func(k string) string { return tc.env[k] }, &out)
			if (err != nil) != tc.wantError {
				t.Fatalf("error=%v, wantError=%v", err, tc.wantError)
			}
			if err == nil && cfg.Interval <= 0 {
				t.Fatal("invalid interval")
			}
		})
	}
}
func TestSecretsAndPrecedence(t *testing.T) {
	values := map[string]string{"GITSYNC_REPOSITORY_TOKEN": "do-not-print", "GITSYNC_REPOSITORY_URL": "https://user:hidden@example.invalid/repo", "GITSYNC_HTTP_AUTH_PASSWORD": "hidden", "GITSYNC_INTERVAL": "2s"}
	var out bytes.Buffer
	_, err := Parse([]string{"--help"}, func(k string) string { return values[k] }, &out)
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "do-not-print") || strings.Contains(out.String(), "hidden") {
		t.Fatal("help leaks secret")
	}
	cfg, err := Parse([]string{"--repo-url=https://example.invalid/r", "--local-path=/tmp/root", "--repo-token=override", "--http-auth-password=", "--sync-interval=3s"}, func(k string) string { return values[k] }, &out)
	if err != nil || cfg.Repository.Token != "override" || cfg.HTTPPassword != "" || cfg.Interval != 3*time.Second {
		t.Fatalf("config=%+v err=%v", cfg, err)
	}
}
