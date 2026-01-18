$ErrorActionPreference = "Stop"

$WrapperBaseUrl = if ($env:WRAPPER_BASE_URL) { $env:WRAPPER_BASE_URL } else { "https://releases.perny.dev/mineframe/plugstep-wrapper" }
$ReleasesBaseUrl = if ($env:RELEASES_BASE_URL) { $env:RELEASES_BASE_URL } else { "https://releases.perny.dev/mineframe/plugstep" }

Write-Host "Installing plugstepw..."

# Download bash wrapper
try {
    Invoke-WebRequest -Uri "$WrapperBaseUrl/plugstepw" -OutFile "plugstepw" -UseBasicParsing
} catch {
    Write-Error "Failed to download bash wrapper: $_"
    exit 1
}

# Download PowerShell wrapper
try {
    Invoke-WebRequest -Uri "$WrapperBaseUrl/plugstepw.ps1" -OutFile "plugstepw.ps1" -UseBasicParsing
} catch {
    Write-Error "Failed to download PowerShell wrapper: $_"
    exit 1
}

# Fetch and set latest version
Write-Host "Fetching latest version..."
try {
    Invoke-WebRequest -Uri "$ReleasesBaseUrl/latest" -OutFile ".plugstep-version" -UseBasicParsing
} catch {
    Write-Error "Failed to fetch latest version: $_"
    exit 1
}

$Version = (Get-Content ".plugstep-version" -Raw).Trim()
Write-Host "Installed plugstepw (bash + PowerShell)"
Write-Host "Set version to $Version"
Write-Host "Run .\plugstepw.ps1 to get started"
