#requires -Version 5.1
<#
==============================================================================
 bgscan-builder installer (Windows)
------------------------------------------------------------------------------
 Project:
   bgscan-builder

 Purpose:
   Downloads the correct bgscan-builder binary for the current Windows
   architecture and installs it into the root directory of a valid
   BgScan project.

 Supported Platforms:
   - Windows AMD64
   - Windows ARM64
   - Windows x86

 Installation Target:
   .\bgscan-builder.exe

 Safety Checks:
   - Must be executed from a BgScan project root
   - Requires go.mod to exist
   - Requires module name to contain "bgscan"

 Example:
   .\install-builder.ps1

 Notes:
   - This is the Windows counterpart to install-builder.sh.
   - Requires PowerShell 5.1+ (ships with Windows 10/11) or PowerShell 7+.
==============================================================================
#>

$ErrorActionPreference = "Stop"

# ==============================================================================
# CONFIGURATION
# ==============================================================================

$RepositoryOwner = "MohsenBg"
$RepositoryName  = "bgscan-builder"
$InstallName     = "bgscan-builder.exe"
$ScriptName      = Split-Path -Leaf $PSCommandPath
$StartTime       = Get-Date
$Step            = 0

# ==============================================================================
# LOGGER
# ==============================================================================

function Get-Timestamp {
    Get-Date -Format "yyyy-MM-dd HH:mm:ss"
}

function Write-Step {
    param([string]$Message)
    $script:Step++
    Write-Host ""
    Write-Host ("-" * 64) -ForegroundColor Blue
    Write-Host ("[STEP {0}] {1}  {2}" -f $script:Step, (Get-Timestamp), $ScriptName) -ForegroundColor Blue
    Write-Host ("-" * 64) -ForegroundColor Blue
    Write-Host "  $Message"
}

function Write-Info {
    param([string]$Message)
    Write-Host ("  {0} [INFO]  {1}" -f (Get-Timestamp), $Message) -ForegroundColor Green
}

function Write-WarnLog {
    param([string]$Message)
    Write-Host ("  {0} [WARN]  {1}" -f (Get-Timestamp), $Message) -ForegroundColor Yellow
}

function Write-Success {
    param([string]$Message)
    Write-Host ("  {0} [ OK ]  {1}" -f (Get-Timestamp), $Message) -ForegroundColor Green
}

function Write-Fail {
    param([string]$Message)
    Write-Host ""
    Write-Host ("-" * 64) -ForegroundColor Red
    Write-Host ("[FATAL] {0}  {1}" -f (Get-Timestamp), $Message) -ForegroundColor Red
    Write-Host ("-" * 64) -ForegroundColor Red
    Write-Host ""
    exit 1
}

# ==============================================================================
# PROJECT VALIDATION
# ==============================================================================

function Validate-Project {
    Write-Step "Validating BgScan project"

    Write-Info "Checking for go.mod in current directory"
    if (-not (Test-Path "go.mod")) {
        Write-Fail "go.mod not found. Run this installer from the project root."
    }
    Write-Success "go.mod located"

    $moduleLine = Select-String -Path "go.mod" -Pattern '^module\s+(\S+)' | Select-Object -First 1
    if (-not $moduleLine) {
        Write-Fail "Unable to determine Go module name from go.mod"
    }
    $moduleName = $moduleLine.Matches[0].Groups[1].Value
    Write-Info "Module name resolved: $moduleName"

    if ($moduleName -notmatch "bgscan") {
        Write-Fail "Unsupported module '$moduleName'. Expected a BgScan project."
    }

    Write-Success "Project validated - module: $moduleName"
    return $moduleName
}

# ==============================================================================
# PLATFORM DETECTION
# ==============================================================================

function Detect-Platform {
    Write-Step "Detecting operating system and architecture"

    $arch = $env:PROCESSOR_ARCHITECTURE
    $archW6432 = $env:PROCESSOR_ARCHITEW6432
    if ($archW6432) { $arch = $archW6432 }

    Write-Info "Raw architecture value: $arch"

    switch -Regex ($arch) {
        "AMD64" { $script:Arch = "64" }
        "ARM64" { $script:Arch = "arm64" }
        "x86"   { $script:Arch = "32" }
        default { Write-Fail "Unsupported architecture: $arch" }
    }

    $script:Platform  = "windows"
    $script:AssetName = "bgscan-builder-$($script:Platform)-$($script:Arch).exe"

    Write-Success "Platform resolved: $($script:Platform) ($arch)"
    Write-Info "Target asset name: $($script:AssetName)"
}

# ==============================================================================
# RELEASE URL RESOLUTION
# ==============================================================================

function Build-DownloadUrl {
    Write-Step "Resolving release download URL"

    $script:DownloadUrl = "https://github.com/$RepositoryOwner/$RepositoryName/releases/latest/download/$($script:AssetName)"

    Write-Info "Repository : $RepositoryOwner/$RepositoryName"
    Write-Info "Asset      : $($script:AssetName)"
    Write-Success "Download URL resolved: $($script:DownloadUrl)"
}

# ==============================================================================
# INSTALLATION
# ==============================================================================

function Install-Builder {
    Write-Step "Downloading bgscan-builder binary"

    Write-Info "Fetching from: $($script:DownloadUrl)"

    try {
        Invoke-WebRequest -Uri $script:DownloadUrl -OutFile $InstallName -UseBasicParsing
    }
    catch {
        Write-Fail "Download failed. Check your network connection and that a release exists for $($script:AssetName). Details: $($_.Exception.Message)"
    }

    Write-Success "Binary downloaded to .\$InstallName"
}

# ==============================================================================
# POST-INSTALL VERIFICATION
# ==============================================================================

function Verify-Installation {
    Write-Step "Verifying installation"

    if (-not (Test-Path $InstallName)) {
        Write-Fail "Installation verification failed - $InstallName not found."
    }

    $fileInfo = Get-Item $InstallName
    $sizeKb = [math]::Round($fileInfo.Length / 1KB, 1)

    Write-Success "Binary verified: $InstallName"
    Write-Info "Details: $sizeKb KB, last modified $($fileInfo.LastWriteTime)"
}

# ==============================================================================
# MAIN
# ==============================================================================

function Main {
    Write-Step "Starting bgscan-builder installation"

    Validate-Project | Out-Null
    Detect-Platform
    Build-DownloadUrl
    Install-Builder
    Verify-Installation

    $elapsed = [math]::Round(((Get-Date) - $StartTime).TotalSeconds, 1)

    Write-Host ""
    Write-Host ("=" * 64) -ForegroundColor Green
    Write-Host "  bgscan-builder installed successfully" -ForegroundColor Green
    Write-Host "  Binary      : .\$InstallName" -ForegroundColor Green
    Write-Host "  Total steps : $Step" -ForegroundColor Green
    Write-Host "  Elapsed time: ${elapsed}s" -ForegroundColor Green
    Write-Host "  Finished at : $(Get-Timestamp)" -ForegroundColor Green
    Write-Host ("=" * 64) -ForegroundColor Green
    Write-Host ""
}

Main
