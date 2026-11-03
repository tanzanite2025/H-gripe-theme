'use strict';
// customer service conversation recovery plan and conflict review. Loaded in index.html order; session credentials stay in memory.
let customerServiceConflictReviewContext = null;
function showCustomerServiceConversationRowInspection(inspection) {
  const panel=getMaintenanceConsoleElement('conversation-row-inspection');panel.hidden=!inspection;
  showCustomerServiceConversationRecoveryPlanDraft(inspection?.conversation_recovery_plan);
  if(!inspection)return;
  const files=inspection.files;
  getMaintenanceConsoleElement('conversation-row-description').textContent=`${inspection.complete?'已完成本轮范围检查':'检查不完整，请先处理限制或缺失信息'}。包内引用行 ${inspection.source_selected_rows} 条，目标引用行 ${inspection.target_selected_rows} 条，阻断问题 ${inspection.blocking_issues} 项。每侧最多检查 ${inspection.row_limit_per_side} 条引用行、${formatBackupByteSize(inspection.projection_byte_limit_per_side)} 引用字段；正文和访客身份值不会写入报告。`;
  getMaintenanceConsoleElement('conversation-file-description').textContent=`包内附件已校验 ${files.package_verified_files}/${files.package_referenced_files} 个；目标同内容 ${files.target_matching_files} 个、缺失 ${files.target_missing_files} 个、文件内容冲突 ${files.target_different_files} 个。目标检查同时覆盖包内需要的路径。`;
  const referenceChecks=inspection.conversation_reference_checks;
  if(referenceChecks)getMaintenanceConsoleElement('conversation-row-description').textContent+=` 自然键匹配核对 ${referenceChecks.natural_key_comparisons} 项，业务快照引用检查 ${referenceChecks.snapshot_reference_checks} 项；历史卡片保留原文，订单、商品与配置保持当前共享数据。`;
  if(referenceChecks?.column_contracts?.length) {
    const observed=referenceChecks.column_contracts.reduce((total,coverage)=>total+coverage.observed_columns,0);
    const unregistered=referenceChecks.column_contracts.reduce((total,coverage)=>total+coverage.unregistered_columns,0);
    getMaintenanceConsoleElement('conversation-row-description').textContent+=` 候选字段引用契约已检查 ${observed} 项，未登记 ${unregistered} 项；未登记字段将阻断候选构造。`;
  }
  if(referenceChecks?.retained_media_owner_column_contracts?.length) {
    const observed=referenceChecks.retained_media_owner_column_contracts.reduce((total,coverage)=>total+coverage.observed_columns,0);
    const unregistered=referenceChecks.retained_media_owner_column_contracts.reduce((total,coverage)=>total+coverage.unregistered_columns,0);
    getMaintenanceConsoleElement('conversation-row-description').textContent+=` 保留头像档案和自动回复字段已检查 ${observed} 项，未登记 ${unregistered} 项；未知引用须先处理。`;
  }
  showCustomerServiceDurableEventInspection(inspection.durable_events);
  const rows=getMaintenanceConsoleElement('conversation-row-tables');rows.replaceChildren();
  for(const table of inspection.tables){const row=document.createElement('tr');for(const value of [table.table,`${table.source_scoped}/${table.source_selected}`,`${table.target_scoped}/${table.target_selected}`,table.matching,table.target_missing,table.content_differs,table.identity_conflicts,table.target_only_scoped+(table.target_only_shared||0)]){const cell=document.createElement('td');cell.textContent=String(value);row.append(cell);}rows.append(row);}
  const counts=getMaintenanceConsoleElement('conversation-row-issue-counts');counts.replaceChildren();
  for(const [code,count] of Object.entries(inspection.issue_counts)){const item=document.createElement('li');item.textContent=`${code}：${count} 项`;counts.append(item);}
  const issues=getMaintenanceConsoleElement('conversation-row-issues');issues.replaceChildren();
  for(const issue of inspection.issues){const item=document.createElement('li');item.textContent=`${issue.severity==='blocking'?'需处理':'需核对'} · ${issue.side==='package'?'包内':issue.side==='target'?'目标':'双方'} · ${issue.table||'范围检查'}${issue.row_identity?` #${issue.row_identity}`:''}：${issue.message}`;issues.append(item);}
  getMaintenanceConsoleElement('conversation-row-sample-description').textContent=inspection.issues_truncated?`问题样例仅展示前 ${inspection.issue_sample_limit} 项；上方分类计数包含本轮已发现的全部问题。`:'已展示本轮发现的问题；这份报告不能用于执行单模块覆盖。';
}
function showCustomerServiceConversationRecoveryPlanDraft(plan) {
  showCustomerServiceConflictDecisionReview(plan);
  const panel=getMaintenanceConsoleElement('customer-service-recovery-plan-draft');panel.hidden=!plan;
  if(!plan)return;
  getMaintenanceConsoleElement('customer-service-recovery-plan-description').textContent=`客服恢复计划草案 v${plan.version} · ${plan.inspection_complete?'本轮引用检查完成':'引用检查不完整'} · 仅供审查，不能执行覆盖。工作副本到期：${plan.binding.work_copy_expires_at}。`;
  getMaintenanceConsoleElement('customer-service-recovery-plan-digest').textContent=`计划 SHA-256：${plan.sha256}`;
  const labels={preserve_target_only:'保留目标新增',preserve_matching_target:'保留相同目标',propose_insert_after_review:'建议插入，待审查',propose_replace_after_review:'建议替换，待审查',propose_add_source_file_after_review:'建议补入文件，待审查',blocked_comparison_incomplete:'比较不完整',blocked_ownership_conflict:'归属冲突',blocked_reference_contract_conflict:'自然键或业务引用冲突',blocked_shared_dependency_missing:'共享依赖缺失',blocked_shared_dependency_differs:'共享依赖不同',blocked_file_observation_incomplete:'文件检查不完整',blocked_same_path_content_conflict:'同路径内容冲突'};
  const actions=getMaintenanceConsoleElement('customer-service-recovery-plan-actions');actions.replaceChildren();
  for(const [kind,counts] of [['行',plan.row_action_counts],['文件',plan.file_action_counts]])for(const [action,count] of Object.entries(counts||{})){const item=document.createElement('li');item.textContent=`${kind} · ${labels[action]||action}：${count} 项`;actions.append(item);}
  getMaintenanceConsoleElement('customer-service-recovery-plan-preservation').textContent=`记录 ${plan.target_preservation.length} 张目标表的保留部分指纹；目标新增记录、所有当前事件与凭据、共享配置继续保留。候选媒体须从当前完整媒体开始构造；同路径不同内容须先处理。`;
  const blockers=getMaintenanceConsoleElement('customer-service-recovery-plan-blockers');blockers.replaceChildren();
  for(const [code,count] of Object.entries(plan.blocking_issue_counts||{})){const item=document.createElement('li');item.textContent=`${code}：${count} 项阻断`;blockers.append(item);}
  getMaintenanceConsoleElement('customer-service-recovery-plan-limitations').textContent=`尚待完成 ${plan.unimplemented_safeguards.length} 类前置条件，包括未登记引用、自定义写入副作用契约、选择后的依赖验证、停写后复核、候选库与媒体构造、保留数据验证及成对切换回退。计划不代表已完成这些步骤。`;
}
function customerServiceConflictDecisionRowKey(row) { return `${row.table}:${row.identity_sha256}`; }
function showCustomerServiceConflictDecisionReview(plan) {
  const panel=getMaintenanceConsoleElement('customer-service-conflict-review');
  customerServiceConflictReviewContext=null;
  panel.hidden=!plan?.conflict_decision_policy;
  if(panel.hidden)return;
  const rows=plan.row_actions.filter(row=>['propose_insert_after_review','propose_replace_after_review'].includes(row.action));
  customerServiceConflictReviewContext={plan,rows,page:0,loaded:false,saving:false,inspectingSelection:false,review:null,edits:new Map()};
  clearCustomerServiceEffectiveSelectionInspection('');
  getMaintenanceConsoleElement('customer-service-conflict-review-confirmation').value='';
  getMaintenanceConsoleElement('customer-service-conflict-review-confirmation-phrase').textContent=`REVIEW ${plan.binding.site_id}/${plan.binding.environment}/${plan.binding.package_id}`;
  getMaintenanceConsoleElement('customer-service-conflict-review-result').textContent='';
  const counts=getMaintenanceConsoleElement('customer-service-conflict-review-repair-counts');counts.replaceChildren();
  for(const [code,count] of Object.entries(plan.conflict_decision_policy.repair_required_issue_counts)){const item=document.createElement('li');item.textContent=`${code}：${count} 项，不能通过内容选择豁免`;counts.append(item);}
  renderCustomerServiceConflictDecisionPage();
}
async function loadCustomerServiceConflictReviewForDisplayedPlan() {
  const context=customerServiceConflictReviewContext;if(!context)return;
  try{
    const result=await maintenanceRequest(`/api/tasks/${context.plan.binding.task_id}/customer-service-conflict-review`);
    if(customerServiceConflictReviewContext!==context)return;
    context.review=result.review;context.loaded=true;renderCustomerServiceConflictDecisionPage();
  }catch(error){
    if(customerServiceConflictReviewContext===context)getMaintenanceConsoleElement('customer-service-conflict-review-result').textContent=`已保存记录读取失败：${error.message}。请重新检查并载入。`;
  }
}
function renderCustomerServiceConflictDecisionPage() {
  const context=customerServiceConflictReviewContext;if(!context)return;
  const {plan,review,edits}=context;
  const samePlan=review?.plan_sha256===plan.sha256;
  const recorded=new Map((samePlan?review.decisions:[]).map(row=>[customerServiceConflictDecisionRowKey(row),row]));
  const container=getMaintenanceConsoleElement('customer-service-conflict-review-rows');container.replaceChildren();
  const pageRows=context.rows.slice(context.page*10,context.page*10+10);
  for(const row of pageRows){
    const key=customerServiceConflictDecisionRowKey(row);
    const value=edits.get(key)||recorded.get(key)||{choice:'undecided',reason:''};
    const article=document.createElement('article');
    const title=document.createElement('strong');title.textContent=`${row.table} #${row.row_identity||row.identity_sha256.slice(0,16)} · ${row.action==='propose_insert_after_review'?'备份独有行':'同归属内容不同'}`;article.append(title);
    const label=document.createElement('label');label.textContent='处置选择';const select=document.createElement('select');select.setAttribute('aria-label',`处置选择 ${row.table} #${row.row_identity||row.identity_sha256}`);
    for(const [choice,text] of [['undecided','待决定'],['preserve_current_row','保留当前内容'],['use_package_row',row.action==='propose_insert_after_review'?'记录补入备份行的选择':'记录采用备份内容的选择']]){if(choice==='preserve_current_row'&&row.action==='propose_insert_after_review')continue;const option=document.createElement('option');option.value=choice;option.textContent=text;select.append(option);}
    select.value=value.choice;label.append(select);article.append(label);
    const reasonLabel=document.createElement('label');reasonLabel.textContent='处置原因或核对凭证';const input=document.createElement('input');input.value=value.reason;input.maxLength=512;input.setAttribute('aria-label',`处置原因 ${row.table} #${row.row_identity||row.identity_sha256}`);reasonLabel.append(input);article.append(reasonLabel);
    const update=()=>{edits.set(key,{table:row.table,identity_sha256:row.identity_sha256,choice:select.value,reason:input.value});clearCustomerServiceEffectiveSelectionInspection('存在未保存的修改，请保存后重新核对关联。');updateCustomerServiceConflictReviewSaveButton();};
    select.addEventListener('change',update);input.addEventListener('input',update);
    select.disabled=input.disabled=context.saving||context.inspectingSelection;container.append(article);
  }
  const saved=samePlan?review.decisions.length:0;
  const fixed=Object.values(plan.conflict_decision_policy.fixed_preservation_issue_counts).reduce((total,count)=>total+count,0);
  getMaintenanceConsoleElement('customer-service-conflict-review-summary').textContent=`需逐行决定 ${context.rows.length} 项，当前计划已保存 ${saved} 项；目标新增记录固定保留 ${fixed} 项。${!context.loaded?'正在读取已保存记录。':review&&!samePlan?'原记录属于旧计划，本计划须重新选择。':samePlan?`保存时间 ${review.recorded_at}；这是保存时的核对结果。`:'尚无处置记录。'} ${plan.inspection_complete?'':'依赖检查不完整，保存选择不会取消阻断。'}`;
  const totalPages=Math.max(1,Math.ceil(context.rows.length/10));
  getMaintenanceConsoleElement('customer-service-conflict-review-page').textContent=`第 ${context.page+1} / ${totalPages} 页；翻页保留未保存的选择，每页单独保存。`;
  getMaintenanceConsoleElement('customer-service-conflict-review-previous').disabled=context.saving||context.page===0;
  getMaintenanceConsoleElement('customer-service-conflict-review-next').disabled=context.saving||context.page+1>=totalPages;
  updateCustomerServiceConflictReviewSaveButton();
}
function updateCustomerServiceConflictReviewSaveButton() {
  const context=customerServiceConflictReviewContext;
  getMaintenanceConsoleElement('customer-service-conflict-review-save').disabled=!context||!context.loaded||context.saving||context.inspectingSelection||!context.plan.conflict_decision_policy.mapping_complete||!context.rows.slice(context.page*10,context.page*10+10).some(row=>context.edits.has(customerServiceConflictDecisionRowKey(row)));
  const decisionsComplete=context&&(context.rows.length===0||context.review?.plan_sha256===context.plan.sha256&&context.review.content_decisions_complete);
  getMaintenanceConsoleElement('customer-service-effective-selection-inspect').disabled=!context||!context.loaded||context.saving||context.inspectingSelection||!context.plan.conflict_decision_policy.mapping_complete||!decisionsComplete||context.edits.size>0;
}
function clearCustomerServiceEffectiveSelectionInspection(message) {
  getMaintenanceConsoleElement('customer-service-effective-selection-result').hidden=true;
  getMaintenanceConsoleElement('customer-service-effective-selection-status').textContent=message;
}
async function inspectDisplayedCustomerServiceEffectiveSelectionDependencies() {
  const context=customerServiceConflictReviewContext;if(!context||getMaintenanceConsoleElement('customer-service-effective-selection-inspect').disabled)return;
  context.inspectingSelection=true;clearCustomerServiceEffectiveSelectionInspection('正在重新读取恢复点与当前数据，核对已保存选择的关联。');renderCustomerServiceConflictDecisionPage();
  const currentReviewSHA256=context.review?.plan_sha256===context.plan.sha256?context.review.sha256:'';
  try {
    const {report}=await postMaintenanceRequest(`/api/tasks/${context.plan.binding.task_id}/customer-service-effective-selection-inspection`,{plan_sha256:context.plan.sha256,review_sha256:currentReviewSHA256});
    if(customerServiceConflictReviewContext!==context)return;
    if(report.plan_sha256!==context.plan.sha256||report.review_sha256!==currentReviewSHA256)throw Error('检查结果与当前显示的选择不符，请重新检查并载入。');
    const counts=report.effective_row_action_counts;
    const additions=report.required_files.filter(file=>file.action==='selected_add_package_file').length;
    getMaintenanceConsoleElement('customer-service-effective-selection-summary').textContent=`${report.selection_dependencies_satisfied?'本次选择的依赖检查通过':'本次选择仍有依赖问题或原始阻断'}。拟补入 ${counts.selected_package_insert||0} 行、采用备份内容 ${counts.selected_package_replacement||0} 行、明确保留当前 ${counts.selected_preserve_current||0} 行；需要文件 ${report.required_files.length} 个，其中拟从备份补入 ${additions} 个。已重算 ${report.target_preservation.length} 张表的保留范围。检查时间 ${report.inspected_at}。尚未构造候选库，也未执行数据覆盖。`;
    const issues=getMaintenanceConsoleElement('customer-service-effective-selection-issues');issues.replaceChildren();
    for(const [label,values] of [['组合依赖',report.dependency_checks.blocking_issue_counts||{}],['原始问题仍需修复',report.repair_required_issue_counts||{}]])for(const [code,count] of Object.entries(values)){const item=document.createElement('li');item.textContent=`${label}：${code} · ${count} 项`;issues.append(item);}
    if(report.dependency_blockers?.length){
      const group=document.createElement('li'),details=document.createElement('details'),summary=document.createElement('summary'),list=document.createElement('ul');
      summary.textContent=`仍需处理的模块级问题：${report.dependency_blockers.length} 项`;details.append(summary);
      for(const blocker of report.dependency_blockers){const item=document.createElement('li');item.textContent=`${blocker.table?`${blocker.table}：`:''}${blocker.message}`;list.append(item);}
      details.append(list);group.append(details);issues.append(group);
    }
    const preserved=getMaintenanceConsoleElement('customer-service-effective-selection-preservation');preserved.replaceChildren();
    for(const table of report.target_preservation){const item=document.createElement('li');item.textContent=`${table.table}：保留 ${table.retained_records} 行，拟替换 ${table.proposed_replacement_records} 行`;preserved.append(item);}
    getMaintenanceConsoleElement('customer-service-effective-selection-result').hidden=false;
    getMaintenanceConsoleElement('customer-service-effective-selection-status').textContent='结果绑定本次已保存选择；内容修改或新写入后须重新检查。单模块恢复仍未开放。';
  } catch(error) {
    if(customerServiceConflictReviewContext===context){context.loaded=false;clearCustomerServiceEffectiveSelectionInspection(`关联检查被拒绝：${error.message}。请重新检查并载入。`);}throw error;
  } finally {context.inspectingSelection=false;if(customerServiceConflictReviewContext===context)renderCustomerServiceConflictDecisionPage();}
}
async function saveDisplayedCustomerServiceConflictDecisions() {
  const context=customerServiceConflictReviewContext;if(!context||!context.loaded||context.saving||context.inspectingSelection)return;
  const decisions=context.rows.slice(context.page*10,context.page*10+10).map(row=>context.edits.get(customerServiceConflictDecisionRowKey(row))).filter(Boolean);
  if(!decisions.length)return;
  for(const decision of decisions)if(!decision.reason.trim()||new TextEncoder().encode(decision.reason.trim()).length>512)throw Error('请为每项选择填写处置原因，最多 512 个 UTF-8 字节。');
  context.saving=true;clearCustomerServiceEffectiveSelectionInspection('选择正在保存，保存后须重新核对关联。');renderCustomerServiceConflictDecisionPage();
  try{
    const {review}=await postMaintenanceRequest(`/api/tasks/${context.plan.binding.task_id}/customer-service-conflict-review`,{plan_sha256:context.plan.sha256,expected_review_sha256:context.review?.sha256||'',confirmation:getMaintenanceConsoleElement('customer-service-conflict-review-confirmation').value,decisions});
    if(customerServiceConflictReviewContext!==context)return;
    context.review=review;for(const decision of decisions)context.edits.delete(customerServiceConflictDecisionRowKey(decision));
    getMaintenanceConsoleElement('customer-service-conflict-review-result').textContent=`已保存本页 ${decisions.length} 项，仍有 ${review.pending_row_decisions} 项待决定。处置记录不能执行覆盖，必须先修复的问题继续保留。`;
  }catch(error){if(customerServiceConflictReviewContext===context){context.loaded=false;getMaintenanceConsoleElement('customer-service-conflict-review-result').textContent=`保存被拒绝：${error.message}。请重新检查并载入后核对。`;}throw error;}
  finally{context.saving=false;if(customerServiceConflictReviewContext===context)renderCustomerServiceConflictDecisionPage();}
}
function showCustomerServiceDurableEventInspection(inspection) { showModuleRecoveryDurableEventInspection('customer-service-durable-event', inspection); }
