'use strict';
// customer profile restore eligibility and confirmation. Loaded in index.html order; session credentials stay in memory.
function isCustomerProfileModuleFormalRestoreActionAvailable(task,status,recoverInterrupted) {
  if(!status?.permissions?.includes('ops:backup:restore')||task?.kind!=='upload'||task.manifest?.scope!=='selected_business_modules'||JSON.stringify(task.manifest.selected_modules)!=='["customers"]'||task.recovery_journal||task.configuration_site_content_restore||task.media_module_restore||task.order_module_restore||task.customer_service_module_restore||['running','queued'].includes(task.state))return false;
  if(recoverInterrupted)return Boolean(task.customer_profile_module_restore&&task.customer_profile_module_restore.phase!=='released'&&['failed','interrupted'].includes(task.state));
  const report=task.selected_module_restore_preview;
  return Boolean(!status.application_fenced&&!task.customer_profile_module_restore&&!task.work_copy_expired&&!task.stage_database&&!task.selected_module_disposable_stage&&new Date(task.expires_at).getTime()>Date.now()&&task.state==='module_isolated_restore_verified'&&report?.sha256&&report.binding&&JSON.stringify(report.modules)==='["customers"]'&&report.target_unchanged&&report.disposable_stage_removed&&report.isolated_restore_verified&&report.online_overwrite_available===false&&report.customer_profile_restore_eligibility?.eligible&&report.customer_profile_restore_eligibility.policy==='restore_missing_customer_profiles_and_closed_new_account_facts_preserve_current_identities_v1');
}
async function startCustomerProfileModuleFormalRestoreOrInterruptedRecovery(taskID,recoverInterrupted) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(!isCustomerProfileModuleFormalRestoreActionAvailable(task,targetStatus,recoverInterrupted))throw new Error('客户恢复资格、任务或权限已变化，请刷新后重新核对。');
  const taskSnapshot=JSON.stringify(task),statusSnapshot=JSON.stringify(targetStatus),credential=maintenanceCredential,revision=maintenanceConnectionRevision;
  const phrase=`${recoverInterrupted?'RECOVER CUSTOMER PROFILES':'RESTORE CUSTOMER PROFILES'} ${targetStatus.site_id}/${targetStatus.environment}/${taskID}`;
  const confirmation=prompt(`${recoverInterrupted?'将核验本次客户恢复的提交凭证；未提交则核对并撤回本次新增文件，已提交则验证结果，核验通过后恢复商城服务。':'将暂停商城，补回缺失客户及符合条件的完整账号关联。当前已有账号、密码、账号限制、积分和订阅状态保留；订单和聊天保留。不会调用注册或奖励服务。缺失积分只支持整个账号组缺失且无当前金融足迹的情况。成功后恢复商城服务。'}\n请输入：${phrase}`);
  if(confirmation===null)return;
  const current=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision||JSON.stringify(current)!==taskSnapshot||JSON.stringify(targetStatus)!==statusSnapshot||!isCustomerProfileModuleFormalRestoreActionAvailable(current,targetStatus,recoverInterrupted))throw new Error('客户恢复状态已变化，请刷新后重新核对。');
  await postMaintenanceRequest(`/api/tasks/${taskID}/${recoverInterrupted?'customer-profile-module-interrupted-restore-recovery':'customer-profile-module-formal-restore'}`,recoverInterrupted?{confirmation}:{confirmation,expected_preview_sha256:current.selected_module_restore_preview.sha256});
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision)return;
  showMessage(recoverInterrupted?'正在核验客户恢复的提交结果，请查看任务最终状态。':'正在恢复缺失客户资料，请查看任务最终状态。');
}
