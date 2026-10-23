'use strict';

const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { join } = require('node:path');
const { test } = require('node:test');
const vm = require('node:vm');

class SyntheticProductHandoffConsoleElement {
  constructor(tag = 'div') { this.tag = tag; this.children = []; this.dataset = {}; this.listeners = {}; this.value = ''; this.textContent = ''; this.hidden = false; }
  append(...children) { this.children.push(...children); }
  replaceChildren(...children) { this.children = children; }
  addEventListener(type, listener) { this.listeners[type] = listener; }
  scrollIntoView() {}
}

function createProductHandoffConsoleHarnessWithVerifiedCandidate() {
  const elements = new Map();
  const findElement = id => {
    if (!elements.has(id)) elements.set(id, new SyntheticProductHandoffConsoleElement());
    return elements.get(id);
  };
  const requests = [];
  const browserWindow = { addEventListener() {} }; browserWindow.parent = browserWindow;
  const context = vm.createContext({
    document: { getElementById: findElement, createElement: tag => new SyntheticProductHandoffConsoleElement(tag) },
    location: { pathname: '/local-backup-maintenance/', origin: 'https://synthetic.invalid' },
    window: browserWindow, Headers, console, clearInterval() {}, setInterval() { return 1; },
    fetch: async (...args) => { requests.push(args); throw new Error('unexpected synthetic network request'); },
  });
  vm.runInContext(readFileSync(join(__dirname, '../internal/localbackup/maintenance_console/maintenance-console.js'), 'utf8'), context);
  const task = { id: 'synthetic-candidate-task', kind: 'upload', state: 'product_module_isolated_verified', expires_at: new Date(Date.now() + 3600000).toISOString(), size: 2048, manifest: { scope: 'product_catalog_with_images', package_id: 'synthetic-package' } };
  const status = { permissions: ['ops:backup:restore'], application_fenced: false, site_id: 'synthetic-store', environment: 'disposable', storage_roots: {}, max_package_bytes: 4096 };
  const report = { version: 1, task_id: task.id, package_id: task.manifest.package_id, scope: task.manifest.scope, expires_at: task.expires_at, captured_at: new Date().toISOString(), online_overwrite_available: false, custody_transfer_available: false, contract_sha256: 'a'.repeat(64), preserved_tables: 201, custody_rules: ['隔离候选仍由维护器保管'], projection_requirements: [{ description: '保留当前浏览增量与客服 Stream' }], blockers: ['production_product_candidate_ownership_handoff_not_implemented', 'product_projection_configuration_unknown_reregister_current_business_build', 'buffered_product_view_delta_drain_and_inflight_idempotency_contract_not_implemented'] };
  Object.assign(context, { syntheticTask: task, syntheticStatus: status, syntheticReport: report });
  vm.runInContext('maintenanceCredential="synthetic-token";targetStatus=syntheticStatus;displayedMaintenanceTasks=[syntheticTask]', context);
  const collectText = element => [element.textContent, ...element.children.map(child => typeof child === 'string' ? child : collectText(child))].join('\n');
  return { context, findElement, task, status, report, requests, run: code => vm.runInContext(code, context), text: id => collectText(findElement(id)) };
}

