'use strict';

const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { test } = require('node:test');

const repository = path.resolve(__dirname, '..');
const bash = process.env.LOCAL_BACKUP_VALIDATION_BASH || (process.platform === 'win32' ? 'C:/Program Files/Git/bin/bash.exe' : 'bash');
const guardScript = path.join(__dirname, 'local-computer-backup-release-overlay-selection-and-existing-maintenance-gate-protection.sh');
const syntheticImageTag = 'sha-' + 'a'.repeat(40);

function createSyntheticBackupReleaseFixture(t) {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'local-computer-backup-release-validation-'));
  t.after(() => fs.rmSync(directory, { recursive: true, force: true }));
  for (const name of ['gate', 'previous-gate', 'workspace']) fs.mkdirSync(path.join(directory, name));
  return directory;
}

function createIsolatedComposeValidationEnvironment() {
  const environment = {};
  for (const name of ['PATH', 'Path', 'SystemRoot', 'SYSTEMROOT', 'WINDIR', 'TEMP', 'TMP', 'HOME', 'USERPROFILE', 'APPDATA', 'LOCALAPPDATA', 'ProgramFiles', 'ProgramData', 'DOCKER_CONFIG', 'DOCKER_CLI_PLUGIN_EXTRA_DIRS']) {
    if (process.env[name] !== undefined) environment[name] = process.env[name];
  }
  return environment;
}

function runSyntheticProductionComposeConfiguration(t, directory, enabled, sentinel, missingGate) {
  let contents = fs.readFileSync(path.join(__dirname, 'production.env.example'), 'utf8')
    .replace(/^IMAGE_NAMESPACE=.*$/m, 'IMAGE_NAMESPACE=synthetic-backup-release-validation')
    .replace(/^IMAGE_TAG=.*$/m, 'IMAGE_TAG=' + syntheticImageTag)
    .replace(/^REDIS_MODE=.*$/m, 'REDIS_MODE=' + (sentinel ? 'sentinel' : 'standalone'));
  const fields = {
    LOCAL_BACKUP_MAINTENANCE_ENABLED: String(enabled), LOCAL_BACKUP_SITE_ID: 'synthetic-release-store', LOCAL_BACKUP_ENVIRONMENT: 'test',
    LOCAL_BACKUP_GATE_DIRECTORY: missingGate ? '' : path.join(directory, 'gate').replace(/\\/g, '/'),
    LOCAL_BACKUP_WORKSPACE_DIRECTORY: path.join(directory, 'workspace').replace(/\\/g, '/'),
    LOCAL_BACKUP_CONFIGURATION_FILE: path.join(directory, 'synthetic-private-configuration.json').replace(/\\/g, '/'),
    LOCAL_BACKUP_AUTHENTICATION_KEY_FILE: path.join(directory, 'synthetic-authentication').replace(/\\/g, '/'),
    LOCAL_BACKUP_PACKAGE_PASSPHRASE_FILE: path.join(directory, 'synthetic-passphrase').replace(/\\/g, '/'),
  };
  for (const [key, value] of Object.entries(fields)) contents = contents.replace(new RegExp('^' + key + '=.*$', 'm'), key + '=' + value);
  const environmentFile = path.join(directory, 'synthetic-release.env');
  fs.writeFileSync(environmentFile, contents, { mode: 0o600 });
  const composeArguments = ['compose', '--env-file', environmentFile, '--project-name', 'codex-local-backup-release-validation', '-f', path.join(repository, 'compose.prod.yml')];
  if (sentinel) composeArguments.push('-f', path.join(__dirname, 'redis-sentinel.compose.yml'));
  if (enabled) composeArguments.push('-f', path.join(repository, 'compose.local-computer-backup-maintenance.yml'));
  composeArguments.push('config', '--format', 'json');
  const result = spawnSync('docker', composeArguments, { cwd: directory, env: createIsolatedComposeValidationEnvironment(), encoding: 'utf8', timeout: 30000, maxBuffer: 2 << 20 });
  if (missingGate) {
    assert.notEqual(result.status, 0, 'enabled backup without an explicit gate must be rejected');
    assert.match(result.stderr, /LOCAL_BACKUP_GATE_DIRECTORY/);
    return;
  }
  assert.equal(result.status, 0, 'synthetic Compose config must succeed; no services are started');
  return JSON.parse(result.stdout);
}

