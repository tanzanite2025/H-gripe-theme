'use strict';

const assert = require('node:assert/strict');
const { createHash, webcrypto } = require('node:crypto');
const { readFileSync } = require('node:fs');
const { join } = require('node:path');
const { test } = require('node:test');
const vm = require('node:vm');

class SyntheticMaintenanceConsoleElement {
  constructor(tag = 'div') { this.tag = tag; this.children = []; this.dataset = {}; this.listeners = {}; this.checked = false; this.files = []; this.value = ''; }
  append(...children) { this.children.push(...children); }
  replaceChildren(...children) { this.children = children; }
  addEventListener(type, listener) { this.listeners[type] = listener; }
  querySelectorAll(selector) {
    const descendants = this.children.flatMap(child => child instanceof SyntheticMaintenanceConsoleElement ? [child, ...child.querySelectorAll('*')] : []);
    if (selector === '*') return descendants;
    if (selector === 'fieldset') return descendants.filter(child => child.tag === selector);
    const property = { '[data-projection-reference]': 'projectionReference', '[data-projection-record]': 'projectionRecord', '[data-projection-confirmed]': 'projectionConfirmed' }[selector];
    return descendants.filter(child => child.dataset[property]);
  }
  querySelector(selector) { return this.querySelectorAll(selector)[0] || null; }
}

function createIsolatedMaintenanceConsoleHarness() {
  const elements = new Map();
  const findElement = id => {
    if (!elements.has(id)) elements.set(id, new SyntheticMaintenanceConsoleElement());
    return elements.get(id);
  };
  const networkRequests = [];
  const browserWindow = { addEventListener() {} }; browserWindow.parent = browserWindow;
  const context = vm.createContext({
    document: { getElementById: findElement, createElement: tag => new SyntheticMaintenanceConsoleElement(tag), createTextNode: text => text },
    location: { pathname: '/local-backup-maintenance/', origin: 'https://synthetic.example.test' },
    window: browserWindow, crypto: webcrypto, Headers, Uint8Array, TextDecoder, atob, console,
    clearInterval() {},
    fetch: async (...args) => { networkRequests.push(args); throw new Error('synthetic external network deliberately unavailable'); },
    prompt: () => { throw new Error('stale or incomplete evidence must never reach confirmation'); },
  });
  vm.runInContext(readFileSync(join(__dirname, '../internal/localbackup/maintenance_console/maintenance-console.js'), 'utf8'), context);
  const contract = {
    contract_sha256: 'a'.repeat(64), blockers: [],
    requirements: ['customer_service_search','redis_cache','customer_service_realtime','media_delivery_and_browser'].map(projection => ({ projection, action: `synthetic-required-${projection}`, description: '独立人工核验要求' })),
  };
  context.syntheticContract = contract;
  vm.runInContext('renderExternalProjectionReconciliationEvidenceRequirements(syntheticContract)', context);
  const fieldsets = findElement('external-projection-reconciliation-evidence').querySelectorAll('fieldset');
  for (const fieldset of fieldsets) {
    fieldset.querySelector('[data-projection-confirmed]').checked = true;
    fieldset.querySelector('[data-projection-reference]').value = 'synthetic-local-evidence-record';
    fieldset.querySelector('[data-projection-record]').files = [{ size: 18, arrayBuffer: async () => new TextEncoder().encode('independent-report').buffer }];
  }
  return { context, findElement, fieldsets, contract, networkRequests, collect: () => vm.runInContext('collectExternalProjectionReconciliationOperatorAttestations(syntheticContract)', context) };
}

test('four required operator records hash the actual local bytes without uploading files', async () => {
  const harness = createIsolatedMaintenanceConsoleHarness();
  const evidence = await harness.collect();
  assert.equal(evidence.contract_sha256, harness.contract.contract_sha256);
  assert.equal(evidence.attestations.length, 4);
  for (let index = 0; index < evidence.attestations.length; index++) {
    assert.equal(evidence.attestations[index].projection, harness.contract.requirements[index].projection);
    assert.equal(evidence.attestations[index].action, harness.contract.requirements[index].action);
    assert.equal(evidence.attestations[index].evidence_sha256, createHash('sha256').update('independent-report').digest('hex'));
  }
  assert.equal(harness.networkRequests.length, 0);
});