for (const scenario of ['new','ready','restarted','expired_retained','draining','cancelling','view_only']) {
  test(`continuous stopped handoff actions preserve isolation: ${scenario}`,()=>{
    const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
    harness.status.product_view_stopped_drain_configured=true;
    if(scenario!=='new'){
      harness.status.application_fenced=true;
      harness.task.state=scenario==='ready'?'product_module_stopped_handoff_prepared':'interrupted';
      harness.task.product_catalog_stopped_handoff_preparation={phase:scenario==='draining'?'preparing':scenario==='cancelling'?'cancelling':'ready'};
      harness.task.product_view_stopped_drain={phase:scenario==='draining'?'draining':'candidate_rebuilt'};
      if(scenario==='expired_retained')harness.task.expires_at=new Date(0).toISOString();
      if(scenario==='view_only')harness.status.permissions=['ops:backup:view'];
    }
    harness.run('renderBackupTasks([syntheticTask])');
    const text=harness.text('tasks');
    assert.equal(text.includes('保持停写准备商品交接'),scenario==='new');
    assert.equal(text.includes('重新核验并保持停写'),!['new','view_only','cancelling'].includes(scenario));
    assert.equal(text.includes('取消交接准备并恢复当前商城'),!['new','view_only','draining'].includes(scenario));
    if(scenario!=='new'){
      assert(!text.includes('解除已重建预览'));
      assert(!text.includes('解除未改数据'));
      assert(!text.includes('重试浏览增量排空并重建预览'));
      assert(!text.includes('验证并演练隔离恢复'));
      assert(!text.includes('检查商品恢复准备'));
    }
    assert.equal(harness.requests.length,0);
  });
}

test('continuous handoff preparation confirms continued downtime and complete writer inventory',async()=>{
  const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
  harness.status.product_view_stopped_drain_configured=true;
  harness.status.product_view_recovery_writer_inventory=['API and background workers','all importer commands'];
  let posted;
  harness.context.prompt=message=>{assert.match(message,/完成后仍持续停写/);assert.match(message,/all importer commands/);return `PREPARE STOPPED PRODUCT HANDOFF synthetic-store/disposable/${harness.task.id}`;};
  harness.context.postMaintenanceRequest=async(path,body)=>{posted={path,body};return {};};
  await harness.run('startContinuouslyStoppedProductCatalogHandoffPreparation(syntheticTask.id)');
  assert.equal(posted.path,`/api/tasks/${harness.task.id}/product-catalog-stopped-handoff-preparation`);
  assert.equal(posted.body.all_listed_writers_stopped_and_use_this_single_host_gate,true);
  assert.equal(harness.findElement('preview-panel').hidden,true);
  assert.match(harness.text('message'),/完成后商城仍保持停写/);
});

for(const scenario of ['submit','cancel_prompt','disconnect','no_permission']){
  test(`verified handoff cancellation controls: ${scenario}`,async()=>{
    const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
    harness.status.application_fenced=true;
    harness.task.state='product_module_stopped_handoff_prepared';
    harness.task.product_catalog_stopped_handoff_preparation={phase:'ready'};
    harness.task.expires_at=new Date(0).toISOString();
    if(scenario==='no_permission')harness.status.permissions=['ops:backup:view'];
    let posted;
    harness.context.prompt=()=>scenario==='cancel_prompt'?null:`CANCEL STOPPED PRODUCT HANDOFF synthetic-store/disposable/${harness.task.id}`;
    harness.context.postMaintenanceRequest=async(path,body)=>{posted={path,body};if(scenario==='disconnect')harness.findElement('disconnect').listeners.click();return {};};
    if(scenario==='no_permission')await assert.rejects(harness.run('cancelStoppedProductCatalogHandoffPreparationAfterVerification(syntheticTask.id)'));
    else await harness.run('cancelStoppedProductCatalogHandoffPreparationAfterVerification(syntheticTask.id)');
    if(['no_permission','cancel_prompt'].includes(scenario))assert.equal(posted,undefined);
    else assert.equal(posted.path,`/api/tasks/${harness.task.id}/product-catalog-stopped-handoff-cancellation`);
    if(scenario==='disconnect')assert(!harness.text('message').includes('正在核验后取消准备'));
  });
}

