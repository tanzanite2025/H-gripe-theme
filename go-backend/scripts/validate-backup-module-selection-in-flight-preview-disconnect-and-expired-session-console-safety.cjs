'use strict';
const assert = require('node:assert/strict');
const vm = require('node:vm');
const { test } = require('node:test');
const { webcrypto } = require('node:crypto');
const { loadMaintenanceConsoleScriptsInDocumentOrder } = require('./load-maintenance-console-classic-scripts-in-html-document-order.cjs');

class BackupInteractionSafetySyntheticElement {
  constructor() { this.value = ''; this.hidden = false; this.disabled = false; this.checked = false; this.children = []; this.dataset = {}; this.listeners = {}; this.files = []; }
  addEventListener(type, callback) { this.listeners[type] = callback; }
  append(...elements) { this.children.push(...elements); }
  replaceChildren(...elements) { this.children = elements; }
  scrollIntoView() {}
  focus() { this.focused = true; }
}

function createDeferredMaintenanceHTTPResponse() {
  let resolve;
  const promise = new Promise(complete => { resolve = complete; });
  return { promise, resolve };
}

function createBackupModuleSelectionAndMaintenanceSessionSafetyHarness(fetchResponse) {
  const elements = new Map();
  const parentMessages = [];
  const requests = [];
  const storedUploadTasks = new Map();
  const parent = {postMessage: message => parentMessages.push(message)};
  const radios = ['whole', 'modules'].map(value => Object.assign(new BackupInteractionSafetySyntheticElement(), {value, checked: value === 'modules'}));
  const checkboxes = ['orders', 'customers', 'customer_service'].map(value => Object.assign(new BackupInteractionSafetySyntheticElement(), {value, checked: value === 'orders'}));
  const findElement = id => { if (!elements.has(id)) elements.set(id, new BackupInteractionSafetySyntheticElement()); return elements.get(id); };
  const context = vm.createContext({
    window: {parent, addEventListener() {}},
    location: {origin: 'http://maintenance.invalid', pathname: '/local-backup-maintenance/'},
    document: {
      getElementById: findElement,
      createElement: () => new BackupInteractionSafetySyntheticElement(),
      querySelectorAll: selector => selector.includes('backup-export-mode') ? radios : selector.includes('input:checked') ? checkboxes.filter(input => input.checked) : selector.includes('input') ? checkboxes : [],
      querySelector: selector => selector.includes('backup-export-mode') ? radios.find(input => input.checked) : null,
    },
    Headers, Blob, Uint8Array, crypto: webcrypto, console,
    prompt: () => { throw new Error('This verification must never request backup generation'); },
    clearInterval() {}, setInterval() { return 1; },
    localStorage: {getItem: key => storedUploadTasks.get(key), setItem: (key, value) => storedUploadTasks.set(key, value), removeItem: key => storedUploadTasks.delete(key)},
    fetch: async (url, options) => { requests.push({url, options}); return fetchResponse(url, options); },
  });
  loadMaintenanceConsoleScriptsInDocumentOrder(context);
  vm.runInContext(`maintenanceCredential='synthetic-in-memory-session';targetStatus={permissions:['ops:backup:view','ops:backup:export','ops:backup:upload'],business_module_backup_options:[{id:'orders'},{id:'customers'},{id:'customer_service'}],application_fenced:false,site_id:'synthetic',environment:'test',max_package_bytes:1048576,chunk_bytes:1024};`, context);
  return {context, elements, findElement, parentMessages, requests, storedUploadTasks, checkboxes,
    receive: data => context.receiveAdminBackupScopeSelectionOrWorkflowNavigation({origin: context.location.origin, source: parent, data}),
    readState: () => JSON.parse(vm.runInContext('JSON.stringify({revision:maintenanceConnectionRevision,credential:maintenanceCredential,scope:selectedBusinessBackupModuleIDs(),preview:selectedBusinessBackupPreview,busy:businessBackupRequestInProgress,uploading,connected:!!targetStatus})', context)),
  };
}

function createSyntheticMaintenanceJSONResponse(status, body) {
  return {ok: status >= 200 && status < 300, status, json: async () => body};
}

test('changing modules during a delayed preview discards the old result and applies the latest requested scope', async () => {
  const response = createDeferredMaintenanceHTTPResponse();
  const h = createBackupModuleSelectionAndMaintenanceSessionSafetyHarness(() => response.promise);
  const preview = h.context.previewCurrentSelectedBusinessBackupModules();
  h.receive({type: 'local-backup-maintenance-select-scope', mode: 'modules', modules: ['customers']});
  h.receive({type: 'local-backup-maintenance-select-scope', mode: 'modules', modules: ['customer_service']});
  response.resolve(createSyntheticMaintenanceJSONResponse(200, {preview: {modules: ['orders'], online_overwrite_available: false, included_dependencies: [], recovery_limitations: [], tables: {}}}));
  await preview;
  assert.deepEqual(h.readState().scope, ['customer_service']);
  assert.equal(h.readState().preview, null);
  assert.equal(h.findElement('generate').disabled, true);
  assert.equal(h.findElement('business-backup-range-preview').hidden, true);
  assert.equal(h.requests.length, 1);
});

