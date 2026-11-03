'use strict';
const { loadMaintenanceConsoleScriptsInDocumentOrder } = require('./load-maintenance-console-classic-scripts-in-html-document-order.cjs');
const assert=require('node:assert/strict');
const vm=require('node:vm');
const {test}=require('node:test');

class SyntheticMediaRestoreConsoleElement {
  constructor(){this.children=[];this.dataset={};this.listeners={};this.value='';this.textContent='';this.hidden=false;}
  append(...children){this.children.push(...children);}
  replaceChildren(...children){this.children=children;}
  addEventListener(type,listener){this.listeners[type]=listener;}
  scrollIntoView(){}
}
function createMediaModuleFormalRestoreConsoleHarness(){
  const elements=new Map();
  const findElement=id=>{if(!elements.has(id))elements.set(id,new SyntheticMediaRestoreConsoleElement());return elements.get(id);};
  const browserWindow={addEventListener(){}};browserWindow.parent=browserWindow;
  const context=vm.createContext({document:{getElementById:findElement,createElement:()=>new SyntheticMediaRestoreConsoleElement(),querySelectorAll:()=>[],querySelector:()=>null},location:{pathname:'/local-backup-maintenance/',origin:'https://synthetic.invalid'},window:browserWindow,Headers,console,clearInterval(){},setInterval(){return 1;},fetch:async()=>{throw new Error('unexpected network request');}});
  loadMaintenanceConsoleScriptsInDocumentOrder(context);
  const task={id:'synthetic-media-task',kind:'upload',state:'module_isolated_restore_verified',size:100,uploaded_bytes:100,expires_at:new Date(Date.now()+3600000).toISOString(),manifest:{scope:'selected_business_modules',selected_modules:['media']},selected_module_restore_preview:{modules:['media'],sha256:'a'.repeat(64),binding:{},tables:[],limitations:[],online_overwrite_available:false,target_unchanged:true,disposable_stage_removed:true,isolated_restore_verified:true,media_restore_eligibility:{eligible:true,policy:'restore_missing_media_and_repair_registered_file_bytes_preserve_current_business_data_v2'}}};
  const status={site_id:'synthetic-store',environment:'disposable',permissions:['ops:backup:restore','ops:backup:upload'],application_fenced:false};
  Object.assign(context,{syntheticTask:task,syntheticStatus:status});
  vm.runInContext('maintenanceCredential="synthetic-token";targetStatus=syntheticStatus;displayedMaintenanceTasks=[syntheticTask]',context);
  const collect=element=>[element.textContent,...element.children.map(collect)].join('\n');
  return {context,task,status,run:source=>vm.runInContext(source,context),text:id=>collect(findElement(id)),findElement};
}