test('production Compose preserves optional backup boundaries for standalone and Sentinel releases', async t => {
  for (const enabled of [false, true]) for (const sentinel of [false, true]) {
    await t.test(`backup=${enabled}, sentinel=${sentinel}`, t => {
      const directory = createSyntheticBackupReleaseFixture(t);
      const configuration = runSyntheticProductionComposeConfiguration(t, directory, enabled, sentinel, false);
      const services = configuration.services;
      assert.equal(Boolean(services['local-backup-maintenance']), enabled);
      assert.equal(Boolean(services['redis-sentinel-1']), sentinel);
      assert.equal(services.api.image.split(':').at(-1), syntheticImageTag);
      if (!enabled) {
        assert.equal(services.api.environment.LOCAL_BACKUP_MAINTENANCE_DIRECTORY, undefined);
        return;
      }
      const maintenance = services['local-backup-maintenance'];
      assert.equal(maintenance.image, `ghcr.io/synthetic-backup-release-validation/commerce-platform-local-backup-maintenance:${syntheticImageTag}`);
      assert.equal(maintenance.build.target, 'local-backup-maintenance');
      assert.equal(maintenance.pull_policy, 'always');
      assert.deepEqual(Object.keys(maintenance.networks).sort(), ['app', 'db']);
      assert.equal(maintenance.depends_on, undefined);
      assert.equal(maintenance.read_only, true);
      const servingStatic = services.storefront.volumes.find(mount => mount.target === '/app/.output/public/company');
      const backupStatic = maintenance.volumes.find(mount => mount.target === '/app/product-static-assets/company');
      assert.equal(servingStatic.source, backupStatic.source);
      assert.equal(servingStatic.read_only, true);
      assert.notEqual(backupStatic.read_only, true);
      assert.equal(backupStatic.volume.nocopy, true);
      assert(configuration.volumes[servingStatic.source].name.endsWith('-' + syntheticImageTag));
      assert.equal(maintenance.environment.SERVER_BASE_URL, services.api.environment.SERVER_BASE_URL);
      assert.equal(maintenance.environment.STOREFRONT_BASE_URL, services.api.environment.STOREFRONT_BASE_URL);
      assert.equal(maintenance.ports.length, 1);
      assert.equal(maintenance.ports[0].host_ip, '127.0.0.1');
      assert.equal(maintenance.ports[0].target, 19090);
      const gates = ['api', 'migrate', 'media-derivatives-backfill', 'local-backup-maintenance'].map(name => {
        const mounts = services[name].volumes.filter(mount => mount.target === '/app/local-backup-maintenance-gate');
        assert.equal(mounts.length, 1);
        assert.equal(mounts[0].type, 'bind');
        assert.notEqual(mounts[0].read_only, true);
        if (name !== 'local-backup-maintenance') assert.equal(services[name].environment.LOCAL_BACKUP_MAINTENANCE_DIRECTORY, mounts[0].target);
        return mounts[0].source;
      });
      assert.equal(new Set(gates).size, 1);
      for (const [apiTarget, maintenanceTarget] of [['/app/uploads', '/app/uploads'], ['/app/uploads-private', '/app/uploads-private'], ['/app/site-logo-uploads', '/app/site-logo-uploads']]) {
        assert.equal(services.api.volumes.find(mount => mount.target === apiTarget).source, maintenance.volumes.find(mount => mount.target === maintenanceTarget).source);
      }
      const databaseCapacity = maintenance.volumes.find(mount => mount.target === '/app/local-backup-postgresql-capacity');
      assert.equal(databaseCapacity.read_only, true);
      assert.equal(databaseCapacity.source, services.db.volumes.find(mount => mount.target === '/var/lib/postgresql/data').source);
    });
  }
});

test('enabled maintenance configuration rejects a missing explicit host gate', t => {
  runSyntheticProductionComposeConfiguration(t, createSyntheticBackupReleaseFixture(t), true, false, true);
});

