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
- Version information via CLI flag and HTTP endpoint.

## Environment Setup

To work with the repository functionality, you need to set up your environment properly. Here are the instructions for different operating systems:

### Installing Go

Go is the main programming language used for this project. You need Go 1.23 or later to build the project locally.

#### Linux

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# CentOS/RHEL/Fedora
sudo yum install golang
# or on newer versions
sudo dnf install golang
```

#### macOS

```bash
# Using Homebrew
brew install go
```

#### Windows

Download and install Go from the official website:
https://golang.org/dl/

### Installing Make

#### Linux

On most Linux distributions, `make` is available in the default package manager:

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install make

# CentOS/RHEL/Fedora
sudo yum install make
# or on newer versions
sudo dnf install make
```

#### macOS

On macOS, you'll need to install Xcode Command Line Tools which includes `make`:

```bash
xcode-select --install
```

#### Windows

For Windows 10/11, there are several options to install `make`:

1. **Using Chocolatey** (recommended):

   ```powershell
   # First install Chocolatey if you haven't already
   # Then install make
   choco install make
   ```

2. **Using Winget**:

   ```powershell
   winget install ezwinports.make
   ```

3. **Using WSL** (Windows Subsystem for Linux):
   Install WSL and then follow the Linux instructions above.

### Installing Docker

Docker is required for building and running the service. Follow these instructions based on your operating system:

#### Linux

Follow the official Docker installation guide for your distribution:
https://docs.docker.com/engine/install/

Also, don't forget to install Docker Compose:
https://docs.docker.com/compose/install/

#### macOS

Download Docker Desktop for Mac:
https://docs.docker.com/docker-for-mac/install/

#### Windows

Download Docker Desktop for Windows:
https://docs.docker.com/docker-for-windows/install/

Note: Docker Desktop on Windows requires WSL 2 backend for best performance.

### Building the Project

Once you have `make` and Docker installed, you can build the project using:

```bash
# Build using Docker (recommended)
make build

# Build locally (for development)
make build-local

# Build Windows executable with .exe extension
make build-windows

# Build for multiple platforms
make build-all

# Run tests
make test

# Run tests with coverage report
make cover

# Clean build artifacts
make clean
```

For Windows users without `make`, you can use the PowerShell script:

```powershell
# Build Windows executable
.\scripts\build-windows.ps1
```

For more build options, run:

```bash
make help
```

## Configuration and Parameters

Service launch parameters can be set via command-line flags, environment variables, and a configuration file.

### Viewing Configuration Options

To view all available configuration options with their descriptions, environment variables, and command-line flags, use:

```bash
git-sync config-help
```

### Generating a Configuration File

To generate a sample configuration file, use the command:

```bash
git-sync gen-config
```

By default, a `config.yaml` file will be created in the current directory. To specify a different file path, use the `-output` flag:

```bash
git-sync gen-config -output /path/to/config.yaml
```

### Using a Configuration File

The service automatically looks for a configuration file in the following locations:

1. Current directory (`./config.yaml`)
2. Config directory (`./config/config.yaml`)
3. User's home directory (`~/.gitsync/config.yaml`)

Configuration file format (YAML):

```yaml
debug: false

# Default settings
defaults:
  gitlab:
    repoauth:
      user: gitlab-user
      token: your-default-gitlab-token  # Default token for all repositories
      token_file: ""  # Default token file path (optional)
    repobranch: main
  sync:
    interval: 30

http_server:
  addr: "0.0.0.0:8080"
  auth:
    username: "admin"
    password: "your-password"

repositories:
  # Main repository - uses most default settings
  main-repo:
    gitlab:
      repourl: "https://gitlab.example.com/username/repository1.git"
      # repoauth not specified - will use defaults
    sync:
      local_path: "./repo1"

  # Repository with custom interval and branch
  secondary-repo:
    gitlab:
      repoauth:
        token: "your-gitlab-token-2"  # Custom token
      repobranch: "develop"  # Override branch
      repourl: "https://gitlab.example.com/username/repository2.git"
    sync:
      interval: 60  # Override interval
      local_path: "./repo2"

  # Repository with different user, but using default token
  external-repo:
    gitlab:
      repoauth:
        user: "different-user"  # Override only user
        # token will be taken from defaults
      repourl: "https://gitlab.example.com/anotheruser/repository.git"
    sync:
      local_path: "./external-repo"

  # Repository with fully custom settings
  custom-repo:
    gitlab:
      repoauth:
        token: "your-gitlab-token-4"
        user: "custom-user"
      repobranch: "feature-branch"
      repourl: "https://gitlab.example.com/custom/repo.git"
    sync:
      interval: 300  # 5 minutes
      local_path: "./custom-repo"
```

