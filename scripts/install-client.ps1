# GOST Panel Client Installer for Windows
# Usage: .\install-client.ps1 -PanelUrl "http://panel:8080" -Token "YOUR_TOKEN"

param(
    [Parameter(Mandatory=$true)]
    [string]$PanelUrl,

    [Parameter(Mandatory=$true)]
    [string]$Token,

    [string]$InstallDir = "C:\gost-panel-client",
    [string]$GostVersion = "3.0.0-rc10"
)

$ErrorActionPreference = "Stop"

function Write-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Green }
function Write-Warn { param($msg) Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Err { param($msg) Write-Host "[ERROR] $msg" -ForegroundColor Red }

Write-Host "========================================"
Write-Host "  GOST Panel Client Installer (Windows)"
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
Write-Info "[1/4] Creating directories..."
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
New-Item -ItemType Directory -Force -Path "$InstallDir\config" | Out-Null
New-Item -ItemType Directory -Force -Path "$InstallDir\logs" | Out-Null

# Download GOST
Write-Info "[2/4] Downloading GOST..."
$gostZip = "$env:TEMP\gost.zip"
$gostUrl = "https://github.com/go-gost/gost/releases/download/v$GostVersion/gost_${GostVersion}_windows_$arch.zip"

if (-not (Test-Path "$InstallDir\gost.exe")) {
    try {
        Invoke-WebRequest -Uri $gostUrl -OutFile $gostZip -UseBasicParsing
        Expand-Archive -Path $gostZip -DestinationPath "$env:TEMP\gost-extract" -Force
        Move-Item -Path "$env:TEMP\gost-extract\gost.exe" -Destination "$InstallDir\gost.exe" -Force
        Remove-Item -Path $gostZip -Force -ErrorAction SilentlyContinue
        Remove-Item -Path "$env:TEMP\gost-extract" -Recurse -Force -ErrorAction SilentlyContinue
        Write-Info "GOST installed to $InstallDir\gost.exe"
    } catch {
        Write-Err "Failed to download GOST: $_"
        exit 1
    }
} else {
    Write-Info "GOST already installed"
}

# Download config
Write-Info "[3/4] Downloading config..."
try {
    Invoke-WebRequest -Uri "$PanelUrl/agent/config/$Token" -OutFile "$InstallDir\config\client.yml" -UseBasicParsing
    Write-Info "Config saved to $InstallDir\config\client.yml"
} catch {
    Write-Err "Failed to download config: $_"
    exit 1
}

# Create Windows Service
Write-Info "[4/5] Installing Windows service..."

$serviceName = "GostClient"
$serviceDisplayName = "GOST Panel Client"

# Stop and remove existing service
$existingService = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
if ($existingService) {
    Write-Info "Stopping existing service..."
    Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue
    sc.exe delete $serviceName | Out-Null
    Start-Sleep -Seconds 2
}

# Download NSSM if not exists
$nssmPath = "$InstallDir\nssm.exe"
if (-not (Test-Path $nssmPath)) {
    $nssmUrl = "https://nssm.cc/release/nssm-2.24.zip"
    try {
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
        Remove-Item -Path $nssmZip -Force -ErrorAction SilentlyContinue
        Remove-Item -Path "$env:TEMP\nssm-extract" -Recurse -Force -ErrorAction SilentlyContinue
    } catch {
        Write-Warn "NSSM download failed, will use scheduled task"
    }
}

if (Test-Path $nssmPath) {
    # Install service with NSSM
    & $nssmPath install $serviceName "$InstallDir\gost.exe"
    & $nssmPath set $serviceName AppParameters "-C `"$InstallDir\config\client.yml`""
    & $nssmPath set $serviceName DisplayName $serviceDisplayName
    & $nssmPath set $serviceName Start SERVICE_AUTO_START
    & $nssmPath set $serviceName AppStdout "$InstallDir\logs\stdout.log"
    & $nssmPath set $serviceName AppStderr "$InstallDir\logs\stderr.log"
    & $nssmPath set $serviceName AppRotateFiles 1
    & $nssmPath set $serviceName AppRotateBytes 10485760

    # Start service
    Start-Service -Name $serviceName
    Write-Info "Service installed and started"
} else {
    # Fallback: Create scheduled task
    Write-Warn "Using scheduled task as fallback..."
    $taskName = "GostClient"

    Unregister-ScheduledTask -TaskName $taskName -Confirm:$false -ErrorAction SilentlyContinue

    $action = New-ScheduledTaskAction -Execute "$InstallDir\gost.exe" -Argument "-C `"$InstallDir\config\client.yml`""
    $trigger = New-ScheduledTaskTrigger -AtStartup
    $principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
    $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -RestartInterval (New-TimeSpan -Minutes 1) -RestartCount 3

    Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Principal $principal -Settings $settings -Force
    Start-ScheduledTask -TaskName $taskName

    Write-Info "Scheduled task created and started"
}

# Install heartbeat with auto-uninstall
Write-Info "[4.5/5] Setting up heartbeat..."

$heartbeatScript = @"
# GOST Client Heartbeat (auto-uninstall on 410 Gone)
try {
    `$response = Invoke-WebRequest -Uri "$PanelUrl/agent/client-heartbeat/$Token" -Method POST -UseBasicParsing -ErrorAction Stop
} catch {
    `$statusCode = `$_.Exception.Response.StatusCode.value__
    if (`$statusCode -eq 410) {
        # Client deleted from panel, auto-uninstall
        Write-Host "[GOST] Client deleted from panel, auto-uninstalling..."

        # Stop and remove service
        `$svc = Get-Service -Name "GostClient" -ErrorAction SilentlyContinue
        if (`$svc) {
            Stop-Service -Name "GostClient" -Force -ErrorAction SilentlyContinue
            if (Test-Path "$InstallDir\nssm.exe") {
                & "$InstallDir\nssm.exe" remove "GostClient" confirm 2>`$null
            } else {
                sc.exe delete "GostClient" 2>`$null
            }
        }

        # Remove scheduled tasks
        Unregister-ScheduledTask -TaskName "GostClient" -Confirm:`$false -ErrorAction SilentlyContinue
        Unregister-ScheduledTask -TaskName "GostHeartbeat" -Confirm:`$false -ErrorAction SilentlyContinue

        # Remove files
        Start-Sleep -Seconds 2
        Remove-Item -Recurse -Force "$InstallDir" -ErrorAction SilentlyContinue

        Write-Host "[GOST] Uninstall complete."
    }
}
"@

