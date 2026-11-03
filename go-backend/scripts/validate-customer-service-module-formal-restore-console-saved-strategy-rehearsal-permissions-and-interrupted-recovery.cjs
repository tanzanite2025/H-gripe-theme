'use strict';
const { loadMaintenanceConsoleScriptsInDocumentOrder } = require('./load-maintenance-console-classic-scripts-in-html-document-order.cjs');
const assert=require('node:assert/strict');
const vm=require('node:vm');
const {test}=require('node:test');

class SyntheticCustomerServiceFormalRestoreConsoleElement {
  constructor(){this.children=[];this.dataset={};this.listeners={};this.value='';this.textContent='';this.hidden=false;}
  append(...children){this.children.push(...children);}
  replaceChildren(...children){this.children=children;}
  addEventListener(type,listener){this.listeners[type]=listener;}
  scrollIntoView(){}
}
function createCustomerServiceModuleFormalRestoreConsoleHarness(){
  const elements=new Map();
  const findElement=id=>{if(!elements.has(id))elements.set(id,new SyntheticCustomerServiceFormalRestoreConsoleElement());return elements.get(id);};
  const browserWindow={addEventListener(){}};browserWindow.parent=browserWindow;
  const context=vm.createContext({document:{getElementById:findElement,createElement:()=>new SyntheticCustomerServiceFormalRestoreConsoleElement(),querySelectorAll:()=>[],querySelector:()=>null},location:{pathname:'/local-backup-maintenance/',origin:'https://synthetic.invalid'},window:browserWindow,Headers,console,clearInterval(){},setInterval(){return 1;},fetch:async()=>{throw new Error('unexpected network request');}});
  loadMaintenanceConsoleScriptsInDocumentOrder(context);
  const strategy={policy:'use_package_conversation_rows',complete:true,domain_checks_satisfied:true,blocking_issues:0};
  const coverage={prerequisites_satisfied:true,package_rows:strategy,preserve_current_rows:{...strategy,policy:'preserve_current_conflicting_conversation_rows'}};
  const task={id:'synthetic-chat-task',kind:'upload',state:'module_chat_candidate_rehearsal_verified',size:100,uploaded_bytes:100,expires_at:new Date(Date.now()+3600000).toISOString(),manifest:{scope:'selected_business_modules',selected_modules:['customer_service']},selected_module_restore_preview:{sha256:'a'.repeat(64),binding:{},tables:[],limitations:[],target_unchanged:true,disposable_stage_removed:true,online_overwrite_available:false,dependency_conflicts:0,file_conflicts:0,customer_service_overwrite_preview:coverage},selected_customer_service_restore_choice:{state:'recorded_not_executable',policy:'use_package_conversation_rows',sha256:'b'.repeat(64),preview_sha256:'a'.repeat(64),execution_available:false},selected_customer_service_candidate_rehearsal:{sha256:'c'.repeat(64),binding:{},saved_choice_sha256:'b'.repeat(64),customer_service_formal_restore_policy:'restore_selected_conversation_rows_preserve_current_shared_facts_v1',target_unchanged:true,disposable_stage_removed:true,isolated_restore_verified:true,online_overwrite_available:false,dependency_conflicts:0,file_conflicts:0,customer_service_overwrite_preview:coverage}};
  const status={site_id:'synthetic-store',environment:'disposable',permissions:['ops:backup:restore','ops:backup:upload'],application_fenced:false};
  Object.assign(context,{syntheticTask:task,syntheticStatus:status});
  vm.runInContext('maintenanceCredential="synthetic-token";targetStatus=syntheticStatus;displayedMaintenanceTasks=[syntheticTask]',context);
  const collect=element=>[element.textContent,...element.children.map(collect)].join('\n');
  return {context,task,status,run:source=>vm.runInContext(source,context),text:id=>collect(findElement(id)),findElement};
}