test('export workflow queued behind a preview uses the latest module selection and starts no backup', async () => {
  const response = createDeferredMaintenanceHTTPResponse();
  const h = createBackupModuleSelectionAndMaintenanceSessionSafetyHarness((url, options) => {
    const modules = JSON.parse(options.body).modules;
    if (h.requests.length === 1) return response.promise;
    assert.deepEqual(modules, ['customers']);
    return createSyntheticMaintenanceJSONResponse(200, {preview: {modules, online_overwrite_available: false, included_dependencies: [], recovery_limitations: [], tables: {}}});
  });
  const firstPreview = h.context.previewCurrentSelectedBusinessBackupModules();
  h.receive({type: 'local-backup-maintenance-select-scope', mode: 'modules', modules: ['customers']});
  h.receive({type: 'local-backup-maintenance-show-workflow', workflow: 'export'});
  response.resolve(createSyntheticMaintenanceJSONResponse(200, {preview: {modules: ['orders'], online_overwrite_available: false, included_dependencies: [], recovery_limitations: [], tables: {}}}));
  await firstPreview;
  await new Promise(resolve => setImmediate(resolve));
  assert.equal(h.requests.length, 2);
  assert(h.requests.every(request => request.url.endsWith('/api/module-backup-preview')));
  assert.deepEqual(h.readState().preview.modules, ['customers']);
});

test('busy and permission state reaches the embedded parent and disables export during upload', () => {
  const h = createBackupModuleSelectionAndMaintenanceSessionSafetyHarness(() => { throw new Error('unexpected request'); });
  vm.runInContext('uploading=true;updateBusinessBackupSelectionControls();', h.context);
  assert.equal(h.findElement('generate').disabled, true);
  assert.equal(h.findElement('upload-button').disabled, true);
  assert.deepEqual(JSON.parse(JSON.stringify(h.parentMessages.at(-1))), {type: 'local-backup-maintenance-backup-controls-state', busy: true, exportAllowed: false, uploadAllowed: false});
});

test('expired maintenance credentials clear pending scope and disable connected actions', async () => {
  const h = createBackupModuleSelectionAndMaintenanceSessionSafetyHarness(() => createSyntheticMaintenanceJSONResponse(401, {error: 'expired'}));
  vm.runInContext('businessBackupRequestInProgress=true;', h.context);
  h.receive({type: 'local-backup-maintenance-select-scope', mode: 'modules', modules: ['customers']});
  await assert.rejects(h.context.maintenanceRequest('/api/status'), /expired/);
  const state = h.readState();
  assert.equal(state.credential, '');
  assert.equal(state.connected, false);
  assert.equal(state.preview, null);
  assert.equal(h.findElement('connected-panel').hidden, true);
  assert.equal(h.findElement('generate').disabled, true);
  assert.equal(h.findElement('upload-button').disabled, true);
  assert.equal(vm.runInContext('pendingAdminBusinessBackupScopeSelection', h.context), null);
  assert.equal(h.parentMessages.at(-1).type, 'local-backup-maintenance-disconnected');
});

for (const oldStatus of [200, 401]) {
  test('late HTTP ' + oldStatus + ' from an old session cannot invalidate the new maintenance session', async () => {
    const response = createDeferredMaintenanceHTTPResponse();
    const h = createBackupModuleSelectionAndMaintenanceSessionSafetyHarness(() => response.promise);
    const request = h.context.maintenanceRequest('/api/status');
    vm.runInContext("maintenanceConnectionRevision++;maintenanceCredential='new-synthetic-session';", h.context);
    response.resolve(createSyntheticMaintenanceJSONResponse(oldStatus, {}));
    await assert.rejects(request, /维护会话已更换/);
    assert.equal(h.readState().credential, 'new-synthetic-session');
    assert.equal(h.readState().connected, true);
    assert(!h.parentMessages.some(message => message.type === 'local-backup-maintenance-disconnected'));
  });
}