test('the actual release preflight rejects exposed ports and mismatched maintenance gates', async t => {
  const directory = createSyntheticBackupReleaseFixture(t);
  const configuration = runSyntheticProductionComposeConfiguration(t, directory, true, false, false);
  const verifier = fs.readFileSync(path.join(__dirname, 'verify-vps-release-boundary.sh'), 'utf8');
  const verifierProgram = [...verifier.matchAll(/<<'PY'\r?\n([\s\S]*?)\r?\nPY/g)].find(match => match[1].includes('expected_service_networks ='))[1];
  const pythonCommand = process.env.LOCAL_BACKUP_VALIDATION_PYTHON_EXECUTABLE || (process.platform === 'win32' ? 'py' : 'python3');
  const pythonPrefix = process.platform === 'win32' && !process.env.LOCAL_BACKUP_VALIDATION_PYTHON_EXECUTABLE ? ['-3'] : [];
  const modifications = {
    valid: () => {},
    public_maintenance_port: data => { data.services['local-backup-maintenance'].ports[0].host_ip = '0.0.0.0'; },
    maintenance_port_range: data => { data.services['local-backup-maintenance'].ports[0].published = '19090-19091'; },
    extra_service_exposes_port: data => { data.services['redis-sentinel-extra'] = { networks: { cache: {} }, ports: [{ host_ip: '127.0.0.1', published: '26379', target: 26379 }] }; },
    different_migration_gate: data => { data.services.migrate.volumes.find(mount => mount.target === '/app/local-backup-maintenance-gate').source += '-different'; },
    readonly_application_gate: data => { data.services.api.volumes.find(mount => mount.target === '/app/local-backup-maintenance-gate').read_only = true; },
    maintenance_depends_on_business: data => { data.services['local-backup-maintenance'].depends_on = { api: { condition: 'service_started' } }; },
    missing_existing_server_url: data => { delete data.services['local-backup-maintenance'].environment.SERVER_BASE_URL; },
    different_storefront_url: data => { data.services['local-backup-maintenance'].environment.STOREFRONT_BASE_URL = 'https://different-storefront.example.invalid'; },
    missing_company_static_volume: data => { data.services['local-backup-maintenance'].volumes = data.services['local-backup-maintenance'].volumes.filter(mount => mount.target !== '/app/product-static-assets/company'); },
    different_company_static_volume: data => { data.services['local-backup-maintenance'].volumes.find(mount => mount.target === '/app/product-static-assets/company').source = 'other-storefront-static'; },
    readonly_company_static_recovery: data => { data.services['local-backup-maintenance'].volumes.find(mount => mount.target === '/app/product-static-assets/company').read_only = true; },
    maintenance_initializes_static_volume: data => { data.services['local-backup-maintenance'].volumes.find(mount => mount.target === '/app/product-static-assets/company').volume.nocopy = false; },
    mutable_storefront_image: data => { data.services.storefront.image = data.services.storefront.image.replace(syntheticImageTag, 'master'); },
    unversioned_static_volume: data => { data.volumes.storefront_company_static_assets.name = 'unversioned-storefront-static'; },
  };
  for (const [scenario, modify] of Object.entries(modifications)) await t.test(scenario, () => {
    const data = structuredClone(configuration);
    modify(data);
    const configurationFile = path.join(directory, 'synthetic-resolved-compose.json');
    fs.writeFileSync(configurationFile, JSON.stringify(data), { mode: 0o600 });
    const result = spawnSync(pythonCommand, [...pythonPrefix, '-', configurationFile], { input: verifierProgram, encoding: 'utf8', timeout: 10000 });
    assert.equal(result.status === 0, scenario === 'valid', 'unexpected actual release preflight decision: ' + result.stderr);
    if (scenario !== 'valid') assert.match(result.stderr, /FAIL:/);
  });
  await t.test('actual deployment metadata selects only project and host gate', () => {
    const deploy = fs.readFileSync(path.join(repository, 'deploy.sh'), 'utf8');
    const metadataProgram = deploy.match(/local_backup_release_metadata="\$\([\s\S]*? -c '\r?\n([\s\S]*?)\r?\n'\)"/)[1];
    const result = spawnSync(pythonCommand, [...pythonPrefix, '-c', metadataProgram], { input: JSON.stringify(configuration), encoding: 'utf8', timeout: 10000 });
    assert.equal(result.status, 0, result.stderr);
    const output = result.stdout.trim().split(/\r?\n/);
    assert.equal(output.length, 2);
    assert.equal(output[0], configuration.name);
    assert.equal(output[1], configuration.services.api.volumes.find(mount => mount.target === '/app/local-backup-maintenance-gate').source);
    assert.doesNotMatch(result.stdout, /CHANGE_ME|JWT_SECRET|PAYMENT_CONFIG_MASTER_KEY|DB_PASSWORD/);
  });
});

