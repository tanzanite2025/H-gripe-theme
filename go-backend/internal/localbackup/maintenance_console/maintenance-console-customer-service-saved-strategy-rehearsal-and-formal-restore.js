'use strict';
// customer service saved strategy rehearsal and formal restore. Loaded in index.html order; session credentials stay in memory.
function isSelectedCustomerServiceModuleStrategySaveAvailable(task,policy) {
  const report=task?.selected_module_restore_preview,coverage=report?.customer_service_overwrite_preview;
  const strategy=policy==='use_package_conversation_rows'?coverage?.package_rows:policy==='preserve_current_conflicting_conversation_rows'?coverage?.preserve_current_rows:null;
  return hasPermission('restore')&&!targetStatus.application_fenced&&task?.kind==='upload'&&task.manifest?.scope==='selected_business_modules'&&JSON.stringify(task.manifest.selected_modules)===JSON.stringify(['customer_service'])&&!task.work_copy_expired&&!task.recovery_journal&&!task.media_module_restore&&!task.order_module_restore&&!task.customer_service_module_restore&&!task.customer_profile_module_restore&&!task.stage_database&&!task.selected_module_disposable_stage&&new Date(task.expires_at).getTime()>Date.now()&&['module_isolated_restore_verified','module_restore_conflicts_found','module_chat_candidate_rehearsal_verified'].includes(task.state)&&report?.sha256&&report.binding&&report.target_unchanged&&report.disposable_stage_removed&&report.online_overwrite_available===false&&coverage?.prerequisites_satisfied&&strategy?.complete&&strategy.domain_checks_satisfied&&strategy.blocking_issues===0&&report.dependency_conflicts===0&&report.file_conflicts===0&&report.tables.every(table=>!table.unsupported_reason&&!table.protected_new_rows&&(['tickets','ticket_messages','customer_service_inbox_states','ticket_message_media_attachments'].includes(table.table)||(!table.missing_rows&&!table.conflicting_rows)));
}
async function saveSelectedCustomerServiceModuleRestoreChoice(taskID,previewSHA256,policy) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(!isSelectedCustomerServiceModuleStrategySaveAvailable(task,policy)||task.selected_module_restore_preview.sha256!==previewSHA256)throw new Error('聊天预览、任务或权限已变化，请重新预览。');
  const revision=maintenanceConnectionRevision;
  await postMaintenanceRequest(`/api/tasks/${taskID}/selected-customer-service-restore-choice`,{preview_sha256:previewSHA256,policy});
  if(revision!==maintenanceConnectionRevision)return;
  showMessage('聊天恢复策略已保存，尚未覆盖商城。可以继续在独立副本验证完整记录。');
}
async function startSelectedCustomerServiceModuleCandidateRehearsal(taskID,choiceSHA256) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID),choice=task?.selected_customer_service_restore_choice;
  if(!choice||!isSelectedCustomerServiceModuleStrategySaveAvailable(task,choice.policy)||choice.sha256!==choiceSHA256||choice.preview_sha256!==task.selected_module_restore_preview.sha256)throw new Error('已保存策略、预览或权限已变化，请重新预览。');
  const phrase=`VERIFY CHAT CHOICE ${targetStatus.site_id}/${targetStatus.environment}/${taskID}/${choiceSHA256}`;
  const confirmation=prompt(`将暂停商城服务，重新核对所选策略绑定的包和当前数据，再在一次性副本写入完整聊天记录。完成后删除副本并恢复当前商城。本次只演练，正式恢复须另行确认。\n请输入：${phrase}`);
  if(confirmation===null)return;
  const current=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(current!==task||!isSelectedCustomerServiceModuleStrategySaveAvailable(current,choice.policy))throw new Error('当前任务已变化，请刷新后重试。');
  const revision=maintenanceConnectionRevision;
  await postMaintenanceRequest(`/api/tasks/${taskID}/selected-customer-service-candidate-rehearsal`,{choice_sha256:choiceSHA256,confirmation});
  if(revision!==maintenanceConnectionRevision)return;
  pendingVerificationTaskID=taskID;getMaintenanceConsoleElement('preview-panel').hidden=true;
  showMessage('正在验证所选聊天策略；完成后查看副本验证结果。');
}
function isCustomerServiceModuleFormalRestoreActionAvailable(task,status,recoverInterrupted) {
  if(!status?.permissions?.includes('ops:backup:restore')||task?.kind!=='upload'||task.manifest?.scope!=='selected_business_modules'||JSON.stringify(task.manifest.selected_modules)!=='["customer_service"]'||task.recovery_journal||task.configuration_site_content_restore||task.media_module_restore||task.order_module_restore||task.customer_profile_module_restore||['running','queued'].includes(task.state))return false;
  if(recoverInterrupted)return Boolean(task.customer_service_module_restore&&task.customer_service_module_restore.phase!=='released'&&['failed','interrupted'].includes(task.state));
  const choice=task.selected_customer_service_restore_choice,rehearsal=task.selected_customer_service_candidate_rehearsal,original=task.selected_module_restore_preview;
  const strategy=choice?.policy==='use_package_conversation_rows'?rehearsal?.customer_service_overwrite_preview?.package_rows:choice?.policy==='preserve_current_conflicting_conversation_rows'?rehearsal?.customer_service_overwrite_preview?.preserve_current_rows:null;
  return Boolean(!status.application_fenced&&!task.customer_service_module_restore&&!task.customer_profile_module_restore&&!task.work_copy_expired&&!task.stage_database&&!task.selected_module_disposable_stage&&new Date(task.expires_at).getTime()>Date.now()&&task.state==='module_chat_candidate_rehearsal_verified'&&choice?.sha256&&choice.state==='recorded_not_executable'&&choice.execution_available===false&&original?.sha256===choice.preview_sha256&&rehearsal?.sha256&&rehearsal.binding&&rehearsal.saved_choice_sha256===choice.sha256&&rehearsal.customer_service_formal_restore_policy==='restore_selected_conversation_rows_preserve_current_shared_facts_v1'&&rehearsal.target_unchanged&&rehearsal.disposable_stage_removed&&rehearsal.isolated_restore_verified&&rehearsal.online_overwrite_available===false&&rehearsal.customer_service_overwrite_preview?.prerequisites_satisfied&&rehearsal.dependency_conflicts===0&&rehearsal.file_conflicts===0&&strategy?.complete&&strategy.domain_checks_satisfied&&strategy.blocking_issues===0);
}
async function startCustomerServiceModuleFormalRestoreOrInterruptedRecovery(taskID,recoverInterrupted) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(!isCustomerServiceModuleFormalRestoreActionAvailable(task,targetStatus,recoverInterrupted))throw new Error('聊天恢复策略、演练或权限已变化，请刷新后重新核对。');
  const taskSnapshot=JSON.stringify(task),statusSnapshot=JSON.stringify(targetStatus),credential=maintenanceCredential,revision=maintenanceConnectionRevision;
  const phrase=recoverInterrupted?`RECOVER CHAT RESTORE ${targetStatus.site_id}/${targetStatus.environment}/${taskID}`:`RESTORE CHAT MODULE ${targetStatus.site_id}/${targetStatus.environment}/${taskID}/${task.selected_customer_service_restore_choice.sha256}`;
  const label=task.selected_customer_service_restore_choice?.policy==='use_package_conversation_rows'?'采用备份中的聊天记录':'保留当前存在冲突的聊天记录';
  const confirmation=prompt(`${recoverInterrupted?'将核验聊天恢复的提交收据和附件；未提交则撤回本次新增文件，已提交则验证完整结果，核验通过后恢复商城服务。':`将暂停商城，按“${label}”恢复四张会话表和缺失附件。保留当前新增会话、消息、订单和客户；保留事件原处理状态。成功后恢复商城服务。`}\n请输入：${phrase}`);
  if(confirmation===null)return;
  const current=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision||JSON.stringify(current)!==taskSnapshot||JSON.stringify(targetStatus)!==statusSnapshot||!isCustomerServiceModuleFormalRestoreActionAvailable(current,targetStatus,recoverInterrupted))throw new Error('聊天恢复状态已变化，请刷新后重新核对。');
  const body=recoverInterrupted?{confirmation}:{confirmation,expected_choice_sha256:current.selected_customer_service_restore_choice.sha256,expected_rehearsal_sha256:current.selected_customer_service_candidate_rehearsal.sha256};
  await postMaintenanceRequest(`/api/tasks/${taskID}/${recoverInterrupted?'customer-service-module-interrupted-restore-recovery':'customer-service-module-formal-restore'}`,body);
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision)return;
  showMessage(recoverInterrupted?'正在核验聊天提交结果与附件，请查看任务最终状态。':'正在按所选策略恢复聊天与附件，请查看任务最终状态。');
}
