// Package config owns the CLI/environment boundary.
package config

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"git-sync/internal/snapshot"
)

type Config struct {
	Repository   snapshot.Config
	Interval     time.Duration
	Timeout      time.Duration
	HTTPAddr     string
	HTTPUsername string
	HTTPPassword string
	HTTPToken    string
}

// Parse applies CLI > environment > defaults, including duration overrides.
func Parse(args []string, env func(string) string, out io.Writer) (Config, error) {
	var c Config
	get := func(key, def string) string {
		if v := env(key); v != "" {
			return v
		}
		return def
	}
	fs := flag.NewFlagSet("git-sync", flag.ContinueOnError)
	fs.SetOutput(out)
	fs.StringVar(&c.Repository.URL, "repo-url", "", "Repository URL or absolute local repository path")
	fs.StringVar(&c.Repository.Branch, "repo-branch", get("GITSYNC_REPOSITORY_BRANCH", "main"), "Branch to publish")
	fs.StringVar(&c.Repository.Root, "local-path", get("GITSYNC_LOCAL_PATH", ""), "Owned snapshot root; consumers read current or pin its target")
	fs.StringVar(&c.Repository.Username, "repo-user", get("GITSYNC_REPOSITORY_USER", ""), "HTTPS Basic username")
	// Secret defaults are deliberately blank in help; env is applied after parsing.
	fs.StringVar(&c.Repository.Token, "repo-token", "", "Repository credential (prefer GITSYNC_REPOSITORY_TOKEN)")
	fs.StringVar(&c.Repository.AuthMode, "repo-auth-mode", get("GITSYNC_REPOSITORY_AUTH_MODE", "basic"), "basic or bearer")
	interval := fs.String("sync-interval", get("GITSYNC_INTERVAL", "30s"), "Delay after each reconciliation")
	timeout := fs.String("sync-timeout", get("GITSYNC_TIMEOUT", "2m"), "Maximum duration of a reconciliation")
	fs.StringVar(&c.HTTPAddr, "http-server-addr", get("GITSYNC_HTTP_SERVER_ADDR", ""), "HTTP bind address; empty disables HTTP")
	fs.StringVar(&c.HTTPUsername, "http-auth-username", get("GITSYNC_HTTP_AUTH_USERNAME", ""), "HTTP Basic username")
	fs.StringVar(&c.HTTPPassword, "http-auth-password", "", "HTTP Basic password (prefer environment)")
	fs.StringVar(&c.HTTPToken, "http-auth-token", "", "HTTP Bearer token (prefer environment)")
	if err := fs.Parse(args); err != nil {
		return c, err
	}
	if fs.NArg() != 0 {
		return c, fmt.Errorf("unexpected positional arguments")
	}
	explicit := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { explicit[f.Name] = true })
	if !explicit["repo-url"] {
		c.Repository.URL = env("GITSYNC_REPOSITORY_URL")
	}
	if !explicit["repo-token"] {
		c.Repository.Token = env("GITSYNC_REPOSITORY_TOKEN")
	}
	if !explicit["http-auth-password"] {
		c.HTTPPassword = env("GITSYNC_HTTP_AUTH_PASSWORD")
	}
	if !explicit["http-auth-token"] {
		c.HTTPToken = env("GITSYNC_HTTP_AUTH_TOKEN")
	}
	var err error
	c.Interval, err = time.ParseDuration(*interval)
	if err != nil || c.Interval <= 0 {
		return c, fmt.Errorf("sync-interval must be a positive duration")
	}
	c.Timeout, err = time.ParseDuration(*timeout)
	if err != nil || c.Timeout <= 0 {
		return c, fmt.Errorf("sync-timeout must be a positive duration")
	}
	if c.Repository.URL == "" || c.Repository.Root == "" || c.Repository.Branch == "" {
		return c, fmt.Errorf("repo-url, repo-branch and local-path are required")
	}
	u, err := url.Parse(c.Repository.URL)
	if err != nil {
		return c, fmt.Errorf("invalid repository URL")
	}
	if u.User != nil {
		return c, fmt.Errorf("credentials in repository URL are forbidden")
	}
	switch u.Scheme {
	case "http", "https":
		if u.Host == "" || u.RawQuery != "" || u.Fragment != "" {
			return c, fmt.Errorf("invalid HTTP repository URL")
		}
	case "file":
		if u.Host != "" || !filepath.IsAbs(u.Path) {
			return c, fmt.Errorf("file URL must identify a local absolute path")
		}
	case "":
		if !filepath.IsAbs(c.Repository.URL) {
			return c, fmt.Errorf("local repository path must be absolute")
		}
	default:
		return c, fmt.Errorf("unsupported repository transport; use HTTP(S) or an absolute local path")
	}
	if c.Repository.AuthMode != "basic" && c.Repository.AuthMode != "bearer" {
		return c, fmt.Errorf("repo-auth-mode must be basic or bearer")
	}
	if c.Repository.Username != "" && c.Repository.Token == "" {
		return c, fmt.Errorf("repo-user requires repo-token")
	}
	if c.HTTPAddr != "" {
		_, port, err := net.SplitHostPort(c.HTTPAddr)
		if err != nil {
			return c, fmt.Errorf("http-server-addr must be host:port")
		}
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return c, fmt.Errorf("invalid HTTP port")
		}
	}
	if (c.HTTPUsername == "") != (c.HTTPPassword == "") {
		return c, fmt.Errorf("both HTTP Basic username and password are required")
	}
	if c.HTTPUsername != "" && c.HTTPToken != "" {
		return c, fmt.Errorf("choose HTTP Basic or Bearer authentication")
	}
	return c, nil
}