for(const scenario of ['eligible','view_only','fenced','expired','invalid_expiry','work_expired','mixed_scope','missing_digest','ineligible','old_policy','running','already_started','chat_journal','order_journal','retained_stage','wrong_report_scope','negative_repairs','missing_repair_digest','unexpected_repair_digest','repair_eligible']){
  test(`media restore requires the current isolated preview and permission: ${scenario}`,()=>{
    const h=createMediaModuleFormalRestoreConsoleHarness();
    if(scenario==='view_only')h.status.permissions=['ops:backup:view'];
    if(scenario==='fenced')h.status.application_fenced=true;
    if(scenario==='expired')h.task.expires_at=new Date(0).toISOString();
    if(scenario==='invalid_expiry')h.task.expires_at='invalid';
    if(scenario==='work_expired')h.task.work_copy_expired=true;
    if(scenario==='mixed_scope')h.task.manifest.selected_modules=['media','orders'];
    if(scenario==='missing_digest')delete h.task.selected_module_restore_preview.sha256;
    if(scenario==='ineligible')h.task.selected_module_restore_preview.media_restore_eligibility.eligible=false;
    if(scenario==='old_policy')delete h.task.selected_module_restore_preview.media_restore_eligibility.policy;
    if(scenario==='running')h.task.state='running';
    if(scenario==='already_started')h.task.media_module_restore={phase:'installing_files'};
    if(scenario==='chat_journal')h.task.customer_service_module_restore={phase:'committing_rows'};
    if(scenario==='order_journal')h.task.order_module_restore={phase:'committing_rows'};
    if(scenario==='retained_stage')h.task.selected_module_disposable_stage={database_oid:1};
    if(scenario==='wrong_report_scope')h.task.selected_module_restore_preview.modules=['customer_service'];
    if(scenario==='negative_repairs')h.task.selected_module_restore_preview.media_restore_eligibility.existing_file_repairs=-1;
    if(scenario==='missing_repair_digest')h.task.selected_module_restore_preview.media_restore_eligibility.existing_file_repairs=4;
    if(scenario==='unexpected_repair_digest')h.task.selected_module_restore_preview.media_restore_eligibility.existing_file_repair_plan_sha256='b'.repeat(64);
    if(scenario==='repair_eligible')Object.assign(h.task.selected_module_restore_preview.media_restore_eligibility,{existing_file_repairs:4,existing_file_repair_plan_sha256:'b'.repeat(64)});
    h.run('renderBackupTasks([syntheticTask])');
    assert.equal(h.text('tasks').includes('恢复图片与附件（暂停商城）'),scenario==='eligible'||scenario==='repair_eligible');
  });
}
for(const scenario of ['interrupted','expired','view_only','running','released','mixed_order_journal','mixed_chat_journal']){
  test(`media interruption recovery retains its dedicated entry: ${scenario}`,()=>{
    const h=createMediaModuleFormalRestoreConsoleHarness();h.task.state='interrupted';h.status.application_fenced=true;h.task.media_module_restore={phase:'committing_rows'};
    if(scenario==='expired'){h.task.expires_at=new Date(0).toISOString();h.task.work_copy_expired=true;}
    if(scenario==='view_only')h.status.permissions=['ops:backup:view'];
    if(scenario==='running')h.task.state='running';
    if(scenario==='released'){h.task.state='completed';h.task.media_module_restore={phase:'released',outcome:'restored',restored_rows:12};}
    if(scenario==='mixed_order_journal')h.task.order_module_restore={phase:'committing_rows'};
    if(scenario==='mixed_chat_journal')h.task.customer_service_module_restore={phase:'committing_rows'};
    h.run('renderBackupTasks([syntheticTask])');const text=h.text('tasks');
    assert.equal(text.includes('核验并恢复图片与附件任务'),['interrupted','expired'].includes(scenario));
    for(const ordinary of ['解除未改数据的维护隔离','校验包并检查可用恢复方式','验证模块恢复（暂停商城）','恢复图片与附件（暂停商城）'])assert(!text.includes(ordinary));
    if(scenario==='released')assert(text.includes('补回 12 条媒体记录'));
  });
}
for(const scenario of ['submit','repair_submit','cancel','changed_preview','changed_task','changed_status','changed_session','disconnect_after_submit','interrupted_expired']){
  test(`media restore submission binds confirmation to the displayed preview and session: ${scenario}`,async()=>{
    const h=createMediaModuleFormalRestoreConsoleHarness();const recover=scenario==='interrupted_expired';
    if(scenario==='repair_submit')Object.assign(h.task.selected_module_restore_preview.media_restore_eligibility,{existing_file_repairs:4,existing_file_repair_plan_sha256:'b'.repeat(64)});
    if(recover){h.task.state='interrupted';h.task.media_module_restore={phase:'committing_rows'};h.task.expires_at=new Date(0).toISOString();h.status.application_fenced=true;}
    let posted;
    h.context.prompt=message=>{
      assert(message.includes(recover?'核验本次媒体恢复':'保留当前商品、订单、客户、聊天和配置'));
      if(scenario==='repair_submit'){assert(message.includes('修复 4 个'));assert(message.includes('原字节保留至提交核验'));}
      if(scenario==='cancel')return null;
      if(scenario==='changed_preview')h.task.selected_module_restore_preview.sha256='b'.repeat(64);
      if(scenario==='changed_task')h.task.state='failed';
      if(scenario==='changed_status')h.status.application_fenced=true;
      if(scenario==='changed_session')h.findElement('disconnect').listeners.click();
      return `${recover?'RECOVER MEDIA':'RESTORE MEDIA'} synthetic-store/disposable/${h.task.id}`;
    };
    h.context.postMaintenanceRequest=async(url,body)=>{posted={url,body};if(scenario==='disconnect_after_submit')h.findElement('disconnect').listeners.click();return {};};
    const operation=h.run(`startMediaModuleFormalRestoreOrInterruptedRecovery(syntheticTask.id,${recover})`);
    if(scenario.startsWith('changed_')){await assert.rejects(operation);assert.equal(posted,undefined);}
    else {await operation;if(scenario==='cancel')assert.equal(posted,undefined);else {assert(posted.url.endsWith(recover?'media-module-interrupted-restore-recovery':'media-module-formal-restore'));if(!recover)assert.equal(posted.body.expected_preview_sha256,'a'.repeat(64));}}
    if(scenario==='disconnect_after_submit')assert(!h.text('message').includes('正在恢复缺失图片与附件'));
  });
}

test('completed media byte repair shows repaired files even when no rows were missing',()=>{
  const h=createMediaModuleFormalRestoreConsoleHarness();
  h.task.state='completed';
  h.task.media_module_restore={phase:'released',outcome:'restored',restored_rows:0,files:[{original_identity:{sha256:'b'.repeat(64)}},{original_identity:{sha256:'c'.repeat(64)}},{}]};
  h.run('renderBackupTasks([syntheticTask])');
  assert(h.text('tasks').includes('补回 0 条媒体记录、修复 2 个文件'));
});