const syntheticDockerInspectionHarness = `
set -euo pipefail
source "$1"
cd "$2"
fixture_directory="$(pwd -P)"
selected_gate="$fixture_directory/gate"
if [[ "$LOCAL_BACKUP_TEST_ENABLED" == false ]]; then selected_gate=-; fi
if [[ "$LOCAL_BACKUP_TEST_SCENARIO" == missing_gate ]]; then selected_gate="$fixture_directory/missing"; fi
if [[ "$LOCAL_BACKUP_TEST_SCENARIO" == active_fence ]]; then : > "$fixture_directory/gate/application-maintenance.json"; fi
docker() {
  printf '%s\\n' "$1" >> "$fixture_directory/mock-docker-inspection.log"
  case "$1" in
    ps)
      [[ "$*" == 'ps --all --quiet --filter label=com.docker.compose.project=synthetic-release-project' ]]
      [[ "$LOCAL_BACKUP_TEST_SCENARIO" != docker_query_failure ]] || return 7
      if [[ "$LOCAL_BACKUP_TEST_EXISTING" == true ]]; then printf '%s\\n' synthetic-stopped-container; fi
      ;;
    inspect)
      [[ "$LOCAL_BACKUP_TEST_SCENARIO" != inspection_failure ]] || return 8
      [[ "$4" == synthetic-stopped-container ]]
      if [[ "$3" == *com.docker.compose.service* ]]; then
        if [[ "$LOCAL_BACKUP_TEST_SCENARIO" == stopped_maintenance_without_gate ]]; then printf '%s\\n' local-backup-maintenance; else printf '%s\\n' api; fi
      elif [[ "$3" == *Mounts* ]]; then
        if [[ "$LOCAL_BACKUP_TEST_SCENARIO" == changed_gate ]]; then printf '%s\\n' "$fixture_directory/previous-gate";
        elif [[ "$LOCAL_BACKUP_TEST_SCENARIO" != stopped_maintenance_without_gate ]]; then printf '%s\\n' "$fixture_directory/gate"; fi
      else return 91; fi
      ;;
    *) return 92 ;;
  esac
}
verify_existing_local_computer_backup_gate_before_application_release synthetic-release-project "$LOCAL_BACKUP_TEST_ENABLED" "$selected_gate"
`;

test('release guard preserves existing gates and refuses omission, replacement, fences and failed inspection', async t => {
  const scenarios = [
    ['disabled_new_project', false, false, true], ['enabled_new_project', true, false, true],
    ['disabled_existing_gate', false, true, false], ['stopped_maintenance_without_gate', false, true, false],
    ['enabled_existing_gate', true, true, true], ['changed_gate', true, true, false],
    ['active_fence', true, false, false], ['missing_gate', true, false, false],
    ['inspection_failure', true, true, false], ['docker_query_failure', true, true, false],
  ];
  for (const [scenario, enabled, existing, expectedSuccess] of scenarios) await t.test(scenario, t => {
    const directory = createSyntheticBackupReleaseFixture(t);
    const result = spawnSync(bash, ['-c', syntheticDockerInspectionHarness, 'synthetic-backup-release-guard', guardScript, directory], {
      cwd: directory, env: { ...createIsolatedComposeValidationEnvironment(), LOCAL_BACKUP_TEST_SCENARIO: scenario, LOCAL_BACKUP_TEST_ENABLED: String(enabled), LOCAL_BACKUP_TEST_EXISTING: String(existing) },
      encoding: 'utf8', timeout: 10000,
    });
    assert.equal(result.status === 0, expectedSuccess, 'unexpected guard decision for ' + scenario + ': ' + result.stderr);
    assert.equal(fs.existsSync(path.join(directory, 'gate', 'application-maintenance.json')), scenario === 'active_fence', 'persistent fence must remain untouched');
    const logFile = path.join(directory, 'mock-docker-inspection.log');
    if (fs.existsSync(logFile)) for (const command of fs.readFileSync(logFile, 'utf8').trim().split('\n')) assert.ok(['ps', 'inspect'].includes(command));
  });
});