for (const scenario of ['unchecked','empty_reference','missing_record','empty_record','oversized_record']) {
  test(`incomplete local operator record is rejected: ${scenario}`, async () => {
    const harness = createIsolatedMaintenanceConsoleHarness();
    const fieldset = harness.fieldsets[2];
    if (scenario === 'unchecked') fieldset.querySelector('[data-projection-confirmed]').checked = false;
    if (scenario === 'empty_reference') fieldset.querySelector('[data-projection-reference]').value = '   ';
    const input = fieldset.querySelector('[data-projection-record]');
    if (scenario === 'missing_record') input.files = [];
    if (scenario === 'empty_record') input.files[0].size = 0;
    if (scenario === 'oversized_record') input.files[0].size = 16 * 1024 * 1024 + 1;
    await assert.rejects(harness.collect(), /请完成每项核对/);
    assert.equal(harness.networkRequests.length, 0);
  });
}

test('changing recovery task while local evidence hashing is pending prevents a release request', async () => {
  const harness = createIsolatedMaintenanceConsoleHarness();
  vm.runInContext('reconciliationTask={id:"original-task"}; inspectedReconciliationReport={external_projection_reconciliation:syntheticContract};', harness.context);
  harness.fieldsets[0].querySelector('[data-projection-record]').files[0].arrayBuffer = async () => {
    vm.runInContext('reconciliationTask={id:"another-task"}', harness.context);
    return new TextEncoder().encode('independent-report').buffer;
  };
  await harness.findElement('reconciliation-form').listeners.submit({ preventDefault() {}, target: {} });
  assert.match(harness.findElement('message').textContent, /对账快照已变化/);
  assert.equal(harness.networkRequests.length, 0);
});

test('a failed snapshot refresh clears previously accepted report and keeps release disabled', async () => {
  const harness = createIsolatedMaintenanceConsoleHarness();
  vm.runInContext('reconciliationTask={id:"original-task"}; inspectedReconciliationReport={external_projection_reconciliation:syntheticContract};', harness.context);
  await harness.findElement('inspect-reconciliation').listeners.click();
  assert.equal(vm.runInContext('inspectedReconciliationReport', harness.context), null);
  assert.equal(harness.findElement('complete-reconciliation').disabled, true);
  assert.equal(harness.findElement('external-projection-reconciliation-evidence').children.length, 0);
  assert.equal(harness.networkRequests.length, 1);
});

test('unknown projection configuration cannot collect completion evidence', async () => {
  const harness = createIsolatedMaintenanceConsoleHarness();
  harness.contract.blockers.push('configuration_unknown');
  await assert.rejects(harness.collect(), /投影配置未登记完整/);
});

