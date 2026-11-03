param([switch]$Stop, [switch]$PrepareOnly, [switch]$UsePreparedExecutables, [switch]$PrintApplicationBuildIdentity)

$ErrorActionPreference = 'Stop'
$repositoryDirectory = $PSScriptRoot
$backendDirectory = Join-Path $repositoryDirectory 'go-backend'
$runtimeDirectory = Join-Path $repositoryDirectory 'output/dev'
$statePath = Join-Path $runtimeDirectory 'local-development-backup-supervisor-processes.json'
$apiExecutable = Join-Path $runtimeDirectory 'local-development-api-with-module-backup.exe'
$maintenanceExecutable = Join-Path $runtimeDirectory 'local-development-backup-maintenance.exe'
$postgresToolsDirectory = Join-Path $repositoryDirectory 'output/tools/postgresql-15-native-backup-tools'
$privateDirectory = Join-Path $runtimeDirectory 'private-local-development-backup'
$configurationPath = Join-Path $privateDirectory 'local-development-backup-configuration.json'
$stopMarker = Join-Path $privateDirectory 'stop-local-api-supervisor'
$buildIdentityPath = Join-Path $privateDirectory 'local-application-source-and-build-identity.json'

function Get-LocalDevelopmentBackupApplicationSourceBuildIdentity {
  Push-Location $backendDirectory
  try {
    $template = '{{if .Module}}{{if .Module.Main}}{{range .GoFiles}}{{$.Dir}}/{{.}}{{println}}{{end}}{{range .EmbedFiles}}{{$.Dir}}/{{.}}{{println}}{{end}}{{end}}{{end}}'
    $sourceFiles = @(& go list -deps -f $template ./cmd/server)
    if ($LASTEXITCODE -ne 0) { throw 'Cannot determine local application build inputs.' }
    $sourceFiles += @((Join-Path $backendDirectory 'go.mod'), (Join-Path $backendDirectory 'go.sum'))
    $sourceFiles += @(Get-ChildItem -LiteralPath (Join-Path $backendDirectory 'migrations') -Filter '*.sql' -File | Select-Object -ExpandProperty FullName)
    $identities = @($sourceFiles | Where-Object { ![string]::IsNullOrWhiteSpace($_) } | Sort-Object -Unique | ForEach-Object {
      $absolute = [IO.Path]::GetFullPath($_)
      $relative = $absolute.Substring($backendDirectory.Length).Replace([char]92, [char]47)
      $relative + ':' + (Get-FileHash -LiteralPath $absolute -Algorithm SHA256).Hash.ToLowerInvariant()
    })
    $hash = [Security.Cryptography.SHA256]::Create()
    try { $sourceSHA256 = [BitConverter]::ToString($hash.ComputeHash([Text.Encoding]::UTF8.GetBytes(($identities -join [char]10)))).Replace('-', '').ToLowerInvariant() } finally { $hash.Dispose() }
    $version = 'local-backup-source-' + $sourceSHA256.Substring(0, 24)
    if (Test-Path -LiteralPath $buildIdentityPath) {
      $previousBuild = Get-Content -LiteralPath $buildIdentityPath -Raw -Encoding UTF8 | ConvertFrom-Json
      if ($previousBuild.source_sha256 -eq $sourceSHA256 -and $previousBuild.application_version) { $version = $previousBuild.application_version }
    }
    return @{ source_sha256=$sourceSHA256; application_version=$version }
  } finally { Pop-Location }
}

if ($PrintApplicationBuildIdentity) { Get-LocalDevelopmentBackupApplicationSourceBuildIdentity | ConvertTo-Json; return }

if ($Stop) {
  if (Test-Path -LiteralPath (Join-Path $privateDirectory 'maintenance-gate/application-maintenance.json')) { throw 'Maintenance is active; finish backup or recovery before stopping its services.' }
  if (Test-Path -LiteralPath $statePath) {
    $state = Get-Content -LiteralPath $statePath -Raw -Encoding UTF8 | ConvertFrom-Json
    if ($state.repository_directory -ne $repositoryDirectory) { throw 'Backup supervisor belongs to another workspace.' }
    $supervisor = Get-CimInstance Win32_Process -Filter "ProcessId = $($state.supervisor_process_id)" -ErrorAction SilentlyContinue
    $supervisorIdentityMatches = $supervisor -and $state.supervisor_start_time_utc -and $supervisor.CreationDate.ToUniversalTime().ToString('o') -eq $state.supervisor_start_time_utc
    [IO.File]::WriteAllText($stopMarker, 'stop')
    # Stop the supervisor first so it cannot create another API while stopping.
    if ($supervisorIdentityMatches) {
      $supervisorProcess = Get-Process -Id $state.supervisor_process_id -ErrorAction SilentlyContinue
      if ($supervisorProcess) {
        $supervisorProcess.Kill()
        if (!$supervisorProcess.WaitForExit(10000)) { throw 'Local API supervisor did not finish stopping.' }
      }
      foreach ($child in @(Get-CimInstance Win32_Process -Filter "ParentProcessId = $($state.supervisor_process_id)")) {
        if ($child.ExecutablePath -eq $apiExecutable) {
          $apiProcess = Get-Process -Id $child.ProcessId -ErrorAction SilentlyContinue
          if ($apiProcess) {
            $apiProcess.Kill()
            if (!$apiProcess.WaitForExit(10000)) { throw 'Local API did not finish stopping.' }
          }
        }
      }
    }
    $maintenance = Get-CimInstance Win32_Process -Filter "ProcessId = $($state.maintenance_process_id)" -ErrorAction SilentlyContinue
    if ($maintenance -and $maintenance.ExecutablePath -eq $maintenanceExecutable) {
      $maintenanceProcess = Get-Process -Id $maintenance.ProcessId -ErrorAction SilentlyContinue
      if ($maintenanceProcess) {
        $maintenanceProcess.Kill()
        if (!$maintenanceProcess.WaitForExit(10000)) { throw 'Local backup maintenance service did not finish stopping.' }
      }
    }
    Remove-Item -LiteralPath $statePath -Force
  }
  return
}

