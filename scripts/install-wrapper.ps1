$ErrorActionPreference = "Stop"

$WrapperBaseUrl = if ($env:WRAPPER_BASE_URL) { $env:WRAPPER_BASE_URL } else { "https://releases.perny.dev/mineframe/plugstep-wrapper" }

Write-Host "Installing plugstepw.ps1..."

try {
    Invoke-WebRequest -Uri "$WrapperBaseUrl/plugstepw.ps1" -OutFile "plugstepw.ps1" -UseBasicParsing
} catch {
    Write-Error "Failed to download wrapper: $_"
    exit 1
}

Write-Host "Installed plugstepw.ps1 to current directory."
Write-Host "Create a .plugstep-version file with your desired version (e.g., v1.0.0)"
