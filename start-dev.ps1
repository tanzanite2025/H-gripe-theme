param(
  [switch]$StopOnly,
  [switch]$PortsOnly
)

$ErrorActionPreference = 'Stop'

$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
$LogDir = Join-Path $Root 'output/dev'

$Ports = [ordered]@{
  Storefront = 9100
  Api        = 9200
  Admin      = 9300
  Postgres   = 9400
  Redis      = 9500
}

$AppPorts = @($Ports.Storefront, $Ports.Api, $Ports.Admin)

function Write-Section($Text) {
  Write-Host ''
  Write-Host '========================================' -ForegroundColor Cyan
  Write-Host $Text -ForegroundColor Cyan
  Write-Host '========================================' -ForegroundColor Cyan
}

function Write-Ok($Text) {
  Write-Host "[OK] $Text" -ForegroundColor Green
}

function Write-Warn($Text) {
  Write-Host "[WARN] $Text" -ForegroundColor Yellow
}

function Write-Fail($Text) {
  Write-Host "[FAIL] $Text" -ForegroundColor Red
}

function Get-ListenersOnPort([int]$Port) {
  Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue |
    Select-Object -ExpandProperty OwningProcess -Unique |
    Where-Object { $_ -and $_ -ne $PID }
}

function Stop-ListenersOnPort([int]$Port) {
  $pids = @(Get-ListenersOnPort $Port)
  if (-not $pids.Count) {
    return
  }

  foreach ($processId in $pids) {
    $process = Get-Process -Id $processId -ErrorAction SilentlyContinue
    if (-not $process) {
      continue
    }

    Write-Warn "Stopping process on port ${Port}: $($process.ProcessName) (PID $processId)"
    Stop-Process -Id $processId -Force -ErrorAction SilentlyContinue
  }
}

function Show-Ports {
  foreach ($entry in $Ports.GetEnumerator()) {
    $port = [int]$entry.Value
    $pids = @(Get-ListenersOnPort $port)
    if ($pids.Count) {
      foreach ($processId in $pids) {
        $process = Get-Process -Id $processId -ErrorAction SilentlyContinue
        Write-Host ("{0,-10} {1} LISTEN PID={2} PROC={3}" -f $entry.Key, $port, $processId, $process.ProcessName) -ForegroundColor Yellow
      }
    } else {
      Write-Host ("{0,-10} {1} FREE" -f $entry.Key, $port) -ForegroundColor Green
    }
  }
}

function Test-CommandExists([string]$Name) {
  return [bool](Get-Command $Name -ErrorAction SilentlyContinue)
}

function Wait-TcpPort([string]$HostName, [int]$Port, [int]$TimeoutSeconds) {
  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    try {
      $client = [System.Net.Sockets.TcpClient]::new()
      $async = $client.BeginConnect($HostName, $Port, $null, $null)
      $connected = $async.AsyncWaitHandle.WaitOne(700)
      if ($connected) {
        $client.EndConnect($async)
        $client.Dispose()
        return $true
      }
      $client.Dispose()
    } catch {
      Start-Sleep -Milliseconds 500
    }
    Start-Sleep -Milliseconds 500
  }
  return $false
}

function Wait-HttpOk([string]$Url, [int]$TimeoutSeconds) {
  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    try {
      Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 2 | Out-Null
      return $true
    } catch {
      Start-Sleep -Seconds 1
    }
  }
  return $false
}

