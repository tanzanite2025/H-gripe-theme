'use strict';
const { loadMaintenanceConsoleScriptsInDocumentOrder } = require('./load-maintenance-console-classic-scripts-in-html-document-order.cjs');
const assert=require('node:assert/strict');
const vm=require('node:vm');
const {test}=require('node:test');

class SyntheticCustomerProfileRestoreConsoleElement {
  constructor(){this.children=[];this.dataset={};this.listeners={};this.value='';this.textContent='';this.hidden=false;}
  append(...children){this.children.push(...children);}
  replaceChildren(...children){this.children=children;}
  addEventListener(type,listener){this.listeners[type]=listener;}
  scrollIntoView(){}
}
function createCustomerProfileModuleFormalRestoreConsoleHarness(){
  const elements=new Map();
  const findElement=id=>{if(!elements.has(id))elements.set(id,new SyntheticCustomerProfileRestoreConsoleElement());return elements.get(id);};
  const browserWindow={addEventListener(){}};browserWindow.parent=browserWindow;
  const context=vm.createContext({document:{getElementById:findElement,createElement:()=>new SyntheticCustomerProfileRestoreConsoleElement(),querySelectorAll:()=>[],querySelector:()=>null},location:{pathname:'/local-backup-maintenance/',origin:'https://synthetic.invalid'},window:browserWindow,Headers,console,clearInterval(){},setInterval(){return 1;},fetch:async()=>{throw new Error('unexpected network request');}});
  loadMaintenanceConsoleScriptsInDocumentOrder(context);
  const task={id:'synthetic-customer-profile-task',kind:'upload',state:'module_isolated_restore_verified',size:100,uploaded_bytes:100,expires_at:new Date(Date.now()+3600000).toISOString(),manifest:{scope:'selected_business_modules',selected_modules:['customers']},selected_module_restore_preview:{modules:['customers'],sha256:'a'.repeat(64),binding:{},tables:[],limitations:[],online_overwrite_available:false,target_unchanged:true,disposable_stage_removed:true,isolated_restore_verified:true,customer_profile_restore_eligibility:{eligible:true,policy:'restore_missing_customer_profiles_and_closed_new_account_facts_preserve_current_identities_v1'}}};
  const status={site_id:'synthetic-store',environment:'disposable',permissions:['ops:backup:restore','ops:backup:upload'],application_fenced:false};
  Object.assign(context,{syntheticTask:task,syntheticStatus:status});
  vm.runInContext('maintenanceCredential="synthetic-token";targetStatus=syntheticStatus;displayedMaintenanceTasks=[syntheticTask]',context);
  const collect=element=>[element.textContent,...element.children.map(collect)].join('\n');
  return {context,task,status,run:source=>vm.runInContext(source,context),text:id=>collect(findElement(id)),findElement};
}

