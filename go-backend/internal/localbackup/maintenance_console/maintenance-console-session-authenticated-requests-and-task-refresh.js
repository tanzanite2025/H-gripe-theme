'use strict';
// session authenticated requests and task refresh. Loaded in index.html order; session credentials stay in memory.
const consolePrefix = location.pathname.includes('/local-backup-maintenance') ? '/local-backup-maintenance' : '';
let maintenanceCredential = '';
let targetStatus = null;
let refreshTimer = null;
let maintenanceConnectionRevision = 0;
let displayedMaintenanceTasks = [];
function disconnectMaintenanceConsoleAndInvalidatePendingClientOperations(message, failed = false) {
  maintenanceCredential = '';
  maintenanceConnectionRevision++;
  displayedMaintenanceTasks = [];
  targetStatus = null;
  businessBackupRequestInProgress = false;
  uploading = false;
  clearPendingAdminBusinessBackupScopeSelectionAndWorkflow();
  invalidateSelectedBusinessBackupPreview();
  clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();
  clearInterval(refreshTimer);
  getMaintenanceConsoleElement('connected-panel').hidden = true;
  getMaintenanceConsoleElement('login-panel').hidden = false;
  showMessage(message, failed);
  if(window.parent!==window)window.parent.postMessage({type:'local-backup-maintenance-disconnected'},location.origin);
}
function requireUnchangedMaintenanceConnectionForPendingOperation(connectionRevision) {
  if(connectionRevision!==maintenanceConnectionRevision)throw new Error('维护会话已更换，请重新操作。');
}
async function maintenanceRequest(path, options = {}) {
  const connectionRevision = maintenanceConnectionRevision;
  const headers = new Headers(options.headers || {});
  headers.set('Authorization', `Bearer ${maintenanceCredential}`);
  if (options.body && typeof options.body === 'string') headers.set('Content-Type', 'application/json');
  const response = await fetch(`${consolePrefix}${path}`, { ...options, headers, cache: 'no-store', credentials: 'same-origin' });
  const result = await response.json().catch(() => ({}));
  requireUnchangedMaintenanceConnectionForPendingOperation(connectionRevision);
  if (!response.ok) {
    if(response.status===401)disconnectMaintenanceConsoleAndInvalidatePendingClientOperations('维护凭据无效或已过期，请重新连接；服务器任务及维护隔离保持原状。', true);
    const error=new Error(result.error || `请求失败 (${response.status})`);
    error.httpStatus=response.status;
    throw error;
  }
  return result;
}
function postMaintenanceRequest(path, body = {}) { return maintenanceRequest(path, {method:'POST',body:JSON.stringify(body)}); }
function hasPermission(permission) { return targetStatus?.permissions.includes(`ops:backup:${permission}`); }
async function connectMaintenanceConsole(credential) {
  const connectionRevision=++maintenanceConnectionRevision;
  clearPendingAdminBusinessBackupScopeSelectionAndWorkflow();
  businessBackupRequestInProgress = false;
  uploading = false;
  targetStatus = null;
  displayedMaintenanceTasks = [];
  clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();
  maintenanceCredential = credential;
  invalidateSelectedBusinessBackupPreview();
  try {
    await refreshMaintenanceTasks();
  } catch(error) {
    if(connectionRevision===maintenanceConnectionRevision)disconnectMaintenanceConsoleAndInvalidatePendingClientOperations('维护入口连接失败，请检查服务状态后重新连接。', true);
    throw error;
  }
  if(connectionRevision!==maintenanceConnectionRevision)return;
  getMaintenanceConsoleElement('credential').value = '';
  getMaintenanceConsoleElement('login-panel').hidden = true;
  getMaintenanceConsoleElement('connected-panel').hidden = false;
  clearInterval(refreshTimer);
  refreshTimer = setInterval(() => { if (!uploading) refreshMaintenanceTasks().catch((error) => showMessage(error.message, true)); }, 3000);
}
async function refreshMaintenanceTasks() {
  const connectionRevision = maintenanceConnectionRevision;
  let status, result;
  try {
    [status, result] = await Promise.all([maintenanceRequest('/api/status'), maintenanceRequest('/api/tasks')]);
  } catch (error) {
    if (connectionRevision === maintenanceConnectionRevision) disconnectMaintenanceConsoleAndInvalidatePendingClientOperations('获取维护服务状态失败，请重新连接；服务器任务及维护隔离保持原状。', true);
    throw error;
  }
  if (connectionRevision !== maintenanceConnectionRevision) return;
  if(stoppedProductProjectionEvidenceContext){const observed=stoppedProductProjectionEvidenceContext;const current=result.tasks.find(task=>task.id===observed.taskID);if(!current||JSON.stringify(current)!==observed.taskSnapshot||JSON.stringify(status)!==observed.statusSnapshot||!isStoppedProductProjectionInspectionAvailable(current,status))clearStoppedProductProjectionEvidencePanel();}
  if (productCatalogStoppedPostSwitchObservedTask) {
    const observed = productCatalogStoppedPostSwitchObservedTask;
    const currentTask = result.tasks.find(task => task.id === observed.taskID);
    if (!currentTask || JSON.stringify(currentTask) !== observed.taskSnapshot || JSON.stringify(status) !== observed.statusSnapshot || !isStoppedProductCatalogPostSwitchInspectionAvailable(currentTask,status)) clearStoppedProductCatalogPostSwitchInspection();
  }
  if (productCatalogHandoffInspectionObservedTask) {
    const observed = productCatalogHandoffInspectionObservedTask;
    const currentTask = result.tasks.find(task => task.id === observed.taskID);
    if (!currentTask || JSON.stringify(currentTask) !== observed.taskSnapshot || JSON.stringify(status) !== observed.statusSnapshot || !isProductCatalogCandidateEligibleForHandoffInspection(currentTask, status)) clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();
  }
  displayedMaintenanceTasks = result.tasks;
  targetStatus = status;
  getMaintenanceConsoleElement('target-title').textContent = `目标商城：${status.site_id} / ${status.environment}`;
  getMaintenanceConsoleElement('target-description').textContent = `数据库 ${status.database_name} · 应用版本 ${status.application_version} · 包版本 ${status.package_version} · 服务器工作副本保留 ${status.work_copy_lifetime_hours} 小时 · 单包上限 ${formatBackupByteSize(status.max_package_bytes)}`;
  getMaintenanceConsoleElement('maintenance-status').textContent = status.application_fenced ? '商城维护隔离中：业务进程、任务和回调处理保持停止。' : '商城未处于备份维护隔离。';
  renderBusinessBackupModuleChoices();
  updateBusinessBackupSelectionControls();
  renderBackupTasks(result.tasks);
  const selectedModulePanel=getMaintenanceConsoleElement('selected-module-restore-preview');
  if(!getMaintenanceConsoleElement('preview-panel').hidden&&selectedModulePanel.dataset.taskId){
    const current=result.tasks.find(task=>task.id===selectedModulePanel.dataset.taskId);
    if(current)renderSelectedBusinessModuleIsolatedRestorePreview(current);
    else {selectedModulePanel.replaceChildren();selectedModulePanel.hidden=true;}
  }
  if (pendingVerificationTaskID) {
    const task=result.tasks.find(task=>task.id===pendingVerificationTaskID);
    if (task && ['configuration_site_content_restored_still_isolated','verified','product_module_isolated_verified','module_package_integrity_verified','module_isolated_restore_verified','module_restore_conflicts_found','module_chat_candidate_rehearsal_verified','failed','interrupted'].includes(task.state)) {
      pendingVerificationTaskID='';
      showMessage(task.state==='configuration_site_content_restored_still_isolated'?'网站内容已写入，商城继续维护，请完成内容与缓存核对后恢复服务。':task.state==='module_chat_candidate_rehearsal_verified'?'所选聊天策略已在副本写入完整记录并验证，副本已删除，当前商城未覆盖。':task.state==='module_isolated_restore_verified'?'所选模块的隔离恢复验证通过，当前商城数据未改变。请查看任务中的可用恢复方式与限制。':task.state==='module_restore_conflicts_found'?'发现模块恢复冲突，请查看验证结果。当前商城数据未改变。':task.state==='module_package_integrity_verified'?'所选模块记录及实际文件完整性校验通过；请继续验证模块恢复，再查看当前范围的恢复资格。':task.state==='product_module_isolated_verified'?'商品与实际图片的隔离恢复验证通过，其他业务数据保留。':task.state==='verified'?'整站包隔离恢复验证通过，请查看覆盖预览。':`恢复验证未通过：${task.error||'请检查任务记录。'}`,['failed','interrupted','module_restore_conflicts_found'].includes(task.state));
    }
  }
}
