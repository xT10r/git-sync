# Git-Sync Service

## General Description

The `Git-Sync Service` provides synchronization of a remote repository with a local repository.
Any changes in the local repository trigger synchronization with the remote repository.
The necessity for synchronization is checked by comparing file hashes (which determines if a file has changed) and by comparing the file trees of the remote and local repositories.
The service provides access to its metrics using Prometheus and supports webhooks for manual synchronization.

## Service Features

- Synchronization of the remote repository with the local repository.
- Checking the necessity for synchronization based on comparing file hashes and the file trees of the remote and local repositories.
- Handling webhooks for manual synchronization.
- Access to metrics via Prometheus.

## Configuration and Parameters

Service launch parameters can be set via command-line flags, environment variables, and a configuration file.

### Command-Line Parameters and Environment Variables

Each environment variable is mapped to the corresponding command-line parameter (flag).

Command-Line Parameters / Environment Variables

`--local-path` / `GITSYNC_LOCAL_PATH`: Path to the local repository.
`--repo-url` / `GITSYNC_REPOSITORY_URL`: URL of the remote repository.
`--repo-branch` / `GITSYNC_REPOSITORY_BRANCH`: Branch of the remote repository.
`--repo-auth-user` / `GITSYNC_REPOSITORY_USER`: User for repository authentication.
`--repo-auth-token` / `GITSYNC_REPOSITORY_TOKEN`: Token for repository authentication.
`--sync-interval` / `GITSYNC_INTERVAL`: Interval for repository synchronization.
`--http-server-addr` / `GITSYNC_HTTP_SERVER_ADDR`: Address and port of the HTTP server.
`--http-server-auth-username` / `GITSYNC_HTTP_SERVER_AUTH_USERNAME`: Username for HTTP server authentication.
`--http-server-auth-password` / `GITSYNC_HTTP_SERVER_AUTH_PASSWORD`: Password for HTTP server authentication.
`--http-server-auth-token` / `GITSYNC_HTTP_SERVER_AUTH_TOKEN`: Bearer token for HTTP server authentication.

### Prometheus Metrics

The service provides the following metrics:

`git_sync_sync_count`: Total number of synchronizations with changes.
`git_sync_sync_total_count`: Total number of synchronizations.
`git_sync_sync_total_error_count`: Total number of synchronization errors.
`git_sync_repo_info`: Information about the synchronized repository with labels for `repository name` and `repository branch`.
`git_sync_commit_info`: Information about the latest commit with labels for `commit hash`, `author name`, `author email`, `commit date`, `commit message`.

### Use Cases

<b>Application Configuration Files</b>: Ensuring a single source of truth for application configuration files that frequently change and need to be synchronized across different instances.

<b>Deployment Scripts</b>: Automatically updating and synchronizing deployment scripts across different servers or environments to ensure all servers use the same version of the scripts.

<b>Server Configuration Files</b>: Synchronizing server configuration files, such as web server or database configuration files, to ensure all servers are configured correctly and uniformly.

<b>Documentation and Instructions</b>: Synchronizing documentation and instructions for developers to ensure they always have access to the most up-to-date information.