async function createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords() {
  const harness=createIsolatedMaintenanceConsoleHarness();
  harness.findElement('stopped-product-read-only-http-query-country').value='AQ';
  harness.findElement('stopped-product-read-only-http-query-currency').value='USD';
  const task={id:'synthetic-product-task',kind:'upload',state:'product_module_switched_still_isolated',expires_at:new Date(Date.now()+3600000).toISOString(),manifest:{scope:'product_catalog_with_images'},recovery_journal:{new_database_oid:42,product_catalog_stopped_switch:{phase:'switched_still_isolated',window_id:'synthetic-window',sha256:'d'.repeat(64)}},product_catalog_stopped_detail_cache_invalidation:{phase:'verified_absent',sha256:'e'.repeat(64)}};
  const status={site_id:'synthetic-store',environment:'disposable',application_fenced:true,permissions:['ops:backup:restore']};
  const report={version:1,scope:'product_catalog_with_images',task_id:task.id,window_id:'synthetic-window',switch_journal_sha256:'d'.repeat(64),detail_cache_progress_sha256:'e'.repeat(64),database_oid:42,application_remains_stopped:true,online_overwrite_available:false,contract_sha256:'f'.repeat(64),requirements:['product_storefront_http_routes','product_public_private_media_delivery','product_html_edge_and_browser_caches'].map(projection=>({projection,action:'synthetic-action-'+projection,description:'独立人工核对；503 不能算通过。'}))};
  harness.context.syntheticProductTask=task;harness.context.syntheticProductStatus=status;harness.context.syntheticProductReport=report;
  vm.runInContext('maintenanceCredential="synthetic-product-token";targetStatus=syntheticProductStatus;displayedMaintenanceTasks=[syntheticProductTask]',harness.context);
  harness.context.postMaintenanceRequest=async()=>({report});
  await vm.runInContext('inspectStoppedProductStorefrontProjectionBaseline(syntheticProductTask.id)',harness.context);
  const fieldsets=harness.findElement('product-catalog-stopped-projection-requirements').querySelectorAll('fieldset');
  for(const fieldset of fieldsets){fieldset.querySelector('[data-projection-confirmed]').checked=true;fieldset.querySelector('[data-projection-reference]').value='synthetic-product-evidence';fieldset.querySelector('[data-projection-record]').files=[{size:18,arrayBuffer:async()=>new TextEncoder().encode('independent-report').buffer}];}
  return {...harness,task,status,report,productFieldsets:fieldsets,submit:()=>harness.findElement('product-catalog-stopped-projection-form').listeners.submit({preventDefault(){}})};
}

test('product projection panel hashes three records separately from whole-storefront evidence and keeps isolation',async()=>{
  const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();
  assert.equal(harness.productFieldsets.length,3);
  assert.equal(harness.findElement('external-projection-reconciliation-evidence').querySelectorAll('fieldset').length,4);
  harness.context.prompt=message=>{assert.match(message,/商城继续停写/);return 'RECORD STOPPED PRODUCT PROJECTIONS synthetic-store/disposable/'+harness.task.id;};
  let saved;
  harness.context.postMaintenanceRequest=async(path,request)=>{saved={path,request};return {record:{task_id:harness.task.id,scope:'product_catalog_with_images',contract_sha256:harness.report.contract_sha256,detail_cache_progress_sha256:harness.report.detail_cache_progress_sha256,verification:'operator_attestation_only_external_results_not_automatically_verified'}};};
  harness.context.refreshMaintenanceTasks=async()=>{};
  await harness.submit();
  assert.match(saved.path,/product-catalog-stopped-projection-evidence$/);
  assert.equal(saved.request.attestations.length,3);
  assert.equal(saved.request.expected_record_sha256,'');
  for(const record of saved.request.attestations)assert.equal(record.evidence_sha256,createHash('sha256').update('independent-report').digest('hex'));
  assert.match(harness.findElement('message').textContent,/不能恢复营业/);
});

for(const scenario of ['unchecked','missing_record','changed_task_during_hash','permission_removed_during_hash','disconnect_during_hash','changed_cache_during_prompt']) {
  test(`product operator evidence refuses incomplete or stale records: ${scenario}`,async()=>{
    const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();
    let posted=0;
    harness.context.postMaintenanceRequest=async()=>{posted++;throw new Error('must not post stale product evidence');};
    if(scenario==='unchecked')harness.productFieldsets[0].querySelector('[data-projection-confirmed]').checked=false;
    if(scenario==='missing_record')harness.productFieldsets[0].querySelector('[data-projection-record]').files=[];
    if(scenario.includes('during_hash'))harness.productFieldsets[0].querySelector('[data-projection-record]').files[0].arrayBuffer=async()=>{
      if(scenario==='changed_task_during_hash')harness.task.recovery_journal.product_catalog_stopped_switch.sha256='changed';
      if(scenario==='permission_removed_during_hash')harness.status.permissions=['ops:backup:view'];
      if(scenario==='disconnect_during_hash')harness.findElement('disconnect').listeners.click();
      return new TextEncoder().encode('independent-report').buffer;
    };
    if(scenario==='changed_cache_during_prompt')harness.context.prompt=()=>{harness.task.product_catalog_stopped_detail_cache_invalidation.sha256='new-observation';return 'synthetic confirmation';};
    await harness.submit();assert.equal(posted,0);
    assert.equal(harness.findElement('product-catalog-stopped-projection-save').disabled,true);
  });
}