### How Default Settings Work

The `defaults` section allows you to define common settings for all repositories. If certain parameters are not specified in a repository's configuration, they will be taken from the `defaults` section.

For example:

- If a repository doesn't specify `repoauth.token`, the token from `defaults.gitlab.repoauth.token` will be used
- If a repository doesn't specify `repoauth.user`, the user from `defaults.gitlab.repoauth.user` will be used
- If a repository doesn't specify `repoauth.token_file`, the token file from `defaults.gitlab.repoauth.token_file` will be used
- If a repository doesn't specify `repobranch`, the branch from `defaults.gitlab.repobranch` will be used
- If a repository doesn't specify `sync.interval`, the interval from `defaults.sync.interval` will be used

This allows you to avoid duplicating identical settings for multiple repositories that use the same credentials.

If you specify a `token_file` in the defaults, it will be used for all repositories that don't have their own `token_file` specified. The token will be read from the file and used for authentication.

### Secure Handling of Sensitive Credentials

For security reasons, sensitive credentials such as tokens and passwords should not be stored directly in configuration files or passed as environment variables in production environments. Here are recommended approaches for different environments:

#### Development Environment

For development environments, you can use direct values for convenience:

```bash
# Using environment variables (for development only)
GITSYNC_REPOSITORY_TOKEN=your-token \
GITSYNC_HTTP_AUTH_USERNAME=admin \
GITSYNC_HTTP_AUTH_PASSWORD=your-password \
git-sync --repo-url=https://gitlab.example.com/username/repository.git --repo-branch=main --local-path=./repo
```

Or in a configuration file (for development only):

```yaml
gitlab:
  repourl: "https://gitlab.example.com/username/repository.git"
  repobranch: "main"
  repoauth:
    user: "gitlab-user"
    token: "your-gitlab-token"

http_server:
  addr: "0.0.0.0:8080"
  auth:
    username: "admin"
    password: "your-password"
```

#### Production Environment

For production environments, use external secret management systems:

1. **Docker Secrets**:

   ```bash
   # Create secret files
   echo "your-gitlab-token" > gitlab-token
   echo "your-http-password" > http-password
   
   # Run with mounted secrets
   docker run -v $(pwd)/gitlab-token:/run/secrets/gitlab-token:ro \
              -v $(pwd)/http-password:/run/secrets/http-password:ro \
              -e GITSYNC_REPOSITORY_TOKEN_FILE=/run/secrets/gitlab-token \
              -e GITSYNC_HTTP_AUTH_PASSWORD_FILE=/run/secrets/http-password \
              git-sync
   ```

