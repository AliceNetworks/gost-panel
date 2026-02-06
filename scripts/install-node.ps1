# GOST Panel Node Installer for Windows
# Usage: irm http://panel:8080/scripts/install-node.ps1 | iex
# Or: .\install-node.ps1 -PanelUrl "http://panel:8080" -Token "YOUR_TOKEN"

param(
    [Parameter(Mandatory=$true)]
    [string]$PanelUrl,

    [Parameter(Mandatory=$true)]
    [string]$Token,

    [string]$InstallDir = "C:\gost-panel",
    [string]$GostVersion = "3.0.0-rc10"
)

$ErrorActionPreference = "Stop"

function Write-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Green }
function Write-Warn { param($msg) Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Err { param($msg) Write-Host "[ERROR] $msg" -ForegroundColor Red }

Write-Host "========================================"
Write-Host "    GOST Panel Node Installer (Windows)"
Write-Host "========================================"
Write-Host ""
Write-Info "Panel: $PanelUrl"

# Detect architecture
$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    "x86"   { "386" }
    default { "amd64" }
}
Write-Info "Architecture: $arch"

# Create install directory
Write-Info "[1/5] Creating directories..."
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
New-Item -ItemType Directory -Force -Path "$InstallDir\config" | Out-Null

# Download GOST
Write-Info "[2/5] Downloading GOST..."
$gostZip = "$env:TEMP\gost.zip"
$gostUrl = "https://github.com/go-gost/gost/releases/download/v$GostVersion/gost_${GostVersion}_windows_$arch.zip"

try {
    Invoke-WebRequest -Uri $gostUrl -OutFile $gostZip -UseBasicParsing
    Expand-Archive -Path $gostZip -DestinationPath "$env:TEMP\gost-extract" -Force
    Move-Item -Path "$env:TEMP\gost-extract\gost.exe" -Destination "$InstallDir\gost.exe" -Force
    Remove-Item -Path $gostZip -Force
    Remove-Item -Path "$env:TEMP\gost-extract" -Recurse -Force
    Write-Info "GOST installed to $InstallDir\gost.exe"
} catch {
    Write-Err "Failed to download GOST: $_"
    exit 1
}

# Download Agent
Write-Info "[3/5] Downloading agent..."
$useAgent = $false

# Stop existing service first to avoid file lock
$existingService = Get-Service -Name "GostNode" -ErrorAction SilentlyContinue
if ($existingService -and $existingService.Status -eq 'Running') {
    Write-Info "Stopping existing service..."
    Stop-Service -Name "GostNode" -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
}

# Remove old agent file
if (Test-Path "$InstallDir\gost-agent.exe") {
    Remove-Item -Path "$InstallDir\gost-agent.exe" -Force -ErrorAction SilentlyContinue
}

# Get latest version from GitHub
$repo = "AliceNetworks/gost-panel"
try {
    $releaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -UseBasicParsing
    $latestVersion = $releaseInfo.tag_name
    Write-Info "Latest version: $latestVersion"
} catch {
    $latestVersion = "v1.1.0"
    Write-Warn "Could not fetch latest version, using $latestVersion"
}

# Try GitHub first
$agentGithubUrl = "https://github.com/$repo/releases/download/$latestVersion/gost-agent-windows-$arch.exe"
try {
    Write-Info "Downloading agent from GitHub..."
    Invoke-WebRequest -Uri $agentGithubUrl -OutFile "$InstallDir\gost-agent.exe" -UseBasicParsing
    Write-Info "Agent downloaded to $InstallDir\gost-agent.exe"
    $useAgent = $true
} catch {
    Write-Warn "GitHub download failed, trying panel..."
    $agentUrl = "$PanelUrl/agent/download/windows/$arch"
    try {
        Invoke-WebRequest -Uri $agentUrl -OutFile "$InstallDir\gost-agent.exe" -UseBasicParsing
        Write-Info "Agent downloaded from panel"
        $useAgent = $true
    } catch {
        Write-Warn "Agent not available, using GOST directly"
    }
}

# Download config
Write-Info "[4/5] Downloading config..."
try {
    Invoke-WebRequest -Uri "$PanelUrl/agent/config/$Token" -OutFile "$InstallDir\config\gost.yml" -UseBasicParsing
    Write-Info "Config saved to $InstallDir\config\gost.yml"
} catch {
    Write-Err "Failed to download config: $_"
    exit 1
}

# Create Windows Service
Write-Info "[5/5] Installing Windows service..."

$serviceName = "GostNode"
$serviceDisplayName = "GOST Panel Node"