$heartbeatPath = "$InstallDir\heartbeat.ps1"
Set-Content -Path $heartbeatPath -Value $heartbeatScript -Force

# Create scheduled task for heartbeat (every minute)
$heartbeatTaskName = "GostHeartbeat"
Unregister-ScheduledTask -TaskName $heartbeatTaskName -Confirm:$false -ErrorAction SilentlyContinue

$hbAction = New-ScheduledTaskAction -Execute "powershell.exe" -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$heartbeatPath`""
$hbTrigger = New-ScheduledTaskTrigger -Once -At (Get-Date) -RepetitionInterval (New-TimeSpan -Minutes 1)
$hbPrincipal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
$hbSettings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -Hidden

Register-ScheduledTask -TaskName $heartbeatTaskName -Action $hbAction -Trigger $hbTrigger -Principal $hbPrincipal -Settings $hbSettings -Force | Out-Null

# Send first heartbeat
try { Invoke-WebRequest -Uri "$PanelUrl/agent/client-heartbeat/$Token" -Method POST -UseBasicParsing -ErrorAction SilentlyContinue | Out-Null } catch {}
Write-Info "Heartbeat configured"

# Extract local port from config
$localPort = 38777
try {
    $configContent = Get-Content "$InstallDir\config\client.yml" -Raw
    if ($configContent -match 'addr:\s*":(\d+)"') {
        $localPort = $Matches[1]
    }
} catch {}

Write-Host ""
Write-Host "========================================"
Write-Host "    Installation Complete!"
Write-Host "========================================"
Write-Host ""
Write-Host "Install directory: $InstallDir"
Write-Host ""
Write-Host "Local SOCKS5 proxy: socks5://127.0.0.1:$localPort"
Write-Host ""
Write-Host "Commands:"
Write-Host "  Get-Service $serviceName              - Check status"
Write-Host "  Restart-Service $serviceName          - Restart"
Write-Host "  Get-Content $InstallDir\logs\*.log    - View logs"
Write-Host ""
Write-Host "Proxy settings (for browsers/apps):"
Write-Host "  Type: SOCKS5"
Write-Host "  Host: 127.0.0.1"
Write-Host "  Port: $localPort"
Write-Host ""
Write-Host "To uninstall:"
Write-Host "  Stop-Service $serviceName"
Write-Host "  sc.exe delete $serviceName"
Write-Host "  Remove-Item -Recurse $InstallDir"
