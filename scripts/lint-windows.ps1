# Copyright 2025 Aleksey Dobshikov
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# PowerShell script to run golangci-lint on Windows
# This script will automatically download and install golangci-lint if not present

param(
    [string]$Version = "2.5.0"  # Default version to download
)

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
    
    Write-Host "Checking for golangci-lint installation..." -ForegroundColor Cyan
    
    # Define installation path in AppData
    $appData = $env:APPDATA
    $lintDir = Join-Path $appData "golangci-lint"
    $lintBin = Join-Path $lintDir "golangci-lint.exe"
    
    # Check if golangci-lint is already installed and in PATH
    $lintInPath = Get-Command golangci-lint -ErrorAction SilentlyContinue
    if ($lintInPath) {
        Write-Host "Found golangci-lint in PATH" -ForegroundColor Green
        $lintCmd = "golangci-lint"
    }
    # Check if we have it in our AppData directory
    elseif (Test-Path $lintBin) {
        Write-Host "Found golangci-lint in AppData" -ForegroundColor Green
        $lintCmd = $lintBin
    }
    else {
        Write-Host "golangci-lint not found, downloading version $Version..." -ForegroundColor Yellow
        
        # Create directory if it doesn't exist
        if (!(Test-Path -Path $lintDir)) {
            New-Item -ItemType Directory -Path $lintDir | Out-Null
        }
        
        # Download URL for Windows
        $downloadUrl = "https://github.com/golangci/golangci-lint/releases/download/v$Version/golangci-lint-$Version-windows-amd64.zip"
        $zipFile = Join-Path $lintDir "golangci-lint.zip"
        
        try {
            Write-Host "Downloading from $downloadUrl" -ForegroundColor Yellow
            Invoke-WebRequest -Uri $downloadUrl -OutFile $zipFile
            
            # Extract the zip file
            Write-Host "Extracting archive..." -ForegroundColor Yellow
            Expand-Archive -Path $zipFile -DestinationPath $lintDir -Force
            
            # Find the extracted binary (it will be in a subdirectory)
            $extractedDir = Get-ChildItem -Path $lintDir -Directory | Where-Object { $_.Name -like "golangci-lint*" }
            if ($extractedDir) {
                $extractedBin = Join-Path $extractedDir.FullName "golangci-lint.exe"
                if (Test-Path $extractedBin) {
                    # Move the binary to our target location
                    Move-Item -Path $extractedBin -Destination $lintBin -Force
                    Write-Host "golangci-lint installed successfully to $lintBin" -ForegroundColor Green
                    $lintCmd = $lintBin
                }
            }
            
            # Clean up
            Remove-Item -Path $zipFile -Force
            if ($extractedDir) {
                Remove-Item -Path $extractedDir.FullName -Recurse -Force
            }
        }
        catch {
            Write-Host "Failed to download or install golangci-lint: $_" -ForegroundColor Red
            Write-Host "Trying to use go run method instead..." -ForegroundColor Yellow
            
            # Try to run via go run as fallback
            try {
                Write-Host "Running golangci-lint via 'go run'..." -ForegroundColor Yellow
                go run github.com/golangci/golangci-lint/cmd/golangci-lint@v$Version run -c ./.golangci.yml
                if ($LASTEXITCODE -eq 0) {
                    Write-Host "Linting completed successfully via go run" -ForegroundColor Green
                    exit 0
                } else {
                    Write-Host "Linting failed with exit code: $LASTEXITCODE" -ForegroundColor Red
                    exit 1
                }
            }
            catch {
                Write-Host "Failed to run golangci-lint via 'go run': $_" -ForegroundColor Red
                exit 1
            }

        }
    }
    
    # Run the linter
    Write-Host "Running golangci-lint..." -ForegroundColor Yellow
    & $lintCmd run -c ./.golangci.yml
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Linting completed successfully!" -ForegroundColor Green
    } else {
        Write-Host "Linting failed with exit code: $LASTEXITCODE" -ForegroundColor Red
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