# Stop and remove existing service
$existingService = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
if ($existingService) {
    Write-Info "Stopping existing service..."
    Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue
    sc.exe delete $serviceName | Out-Null
    Start-Sleep -Seconds 2
}

# Use NSSM or create scheduled task
$nssmPath = "$InstallDir\nssm.exe"
$nssmUrl = "https://nssm.cc/release/nssm-2.24.zip"

try {
    # Download NSSM
    Write-Info "Downloading NSSM service manager..."
    $nssmZip = "$env:TEMP\nssm.zip"
    Invoke-WebRequest -Uri $nssmUrl -OutFile $nssmZip -UseBasicParsing
    Expand-Archive -Path $nssmZip -DestinationPath "$env:TEMP\nssm-extract" -Force

    $nssmExe = if ([Environment]::Is64BitOperatingSystem) {
        Get-ChildItem "$env:TEMP\nssm-extract" -Recurse -Filter "nssm.exe" | Where-Object { $_.DirectoryName -match "win64" } | Select-Object -First 1
    } else {
        Get-ChildItem "$env:TEMP\nssm-extract" -Recurse -Filter "nssm.exe" | Where-Object { $_.DirectoryName -match "win32" } | Select-Object -First 1
    }

    Copy-Item -Path $nssmExe.FullName -Destination $nssmPath -Force
    Remove-Item -Path $nssmZip -Force
    Remove-Item -Path "$env:TEMP\nssm-extract" -Recurse -Force

    # Install service with NSSM
    if ($useAgent) {
        & $nssmPath install $serviceName "$InstallDir\gost-agent.exe"
        & $nssmPath set $serviceName AppParameters "-panel `"$PanelUrl`" -token `"$Token`""
    } else {
        & $nssmPath install $serviceName "$InstallDir\gost.exe"
        & $nssmPath set $serviceName AppParameters "-C `"$InstallDir\config\gost.yml`""
    }

    & $nssmPath set $serviceName DisplayName $serviceDisplayName
    & $nssmPath set $serviceName Start SERVICE_AUTO_START
    & $nssmPath set $serviceName AppStdout "$InstallDir\logs\stdout.log"
    & $nssmPath set $serviceName AppStderr "$InstallDir\logs\stderr.log"
    & $nssmPath set $serviceName AppRotateFiles 1
    & $nssmPath set $serviceName AppRotateBytes 10485760

    New-Item -ItemType Directory -Force -Path "$InstallDir\logs" | Out-Null

    # Start service
    Start-Service -Name $serviceName
    Write-Info "Service installed and started"

} catch {
    Write-Warn "NSSM installation failed, creating scheduled task instead..."

    # Fallback: Create scheduled task
    $taskName = "GostNode"

    Unregister-ScheduledTask -TaskName $taskName -Confirm:$false -ErrorAction SilentlyContinue

    if ($useAgent) {
        $action = New-ScheduledTaskAction -Execute "$InstallDir\gost-agent.exe" -Argument "-panel `"$PanelUrl`" -token `"$Token`""
    } else {
        $action = New-ScheduledTaskAction -Execute "$InstallDir\gost.exe" -Argument "-C `"$InstallDir\config\gost.yml`""
    }

    $trigger = New-ScheduledTaskTrigger -AtStartup
    $principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
    $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -RestartInterval (New-TimeSpan -Minutes 1) -RestartCount 3

    Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Principal $principal -Settings $settings -Force
    Start-ScheduledTask -TaskName $taskName

    Write-Info "Scheduled task created and started"
}

# Add firewall rule
Write-Info "Adding firewall rules..."
$ports = @(38567, 18080)  # Default proxy and API ports
foreach ($port in $ports) {
    New-NetFirewallRule -DisplayName "GOST Node Port $port" -Direction Inbound -Protocol TCP -LocalPort $port -Action Allow -ErrorAction SilentlyContinue | Out-Null
}

Write-Host ""
Write-Host "========================================"
Write-Host "    Installation Complete!"
Write-Host "========================================"
Write-Host ""
Write-Host "Install directory: $InstallDir"
Write-Host ""
Write-Host "Commands:"
Write-Host "  Get-Service $serviceName              - Check status"
Write-Host "  Restart-Service $serviceName          - Restart"
Write-Host "  Get-Content $InstallDir\logs\*.log    - View logs"
Write-Host ""
Write-Host "To uninstall:"
Write-Host "  Stop-Service $serviceName"
Write-Host "  sc.exe delete $serviceName"
Write-Host "  Remove-Item -Recurse $InstallDir"
