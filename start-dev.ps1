param(
  [switch]$StopOnly,
  [switch]$PortsOnly
)

$ErrorActionPreference = 'Stop'

$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
$LogDir = Join-Path $Root 'output/dev'

$Ports = [ordered]@{
  Storefront = 9199
  Api        = 9200
  Admin      = 9300
  SiteQualityRunner = 9240
  Postgres   = 9400
  Redis      = 9510
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

  return $process
}

function Get-EnvOrDefault([string]$Name, [string]$Default) {
  $value = [Environment]::GetEnvironmentVariable($Name, 'Process')
  if ([string]::IsNullOrWhiteSpace($value)) {
    return $Default
  }
  return $value
}

function Invoke-DevAdminBootstrap {
  Write-Section 'Ensuring DEV admin account'

  $devAdminEmail = Get-EnvOrDefault 'DEV_ADMIN_EMAIL' 'admin@example.com'
  $devAdminUsername = Get-EnvOrDefault 'DEV_ADMIN_USERNAME' 'admin'
  $devAdminPassword = Get-EnvOrDefault 'DEV_ADMIN_PASSWORD' 'Admin123456!'
  $devAdminRole = Get-EnvOrDefault 'DEV_ADMIN_ROLE' 'admin'
  $devAdminReset = Get-EnvOrDefault 'DEV_ADMIN_RESET' ''

  $envValues = [ordered]@{
    SERVER_MODE         = 'debug'
    DB_HOST             = 'localhost'
    DB_PORT             = [string]$Ports.Postgres
    DB_USERNAME         = 'commerce_platform'
    DB_PASSWORD         = 'commerce_platform_password'
    DB_NAME             = 'commerce_platform'
    DB_LOG_LEVEL        = 'silent'
    DEV_ADMIN_BOOTSTRAP = 'true'
    DEV_ADMIN_EMAIL     = $devAdminEmail
    DEV_ADMIN_USERNAME  = $devAdminUsername
    DEV_ADMIN_PASSWORD  = $devAdminPassword
    DEV_ADMIN_ROLE      = $devAdminRole
    DEV_ADMIN_RESET     = $devAdminReset
  }

  $previousEnv = @{}
  foreach ($key in $envValues.Keys) {
    $previousEnv[$key] = [Environment]::GetEnvironmentVariable($key, 'Process')
    [Environment]::SetEnvironmentVariable($key, $envValues[$key], 'Process')
  }

  Push-Location (Join-Path $Root 'go-backend')
  try {
    go run ./cmd/dev-admin | Out-Host
    Write-Ok "DEV admin account ready: $devAdminEmail"
  } catch {
    Write-Fail "DEV admin bootstrap failed: $($_.Exception.Message)"
    exit 1
  } finally {
    Pop-Location
    foreach ($key in $previousEnv.Keys) {
      [Environment]::SetEnvironmentVariable($key, $previousEnv[$key], 'Process')
    }
  }
}

function Show-LogTail([string]$Path) {
  if (-not (Test-Path -LiteralPath $Path)) {
    Write-Warn "Log not found: $Path"
    return
  }

  Get-Content -LiteralPath $Path -Tail 40 | ForEach-Object {
    Write-Host "  $_" -ForegroundColor DarkGray
  }
}