for(const scenario of ['disconnect','cache_progress_changes','permission_removed']) {
  test(`pending product baseline is discarded after ${scenario}`,async()=>{
    const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();
    let finish;
    harness.context.postMaintenanceRequest=()=>new Promise(resolve=>{finish=resolve;});
    const pending=vm.runInContext('inspectStoppedProductStorefrontProjectionBaseline(syntheticProductTask.id)',harness.context);
    if(scenario==='disconnect')harness.findElement('disconnect').listeners.click();
    if(scenario==='cache_progress_changes')harness.task.product_catalog_stopped_detail_cache_invalidation.sha256='changed';
    if(scenario==='permission_removed')harness.status.permissions=['ops:backup:view'];
    finish({report:harness.report});await pending;
    assert.equal(harness.findElement('product-catalog-stopped-projection-panel').hidden,true);
    assert.equal(harness.findElement('product-catalog-stopped-projection-save').disabled,true);
  });
}

function createBoundStoppedStorefrontConsoleSyntheticResult(harness,request,status=200) {
  const bytes=Buffer.from(JSON.stringify({code:0,data:{name:'恢复商品',slug:'recovered-product'}}));
  return {version:1,task_id:harness.task.id,scope:'product_catalog_with_images',window_id:harness.report.window_id,contract_sha256:harness.report.contract_sha256,database_oid:42,detail_cache_progress_sha256:harness.report.detail_cache_progress_sha256,sql_read_only:true,application_remains_stopped:true,online_overwrite_available:false,verification:'real_storefront_get_handlers_bound_read_only_sql_without_checkout_snapshots_or_network',request,http_status:status,response_bytes_base64:bytes.toString('base64'),response_sha256:createHash('sha256').update(bytes).digest('hex')};
}

for(const scenario of ['disconnect','changed_cache','permission_removed']) {
  test(`storefront response hashing discards a changed recovery context: ${scenario}`,async()=>{
    const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();
    harness.findElement('stopped-product-read-only-http-query-operation').value='context';
    harness.findElement('stopped-product-read-only-http-query-locale').value='en';
    harness.context.postMaintenanceRequest=async(path,request)=>({result:createBoundStoppedStorefrontConsoleSyntheticResult(harness,request)});
    harness.context.crypto={subtle:{digest:async(algorithm,bytes)=>{
      if(scenario==='disconnect')harness.findElement('disconnect').listeners.click();
      if(scenario==='changed_cache')harness.task.product_catalog_stopped_detail_cache_invalidation.sha256='changed';
      if(scenario==='permission_removed')harness.status.permissions=['ops:backup:view'];
      return webcrypto.subtle.digest(algorithm,bytes);
    }}};
    await harness.findElement('stopped-product-read-only-http-query-form').listeners.submit({preventDefault(){}});
    assert.equal(harness.findElement('stopped-product-read-only-http-query-result').textContent,'');
    assert.equal(harness.findElement('stopped-product-read-only-http-query-submit').disabled,true);
  });
}

