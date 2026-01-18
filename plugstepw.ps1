$ErrorActionPreference = "Stop"

$PlugstepBaseUrl = if ($env:PLUGSTEP_BASE_URL) { $env:PLUGSTEP_BASE_URL } else { "https://releases.perny.dev/mineframe/plugstep" }
$PlugstepCacheDir = if ($env:PLUGSTEP_CACHE_DIR) { $env:PLUGSTEP_CACHE_DIR } else { "$env:LOCALAPPDATA\plugstep" }

$VersionFile = ".plugstep-version"

if (-not (Test-Path $VersionFile)) {
    Write-Error "Error: $VersionFile not found`nCreate a $VersionFile file containing the version (e.g., v1.0.0)"
    exit 1
}

$Version = (Get-Content $VersionFile -Raw).Trim()
if ([string]::IsNullOrEmpty($Version)) {
    Write-Error "Error: $VersionFile is empty"
    exit 1
}

$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$BinaryName = "plugstep-windows-$Arch.exe"
$CachedBinary = Join-Path $PlugstepCacheDir "$Version\$BinaryName"

if (-not (Test-Path $CachedBinary)) {
    Write-Host "Downloading plugstep $Version for windows/$Arch..." -ForegroundColor Cyan

    $CacheVersionDir = Join-Path $PlugstepCacheDir $Version
    if (-not (Test-Path $CacheVersionDir)) {
        New-Item -ItemType Directory -Path $CacheVersionDir -Force | Out-Null
    }

    $DownloadUrl = "$PlugstepBaseUrl/$Version/$BinaryName"

    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $CachedBinary -UseBasicParsing
    } catch {
        Write-Error "Failed to download from $DownloadUrl`n$_"
        exit 1
    }
}

& $CachedBinary @args
exit $LASTEXITCODE
