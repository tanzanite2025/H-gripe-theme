'use strict';
const { loadMaintenanceConsoleScriptsInDocumentOrder } = require('./load-maintenance-console-classic-scripts-in-html-document-order.cjs');
const assert=require('node:assert/strict');
const vm=require('node:vm');
const {test}=require('node:test');

class SyntheticConfigurationContentConsoleElement {
  constructor(){this.children=[];this.dataset={};this.listeners={};this.value='';this.textContent='';this.hidden=false;}
  append(...children){this.children.push(...children);}
  replaceChildren(...children){this.children=children;}
  addEventListener(type,listener){this.listeners[type]=listener;}
  scrollIntoView(){}
}
function createConfigurationContentRestoreConsoleHarness(){
  const elements=new Map();
  const findElement=id=>{if(!elements.has(id))elements.set(id,new SyntheticConfigurationContentConsoleElement());return elements.get(id);};
  const browserWindow={addEventListener(){}};browserWindow.parent=browserWindow;
  const context=vm.createContext({document:{getElementById:findElement,createElement:()=>new SyntheticConfigurationContentConsoleElement(),querySelectorAll:()=>[],querySelector:()=>null},location:{pathname:'/local-backup-maintenance/',origin:'https://synthetic.invalid'},window:browserWindow,Headers,console,clearInterval(){},setInterval(){return 1;},fetch:async()=>{throw new Error('unexpected network request');}});
  loadMaintenanceConsoleScriptsInDocumentOrder(context);
  const task={id:'synthetic-configuration-task',kind:'upload',state:'module_isolated_restore_verified',size:100,uploaded_bytes:100,expires_at:new Date(Date.now()+3600000).toISOString(),manifest:{scope:'selected_business_modules',selected_modules:['configuration']},selected_module_restore_preview:{modules:['configuration'],sha256:'a'.repeat(64),binding:{},tables:[],limitations:[],online_overwrite_available:false,target_unchanged:true,disposable_stage_removed:true,isolated_restore_verified:true,configuration_site_content_restore_eligibility:{eligible:true,policy:'restore_missing_business_configuration_including_shipping_reference_packaging_and_customs_preserve_existing_rows_and_registered_files_v6'}}};
  const status={site_id:'synthetic-store',environment:'disposable',permissions:['ops:backup:restore','ops:backup:upload'],application_fenced:false};
  Object.assign(context,{syntheticTask:task,syntheticStatus:status});
  vm.runInContext('maintenanceCredential="synthetic-token";targetStatus=syntheticStatus;displayedMaintenanceTasks=[syntheticTask]',context);
  const collect=element=>[element.textContent,...element.children.map(collect)].join('\n');
  return {context,task,status,run:source=>vm.runInContext(source,context),text:id=>collect(findElement(id)),findElement};
}


