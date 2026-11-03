'use strict';
const { loadMaintenanceConsoleScriptsInDocumentOrder } = require('./load-maintenance-console-classic-scripts-in-html-document-order.cjs');
const assert=require('node:assert/strict');
const vm=require('node:vm');
const {test}=require('node:test');

class SyntheticOrderModuleRestoreConsoleElement {
  constructor(){this.children=[];this.dataset={};this.listeners={};this.value='';this.textContent='';this.hidden=false;}
  append(...children){this.children.push(...children);}
  replaceChildren(...children){this.children=children;}
  addEventListener(type,listener){this.listeners[type]=listener;}
  scrollIntoView(){}
}
function createOrderModuleHistoryRestoreConsoleHarness(){
  const elements=new Map();const findElement=id=>{if(!elements.has(id))elements.set(id,new SyntheticOrderModuleRestoreConsoleElement());return elements.get(id);};
  const browserWindow={addEventListener(){}};browserWindow.parent=browserWindow;
  const context=vm.createContext({document:{getElementById:findElement,createElement:()=>new SyntheticOrderModuleRestoreConsoleElement(),querySelectorAll:()=>[],querySelector:()=>null},location:{pathname:'/local-backup-maintenance/',origin:'https://synthetic.invalid'},window:browserWindow,Headers,console,clearInterval(){},setInterval(){return 1;},fetch:async()=>{throw new Error('unexpected network request');}});
  loadMaintenanceConsoleScriptsInDocumentOrder(context);
  const task={id:'synthetic-order-task',kind:'upload',state:'module_isolated_restore_verified',size:100,uploaded_bytes:100,expires_at:new Date(Date.now()+3600000).toISOString(),manifest:{scope:'selected_business_modules',selected_modules:['orders']},selected_module_restore_preview:{sha256:'a'.repeat(64),binding:{},target_unchanged:true,disposable_stage_removed:true,isolated_restore_verified:true,order_history_restore_eligibility:{eligible:true,policy:'restore_terminal_and_settled_active_orders_with_normal_constraints_and_original_inventory_dependencies_v6'}}};
  const status={site_id:'synthetic-store',environment:'disposable',permissions:['ops:backup:restore','ops:backup:upload'],application_fenced:false};
  Object.assign(context,{syntheticTask:task,syntheticStatus:status});
  vm.runInContext('maintenanceCredential="synthetic-token";targetStatus=syntheticStatus;displayedMaintenanceTasks=[syntheticTask]',context);
  const collect=element=>[element.textContent,...element.children.map(collect)].join('\n');
  return{context,task,status,run:code=>vm.runInContext(code,context),text:id=>collect(findElement(id)),findElement};
}
for(const scenario of ['eligible','view_only','fenced','expired','invalid_expiry','work_expired','mixed_scope','missing_digest','ineligible','running','already_started','obsolete_detail_only_policy','obsolete_completed_facts_policy','obsolete_global_payment_policy','obsolete_settled_active_policy','obsolete_replica_writer_policy']){
  test(`order history restore button follows actual eligibility: ${scenario}`,()=>{
    const h=createOrderModuleHistoryRestoreConsoleHarness();
    if(scenario==='view_only')h.status.permissions=['ops:backup:view'];
    if(scenario==='fenced')h.status.application_fenced=true;
    if(scenario==='expired')h.task.expires_at=new Date(0).toISOString();
    if(scenario==='invalid_expiry')h.task.expires_at='invalid';
    if(scenario==='work_expired')h.task.work_copy_expired=true;
    if(scenario==='mixed_scope')h.task.manifest.selected_modules=['customer_service','orders'];
    if(scenario==='missing_digest')delete h.task.selected_module_restore_preview.sha256;
    if(scenario==='ineligible')h.task.selected_module_restore_preview.order_history_restore_eligibility.eligible=false;
    if(scenario==='obsolete_detail_only_policy')h.task.selected_module_restore_preview.order_history_restore_eligibility.policy='restore_missing_terminal_order_history_preserve_current_facts';
    if(scenario==='obsolete_completed_facts_policy')h.task.selected_module_restore_preview.order_history_restore_eligibility.policy='restore_terminal_order_history_and_completed_facts_preserve_current_notes_v2';
    if(scenario==='obsolete_global_payment_policy')h.task.selected_module_restore_preview.order_history_restore_eligibility.policy='restore_terminal_order_history_completed_facts_and_global_payment_receipts_preserve_current_notes_v3';
    if(scenario==='obsolete_settled_active_policy')h.task.selected_module_restore_preview.order_history_restore_eligibility.policy='restore_terminal_and_settled_active_orders_with_completed_facts_and_original_payment_receipts_v4';
    if(scenario==='obsolete_replica_writer_policy')h.task.selected_module_restore_preview.order_history_restore_eligibility.policy='restore_terminal_and_settled_active_orders_with_original_component_and_master_inventory_dependencies_v5';
    if(scenario==='running')h.task.state='running';
    if(scenario==='already_started')h.task.order_module_restore={phase:'installing_files'};
    h.run('renderBackupTasks([syntheticTask])');
    assert.equal(h.text('tasks').includes('恢复缺失订单记录（暂停商城）'),scenario==='eligible');
  });
}
for(const scenario of ['interrupted','expired','view_only','running','released']){
  test(`interrupted order recovery retains its dedicated entry: ${scenario}`,()=>{
    const h=createOrderModuleHistoryRestoreConsoleHarness();h.task.state='interrupted';h.status.application_fenced=true;h.task.order_module_restore={phase:'committing_rows'};
    if(scenario==='expired'){h.task.expires_at=new Date(0).toISOString();h.task.work_copy_expired=true;}
    if(scenario==='view_only')h.status.permissions=['ops:backup:view'];
    if(scenario==='running')h.task.state='running';
    if(scenario==='released'){h.task.state='completed';h.task.order_module_restore={phase:'released',outcome:'restored',restored_rows:6};}
    h.run('renderBackupTasks([syntheticTask])');const text=h.text('tasks');
    assert.equal(text.includes('核验并恢复订单任务'),['interrupted','expired'].includes(scenario));
    for(const ordinary of ['解除未改数据的维护隔离','校验包并检查可用恢复方式','验证模块恢复（暂停商城）','恢复缺失订单记录（暂停商城）'])assert(!text.includes(ordinary));
    if(scenario==='released')assert(text.includes('共补回 6 条记录'));
  });
}
for(const scenario of ['submit','cancel','changed_preview','changed_task','changed_status','changed_session','disconnect_after_submit','interrupted_expired']){
  test(`order restore submission binds current preview and session: ${scenario}`,async()=>{
    const h=createOrderModuleHistoryRestoreConsoleHarness();const recover=scenario==='interrupted_expired';
    if(recover){h.task.state='interrupted';h.task.order_module_restore={phase:'committing_rows'};h.task.expires_at=new Date(0).toISOString();h.status.application_fenced=true;}
    let posted;
    h.context.prompt=message=>{
      assert(message.includes(recover?'核验本次订单恢复':'将暂停商城'));
      if(scenario==='cancel')return null;
      if(scenario==='changed_preview')h.task.selected_module_restore_preview.sha256='b'.repeat(64);
      if(scenario==='changed_task')h.task.state='failed';
      if(scenario==='changed_status')h.status.application_fenced=true;
      if(scenario==='changed_session')h.findElement('disconnect').listeners.click();
      return `${recover?'RECOVER ORDER RESTORE':'RESTORE ORDER HISTORY'} synthetic-store/disposable/${h.task.id}`;
    };
    h.context.postMaintenanceRequest=async(url,body)=>{posted={url,body};if(scenario==='disconnect_after_submit')h.findElement('disconnect').listeners.click();return {};};
    const operation=h.run(`startOrderModuleHistoryRestoreOrInterruptedRecovery(syntheticTask.id,${recover})`);
    if(scenario.startsWith('changed_')){await assert.rejects(operation);assert.equal(posted,undefined);}
    else{await operation;if(scenario==='cancel')assert.equal(posted,undefined);else{assert(posted.url.endsWith(recover?'order-module-interrupted-restore-recovery':'order-module-formal-history-restore'));if(!recover)assert.equal(posted.body.expected_preview_sha256,'a'.repeat(64));}}
    if(scenario==='disconnect_after_submit')assert(!h.text('message').includes('正在补回历史订单'));
  });
}
test('order preview shows completed-history scope and the number of retained newer notes',()=>{
  const h=createOrderModuleHistoryRestoreConsoleHarness();
  h.task.selected_module_restore_preview.order_history_restore_eligibility.settled_active_order_owners=3;
  Object.assign(h.task.selected_module_restore_preview,{tables:[{table:'orders',identical_rows:0,preserved_current_order_note_rows:2,missing_rows:1,conflicting_rows:0,protected_new_rows:0}],restored_rows:9,verified_files:1,preserved_tables:228,dependency_conflicts:0,file_conflicts:0,limitations:[]});
  h.run('renderSelectedBusinessModuleIsolatedRestorePreview(syntheticTask)');
  const text=h.text('selected-module-restore-preview');
  for(const phrase of ['具备完成凭据','已处理事件','已发送通知','已完成支付请求凭据','保留较新订单备注','恢复 9 条缺失记录','已付款、处理中、已发货','符合恢复条件的未结束订单 3 个'])assert(text.includes(phrase));
  assert(text.includes('\n2\n'));
});
test('settled active order blockers explain payment shipping and inventory limits',()=>{
  const h=createOrderModuleHistoryRestoreConsoleHarness();
  h.task.state='module_restore_conflicts_found';
  Object.assign(h.task.selected_module_restore_preview,{tables:[],limitations:[],order_history_restore_eligibility:{eligible:false,blockers:['active_order_requires_one_matching_settled_transaction','active_order_requires_matching_shipping_facts','active_order_component_inventory_requires_separate_dependency_contract','active_order_refund_dispute_or_review_workflow_requires_reconciliation']}});
  h.run('renderSelectedBusinessModuleIsolatedRestorePreview(syntheticTask)');
  const text=h.text('selected-module-restore-preview');
  for(const phrase of ['唯一且金额、币种、渠道一致的已完成收款凭据','发货状态、时间或原发货证据不对应','配置组件库存分配尚需单独核对','退款、售后、争议或支付审核'])assert(text.includes(phrase));
  assert(!text.includes('active_order_'));
  h.run('renderBackupTasks([syntheticTask])');
  assert(!h.text('tasks').includes('恢复缺失订单记录（暂停商城）'));
});
test('order preview explains missing receipts and domain conflicts in readable Chinese',()=>{
  const h=createOrderModuleHistoryRestoreConsoleHarness();
  h.task.state='module_restore_conflicts_found';
  Object.assign(h.task.selected_module_restore_preview,{tables:[],dependency_conflicts:0,file_conflicts:0,limitations:[],order_history_restore_eligibility:{eligible:false,blockers:['transaction_requires_recorded_completion_and_provider_identity:transactions','restored_order_domain_or_receipt_validation_failed','global_payment_receipt_requires_completed_state_without_execution_lease:payment_operation_idempotencies','global_payment_receipt_recorded_response_invalid:payment_operation_idempotencies']}});
  h.run('renderSelectedBusinessModuleIsolatedRestorePreview(syntheticTask)');
  const text=h.text('selected-module-restore-preview');
  assert(text.includes('交易缺少完成时间或支付平台交易编号（交易）'));
  assert(text.includes('副本恢复后的订单关联、事件载荷或通知凭据核对未通过'));
  assert(text.includes('支付请求尚未完成或仍有执行租约（全局支付请求凭据）'));
  assert(text.includes('支付请求缺少可安全返回的原始 HTTP 结果（全局支付请求凭据）'));
  assert(!text.includes('transaction_requires_'));
});

test('inventory dependency and pending reservation blockers remain explicit',()=>{
 const h=createOrderModuleHistoryRestoreConsoleHarness();
 h.task.state='module_restore_conflicts_found';
 Object.assign(h.task.selected_module_restore_preview,{tables:[],limitations:[],order_history_restore_eligibility:{eligible:false,blockers:['pending_order_inventory_reservation_requires_durable_release_reconciliation','active_order_component_inventory_configuration_invalid','active_order_component_inventory_original_dependency_missing']}});
 h.run('renderSelectedBusinessModuleIsolatedRestorePreview(syntheticTask)');
 const text=h.text('selected-module-restore-preview');
 for(const phrase of ['库存预留与返还状态','库存分配结构、SKU 或数量无效','原包缺少配置组件 SKU 的依赖摘要'])assert(text.includes(phrase));
 assert(!text.includes('pending_order_'));
 assert(!text.includes('active_order_'));
});