2. **Kubernetes Secrets**:

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: git-sync-secrets
   type: Opaque
   data:
     gitlab-token: eW91ci1naXRsYWItdG9rZW4=  # base64 encoded
     http-password: eW91ci1odHRwLXBhc3N3b3Jk  # base64 encoded
   ---
   apiVersion: v1
   kind: Pod
   metadata:
     name: git-sync
   spec:
     containers:
     - name: git-sync
       image: git-sync
       env:
       - name: GITSYNC_REPOSITORY_TOKEN_FILE
         value: /run/secrets/gitlab-token
       - name: GITSYNC_HTTP_AUTH_PASSWORD_FILE
         value: /run/secrets/http-password
       volumeMounts:
       - name: secrets
         mountPath: /run/secrets
         readOnly: true
     volumes:
     - name: secrets
       secret:
         secretName: git-sync-secrets
   ```

3. **Environment Variable Files**:

   ```bash
   # Create a secure env file
   echo "GITSYNC_REPOSITORY_TOKEN=your-token" > .env.secure
   echo "GITSYNC_HTTP_AUTH_PASSWORD=your-password" >> .env.secure
   chmod 600 .env.secure
   
   # Load and run
   source .env.secure
   git-sync --repo-url=https://gitlab.example.com/username/repository.git --repo-branch=main --local-path=./repo
   ```

### Logging

The service uses a unified logging approach throughout the codebase. Previously, logging was done using `logger.GetLogger().Info()`, `logger.GetLogger().Error()`, etc. This has been unified to use global logger functions like `logger.Info()`, `logger.Error()`, etc. for consistency and better maintainability.

#### Debug Logging

The service supports debug logging which can be enabled for troubleshooting purposes. Debug logging can be enabled using:

1. Command-line flag: `--debug`
2. Environment variable: `GITSYNC_DEBUG=true`
3. Configuration file: `debug: true`

When debug logging is enabled, additional debug messages will be output to help with troubleshooting.

Example with command-line flag:

```bash
git-sync --debug --repo-url=https://gitlab.example.com/username/repository.git --repo-branch=main --local_path=./repo
```

Example with environment variable:

```bash
GITSYNC_DEBUG=true git-sync --repo-url=https://gitlab.example.com/username/repository.git --repo-branch=main --local_path=./repo
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

### Version Information

The service includes version information that can be accessed in multiple ways:

1. **CLI Flag**: Use `--version` to display version information:

   ```bash
   git-sync --version
   ```

2. **HTTP Endpoint**: Access the `/version` endpoint to get detailed version information in JSON format:

   ```bash
   curl http://localhost:8080/version
   ```

3. **Prometheus Metrics**: The service exposes build information via the `git_sync_build_info` metric with labels for version, commit, date, and dirty status.

The version information includes:

- **Version**: The semantic version (e.g., v1.0.0)
- **Commit**: The git commit hash at build time
- **Date**: The build date in UTC
- **Dirty**: Indicates if there were uncommitted changes at build time

### Command-Line Parameters and Environment Variables

Each environment variable is mapped to the corresponding command-line parameter (flag).

Command-Line Parameters / Environment Variables

|Command Argument (flag)|Environment variable|Description|
|-|-|-|
|`--local-path`|`GITSYNC_LOCAL_PATH`|Path to the local repository.|
|`--repo-url`|`GITSYNC_REPOSITORY_URL`|URL of the remote repository.|
|`--repo-branch`|`GITSYNC_REPOSITORY_BRANCH`|Branch of the remote repository.|
|`--repo-user`|`GITSYNC_REPOSITORY_USER`|User for repository authentication.|
|`--repo-token`|`GITSYNC_REPOSITORY_TOKEN`|Token for repository authentication.|
|`--repo-token-file`|`GITSYNC_REPOSITORY_TOKEN_FILE`|Path to file containing repository authentication token.|
|`--sync-interval`|`GITSYNC_INTERVAL`|Interval for repository synchronization.|
|`--http-server-addr`|`GITSYNC_HTTP_SERVER_ADDR`|Address and port of the HTTP server.|
|`--http-auth-username`|`GITSYNC_HTTP_AUTH_USERNAME`|Username for HTTP server authentication.|
|`--http-auth-password`|`GITSYNC_HTTP_AUTH_PASSWORD`|Password for HTTP server authentication.|
|`--http-auth-token`|`GITSYNC_HTTP_AUTH_TOKEN`|Token for HTTP server authentication.|
|`--http-auth-token-file`|`GITSYNC_HTTP_AUTH_TOKEN_FILE`|Path to file containing HTTP server authentication token.|
|`--debug`|`GITSYNC_DEBUG`|Enable debug logging.|
|`--version`||Display version information.|

### Configuration Precedence