function Start-DevProcess(
  [string]$Name,
  [string]$WorkingDirectory,
  [string]$Command,
  [string]$LogName
) {
  New-Item -ItemType Directory -Force -Path $LogDir | Out-Null

  $stdout = Join-Path $LogDir "$LogName.out.log"
  $stderr = Join-Path $LogDir "$LogName.err.log"
  Remove-Item -LiteralPath $stdout, $stderr -Force -ErrorAction SilentlyContinue

  $pwshCommand = Get-Command pwsh -ErrorAction SilentlyContinue
  $shell = $null
  if ($pwshCommand) {
    $shell = $pwshCommand.Source
  }
  if (-not $shell) {
    $shell = (Get-Command powershell -ErrorAction Stop).Source
  }

  $process = Start-Process `
    -FilePath $shell `
    -ArgumentList @('-NoLogo', '-NoProfile', '-ExecutionPolicy', 'Bypass', '-Command', $Command) `
    -WorkingDirectory $WorkingDirectory `
    -RedirectStandardOutput $stdout `
    -RedirectStandardError $stderr `
    -PassThru `
    -WindowStyle Hidden

  Write-Ok "$Name started, PID $($process.Id)"
  Write-Host "  stdout: $stdout" -ForegroundColor DarkGray
  Write-Host "  stderr: $stderr" -ForegroundColor DarkGray
}

function Stop-Dev {
  Write-Section 'Stopping Tanzanite local DEV'

  foreach ($port in $AppPorts) {
    Stop-ListenersOnPort $port
  }

  if (Test-CommandExists 'docker') {
    Push-Location $Root
    try {
      docker compose stop postgres redis | Out-Host
    } catch {
      Write-Warn "Docker infra was not stopped: $($_.Exception.Message)"
    } finally {
      Pop-Location
    }
  }

  Write-Ok 'Local DEV stopped'
}

if ($PortsOnly) {
  Write-Section 'Tanzanite local DEV ports'
  Show-Ports
  exit 0
}

if ($StopOnly) {
  Stop-Dev
  exit 0
}

Write-Section 'Starting Tanzanite local DEV'
Write-Host 'Port plan:' -ForegroundColor Cyan
Write-Host "  Storefront Nuxt : http://localhost:$($Ports.Storefront)" -ForegroundColor White
Write-Host "  Go API          : http://localhost:$($Ports.Api)" -ForegroundColor White
Write-Host "  Admin Console   : http://localhost:$($Ports.Admin)" -ForegroundColor White
Write-Host "  PostgreSQL      : localhost:$($Ports.Postgres)" -ForegroundColor White
Write-Host "  Redis           : localhost:$($Ports.Redis)" -ForegroundColor White

Write-Section 'Checking local tools'
foreach ($commandName in @('node', 'npm', 'go', 'docker')) {
  if (-not (Test-CommandExists $commandName)) {
    Write-Fail "$commandName is not installed or is not in PATH"
    exit 1
  }
  $source = (Get-Command $commandName).Source
  Write-Ok "${commandName}: $source"
}

Write-Section 'Cleaning app ports'
foreach ($port in $AppPorts) {
  Stop-ListenersOnPort $port
}

Write-Section 'Starting PostgreSQL / Redis'
Push-Location $Root
try {
  docker compose up -d postgres redis | Out-Host
} finally {
  Pop-Location
}

$postgresReady = Wait-TcpPort -HostName '127.0.0.1' -Port $Ports.Postgres -TimeoutSeconds 60
if (-not $postgresReady) {
  Write-Fail "PostgreSQL is not ready on localhost:$($Ports.Postgres)"
  exit 1
}
Write-Ok "PostgreSQL ready: localhost:$($Ports.Postgres)"

$redisReady = Wait-TcpPort -HostName '127.0.0.1' -Port $Ports.Redis -TimeoutSeconds 60
if (-not $redisReady) {
  Write-Fail "Redis is not ready on localhost:$($Ports.Redis)"
  exit 1
}
Write-Ok "Redis ready: localhost:$($Ports.Redis)"

Write-Section 'Starting Go API'
$backendCommand = @"
`$env:SERVER_PORT=':$($Ports.Api)'
`$env:SERVER_BASE_URL='http://localhost:$($Ports.Api)'
`$env:DB_HOST='localhost'
`$env:DB_PORT='$($Ports.Postgres)'
`$env:DB_USERNAME='tanzanite'
`$env:DB_PASSWORD='tanzanite_password'
`$env:DB_NAME='tanzanite'
`$env:REDIS_HOST='localhost'
`$env:REDIS_PORT='$($Ports.Redis)'
`$env:CORS_ORIGINS='http://localhost:$($Ports.Storefront),http://127.0.0.1:$($Ports.Storefront),http://localhost:$($Ports.Admin),http://127.0.0.1:$($Ports.Admin),http://localhost:$($Ports.Api),http://127.0.0.1:$($Ports.Api)'
`$env:STOREFRONT_HTML_CACHE_PURGE_URL='http://localhost:$($Ports.Storefront)/_internal/html-cache/purge'
`$env:STOREFRONT_HTML_CACHE_PURGE_TOKEN='dev-html-cache-purge-token'
go run ./cmd/server
"@
Start-DevProcess -Name 'Go API' -WorkingDirectory (Join-Path $Root 'go-backend') -Command $backendCommand -LogName 'go-api'

$apiHealthUrl = "http://localhost:$($Ports.Api)/health"
$apiReady = Wait-HttpOk -Url $apiHealthUrl -TimeoutSeconds 90
if ($apiReady) {
  Write-Ok "Go API health check passed: http://localhost:$($Ports.Api)/health"
} else {
  Write-Warn "Go API health check is not ready yet. Check log: $(Join-Path $LogDir 'go-api.err.log')"
}

Write-Section 'Starting Nuxt Storefront'
$storefrontCommand = @"
`$env:NUXT_PUBLIC_API_BASE='http://localhost:$($Ports.Api)'
`$env:API_INTERNAL_ORIGIN='http://localhost:$($Ports.Api)'
`$env:NUXT_HTML_CACHE_DRIVER='redis'
`$env:NUXT_HTML_CACHE_REDIS_HOST='localhost'
`$env:NUXT_HTML_CACHE_REDIS_PORT='$($Ports.Redis)'
`$env:NUXT_HTML_CACHE_REDIS_DB='1'
`$env:NUXT_HTML_CACHE_PURGE_TOKEN='dev-html-cache-purge-token'
npm run dev
"@
Start-DevProcess -Name 'Nuxt Storefront' -WorkingDirectory (Join-Path $Root 'nuxt-i18n') -Command $storefrontCommand -LogName 'storefront'

Write-Section 'Starting Admin Console'
$adminCommand = @"
`$env:VITE_API_BASE_URL=''
npm run dev
"@
Start-DevProcess -Name 'Admin Console' -WorkingDirectory (Join-Path $Root 'go-backend/web/admin') -Command $adminCommand -LogName 'admin'

Write-Section 'Local DEV is up'
Write-Host "Storefront : http://localhost:$($Ports.Storefront)" -ForegroundColor White
Write-Host "Go API     : http://localhost:$($Ports.Api)" -ForegroundColor White
Write-Host "Admin      : http://localhost:$($Ports.Admin)" -ForegroundColor White
Write-Host "Logs       : $LogDir" -ForegroundColor White
Write-Host ''
Write-Host 'Stop with: npm run dev:stop' -ForegroundColor Yellow