for(const operation of ['detail','list','search','missing_detail','context','site_settings','categories','redirects','seo_home','social_settings','website_profile','website_name','specification_templates','currency_policy','product_page_faq','product_reviews','product_review_summary']) {
  test(`stopped product read-only HTTP query uses the current baseline and preserves status: ${operation}`,async()=>{
    const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();
    const queryOperation=operation==='missing_detail'?'detail':operation;
    harness.findElement('stopped-product-read-only-http-query-operation').value=queryOperation;
    harness.findElement('stopped-product-read-only-http-query-locale').value='en';
    harness.findElement('stopped-product-read-only-http-query-slug').value='recovered-product';
    harness.findElement('stopped-product-read-only-http-query-keyword').value='product';
    harness.findElement('stopped-product-read-only-http-query-product-id').value='42';
    harness.findElement('stopped-product-read-only-http-query-page').value='2';
    harness.findElement('stopped-product-read-only-http-query-page-size').value='5';
    let posted;
    harness.context.postMaintenanceRequest=async(path,request)=>{
      posted={path,request};
      return {result:createBoundStoppedStorefrontConsoleSyntheticResult(harness,request,operation==='missing_detail'?404:200)};
    };
    await harness.findElement('stopped-product-read-only-http-query-form').listeners.submit({preventDefault(){}});
    assert.match(posted.path,/product-catalog-stopped-read-only-storefront-http-query$/);
    assert.equal(posted.request.contract_sha256,harness.report.contract_sha256);
    assert.equal(posted.request.slug,queryOperation==='detail'?'recovered-product':'');
    assert.equal(posted.request.keyword,queryOperation==='search'?'product':'');
    assert.equal(posted.request.country,'AQ');assert.equal(posted.request.currency,'USD');
    assert.equal(posted.request.product_id,queryOperation.startsWith('product_review')?42:undefined);
    assert.equal(posted.request.page,queryOperation==='product_reviews'?2:undefined);
    assert.equal(posted.request.page_size,queryOperation==='product_reviews'?5:undefined);
    assert.match(harness.findElement('stopped-product-read-only-http-query-status').textContent,operation==='missing_detail'?/HTTP 404/:/HTTP 200/);
    assert.match(harness.findElement('stopped-product-read-only-http-query-status').textContent,/仍待完成/);
    assert.equal(harness.findElement('stopped-product-read-only-http-query-submit').disabled,false);
  });
}

for(const scenario of ['zero_product','oversized_product','zero_page','oversized_page','oversized_page_size','mismatched_product_id','mismatched_page','mismatched_page_size','extra_response_selector']){
  test(`stopped product review console refuses invalid selectors and mismatched response bindings: ${scenario}`,async()=>{
    const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();
    harness.findElement('stopped-product-read-only-http-query-operation').value='product_reviews';
    harness.findElement('stopped-product-read-only-http-query-locale').value='en';
    harness.findElement('stopped-product-read-only-http-query-product-id').value=scenario==='zero_product'?'0':scenario==='oversized_product'?'4294967296':'42';
    harness.findElement('stopped-product-read-only-http-query-page').value=scenario==='zero_page'?'0':scenario==='oversized_page'?'1001':'2';
    harness.findElement('stopped-product-read-only-http-query-page-size').value=scenario==='oversized_page_size'?'21':'5';
    let calls=0;
    harness.context.postMaintenanceRequest=async(path,request)=>{
      calls++;
      const result=createBoundStoppedStorefrontConsoleSyntheticResult(harness,request);
      result.request={...request};
      if(scenario.startsWith('mismatched_'))result.request[scenario.slice('mismatched_'.length)]++;
      if(scenario==='extra_response_selector')result.request.status='pending';
      return {result};
    };
    await harness.findElement('stopped-product-read-only-http-query-form').listeners.submit({preventDefault(){}});
    assert.equal(calls,scenario.startsWith('mismatched_')||scenario==='extra_response_selector'?1:0);
    assert.equal(harness.findElement('stopped-product-read-only-http-query-result').textContent,'');
    assert.equal(harness.findElement('stopped-product-read-only-http-query-submit').disabled,true);
  });
}

