# Copyright 2025 Aleksey Dobshikov
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     https://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# PowerShell script to build git-sync with .exe extension on Windows
# This script can be run from any location and will properly navigate to the project root

# Store the current location
$originalLocation = Get-Location

try {
    # Determine the script's directory
    $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    
    # Navigate to the script's directory first
    Set-Location -Path $scriptDir
    
    # Then navigate to the project root (parent directory)
    Set-Location -Path ".."
    
    # Verify we're in the correct directory by checking for go.mod file
    if (-not (Test-Path -Path "go.mod")) {
        Write-Host "Error: Could not find go.mod file. Please ensure you're running this script from the correct location." -ForegroundColor Red
        exit 1
    }
    
    Write-Host "Building git-sync.exe in project root directory..." -ForegroundColor Cyan
    
    # Create bin directory if it doesn't exist
    if (!(Test-Path -Path "bin")) {
        New-Item -ItemType Directory -Path "bin" | Out-Null
    }
    
    # Get git information for versioning
    $gitTag = try { git describe --tags --abbrev=0 2>$null } catch { "dev" }
    if (-not $gitTag) { $gitTag = "dev" }
    
    $gitCommit = try { git rev-parse --short HEAD 2>$null } catch { "none" }
    if (-not $gitCommit) { $gitCommit = "none" }
    
    $buildDate = try { Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ" } catch { "unknown" }
    
    $gitDirty = try { 
        $status = git status --porcelain 2>$null
        if ($status) { "true" } else { "false" }
    } catch { "false" }
    
    # Set ldflags for version information
    $ldflags = @"
-s -w -X 'git-sync/internal/version.Version=$gitTag' -X 'git-sync/internal/version.Commit=$gitCommit' -X 'git-sync/internal/version.Date=$buildDate' -X 'git-sync/internal/version.Dirty=$gitDirty'
"@
    
    # Build the application
    Write-Host "Building git-sync.exe with version info..." -ForegroundColor Yellow
    go build -trimpath -ldflags $ldflags -o bin/git-sync.exe ./cmd
    
    # Check if the build was successful
    if ($LASTEXITCODE -eq 0) {
        # Verify the binary was created
        if (Test-Path -Path "bin/git-sync.exe") {
            Write-Host "Build successful! Binary created at bin/git-sync.exe" -ForegroundColor Green
            Write-Host "Version info: $gitTag ($gitCommit, $buildDate, dirty=$gitDirty)" -ForegroundColor Cyan
        } else {
            Write-Host "Build command succeeded but binary was not found at bin/git-sync.exe" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "Build failed with exit code: $LASTEXITCODE" -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "An error occurred: $_" -ForegroundColor Red
    exit 1
}
finally {
    # Return to the original location
    Set-Location -Path $originalLocation
}