test('a preview finishing after disconnect cannot reenable controls or reapply queued modules', async () => {
  const response = createDeferredMaintenanceHTTPResponse();
  const h = createBackupModuleSelectionAndMaintenanceSessionSafetyHarness(() => response.promise);
  const preview = h.context.previewCurrentSelectedBusinessBackupModules();
  h.receive({type: 'local-backup-maintenance-select-scope', mode: 'modules', modules: ['customers']});
  h.context.disconnectMaintenanceConsoleAndInvalidatePendingClientOperations('synthetic disconnect');
  response.resolve(createSyntheticMaintenanceJSONResponse(200, {preview: {modules: ['orders']}}));
  await assert.rejects(preview, /维护会话已更换/);
  assert.equal(h.findElement('generate').disabled, true);
  assert.equal(h.readState().connected, false);
  assert.deepEqual(h.readState().scope, ['orders']);
});

test('resume inspection service errors keep the existing upload instead of creating another task', async () => {
  const h = createBackupModuleSelectionAndMaintenanceSessionSafetyHarness(url => {
    if (url.endsWith('/api/tasks/existing-upload')) return createSyntheticMaintenanceJSONResponse(503, {error: 'temporarily unavailable'});
    throw new Error('Unexpected mutation or refresh: ' + url);
  });
  h.findElement('upload-file').files = [new Blob(['synthetic encrypted package bytes'])];
  h.storedUploadTasks.get = () => 'existing-upload';
  vm.runInContext('refreshMaintenanceTasks=async()=>{};', h.context);
  await assert.rejects(h.context.uploadLocalBackupFile(), /temporarily unavailable/);
  assert.equal(h.requests.length, 1);
  assert.equal(h.requests[0].url, '/local-backup-maintenance/api/tasks/existing-upload');
  assert.equal(h.readState().uploading, false);
});

test('unavailable maintenance status disables stale connected actions and requests reconnection', async () => {
  const h = createBackupModuleSelectionAndMaintenanceSessionSafetyHarness(() => createSyntheticMaintenanceJSONResponse(503, {error: 'maintenance unavailable'}));
  await assert.rejects(h.context.refreshMaintenanceTasks(), /maintenance unavailable/);
  assert.equal(h.readState().connected, false);
  assert.equal(h.findElement('connected-panel').hidden, true);
  assert.equal(h.findElement('generate').disabled, true);
  assert.equal(h.findElement('upload-button').disabled, true);
  assert.equal(h.parentMessages.at(-1).type, 'local-backup-maintenance-disconnected');
});

test('a missing resumable work copy can still start a fresh chunked upload and commit', async () => {
  const file = new Blob(['synthetic encrypted package bytes']);
  const h = createBackupModuleSelectionAndMaintenanceSessionSafetyHarness((url, options) => {
    if (url.endsWith('/api/tasks/existing-upload')) return createSyntheticMaintenanceJSONResponse(404, {});
    if (url.endsWith('/api/uploads')) return createSyntheticMaintenanceJSONResponse(200, {task: {id: 'fresh-upload', uploaded_bytes: 0}});
    if (url.endsWith('/api/uploads/fresh-upload/chunks')) return createSyntheticMaintenanceJSONResponse(200, {task: {uploaded_bytes: file.size}});
    if (url.endsWith('/api/uploads/fresh-upload/commit')) return createSyntheticMaintenanceJSONResponse(200, {});
    throw new Error('Unexpected request: ' + options.method + ' ' + url);
  });
  h.findElement('upload-file').files = [file];
  h.storedUploadTasks.get = () => 'existing-upload';
  vm.runInContext('refreshMaintenanceTasks=async()=>{};', h.context);
  await h.context.uploadLocalBackupFile();
  assert.deepEqual(h.requests.map(request => request.options.method || 'GET'), ['GET', 'POST', 'PUT', 'POST']);
  assert.equal(h.findElement('upload-progress').value, 100);
  assert.equal(h.readState().uploading, false);
});

test('disconnect while hashing a local file cannot upload it using a replacement session', async () => {
  const hashStarted = createDeferredMaintenanceHTTPResponse();
  const hashFinished = createDeferredMaintenanceHTTPResponse();
  const h = createBackupModuleSelectionAndMaintenanceSessionSafetyHarness(() => { throw new Error('No upload may start after disconnect'); });
  h.context.crypto = {subtle: {digest: async () => { hashStarted.resolve(); return hashFinished.promise; }}};
  h.findElement('upload-file').files = [new Blob(['synthetic encrypted package bytes'])];
  const upload = h.context.uploadLocalBackupFile();
  await hashStarted.promise;
  h.context.disconnectMaintenanceConsoleAndInvalidatePendingClientOperations('synthetic disconnect');
  vm.runInContext("maintenanceCredential='replacement-synthetic-session';", h.context);
  hashFinished.resolve(new Uint8Array(32).buffer);
  await assert.rejects(upload, /维护会话已更换/);
  assert.equal(h.requests.length, 0);
  assert.equal(h.readState().credential, 'replacement-synthetic-session');
  assert.equal(h.readState().uploading, false);
});