for(const scenario of ['eligible','view_only','fenced','expired','invalid_expiry','work_expired','mixed_scope','missing_digest','ineligible','old_policy','running','already_started','chat_journal','order_journal','retained_stage','wrong_report_scope']){
  test(`customer restore requires the current isolated preview and permission: ${scenario}`,()=>{
    const h=createCustomerProfileModuleFormalRestoreConsoleHarness();
    if(scenario==='view_only')h.status.permissions=['ops:backup:view'];
    if(scenario==='fenced')h.status.application_fenced=true;
    if(scenario==='expired')h.task.expires_at=new Date(0).toISOString();
    if(scenario==='invalid_expiry')h.task.expires_at='invalid';
    if(scenario==='work_expired')h.task.work_copy_expired=true;
    if(scenario==='mixed_scope')h.task.manifest.selected_modules=['customers','orders'];
    if(scenario==='missing_digest')delete h.task.selected_module_restore_preview.sha256;
    if(scenario==='ineligible')h.task.selected_module_restore_preview.customer_profile_restore_eligibility.eligible=false;
    if(scenario==='old_policy')delete h.task.selected_module_restore_preview.customer_profile_restore_eligibility.policy;
    if(scenario==='running')h.task.state='running';
    if(scenario==='already_started')h.task.customer_profile_module_restore={phase:'installing_files'};
    if(scenario==='chat_journal')h.task.customer_service_module_restore={phase:'committing_rows'};
    if(scenario==='order_journal')h.task.order_module_restore={phase:'committing_rows'};
    if(scenario==='retained_stage')h.task.selected_module_disposable_stage={database_oid:1};
    if(scenario==='wrong_report_scope')h.task.selected_module_restore_preview.modules=['customer_service'];
    h.run('renderBackupTasks([syntheticTask])');
    assert.equal(h.text('tasks').includes('恢复缺失客户资料（暂停商城）'),scenario==='eligible');
  });
}
for(const scenario of ['interrupted','expired','view_only','running','released','mixed_order_journal','mixed_chat_journal']){
  test(`customer interruption recovery retains its dedicated entry: ${scenario}`,()=>{
    const h=createCustomerProfileModuleFormalRestoreConsoleHarness();h.task.state='interrupted';h.status.application_fenced=true;h.task.customer_profile_module_restore={phase:'committing_rows'};
    if(scenario==='expired'){h.task.expires_at=new Date(0).toISOString();h.task.work_copy_expired=true;}
    if(scenario==='view_only')h.status.permissions=['ops:backup:view'];
    if(scenario==='running')h.task.state='running';
    if(scenario==='released'){h.task.state='completed';h.task.customer_profile_module_restore={phase:'released',outcome:'restored',restored_rows:12};}
    if(scenario==='mixed_order_journal')h.task.order_module_restore={phase:'committing_rows'};
    if(scenario==='mixed_chat_journal')h.task.customer_service_module_restore={phase:'committing_rows'};
    h.run('renderBackupTasks([syntheticTask])');const text=h.text('tasks');
    assert.equal(text.includes('核验并恢复客户资料任务'),['interrupted','expired'].includes(scenario));
    for(const ordinary of ['解除未改数据的维护隔离','校验包并检查可用恢复方式','验证模块恢复（暂停商城）','恢复缺失客户资料（暂停商城）'])assert(!text.includes(ordinary));
    if(scenario==='released')assert(text.includes('补入 12 条完整记录'));
  });
}
for(const scenario of ['submit','cancel','changed_preview','changed_task','changed_status','changed_session','disconnect_after_submit','interrupted_expired']){
  test(`customer restore submission binds confirmation to the displayed preview and session: ${scenario}`,async()=>{
    const h=createCustomerProfileModuleFormalRestoreConsoleHarness();const recover=scenario==='interrupted_expired';
    if(recover){h.task.state='interrupted';h.task.customer_profile_module_restore={phase:'committing_rows'};h.task.expires_at=new Date(0).toISOString();h.status.application_fenced=true;}
    let posted;
    h.context.prompt=message=>{
      assert(message.includes(recover?'核验本次客户恢复':'当前已有账号、密码、账号限制、积分和订阅状态保留'));
      if(scenario==='cancel')return null;
      if(scenario==='changed_preview')h.task.selected_module_restore_preview.sha256='b'.repeat(64);
      if(scenario==='changed_task')h.task.state='failed';
      if(scenario==='changed_status')h.status.application_fenced=true;
      if(scenario==='changed_session')h.findElement('disconnect').listeners.click();
      return `${recover?'RECOVER CUSTOMER PROFILES':'RESTORE CUSTOMER PROFILES'} synthetic-store/disposable/${h.task.id}`;
    };
    h.context.postMaintenanceRequest=async(url,body)=>{posted={url,body};if(scenario==='disconnect_after_submit')h.findElement('disconnect').listeners.click();return {};};
    const operation=h.run(`startCustomerProfileModuleFormalRestoreOrInterruptedRecovery(syntheticTask.id,${recover})`);
    if(scenario.startsWith('changed_')){await assert.rejects(operation);assert.equal(posted,undefined);}
    else {await operation;if(scenario==='cancel')assert.equal(posted,undefined);else {assert(posted.url.endsWith(recover?'customer-profile-module-interrupted-restore-recovery':'customer-profile-module-formal-restore'));if(!recover)assert.equal(posted.body.expected_preview_sha256,'a'.repeat(64));}}
    if(scenario==='disconnect_after_submit')assert(!h.text('message').includes('正在恢复缺失客户资料'));
  });
}
test('customer preview explains current balance and identity blockers in Chinese',()=>{
  const h=createCustomerProfileModuleFormalRestoreConsoleHarness();h.task.state='module_restore_conflicts_found';h.task.selected_module_restore_preview.customer_profile_restore_eligibility={eligible:false,blockers:['customer_email_or_username_already_has_another_identity','missing_points_or_referral_facts_require_entirely_missing_account_cohort','missing_account_balance_requires_current_order_reconciliation','unfinished_referral_reward_cannot_be_restored']};
  h.run('renderSelectedBusinessModuleIsolatedRestorePreview(syntheticTask)');const text=h.text('selected-module-restore-preview');
  for(const phrase of ['邮箱或用户名已属于另一个账号','不能回填当前账号的旧余额','旧积分余额须先与当前订单核对','尚未完成的推荐奖励不能恢复执行'])assert(text.includes(phrase));
  assert(!text.includes('missing_account_'));
});