for(const scenario of ['disconnect','changed_task','changed_cache','permission_removed','new_baseline']) {
  test(`late stopped product read-only HTTP results are discarded: ${scenario}`,async()=>{
    const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();
    harness.findElement('stopped-product-read-only-http-query-operation').value='list';
    harness.findElement('stopped-product-read-only-http-query-locale').value='en';
    let finish;
    harness.context.postMaintenanceRequest=()=>new Promise(resolve=>{finish=resolve;});
    const pending=harness.findElement('stopped-product-read-only-http-query-form').listeners.submit({preventDefault(){}});
    if(scenario==='disconnect')harness.findElement('disconnect').listeners.click();
    if(scenario==='changed_task')harness.task.state='rolled_back';
    if(scenario==='changed_cache')harness.task.product_catalog_stopped_detail_cache_invalidation.sha256='new-cache-observation';
    if(scenario==='permission_removed')harness.status.permissions=['ops:backup:view'];
    if(scenario==='new_baseline')vm.runInContext('clearStoppedProductProjectionEvidencePanel()',harness.context);
    finish({result:{task_id:harness.task.id,response:{secret:'stale-result-must-never-render'}}});
    await pending;
    assert.equal(harness.findElement('stopped-product-read-only-http-query-result').textContent,'');
    assert.equal(harness.findElement('stopped-product-read-only-http-query-submit').disabled,true);
  });
}

for(const scenario of ['wrong_contract','wrong_window','writeable_transaction','claims_release','changed_query','changed_country','changed_currency','wrong_digest','legacy_handler_result']) {
  test(`stopped product read-only HTTP results reject inconsistent binding: ${scenario}`,async()=>{
    const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();
    harness.findElement('stopped-product-read-only-http-query-operation').value='list';
    harness.findElement('stopped-product-read-only-http-query-locale').value='en';
    harness.context.postMaintenanceRequest=async(path,request)=>{
      const result=createBoundStoppedStorefrontConsoleSyntheticResult(harness,{...request});
      if(scenario==='wrong_contract')result.contract_sha256='different';
      if(scenario==='wrong_window')result.window_id='different';
      if(scenario==='writeable_transaction')result.sql_read_only=false;
      if(scenario==='claims_release')result.online_overwrite_available=true;
      if(scenario==='changed_query')result.request.operation='detail';
      if(scenario==='changed_country')result.request.country='US';
      if(scenario==='changed_currency')result.request.currency='EUR';
      if(scenario==='wrong_digest')result.response_sha256='a'.repeat(64);
      if(scenario==='legacy_handler_result')result.verification='in_process_public_product_http_handler_enforced_sql_read_only_no_network';
      return {result};
    };
    await harness.findElement('stopped-product-read-only-http-query-form').listeners.submit({preventDefault(){}});
    assert.equal(harness.findElement('stopped-product-read-only-http-query-result').textContent,'');
    assert.match(harness.findElement('message').textContent,/不一致/);
    assert.equal(harness.findElement('product-catalog-stopped-projection-panel').hidden,true);
  });
}

function configureStoppedProductMediaQueryHarness(harness) {
  for(const [name,value] of Object.entries({operation:'conversation_attachment',conversation:'guest-conversation',message:'52',index:'0',preset:'thumbnail',range:'bytes=0-7','visitor-cookie':'synthetic-existing-signed-cookie'}))harness.findElement('stopped-product-read-only-media-query-'+name).value=value;
}
function createStoppedProductMediaQueryResponse(harness,request) {
  return {version:2,task_id:harness.task.id,scope:'product_catalog_with_images',window_id:harness.report.window_id,contract_sha256:harness.report.contract_sha256,database_oid:42,detail_cache_progress_sha256:harness.report.detail_cache_progress_sha256,sql_read_only:true,application_remains_stopped:true,online_overwrite_available:false,verification:'in_process_business_local_media_http_handlers_active_account_role_assignment_read_only_no_network',visitor_credential_supplied:!!request.visitor_cookie,storefront_credential_supplied:!!request.storefront_auth_cookie,admin_credential_supplied:!!request.admin_auth_cookie,target:{...request,visitor_cookie:'',storefront_auth_cookie:'',admin_auth_cookie:''},http_status:206,response_bytes:8,response_sha256:'a'.repeat(64)};
}
test('stopped media query clears visitor credential and only renders bound metadata',async()=>{
  const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();configureStoppedProductMediaQueryHarness(harness);
  harness.context.postMaintenanceRequest=async(path,request)=>{assert.match(path,/read-only-media-http-query$/);assert.equal(request.visitor_cookie,'synthetic-existing-signed-cookie');return {result:createStoppedProductMediaQueryResponse(harness,request)};};
  await harness.findElement('stopped-product-read-only-media-query-form').listeners.submit({preventDefault(){}});
  assert.match(harness.findElement('stopped-product-read-only-media-query-status').textContent,/HTTP 206/);
  assert.equal(harness.findElement('stopped-product-read-only-media-query-visitor-cookie').value,'');
  assert(!harness.findElement('stopped-product-read-only-media-query-result').textContent.includes('synthetic-existing-signed-cookie'));
});