for (const scenario of ['configured','unconfigured','view_only','interrupted','interrupted_expired','expired']) {
  test(`stopped view drain button and unresolved release guards: ${scenario}`,()=>{
    const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
    harness.status.product_view_stopped_drain_configured=scenario!=='unconfigured';
    if(scenario==='view_only')harness.status.permissions=['ops:backup:view'];
    if(scenario==='expired')harness.task.work_copy_expired=true;
    if(scenario.startsWith('interrupted')){
      harness.status.application_fenced=true;
      harness.task.state='interrupted';
      harness.task.product_view_stopped_drain={phase:'draining'};
      if(scenario==='interrupted_expired')harness.task.expires_at=new Date(0).toISOString();
    }
    harness.run('renderBackupTasks([syntheticTask])');
    assert.equal(harness.text('tasks').includes('停写保存浏览增量并重建预览'),scenario==='configured');
    if(scenario.startsWith('interrupted')){
      assert(harness.text('tasks').includes('重试浏览增量排空并重建预览'));
      assert(!harness.text('tasks').includes('解除未改数据'));
      assert(!harness.text('tasks').includes('验证并演练隔离恢复'));
      assert(!harness.text('tasks').includes('查看范围与覆盖预览'));
    }
    assert.equal(harness.requests.length,0);
  });
}

test('stopped view drain requires typed inventory confirmation and clears the old preview',async()=>{
  const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
  harness.status.product_view_stopped_drain_configured=true;
  harness.status.product_view_recovery_writer_inventory=['API and workers','importer'];
  let posted;
  harness.context.prompt=message=>{assert.match(message,/API and workers/);assert.match(message,/多主机写入不支持/);return `DRAIN PRODUCT VIEWS synthetic-store/disposable/${harness.task.id}`;};
  harness.context.postMaintenanceRequest=async(path,body)=>{posted={path,body};return {};};
  await harness.run('startStoppedProductViewDrainAndFreshCandidateRebuild(syntheticTask.id)');
  assert.equal(posted.path,`/api/tasks/${harness.task.id}/product-view-stopped-drain-and-rebuild`);
  assert.equal(posted.body.all_listed_writers_stopped_and_use_this_single_host_gate,true);
  assert.equal(harness.findElement('preview-panel').hidden,true);
  assert.equal(harness.findElement('product-catalog-handoff-inspection-panel').hidden,true);
  assert(!JSON.stringify(posted).includes('redis_password'));
});

test('cancelled stopped view drain does not start maintenance',async()=>{
  const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
  harness.status.product_view_stopped_drain_configured=true;
  harness.context.prompt=()=>null;
  await harness.run('startStoppedProductViewDrainAndFreshCandidateRebuild(syntheticTask.id)');
  assert.equal(harness.requests.length,0);
});

for (const scenario of ['eligible', 'view_only', 'fenced', 'expired', 'invalid_expiry', 'journal', 'wrong_scope', 'unverified', 'removed']) {
  test(`product inspection button follows permission and current candidate state: ${scenario}`, () => {
    const harness = createProductHandoffConsoleHarnessWithVerifiedCandidate();
    if (scenario === 'view_only') harness.status.permissions = ['ops:backup:view'];
    if (scenario === 'fenced') harness.status.application_fenced = true;
    if (scenario === 'expired') harness.task.expires_at = new Date(0).toISOString();
    if (scenario === 'invalid_expiry') harness.task.expires_at = 'not-a-date';
    if (scenario === 'journal') harness.task.recovery_journal = {};
    if (scenario === 'wrong_scope') harness.task.manifest.scope = 'whole_application';
    if (scenario === 'unverified') harness.task.state = 'uploaded';
    if (scenario === 'removed') harness.task.work_copy_expired = true;
    harness.run('renderBackupTasks([syntheticTask])');
    assert.equal(harness.text('tasks').includes('检查商品恢复准备'), scenario === 'eligible');
    assert(!harness.text('tasks').includes('覆盖恢复'));
    assert.equal(harness.requests.length, 0);
  });
}