test('release overlay selection is explicit and keeps earlier Compose files', async t => {
  for (const setting of ['false', 'true', 'invalid']) await t.test(setting, t => {
    const directory = createSyntheticBackupReleaseFixture(t);
    const script = 'set -euo pipefail; source "$1"; compose=(docker compose -f compose.prod.yml -f redis-sentinel.compose.yml); append_local_computer_backup_overlay_for_explicit_release_setting "$2" "$3" compose; printf "%s\\n" "${compose[@]}"';
    const result = spawnSync(bash, ['-c', script, 'synthetic-overlay-selection', guardScript, repository, setting], { encoding: 'utf8', timeout: 10000 });
    assert.equal(result.status === 0, setting !== 'invalid');
    if (setting === 'invalid') return;
    const composeArguments = result.stdout.trim().split('\n');
    assert.deepEqual(composeArguments.slice(0, 6), ['docker', 'compose', '-f', 'compose.prod.yml', '-f', 'redis-sentinel.compose.yml']);
    assert.equal(composeArguments.length, setting === 'true' ? 8 : 6);
    if (setting === 'true') assert.match(composeArguments[7], /compose\.local-computer-backup-maintenance\.yml$/);
  });
  await t.test('enabled release missing overlay', t => {
    const directory = createSyntheticBackupReleaseFixture(t);
    const result = spawnSync(bash, ['-c', 'set -euo pipefail; source "$1"; compose=(docker compose); append_local_computer_backup_overlay_for_explicit_release_setting "$2" true compose', 'synthetic-missing-overlay', guardScript, directory], { encoding: 'utf8', timeout: 10000 });
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /does not include/);
  });
});

test('published maintenance image uses its isolated Docker target and the shared scan-and-push gate', () => {
  const workflow = fs.readFileSync(path.join(repository, '.github/workflows/publish-images.yml'), 'utf8');
  assert.match(workflow, /image: commerce-platform-local-backup-maintenance\s+context: \.\/go-backend\s+dockerfile: \.\/go-backend\/Dockerfile\s+target: local-backup-maintenance/);
  assert.match(workflow, /target: \$\{\{ matrix\.target \|\| '' \}\}/);
  assert.match(workflow, /local\/\$\{\{ matrix\.image \}\}:\$\{\{ github\.sha \}\}/);
  assert.match(workflow, /name: Block release on high or critical findings/);
  const deploy = fs.readFileSync(path.join(repository, 'deploy.sh'), 'utf8');
  assert.ok(deploy.indexOf('verify_existing_local_computer_backup_gate_before_application_release "') < deploy.indexOf('pull_succeeded=false'));
  assert.match(deploy, /"\$\{compose\[@\]\}" up -d --remove-orphans api storefront admin web/);
  assert.doesNotMatch(deploy, /"\$\{compose\[@\]\}" (?:up|run|restart)[^\n]*local-backup-maintenance/);
});

test('deployment and preflight share ordered Redis and backup overlay selection', async t => {
  const verifier = fs.readFileSync(path.join(__dirname, 'verify-vps-release-boundary.sh'), 'utf8');
  const deploy = fs.readFileSync(path.join(repository, 'deploy.sh'), 'utf8');
  for (const script of [verifier, deploy]) {
    assert.ok(script.indexOf('append_redis_sentinel_overlay_for_explicit_release_mode "') < script.indexOf('append_local_computer_backup_overlay_for_explicit_release_setting "'));
  }
  for (const mode of ['', 'standalone', 'single', 'sentinel', 'failover', 'SeNtInEl', 'invalid']) await t.test(mode || 'empty defaults to standalone', () => {
    const script = 'set -euo pipefail; source "$1"; compose=(docker compose -f compose.prod.yml); append_redis_sentinel_overlay_for_explicit_release_mode "$2" "$3" compose; append_local_computer_backup_overlay_for_explicit_release_setting "$2" true compose; printf "%s\\n" "${compose[@]}"';
    const result = spawnSync(bash, ['-c', script, 'synthetic-redis-overlay-selection', guardScript, repository, mode], { encoding: 'utf8', timeout: 10000 });
    assert.equal(result.status === 0, mode !== 'invalid', result.stderr);
    if (mode === 'invalid') return;
    const argumentsList = result.stdout.trim().split('\n');
    const sentinel = ['sentinel', 'failover'].includes(mode.toLowerCase());
    assert.equal(argumentsList.length, sentinel ? 8 : 6);
    if (sentinel) assert.match(argumentsList[5], /redis-sentinel\.compose\.yml$/);
    assert.match(argumentsList.at(-1), /compose\.local-computer-backup-maintenance\.yml$/);
  });
});

