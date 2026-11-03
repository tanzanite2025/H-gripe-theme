'use strict';
// selected module isolated restore preview and dependency reports. Loaded in index.html order; session credentials stay in memory.
let pendingVerificationTaskID = '';
async function startSelectedBusinessModuleIsolatedRestorePreview(task) {
  const phrase=`VERIFY MODULE RESTORE ${targetStatus.site_id}/${targetStatus.environment}/${task.id}`;
  const confirmation=prompt(`将暂停商城服务，在一次性副本验证 ${describeSelectedBusinessBackupModules(task.manifest.selected_modules)} 的恢复，完成后恢复当前商城服务。当前商城数据不会被覆盖。请输入：${phrase}`);
  if(confirmation===null)return;
  await postMaintenanceRequest(`/api/tasks/${task.id}/selected-module-isolated-restore-preview`,{confirmation});
  pendingVerificationTaskID=task.id;
  getMaintenanceConsoleElement('preview-panel').hidden=true;
  showMessage('正在验证模块隔离恢复。结束后请查看范围与验证结果。');
}
function renderSelectedBusinessModuleIsolatedRestorePreview(task) {
  const panel=getMaintenanceConsoleElement('selected-module-restore-preview');
  const report=task.selected_module_restore_preview;
  panel.dataset.taskId=task.id;
  panel.replaceChildren();panel.hidden=!report;
  if(!report)return;
  if(report.media_restore_eligibility){const eligibility=report.media_restore_eligibility;const note=document.createElement('p');note.textContent=eligibility.eligible?`图片与附件规则通过：可补缺，并修复 ${eligibility.existing_file_repairs||0} 个登记一致的文件；原字节保留至提交核验，商品、订单、聊天和配置保留。`:'图片与附件恢复存在阻断：';panel.append(note);for(const blocker of eligibility.blockers||[]){const reason=document.createElement('p');reason.className='error';reason.textContent=blocker;panel.append(reason);}}
  if(report.configuration_site_content_restore_eligibility){const eligibility=report.configuration_site_content_restore_eligibility;const note=document.createElement("p");note.textContent=eligibility.eligible?"内容、目录、问卷及市场/汇率/税率、邮件模板/历史版本和会员等级、承运商/配送区域、包装规格/商品SKU绑定及报关分类补缺规则通过：支持文章分类、文章、图库、FAQ、视觉展示、跳转及车架/前叉/花鼓、辐条计算和轮胎适配目录、选购流程/问卷及市场/汇率/税率、邮件模板/历史版本和会员等级、承运商/配送区域、包装规格/商品SKU绑定及报关分类；相关图片须已恢复。写入后保持维护，另行核对内容、目录/计算结果与缓存再恢复服务。":"内容、目录、问卷及市场/汇率/税率、邮件模板/历史版本和会员等级、承运商/配送区域、包装规格/商品SKU绑定及报关分类补缺存在阻断：";panel.append(note);for(const blocker of eligibility.blockers||[]){const reason=document.createElement("p");reason.className="error";reason.textContent=blocker;panel.append(reason);}}
  const coverage=report.customer_service_overwrite_preview;
  if(report.order_history_restore_eligibility){const note=document.createElement('p');note.textContent=report.order_history_restore_eligibility.eligible?`订单历史恢复资格通过：可恢复缺失终态订单、已付款且执行事实完整的未结束订单（已付款、处理中、已发货）、证据附件及具备完成凭据的交易、退款、发货证据、已处理事件、已发送通知和已完成支付请求凭据；保留当前记录及较新备注，不重新执行业务操作。符合恢复条件的未结束订单 ${report.order_history_restore_eligibility.settled_active_order_owners||0} 个。配置组件与主库存 SKU 已按原包依赖核对，当前库存保留。`:`订单历史恢复被阻断：${report.order_history_restore_eligibility.blockers.map(describeOrderHistoryRestoreEligibilityBlocker).join('；')}。待付款、退款/售后/争议/审核中的订单、组件或主库存依赖缺失/变化、凭据不完整的执行记录及金额/状态/地址差异仍需核对。`;panel.append(note);}
  const summary=document.createElement('p');
  const passed=['module_isolated_restore_verified','module_chat_candidate_rehearsal_verified'].includes(task.state)&&report.isolated_restore_verified&&report.target_unchanged&&report.disposable_stage_removed;
  const contentConflicts=report.tables.reduce((count,item)=>count+item.conflicting_rows,0);
  const selectedCandidate=task.selected_customer_service_candidate_rehearsal;
  const chosenCandidateVerified=task.state==='module_chat_candidate_rehearsal_verified'&&selectedCandidate?.isolated_restore_verified&&selectedCandidate.target_unchanged&&selectedCandidate.disposable_stage_removed&&selectedCandidate.saved_choice_sha256===task.selected_customer_service_restore_choice?.sha256&&selectedCandidate.online_overwrite_available===false;
  summary.textContent=chosenCandidateVerified?`原始模块预览发现 ${contentConflicts} 条内容冲突；已按保存的策略完成副本验证，详细结果见下方。当前商城未覆盖。`:passed?`隔离恢复通过：恢复 ${report.restored_rows} 条缺失记录，验证 ${report.verified_files} 个文件，核对 ${report.preserved_tables} 张表的已有记录保留；演练库已删除，当前商城未覆盖。`:`验证未通过或存在冲突：内容冲突 ${contentConflicts} 条，依赖冲突 ${report.dependency_conflicts} 项，文件冲突 ${report.file_conflicts} 项${coverage?.package_rows?`，采用备份记录时聊天规则阻断 ${coverage.package_rows.blocking_issues} 项`:''}。完整结果以本次任务状态为准。`;
  if(!passed&&!chosenCandidateVerified)summary.className='error';panel.append(summary);
  if(report.customer_profile_restore_eligibility){
    const eligibility=report.customer_profile_restore_eligibility;
    const description=document.createElement('p');description.textContent=eligibility.eligible?'客户恢复规则通过：可另行确认补回缺失资料，当前账号、密码、限制、积分和订阅状态保留。':'客户恢复存在阻断，请核对身份、完整积分事实和共享依赖。';panel.append(description);
    const reasons={missing_account_cohort_requires_original_balance_summary:'原包缺少完整账号组的原始积分汇总，不能补造余额',customer_contact_or_visitor_identity_already_exists:'订阅邮箱或访客身份已属于另一条当前记录',standalone_customer_package_required:'请使用单独客户资料包',only_storefront_customer_accounts_can_be_restored:'仅支持商城客户账号，后台账号不在恢复范围',customer_email_or_username_already_has_another_identity:'邮箱或用户名已属于另一个账号',shared_media_or_execution_facts_must_already_exist:'共享媒体和执行凭据须已存在且一致',missing_points_or_referral_facts_require_entirely_missing_account_cohort:'缺失积分或推荐事实需要完整缺失的账号组，不能回填当前账号的旧余额',referral_transition_requires_original_record:'推荐审计缺少原始推荐记录',active_referral_execution_cannot_be_restored:'进行中的推荐流程尚不支持恢复',unfinished_referral_reward_cannot_be_restored:'尚未完成的推荐奖励不能恢复执行',active_legacy_referral_cannot_be_restored:'进行中的历史推荐尚不支持恢复',missing_account_has_current_points_or_referral_facts:'缺失账号仍有当前积分或推荐事实，须先核对完整状态',missing_account_balance_requires_current_order_reconciliation:'缺失账号仍有当前订单，旧积分余额须先与当前订单核对',customer_business_reference_requires_original_storefront_account:'关联资料缺少原始商城客户身份'};
    for(const blocker of eligibility.blockers||[]){const paragraph=document.createElement('p');paragraph.className='error';paragraph.textContent=reasons[blocker.split(':')[0]]||'存在尚未支持的客户恢复条件，请保留任务并核对完整状态。';panel.append(paragraph);}
  }
  const table=document.createElement('table');
  const head=document.createElement('tr');
  for(const value of ['数据表','相同记录','保留当前运行状态','保留较新订单备注','保留当前客户资料','缺失记录','内容冲突','受保护缺失记录','限制']){const cell=document.createElement('th');cell.textContent=value;head.append(cell);}table.append(head);
  for(const item of report.tables){const row=document.createElement('tr');for(const value of [item.table,item.identical_rows,item.preserved_current_runtime_rows||0,item.preserved_current_order_note_rows||0,item.preserved_current_customer_rows||0,item.missing_rows,item.conflicting_rows,item.protected_new_rows,item.unsupported_reason||'']){const cell=document.createElement('td');cell.textContent=String(value);row.append(cell);}table.append(row);}
  const scroll=document.createElement('div');scroll.className='scroll';scroll.append(table);panel.append(scroll);
  const dependencyReasons={visitor_browsing_context_absent_at_backup_and_target:'备份与当前均无访客浏览档案；会话身份仍按会话自身哈希与签名密钥核对',dependency_absent_at_backup_and_target:'备份与当前均缺少必需依赖',target_dependency_missing:'当前缺少备份时存在的依赖',package_dependency_missing:'备份缺少当前存在的依赖，需核对身份',dependency_content_changed:'依赖内容已变化',unexpected_target_reference:'当前发现未登记在包内的引用',expected_reference_not_reproduced:'包内依赖未能按当前引用结构重现'};
  for(const issue of report.dependency_issues||[]){const paragraph=document.createElement('p');paragraph.textContent=`${issue.severity==='blocking'?'依赖阻断':'依赖提示'}：${issue.table}（${issue.referenced_columns.join('、')}）${issue.references} 项；${dependencyReasons[issue.reason]||issue.reason}。`;if(issue.severity==='blocking')paragraph.className='error';panel.append(paragraph);}
  if(coverage){
    const explanation=document.createElement('p');
    explanation.textContent=coverage.prerequisites_satisfied?'聊天覆盖策略预览：分别核对会话版本、阅读位置、附件归属及上传人、保留事件状态。两种策略均保留当前新增会话和消息；保存策略并完成本版完整行演练后，可另行确认正式恢复。':'聊天覆盖策略尚未检查：请先处理共享依赖、文件、受保护执行事实或写入契约冲突。';
    panel.append(explanation);
    for(const strategy of [coverage.package_rows,coverage.preserve_current_rows]){
      if(!strategy)continue;
      const label=strategy.policy==='use_package_conversation_rows'?'采用备份中的聊天记录':'保留当前存在冲突的聊天记录';
      const paragraph=document.createElement('p');
      paragraph.textContent=`${label}：${strategy.domain_checks_satisfied?'聊天规则检查通过':strategy.complete?'存在阻断':'检查未完成'}；采用包内 ${strategy.package_rows_selected} 条，保留冲突 ${strategy.conflicting_rows_preserved} 条，保留当前新增 ${strategy.target_only_rows_preserved} 条；核对事件 ${strategy.preserved_events_checked} 条、所需文件 ${strategy.required_files_checked} 个。正式恢复须另行确认。`;
      if(!strategy.domain_checks_satisfied)paragraph.className='error';panel.append(paragraph);
      if(isSelectedCustomerServiceModuleStrategySaveAvailable(task,strategy.policy))taskActionButton(panel,`保存策略：${label}`,()=>saveSelectedCustomerServiceModuleRestoreChoice(task.id,report.sha256,strategy.policy));
      for(const issue of strategy.issues||[]){const detail=document.createElement('p');detail.textContent=issue.message;detail.className='error';panel.append(detail);}
      if(strategy.issues_truncated){const detail=document.createElement('p');detail.textContent=`问题样例达到上限，共有 ${strategy.blocking_issues} 项阻断；不能将未显示的问题视为通过。`;panel.append(detail);}
    }
  }
  const choice=task.selected_customer_service_restore_choice;
  if(choice){
    const paragraph=document.createElement('p');
    paragraph.textContent=`已保存策略：${choice.policy==='use_package_conversation_rows'?'采用备份中的聊天记录':'保留当前存在冲突的聊天记录'}。已绑定本次包、数据库、附件及配置摘要；任何变化都需重新预览。保存选择未覆盖商城。`;
    panel.append(paragraph);
    if(isSelectedCustomerServiceModuleStrategySaveAvailable(task,choice.policy)&&choice.preview_sha256===report.sha256&&choice.sha256&&choice.execution_available===false)taskActionButton(panel,'验证所选聊天策略（暂停商城）',()=>startSelectedCustomerServiceModuleCandidateRehearsal(task.id,choice.sha256),'danger');
  }
  const candidate=task.selected_customer_service_candidate_rehearsal;
  if(candidate){
    const paragraph=document.createElement('p');
    const verified=task.state==='module_chat_candidate_rehearsal_verified'&&candidate.isolated_restore_verified&&candidate.target_unchanged&&candidate.disposable_stage_removed&&candidate.online_overwrite_available===false&&candidate.saved_choice_sha256===choice?.sha256;
    paragraph.textContent=verified?`所选策略副本验证通过：补入 ${candidate.restored_rows} 条、替换 ${candidate.replaced_rows} 条完整聊天记录，核对 ${candidate.preserved_tables} 张表的保留数据。副本已删除，当前商城未覆盖；${candidate.customer_service_formal_restore_policy==='restore_selected_conversation_rows_preserve_current_shared_facts_v1'?'可在任务列表另行确认正式恢复。':'旧版演练须重新验证后才能正式恢复。'}`:'所选策略副本验证未完成；不能沿用旧结果进行覆盖，请检查任务记录并重新预览。';
    if(!verified)paragraph.className='error';panel.append(paragraph);
  }
  for(const text of report.limitations){const paragraph=document.createElement('p');paragraph.textContent=text;panel.append(paragraph);}
}
async function inspectModuleRecoveryDependencies(taskID, module) {
  getMaintenanceConsoleElement('module-dependency-panel').hidden=true;
  showMessage('正在读取模块与共享依赖，数据较多时需要等待；检查期间商城保持当前运行状态。');
  const {report}=await postMaintenanceRequest(`/api/tasks/${taskID}/module-dependency-report`,{module});
  getMaintenanceConsoleElement('module-dependency-description').textContent=`${report.module_label} · 包 ${report.package_id} · 恢复点 ${report.package_captured_at} · 检查时间 ${report.inspected_at}。${report.target_matches_whole_storefront_preview?'数据、序列和附件摘要与原整体预览一致；字段和约束另见检查结果。':'目标已变化，原整体覆盖预览需要重新验证。'}`;
  const blockers=getMaintenanceConsoleElement('module-dependency-blockers');blockers.replaceChildren();
  for(const blocker of report.blockers){const item=document.createElement('li');item.textContent=`${blocker.table?`${blocker.table}：`:''}${blocker.message}`;blockers.append(item);}
  const roleLabels={module_owned:'模块数据',shared_dependency:'共享依赖',preserved:'保留数据'};
  const statusLabels={not_compared:'未比较内容',source_missing:'包内缺表',target_missing:'目标缺表',schema_conflict:'字段或约束不同',content_differs:'内容不同',table_contents_match:'表内容一致，仍须核对行归属'};
  const rows=getMaintenanceConsoleElement('module-dependency-tables');rows.replaceChildren();
  for(const table of report.tables){const row=document.createElement('tr');for(const value of [table.table,roleLabels[table.role]||table.role,table.target_records??'无此表',table.source_records??'无此表',statusLabels[table.status]||table.status]){const cell=document.createElement('td');cell.textContent=String(value);row.append(cell);}rows.append(row);}
  showCustomerServiceConversationRowInspection(report.conversation_rows);
  showOrderTransactionRowInspection(report.order_rows);
  const reviews=getMaintenanceConsoleElement('module-dependency-reviews');reviews.replaceChildren();
  for(const review of report.required_reviews){const item=document.createElement('li');item.textContent=review;reviews.append(item);}
  getMaintenanceConsoleElement('module-dependency-panel').hidden=false;
  getMaintenanceConsoleElement('module-dependency-panel').scrollIntoView({behavior:'smooth'});
  showMessage('模块依赖报告已生成并保存在服务器任务目录；单模块覆盖仍未开放。');
  if(report.conversation_rows?.conversation_recovery_plan?.conflict_decision_policy)await loadCustomerServiceConflictReviewForDisplayedPlan();
}
