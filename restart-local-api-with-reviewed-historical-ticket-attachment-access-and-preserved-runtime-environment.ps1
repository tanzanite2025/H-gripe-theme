param(
  [Parameter(Mandatory = $true)][string]$PreparedApiExecutablePath,
  [Parameter(Mandatory = $true)][string]$ReviewedHistoricalAttachmentOrigins,
  [int]$ApiPort = 10241,
  [int]$ExpectedExistingApiProcessId = 0,
  [switch]$ValidateOnly
)

$ErrorActionPreference = 'Stop'
$repositoryDirectory = $PSScriptRoot
$backendDirectory = Join-Path $repositoryDirectory 'go-backend'
$runtimeOutputDirectory = Join-Path $repositoryDirectory 'output/dev'
$preparedExecutable = (Resolve-Path -LiteralPath $PreparedApiExecutablePath).Path
$allowedExecutableDirectory = [IO.Path]::GetFullPath($runtimeOutputDirectory) + [IO.Path]::DirectorySeparatorChar
if (-not $preparedExecutable.StartsWith($allowedExecutableDirectory, [StringComparison]::OrdinalIgnoreCase)) {
  throw 'The prepared API executable must be inside this workspace output/dev directory.'
}
foreach ($reviewedOrigin in $ReviewedHistoricalAttachmentOrigins.Split(',')) {
  $parsedReviewedOrigin = $null
  if ($reviewedOrigin.Trim() -notmatch '\Ahttps?://[^/\\?#\s]+/?\z' -or
      -not [Uri]::TryCreate($reviewedOrigin.Trim(), [UriKind]::Absolute, [ref]$parsedReviewedOrigin) -or
      $parsedReviewedOrigin.Scheme -notin @('http', 'https') -or
      [string]::IsNullOrWhiteSpace($parsedReviewedOrigin.Host) -or
      $parsedReviewedOrigin.Host.Contains('*') -or $parsedReviewedOrigin.UserInfo -or
      $parsedReviewedOrigin.AbsolutePath -ne '/' -or $parsedReviewedOrigin.Query -or
      $reviewedOrigin.Contains('?') -or $reviewedOrigin.Contains('#')) {
    throw 'Reviewed historical attachment sources must be explicit HTTP(S) origins.'
  }
}

