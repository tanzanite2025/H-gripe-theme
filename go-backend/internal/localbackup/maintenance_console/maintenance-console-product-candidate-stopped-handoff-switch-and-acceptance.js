'use strict';
// product candidate stopped handoff switch and acceptance. Loaded in index.html order; session credentials stay in memory.
let productCatalogHandoffInspectionRequestRevision = 0;
let productCatalogHandoffInspectionObservedTask = null;
let productCatalogStoppedPostSwitchInspectionRevision = 0;
let productCatalogStoppedPostSwitchObservedTask = null;
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
function isProductCatalogCandidateEligibleForHandoffInspection(task, status) {
  return Boolean(status?.permissions?.includes('ops:backup:restore') && !status.application_fenced && task?.kind==='upload' && task.state==='product_module_isolated_verified' && task.manifest?.scope==='product_catalog_with_images' && !task.work_copy_expired && !task.recovery_journal && (!task.product_catalog_stopped_handoff_preparation||task.product_catalog_stopped_handoff_preparation.phase==='cancelled') && (!task.product_view_stopped_drain||task.product_view_stopped_drain.phase==='candidate_rebuilt') && new Date(task.expires_at).getTime()>Date.now());
}
function isContinuouslyStoppedProductCatalogIsolatedSwitchAvailable(task,status) {
  return Boolean(status?.permissions?.includes('ops:backup:restore')&&status.application_fenced&&status.product_view_stopped_drain_configured&&task?.state==='product_module_stopped_handoff_prepared'&&task.manifest?.scope==='product_catalog_with_images'&&task.product_catalog_stopped_handoff_preparation?.phase==='ready'&&!task.recovery_journal&&!task.work_copy_expired&&new Date(task.expires_at).getTime()>Date.now());
}
function isProductCatalogFormalRestoreAcceptanceAvailable(task,status) {
  return Boolean(status?.permissions?.includes('ops:backup:restore')&&status.application_fenced&&task?.kind==='upload'&&task.manifest?.scope==='product_catalog_with_images'&&task.state==='product_module_switched_still_isolated'&&task.recovery_journal?.product_catalog_stopped_switch?.phase==='switched_still_isolated'&&task.product_catalog_stopped_detail_cache_invalidation?.phase==='verified_absent'&&task.product_catalog_stopped_projection_evidence?.sha256&&!task.work_copy_expired&&new Date(task.expires_at).getTime()>Date.now());
}
async function acceptProductCatalogRestoreAndResumeService(taskID) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(!isProductCatalogFormalRestoreAcceptanceAvailable(task,targetStatus))throw new Error('当前商品恢复尚未完成核对，或权限和任务已变化，请刷新。');
  const taskSnapshot=JSON.stringify(task),statusSnapshot=JSON.stringify(targetStatus),credential=maintenanceCredential,revision=maintenanceConnectionRevision;
  const phrase=`ACCEPT PRODUCT RESTORE ${targetStatus.site_id}/${targetStatus.environment}/${taskID}`;
  const confirmation=prompt(`将正式接收本次商品及图片恢复，并恢复商城服务。请确认已实际核对商品详情、图片及页面缓存；库存和未选择的订单、客户、聊天按预览保留。恢复服务后不能再用旧任务回退。\n请输入：${phrase}`);
  if(confirmation===null)return;
  const current=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision||JSON.stringify(current)!==taskSnapshot||JSON.stringify(targetStatus)!==statusSnapshot||!isProductCatalogFormalRestoreAcceptanceAvailable(current,targetStatus))throw new Error('商品恢复状态已变化，请刷新后重新核对。');
  await postMaintenanceRequest(`/api/tasks/${taskID}/product-catalog-formal-restore-acceptance`,{confirmation,expected_evidence_sha256:current.product_catalog_stopped_projection_evidence.sha256});
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision)return;
  clearProductCatalogHandoffInspectionForChangedSessionOrCandidate();
  showMessage('正在正式接收商品恢复；系统复核数据、图片和凭证后恢复服务。请查看任务最终状态。');
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
async function startProductCatalogIsolatedSwitchOrVerifiedPairedRollback(taskID,rollback) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(!hasPermission('restore')||!task||(rollback?(!task.recovery_journal?.product_catalog_stopped_switch||['running','queued','completed','rolled_back'].includes(task.state)):!isContinuouslyStoppedProductCatalogIsolatedSwitchAvailable(task,targetStatus)))throw new Error('商品切换状态或维护权限已变化，请刷新后重试。');
  const credential=maintenanceCredential,connectionRevision=maintenanceConnectionRevision,taskSnapshot=JSON.stringify(task),statusSnapshot=JSON.stringify(targetStatus);
  const phrase=`${rollback?'ROLLBACK':'SWITCH'} STOPPED PRODUCT CATALOG ${targetStatus.site_id}/${targetStatus.environment}/${taskID}`;
  const confirmation=prompt(`${rollback?'将数据库和完整媒体一并回退；核对通过后恢复原商城服务。':'将商品数据库和完整媒体一起切换到候选，商城仍持续停写。核对页面和图片、保存凭证后，可正式接收并恢复服务；也可成对回退。'}\n请输入：${phrase}`);
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