for(const scenario of ['eligible','view_only','fenced','expired','mixed_scope','missing_preview','wrong_policy','legacy_policy','legacy_v2_policy','legacy_v3_policy','legacy_v4_policy','already_started','mixed_media_journal','missing_binding']){
  test(`configuration content restore scope and permission: ${scenario}`,()=>{
    const h=createConfigurationContentRestoreConsoleHarness();
    if(scenario==='view_only')h.status.permissions=['ops:backup:view'];
    if(scenario==='fenced')h.status.application_fenced=true;
    if(scenario==='expired')h.task.expires_at=new Date(0).toISOString();
    if(scenario==='mixed_scope')h.task.manifest.selected_modules=['configuration','media'];
    if(scenario==='missing_preview')delete h.task.selected_module_restore_preview;
    if(scenario==='wrong_policy')h.task.selected_module_restore_preview.configuration_site_content_restore_eligibility.policy='arbitrary_config_overwrite';
    if(scenario==='legacy_policy')h.task.selected_module_restore_preview.configuration_site_content_restore_eligibility.policy='restore_missing_site_content_preserve_existing_configuration_and_registered_files_v1';
    if(scenario==='legacy_v2_policy')h.task.selected_module_restore_preview.configuration_site_content_restore_eligibility.policy='restore_missing_site_content_and_fitment_catalog_preserve_existing_configuration_and_registered_files_v2';
    if(scenario==='legacy_v3_policy')h.task.selected_module_restore_preview.configuration_site_content_restore_eligibility.policy='restore_missing_content_catalog_selection_flows_and_questionnaires_preserve_existing_configuration_and_registered_files_v3';
    if(scenario==='legacy_v4_policy')h.task.selected_module_restore_preview.configuration_site_content_restore_eligibility.policy='restore_missing_content_catalog_selection_and_market_currency_tax_preserve_existing_configuration_and_registered_files_v4';
    if(scenario==='already_started')h.task.configuration_site_content_restore={phase:'rows_committed',sha256:'c'.repeat(64)};
    if(scenario==='mixed_media_journal')h.task.media_module_restore={phase:'rows_committed'};
    if(scenario==='missing_binding')delete h.task.selected_module_restore_preview.binding;
    h.run('renderBackupTasks([syntheticTask])');
    assert.equal(h.text('tasks').includes('补回缺失业务配置（暂停商城）'),scenario==='eligible');
  });
}
for(const scenario of ['awaiting_review','expired_review','view_only','running','uncommitted','released','not_fenced','mixed_chat_journal']){
  test(`committed configuration remains isolated until reviewed release: ${scenario}`,()=>{
    const h=createConfigurationContentRestoreConsoleHarness();h.task.state='configuration_site_content_restored_still_isolated';h.status.application_fenced=true;h.task.configuration_site_content_restore={phase:'rows_committed',sha256:'c'.repeat(64),restored_rows:12};
    if(scenario==='expired_review'){h.task.expires_at=new Date(0).toISOString();h.task.work_copy_expired=true;}
    if(scenario==='view_only')h.status.permissions=['ops:backup:view'];
    if(scenario==='running')h.task.state='running';
    if(scenario==='uncommitted')h.task.configuration_site_content_restore.phase='installing_files';
    if(scenario==='released')h.task.configuration_site_content_restore.phase='released';
    if(scenario==='not_fenced')h.status.application_fenced=false;
    if(scenario==='mixed_chat_journal')h.task.customer_service_module_restore={phase:'rows_committed'};
    h.run('renderBackupTasks([syntheticTask])');
    assert.equal(h.text('tasks').includes('记录内容与缓存核对并恢复服务'),['awaiting_review','expired_review'].includes(scenario));
    assert(!h.text('tasks').includes('解除未改数据的维护隔离'));
  });
}
for(const scenario of ['restore','recover','release','cancel','short_review','changed_task','changed_session']){
  test(`configuration confirmation binds current task and session: ${scenario}`,async()=>{
    const h=createConfigurationContentRestoreConsoleHarness();let posted,prompts=0;
    const release=['release','short_review','changed_task','changed_session'].includes(scenario),recover=scenario==='recover';
    if(release||recover){h.task.state=release?'configuration_site_content_restored_still_isolated':'interrupted';h.status.application_fenced=true;h.task.configuration_site_content_restore={phase:'rows_committed',sha256:'c'.repeat(64)};}
    h.context.prompt=message=>{
      prompts++;
      if(scenario==='cancel')return null;
      if(release&&prompts===1){assert(message.includes('不会自动清理缓存'));assert(message.includes('目录与计算结果'));return scenario==='short_review'?'x':'local-review-2026-10-30';}
      if(scenario==='changed_task')h.task.configuration_site_content_restore.sha256='d'.repeat(64);
      if(scenario==='changed_session')h.findElement('disconnect').listeners.click();
      return `${release?'RELEASE':recover?'RECOVER':'RESTORE'} CONFIGURATION CONTENT synthetic-store/disposable/${h.task.id}`;
    };
    h.context.postMaintenanceRequest=async(url,body)=>{posted={url,body};return {};};
    const operation=h.run(`startConfigurationSiteContentRestoreOrReviewedRelease(syntheticTask.id,'${release?'release':recover?'recover':'restore'}')`);
    if(['short_review','changed_task','changed_session'].includes(scenario)){await assert.rejects(operation);assert.equal(posted,undefined);}
    else {await operation;if(scenario==='cancel')assert.equal(posted,undefined);else {assert(posted);if(release){assert.equal(posted.body.expected_journal_sha256,'c'.repeat(64));assert.equal(posted.body.content_and_cache_review_reference,'local-review-2026-10-30');}else if(!recover)assert.equal(posted.body.expected_preview_sha256,'a'.repeat(64));}}
  });
}

test('configuration preview names the bounded scope and post-commit review',()=>{
  const h=createConfigurationContentRestoreConsoleHarness();
  h.run('renderSelectedBusinessModuleIsolatedRestorePreview(syntheticTask)');
  assert(h.text('selected-module-restore-preview').includes('车架/前叉/花鼓、辐条计算和轮胎适配目录'));
  assert(h.text('selected-module-restore-preview').includes('选购流程/问卷'));
  assert(h.text('selected-module-restore-preview').includes('市场/汇率/税率'));
  assert(h.text('selected-module-restore-preview').includes('邮件模板/历史版本和会员等级、承运商/配送区域、包装规格/商品SKU绑定及报关分类'));
  assert(h.text('selected-module-restore-preview').includes('另行核对内容、目录/计算结果与缓存'));
});

test('configuration refresh explains the committed content still awaiting review',async()=>{
  const h=createConfigurationContentRestoreConsoleHarness();
  h.task.state='configuration_site_content_restored_still_isolated';
  h.status.application_fenced=true;
  h.task.configuration_site_content_restore={phase:'rows_committed',sha256:'c'.repeat(64),restored_rows:12};
  h.context.maintenanceRequest=async path=>path==='/api/status'?h.status:{tasks:[h.task]};
  let message;
  h.context.showMessage=(text,isError)=>{message={text,isError};};
  h.run('pendingVerificationTaskID=syntheticTask.id');
  await h.run('refreshMaintenanceTasks()');
  assert(message.text.includes('网站内容已写入，商城继续维护'));
  assert(message.text.includes('内容与缓存核对'));
  assert.equal(message.isError,false);
  assert.equal(h.run('pendingVerificationTaskID'),'');
});