The service uses the following precedence order for configuration values:

1. Command-line flags (highest precedence)
2. Environment variables
3. Configuration file values
4. Default values (lowest precedence)

### License Header Maintenance

The git-sync project uses Apache License 2.0 for all source code files. Each Go source file should have a license header at the top with the copyright notice.

#### Tools Used

We use the VSCode licenser extension (`ymotongpoo.licenser`) to automatically add and update license headers in source files.

#### Adding License Headers

To add a license header to a file:

1. Open the file in VSCode
2. Open the Command Palette (`Ctrl+Shift+P` or `Cmd+Shift+P`)
3. Type "licenser: Insert license header"
4. Press Enter

The licenser will automatically insert the appropriate license header at the top of the file.

The license headers will be automatically updated with the current year. For more information, see [License Header Maintenance](./LICENSE-HEADER-MAINTENANCE.md).

### Prometheus Metrics

The service provides the following metrics with improved naming for better readability:

|Name|Description|
|-|-|
|`git_sync_changes_count`|Total number of synchronizations with changes.|
|`git_sync_total_count`|Total number of synchronizations.|
|`git_sync_error_total`|Total number of synchronization errors.|
|`git_sync_repo_info`|Information about the synchronized repository with labels for `repository name` and `repository branch`.|
|`git_sync_commit_info`|Information about the latest commit with labels for `commit hash`, `author name`, `author email`, `commit date`, `commit message`.|
|`git_sync_build_info`|Build information with labels for `version`, `commit`, `date`, and `dirty` status.|

Note: The metric names have been improved to avoid repetition and enhance readability. For example, `git_sync_sync_count` was renamed to `git_sync_changes_count` to eliminate the redundant "sync" word.

### Web Interface

The service provides an enhanced web dashboard at the root endpoint (`/`) that offers a user-friendly interface for monitoring and managing the service. The dashboard includes:

- **System Status Overview**: Immediate visual indication of service health
- **Repository Sync Status**: Information about the last synchronization
- **Manual Sync Button**: Quick access to trigger synchronization manually
- **Health Card**: Displays service health status with link to detailed health information
- **Metrics Card**: Shows key metrics with link to full Prometheus metrics
- **Version Card**: Displays version information with link to detailed version endpoint
- **Endpoint List**: Complete list of all available API endpoints

The dashboard features a modern, responsive design with clear visual indicators for different service states (operational, warning, error) and follows best practices for web-based monitoring interfaces.

### API Endpoints

The service provides several API endpoints for monitoring and management:

|Endpoint|Method|Description|
|-|-|-|
|`/status`|GET|Returns basic status information about the service|
|`/health`|GET|Returns detailed health information including Go runtime stats|
|`/ready`|GET|Returns readiness status for Kubernetes probes|
|`/version`|GET|Returns version information about the service|
|`/metrics`|GET|Returns Prometheus metrics|
|`/webhook`|POST|Triggers manual synchronization|
|`/`|GET|Returns the enhanced web dashboard|

All endpoints support optional authentication via:

- Basic authentication (username/password)
- Bearer token authentication

When authentication is configured, all endpoints require valid credentials to access.

The service uses a modular architecture with the following internal components:

1. **Server Package**: Manages the HTTP server lifecycle and coordinates route registration
2. **Router Package**: Provides centralized routing and middleware handling
3. **API Package**: Implements monitoring and management endpoints
4. **Handlers Package**: Implements specialized functionality like webhooks and metrics

### Use Cases

<b>Application Configuration Files</b>: Ensuring a single source of truth for application configuration files that frequently change and need to be synchronized across different instances.

<b>Deployment Scripts</b>: Automatically updating and synchronizing deployment scripts across different servers or environments to ensure all servers use the same version of the scripts.

<b>Server Configuration Files</b>: Synchronizing server configuration files, such as web server or database configuration files, to ensure all servers are configured correctly and uniformly.

<b>Documentation and Instructions</b>: Synchronizing documentation and instructions for developers to ensure they always have access to the most up-to-date information.