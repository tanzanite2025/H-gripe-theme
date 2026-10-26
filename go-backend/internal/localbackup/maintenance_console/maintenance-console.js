'use strict';
let selectedBusinessBackupPreview = null;
let selectedBusinessBackupRevision = 0;
let businessBackupRequestInProgress = false;
const consolePrefix = location.pathname.includes('/local-backup-maintenance') ? '/local-backup-maintenance' : '';
let maintenanceCredential = '';
let targetStatus = null;
let reconciliationTask = null;
let inspectedReconciliationReport = null;
let refreshTimer = null;
let uploading = false;
let pendingVerificationTaskID = '';
let customerServiceConflictReviewContext = null;
let productCatalogHandoffInspectionRequestRevision = 0;
let maintenanceConnectionRevision = 0;
let displayedMaintenanceTasks = [];
let productCatalogHandoffInspectionObservedTask = null;
let productCatalogStoppedPostSwitchInspectionRevision = 0;
let productCatalogStoppedPostSwitchObservedTask = null;
let stoppedProductProjectionEvidenceContext = null;
let stoppedProductProjectionInspectionRevision = 0;
let stoppedProductReadOnlyHTTPQueryRevision = 0;
const getMaintenanceConsoleElement = (id) => document.getElementById(id);
const formatBackupByteSize = (value) => value < 1024 * 1024 ? `${(value / 1024).toFixed(1)} KB` : `${(value / 1024 / 1024).toFixed(1)} MB`;
const productCatalogPhaseNames = {capturing_product_catalog_and_images:'正在读取商品与图片',product_catalog_and_actual_images_encrypted:'商品和实际图片已加密保存',restoring_product_module_in_isolated_target_clone:'正在隔离副本恢复商品',product_rows_images_and_unselected_data_preservation_verified:'商品、图片和其他数据保留验证完成'};
const stateNames = {queued:'等待执行',running:'执行中',ready_to_download:'已生成，可下载',generated_with_gaps:'已生成，但缺口阻止恢复',uploading:'上传中',uploaded:'上传完成，尚未验证',verified:'隔离恢复验证通过',awaiting_reconciliation:'已切换，等待对账',resuming:'对账已记录，正在恢复服务',releasing_rollback:'回退已核对，正在解除隔离',completed:'已记录对账并解除隔离',rolled_back:'已成对回退',failed:'失败',interrupted:'执行中断，请检查阶段',expired:'服务器工作副本已过期'};
function showMessage(message, failed = false) { const output = getMaintenanceConsoleElement('message'); output.textContent = message; output.className = failed ? 'error' : ''; }
stateNames.product_module_isolated_verified = '商品与图片隔离恢复通过';
stateNames.module_package_integrity_verified = '模块内容完整性校验通过，尚未恢复';
stateNames.module_isolated_restore_verified = '模块隔离恢复验证通过，未覆盖当前商城';
stateNames.module_restore_conflicts_found = '模块恢复存在冲突，请查看结果';
stateNames.module_chat_candidate_rehearsal_verified = '所选聊天策略副本验证通过，未覆盖当前商城';
Object.assign(productCatalogPhaseNames,{preparing_selected_chat_choice_candidate_rehearsal:'准备所选聊天策略副本验证',selected_chat_choice_full_rows_and_retained_target_verified:'完整聊天记录及其他数据保留核对通过'});
Object.assign(productCatalogPhaseNames,{preparing_selected_module_isolated_restore:'准备模块隔离恢复',checking_selected_module_target_dependencies:'检查模块依赖与内容冲突',selected_module_disposable_restore_and_target_preservation_verified:'隔离恢复及当前商城数据保留验证通过',selected_module_dependency_or_content_conflicts_require_resolution:'依赖、记录或文件存在冲突，恢复已阻断'});
Object.assign(productCatalogPhaseNames,{capturing_selected_business_modules:'正在读取所选业务模块和实际附件',selected_business_modules_encrypted:'所选业务模块及实际文件已加密保存',selected_module_contents_verified_restore_unavailable:'模块内容已校验，正式覆盖导入尚未开放'});
stateNames.product_module_stopped_handoff_prepared = '商品交接准备完成，商城持续停写';
stateNames.product_module_switched_still_isolated = '商品已隔离切换，商城继续停写';
Object.assign(productCatalogPhaseNames,{invalidating_stopped_product_detail_cache_and_read_leases:'正在核验并处理商品详情缓存与读取锁',product_detail_cache_and_read_leases_verified_absent_still_isolated:'商品详情缓存与读取锁已核验为空，继续停写'});
Object.assign(productCatalogPhaseNames,{switching_product_with_application_continuously_stopped:'正在持续停写窗口内切换商品',product_stopped_switch_journal_persisted:'商品切换意图已保存，失败时须成对回退',product_database_and_complete_media_switched_application_stays_stopped:'商品数据库与完整媒体已切换，继续停写',rolling_back_stopped_product_database_and_complete_media:'正在成对回退商品数据库与完整媒体',product_stopped_switch_requires_explicit_verified_rollback:'商品切换中断，需明确执行核验回退',product_original_database_and_complete_media_verified_before_release:'原商城数据库与完整媒体已核验，正在解除隔离',product_stopped_switch_original_service_restored_after_verified_paired_rollback:'商品切换已成对回退，原商城已恢复服务'});
Object.assign(productCatalogPhaseNames,{preparing_product_handoff_with_continuous_isolation:'正在持续停写窗口内准备商品交接',product_handoff_prepared_application_remains_stopped:'商品候选已重建，继续保持停写',product_stopped_handoff_requires_explicit_revalidation:'交接准备中断，需重新核验或取消',verifying_original_before_product_handoff_cancellation:'正在核对当前商城后取消准备',product_handoff_cancelled_original_service_released:'交接准备已取消，当前商城已解除隔离'});
Object.assign(productCatalogPhaseNames,{waiting_for_product_view_writers_to_stop:'等待商品浏览相关进程停止',draining_idempotent_product_view_batches:'正在幂等保存浏览增量',rebuilding_product_candidate_after_view_drain:'正在基于最新浏览数重建商品预览',product_view_drain_and_fresh_candidate_rebuild_verified:'浏览增量已保存，商品预览已重建',product_view_drain_retained_for_explicit_retry:'浏览排空中断，维护隔离保留，需检查后重试'});
async function maintenanceRequest(path, options = {}) {
  const headers = new Headers(options.headers || {});
  headers.set('Authorization', `Bearer ${maintenanceCredential}`);
  if (options.body && typeof options.body === 'string') headers.set('Content-Type', 'application/json');
  const response = await fetch(`${consolePrefix}${path}`, { ...options, headers, cache: 'no-store', credentials: 'same-origin' });
  const result = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(result.error || `请求失败 (${response.status})`);
  return result;
}
function postMaintenanceRequest(path, body = {}) { return maintenanceRequest(path, {method:'POST',body:JSON.stringify(body)}); }
function hasPermission(permission) { return targetStatus?.permissions.includes(`ops:backup:${permission}`); }
function selectedBusinessBackupModuleIDs() {
  return Array.from(document.querySelectorAll('#business-backup-module-options input:checked'),input=>input.value).sort();
}
function isBusinessBackupModuleModeSelected() {
  return document.querySelector('input[name="backup-export-mode"]:checked')?.value==='modules';
}
function describeSelectedBusinessBackupModules(modules) {
  const labels=new Map((targetStatus?.business_module_backup_options||[]).map(option=>[option.id,option.label]));
  return modules.map(module=>labels.get(module)||module).join('、');
}
function invalidateSelectedBusinessBackupPreview() {
  selectedBusinessBackupRevision++;
  selectedBusinessBackupPreview=null;
  getMaintenanceConsoleElement('business-backup-range-preview').hidden=true;
  updateBusinessBackupSelectionControls();
}
function renderBusinessBackupModuleChoices() {
  const container=getMaintenanceConsoleElement('business-backup-module-options');
  const options=targetStatus?.business_module_backup_options||[];
  const signature=JSON.stringify(options);
  if(container.dataset.signature===signature)return;
  const previous=new Set(selectedBusinessBackupModuleIDs());
  container.replaceChildren();container.dataset.signature=signature;
  for(const option of options){
    const label=document.createElement('label');label.className='backup-module-option';
    const checkbox=document.createElement('input');checkbox.type='checkbox';checkbox.value=option.id;checkbox.checked=previous.has(option.id);checkbox.addEventListener('change',invalidateSelectedBusinessBackupPreview);
    const title=document.createElement('strong');title.textContent=option.label;
    const description=document.createElement('span');description.textContent=option.description;
    const text=document.createElement('span');text.append(title,description);label.append(checkbox,text);container.append(label);
  }
  invalidateSelectedBusinessBackupPreview();
}
function updateBusinessBackupSelectionControls() {
  const moduleMode=isBusinessBackupModuleModeSelected();
  const modules=selectedBusinessBackupModuleIDs();
  const allowed=Boolean(hasPermission('export')&&!targetStatus.application_fenced&&!businessBackupRequestInProgress);
  getMaintenanceConsoleElement('business-backup-module-picker').hidden=!moduleMode;
  getMaintenanceConsoleElement('business-backup-selection-summary').textContent=modules.length?`已选择 ${modules.length} 个模块：${describeSelectedBusinessBackupModules(modules)}`:'请勾选要保存的模块。';
  getMaintenanceConsoleElement('preview-selected-business-modules').disabled=!allowed||!modules.length;
  getMaintenanceConsoleElement('generate').textContent=moduleMode?'生成所选模块备份':'生成整站备份';
  getMaintenanceConsoleElement('generate').disabled=!allowed||(moduleMode&&(!modules.length||!selectedBusinessBackupPreview));
  for(const input of document.querySelectorAll('#business-backup-selection input'))input.disabled=!allowed;
  for(const id of ['select-all-business-backup-modules','clear-business-backup-modules'])getMaintenanceConsoleElement(id).disabled=!allowed;
}
async function previewCurrentSelectedBusinessBackupModules() {
  if(!hasPermission('export')||businessBackupRequestInProgress||targetStatus.application_fenced)return;
  const modules=selectedBusinessBackupModuleIDs();
  if(!modules.length)return;
  const revision=++selectedBusinessBackupRevision;
  const connectionRevision=maintenanceConnectionRevision;
  selectedBusinessBackupPreview=null;getMaintenanceConsoleElement('business-backup-range-preview').hidden=true;businessBackupRequestInProgress=true;updateBusinessBackupSelectionControls();
  try {
    const {preview}=await postMaintenanceRequest('/api/module-backup-preview',{modules});
    if(connectionRevision!==maintenanceConnectionRevision||revision!==selectedBusinessBackupRevision)return;
    if(JSON.stringify(preview?.modules)!==JSON.stringify(modules)||preview.online_overwrite_available!==false)throw new Error('范围预览与当前选择不符，请重新预览。');
    selectedBusinessBackupPreview=preview;
    getMaintenanceConsoleElement('business-backup-range-description').textContent=`本次保存：${describeSelectedBusinessBackupModules(modules)}。`;
    const notes=getMaintenanceConsoleElement('business-backup-range-notes');notes.replaceChildren();
    for(const text of [...preview.included_dependencies,...preview.recovery_limitations]){const item=document.createElement('li');item.textContent=text;notes.append(item);}
    const rows=getMaintenanceConsoleElement('business-backup-range-tables');rows.replaceChildren();
    for(const [table,count] of Object.entries(preview.tables).sort(([left],[right])=>left.localeCompare(right))){const row=document.createElement('tr');for(const value of [table,count]){const cell=document.createElement('td');cell.textContent=String(value);row.append(cell);}rows.append(row);}
    getMaintenanceConsoleElement('business-backup-range-preview').hidden=false;
  } finally {businessBackupRequestInProgress=false;updateBusinessBackupSelectionControls();}
}
async function generateBackupForCurrentBusinessModuleSelection() {
  if(!hasPermission('export')||businessBackupRequestInProgress||targetStatus.application_fenced)return;
  const moduleMode=isBusinessBackupModuleModeSelected();
  const modules=selectedBusinessBackupModuleIDs();
  if(moduleMode&&(!selectedBusinessBackupPreview||JSON.stringify(selectedBusinessBackupPreview.modules)!==JSON.stringify(modules)))throw new Error('请先预览所选模块范围。');
  const phrase=moduleMode?`BACKUP MODULES ${targetStatus.site_id}/${targetStatus.environment}/${modules.join(',')}`:`BACKUP ${targetStatus.site_id}/${targetStatus.environment}`;
  const confirmation=prompt(`${moduleMode?`本次导出：${describeSelectedBusinessBackupModules(modules)}。`:'本次导出全部商城数据。'}生成期间商城将暂时停止服务。请确认维护窗口，并输入：${phrase}`);
  if(confirmation===null)return;
  businessBackupRequestInProgress=true;updateBusinessBackupSelectionControls();
  try {
    await postMaintenanceRequest(moduleMode?'/api/module-backups':'/api/backups',moduleMode?{modules,confirmation}:{confirmation});
    showMessage(moduleMode?'已启动所选模块备份，生成后从记录下载到电脑。':'已启动整站备份，生成后从记录下载到电脑。');
    invalidateSelectedBusinessBackupPreview();
    await refreshMaintenanceTasks();
  } finally {businessBackupRequestInProgress=false;updateBusinessBackupSelectionControls();}
}
async function connectMaintenanceConsole(credential) {
  const connectionRevision=++maintenanceConnectionRevision;
  clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();
  maintenanceCredential = credential;
  invalidateSelectedBusinessBackupPreview();
  await refreshMaintenanceTasks();
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
    if (connectionRevision === maintenanceConnectionRevision) clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();
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
  getMaintenanceConsoleElement('upload-button').disabled = !hasPermission('upload') || uploading;
  renderBackupTasks(result.tasks);
  const selectedModulePanel=getMaintenanceConsoleElement('selected-module-restore-preview');
  if(!getMaintenanceConsoleElement('preview-panel').hidden&&selectedModulePanel.dataset.taskId){
    const current=result.tasks.find(task=>task.id===selectedModulePanel.dataset.taskId);
    if(current)renderSelectedBusinessModuleIsolatedRestorePreview(current);
    else {selectedModulePanel.replaceChildren();selectedModulePanel.hidden=true;}
  }
  if (pendingVerificationTaskID) {
    const task=result.tasks.find(task=>task.id===pendingVerificationTaskID);
    if (task && ['verified','product_module_isolated_verified','module_package_integrity_verified','module_isolated_restore_verified','module_restore_conflicts_found','module_chat_candidate_rehearsal_verified','failed','interrupted'].includes(task.state)) {
      pendingVerificationTaskID='';
      showMessage(task.state==='module_chat_candidate_rehearsal_verified'?'所选聊天策略已在副本写入完整记录并验证，副本已删除，当前商城未覆盖。':task.state==='module_isolated_restore_verified'?'所选模块的隔离恢复验证通过，当前商城数据未改变。请查看结果；正式覆盖导入尚未开放。':task.state==='module_restore_conflicts_found'?'发现模块恢复冲突，请查看验证结果。当前商城数据未改变。':task.state==='module_package_integrity_verified'?'所选模块记录及实际文件完整性校验通过；可继续验证模块恢复，正式覆盖导入未开放。':task.state==='product_module_isolated_verified'?'商品与实际图片的隔离恢复验证通过，其他业务数据保留。':task.state==='verified'?'整站包隔离恢复验证通过，请查看覆盖预览。':`恢复验证未通过：${task.error||'请检查任务记录。'}`,['failed','interrupted','module_restore_conflicts_found'].includes(task.state));
    }
  }
}
function taskActionButton(container, label, action, style = 'secondary') {
  const button = document.createElement('button'); button.textContent = label; button.className = style;
  button.addEventListener('click', async () => {button.disabled = true; try {await action();await refreshMaintenanceTasks();} catch (error) {showMessage(error.message,true);} finally {button.disabled=false;}});
  container.append(button);
}
function renderBackupTasks(tasks) {
  const container = getMaintenanceConsoleElement('tasks'); container.replaceChildren();
  if (!tasks.length) {container.textContent='还没有本地备份或上传记录。';return;}
  for (const task of tasks) {
    const article = document.createElement('article'); const title = document.createElement('strong');
    const stoppedDrain=task.product_view_stopped_drain;
    const unresolvedStoppedDrain=Boolean(stoppedDrain&&stoppedDrain.phase!=='candidate_rebuilt');
    const stoppedHandoff=task.product_catalog_stopped_handoff_preparation;
    const stoppedSwitch=task.recovery_journal?.product_catalog_stopped_switch;
    const unresolvedStoppedHandoff=Boolean(stoppedHandoff&&stoppedHandoff.phase!=='cancelled'&&stoppedSwitch?.phase!=='original_released');
    title.textContent = `${task.kind === 'selected_business_modules_backup' ? '所选业务模块备份' : task.kind === 'product_catalog_backup' ? '商品与图片备份' : task.kind === 'backup' ? '整站商城备份' : '上传恢复'} · ${new Date(task.created_at).toLocaleString()}`;article.append(title);
    const selectedModules=task.manifest?.selected_modules||task.selected_backup_modules;
    if(selectedModules?.length){const paragraph=document.createElement('p');paragraph.textContent=`包含模块：${describeSelectedBusinessBackupModules(selectedModules)}。正式分模块覆盖导入尚未开放。`;article.append(paragraph);}
    if (task.product_catalog_recovery && !unresolvedStoppedDrain && !unresolvedStoppedHandoff) {
      const result=task.product_catalog_recovery;
      const paragraph=document.createElement('p');
      paragraph.textContent=`商品隔离恢复通过：恢复 ${result.restored_rows} 行，验证 ${result.verified_image_files} 个图片文件，保留 ${result.preserved_tables} 张其他业务表及目标新增记录。在线商品覆盖尚未开放。`;
      article.append(paragraph);
      const inventory=result.inventory_preservation;
      if (inventory) {
        const inventorySummary=document.createElement('p');
        inventorySummary.textContent=`库存处理：保留 ${inventory.existing_variant_stocks_preserved} 个已有 SKU 的当前库存，其中 ${inventory.different_snapshot_stock_values_not_applied} 个与备份不同；另有 ${inventory.new_variant_stocks_restored_from_snapshot} 个新增 SKU 使用备份库存。已有商品浏览数保持当前值。`;
        article.append(inventorySummary);
      } else {
        const inventorySummary=document.createElement('p');
        inventorySummary.textContent='这份旧验证记录尚未核对当前库存保留规则，请重新验证上传包。';
        article.append(inventorySummary);
      }
    }
    for (const text of [`${stateNames[task.state] || task.state} · ${productCatalogPhaseNames[task.phase] || task.phase}`,`任务 ${task.id} · ${formatBackupByteSize(task.size)} · 到期 ${new Date(task.expires_at).toLocaleString()}`,`下载请求 ${task.download_requests} 次；电脑保存情况需要自行确认。`]) {const paragraph=document.createElement('p');paragraph.textContent=text;article.append(paragraph);}
    if (task.error) {const error=document.createElement('p');error.className='error';error.textContent=task.error;article.append(error);}
    if (task.ciphertext_sha256) {const checksum=document.createElement('p');checksum.textContent=`包校验 SHA-256：${task.ciphertext_sha256}`;article.append(checksum);}
    if (hasPermission('export') && ['ready_to_download','generated_with_gaps'].includes(task.state)) taskActionButton(article,'下载到电脑',async()=>{const ticket=await postMaintenanceRequest(`/api/tasks/${task.id}/download-ticket`);const link=document.createElement('a');link.href=ticket.download_path;link.download='';document.body.append(link);link.click();link.remove();showMessage('已发起浏览器下载，请检查电脑中的保存结果。');});
    if (task.manifest && !unresolvedStoppedDrain && !unresolvedStoppedHandoff) taskActionButton(article,'查看范围与覆盖预览',()=>showOverwritePreview(task));
    if (hasPermission('restore') && !targetStatus.application_fenced && task.kind==='upload' && task.manifest?.scope==='selected_business_modules' && !task.work_copy_expired && !task.recovery_journal && new Date(task.expires_at).getTime()>Date.now() && (['module_package_integrity_verified','module_isolated_restore_verified','module_restore_conflicts_found','module_chat_candidate_rehearsal_verified'].includes(task.state)||(['failed','interrupted'].includes(task.state)&&task.selected_module_restore_preview))) taskActionButton(article,'验证模块恢复（暂停商城）',()=>startSelectedBusinessModuleIsolatedRestorePreview(task));
    if (isProductCatalogCandidateEligibleForHandoffInspection(task, targetStatus)) taskActionButton(article,'检查商品恢复准备',()=>inspectProductCatalogCandidateHandoff(task.id));
    if (stoppedDrain) { const description=document.createElement('p');description.textContent=unresolvedStoppedDrain?'浏览增量处理尚未完成，旧商品预览失效；维护隔离会保留，请检查失败原因后重试。':`浏览增量已幂等保存并重建预览，共执行 ${stoppedDrain.flush_calls} 次落库检查；这份记录不批准在线覆盖。`;article.append(description); }
    if (hasPermission('restore') && !unresolvedStoppedHandoff && task.manifest?.scope==='product_catalog_with_images' && !task.work_copy_expired && !task.recovery_journal && (isProductCatalogCandidateEligibleForHandoffInspection(task,targetStatus)||(unresolvedStoppedDrain&&['failed','interrupted'].includes(task.state)))) {
      if(targetStatus.product_view_stopped_drain_configured)taskActionButton(article,unresolvedStoppedDrain?'重试浏览增量排空并重建预览':'停写保存浏览增量并重建预览',()=>startStoppedProductViewDrainAndFreshCandidateRebuild(task.id),'danger');
      else { const description=document.createElement('p');description.textContent='浏览增量维护尚未配置：需先登记 Redis 归属，并由运维填写独立连接和本机全部写入进程清单。';article.append(description); }
    }
    if (unresolvedStoppedHandoff && !stoppedSwitch) {
      const description=document.createElement('p');description.textContent=task.state==='product_module_stopped_handoff_prepared'?'商品候选已按最新数据准备，商城继续停写。可以继续隔离切换并核对商品，也可以核验后取消准备。隔离切换后仍不能恢复新商品营业。':'商品交接准备保留了维护隔离；中断后须重新核验，不能沿用旧准备结果。原工作副本到期后仍可重试或取消，但不会延长覆盖资格。';article.append(description);
    }
    if (hasPermission('restore') && !task.recovery_journal && targetStatus.product_view_stopped_drain_configured && (isProductCatalogCandidateEligibleForHandoffInspection(task,targetStatus)||(unresolvedStoppedHandoff&&['failed','interrupted','product_module_stopped_handoff_prepared'].includes(task.state)&&stoppedHandoff.phase!=='cancelling'))) taskActionButton(article,unresolvedStoppedHandoff?'重新核验并保持停写':'保持停写准备商品交接',()=>startContinuouslyStoppedProductCatalogHandoffPreparation(task.id),'danger');
    if (hasPermission('restore') && !task.recovery_journal && unresolvedStoppedHandoff && ['ready','cancelling'].includes(stoppedHandoff.phase) && ['failed','interrupted','product_module_stopped_handoff_prepared'].includes(task.state)) taskActionButton(article,'取消交接准备并恢复当前商城',()=>cancelStoppedProductCatalogHandoffPreparationAfterVerification(task.id),'danger');
    if (isContinuouslyStoppedProductCatalogIsolatedSwitchAvailable(task,targetStatus)) taskActionButton(article,'继续隔离切换商品（仍停写）',()=>startProductCatalogIsolatedSwitchOrVerifiedPairedRollback(task.id,false),'danger');
    if (stoppedSwitch && stoppedSwitch.phase!=='original_released') { const description=document.createElement('p');description.textContent='商品和图片已进入隔离切换流程，商城继续停写。当前可以核对结果或成对回退，最终恢复营业尚未开放；中断后请核验回退至原商城。';article.append(description); }
    if (isStoppedProductCatalogPostSwitchInspectionAvailable(task,targetStatus)) taskActionButton(article,'核对本次商品切换结果',()=>inspectStoppedProductCatalogPostSwitch(task.id));
    const cacheProgress=task.product_catalog_stopped_detail_cache_invalidation;
    if(cacheProgress){const description=document.createElement('p');description.textContent=stoppedSwitch?.phase==='original_released'?'本次商品切换的缓存处理记录已保留；原商城已成对回退，详情缓存由正常读取重建。':cacheProgress.phase==='verified_absent'?(task.state==='product_module_switched_still_isolated'?`商品详情缓存与读取锁已核对为空（计划详情键 ${cacheProgress.planned_detail_cache_keys}、读取锁 ${cacheProgress.planned_read_lease_keys}）。商城继续停写；页面、媒体及营业接收仍待完成。`:'旧商品详情缓存处理记录已保留；当前仍需核验，商城继续停写，请检查任务状态和失败原因。'):'商品详情缓存处理尚未完成，已保存原处理清单；可在当前连续停写窗口核验重试，重启后请核验回退。';article.append(description);}
    if(isStoppedProductDetailCacheInvalidationAvailable(task,targetStatus))taskActionButton(article,cacheProgress?.phase==='verified_absent'?'复核商品详情缓存（仍停写）':cacheProgress?'重试商品详情缓存处理（仍停写）':'处理商品详情缓存（仍停写）',()=>startStoppedProductDetailCacheInvalidation(task.id),'danger');
    if(isStoppedProductProjectionInspectionAvailable(task,targetStatus))taskActionButton(article,'核对商品页面、媒体与边缘缓存',()=>inspectStoppedProductStorefrontProjectionBaseline(task.id));
    if(task.product_catalog_stopped_projection_evidence){const note=document.createElement('p');note.textContent='已保留本次商品的人工投影核对凭证；外部结果由操作员核验，凭证记录不批准恢复营业。';article.append(note);}
    if (hasPermission('restore') && stoppedSwitch && !['completed','rolled_back','running','queued'].includes(task.state)) taskActionButton(article,'成对回退商品并恢复原商城',()=>startProductCatalogIsolatedSwitchOrVerifiedPairedRollback(task.id,true),'danger');
    if (hasPermission('upload') && task.kind==='upload' && task.uploaded_bytes===task.size && ['uploaded','failed','interrupted'].includes(task.state) && !task.recovery_journal && !unresolvedStoppedDrain && !unresolvedStoppedHandoff) taskActionButton(article,'校验包并检查可用恢复方式',async()=>{await postMaintenanceRequest(`/api/tasks/${task.id}/verify`);pendingVerificationTaskID=task.id;showMessage('正在校验包。整站及单独商品包会演练隔离恢复，模块组合包检查记录与实际文件完整性。');});
    if (hasPermission('restore') && task.kind==='upload' && task.state==='verified' && !task.work_copy_expired) {
      for (const [module, label] of [['orders','订单'],['customer_service','客服']]) taskActionButton(article,`检查${label}单独恢复依赖`,()=>inspectModuleRecoveryDependencies(task.id,module));
    }
    if (hasPermission('restore') && task.state==='verified' && task.manifest?.scope==='whole_application') taskActionButton(article,'覆盖恢复',async()=>{showOverwritePreview(task);const phrase=`RESTORE ${targetStatus.site_id}/${targetStatus.environment}/${task.manifest.package_id}`;const confirmation=prompt(`将整体商城恢复到 ${task.manifest.captured_at}。该时间点之后的在线数据会被替换。操作期间商城停止服务，并保留成对回退数据。\n请输入：${phrase}`);if(confirmation===null)return;await postMaintenanceRequest(`/api/tasks/${task.id}/apply`,{confirmation});showMessage('正在进行成对切换，完成后仍保持隔离，等待对账。');},'danger');
    if (hasPermission('restore') && task.state==='awaiting_reconciliation') {
      taskActionButton(article,'开启维护连接进行对账',async()=>{const phrase=`RECONCILE ${targetStatus.site_id}/${targetStatus.environment}/${task.id}`;const confirmation=prompt(`商城进程保持停止。此窗口仅供授权维护人员通过数据库工具核对和修正支付、物流、持久化任务等事实。恢复服务前须关闭维护 SQL 连接。\n请输入：${phrase}`);if(confirmation===null)return;await postMaintenanceRequest(`/api/tasks/${task.id}/reconciliation-connections`,{confirmation});showMessage('已开放维护数据库连接，商城进程继续隔离。完成核对后关闭 SQL 连接并读取对账快照。');});
      taskActionButton(article,'填写对账凭证',()=>{reconciliationTask=task;inspectedReconciliationReport=null;getMaintenanceConsoleElement('complete-reconciliation').disabled=true;getMaintenanceConsoleElement('reconciliation-report').textContent='';getMaintenanceConsoleElement('reconciliation-panel').hidden=false;getMaintenanceConsoleElement('reconciliation-panel').scrollIntoView({behavior:'smooth'});});
    }
    if (hasPermission('restore') && task.recovery_journal && !stoppedSwitch && !['completed','rolled_back','running','resuming','releasing_rollback'].includes(task.state)) taskActionButton(article,'成对回退',async()=>{const phrase=`ROLLBACK ${targetStatus.site_id}/${task.id}`;const confirmation=prompt(`将数据库和附件一并回退至覆盖前的状态。\n请输入：${phrase}`);if(confirmation===null)return;await postMaintenanceRequest(`/api/tasks/${task.id}/rollback`,{confirmation});showMessage('正在成对回退，核对通过后解除隔离。');},'danger');
    if (hasPermission('restore') && targetStatus.application_fenced && !task.recovery_journal && !['running','queued'].includes(task.state) && !unresolvedStoppedDrain && !unresolvedStoppedHandoff) taskActionButton(article,stoppedDrain?'解除已重建预览的维护隔离':'解除未改数据的维护隔离',async()=>{await postMaintenanceRequest(`/api/tasks/${task.id}/release`);showMessage('已解除未发生切换的维护隔离。');});
    container.append(article);
  }
}
function showOverwritePreview(task) {
  const manifest=task.manifest;getMaintenanceConsoleElement('preview-panel').hidden=false;
  renderSelectedBusinessModuleIsolatedRestorePreview(task);
  if (manifest.scope==='selected_business_modules') {
    getMaintenanceConsoleElement('preview-description').textContent=`业务模块包：${describeSelectedBusinessBackupModules(manifest.selected_modules)} · 恢复点 ${manifest.captured_at} · ${manifest.site_id}/${manifest.environment}。先校验包，再单独执行隔离恢复验证；依赖或内容冲突会阻断。正式覆盖导入尚未开放。`;
  } else if (manifest.scope==='product_catalog_with_images') {
    getMaintenanceConsoleElement('preview-description').textContent=`商品与图片独立包 · 恢复点 ${manifest.captured_at} · ${manifest.site_id}/${manifest.environment}。上传后先在当前商城的隔离副本恢复商品；保留未选择的业务记录、目标新增商品及已有 SKU 的当前库存。新增 SKU 使用备份库存，结果会单独列出。SKU 或库存归属冲突会阻断。在线商品覆盖尚未开放。`;
  } else {
  getMaintenanceConsoleElement('preview-description').textContent=`恢复点 ${manifest.captured_at} · 包 ${manifest.package_id} · ${manifest.site_id}/${manifest.environment} · 迁移版本 ${manifest.migration_version}。目标数据库 ${targetStatus.database_name}；备份时数据库占用 ${formatBackupByteSize(manifest.database_size_bytes)}。附件目标：${Object.entries(targetStatus.storage_roots).map(([name,path])=>`${name} → ${path}`).join('；')}。覆盖将替换恢复点之后的数据。`;
  }
  const rows=getMaintenanceConsoleElement('preview-tables');rows.replaceChildren();
  const previewTableNames=manifest.scope!=='whole_application'?Object.keys(manifest.tables):Array.from(new Set([...Object.keys(task.preview_target_tables||{}),...Object.keys(manifest.tables)]));
  for(const name of previewTableNames.sort()){const row=document.createElement('tr');for(const value of [name,task.preview_target_tables?.[name]??'未预检',manifest.tables[name]??'不包含']){const cell=document.createElement('td');cell.textContent=String(value);row.append(cell);}rows.append(row);}
  getMaintenanceConsoleElement('preview-files').textContent=`媒体文件 ${manifest.media_file_count} 个 · ${formatBackupByteSize(manifest.media_file_bytes)} · 整包的加密与文件清单均须通过校验。`;
  getMaintenanceConsoleElement('preview-gaps').textContent=manifest.gaps.length?`恢复被阻断：${manifest.gaps.join('；')}`:'包内未报告附件引用缺口。隔离演练结果以任务状态为准。';getMaintenanceConsoleElement('preview-panel').scrollIntoView({behavior:'smooth'});
}
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
  const coverage=report.customer_service_overwrite_preview;
  const summary=document.createElement('p');
  const passed=['module_isolated_restore_verified','module_chat_candidate_rehearsal_verified'].includes(task.state)&&report.isolated_restore_verified&&report.target_unchanged&&report.disposable_stage_removed;
  const contentConflicts=report.tables.reduce((count,item)=>count+item.conflicting_rows,0);
  const selectedCandidate=task.selected_customer_service_candidate_rehearsal;
  const chosenCandidateVerified=task.state==='module_chat_candidate_rehearsal_verified'&&selectedCandidate?.isolated_restore_verified&&selectedCandidate.target_unchanged&&selectedCandidate.disposable_stage_removed&&selectedCandidate.saved_choice_sha256===task.selected_customer_service_restore_choice?.sha256&&selectedCandidate.online_overwrite_available===false;
  summary.textContent=chosenCandidateVerified?`原始模块预览发现 ${contentConflicts} 条内容冲突；已按保存的策略完成副本验证，详细结果见下方。当前商城未覆盖。`:passed?`隔离恢复通过：恢复 ${report.restored_rows} 条缺失记录，验证 ${report.verified_files} 个文件，核对 ${report.preserved_tables} 张表的已有记录保留；演练库已删除，当前商城未覆盖。`:`验证未通过或存在冲突：内容冲突 ${contentConflicts} 条，依赖冲突 ${report.dependency_conflicts} 项，文件冲突 ${report.file_conflicts} 项${coverage?.package_rows?`，采用备份记录时聊天规则阻断 ${coverage.package_rows.blocking_issues} 项`:''}。完整结果以本次任务状态为准。`;
  if(!passed&&!chosenCandidateVerified)summary.className='error';panel.append(summary);
  const table=document.createElement('table');
  const head=document.createElement('tr');
  for(const value of ['数据表','相同记录','保留当前运行状态','缺失记录','内容冲突','受保护缺失记录','限制']){const cell=document.createElement('th');cell.textContent=value;head.append(cell);}table.append(head);
  for(const item of report.tables){const row=document.createElement('tr');for(const value of [item.table,item.identical_rows,item.preserved_current_runtime_rows||0,item.missing_rows,item.conflicting_rows,item.protected_new_rows,item.unsupported_reason||'']){const cell=document.createElement('td');cell.textContent=String(value);row.append(cell);}table.append(row);}
  const scroll=document.createElement('div');scroll.className='scroll';scroll.append(table);panel.append(scroll);
  const dependencyReasons={visitor_browsing_context_absent_at_backup_and_target:'备份与当前均无访客浏览档案；会话身份仍按会话自身哈希与签名密钥核对',dependency_absent_at_backup_and_target:'备份与当前均缺少必需依赖',target_dependency_missing:'当前缺少备份时存在的依赖',package_dependency_missing:'备份缺少当前存在的依赖，需核对身份',dependency_content_changed:'依赖内容已变化',unexpected_target_reference:'当前发现未登记在包内的引用',expected_reference_not_reproduced:'包内依赖未能按当前引用结构重现'};
  for(const issue of report.dependency_issues||[]){const paragraph=document.createElement('p');paragraph.textContent=`${issue.severity==='blocking'?'依赖阻断':'依赖提示'}：${issue.table}（${issue.referenced_columns.join('、')}）${issue.references} 项；${dependencyReasons[issue.reason]||issue.reason}。`;if(issue.severity==='blocking')paragraph.className='error';panel.append(paragraph);}
  if(coverage){
    const explanation=document.createElement('p');
    explanation.textContent=coverage.prerequisites_satisfied?'聊天覆盖策略预览：分别核对会话版本、阅读位置、附件归属及上传人、保留事件状态。两种策略均保留当前新增会话和消息；可保存通过的策略并验证完整记录，正式覆盖尚未开放。':'聊天覆盖策略尚未检查：请先处理共享依赖、文件、受保护执行事实或写入契约冲突。';
    panel.append(explanation);
    for(const strategy of [coverage.package_rows,coverage.preserve_current_rows]){
      if(!strategy)continue;
      const label=strategy.policy==='use_package_conversation_rows'?'采用备份中的聊天记录':'保留当前存在冲突的聊天记录';
      const paragraph=document.createElement('p');
      paragraph.textContent=`${label}：${strategy.domain_checks_satisfied?'聊天规则检查通过':strategy.complete?'存在阻断':'检查未完成'}；采用包内 ${strategy.package_rows_selected} 条，保留冲突 ${strategy.conflicting_rows_preserved} 条，保留当前新增 ${strategy.target_only_rows_preserved} 条；核对事件 ${strategy.preserved_events_checked} 条、所需文件 ${strategy.required_files_checked} 个。正式覆盖未开放。`;
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
    paragraph.textContent=verified?`所选策略副本验证通过：补入 ${candidate.restored_rows} 条、替换 ${candidate.replaced_rows} 条完整聊天记录，核对 ${candidate.preserved_tables} 张表的保留数据。副本已删除，当前商城未覆盖；正式导入尚未开放。`:'所选策略副本验证未完成；不能沿用旧结果进行覆盖，请检查任务记录并重新预览。';
    if(!verified)paragraph.className='error';panel.append(paragraph);
  }
  for(const text of report.limitations){const paragraph=document.createElement('p');paragraph.textContent=text;panel.append(paragraph);}
}
function isSelectedCustomerServiceModuleStrategySaveAvailable(task,policy) {
  const report=task?.selected_module_restore_preview,coverage=report?.customer_service_overwrite_preview;
  const strategy=policy==='use_package_conversation_rows'?coverage?.package_rows:policy==='preserve_current_conflicting_conversation_rows'?coverage?.preserve_current_rows:null;
  return hasPermission('restore')&&!targetStatus.application_fenced&&task?.kind==='upload'&&task.manifest?.scope==='selected_business_modules'&&JSON.stringify(task.manifest.selected_modules)===JSON.stringify(['customer_service'])&&!task.work_copy_expired&&!task.recovery_journal&&!task.stage_database&&!task.selected_module_disposable_stage&&new Date(task.expires_at).getTime()>Date.now()&&['module_isolated_restore_verified','module_restore_conflicts_found','module_chat_candidate_rehearsal_verified'].includes(task.state)&&report?.sha256&&report.binding&&report.target_unchanged&&report.disposable_stage_removed&&report.online_overwrite_available===false&&coverage?.prerequisites_satisfied&&strategy?.complete&&strategy.domain_checks_satisfied&&strategy.blocking_issues===0&&report.dependency_conflicts===0&&report.file_conflicts===0&&report.tables.every(table=>!table.unsupported_reason&&!table.protected_new_rows&&(['tickets','ticket_messages','customer_service_inbox_states','ticket_message_media_attachments'].includes(table.table)||(!table.missing_rows&&!table.conflicting_rows)));
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
  const confirmation=prompt(`将暂停商城服务，重新核对所选策略绑定的包和当前数据，再在一次性副本写入完整聊天记录。完成后删除副本并恢复当前商城。正式覆盖尚未开放。\n请输入：${phrase}`);
  if(confirmation===null)return;
  const current=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(current!==task||!isSelectedCustomerServiceModuleStrategySaveAvailable(current,choice.policy))throw new Error('当前任务已变化，请刷新后重试。');
  const revision=maintenanceConnectionRevision;
  await postMaintenanceRequest(`/api/tasks/${taskID}/selected-customer-service-candidate-rehearsal`,{choice_sha256:choiceSHA256,confirmation});
  if(revision!==maintenanceConnectionRevision)return;
  pendingVerificationTaskID=taskID;getMaintenanceConsoleElement('preview-panel').hidden=true;
  showMessage('正在验证所选聊天策略；完成后查看副本验证结果。');
}
async function startStoppedProductViewDrainAndFreshCandidateRebuild(taskID) {
  const task=displayedMaintenanceTasks.find(task=>task.id===taskID);
  const unresolvedDrain=Boolean(task?.product_view_stopped_drain&&task.product_view_stopped_drain.phase!=='candidate_rebuilt');
  if(!hasPermission('restore')||!targetStatus.product_view_stopped_drain_configured||!task||task.recovery_journal||task.work_copy_expired||(task.product_catalog_stopped_handoff_preparation&&task.product_catalog_stopped_handoff_preparation.phase!=='cancelled')||(!unresolvedDrain&&!(new Date(task.expires_at).getTime()>Date.now())))throw new Error('商品任务或维护权限已变化，请刷新后重试。');
  const phrase=`DRAIN PRODUCT VIEWS ${targetStatus.site_id}/${targetStatus.environment}/${taskID}`;
  const inventory=(targetStatus.product_view_recovery_writer_inventory||[]).join('\n');
  const confirmation=prompt(`商城会进入维护并停止服务，将 Redis 浏览增量幂等保存到当前数据库，再重建商品隔离预览。失败时保留隔离和原批次。\n请先核对下列清单覆盖全部 API、任务、调度器、导入器和脚本，均在同一主机使用相同维护目录，且所有其他 Redis 客户端已停止；多主机写入不支持。\n${inventory}\n输入下列内容即确认已核实上述条件：\n${phrase}`);
  if(confirmation===null)return;
  const revision=maintenanceConnectionRevision;
  clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();
  getMaintenanceConsoleElement('preview-panel').hidden=true;
  await postMaintenanceRequest(`/api/tasks/${taskID}/product-view-stopped-drain-and-rebuild`,{confirmation,all_listed_writers_stopped_and_use_this_single_host_gate:true});
  if(revision!==maintenanceConnectionRevision)return;
  pendingVerificationTaskID=taskID;
  showMessage('正在取得停写租约、核对 Redis 归属并保存浏览增量，随后重建商品预览。失败时保持维护隔离。');
}
async function startContinuouslyStoppedProductCatalogHandoffPreparation(taskID) {
  const task=displayedMaintenanceTasks.find(task=>task.id===taskID);
  const preparation=task?.product_catalog_stopped_handoff_preparation;
  const retry=Boolean(preparation&&preparation.phase!=='cancelled'&&preparation.phase!=='cancelling'&&['failed','interrupted','product_module_stopped_handoff_prepared'].includes(task.state));
  if(!hasPermission('restore')||!targetStatus.product_view_stopped_drain_configured||!task||task.work_copy_expired||task.recovery_journal||(!retry&&!isProductCatalogCandidateEligibleForHandoffInspection(task,targetStatus)))throw new Error('商品任务或维护权限已变化，请刷新后重试。');
  const phrase=`PREPARE STOPPED PRODUCT HANDOFF ${targetStatus.site_id}/${targetStatus.environment}/${taskID}`;
  const inventory=(targetStatus.product_view_recovery_writer_inventory||[]).join('\n');
  const confirmation=prompt(`商城将停止服务，保存浏览增量并重建商品候选，完成后仍持续停写。正式覆盖尚未开放，可核验后取消准备并恢复当前商城。\n确认以下清单覆盖全部 API、任务、调度器、导入器和脚本，均在同一主机使用相同维护目录，且其他 Redis 客户端均已停止；多主机写入不支持。\n${inventory}\n请输入：${phrase}`);
  if(confirmation===null)return;
  const revision=maintenanceConnectionRevision;
  clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();getMaintenanceConsoleElement('preview-panel').hidden=true;
  await postMaintenanceRequest(`/api/tasks/${taskID}/product-catalog-stopped-handoff-preparation`,{confirmation,all_listed_writers_stopped_and_use_this_single_host_gate:true});
  if(revision!==maintenanceConnectionRevision)return;
  showMessage('正在准备商品交接，完成后商城仍保持停写。请在完成后选择继续核验或取消准备。');
}
async function cancelStoppedProductCatalogHandoffPreparationAfterVerification(taskID) {
  const task=displayedMaintenanceTasks.find(task=>task.id===taskID);
  if(!hasPermission('restore')||!task||task.recovery_journal||!['ready','cancelling'].includes(task.product_catalog_stopped_handoff_preparation?.phase)||!['failed','interrupted','product_module_stopped_handoff_prepared'].includes(task.state))throw new Error('交接准备状态或维护权限已变化，请刷新后重试。');
  const phrase=`CANCEL STOPPED PRODUCT HANDOFF ${targetStatus.site_id}/${targetStatus.environment}/${taskID}`;
  const confirmation=prompt(`将核对当前数据库、商品候选、完整媒体和浏览增量；通过后取消交接准备并恢复当前商城。已保存的浏览数继续保留，本操作不导入商品。\n请输入：${phrase}`);
  if(confirmation===null)return;
  const revision=maintenanceConnectionRevision;
  clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();getMaintenanceConsoleElement('preview-panel').hidden=true;
  await postMaintenanceRequest(`/api/tasks/${taskID}/product-catalog-stopped-handoff-cancellation`,{confirmation});
  if(revision!==maintenanceConnectionRevision)return;
  showMessage('正在核验后取消准备；仅在核验通过后解除当前商城隔离。');
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
function isProductCatalogCandidateEligibleForHandoffInspection(task, status) {
  return Boolean(status?.permissions?.includes('ops:backup:restore') && !status.application_fenced && task?.kind==='upload' && task.state==='product_module_isolated_verified' && task.manifest?.scope==='product_catalog_with_images' && !task.work_copy_expired && !task.recovery_journal && (!task.product_catalog_stopped_handoff_preparation||task.product_catalog_stopped_handoff_preparation.phase==='cancelled') && (!task.product_view_stopped_drain||task.product_view_stopped_drain.phase==='candidate_rebuilt') && new Date(task.expires_at).getTime()>Date.now());
}
function isContinuouslyStoppedProductCatalogIsolatedSwitchAvailable(task,status) {
  return Boolean(status?.permissions?.includes('ops:backup:restore')&&status.application_fenced&&status.product_view_stopped_drain_configured&&task?.state==='product_module_stopped_handoff_prepared'&&task.manifest?.scope==='product_catalog_with_images'&&task.product_catalog_stopped_handoff_preparation?.phase==='ready'&&!task.recovery_journal&&!task.work_copy_expired&&new Date(task.expires_at).getTime()>Date.now());
}
function isStoppedProductCatalogPostSwitchInspectionAvailable(task,status) {
  return Boolean(status?.permissions?.includes('ops:backup:restore')&&status.application_fenced&&task?.state==='product_module_switched_still_isolated'&&task.recovery_journal?.product_catalog_stopped_switch?.phase==='switched_still_isolated');
}
function isStoppedProductDetailCacheInvalidationAvailable(task,status) {
  return Boolean(status?.permissions?.includes('ops:backup:restore')&&status.application_fenced&&task?.manifest?.scope==='product_catalog_with_images'&&!task.work_copy_expired&&new Date(task.expires_at).getTime()>Date.now()&&task.recovery_journal?.product_catalog_stopped_switch?.phase==='switched_still_isolated'&&['product_module_switched_still_isolated','failed'].includes(task.state));
}
async function startStoppedProductDetailCacheInvalidation(taskID) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(!isStoppedProductDetailCacheInvalidationAvailable(task,targetStatus))throw new Error('商品缓存处理状态或维护权限已变化，请刷新后重试。');
  const credential=maintenanceCredential,connectionRevision=maintenanceConnectionRevision,taskSnapshot=JSON.stringify(task),statusSnapshot=JSON.stringify(targetStatus);
  const phrase=`INVALIDATE STOPPED PRODUCT CACHE ${targetStatus.site_id}/${targetStatus.environment}/${taskID}`;
  const confirmation=prompt(`商城继续停写。核对本次商品与图片切换、Redis 归属及其他客户端后，保存逐键清单并处理商品详情缓存和读取锁，包括遗留地址。浏览增量、客服消息和任务数据会保留。此步骤完成后仍需页面核对和营业接收。\n请输入：${phrase}`);
  if(confirmation===null)return;
  const current=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(credential!==maintenanceCredential||connectionRevision!==maintenanceConnectionRevision||JSON.stringify(current)!==taskSnapshot||JSON.stringify(targetStatus)!==statusSnapshot||!isStoppedProductDetailCacheInvalidationAvailable(current,targetStatus))throw new Error('操作期间商品状态已变化，请刷新后重试。');
  clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();
  await postMaintenanceRequest(`/api/tasks/${taskID}/product-catalog-stopped-detail-cache-invalidation`,{confirmation});
  if(credential!==maintenanceCredential||connectionRevision!==maintenanceConnectionRevision)return;
  showMessage('正在核验并处理本次商品详情缓存与读取锁；完成后商城继续停写。');
}
function clearStoppedProductCatalogPostSwitchInspection() {
  productCatalogStoppedPostSwitchInspectionRevision++;
  productCatalogStoppedPostSwitchObservedTask=null;
  getMaintenanceConsoleElement('product-catalog-stopped-post-switch-inspection-panel').hidden=true;
  getMaintenanceConsoleElement('product-catalog-stopped-post-switch-inspection-contents').replaceChildren();
  getMaintenanceConsoleElement('product-catalog-stopped-post-switch-inspection-status').textContent='';
}
function isStoppedProductProjectionInspectionAvailable(task,status) {
  return isStoppedProductCatalogPostSwitchInspectionAvailable(task,status)&&task.product_catalog_stopped_detail_cache_invalidation?.phase==='verified_absent'&&!task.work_copy_expired&&new Date(task.expires_at).getTime()>Date.now();
}
function clearStoppedProductProjectionEvidencePanel() {
  clearStoppedProductReadOnlyMediaHTTPQueryResult();
  clearStoppedProductReadOnlyHTTPQueryResult();
  stoppedProductProjectionInspectionRevision++;
  stoppedProductProjectionEvidenceContext=null;
  getMaintenanceConsoleElement('product-catalog-stopped-projection-panel').hidden=true;
  getMaintenanceConsoleElement('product-catalog-stopped-projection-status').textContent='';
  getMaintenanceConsoleElement('product-catalog-stopped-projection-baseline').textContent='';
  getMaintenanceConsoleElement('product-catalog-stopped-projection-requirements').replaceChildren();
  getMaintenanceConsoleElement('product-catalog-stopped-projection-save').disabled=true;
}
async function inspectStoppedProductStorefrontProjectionBaseline(taskID) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(!isStoppedProductProjectionInspectionAvailable(task,targetStatus))throw new Error('请先完成本次商品详情缓存核验，并保持原停写窗口。');
  clearStoppedProductProjectionEvidencePanel();
  const revision=stoppedProductProjectionInspectionRevision,credential=maintenanceCredential,taskSnapshot=JSON.stringify(task),statusSnapshot=JSON.stringify(targetStatus);
  getMaintenanceConsoleElement('product-catalog-stopped-projection-panel').hidden=false;
  getMaintenanceConsoleElement('product-catalog-stopped-projection-status').textContent='正在读取原/最终商品地址与图片基线，并重新核对数据库、媒体和缓存。';
  try{
    const {report}=await postMaintenanceRequest(`/api/tasks/${taskID}/product-catalog-stopped-projection-inspection`);
    if(revision!==stoppedProductProjectionInspectionRevision||credential!==maintenanceCredential)return;
    const current=displayedMaintenanceTasks.find(item=>item.id===taskID);
    if(!isStoppedProductProjectionInspectionAvailable(current,targetStatus)||JSON.stringify(current)!==taskSnapshot||JSON.stringify(targetStatus)!==statusSnapshot||report?.task_id!==taskID||report.scope!=='product_catalog_with_images'||report.window_id!==current.recovery_journal.product_catalog_stopped_switch.window_id||report.switch_journal_sha256!==current.recovery_journal.product_catalog_stopped_switch.sha256||report.detail_cache_progress_sha256!==current.product_catalog_stopped_detail_cache_invalidation.sha256||report.database_oid!==current.recovery_journal.new_database_oid||!report.application_remains_stopped||report.online_overwrite_available!==false||report.requirements?.length!==3||!report.contract_sha256)throw new Error('投影快照与当前商品切换或缓存记录不一致，请重新检查。');
    stoppedProductProjectionEvidenceContext={taskID,taskSnapshot,statusSnapshot,credential,report,expectedRecordSHA256:current.product_catalog_stopped_projection_evidence?.sha256||''};
    getMaintenanceConsoleElement('product-catalog-stopped-projection-baseline').textContent=JSON.stringify(report,null,2);
    renderExternalProjectionReconciliationEvidenceRequirements(report,'product-catalog-stopped-projection-requirements','三项独立人工核对记录');
    getMaintenanceConsoleElement('product-catalog-stopped-projection-save').disabled=false;
    getMaintenanceConsoleElement('stopped-product-read-only-http-query-submit').disabled=false;
    getMaintenanceConsoleElement('stopped-product-read-only-media-query-submit').disabled=false;
    getMaintenanceConsoleElement('product-catalog-stopped-projection-status').textContent='SQL 商品地址、图片与空详情缓存基线已核对。请独立完成下列 HTTP、权限和边缘/浏览器核对；保存记录后商城继续停写。';
  }catch(error){if(revision===stoppedProductProjectionInspectionRevision&&credential===maintenanceCredential){clearStoppedProductProjectionEvidencePanel();showMessage(`商品投影快照未完成：${error.message}`,true);}}
}
function isStoppedProductProjectionEvidenceContextCurrent(context) {
  const task=displayedMaintenanceTasks.find(item=>item.id===context.taskID);
  return stoppedProductProjectionEvidenceContext===context&&maintenanceCredential===context.credential&&JSON.stringify(task)===context.taskSnapshot&&JSON.stringify(targetStatus)===context.statusSnapshot&&isStoppedProductProjectionInspectionAvailable(task,targetStatus);
}
function clearStoppedProductReadOnlyHTTPQueryResult() {
  stoppedProductReadOnlyHTTPQueryRevision++;
  getMaintenanceConsoleElement('stopped-product-read-only-http-query-result').textContent='';
  getMaintenanceConsoleElement('stopped-product-read-only-http-query-status').textContent='';
  getMaintenanceConsoleElement('stopped-product-read-only-http-query-submit').disabled=true;
}
getMaintenanceConsoleElement('stopped-product-read-only-http-query-form').addEventListener('submit',async event=>{
  event.preventDefault();
  const context=stoppedProductProjectionEvidenceContext;
  if(!context||!isStoppedProductProjectionEvidenceContextCurrent(context)){clearStoppedProductProjectionEvidencePanel();showMessage('请重新读取当前商品核对基线。',true);return;}
  clearStoppedProductReadOnlyHTTPQueryResult();
  const revision=stoppedProductReadOnlyHTTPQueryRevision;
  const operation=getMaintenanceConsoleElement('stopped-product-read-only-http-query-operation').value;
  const request={contract_sha256:context.report.contract_sha256,operation,locale:getMaintenanceConsoleElement('stopped-product-read-only-http-query-locale').value.trim(),country:getMaintenanceConsoleElement('stopped-product-read-only-http-query-country').value.trim(),currency:getMaintenanceConsoleElement('stopped-product-read-only-http-query-currency').value.trim(),slug:operation==='detail'?getMaintenanceConsoleElement('stopped-product-read-only-http-query-slug').value.trim():'',keyword:operation==='search'?getMaintenanceConsoleElement('stopped-product-read-only-http-query-keyword').value.trim():''};
  getMaintenanceConsoleElement('stopped-product-read-only-http-query-status').textContent='正在只读试查并重新核对 SQL、媒体和停写边界。';
  try{
    if(operation==='product_reviews'||operation==='product_review_summary'){
      const productID=getMaintenanceConsoleElement('stopped-product-read-only-http-query-product-id').value.trim();
      if(!/^[1-9][0-9]{0,9}$/.test(productID)||Number(productID)>4294967295)throw new Error('评价查询需要填写正整数商品 ID（不超过 4294967295）。');
      request.product_id=Number(productID);
      if(operation==='product_reviews'){
        const page=getMaintenanceConsoleElement('stopped-product-read-only-http-query-page').value.trim(),pageSize=getMaintenanceConsoleElement('stopped-product-read-only-http-query-page-size').value.trim();
        if(!/^[1-9][0-9]{0,3}$/.test(page)||Number(page)>1000||!/^[1-9][0-9]?$/.test(pageSize)||Number(pageSize)>20)throw new Error('评价页码限 1–1000，每页限 1–20 条。');
        request.page=Number(page);request.page_size=Number(pageSize);
      }
    }
    const {result}=await postMaintenanceRequest(`/api/tasks/${context.taskID}/product-catalog-stopped-read-only-storefront-http-query`,request);
    if(revision!==stoppedProductReadOnlyHTTPQueryRevision)return;
    if(!isStoppedProductProjectionEvidenceContextCurrent(context)){if(stoppedProductProjectionEvidenceContext===context)clearStoppedProductProjectionEvidencePanel();return;}
    if(result?.version!==1||result.task_id!==context.taskID||result.scope!=='product_catalog_with_images'||result.window_id!==context.report.window_id||result.contract_sha256!==context.report.contract_sha256||result.database_oid!==context.report.database_oid||result.detail_cache_progress_sha256!==context.report.detail_cache_progress_sha256||result.sql_read_only!==true||result.application_remains_stopped!==true||result.online_overwrite_available!==false||result.verification!=='real_storefront_get_handlers_bound_read_only_sql_without_checkout_snapshots_or_network'||Object.keys(result.request||{}).length!==Object.keys(request).length||Object.keys(request).some(key=>result.request?.[key]!==request[key])||![200,404].includes(result.http_status)||typeof result.response_bytes_base64!=='string'||result.response_bytes_base64.length>5592408)throw new Error('试查结果与当前商品基线或查询不一致。');
    const bytes=Uint8Array.from(atob(result.response_bytes_base64),character=>character.charCodeAt(0));
    if(bytes.length>4*1024*1024)throw new Error('接口内容超过核对上限。');
    const digest=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',bytes)),byte=>byte.toString(16).padStart(2,'0')).join('');
    if(digest!==result.response_sha256)throw new Error('接口内容摘要不一致，请重新核对。');
    const decoded=JSON.parse(new TextDecoder('utf-8',{fatal:true}).decode(bytes));
    if(revision!==stoppedProductReadOnlyHTTPQueryRevision)return;
    if(!isStoppedProductProjectionEvidenceContextCurrent(context)){if(stoppedProductProjectionEvidenceContext===context)clearStoppedProductProjectionEvidencePanel();return;}
    const {response_bytes_base64,...displayResult}=result;
    getMaintenanceConsoleElement('stopped-product-read-only-http-query-result').textContent=JSON.stringify({...displayResult,response:decoded},null,2);
    getMaintenanceConsoleElement('stopped-product-read-only-http-query-status').textContent=`前台接口返回 HTTP ${result.http_status}；请对照实际内容核对。商城继续停写，页面和媒体访问验收仍待完成。`;
  }catch(error){if(revision===stoppedProductReadOnlyHTTPQueryRevision&&isStoppedProductProjectionEvidenceContextCurrent(context)){clearStoppedProductProjectionEvidencePanel();showMessage(`只读试查未完成：${error.message}`,true);}}
  finally{if(isStoppedProductProjectionEvidenceContextCurrent(context)&&revision===stoppedProductReadOnlyHTTPQueryRevision)getMaintenanceConsoleElement('stopped-product-read-only-http-query-submit').disabled=false;}
});