test('product report shows custody and unknown/buffered blockers without a release or acceptance action', async () => {
  const harness = createProductHandoffConsoleHarnessWithVerifiedCandidate();
  harness.context.postMaintenanceRequest = async (path, body) => {
    assert.equal(path, `/api/tasks/${harness.task.id}/product-catalog-handoff-inspection`);
    assert.equal(body, undefined);
    return { report: harness.report };
  };
  await harness.run('inspectProductCatalogCandidateHandoff(syntheticTask.id)');
  assert.match(harness.text('product-catalog-handoff-inspection-status'), /在线覆盖尚未开放/);
  const text = harness.text('product-catalog-handoff-inspection-contents');
  assert.match(text, /隔离候选仍由维护器保管/);
  assert.match(text, /尚未完整登记商品缓存和浏览数配置/);
  assert.match(text, /不能证明不会丢失或重复累计/);
  assert.equal(harness.findElement('product-catalog-handoff-inspection-contents').children.filter(element => element.tag === 'button').length, 0);
});

for (const scenario of ['expired', 'invalid_expiry', 'wrong_package', 'wrong_task', 'unsafe_overwrite_flag']) {
  test(`unusable server inspection cannot appear as a completed candidate check: ${scenario}`, async () => {
    const harness = createProductHandoffConsoleHarnessWithVerifiedCandidate();
    if (scenario === 'expired') harness.report.expires_at = new Date(0).toISOString();
    if (scenario === 'invalid_expiry') harness.report.expires_at = 'not-a-date';
    if (scenario === 'wrong_package') harness.report.package_id = 'different-package';
    if (scenario === 'wrong_task') harness.report.task_id = 'different-task';
    if (scenario === 'unsafe_overwrite_flag') harness.report.online_overwrite_available = true;
    harness.context.postMaintenanceRequest = async () => ({ report: harness.report });
    await harness.run('inspectProductCatalogCandidateHandoff(syntheticTask.id)');
    assert(!harness.text('product-catalog-handoff-inspection-status').includes('已核对当前商品候选'));
    assert.equal(harness.findElement('product-catalog-handoff-inspection-contents').children.length, 0);
  });
}

for (const scenario of ['disconnect', 'reconnect_same_credential', 'candidate_changes', 'permission_removed', 'refresh_failed']) {
  test(`a pending inspection is discarded after ${scenario}`, async () => {
    const harness = createProductHandoffConsoleHarnessWithVerifiedCandidate();
    let finishInspection;
    harness.context.postMaintenanceRequest = () => new Promise(resolve => { finishInspection = resolve; });
    const pending = harness.run('inspectProductCatalogCandidateHandoff(syntheticTask.id)');
    if (scenario === 'disconnect') harness.findElement('disconnect').listeners.click();
    if (scenario === 'reconnect_same_credential') {
      harness.context.maintenanceRequest = async path => path === '/api/status' ? harness.status : { tasks: [harness.task] };
      await harness.run('connectMaintenanceConsole("synthetic-token")');
    }
    if (scenario === 'candidate_changes' || scenario === 'permission_removed') {
      const refreshedTask = structuredClone(harness.task);
      const refreshedStatus = structuredClone(harness.status);
      if (scenario === 'candidate_changes') refreshedTask.ciphertext_sha256 = 'changed';
      if (scenario === 'permission_removed') refreshedStatus.permissions = ['ops:backup:view'];
      harness.context.maintenanceRequest = async path => path === '/api/status' ? refreshedStatus : { tasks: [refreshedTask] };
      await harness.run('refreshMaintenanceTasks()');
    }
    if (scenario === 'refresh_failed') {
      harness.context.maintenanceRequest = async () => { throw new Error('synthetic refresh failed'); };
      await assert.rejects(harness.run('refreshMaintenanceTasks()'), /synthetic refresh failed/);
    }
    finishInspection({ report: harness.report });
    await pending;
    assert.equal(harness.findElement('product-catalog-handoff-inspection-panel').hidden, true);
    assert.equal(harness.findElement('product-catalog-handoff-inspection-contents').children.length, 0);
  });
}

