# Git-Sync snapshot provider

Git-Sync publishes immutable file snapshots from a configured Git branch. One process owns one repository root. Build with Go 1.26+; CI uses Go 1.26.9 and 1.27.2. See [the detailed contract and operations guide](readme.ru.md).

## Contract

The configured `--local-path` is service-owned data, not a developer checkout:

- `cache/`: bare Git object cache;
- `snapshots/<SHA>/`: complete, immutable snapshots;
- `current`: relative symlink to the published snapshot;
- `.git-sync.json` / `.git-sync.lock`: ownership identity and single-writer lock.

A new SHA is prepared separately, then published by replacing the symlink. Failed fetch/preparation/publication preserves the last good version. A new SHA with unchanged file content is still a new revision. Force-push is supported without merge/fast-forward semantics.

Readers must resolve `current` once per group of related reads and pin that snapshot path. Re-resolving the symlink for every file can mix versions. Consumers must not modify service-owned data; mount the volume read-only in consumers. Old snapshots are currently retained indefinitely. Monitor volume size; automatic retention/GC is not implemented.

Linux/macOS/FreeBSD publication requires a local filesystem with flock, symlinks and atomic symlink replacement. Runtime integration tests were run on Linux only. Windows cross-compiles but explicitly refuses startup. NFS/SMB semantics and power-loss durability are not guaranteed. Interrupted temporary preparation is cleaned on restart; the published pointer restores the available snapshot. Corrupt Git-cache repair and snapshot integrity verification after unauthorized writes are not implemented.

## Usage

```sh
go build -o git-sync ./cmd/
./git-sync --repo-url=https://git.example.org/team/config.git \
  --repo-branch=main --local-path=/srv/git-sync/config \
  --sync-interval=30s --sync-timeout=2m \
  --http-server-addr=127.0.0.1:8080
```

The root must be absent, empty or already owned by the same repository/branch. Existing developer checkouts and unrelated directories are refused without clearing data. Use a fresh root when migrating from v1.

| Flag | Environment | Default |
|---|---|---|
| `--repo-url` | `GITSYNC_REPOSITORY_URL` | required |
| `--repo-branch` | `GITSYNC_REPOSITORY_BRANCH` | `main` |
| `--local-path` | `GITSYNC_LOCAL_PATH` | required; `/git` in image |
| `--repo-user` | `GITSYNC_REPOSITORY_USER` | empty; Basic token falls back to `git` username |
| `--repo-token` | `GITSYNC_REPOSITORY_TOKEN` | empty |
| `--repo-auth-mode` | `GITSYNC_REPOSITORY_AUTH_MODE` | `basic`; or `bearer` |
| `--sync-interval` | `GITSYNC_INTERVAL` | `30s` after previous attempt |
| `--sync-timeout` | `GITSYNC_TIMEOUT` | `2m` |
| `--http-server-addr` | `GITSYNC_HTTP_SERVER_ADDR` | disabled; `0.0.0.0:8080` in image |
| `--http-auth-username` | `GITSYNC_HTTP_AUTH_USERNAME` | empty |
| `--http-auth-password` | `GITSYNC_HTTP_AUTH_PASSWORD` | empty |
| `--http-auth-token` | `GITSYNC_HTTP_AUTH_TOKEN` | empty |

CLI overrides environment, which overrides defaults. No config file is supported. Invalid durations and incomplete/conflicting HTTP credentials are rejected. Prefer runtime environment/Secret credentials; environment secrets are omitted from help. Credentials embedded in repository URLs are forbidden.

Transports: HTTP(S), absolute local paths and local `file:///` URLs. Select Basic or Bearer according to the Git server. SSH, LFS, symlinks and submodules are not supported; unsupported Git entries fail preparation without replacing current. Executable regular files preserve their executable mode.

The first reconciliation starts immediately. Later attempts run after the configured interval or an incoming trigger; bursts are coalesced. Failed attempts retry at the normal interval, without exponential backoff/jitter yet. A failed attempt keeps the process alive and marks status degraded. Configuration errors exit 2; fatal startup/HTTP errors exit 1; help/SIGTERM exit 0. Deadlines cancel network IO and are checked during export; arbitrary blocking filesystem operations may not be immediately interruptible.

## HTTP

| Endpoint | Method | Meaning |
|---|---|---|
| `/livez` | GET | process answers; no auth |
| `/readyz` | GET | 200 if a published snapshot exists, else 503; no auth |
| `/status` | GET | revision/path, readiness, degraded/syncing, timestamps and safe error stage |
| `/webhook` | POST | 202 accepted/coalesced, not sync completion; 503 when stopping |
| `/metrics` | GET | Prometheus |

Status/trigger/metrics use Basic or Bearer if configured; otherwise they are accessible on the bind address. No built-in TLS: use network isolation or a TLS proxy. The webhook is a manual trigger, not a GitHub/GitLab signature verifier. Ready denotes available data, not remote freshness. After restart `current` restores readiness, while last-success remains unset until a successful remote check.

Metrics: `git_sync_attempts_total{outcome}`, `git_sync_duration_seconds`, `git_sync_last_success_timestamp_seconds`, `git_sync_last_publication_timestamp_seconds`, `git_sync_ready`, and Go/process collectors. Outcomes: published, unchanged, error. SHA/email/message/URL are not metric labels. Update dashboards from v1 metrics.

## Container and migration

Build with `docker build -t git-sync:snapshots .`. The image contains a static binary plus CA certificates and runs as UID/GID 65532. Mount a writable owned root at `/git` and set `GITSYNC_LOCAL_PATH=/git/root`; consumer mounts should be read-only. No credential build args are used. Kubernetes volume permissions must permit UID/GID 65532; validate the storage class and avoid subPath mounts that pin a resolved version.

This is a breaking file/API/metrics change: deploy to a new root, verify readiness/SHA/content, switch readers to pinned snapshots, then stop the old process. Keep the old image/volume for manual rollback. Legacy public Go packages and unfinished daemon/API packages were removed.

## Checks

```sh
go vet ./...
go test -race -count=2 -coverprofile=coverage.out ./...
go test -run '^$' -bench BenchmarkUnchangedSnapshot -benchmem ./internal/snapshot
govulncheck ./...
```

Tests use local Git repositories and HTTP servers after module download. They cover branch selection, force-push, identical trees at different SHAs, no-op, failed publication/auth, recovery/locking, pinned-reader consistency, lifecycle/metrics and a real subprocess receiving SIGTERM.
