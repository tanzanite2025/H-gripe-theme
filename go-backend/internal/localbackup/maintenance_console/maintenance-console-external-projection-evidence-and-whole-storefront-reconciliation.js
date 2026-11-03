'use strict';
// external projection evidence and whole storefront reconciliation. Loaded in index.html order; session credentials stay in memory.
let reconciliationTask = null;
let inspectedReconciliationReport = null;
function renderExternalProjectionReconciliationEvidenceRequirements(contract,containerID='external-projection-reconciliation-evidence',headingText='搜索、缓存、实时消息与附件投影核对') {
  const container = getMaintenanceConsoleElement(containerID);
  container.replaceChildren();
  const heading = document.createElement('h3'); heading.textContent = headingText; container.append(heading);
  if (!contract || contract.blockers?.length) {
    const blocked = document.createElement('p'); blocked.className = 'error';
    blocked.textContent = '业务配置尚未完整登记或与当前恢复不符。请使用当前业务构建重新登记配置，再读取对账快照。'; container.append(blocked); return;
  }
  const labels = {customer_service_search:'客服搜索',redis_cache:'商城缓存与任务隔离',customer_service_realtime:'实时消息与重新连接',media_delivery_and_browser:'附件、边缘缓存与浏览器',product_storefront_http_routes:'商品页面、搜索与当前库存',product_public_private_media_delivery:'公开图片与私有附件权限',product_html_edge_and_browser_caches:'HTML、边缘缓存与全新浏览器'};
  const note = document.createElement('p'); note.textContent = '每项都须实际核对并提供电脑上的独立记录。记录文件只在浏览器中计算摘要，不会上传；请自行保留原文件。未知归属或结果未确认时，请保持维护隔离。'; container.append(note);
  for (const requirement of contract.requirements) {
    const fieldset = document.createElement('fieldset'); fieldset.dataset.projection = requirement.projection;
    const legend = document.createElement('legend'); legend.textContent = labels[requirement.projection] || requirement.projection;
    const description = document.createElement('p'); description.textContent = requirement.description;
    const checkboxLabel = document.createElement('label'); const checkbox = document.createElement('input'); checkbox.type = 'checkbox'; checkbox.required = true; checkbox.dataset.projectionConfirmed = 'true'; checkboxLabel.append(checkbox,document.createTextNode(' 已完成上述核对，独立记录与当前恢复快照一致'));
    const referenceLabel = document.createElement('label'); referenceLabel.textContent = '核对记录编号或保存位置'; const reference = document.createElement('input'); reference.required = true; reference.minLength = 4; reference.maxLength = 2000; reference.dataset.projectionReference = 'true'; referenceLabel.append(reference);
    const fileLabel = document.createElement('label'); fileLabel.textContent = '选择独立核对记录（仅计算摘要，最大 16 MB）'; const file = document.createElement('input'); file.type = 'file'; file.required = true; file.dataset.projectionRecord = 'true'; fileLabel.append(file);
    fieldset.append(legend,description,checkboxLabel,referenceLabel,fileLabel); container.append(fieldset);
  }
}
async function collectExternalProjectionReconciliationOperatorAttestations(contract,containerID='external-projection-reconciliation-evidence') {
  if (!contract || contract.blockers?.length) throw new Error('投影配置未登记完整，无法记录核对凭证。');
  const attestations = [];
  for (const requirement of contract.requirements) {
    const fieldset = Array.from(getMaintenanceConsoleElement(containerID).querySelectorAll('fieldset')).find(element => element.dataset.projection === requirement.projection);
    const reference = fieldset?.querySelector('[data-projection-reference]')?.value?.trim();
    const file = fieldset?.querySelector('[data-projection-record]')?.files?.[0];
    if (!fieldset?.querySelector('[data-projection-confirmed]')?.checked || !reference || reference.length < 4 || !file || file.size === 0 || file.size > 16 * 1024 * 1024) throw new Error('请完成每项核对并选择非空、最大 16 MB 的独立记录。');
    const digest = await crypto.subtle.digest('SHA-256', await file.arrayBuffer());
    attestations.push({projection:requirement.projection,action:requirement.action,evidence_reference:reference,evidence_sha256:Array.from(new Uint8Array(digest),byte=>byte.toString(16).padStart(2,'0')).join('')});
  }
  return {contract_sha256:contract.contract_sha256,attestations};
}
async function handleWholeStorefrontReconciliationSnapshotInspection() {
  if(!reconciliationTask)return;
  inspectedReconciliationReport=null;
  getMaintenanceConsoleElement('complete-reconciliation').disabled=true;
  getMaintenanceConsoleElement('external-projection-reconciliation-evidence').replaceChildren();
  try {
    const taskID=reconciliationTask.id;
    const result=await maintenanceRequest(`/api/tasks/${taskID}/reconciliation-report`);
    if(reconciliationTask?.id!==taskID)return;
    inspectedReconciliationReport=result.report;
    getMaintenanceConsoleElement('reconciliation-report').textContent=JSON.stringify(result.report,null,2);
    const contract=result.report.external_projection_reconciliation;
    renderExternalProjectionReconciliationEvidenceRequirements(contract);
    const blocked=result.report.gaps.length>0||!contract||contract.blockers.length>0;
    getMaintenanceConsoleElement('complete-reconciliation').disabled=blocked;
    showMessage(blocked?'对账快照仍有数据缺口或未登记的外部投影，恢复服务被阻断。':'已读取对账快照，请完成各项核验并提供独立记录。');
  } catch(error) {showMessage(error.message,true);}
}
async function handleWholeStorefrontReconciliationEvidenceSubmission(event) {
  event.preventDefault();if(!reconciliationTask||!inspectedReconciliationReport)return;
  const task=reconciliationTask;const report=inspectedReconciliationReport;
  try {
    const externalProjections=await collectExternalProjectionReconciliationOperatorAttestations(report.external_projection_reconciliation);
    if(reconciliationTask!==task||inspectedReconciliationReport!==report)throw new Error('对账快照已变化，请重新核对。');
    const confirmation=prompt(`人工核对结果和独立记录摘要将永久记入恢复日志。请输入：RESUME ${targetStatus.site_id}/${targetStatus.environment}/${task.id}`);if(confirmation===null)return;
    const evidence={...Object.fromEntries(new FormData(event.target)),reconciled_database_fingerprint:report.database_fingerprint,external_projections:externalProjections};
    await postMaintenanceRequest(`/api/tasks/${task.id}/complete`,{confirmation,evidence});
    getMaintenanceConsoleElement('reconciliation-panel').hidden=true;showMessage('核对结果已记录，已解除维护隔离。请检查商城实际启动和业务访问结果。');await refreshMaintenanceTasks();
  } catch(error) {showMessage(error.message,true);}
}
async function confirmAndApplyWholeStorefrontPairedRecovery(task) {showOverwritePreview(task);const phrase=`RESTORE ${targetStatus.site_id}/${targetStatus.environment}/${task.manifest.package_id}`;const confirmation=prompt(`将整体商城恢复到 ${task.manifest.captured_at}。该时间点之后的在线数据会被替换。操作期间商城停止服务，并保留成对回退数据。\n请输入：${phrase}`);if(confirmation===null)return;await postMaintenanceRequest(`/api/tasks/${task.id}/apply`,{confirmation});showMessage('正在进行成对切换，完成后仍保持隔离，等待对账。');}
async function confirmAndOpenWholeStorefrontReconciliationConnections(task) {const phrase=`RECONCILE ${targetStatus.site_id}/${targetStatus.environment}/${task.id}`;const confirmation=prompt(`商城进程保持停止。此窗口仅供授权维护人员通过数据库工具核对和修正支付、物流、持久化任务等事实。恢复服务前须关闭维护 SQL 连接。\n请输入：${phrase}`);if(confirmation===null)return;await postMaintenanceRequest(`/api/tasks/${task.id}/reconciliation-connections`,{confirmation});showMessage('已开放维护数据库连接，商城进程继续隔离。完成核对后关闭 SQL 连接并读取对账快照。');}
function openWholeStorefrontReconciliationEvidencePanel(task) {reconciliationTask=task;inspectedReconciliationReport=null;getMaintenanceConsoleElement('complete-reconciliation').disabled=true;getMaintenanceConsoleElement('reconciliation-report').textContent='';getMaintenanceConsoleElement('reconciliation-panel').hidden=false;getMaintenanceConsoleElement('reconciliation-panel').scrollIntoView({behavior:'smooth'});}
async function confirmAndRollBackWholeStorefrontDatabaseAndMedia(task) {const phrase=`ROLLBACK ${targetStatus.site_id}/${task.id}`;const confirmation=prompt(`将数据库和附件一并回退至覆盖前的状态。\n请输入：${phrase}`);if(confirmation===null)return;await postMaintenanceRequest(`/api/tasks/${task.id}/rollback`,{confirmation});showMessage('正在成对回退，核对通过后解除隔离。');}