const syntheticFullPreflightDockerQueryHarness = `
set -euo pipefail
python3() { "$LOCAL_BACKUP_TEST_PYTHON_EXECUTABLE" "$@"; }
docker() {
  printf '%s\\n' "$*" >> "$LOCAL_BACKUP_TEST_FIXTURE/mock-docker-preflight.log"
  if [[ "$1" == compose ]]; then
    shift
    if [[ "$1" == version ]]; then return 0; fi
    [[ "$*" == *' --env-file '* || "$1" == --env-file ]] || return 91
    case " $* " in
      *' config --format json '*)
        if [[ "$LOCAL_BACKUP_TEST_SENTINEL" == true ]]; then [[ "$*" == *redis-sentinel.compose.yml* ]] || return 92;
        else [[ "$*" != *redis-sentinel.compose.yml* ]] || return 93; fi
        if [[ "$LOCAL_BACKUP_TEST_ENABLED" == true ]]; then [[ "$*" == *compose.local-computer-backup-maintenance.yml* ]] || return 94;
        else [[ "$*" != *compose.local-computer-backup-maintenance.yml* ]] || return 95; fi
        cat "$LOCAL_BACKUP_TEST_FIXTURE/synthetic-resolved-compose.json"
        ;;
      *' ps --all --format json '*)
        [[ "$LOCAL_BACKUP_TEST_SCENARIO" != failed_container_query ]] || return 7
        cat "$LOCAL_BACKUP_TEST_FIXTURE/synthetic-runtime-compose-ps.json"
        ;;
      *) return 96 ;;
    esac
  elif [[ "$1" == network && "$2" == inspect ]]; then
    [[ "$LOCAL_BACKUP_TEST_SCENARIO" != failed_network_query ]] || return 8
    cat "$LOCAL_BACKUP_TEST_FIXTURE/synthetic-runtime-networks.json"
  elif [[ "$1" == run && "$CHECK_CONNECTIVITY" == true ]]; then
    # Simulate the existing four probes; never invoke an actual Docker process.
    if [[ "$4" == *_app ]]; then return 1; fi
    return 0
  else
    return 97
  fi
}
source "$1"
`;