$existingListeners = @(Get-NetTCPConnection -LocalPort $ApiPort -State Listen)
$existingProcessIds = @($existingListeners.OwningProcess | Sort-Object -Unique)
if ($existingProcessIds.Count -ne 1) { throw 'Expected one existing local API process.' }
$existingApiProcess = Get-CimInstance Win32_Process -Filter "ProcessId = $($existingProcessIds[0])"
if ($ExpectedExistingApiProcessId -gt 0 -and $existingApiProcess.ProcessId -ne $ExpectedExistingApiProcessId) {
  throw 'The local API process changed; the prepared switch no longer matches its target.'
}
$previousExecutable = $existingApiProcess.ExecutablePath
if (-not $previousExecutable.StartsWith($allowedExecutableDirectory, [StringComparison]::OrdinalIgnoreCase)) {
  throw 'The current API executable belongs to a different workspace.'
}
$existingLauncher = Get-CimInstance Win32_Process -Filter "ProcessId = $($existingApiProcess.ParentProcessId)"
$existingLaunchCommand = $existingLauncher.CommandLine
$encodedLaunchMatch = [regex]::Match($existingLaunchCommand, '(?i)-(?:EncodedCommand|enc)\s+([A-Za-z0-9+/=]+)')
if ($encodedLaunchMatch.Success) {
  $existingLaunchCommand = [Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($encodedLaunchMatch.Groups[1].Value))
}
$environmentAssignments = [regex]::Matches($existingLaunchCommand, '\$env:(?<name>[A-Za-z0-9_]+)\s*=\s*(?:''(?<single>(?:[^'']|'''')*)''|"(?<double>[^"]*)"|(?<plain>[^;\s]+))')
$preservedEnvironment = @{}
foreach ($assignment in $environmentAssignments) {
  $value = if ($assignment.Groups['single'].Success) { $assignment.Groups['single'].Value.Replace("''", "'") }
    elseif ($assignment.Groups['double'].Success) { $assignment.Groups['double'].Value }
    else { $assignment.Groups['plain'].Value }
  $preservedEnvironment[$assignment.Groups['name'].Value] = $value
}
foreach ($requiredName in @('SERVER_PORT', 'SERVER_BASE_URL', 'DB_HOST', 'DB_PORT', 'DB_USERNAME', 'DB_PASSWORD', 'DB_NAME', 'REDIS_HOST', 'REDIS_PORT', 'STOREFRONT_BASE_URL')) {
  if (-not $preservedEnvironment.ContainsKey($requiredName)) { throw "Cannot preserve required runtime variable: $requiredName" }
}
if ($preservedEnvironment.SERVER_PORT -ne ":$ApiPort" -or $preservedEnvironment.DB_HOST -notin @('localhost', '127.0.0.1', '::1')) {
  throw 'This switch supports the reviewed local runtime only.'
}
$preparedExecutableSHA256 = (Get-FileHash -LiteralPath $preparedExecutable -Algorithm SHA256).Hash.ToLowerInvariant()
$previousExecutableSHA256 = (Get-FileHash -LiteralPath $previousExecutable -Algorithm SHA256).Hash.ToLowerInvariant()
$preservedEnvironment.DB_AUTO_MIGRATE = 'false'
$preservedEnvironment.MEDIA_REVIEWED_HISTORICAL_CUSTOMER_SERVICE_ATTACHMENT_ORIGINS = $ReviewedHistoricalAttachmentOrigins

function Start-LocalApiWithPreservedEnvironmentAndHiddenLauncher([string]$Executable, [string]$LogLabel) {
  $launchStatements = [Collections.Generic.List[string]]::new()
  $launchStatements.Add("`$ErrorActionPreference = 'Stop'")
  $launchStatements.Add("Set-Location -LiteralPath '" + $backendDirectory.Replace("'", "''") + "'")
  foreach ($name in ($preservedEnvironment.Keys | Sort-Object)) {
    $launchStatements.Add("`$env:${name}='" + ([string]$preservedEnvironment[$name]).Replace("'", "''") + "'")
  }
  $launchStatements.Add("& '" + $Executable.Replace("'", "''") + "'")
  $launchStatements.Add('exit $LASTEXITCODE')
  $encodedLaunch = [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes(($launchStatements -join "`r`n")))
  $windowsPowerShell = Join-Path $env:SystemRoot 'System32/WindowsPowerShell/v1.0/powershell.exe'
  return Start-Process -FilePath $windowsPowerShell -ArgumentList @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-EncodedCommand', $encodedLaunch) -WorkingDirectory $backendDirectory -WindowStyle Hidden -RedirectStandardOutput (Join-Path $runtimeOutputDirectory "$LogLabel.out.log") -RedirectStandardError (Join-Path $runtimeOutputDirectory "$LogLabel.err.log") -PassThru
}

function Wait-ForExpectedLocalApiExecutableAndHealthyEndpoint([string]$Executable, [int]$TimeoutSeconds = 45) {
  $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
  do {
    $listeners = @(Get-NetTCPConnection -LocalPort $ApiPort -State Listen -ErrorAction SilentlyContinue)
    foreach ($listenerProcessId in @($listeners.OwningProcess | Sort-Object -Unique)) {
      $listenerProcess = Get-CimInstance Win32_Process -Filter "ProcessId = $listenerProcessId"
      if ($listenerProcess.ExecutablePath -ne $Executable) { continue }
      try {
        if ((Invoke-WebRequest -UseBasicParsing -Uri "http://localhost:$ApiPort/health" -TimeoutSec 3).StatusCode -eq 200) {
          return [int]$listenerProcess.ProcessId
        }
      } catch { }
    }
    Start-Sleep -Milliseconds 500
  } while ([DateTime]::UtcNow -lt $deadline)
  throw 'The expected local API executable did not become healthy before the deadline.'
}

$switchEvidence = [ordered]@{
  local_api_port = $ApiPort
  previous_process_id = [int]$existingApiProcess.ProcessId
  previous_executable = $previousExecutable
  previous_executable_sha256 = $previousExecutableSHA256
  prepared_executable = $preparedExecutable
  prepared_executable_sha256 = $preparedExecutableSHA256
  preserved_environment_variable_count = $environmentAssignments.Count
  reviewed_historical_attachment_origins = $ReviewedHistoricalAttachmentOrigins
  database_auto_migrate = $false
  validate_only = [bool]$ValidateOnly
}
if ($ValidateOnly) { $switchEvidence | ConvertTo-Json; exit 0 }

$switchTimestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$newLauncher = $null
try {
  Stop-Process -Id $existingApiProcess.ProcessId
  $newLauncher = Start-LocalApiWithPreservedEnvironmentAndHiddenLauncher $preparedExecutable "go-api-historical-ticket-attachment-access-$switchTimestamp"
  $newApiProcessId = Wait-ForExpectedLocalApiExecutableAndHealthyEndpoint $preparedExecutable
  $switchEvidence.new_launcher_process_id = $newLauncher.Id
  $switchEvidence.new_process_id = $newApiProcessId
  $switchEvidence.health_status = 200
  $switchEvidence.switched_at_utc = [DateTime]::UtcNow.ToString('o')
  $switchEvidence.rollback_performed = $false
  $switchEvidence | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $runtimeOutputDirectory 'historical-ticket-attachment-reviewed-local-runtime-switch-result.json') -Encoding UTF8
  $switchEvidence | ConvertTo-Json
} catch {
  $switchFailure = $_
  if ($null -ne $newLauncher) {
    $newChildren = @(Get-CimInstance Win32_Process -Filter "ParentProcessId = $($newLauncher.Id)")
    foreach ($child in $newChildren) {
      if ($child.ExecutablePath -eq $preparedExecutable) { Stop-Process -Id $child.ProcessId -ErrorAction SilentlyContinue }
    }
    Stop-Process -Id $newLauncher.Id -ErrorAction SilentlyContinue
  }
  $rollbackLauncher = Start-LocalApiWithPreservedEnvironmentAndHiddenLauncher $previousExecutable "go-api-historical-ticket-attachment-access-rollback-$switchTimestamp"
  $restoredProcessId = Wait-ForExpectedLocalApiExecutableAndHealthyEndpoint $previousExecutable
  throw "The new local API did not become healthy; the previous executable was restored as process $restoredProcessId. Failure: $($switchFailure.Exception.Message)"
}