test('stable task refresh preserves an observation but does not request a switch', async () => {
  const harness = createProductHandoffConsoleHarnessWithVerifiedCandidate();
  harness.context.postMaintenanceRequest = async () => ({ report: harness.report });
  await harness.run('inspectProductCatalogCandidateHandoff(syntheticTask.id)');
  harness.context.maintenanceRequest = async path => path === '/api/status' ? structuredClone(harness.status) : { tasks: [structuredClone(harness.task)] };
  await harness.run('refreshMaintenanceTasks()');
  assert.equal(harness.findElement('product-catalog-handoff-inspection-panel').hidden, false);
  assert.match(harness.text('product-catalog-handoff-inspection-status'), /已核对当前商品候选/);
  assert.equal(harness.requests.length, 0);
});

test('registered SQL batch receipts still show stopped handoff as required', async () => {
  const harness = createProductHandoffConsoleHarnessWithVerifiedCandidate();
  harness.report.blockers = ['buffered_product_view_delta_stopped_handoff_and_target_verification_not_implemented'];
  harness.context.postMaintenanceRequest = async () => ({ report: harness.report });
  await harness.run('inspectProductCatalogCandidateHandoff(syntheticTask.id)');
  assert.match(harness.text('product-catalog-handoff-inspection-contents'), /已有批次幂等保护/);
  assert.match(harness.text('product-catalog-handoff-inspection-contents'), /停止全部写入/);
  assert.match(harness.text('product-catalog-handoff-inspection-contents'), /当前不能接收/);
});

function configureSyntheticStoppedProductSwitchConsoleScenario(harness,switched=true) {
  harness.status.application_fenced=true;
  harness.status.product_view_stopped_drain_configured=true;
  harness.task.state=switched?'product_module_switched_still_isolated':'product_module_stopped_handoff_prepared';
  harness.task.product_catalog_stopped_handoff_preparation={phase:'ready',window_id:'synthetic-stopped-window'};
  if(switched)harness.task.recovery_journal={new_database_oid:42,product_catalog_stopped_switch:{phase:'switched_still_isolated',window_id:'synthetic-stopped-window',sha256:'b'.repeat(64)}};
  harness.context.syntheticStoppedSwitchReport={task_id:harness.task.id,scope:'product_catalog_with_images',package_id:harness.task.manifest.package_id,window_id:'synthetic-stopped-window',switch_journal_sha256:'b'.repeat(64),database_oid:42,application_remains_stopped:true,candidate_remains_isolated:true,online_overwrite_available:false,external_actions_executed:false,contract_sha256:'c'.repeat(64),projection_requirements:[{description:'保留订单、客服和浏览收据；尚需核对缓存与商城页面。'}]};
}

for(const scenario of ['prepared','prepared_expired','switched','interrupted','switched_expired','view_only','rolled_back']) {
  test(`stopped product switch and dedicated paired rollback buttons: ${scenario}`,()=>{
    const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
    configureSyntheticStoppedProductSwitchConsoleScenario(harness,!scenario.startsWith('prepared'));
    if(scenario.includes('expired'))harness.task.expires_at=new Date(0).toISOString();
    if(scenario==='interrupted')harness.task.state='interrupted';
    if(scenario==='view_only')harness.status.permissions=['ops:backup:view'];
    if(scenario==='rolled_back'){harness.task.state='rolled_back';harness.task.recovery_journal.product_catalog_stopped_switch.phase='original_released';harness.status.application_fenced=false;}
    harness.run('renderBackupTasks([syntheticTask])');
    const text=harness.text('tasks');
    assert.equal(text.includes('继续隔离切换商品（仍停写）'),scenario==='prepared');
    assert.equal(text.includes('核对本次商品切换结果'),['switched','switched_expired'].includes(scenario));
    assert.equal(text.includes('成对回退商品并恢复原商城'),['switched','switched_expired','interrupted'].includes(scenario));
    if(!scenario.startsWith('prepared')){assert(!text.includes('取消交接准备并恢复当前商城'));assert(!text.includes('重新核验并保持停写'));}
  });
}