test('complete release preflight separates read-only runtime metadata from active probes', async t => {
  const fixturesDirectory = createSyntheticBackupReleaseFixture(t);
  const configurations = new Map();
  for (const enabled of [false, true]) for (const sentinel of [false, true]) {
    const configuration = runSyntheticProductionComposeConfiguration(t, fixturesDirectory, enabled, sentinel, false);
    if (enabled) configuration.services['local-backup-maintenance'].ports[0].published = '29090';
    configurations.set(`${enabled}:${sentinel}`, { configuration, environmentContents: fs.readFileSync(path.join(fixturesDirectory, 'synthetic-release.env'), 'utf8').replace(/^LOCAL_BACKUP_LOOPBACK_PORT=.*$/m, 'LOCAL_BACKUP_LOOPBACK_PORT=29090') });
  }
  const pythonCommand = process.env.LOCAL_BACKUP_VALIDATION_PYTHON_EXECUTABLE || (process.platform === 'win32' ? 'py' : 'python3');
  // The full shell must resolve a real interpreter, not the Windows py launcher.
  const pythonResolution = spawnSync(pythonCommand, [...(process.platform === 'win32' && !process.env.LOCAL_BACKUP_VALIDATION_PYTHON_EXECUTABLE ? ['-3'] : []), '-c', 'import sys; print(sys.executable)'], { encoding: 'utf8', timeout: 10000 });
  assert.equal(pythonResolution.status, 0, pythonResolution.stderr);
  const pythonExecutable = pythonResolution.stdout.trim().replace(/\\/g, '/');
  const cases = [
    ...[false, true].flatMap(enabled => [false, true].flatMap(sentinel => ['static', 'readonly'].map(mode => ({ name: `${mode}_backup_${enabled}_sentinel_${sentinel}`, enabled, sentinel, mode, succeeds: true })))),
    ...['ndjson', 'single_object', 'lowercase_fields'].map(name => ({ name, mode: 'readonly', succeeds: true })),
    { name: 'explicit_active_probes', mode: 'active', succeeds: true },
    ...['public_ipv4', 'public_ipv6', 'different_loopback', 'different_host_port', 'different_target_port', 'udp_port', 'business_loopback_port', 'missing_project', 'different_project', 'empty_runtime', 'network_boundary_mismatch', 'missing_network', 'failed_container_query', 'failed_network_query', 'malformed_port', 'orphan_maintenance_without_overlay', 'invalid_metadata_flag', 'invalid_connectivity_flag', 'existing_report_directory'].map(name => ({ name, mode: 'readonly', succeeds: false })),
  ];
  for (const scenario of cases) await t.test(scenario.name, t => {
    const directory = createSyntheticBackupReleaseFixture(t);
    const enabled = scenario.name === 'orphan_maintenance_without_overlay' ? false : scenario.enabled ?? true;
    const sentinel = scenario.sentinel ?? true;
    const { configuration, environmentContents } = configurations.get(`${enabled}:${sentinel}`);
    const reportDirectory = path.join(directory, 'report');
    let records = ['db', 'api', 'migrate', ...(enabled ? ['local-backup-maintenance'] : [])].map((name, index) => ({
      ID: 'synthetic-container-' + index, Name: 'synthetic-' + name, Image: configuration.services[name].image,
      Project: configuration.name, Service: name, State: name === 'migrate' ? 'exited' : 'running', Health: '', ExitCode: 0,
      Command: 'synthetic-command-secret-must-not-be-exported', Labels: { CREDENTIAL: 'synthetic-label-secret-must-not-be-exported' },
      Publishers: name === 'local-backup-maintenance' ? [{ URL: '127.0.0.1', PublishedPort: 29090, TargetPort: 19090, Protocol: 'tcp', CREDENTIAL: 'synthetic-publisher-secret-must-not-be-exported' }] : [],
    }));
    let networks = Object.values(configuration.networks).map(network => ({ Name: network.name, Id: 'synthetic-network-' + network.name, Driver: 'bridge', Scope: 'local', Internal: network.internal || false, EnableIPv6: false, Labels: { TOKEN: 'synthetic-network-secret-must-not-be-exported' }, Containers: { unrelated: { Name: 'must-not-be-exported' } } }));
    const maintenanceRecord = records.find(record => record.Service === 'local-backup-maintenance');
    const binding = maintenanceRecord?.Publishers[0];
    switch (scenario.name) {
      case 'public_ipv4': binding.URL = '0.0.0.0'; break;
      case 'public_ipv6': binding.URL = '::'; break;
      case 'different_loopback': binding.URL = '127.0.0.2'; break;
      case 'different_host_port': binding.PublishedPort = 19090; break;
      case 'different_target_port': binding.TargetPort = 19091; break;
      case 'udp_port': binding.Protocol = 'udp'; break;
      case 'malformed_port': binding.PublishedPort = true; break;
      case 'business_loopback_port': records[1].Publishers = [{ URL: '127.0.0.1', PublishedPort: 29090, TargetPort: 19090, Protocol: 'tcp' }]; break;
      case 'missing_project': delete records[0].Project; break;
      case 'different_project': records[0].Project += '-other'; break;
      case 'empty_runtime': records = []; break;
      case 'network_boundary_mismatch': networks.find(network => network.Name === configuration.networks.db.name).Internal = false; break;
      case 'missing_network': networks.pop(); break;
      case 'orphan_maintenance_without_overlay': records.push({ Project: configuration.name, Service: 'local-backup-maintenance', Publishers: [{ URL: '127.0.0.1', PublishedPort: 29090, TargetPort: 19090, Protocol: 'tcp' }] }); break;
      case 'single_object': records = [maintenanceRecord]; break;
      case 'lowercase_fields': records = records.map(record => Object.fromEntries(Object.entries(record).map(([key, value]) => [key.toLowerCase(), key === 'Publishers' ? value.map(port => ({ url: port.URL, published_port: port.PublishedPort, target_port: port.TargetPort, protocol: port.Protocol })) : value]))); break;
      case 'existing_report_directory': fs.mkdirSync(reportDirectory); fs.writeFileSync(path.join(reportDirectory, 'compose-config.json'), 'preserve-previous-evidence'); break;
    }
    const runtimeJson = scenario.name === 'ndjson' ? records.map(record => JSON.stringify(record)).join('\n') : JSON.stringify(scenario.name === 'single_object' ? records[0] : records);
    fs.writeFileSync(path.join(directory, 'synthetic-resolved-compose.json'), JSON.stringify(configuration));
    fs.writeFileSync(path.join(directory, 'synthetic-runtime-compose-ps.json'), runtimeJson);
    fs.writeFileSync(path.join(directory, 'synthetic-runtime-networks.json'), JSON.stringify(networks));
    fs.writeFileSync(path.join(directory, 'synthetic-release.env'), environmentContents);
    const result = spawnSync(bash, ['-c', syntheticFullPreflightDockerQueryHarness, 'synthetic-full-release-preflight', path.join(__dirname, 'verify-vps-release-boundary.sh')], {
      cwd: directory, encoding: 'utf8', timeout: 15000,
      env: { ...createIsolatedComposeValidationEnvironment(), ENV_FILE: path.join(directory, 'synthetic-release.env').replace(/\\/g, '/'), REPORT_DIR: reportDirectory.replace(/\\/g, '/'), CHECK_RUNTIME_METADATA: scenario.name === 'invalid_metadata_flag' ? 'invalid' : String(scenario.mode === 'readonly'), CHECK_CONNECTIVITY: scenario.name === 'invalid_connectivity_flag' ? 'invalid' : String(scenario.mode === 'active'), LOCAL_BACKUP_TEST_FIXTURE: directory.replace(/\\/g, '/'), LOCAL_BACKUP_TEST_PYTHON_EXECUTABLE: pythonExecutable, LOCAL_BACKUP_TEST_SENTINEL: String(sentinel), LOCAL_BACKUP_TEST_ENABLED: String(enabled), LOCAL_BACKUP_TEST_SCENARIO: scenario.name },
    });
    assert.equal(result.status === 0, scenario.succeeds, `${scenario.name}: ${result.stdout}\n${result.stderr}`);
    const logFile = path.join(directory, 'mock-docker-preflight.log');
    const queries = fs.existsSync(logFile) ? fs.readFileSync(logFile, 'utf8').trim().split('\n') : [];
    const activeProbeQueries = queries.filter(query => query.startsWith('run '));
    assert.equal(activeProbeQueries.length, scenario.mode === 'active' ? 4 : 0, 'only explicit connectivity mode may request probe containers');
    assert.ok(queries.every(query => /^(compose |network inspect |run )/.test(query)), 'preflight must not start business services or query the database');
    if (scenario.mode === 'static') assert.ok(queries.every(query => !/ ps |^network |^run /.test(query)), 'static mode must not query runtime resources');
    if (scenario.succeeds && scenario.mode !== 'static') {
      const summary = JSON.parse(fs.readFileSync(path.join(reportDirectory, 'runtime-metadata-boundary-summary.json'), 'utf8'));
      assert.equal(summary.target_environment_acceptance_complete, false);
      assert.equal(summary.runtime_metadata_boundary_valid, true);
      assert.equal(summary.container_count, records.length);
      assert.equal(summary.network_count, 5);
      const sanitizedRecords = JSON.parse(fs.readFileSync(path.join(reportDirectory, 'compose-ps.json'), 'utf8'));
      assert.ok(sanitizedRecords.every(record => record.Project === configuration.name && record.Service));
      const reportText = fs.readdirSync(reportDirectory).map(filename => fs.readFileSync(path.join(reportDirectory, filename), 'utf8')).join('\n');
      assert.doesNotMatch(reportText, /must-not-be-exported|CHANGE_ME_DATABASE_PASSWORD|CHANGE_ME_REDIS_PASSWORD/);
      if (process.env.LOCAL_BACKUP_VALIDATION_RUNTIME_EVIDENCE_DIRECTORY && scenario.name === 'readonly_backup_true_sentinel_true') {
        const evidenceDirectory = path.resolve(process.env.LOCAL_BACKUP_VALIDATION_RUNTIME_EVIDENCE_DIRECTORY);
        fs.mkdirSync(evidenceDirectory, { recursive: true });
        for (const filename of fs.readdirSync(reportDirectory)) fs.copyFileSync(path.join(reportDirectory, filename), path.join(evidenceDirectory, filename));
        fs.writeFileSync(path.join(evidenceDirectory, 'simulated-docker-queries.log'), queries.join('\n') + '\n');
      }
    } else if (!scenario.succeeds) {
      assert.equal(fs.existsSync(path.join(reportDirectory, 'runtime-metadata-boundary-summary.json')), false);
      if (scenario.name === 'existing_report_directory') assert.equal(fs.readFileSync(path.join(reportDirectory, 'compose-config.json'), 'utf8'), 'preserve-previous-evidence');
    }
  });
});