function Clear-StorefrontIconCache {
  $cachePath = Join-Path $Root 'nuxt-i18n/.nuxt/cache/nuxt/icon'
  if (-not (Test-Path -LiteralPath $cachePath)) {
    return
  }

  $resolvedRoot = (Resolve-Path -LiteralPath $Root).Path
  $resolvedCache = (Resolve-Path -LiteralPath $cachePath).Path
  if (-not $resolvedCache.StartsWith($resolvedRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    Write-Warn "Skipping Nuxt Icon cache outside repository: $resolvedCache"
    return
  }

  Remove-Item -LiteralPath $resolvedCache -Recurse -Force -ErrorAction SilentlyContinue
  Write-Ok 'Cleared Nuxt Icon cache'
}

function Wait-DevHttpReady(
  [string]$Name,
  [System.Diagnostics.Process]$Process,
  [string]$Url,
  [string]$ErrorLog,
  [int]$TimeoutSeconds,
  [int]$RequestTimeoutSeconds = 5
) {
  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  $lastError = $null
  while ((Get-Date) -lt $deadline) {
    $Process.Refresh()
    if ($Process.HasExited) {
      Write-Fail "$Name exited before it was reachable: $Url"
      Show-LogTail $ErrorLog
      return $false
    }

    try {
      Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec $RequestTimeoutSeconds | Out-Null
      Write-Ok "$Name is reachable: $Url"
      return $true
    } catch {
      $lastError = $_.Exception.Message
      Start-Sleep -Seconds 1
    }
  }

  Write-Warn "$Name is not reachable yet: $Url"
  if (-not [string]::IsNullOrWhiteSpace($lastError)) {
    Write-Warn "Last check error: $lastError"
  }
  Write-Warn "Check log: $ErrorLog"
  return $false
}

function Stop-Dev {
  Write-Section 'Stopping local DEV'

  foreach ($port in $AppPorts) {
    Stop-ListenersOnPort $port
  }

  if (Test-CommandExists 'docker') {
    Push-Location $Root
    try {
      docker compose stop postgres redis site-quality-runner | Out-Host
    } catch {
      Write-Warn "Docker infra was not stopped: $($_.Exception.Message)"
    } finally {
      Pop-Location
    }
  }

  Write-Ok 'Local DEV stopped'
}

if ($PortsOnly) {
  Write-Section 'Local DEV ports'
  Show-Ports
  exit 0
}

if ($StopOnly) {
  Stop-Dev
  exit 0
}

Write-Section 'Starting local DEV'
Write-Host 'Port plan:' -ForegroundColor Cyan
Write-Host "  Storefront Nuxt : http://localhost:$($Ports.Storefront)" -ForegroundColor White
Write-Host "  Go API          : http://localhost:$($Ports.Api)" -ForegroundColor White
Write-Host "  Admin Console   : http://localhost:$($Ports.Admin)" -ForegroundColor White
Write-Host "  Quality Runner  : http://localhost:$($Ports.SiteQualityRunner)/healthz" -ForegroundColor White
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

Write-Section 'Starting PostgreSQL / Redis / Site Quality runner'
Push-Location $Root
try {
  docker compose up -d postgres redis site-quality-runner | Out-Host
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

if (-not (Wait-HttpOk -Url "http://localhost:$($Ports.SiteQualityRunner)/healthz" -TimeoutSeconds 120)) {
  Write-Fail "Site Quality runner is not ready on http://localhost:$($Ports.SiteQualityRunner)/healthz"
  exit 1
}
Write-Ok "Site Quality runner ready: http://localhost:$($Ports.SiteQualityRunner)/healthz"

Write-Section 'Starting Go API'
$corsOrigins = @(
  "http://localhost:$($Ports.Storefront)",
  "http://127.0.0.1:$($Ports.Storefront)",
  "http://localhost:3080",
  "http://127.0.0.1:3080",
  "http://localhost:$($Ports.Admin)",
  "http://127.0.0.1:$($Ports.Admin)",
  "http://localhost:$($Ports.Api)",
  "http://127.0.0.1:$($Ports.Api)"
) -join ','
$backendCommand = @"
`$env:SERVER_PORT=':$($Ports.Api)'
`$env:SERVER_BASE_URL='http://localhost:$($Ports.Api)'
`$env:DB_HOST='localhost'
`$env:DB_PORT='$($Ports.Postgres)'
`$env:DB_USERNAME='commerce_platform'
`$env:DB_PASSWORD='commerce_platform_password'
`$env:DB_NAME='commerce_platform'
`$env:REDIS_HOST='localhost'
`$env:REDIS_PORT='$($Ports.Redis)'
`$env:STOREFRONT_BASE_URL='http://host.docker.internal:$($Ports.Storefront)'
`$env:SITE_QUALITY_RUNNER_URL='http://localhost:$($Ports.SiteQualityRunner)'
`$env:SITE_QUALITY_RUNNER_TOKEN='dev-site-quality-runner-token-0123456789'
`$env:WORKER_SITE_QUALITY_ENABLED='true'
`$env:WORKER_SITE_QUALITY_DISPATCH_INTERVAL_SECONDS='30'
`$env:WORKER_SITE_QUALITY_BATCH_LIMIT='2'
`$env:WORKER_SITE_QUALITY_LEASE_TIMEOUT_SECONDS='900'
`$env:WORKER_SITE_QUALITY_SAMPLE_COUNT='3'
`$env:WORKER_SITE_QUALITY_CONFIRMATIONS='2'
`$env:WORKER_SITE_QUALITY_CLEAN_EVALUATIONS='2'
`$env:WORKER_SITE_QUALITY_PROVIDER_CONCURRENCY='1'
`$env:WORKER_SITE_QUALITY_PROVIDER_SPACING_SECONDS='5'
`$env:PAYMENT_CONFIG_MASTER_KEY='dev-payment-config-master-key-change-before-production'
`$env:CORS_ORIGINS='$corsOrigins'
`$env:STOREFRONT_HTML_CACHE_PURGE_URL='http://localhost:$($Ports.Storefront)/_internal/html-cache/purge'
`$env:STOREFRONT_HTML_CACHE_PURGE_TOKEN='dev-html-cache-purge-token'
go run ./cmd/server
"@
$apiProcess = Start-DevProcess -Name 'Go API' -WorkingDirectory (Join-Path $Root 'go-backend') -Command $backendCommand -LogName 'go-api'

$apiHealthUrl = "http://localhost:$($Ports.Api)/health"
$apiReady = Wait-HttpOk -Url $apiHealthUrl -TimeoutSeconds 90
if ($apiReady) {
  Write-Ok "Go API health check passed: http://localhost:$($Ports.Api)/health"
  Invoke-DevAdminBootstrap
} else {
  Write-Warn "Go API health check is not ready yet. Check log: $(Join-Path $LogDir 'go-api.err.log')"
  Write-Warn 'Skipping DEV admin bootstrap until the API and database are ready.'
}

Write-Section 'Starting Nuxt Storefront'
Clear-StorefrontIconCache
$storefrontCommand = @"
`$env:NUXT_PUBLIC_API_BASE='http://localhost:$($Ports.Api)'
`$env:API_INTERNAL_ORIGIN='http://localhost:$($Ports.Api)'
`$env:NUXT_HTML_CACHE_ENABLED='false'
`$env:NUXT_HTML_CACHE_DRIVER='memory'
`$env:NUXT_HTML_CACHE_PURGE_TOKEN='dev-html-cache-purge-token'
npm run dev
"@
$storefrontProcess = Start-DevProcess -Name 'Nuxt Storefront' -WorkingDirectory (Join-Path $Root 'nuxt-i18n') -Command $storefrontCommand -LogName 'storefront'
$storefrontErrorLog = Join-Path $LogDir 'storefront.err.log'
if (-not (Wait-DevHttpReady -Name 'Nuxt Storefront' -Process $storefrontProcess -Url "http://localhost:$($Ports.Storefront)" -ErrorLog $storefrontErrorLog -TimeoutSeconds 120 -RequestTimeoutSeconds 15)) {
  exit 1
}

Write-Section 'Starting Admin Console'
$adminCommand = @"
`$env:VITE_API_BASE_URL=''
npm run dev
"@
$adminProcess = Start-DevProcess -Name 'Admin Console' -WorkingDirectory (Join-Path $Root 'go-backend/web/admin') -Command $adminCommand -LogName 'admin'
$adminErrorLog = Join-Path $LogDir 'admin.err.log'
if (-not (Wait-DevHttpReady -Name 'Admin Console' -Process $adminProcess -Url "http://localhost:$($Ports.Admin)" -ErrorLog $adminErrorLog -TimeoutSeconds 60)) {
  exit 1
}

Write-Section 'Local DEV is up'
Write-Host "Storefront : http://localhost:$($Ports.Storefront)" -ForegroundColor White
Write-Host "Go API     : http://localhost:$($Ports.Api)" -ForegroundColor White
Write-Host "Admin      : http://localhost:$($Ports.Admin)" -ForegroundColor White
Write-Host "Quality    : http://localhost:$($Ports.SiteQualityRunner)/healthz" -ForegroundColor White
Write-Host "Logs       : $LogDir" -ForegroundColor White
Write-Host ''
Write-Host 'Stop with: npm run dev:stop' -ForegroundColor Yellow
