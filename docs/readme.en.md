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
- Debug logging support for troubleshooting.

## Configuration and Parameters

Service launch parameters can be set via command-line flags, environment variables, and a configuration file.

### Viewing Configuration Options

To view all available configuration options with their descriptions, environment variables, and command-line flags, use:

```
git-sync config-help
```

### Generating a Configuration File

To generate a sample configuration file, use the command:

```
git-sync gen-config
```

By default, a `config.yaml` file will be created in the current directory. To specify a different file path, use the `-output` flag:

```
git-sync gen-config -output /path/to/config.yaml
```

### Using a Configuration File

The service automatically looks for a configuration file in the following locations:
1. Current directory (`./config.yaml`)
2. Config directory (`./config/config.yaml`)
3. User's home directory (`~/.gitsync/config.yaml`)

Configuration file format (YAML):

```yaml
gitlab:
  repourl: "https://gitlab.example.com/username/repository.git"
  repobranch: "main"
  repoauth:
    user: "gitlab-user"
    token: "your-gitlab-token"

sync:
  local_path: "./repo"
  interval: 30

http_server:
  addr: "0.0.0.0:8080"
  auth:
    username: "admin"
    password: "password"
```

### Debug Logging

The service supports debug logging which can be enabled for troubleshooting purposes. Debug logging can be enabled using:

1. Command-line flag: `--debug`
2. Environment variable: `GITSYNC_DEBUG=true`
3. Configuration file: `debug: true`

When debug logging is enabled, additional debug messages will be output to help with troubleshooting.

Example with command-line flag:
```
git-sync --debug --repo-url=https://gitlab.example.com/username/repository.git --repo-branch=main --local-path=./repo
```

Example with environment variable:
```
GITSYNC_DEBUG=true git-sync --repo-url=https://gitlab.example.com/username/repository.git --repo-branch=main --local-path=./repo
```

Example with configuration file:
```yaml
debug: true
gitlab:
  repourl: "https://gitlab.example.com/username/repository.git"
  repobranch: "main"
  repoauth:
    user: "gitlab-user"
    token: "your-gitlab-token"

sync:
  local_path: "./repo"
  interval: 30

http_server:
  addr: "0.0.0.0:8080"
```

### Command-Line Parameters and Environment Variables

Each environment variable is mapped to the corresponding command-line parameter (flag).

Command-Line Parameters / Environment Variables

|Command Argument (flag)|Environment variable|Description|
|-|-|-|
|`--local-path`|`GITSYNC_LOCAL_PATH`|Path to the local repository.|
|`--repo-url`|`GITSYNC_REPOSITORY_URL`|URL of the remote repository.|
|`--repo-branch`|`GITSYNC_REPOSITORY_BRANCH`|Branch of the remote repository.|
|`--repo-auth-user`|`GITSYNC_REPOSITORY_USER`|User for repository authentication.|
|`--repo-auth-token`|`GITSYNC_REPOSITORY_TOKEN`|Token for repository authentication.|
|`--sync-interval`|`GITSYNC_INTERVAL`|Interval for repository synchronization.|
|`--http-server-addr`|`GITSYNC_HTTP_SERVER_ADDR`|Address and port of the HTTP server.|
|`--http-server-auth-username`|`GITSYNC_HTTP_SERVER_AUTH_USERNAME`|Username for HTTP server authentication.|
|`--http-server-auth-password`|`GITSYNC_HTTP_SERVER_AUTH_PASSWORD`|Password for HTTP server authentication.|
|`--http-server-auth-token`|`GITSYNC_HTTP_SERVER_AUTH_TOKEN`|Token for HTTP server authentication.|
|`--debug`|`GITSYNC_DEBUG`|Enable debug logging.|

### Configuration Precedence

The service uses the following precedence order for configuration values:
1. Command-line flags (highest precedence)
2. Environment variables
3. Configuration file
4. Default values (lowest precedence)

This means that command-line flags will override environment variables, which will override configuration file values, which will override default values.

### Prometheus Metrics

The service provides the following metrics:

|Name|Description|
|-|-|
|`git_sync_sync_count`|Total number of synchronizations with changes.|
|`git_sync_sync_total_count`|Total number of synchronizations.|
|`git_sync_sync_total_error_count`|Total number of synchronization errors.|
|`git_sync_repo_info`|Information about the synchronized repository with labels for `repository name` and `repository branch`.|
|`git_sync_commit_info`|Information about the latest commit with labels for `commit hash`, `author name`, `author email`, `commit date`, `commit message`.|

### Use Cases

<b>Application Configuration Files</b>: Ensuring a single source of truth for application configuration files that frequently change and need to be synchronized across different instances.

<b>Deployment Scripts</b>: Automatically updating and synchronizing deployment scripts across different servers or environments to ensure all servers use the same version of the scripts.

<b>Server Configuration Files</b>: Synchronizing server configuration files, such as web server or database configuration files, to ensure all servers are configured correctly and uniformly.

<b>Documentation and Instructions</b>: Synchronizing documentation and instructions for developers to ensure they always have access to the most up-to-date information.