for(const scenario of ['eligible_use_package','eligible_preserve_current','view_only','fenced','expired','invalid_expiry','work_expired','mixed_scope','missing_choice','changed_choice','changed_original_preview','missing_rehearsal_digest','old_rehearsal_policy','unverified_rehearsal','retained_disposable_stage','order_journal','already_started','running','domain_blocked','dependency_conflict','file_conflict','invalid_strategy']){
  test(`formal chat restore requires a fresh saved strategy and rehearsal: ${scenario}`,()=>{
    const h=createCustomerServiceModuleFormalRestoreConsoleHarness();
    if(scenario==='eligible_preserve_current')h.task.selected_customer_service_restore_choice.policy='preserve_current_conflicting_conversation_rows';
    if(scenario==='view_only')h.status.permissions=['ops:backup:view'];
    if(scenario==='fenced')h.status.application_fenced=true;
    if(scenario==='expired')h.task.expires_at=new Date(0).toISOString();
    if(scenario==='invalid_expiry')h.task.expires_at='invalid';
    if(scenario==='work_expired')h.task.work_copy_expired=true;
    if(scenario==='mixed_scope')h.task.manifest.selected_modules=['customer_service','orders'];
    if(scenario==='missing_choice')delete h.task.selected_customer_service_restore_choice;
    if(scenario==='changed_choice')h.task.selected_customer_service_restore_choice.sha256='d'.repeat(64);
    if(scenario==='changed_original_preview')h.task.selected_module_restore_preview.sha256='d'.repeat(64);
    if(scenario==='missing_rehearsal_digest')delete h.task.selected_customer_service_candidate_rehearsal.sha256;
    if(scenario==='old_rehearsal_policy')delete h.task.selected_customer_service_candidate_rehearsal.customer_service_formal_restore_policy;
    if(scenario==='unverified_rehearsal')h.task.selected_customer_service_candidate_rehearsal.isolated_restore_verified=false;
    if(scenario==='retained_disposable_stage')h.task.selected_module_disposable_stage={database_oid:1};
    if(scenario==='order_journal')h.task.order_module_restore={phase:'committing_rows'};
    if(scenario==='already_started')h.task.customer_service_module_restore={phase:'installing_files'};
    if(scenario==='running')h.task.state='running';
    if(scenario==='domain_blocked')h.task.selected_customer_service_candidate_rehearsal.customer_service_overwrite_preview.package_rows.domain_checks_satisfied=false;
    if(scenario==='dependency_conflict')h.task.selected_customer_service_candidate_rehearsal.dependency_conflicts=1;
    if(scenario==='file_conflict')h.task.selected_customer_service_candidate_rehearsal.file_conflicts=1;
    if(scenario==='invalid_strategy')h.task.selected_customer_service_restore_choice.policy='unregistered';
    h.run('renderBackupTasks([syntheticTask])');
    assert.equal(h.text('tasks').includes('按所选策略恢复聊天（暂停商城）'),scenario.startsWith('eligible_'));
  });
}
for(const scenario of ['interrupted','expired','view_only','running','released','mixed_journal']){
  test(`interrupted chat restore retains its dedicated recovery entry: ${scenario}`,()=>{
    const h=createCustomerServiceModuleFormalRestoreConsoleHarness();h.task.state='interrupted';h.status.application_fenced=true;h.task.customer_service_module_restore={phase:'committing_rows'};
    if(scenario==='expired'){h.task.expires_at=new Date(0).toISOString();h.task.work_copy_expired=true;}
    if(scenario==='view_only')h.status.permissions=['ops:backup:view'];
    if(scenario==='running')h.task.state='running';
    if(scenario==='released'){h.task.state='completed';h.task.customer_service_module_restore={phase:'released',outcome:'restored',restored_rows:2,replaced_rows:1};}
    if(scenario==='mixed_journal')h.task.order_module_restore={phase:'committing_rows'};
    h.run('renderBackupTasks([syntheticTask])');
    const text=h.text('tasks');
    assert.equal(text.includes('核验并恢复聊天任务'),['interrupted','expired'].includes(scenario));
    for(const ordinary of ['解除未改数据的维护隔离','校验包并检查可用恢复方式','验证模块恢复（暂停商城）','按所选策略恢复聊天（暂停商城）'])assert(!text.includes(ordinary));
    if(scenario==='released')assert(text.includes('补入 2 条、替换 1 条完整记录'));
  });
}
for(const scenario of ['submit','preserve_current','cancel','changed_choice','changed_rehearsal','changed_task','changed_status','changed_session','disconnect_after_submit','interrupted_expired']){
  test(`formal chat submission binds confirmation to the displayed choice, rehearsal and session: ${scenario}`,async()=>{
    const h=createCustomerServiceModuleFormalRestoreConsoleHarness();const recover=scenario==='interrupted_expired';
    if(scenario==='preserve_current')h.task.selected_customer_service_restore_choice.policy='preserve_current_conflicting_conversation_rows';
    if(recover){h.task.state='interrupted';h.task.customer_service_module_restore={phase:'committing_rows'};h.task.expires_at=new Date(0).toISOString();h.status.application_fenced=true;}
    let posted;
    h.context.prompt=message=>{
      assert(message.includes(recover?'核验聊天恢复':'将暂停商城'));
      if(scenario==='preserve_current')assert(message.includes('保留当前存在冲突的聊天记录'));
      if(scenario==='cancel')return null;
      if(scenario==='changed_choice')h.task.selected_customer_service_restore_choice.sha256='d'.repeat(64);
      if(scenario==='changed_rehearsal')h.task.selected_customer_service_candidate_rehearsal.sha256='d'.repeat(64);
      if(scenario==='changed_task')h.task.state='failed';
      if(scenario==='changed_status')h.status.application_fenced=true;
      if(scenario==='changed_session')h.findElement('disconnect').listeners.click();
      return recover?`RECOVER CHAT RESTORE synthetic-store/disposable/${h.task.id}`:`RESTORE CHAT MODULE synthetic-store/disposable/${h.task.id}/${'b'.repeat(64)}`;
    };
    h.context.postMaintenanceRequest=async(url,body)=>{posted={url,body};if(scenario==='disconnect_after_submit')h.findElement('disconnect').listeners.click();return {};};
    const operation=h.run(`startCustomerServiceModuleFormalRestoreOrInterruptedRecovery(syntheticTask.id,${recover})`);
    if(scenario.startsWith('changed_')){await assert.rejects(operation);assert.equal(posted,undefined);}
    else {await operation;if(scenario==='cancel')assert.equal(posted,undefined);else {assert(posted.url.endsWith(recover?'customer-service-module-interrupted-restore-recovery':'customer-service-module-formal-restore'));if(!recover){assert.equal(posted.body.expected_choice_sha256,'b'.repeat(64));assert.equal(posted.body.expected_rehearsal_sha256,'c'.repeat(64));}}}
    if(scenario==='disconnect_after_submit')assert(!h.text('message').includes('正在按所选策略恢复聊天'));
  });
}