let stoppedProductReadOnlyMediaHTTPQueryRevision=0;
function clearStoppedProductReadOnlyMediaHTTPQueryResult() {
  stoppedProductReadOnlyMediaHTTPQueryRevision++;
  getMaintenanceConsoleElement('stopped-product-read-only-media-query-result').textContent='';
  getMaintenanceConsoleElement('stopped-product-read-only-media-query-status').textContent='';
  getMaintenanceConsoleElement('stopped-product-read-only-media-query-submit').disabled=true;
  getMaintenanceConsoleElement('stopped-product-read-only-media-query-visitor-cookie').value='';
  getMaintenanceConsoleElement('stopped-product-read-only-media-query-storefront-cookie').value='';
  getMaintenanceConsoleElement('stopped-product-read-only-media-query-admin-cookie').value='';
}
getMaintenanceConsoleElement('stopped-product-read-only-media-query-form').addEventListener('submit',async event=>{
  event.preventDefault();
  const context=stoppedProductProjectionEvidenceContext;
  if(!context||!isStoppedProductProjectionEvidenceContextCurrent(context)){clearStoppedProductProjectionEvidencePanel();showMessage('请重新读取当前商品核对基线。',true);return;}
  const operation=getMaintenanceConsoleElement('stopped-product-read-only-media-query-operation').value;
  const cookieInput=getMaintenanceConsoleElement('stopped-product-read-only-media-query-visitor-cookie');
  const storefrontCookieInput=getMaintenanceConsoleElement('stopped-product-read-only-media-query-storefront-cookie');
  const adminCookieInput=getMaintenanceConsoleElement('stopped-product-read-only-media-query-admin-cookie');
  const isAttachment=operation==='conversation_attachment'||operation==='agent_conversation_attachment';
  const request={contract_sha256:context.report.contract_sha256,operation,object_key:operation==='public_upload'?getMaintenanceConsoleElement('stopped-product-read-only-media-query-object-key').value.trim():'',conversation_id:isAttachment?getMaintenanceConsoleElement('stopped-product-read-only-media-query-conversation').value.trim():'',message_id:isAttachment?getMaintenanceConsoleElement('stopped-product-read-only-media-query-message').value.trim():'',attachment_index:isAttachment?Number(getMaintenanceConsoleElement('stopped-product-read-only-media-query-index').value):0,preset:isAttachment?getMaintenanceConsoleElement('stopped-product-read-only-media-query-preset').value.trim():'',range:getMaintenanceConsoleElement('stopped-product-read-only-media-query-range').value.trim(),visitor_cookie:operation==='conversation_attachment'?cookieInput.value:'',storefront_auth_cookie:operation==='conversation_attachment'?storefrontCookieInput.value:'',admin_auth_cookie:operation==='agent_conversation_attachment'?adminCookieInput.value:''};
  const visitorCredentialSupplied=!!request.visitor_cookie;
  const storefrontCredentialSupplied=!!request.storefront_auth_cookie;
  const adminCredentialSupplied=!!request.admin_auth_cookie;
  clearStoppedProductReadOnlyMediaHTTPQueryResult();
  const revision=stoppedProductReadOnlyMediaHTTPQueryRevision;
  getMaintenanceConsoleElement('stopped-product-read-only-media-query-status').textContent='正在核对文件响应与访问权限，商城继续停写。';
  try{
    const responsePromise=postMaintenanceRequest('/api/tasks/'+context.taskID+'/product-catalog-stopped-read-only-media-http-query',request);
    request.visitor_cookie='';
    request.storefront_auth_cookie='';
    request.admin_auth_cookie='';
    const {result}=await responsePromise;
    if(revision!==stoppedProductReadOnlyMediaHTTPQueryRevision)return;
    if(!isStoppedProductProjectionEvidenceContextCurrent(context)){if(stoppedProductProjectionEvidenceContext===context)clearStoppedProductProjectionEvidencePanel();return;}
    if(result?.version!==2||result.task_id!==context.taskID||result.scope!=='product_catalog_with_images'||result.window_id!==context.report.window_id||result.contract_sha256!==context.report.contract_sha256||result.database_oid!==context.report.database_oid||result.detail_cache_progress_sha256!==context.report.detail_cache_progress_sha256||result.sql_read_only!==true||result.application_remains_stopped!==true||result.online_overwrite_available!==false||result.verification!=='in_process_business_local_media_http_handlers_active_account_role_assignment_read_only_no_network'||result.visitor_credential_supplied!==visitorCredentialSupplied||result.storefront_credential_supplied!==storefrontCredentialSupplied||result.admin_credential_supplied!==adminCredentialSupplied||Object.keys(request).some(key=>result.target?.[key]!==request[key]))throw new Error('媒体试查结果与当前基线或文件目标不一致。');
    getMaintenanceConsoleElement('stopped-product-read-only-media-query-result').textContent=JSON.stringify(result,null,2);
    getMaintenanceConsoleElement('stopped-product-read-only-media-query-status').textContent='文件接口返回 HTTP '+result.http_status+'，读取 '+result.response_bytes+' 字节。请核对摘要和权限结果；真实域名、CDN 和浏览器仍待核验。';
  }catch(error){if(revision===stoppedProductReadOnlyMediaHTTPQueryRevision&&isStoppedProductProjectionEvidenceContextCurrent(context)){clearStoppedProductProjectionEvidencePanel();showMessage('媒体只读试查未完成：'+error.message,true);}}
  finally{request.visitor_cookie='';request.storefront_auth_cookie='';request.admin_auth_cookie='';cookieInput.value='';storefrontCookieInput.value='';adminCookieInput.value='';if(isStoppedProductProjectionEvidenceContextCurrent(context)&&revision===stoppedProductReadOnlyMediaHTTPQueryRevision)getMaintenanceConsoleElement('stopped-product-read-only-media-query-submit').disabled=false;}
});