for(const rollback of [false,true]) {
  test(`stopped product operation binds target confirmation and continued downtime: rollback=${rollback}`,async()=>{
    const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
    configureSyntheticStoppedProductSwitchConsoleScenario(harness,rollback);
    let posted;
    harness.context.prompt=message=>{assert.match(message,rollback?/核对通过后恢复原商城/:/仍持续停写/);return `${rollback?'ROLLBACK':'SWITCH'} STOPPED PRODUCT CATALOG synthetic-store/disposable/${harness.task.id}`;};
    harness.context.postMaintenanceRequest=async(path,body)=>{posted={path,body};return {};};
    await harness.run(`startProductCatalogIsolatedSwitchOrVerifiedPairedRollback(syntheticTask.id,${rollback})`);
    assert.match(posted.path,rollback?/product-catalog-stopped-switch-rollback$/:/product-catalog-stopped-isolated-switch$/);
    assert.match(harness.text('message'),rollback?/正在成对回退商品/:/商城仍保持停写/);
  });
}

for(const scenario of ['valid','wrong_window','wrong_journal','wrong_database','claims_release','claims_external_actions','disconnect','changed_candidate','permission_removed']) {
  test(`post-switch product report binds current window and rejects stale approval claims: ${scenario}`,async()=>{
    const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
    configureSyntheticStoppedProductSwitchConsoleScenario(harness);
    const report=harness.context.syntheticStoppedSwitchReport;
    if(scenario==='wrong_window')report.window_id='historical-window';
    if(scenario==='wrong_journal')report.switch_journal_sha256='old-journal';
    if(scenario==='wrong_database')report.database_oid=1;
    if(scenario==='claims_release')report.online_overwrite_available=true;
    if(scenario==='claims_external_actions')report.external_actions_executed=true;
    let finish;
    harness.context.postMaintenanceRequest=()=>new Promise(resolve=>{finish=resolve;});
    const pending=harness.run('inspectStoppedProductCatalogPostSwitch(syntheticTask.id)');
    if(scenario==='disconnect')harness.findElement('disconnect').listeners.click();
    if(['changed_candidate','permission_removed'].includes(scenario)){
      const task=structuredClone(harness.task),status=structuredClone(harness.status);
      if(scenario==='changed_candidate')task.recovery_journal.product_catalog_stopped_switch.sha256='new-journal';
      else status.permissions=['ops:backup:view'];
      harness.context.maintenanceRequest=async path=>path==='/api/status'?status:{tasks:[task]};
      await harness.run('refreshMaintenanceTasks()');
    }
    finish({report});await pending;
    if(scenario==='valid'){
      assert.match(harness.text('product-catalog-stopped-post-switch-inspection-status'),/不能恢复营业/);
      assert.match(harness.text('product-catalog-stopped-post-switch-inspection-contents'),/保留订单、客服/);
    }else if(['disconnect','changed_candidate','permission_removed'].includes(scenario)){
      assert.equal(harness.findElement('product-catalog-stopped-post-switch-inspection-panel').hidden,true);
      assert.equal(harness.findElement('product-catalog-stopped-post-switch-inspection-contents').children.length,0);
    }else{
      assert.match(harness.text('product-catalog-stopped-post-switch-inspection-status'),/不一致/);
      assert.equal(harness.findElement('product-catalog-stopped-post-switch-inspection-contents').children.length,0);
    }
  });
}

