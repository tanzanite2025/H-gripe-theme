'use strict';
// standalone media restore eligibility and confirmation. Loaded in index.html order; session credentials stay in memory.
function isMediaModuleFormalRestoreActionAvailable(task,status,recoverInterrupted) {
  if(!status?.permissions?.includes('ops:backup:restore')||task?.kind!=='upload'||task.manifest?.scope!=='selected_business_modules'||JSON.stringify(task.manifest.selected_modules)!=='["media"]'||task.recovery_journal||task.configuration_site_content_restore||task.order_module_restore||task.customer_service_module_restore||task.customer_profile_module_restore||['running','queued'].includes(task.state))return false;
  if(recoverInterrupted)return Boolean(task.media_module_restore&&task.media_module_restore.phase!=='released'&&['failed','interrupted'].includes(task.state));
  const report=task.selected_module_restore_preview;
  const repairs=report?.media_restore_eligibility?.existing_file_repairs??0,repairDigest=report?.media_restore_eligibility?.existing_file_repair_plan_sha256??'';
  if(!Number.isSafeInteger(repairs)||repairs<0||(repairs===0?repairDigest!=='':!/^[a-f0-9]{64}$/.test(repairDigest)))return false;
  return Boolean(!status.application_fenced&&!task.media_module_restore&&!task.work_copy_expired&&!task.stage_database&&!task.selected_module_disposable_stage&&new Date(task.expires_at).getTime()>Date.now()&&task.state==='module_isolated_restore_verified'&&report?.sha256&&report.binding&&JSON.stringify(report.modules)==='["media"]'&&report.target_unchanged&&report.disposable_stage_removed&&report.isolated_restore_verified&&report.online_overwrite_available===false&&report.media_restore_eligibility?.eligible&&report.media_restore_eligibility.policy==='restore_missing_media_and_repair_registered_file_bytes_preserve_current_business_data_v2');
}
async function startMediaModuleFormalRestoreOrInterruptedRecovery(taskID,recoverInterrupted) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(!isMediaModuleFormalRestoreActionAvailable(task,targetStatus,recoverInterrupted))throw new Error('媒体恢复资格、任务或权限已变化，请刷新后重新核对。');
  const taskSnapshot=JSON.stringify(task),statusSnapshot=JSON.stringify(targetStatus),credential=maintenanceCredential,revision=maintenanceConnectionRevision;
  const phrase=`${recoverInterrupted?'RECOVER MEDIA':'RESTORE MEDIA'} ${targetStatus.site_id}/${targetStatus.environment}/${taskID}`;
  const confirmation=prompt(`${recoverInterrupted?'将核验本次媒体恢复的提交凭证；未提交则撤回新增文件并恢复本次替换的原字节，已提交则验证结果，核验通过后恢复商城服务。':`将暂停商城，补回缺失原件、派生文件、媒体登记和业务绑定。保留当前商品、订单、客户、聊天和配置。本次将修复 ${task.selected_module_restore_preview?.media_restore_eligibility?.existing_file_repairs||0} 个登记一致的文件，原字节保留至提交核验；同 ID 内容不同、未登记字节冲突、身份冲突或未完成任务会阻断。成功后恢复商城服务。`}\n请输入：${phrase}`);
  if(confirmation===null)return;
  const current=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision||JSON.stringify(current)!==taskSnapshot||JSON.stringify(targetStatus)!==statusSnapshot||!isMediaModuleFormalRestoreActionAvailable(current,targetStatus,recoverInterrupted))throw new Error('媒体恢复状态已变化，请刷新后重新核对。');
  await postMaintenanceRequest(`/api/tasks/${taskID}/${recoverInterrupted?'media-module-interrupted-restore-recovery':'media-module-formal-restore'}`,recoverInterrupted?{confirmation}:{confirmation,expected_preview_sha256:current.selected_module_restore_preview.sha256});
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision)return;
  showMessage(recoverInterrupted?'正在核验媒体恢复的提交结果，请查看任务最终状态。':'正在恢复图片与附件，请查看任务最终状态。');
}