getMaintenanceConsoleElement('product-catalog-stopped-projection-form').addEventListener('submit',async event=>{
  event.preventDefault();
  const context=stoppedProductProjectionEvidenceContext;if(!context)return;
  const save=getMaintenanceConsoleElement('product-catalog-stopped-projection-save');save.disabled=true;
  try{
    const evidence=await collectExternalProjectionReconciliationOperatorAttestations(context.report,'product-catalog-stopped-projection-requirements');
    if(!isStoppedProductProjectionEvidenceContextCurrent(context))throw new Error('商品投影快照已变化，请重新读取后记录。');
    const phrase=`RECORD STOPPED PRODUCT PROJECTIONS ${targetStatus.site_id}/${targetStatus.environment}/${context.taskID}`;
    const confirmation=prompt(`将记录三项独立人工核对凭证。系统复核其对应的商品快照，外部结果仍由操作员核验；商城继续停写。\n请输入：${phrase}`);
    if(confirmation===null)return;
    if(!isStoppedProductProjectionEvidenceContextCurrent(context))throw new Error('商品投影快照已变化，请重新读取后记录。');
    const {record}=await postMaintenanceRequest(`/api/tasks/${context.taskID}/product-catalog-stopped-projection-evidence`,{confirmation,...evidence,expected_record_sha256:context.expectedRecordSHA256});
    if(!isStoppedProductProjectionEvidenceContextCurrent(context))return;
    if(record?.task_id!==context.taskID||record.scope!=='product_catalog_with_images'||record.contract_sha256!==context.report.contract_sha256||record.detail_cache_progress_sha256!==context.report.detail_cache_progress_sha256||record.verification!=='operator_attestation_only_external_results_not_automatically_verified')throw new Error('保存结果与当前商品投影快照不一致。');
    clearStoppedProductProjectionEvidencePanel();
    showMessage('已保存三项商品人工核对凭证，商城继续停写；此记录不能恢复营业。');
    await refreshMaintenanceTasks();
  }catch(error){if(stoppedProductProjectionEvidenceContext===context){clearStoppedProductProjectionEvidencePanel();showMessage(error.message,true);}}
  finally{if(stoppedProductProjectionEvidenceContext===context)save.disabled=false;}
});
async function startProductCatalogIsolatedSwitchOrVerifiedPairedRollback(taskID,rollback) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(!hasPermission('restore')||!task||(rollback?(!task.recovery_journal?.product_catalog_stopped_switch||['running','queued','completed','rolled_back'].includes(task.state)):!isContinuouslyStoppedProductCatalogIsolatedSwitchAvailable(task,targetStatus)))throw new Error('商品切换状态或维护权限已变化，请刷新后重试。');
  const credential=maintenanceCredential,connectionRevision=maintenanceConnectionRevision,taskSnapshot=JSON.stringify(task),statusSnapshot=JSON.stringify(targetStatus);
  const phrase=`${rollback?'ROLLBACK':'SWITCH'} STOPPED PRODUCT CATALOG ${targetStatus.site_id}/${targetStatus.environment}/${taskID}`;
  const confirmation=prompt(`${rollback?'将数据库和完整媒体一并回退；核对通过后恢复原商城服务。':'将商品数据库和完整媒体一起切换到候选，商城仍持续停写。当前最终营业接收尚未开放；可以核对结果或成对回退。'}\n请输入：${phrase}`);
  if(confirmation===null)return;
  const current=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(credential!==maintenanceCredential||connectionRevision!==maintenanceConnectionRevision||JSON.stringify(current)!==taskSnapshot||JSON.stringify(targetStatus)!==statusSnapshot||(!rollback&&!isContinuouslyStoppedProductCatalogIsolatedSwitchAvailable(current,targetStatus)))throw new Error('操作期间商品状态已变化，请刷新后重试。');
  clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();
  await postMaintenanceRequest(`/api/tasks/${taskID}/${rollback?'product-catalog-stopped-switch-rollback':'product-catalog-stopped-isolated-switch'}`,{confirmation});
  if(credential!==maintenanceCredential||connectionRevision!==maintenanceConnectionRevision)return;
  showMessage(rollback?'正在成对回退商品；核对原数据库、完整媒体和浏览增量后恢复原商城。':'正在隔离切换商品，完成后商城仍保持停写。');
}
async function inspectStoppedProductCatalogPostSwitch(taskID) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(!isStoppedProductCatalogPostSwitchInspectionAvailable(task,targetStatus))throw new Error('当前商品切换状态不可核对，请刷新。');
  clearStoppedProductCatalogPostSwitchInspection();
  const revision=productCatalogStoppedPostSwitchInspectionRevision,credential=maintenanceCredential;
  productCatalogStoppedPostSwitchObservedTask={taskID,taskSnapshot:JSON.stringify(task),statusSnapshot:JSON.stringify(targetStatus)};
  const panel=getMaintenanceConsoleElement('product-catalog-stopped-post-switch-inspection-panel'),status=getMaintenanceConsoleElement('product-catalog-stopped-post-switch-inspection-status'),contents=getMaintenanceConsoleElement('product-catalog-stopped-post-switch-inspection-contents');
  panel.hidden=false;status.textContent='正在核对本次切换的数据库、完整媒体与浏览增量。';
  try {
    const {report}=await postMaintenanceRequest(`/api/tasks/${taskID}/product-catalog-stopped-post-switch-inspection`);
    if(revision!==productCatalogStoppedPostSwitchInspectionRevision||credential!==maintenanceCredential)return;
    const current=displayedMaintenanceTasks.find(item=>item.id===taskID),journal=current?.recovery_journal?.product_catalog_stopped_switch;
    if(!isStoppedProductCatalogPostSwitchInspectionAvailable(current,targetStatus)||JSON.stringify(current)!==productCatalogStoppedPostSwitchObservedTask.taskSnapshot||report?.task_id!==taskID||report.scope!=='product_catalog_with_images'||report.package_id!==current.manifest?.package_id||report.window_id!==journal.window_id||report.switch_journal_sha256!==journal.sha256||report.database_oid!==current.recovery_journal.new_database_oid||!report.application_remains_stopped||!report.candidate_remains_isolated||report.online_overwrite_available!==false||report.external_actions_executed!==false)throw new Error('核对结果与当前商品切换不一致，请重新检查。');
    status.textContent='数据库、完整媒体和空浏览缓冲已核验；商城继续停写。详情缓存处理记录另见任务，页面访问等核对尚未完成，此记录不能恢复营业。';
    for(const value of [`任务 ${report.task_id} · 停写窗口 ${report.window_id}`,`商品切换核对摘要：${report.contract_sha256}`,...(report.projection_requirements||[]).map(requirement=>requirement.description)]){const paragraph=document.createElement('p');paragraph.textContent=value;contents.append(paragraph);}
  } catch(error) {
    if(revision!==productCatalogStoppedPostSwitchInspectionRevision||credential!==maintenanceCredential)return;
    contents.replaceChildren();status.textContent=`核对未完成：${error.message}`;
  }
}
function clearProductCatalogHandoffInspectionForChangedSessionOrCandidate() {
  clearStoppedProductProjectionEvidencePanel();
  clearStoppedProductCatalogPostSwitchInspection();
  productCatalogHandoffInspectionRequestRevision++;
  productCatalogHandoffInspectionObservedTask=null;
  getMaintenanceConsoleElement('product-catalog-handoff-inspection-panel').hidden=true;
  getMaintenanceConsoleElement('product-catalog-handoff-inspection-contents').replaceChildren();
  getMaintenanceConsoleElement('product-catalog-handoff-inspection-status').textContent='';
}
async function inspectProductCatalogCandidateHandoff(taskID) {
  const task=displayedMaintenanceTasks.find(task=>task.id===taskID);
  if(!isProductCatalogCandidateEligibleForHandoffInspection(task,targetStatus)){clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();showMessage('商品候选或维护会话已变化，请刷新后重新检查。',true);return;}
  const revision=++productCatalogHandoffInspectionRequestRevision;
  const credential=maintenanceCredential;
  productCatalogHandoffInspectionObservedTask={taskID,taskSnapshot:JSON.stringify(task),statusSnapshot:JSON.stringify(targetStatus)};
  const panel=getMaintenanceConsoleElement('product-catalog-handoff-inspection-panel');
  const contents=getMaintenanceConsoleElement('product-catalog-handoff-inspection-contents');
  panel.hidden=false;contents.replaceChildren();
  getMaintenanceConsoleElement('product-catalog-handoff-inspection-status').textContent='正在重新检查商品候选与当前商城，请稍候。';
  try {
    const result=await postMaintenanceRequest(`/api/tasks/${taskID}/product-catalog-handoff-inspection`);
    if(revision!==productCatalogHandoffInspectionRequestRevision||credential!==maintenanceCredential)return;
    if(!isProductCatalogCandidateEligibleForHandoffInspection(task,targetStatus)||result.report?.task_id!==taskID||result.report?.package_id!==task.manifest.package_id||result.report?.scope!=='product_catalog_with_images'||result.report?.expires_at!==task.expires_at)throw new Error('检查记录与当前商品任务不符，请重新检查。');
    renderProductCatalogCandidateHandoffInspection(result.report);
    panel.scrollIntoView({behavior:'smooth'});
  } catch(error) {
    if(revision!==productCatalogHandoffInspectionRequestRevision||credential!==maintenanceCredential)return;
    contents.replaceChildren();
    getMaintenanceConsoleElement('product-catalog-handoff-inspection-status').textContent=`检查未完成：${error.message}`;
  }
}
function renderProductCatalogCandidateHandoffInspection(report) {
  const contents=getMaintenanceConsoleElement('product-catalog-handoff-inspection-contents');contents.replaceChildren();
  const status=getMaintenanceConsoleElement('product-catalog-handoff-inspection-status');
  if(!report||report.version!==1||report.scope!=='product_catalog_with_images'||report.online_overwrite_available!==false||report.custody_transfer_available!==false){status.textContent='检查记录格式不受支持，请更新维护工具后重新检查。';return;}
  if(!(new Date(report.expires_at).getTime()>Date.now())){status.textContent='商品候选已过期或有效期无效，请重新验证上传包。';return;}
  status.textContent='已核对当前商品候选。本次只列出恢复准备要求，在线覆盖尚未开放。';
  const appendParagraph=(text)=>{const paragraph=document.createElement('p');paragraph.textContent=text;contents.append(paragraph);};
  appendParagraph(`任务 ${report.task_id} · 包 ${report.package_id} · 检查时间 ${new Date(report.captured_at).toLocaleString()}`);
  appendParagraph(`检查记录 ${report.contract_sha256} · 保留 ${report.preserved_tables} 张其他业务表。正式接收前仍须停写重验，切换后重新读取核对快照。`);
  const addList=(title,items)=>{const heading=document.createElement('h3');heading.textContent=title;contents.append(heading);const list=document.createElement('ul');for(const text of items){const row=document.createElement('li');row.textContent=text;list.append(row);}contents.append(list);};
  addList('候选保管与退出',report.custody_rules||[]);
  addList('恢复前后分别核对',(report.projection_requirements||[]).map(requirement=>requirement.description));
  const blockerLabels={production_product_candidate_ownership_handoff_not_implemented:'正式接收商品候选的流程尚未完成。',product_post_switch_external_projection_reconciliation_not_executed:'尚未取得切换后针对本次商品结果的核对记录。',production_product_restore_eligibility_and_final_acceptance_not_implemented:'生产恢复资格和最终接收验收尚未完成。',product_projection_configuration_unknown_reregister_current_business_build:'当前业务构建尚未完整登记商品缓存和浏览数配置，请更新登记后重新检查。',buffered_product_view_delta_drain_and_inflight_idempotency_contract_not_implemented:'Redis 中的浏览增量和在途落库还需要单独处理，当前不能证明不会丢失或重复累计。'};
  blockerLabels.buffered_product_view_delta_stopped_handoff_and_target_verification_not_implemented='浏览数已有批次幂等保护；仍需在目标环境证明 Redis 归属、停止全部写入、核对并排空增量后重建候选，当前不能接收。';
  addList('仍未满足的条件',(report.blockers||[]).map(code=>blockerLabels[code]||'存在当前维护工具尚未识别的阻断条件，请核对完整记录。'));
}
function showOrderTransactionRowInspection(inspection) {
  const panel=getMaintenanceConsoleElement('order-transaction-row-inspection');panel.hidden=!inspection;
  if(!inspection)return;
  getMaintenanceConsoleElement('order-transaction-row-description').textContent=`${inspection.complete?'已完成本轮订单范围检查':'订单检查不完整，请先处理限制或缺失信息'}。包内引用行 ${inspection.source_selected_rows} 条，目标引用行 ${inspection.target_selected_rows} 条，阻断问题 ${inspection.blocking_issues} 项。每侧最多观察 ${inspection.row_limit_per_side} 个唯一对象、${formatBackupByteSize(inspection.projection_byte_limit_per_side)} 引用字段；订单号、地址、付款响应和正文不会写入报告。`;
  showModuleRecoveryDurableEventInspection('order-transaction-durable-event', inspection.durable_events);
  showModuleRecoverySideEffectReceiptInspection(inspection.side_effect_receipts);
  const files=inspection.files;
  getMaintenanceConsoleElement('order-transaction-file-description').textContent=`包内证据与售后文件已校验 ${files.package_verified_files}/${files.package_referenced_files} 个；目标同内容 ${files.target_matching_files} 个、缺失 ${files.target_missing_files} 个、文件冲突 ${files.target_different_files} 个。目标同时检查包内需要的路径。`;
  const rows=getMaintenanceConsoleElement('order-transaction-row-tables');rows.replaceChildren();
  for(const table of inspection.tables){const row=document.createElement('tr');for(const value of [table.table,`${table.source_scoped}/${table.source_selected}`,`${table.target_scoped}/${table.target_selected}`,table.matching,table.target_missing,table.content_differs,table.identity_conflicts,table.target_only_scoped+(table.target_only_shared||0)]){const cell=document.createElement('td');cell.textContent=String(value);row.append(cell);}rows.append(row);}
  const counts=getMaintenanceConsoleElement('order-transaction-row-issue-counts');counts.replaceChildren();
  for(const [code,count] of Object.entries(inspection.issue_counts)){const item=document.createElement('li');item.textContent=`${code}：${count} 项`;counts.append(item);}
  const issues=getMaintenanceConsoleElement('order-transaction-row-issues');issues.replaceChildren();
  for(const issue of inspection.issues){const item=document.createElement('li');item.textContent=`${issue.severity==='blocking'?'需处理':'需核对'} · ${issue.side==='package'?'包内':issue.side==='target'?'目标':'双方'} · ${issue.table||'范围检查'}${issue.row_identity?` #${issue.row_identity}`:''}：${issue.message}`;issues.append(item);}
  getMaintenanceConsoleElement('order-transaction-row-sample-description').textContent=inspection.issues_truncated?`问题样例仅展示前 ${inspection.issue_sample_limit} 项；分类计数包含本轮已发现的全部问题。`:'已展示本轮发现的问题；单模块上传覆盖尚未开放。';
}
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
function showModuleRecoverySideEffectReceiptInspection(inspection) {
  const panel=getMaintenanceConsoleElement('module-recovery-side-effect-receipt-inspection');panel.hidden=!inspection;
  if(!inspection)return;
  getMaintenanceConsoleElement('module-recovery-side-effect-receipt-description').textContent=`${inspection.complete?'已完成本轮凭据检查':'凭据检查不完整，请先处理缺失结构或容量上限'}。按自然幂等身份比较 ${inspection.compared_identities} 项；目标中全部凭据，包括其他客户及未绑定退款的记录，均须保留。`;
  getMaintenanceConsoleElement('module-recovery-side-effect-receipt-audit-description').textContent=`缺少邮件审计的通知事件：包内 ${inspection.source_notification_events_without_audit} 条，目标 ${inspection.target_notification_events_without_audit} 条。邮件审计为尽力写入，缺行不能证明未发送；已发送记录也不能代替事件消费状态。`;
  const tableNames={email_delivery_records:'邮件发送审计',payment_operation_idempotencies:'支付操作幂等',refund_idempotencies:'退款请求幂等'};
  const tables=getMaintenanceConsoleElement('module-recovery-side-effect-receipt-tables');tables.replaceChildren();
  for(const table of inspection.tables){const row=document.createElement('tr');for(const value of [tableNames[table.table]||table.table,table.source_scanned,table.target_scanned,table.matching,table.source_only,table.target_only,table.conflicts,table.blocking_issues]){const cell=document.createElement('td');cell.textContent=String(value);row.append(cell);}tables.append(row);}
  const results={same_record_preserved:'相同，保留',target_only_preserved:'目标新增，保留',source_only_requires_reconciliation:'目标缺失，须对账',natural_key_identity_conflict:'同键请求或业务身份冲突',receipt_identifier_moved:'凭据编号不同',target_state_preserved_requires_reconciliation:'当前内容变化，保留并对账'};
  const comparisons=getMaintenanceConsoleElement('module-recovery-side-effect-receipt-comparisons');comparisons.replaceChildren();
  for(const comparison of inspection.comparisons){const row=document.createElement('tr');for(const value of [tableNames[comparison.table]||comparison.table,comparison.identity_sha256,comparison.source_receipt_id||'-',comparison.target_receipt_id||'-',results[comparison.result]||'未登记诊断结果']){const cell=document.createElement('td');cell.textContent=String(value);row.append(cell);}comparisons.append(row);}
  getMaintenanceConsoleElement('module-recovery-side-effect-receipt-samples').textContent=inspection.comparisons_truncated?`仅展示前 ${inspection.comparison_sample_limit} 项凭据对比；全部已发现问题计入下方分类。`:'已展示本轮凭据对比；检查不会导入、覆盖、重置凭据或重发通知。';
}
function showCustomerServiceDurableEventInspection(inspection) { showModuleRecoveryDurableEventInspection('customer-service-durable-event', inspection); }
function showModuleRecoveryDurableEventInspection(elementPrefix, inspection) {
  const panel=getMaintenanceConsoleElement(`${elementPrefix}-inspection`);panel.hidden=!inspection;
  if(!inspection)return;
  getMaintenanceConsoleElement(`${elementPrefix}-description`).textContent=`${inspection.complete?'已完成本轮事件诊断':'范围检查不完整，请先处理缺失信息或容量上限'}。包内扫描 ${inspection.source_scanned_events} 条、相关 ${inspection.source_related_events} 条；目标扫描 ${inspection.target_scanned_events} 条、相关 ${inspection.target_related_events} 条。按幂等身份比较 ${inspection.compared_event_identities} 项；这里只显示幂等键摘要。`;
  const statusLabels={pending:'待处理',processing:'执行中',processed:'已处理',failed:'失败',unknown:'结果未知',dead_letter:'停止重试',ignored:'已忽略',invalid:'未登记状态'};
  getMaintenanceConsoleElement(`${elementPrefix}-statuses`).textContent=`包内状态：${Object.entries(inspection.source_statuses).map(([status,count])=>`${statusLabels[status]||'未登记状态'} ${count}`).join('，')||'无相关事件'}；目标状态：${Object.entries(inspection.target_statuses).map(([status,count])=>`${statusLabels[status]||'未登记状态'} ${count}`).join('，')||'无相关事件'}。`;
  const resultLabels={same_record_preserved:'记录相同，保留状态',target_only_preserved:'目标新增，保留',source_only_requires_reconciliation:'目标缺失，须对账',idempotency_identity_conflict:'幂等身份冲突',event_identifier_moved:'事件编号不同',current_state_must_be_preserved:'状态或内容变化，保留当前记录'};
  const rows=getMaintenanceConsoleElement(`${elementPrefix}-tables`);rows.replaceChildren();
  for(const comparison of inspection.comparisons){const row=document.createElement('tr');for(const value of [comparison.event_key_sha256,comparison.source_event_id||'无',comparison.target_event_id||'无',statusLabels[comparison.source_status]||'无',statusLabels[comparison.target_status]||'无',resultLabels[comparison.result]||'须核对']){const cell=document.createElement('td');cell.textContent=value;row.append(cell);}rows.append(row);}
  getMaintenanceConsoleElement(`${elementPrefix}-samples`).textContent=inspection.comparisons_truncated?`仅展示前 ${inspection.comparison_sample_limit} 项比较；总数与问题分类计数见报告。`:'已展示本轮全部幂等身份比较；异常对象见下方问题明细。';
}
async function uploadLocalBackupFile() {
  const file=getMaintenanceConsoleElement('upload-file').files[0];if(!file)return;
  if(file.size>targetStatus.max_package_bytes)throw new Error('文件超过已配置的备份包上限。');
  uploading=true;getMaintenanceConsoleElement('upload-button').disabled=true;
  try{
    const sample=new Uint8Array(await new Blob([file.slice(0,65536),file.slice(Math.max(0,file.size-65536))]).arrayBuffer());
    const fingerprint=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',sample)),byte=>byte.toString(16).padStart(2,'0')).join('');
    const storageKey=`local-backup-upload:${targetStatus.site_id}:${targetStatus.environment}:${file.size}:${fingerprint}`;
    let task=null;const previousID=localStorage.getItem(storageKey);
    if(previousID){try{const previous=await maintenanceRequest(`/api/tasks/${previousID}`);if(previous.task.state==='uploading')task=previous.task;}catch{ /* Missing or expired work copies start a new upload. */ }}
    if(!task){const result=await postMaintenanceRequest('/api/uploads',{size:file.size});task=result.task;localStorage.setItem(storageKey,task.id);}
    let offset=task.uploaded_bytes;
    while(offset<file.size){const chunk=await file.slice(offset,Math.min(offset+targetStatus.chunk_bytes,file.size)).arrayBuffer();const digest=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',chunk)),byte=>byte.toString(16).padStart(2,'0')).join('');const result=await maintenanceRequest(`/api/uploads/${task.id}/chunks`,{method:'PUT',headers:{'Upload-Offset':String(offset),'X-Chunk-SHA256':digest,'Content-Type':'application/octet-stream'},body:chunk});offset=result.task.uploaded_bytes;getMaintenanceConsoleElement('upload-progress').value=offset/file.size*100;getMaintenanceConsoleElement('upload-status').textContent=`已上传 ${formatBackupByteSize(offset)} / ${formatBackupByteSize(file.size)}；任务 ${task.id}`;}
    await postMaintenanceRequest(`/api/uploads/${task.id}/commit`);localStorage.removeItem(storageKey);showMessage('本地包已上传。请执行“校验包并检查可用恢复方式”，查看范围与校验结果。');
  }finally{uploading=false;getMaintenanceConsoleElement('upload-button').disabled=false;await refreshMaintenanceTasks();}
}
getMaintenanceConsoleElement('login-form').addEventListener('submit',async(event)=>{event.preventDefault();try{await connectMaintenanceConsole(getMaintenanceConsoleElement('credential').value);showMessage('已连接独立维护入口。');}catch(error){maintenanceCredential='';showMessage(error.message,true);}});
getMaintenanceConsoleElement('disconnect').addEventListener('click',()=>{maintenanceCredential='';maintenanceConnectionRevision++;displayedMaintenanceTasks=[];targetStatus=null;invalidateSelectedBusinessBackupPreview();clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();clearInterval(refreshTimer);getMaintenanceConsoleElement('connected-panel').hidden=true;getMaintenanceConsoleElement('login-panel').hidden=false;showMessage('维护会话已断开，正在执行的服务器任务和维护隔离仍会保留。');});
getMaintenanceConsoleElement('refresh').addEventListener('click',()=>refreshMaintenanceTasks().catch((error)=>showMessage(error.message,true)));
getMaintenanceConsoleElement('customer-service-conflict-review-form').addEventListener('submit',event=>{event.preventDefault();saveDisplayedCustomerServiceConflictDecisions().catch(error=>showMessage(error.message,true));});
getMaintenanceConsoleElement('customer-service-effective-selection-inspect').addEventListener('click',()=>inspectDisplayedCustomerServiceEffectiveSelectionDependencies().catch(error=>showMessage(error.message,true)));
for(const [id,direction] of [['previous',-1],['next',1]])getMaintenanceConsoleElement(`customer-service-conflict-review-${id}`).addEventListener('click',()=>{const context=customerServiceConflictReviewContext;if(context&&!context.saving){context.page+=direction;renderCustomerServiceConflictDecisionPage();}});
getMaintenanceConsoleElement('customer-service-conflict-review-reinspect').addEventListener('click',()=>{const context=customerServiceConflictReviewContext;if(context&&!context.saving)inspectModuleRecoveryDependencies(context.plan.binding.task_id,'customer_service').catch(error=>showMessage(error.message,true));});
getMaintenanceConsoleElement('generate').addEventListener('click',()=>generateBackupForCurrentBusinessModuleSelection().catch(error=>showMessage(error.message,true)));
for(const input of document.querySelectorAll('input[name="backup-export-mode"]'))input.addEventListener('change',invalidateSelectedBusinessBackupPreview);
for(const [id,checked] of [['select-all-business-backup-modules',true],['clear-business-backup-modules',false]])getMaintenanceConsoleElement(id).addEventListener('click',()=>{for(const input of document.querySelectorAll('#business-backup-module-options input'))input.checked=checked;invalidateSelectedBusinessBackupPreview();});
getMaintenanceConsoleElement('preview-selected-business-modules').addEventListener('click',()=>previewCurrentSelectedBusinessBackupModules().catch(error=>showMessage(error.message,true)));
getMaintenanceConsoleElement('upload-form').addEventListener('submit',(event)=>{event.preventDefault();uploadLocalBackupFile().catch((error)=>showMessage(error.message,true));});
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
getMaintenanceConsoleElement('inspect-reconciliation').addEventListener('click',async()=>{
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
});
getMaintenanceConsoleElement('reconciliation-form').addEventListener('submit',async(event)=>{
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
});
window.addEventListener('message',async(event)=>{if(event.origin!==location.origin||event.source!==window.parent||event.data?.type!=='local-backup-maintenance-session'||typeof event.data.token!=='string')return;try{await connectMaintenanceConsole(event.data.token);showMessage('已通过后台权限连接维护入口。');window.parent.postMessage({type:'local-backup-maintenance-connected'},location.origin);}catch(error){maintenanceCredential='';showMessage(error.message,true);window.parent.postMessage({type:'local-backup-maintenance-connection-failed'},location.origin);}});
if(window.parent!==window)window.parent.postMessage({type:'local-backup-maintenance-ready'},location.origin);