for(const operation of ['public_upload','conversation_attachment','agent_conversation_attachment'])test('media query scopes and immediately clears all three business cookie inputs: '+operation,async()=>{
  const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();configureStoppedProductMediaQueryHarness(harness);
  harness.findElement('stopped-product-read-only-media-query-operation').value=operation;
  harness.findElement('stopped-product-read-only-media-query-conversation').value=operation==='agent_conversation_attachment'?'123':'member-chat';
  harness.findElement('stopped-product-read-only-media-query-object-key').value='products/image.png';
  harness.findElement('stopped-product-read-only-media-query-storefront-cookie').value='synthetic-existing-member-cookie';
  harness.findElement('stopped-product-read-only-media-query-admin-cookie').value='synthetic-existing-admin-cookie';
  let finish,submitted;
  harness.context.postMaintenanceRequest=async(path,request)=>{
    submitted=request;
    assert.equal(request.visitor_cookie,operation==='conversation_attachment'?'synthetic-existing-signed-cookie':'');
    assert.equal(request.storefront_auth_cookie,operation==='conversation_attachment'?'synthetic-existing-member-cookie':'');
    assert.equal(request.admin_auth_cookie,operation==='agent_conversation_attachment'?'synthetic-existing-admin-cookie':'');
    const result=createStoppedProductMediaQueryResponse(harness,request);
    return new Promise(resolve=>{finish=()=>resolve({result});});
  };
  const pending=harness.findElement('stopped-product-read-only-media-query-form').listeners.submit({preventDefault(){}});
  for(const field of ['visitor','storefront','admin'])assert.equal(harness.findElement('stopped-product-read-only-media-query-'+field+'-cookie').value,'');
  for(const field of ['visitor_cookie','storefront_auth_cookie','admin_auth_cookie'])assert.equal(submitted[field],'');
  finish();await pending;
  assert.match(harness.findElement('stopped-product-read-only-media-query-status').textContent,/HTTP 206/);
  assert(!harness.findElement('stopped-product-read-only-media-query-result').textContent.includes('synthetic-existing-'));
});