for(const scenario of ['switched','failed_before_intent','deleting','verified_absent','failed_after_verified','expired','interrupted','view_only','rolling_back','rolled_back']) {
  test(`stopped product detail cache action and durable progress visibility: ${scenario}`,()=>{
    const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
    configureSyntheticStoppedProductSwitchConsoleScenario(harness);
    if(scenario==='failed_before_intent')harness.task.state='failed';
    if(['deleting','verified_absent'].includes(scenario))harness.task.product_catalog_stopped_detail_cache_invalidation={phase:scenario,planned_detail_cache_keys:134,planned_read_lease_keys:2};
    if(scenario==='deleting')harness.task.state='failed';
    if(scenario==='failed_after_verified'){harness.task.state='failed';harness.task.product_catalog_stopped_detail_cache_invalidation={phase:'verified_absent',planned_detail_cache_keys:134,planned_read_lease_keys:2};}
    if(scenario==='expired')harness.task.expires_at=new Date(0).toISOString();
    if(scenario==='interrupted')harness.task.state='interrupted';
    if(scenario==='view_only')harness.status.permissions=['ops:backup:view'];
    if(scenario==='rolling_back')harness.task.recovery_journal.product_catalog_stopped_switch.phase='rolling_back';
    if(scenario==='rolled_back'){harness.task.state='rolled_back';harness.task.recovery_journal.product_catalog_stopped_switch.phase='original_released';harness.status.application_fenced=false;harness.task.product_catalog_stopped_detail_cache_invalidation={phase:'verified_absent'};}
    harness.run('renderBackupTasks([syntheticTask])');
    const text=harness.text('tasks');
    assert.equal(text.includes('处理商品详情缓存（仍停写）'),['switched','failed_before_intent'].includes(scenario));
    assert.equal(text.includes('重试商品详情缓存处理（仍停写）'),scenario==='deleting');
    assert.equal(text.includes('复核商品详情缓存（仍停写）'),['verified_absent','failed_after_verified'].includes(scenario));
    if(scenario==='verified_absent')assert.match(text,/详情键 134、读取锁 2/);
    if(scenario==='failed_after_verified'){assert.match(text,/旧商品详情缓存处理记录/);assert(!text.includes('已核对为空'));}
    if(scenario==='rolled_back'){assert.match(text,/原商城已成对回退/);assert(!text.includes('商城继续停写'));}
    assert(!text.includes('恢复新商品营业'));
  });
}

test('stopped detail cache action sends only a target confirmation and keeps service stopped',async()=>{
  const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
  configureSyntheticStoppedProductSwitchConsoleScenario(harness);
  let posted;
  harness.context.prompt=message=>{assert.match(message,/商城继续停写/);assert.match(message,/浏览增量、客服消息和任务数据会保留/);return `INVALIDATE STOPPED PRODUCT CACHE synthetic-store/disposable/${harness.task.id}`;};
  harness.context.postMaintenanceRequest=async(path,body)=>{posted={path,body};return {};};
  await harness.run('startStoppedProductDetailCacheInvalidation(syntheticTask.id)');
  assert.equal(posted.path,`/api/tasks/${harness.task.id}/product-catalog-stopped-detail-cache-invalidation`);
  assert.deepEqual(Object.keys(posted.body),['confirmation']);
  assert.match(harness.text('message'),/完成后商城继续停写/);
});

for(const scenario of ['changed_task','permission_removed','cancelled','disconnect_during_request']) {
  test(`stopped detail cache action discards changed state and stale response: ${scenario}`,async()=>{
    const harness=createProductHandoffConsoleHarnessWithVerifiedCandidate();
    configureSyntheticStoppedProductSwitchConsoleScenario(harness);
    let posted=0;
    harness.context.prompt=()=>{
      if(scenario==='changed_task')harness.task.recovery_journal.product_catalog_stopped_switch.sha256='changed';
      if(scenario==='permission_removed')harness.status.permissions=['ops:backup:view'];
      if(scenario==='cancelled')return null;
      return 'synthetic confirmation';
    };
    let finish;
    harness.context.postMaintenanceRequest=()=>{posted++;return new Promise(resolve=>{finish=resolve;});};
    const pending=harness.run('startStoppedProductDetailCacheInvalidation(syntheticTask.id)');
    if(scenario==='disconnect_during_request'){
      harness.findElement('disconnect').listeners.click();
      finish({});await pending;
      assert(!harness.text('message').includes('正在核验并处理'));
    }else if(scenario==='cancelled'){
      await pending;assert.equal(posted,0);
    }else{
      await assert.rejects(pending,/商品状态已变化/);assert.equal(posted,0);
    }
  });
}