if (Test-Path -LiteralPath (Join-Path $privateDirectory 'maintenance-gate/application-maintenance.json')) {
  throw 'A backup or recovery maintenance fence is active. Use its existing recovery console before rebuilding or restarting.'
}
New-Item -ItemType Directory -Force -Path $privateDirectory, $runtimeDirectory | Out-Null
# Keep one supervisor per workspace, including during build and preparation.
$supervisorLeasePath = Join-Path $runtimeDirectory 'local-development-api-supervisor-exclusive-start-and-runtime.lock'
try { $supervisorLease = [IO.File]::Open($supervisorLeasePath, 'OpenOrCreate', 'ReadWrite', 'None') }
catch { throw 'A local API backup supervisor is already starting or running. Stop it before starting another.' }
try {
if (!$PrepareOnly) {
  if (Get-NetTCPConnection -LocalPort 19090 -State Listen -ErrorAction SilentlyContinue) { throw 'Port 19090 is already in use; stop the existing backup service before rebuilding.' }
  $apiPort = 0
  if ($env:SERVER_PORT -match ':?(\d+)$') { $apiPort = [int]$Matches[1] }
  if ($apiPort -and (Get-NetTCPConnection -LocalPort $apiPort -State Listen -ErrorAction SilentlyContinue)) { throw 'The configured local API port is already in use; no existing process was replaced.' }
}
# Windows ignores Unix 0600. Limit the entire credential directory to this user.
$identity = [Security.Principal.WindowsIdentity]::GetCurrent().User
$access = [Security.AccessControl.DirectorySecurity]::new()
$access.SetOwner($identity)
$access.SetAccessRuleProtection($true, $false)
$access.AddAccessRule([Security.AccessControl.FileSystemAccessRule]::new($identity, 'FullControl', 'ContainerInherit,ObjectInherit', 'None', 'Allow'))
Set-Acl -LiteralPath $privateDirectory -AclObject $access

if (!$UsePreparedExecutables) {
  $buildIdentity = Get-LocalDevelopmentBackupApplicationSourceBuildIdentity
  Push-Location $backendDirectory
  try {
    go build -ldflags "-X main.Version=$($buildIdentity.application_version)" -o $apiExecutable ./cmd/server
    if ($LASTEXITCODE -ne 0) { throw 'Local API build failed.' }
    go build -o $maintenanceExecutable ./cmd/local-backup-maintenance
    if ($LASTEXITCODE -ne 0) { throw 'Local backup maintenance build failed.' }
    $buildIdentity | ConvertTo-Json | Set-Content -LiteralPath $buildIdentityPath -Encoding UTF8
  } finally { Pop-Location }
}

if (!(Test-Path -LiteralPath (Join-Path $postgresToolsDirectory 'pg_dump.exe'))) {
  if (!(Get-Command pg_dump -ErrorAction SilentlyContinue) -or !(Get-Command pg_restore -ErrorAction SilentlyContinue)) {
    New-Item -ItemType Directory -Force -Path $postgresToolsDirectory | Out-Null
    $archivePath = Join-Path $postgresToolsDirectory 'postgresql-15.10-1-windows-x64-binaries.zip'
    Invoke-WebRequest -UseBasicParsing -Uri 'https://get.enterprisedb.com/postgresql/postgresql-15.10-1-windows-x64-binaries.zip' -OutFile $archivePath -TimeoutSec 600
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $archive = [IO.Compression.ZipFile]::OpenRead($archivePath)
    try {
      foreach ($entry in $archive.Entries) {
        if ($entry.FullName -match '^pgsql/bin/(pg_dump[.]exe|pg_restore[.]exe|[^/]+[.]dll)$') {
          [IO.Compression.ZipFileExtensions]::ExtractToFile($entry, (Join-Path $postgresToolsDirectory $entry.Name), $true)
        }
      }
    } finally { $archive.Dispose() }
  }
}
if (Test-Path -LiteralPath (Join-Path $postgresToolsDirectory 'pg_dump.exe')) { $env:PATH = "$postgresToolsDirectory;$env:PATH" }
& pg_dump --version
if ($LASTEXITCODE -ne 0) { throw 'PostgreSQL backup tools are unavailable.' }
& pg_restore --version
if ($LASTEXITCODE -ne 0) { throw 'PostgreSQL restore tools are unavailable.' }

$env:LOCAL_DEVELOPMENT_BACKUP_CONFIGURATION_PATH = $configurationPath
$env:LOCAL_DEVELOPMENT_BACKUP_STATIC_ASSETS_DIRECTORY = Join-Path $repositoryDirectory 'nuxt-i18n/public'
$env:LOCAL_BACKUP_MAINTENANCE_DIRECTORY = ''
Push-Location $backendDirectory
try { & $apiExecutable prepare-local-development-backup } finally { Pop-Location }
if ($LASTEXITCODE -ne 0) { throw 'Local backup preparation failed; existing business data was retained.' }
$configuration = Get-Content -LiteralPath $configurationPath -Raw -Encoding UTF8 | ConvertFrom-Json
$env:LOCAL_BACKUP_SITE_ID = $configuration.site_id
$env:LOCAL_BACKUP_ENVIRONMENT = $configuration.environment
$env:LOCAL_BACKUP_MAINTENANCE_DIRECTORY = $configuration.maintenance_directory
$env:LOCAL_BACKUP_AUTHENTICATION_KEY_FILE = $configuration.authentication_key_file
$env:OPS_CONNECTOR_MASTER_KEY = $configuration.recovery_secrets.OPS_CONNECTOR_MASTER_KEY
$env:STORAGE_LOCAL_PATH = $configuration.storage_roots.uploads
$env:STORAGE_PRIVATE_LOCAL_PATH = $configuration.storage_roots.uploads_private
$env:SITE_LOGO_STORAGE_LOCAL_PATH = $configuration.storage_roots.site_logo_uploads
$env:DB_AUTO_MIGRATE = 'false'
# The native development server serves this workspace's company directory.
# Stamp its actual local build when starting that build, as the image build does.
$companyReleaseMarker = Join-Path $env:LOCAL_DEVELOPMENT_BACKUP_STATIC_ASSETS_DIRECTORY 'company/.storefront-release-version'
[IO.File]::WriteAllText($companyReleaseMarker, $configuration.application_version)
& $maintenanceExecutable -config $configurationPath -check-configuration-only
if ($LASTEXITCODE -ne 0) { throw 'Independent backup configuration validation failed.' }
if ($PrepareOnly) { return }

if (Get-NetTCPConnection -LocalPort 19090 -State Listen -ErrorAction SilentlyContinue) { throw 'Port 19090 is already in use; no existing process was replaced.' }
Remove-Item -LiteralPath $stopMarker -Force -ErrorAction SilentlyContinue
$maintenanceProcess = Start-Process -FilePath $maintenanceExecutable -ArgumentList @('-config', ('"' + $configurationPath + '"'), '-listen', '127.0.0.1:19090') -WorkingDirectory $backendDirectory -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $runtimeDirectory 'local-backup-maintenance.out.log') -RedirectStandardError (Join-Path $runtimeDirectory 'local-backup-maintenance.err.log')
$supervisorStartTime = (Get-CimInstance Win32_Process -Filter "ProcessId = $PID").CreationDate.ToUniversalTime().ToString('o')
@{ repository_directory=$repositoryDirectory; supervisor_process_id=$PID; supervisor_start_time_utc=$supervisorStartTime; maintenance_process_id=$maintenanceProcess.Id; private_configuration_directory=$privateDirectory } | ConvertTo-Json | Set-Content -LiteralPath $statePath -Encoding UTF8
Write-Host 'Local module backup service: http://127.0.0.1:19090/local-backup-maintenance/'
Write-Host "Local recovery credentials (private): $privateDirectory"
Push-Location $backendDirectory
try {
  while (!(Test-Path -LiteralPath $stopMarker)) {
    & $apiExecutable
    if ($LASTEXITCODE -ne 0) { throw 'Local API exited unexpectedly. Check the API log; automatic crash looping was stopped.' }
    while (!(Test-Path -LiteralPath $stopMarker) -and (Test-Path -LiteralPath (Join-Path $env:LOCAL_BACKUP_MAINTENANCE_DIRECTORY 'application-maintenance.json'))) { Start-Sleep -Milliseconds 500 }
    Start-Sleep -Milliseconds 500
  }
} finally { Pop-Location }
} finally { $supervisorLease.Dispose() }