for(const operation of ['conversation_attachment','agent_conversation_attachment'])for(const scenario of ['wrong_member_flag','wrong_admin_flag','echoed_member_cookie','echoed_admin_cookie','old_guest_version','failed_request','lost_permission'])test('member and agent query refuses unsafe or stale results: '+operation+'/'+scenario,async()=>{
  const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();configureStoppedProductMediaQueryHarness(harness);
  harness.findElement('stopped-product-read-only-media-query-operation').value=operation;
  harness.findElement('stopped-product-read-only-media-query-storefront-cookie').value='synthetic-member-secret';
  harness.findElement('stopped-product-read-only-media-query-admin-cookie').value='synthetic-admin-secret';
  let finish;
  harness.context.postMaintenanceRequest=async(path,request)=>{
    const result=createStoppedProductMediaQueryResponse(harness,request);
    if(scenario==='wrong_member_flag')result.storefront_credential_supplied=!result.storefront_credential_supplied;
    if(scenario==='wrong_admin_flag')result.admin_credential_supplied=!result.admin_credential_supplied;
    if(scenario==='echoed_member_cookie')result.target.storefront_auth_cookie='synthetic-member-secret';
    if(scenario==='echoed_admin_cookie')result.target.admin_auth_cookie='synthetic-admin-secret';
    if(scenario==='old_guest_version')result.version=1;
    if(scenario==='failed_request')throw new Error('synthetic request rejected');
    if(scenario==='lost_permission')return new Promise(resolve=>{finish=()=>resolve({result});});
    return {result};
  };
  const pending=harness.findElement('stopped-product-read-only-media-query-form').listeners.submit({preventDefault(){}});
  if(scenario==='lost_permission'){vm.runInContext('targetStatus.permissions=[]',harness.context);finish();}
  await pending;
  assert.equal(harness.findElement('stopped-product-read-only-media-query-result').textContent,'');
  for(const field of ['visitor','storefront','admin'])assert.equal(harness.findElement('stopped-product-read-only-media-query-'+field+'-cookie').value,'');
});
for(const scenario of ['different_task','lost_permission','cache_changed','disconnect','new_baseline'])test('late media response discarded: '+scenario,async()=>{
  const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();configureStoppedProductMediaQueryHarness(harness);let finish;
  harness.context.postMaintenanceRequest=async(path,request)=>{const result=createStoppedProductMediaQueryResponse(harness,request);return new Promise(resolve=>{finish=()=>resolve({result})});};
  const pending=harness.findElement('stopped-product-read-only-media-query-form').listeners.submit({preventDefault(){}});
  assert.equal(harness.findElement('stopped-product-read-only-media-query-visitor-cookie').value,'');
  if(scenario==='different_task')vm.runInContext('displayedMaintenanceTasks=[]',harness.context);
  if(scenario==='lost_permission')vm.runInContext('targetStatus.permissions=[]',harness.context);
  if(scenario==='cache_changed')vm.runInContext('displayedMaintenanceTasks[0].product_catalog_stopped_detail_cache_invalidation.sha256="changed"',harness.context);
  if(scenario==='disconnect')vm.runInContext('maintenanceCredential=""',harness.context);
  if(scenario==='new_baseline')vm.runInContext('clearStoppedProductProjectionEvidencePanel()',harness.context);
  finish();await pending;
  assert.equal(harness.findElement('stopped-product-read-only-media-query-result').textContent,'');assert.equal(harness.findElement('stopped-product-read-only-media-query-submit').disabled,true);
});
for(const scenario of ['wrong_contract','wrong_window','writeable_transaction','claims_release','changed_target','echoed_cookie'])test('media response refuses invalid binding or credential reflection: '+scenario,async()=>{
  const harness=await createStoppedProductProjectionConsoleHarnessWithThreeIndependentRecords();configureStoppedProductMediaQueryHarness(harness);
  harness.context.postMaintenanceRequest=async(path,request)=>{const result=createStoppedProductMediaQueryResponse(harness,request);
    if(scenario==='wrong_contract')result.contract_sha256='wrong';if(scenario==='wrong_window')result.window_id='wrong';if(scenario==='writeable_transaction')result.sql_read_only=false;if(scenario==='claims_release')result.online_overwrite_available=true;if(scenario==='changed_target')result.target.message_id='other';if(scenario==='echoed_cookie')result.target.visitor_cookie=request.visitor_cookie;return {result};};
  await harness.findElement('stopped-product-read-only-media-query-form').listeners.submit({preventDefault(){}});
  assert.equal(harness.findElement('stopped-product-read-only-media-query-result').textContent,'');assert.match(harness.findElement('message').textContent,/不一致/);assert.equal(harness.findElement('stopped-product-read-only-media-query-visitor-cookie').value,'');
});
