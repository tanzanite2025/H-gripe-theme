'use strict';
// task list actions and package scope preview. Loaded in index.html order; session credentials stay in memory.
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
    const unresolvedStoppedHandoff=Boolean(stoppedHandoff&&stoppedHandoff.phase!=='cancelled'&&!['original_released','product_released'].includes(stoppedSwitch?.phase));
    const orderRestore=task.order_module_restore;
    const chatRestore=task.customer_service_module_restore;
    const profileRestore=task.customer_profile_module_restore;
    const mediaRestore=task.media_module_restore;
    const configurationRestore=task.configuration_site_content_restore;
    title.textContent = `${task.kind === 'selected_business_modules_backup' ? '所选业务模块备份' : task.kind === 'product_catalog_backup' ? '商品与图片备份' : task.kind === 'backup' ? '整站商城备份' : '上传恢复'} · ${new Date(task.created_at).toLocaleString()}`;article.append(title);
    const selectedModules=task.manifest?.selected_modules||task.selected_backup_modules;
    if(selectedModules?.length){const paragraph=document.createElement('p');paragraph.textContent=`包含模块：${describeSelectedBusinessBackupModules(selectedModules)}。${selectedModules.length===1&&selectedModules[0]==='orders'?'单独订单包可在验证后补回缺失终态订单、符合条件的已付款未结束订单与证据附件；内容冲突会阻断。':selectedModules.length===1&&selectedModules[0]==='customer_service'?'单独聊天包可在保存策略、完整行演练通过后正式恢复；共享身份、附件和事件冲突会阻断。':selectedModules.length===1&&selectedModules[0]==='customers'?'单独客户包可在演练通过后补回缺失客户；当前账号、积分、订阅、订单和聊天保留。':selectedModules.length===1&&selectedModules[0]==='media'?'单独图片与附件包可在演练通过后补缺，或修复登记与原包一致的文件字节；当前商品、订单、聊天和配置保留。':selectedModules.length===1&&selectedModules[0]==='configuration'?'单独配置包目前补回缺失网站内容、适配/辐条/轮胎目录、选购流程与问卷及市场/汇率/税率、邮件模板/历史版本和会员等级、承运商/配送区域、包装规格/商品SKU绑定及报关分类；其他配置和媒体登记须兼容，相关图片须已恢复。写入后核对内容、目录/计算结果与缓存再恢复服务。':'此组合的正式覆盖导入尚未开放。'}`;article.append(paragraph);}
    if(orderRestore){const paragraph=document.createElement('p');paragraph.textContent=orderRestore.phase==='released'?(orderRestore.outcome==='restored'?`订单历史已恢复并恢复服务，共补回 ${orderRestore.restored_rows} 条记录；恢复审计已同事务保存。`:'本次订单恢复未提交，新增附件已撤回，商城服务已恢复。'):'订单恢复尚未完成，商城保持维护。请核验提交结果并恢复任务；工作副本到期后仍可处理本次中断。';article.append(paragraph);}
    if(chatRestore){const paragraph=document.createElement('p');paragraph.textContent=chatRestore.phase==='released'?(chatRestore.outcome==='restored'?`聊天已按所选策略恢复并恢复服务，补入 ${chatRestore.restored_rows} 条、替换 ${chatRestore.replaced_rows||0} 条完整记录；恢复审计已同事务保存。`:'本次聊天恢复未提交，新增附件已撤回，商城服务已恢复。'):'聊天恢复尚未完成，请核验提交结果和附件后恢复任务；工作副本到期后仍可处理本次中断。';article.append(paragraph);}
    if(mediaRestore){const paragraph=document.createElement('p');paragraph.textContent=mediaRestore.phase==='released'?(mediaRestore.outcome==='restored'?`图片与附件已恢复并恢复服务，补回 ${mediaRestore.restored_rows} 条媒体记录、修复 ${mediaRestore.files?.filter(file=>file.original_identity).length||0} 个文件；审计已同事务保存。`:'媒体恢复未提交，新增文件已撤回、替换的原字节已核验恢复，商城服务已恢复。'):'媒体恢复尚未完成，请核验提交结果与文件；工作副本到期后仍可处理本次中断。';article.append(paragraph);}
    if (task.product_catalog_recovery && !unresolvedStoppedDrain && !unresolvedStoppedHandoff) {
      const result=task.product_catalog_recovery;
      const paragraph=document.createElement('p');
      paragraph.textContent=stoppedSwitch?.phase==='product_released'?`商品与图片已正式恢复并恢复服务；恢复 ${result.restored_rows} 行，验证 ${result.verified_image_files} 个图片文件，其他业务数据按预览保留。`:`商品隔离恢复通过：恢复 ${result.restored_rows} 行，验证 ${result.verified_image_files} 个图片文件，保留 ${result.preserved_tables} 张其他业务表及目标新增记录。正式恢复须继续停写准备、切换与核对后接收。`;
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
    for (const text of [`${(orderRestore&&task.state==='completed'?'订单历史恢复完成':orderRestore&&task.state==='rolled_back'?'未提交恢复已撤销':stateNames[task.state]) || task.state} · ${productCatalogPhaseNames[task.phase] || task.phase}`,`任务 ${task.id} · ${formatBackupByteSize(task.size)} · 到期 ${new Date(task.expires_at).toLocaleString()}`,`下载请求 ${task.download_requests} 次；电脑保存情况需要自行确认。`]) {const paragraph=document.createElement('p');paragraph.textContent=text;article.append(paragraph);}
    if (task.error) {const error=document.createElement('p');error.className='error';error.textContent=task.error;article.append(error);}
    if (task.ciphertext_sha256) {const checksum=document.createElement('p');checksum.textContent=`包校验 SHA-256：${task.ciphertext_sha256}`;article.append(checksum);}
    if(configurationRestore){const paragraph=document.createElement('p');paragraph.textContent=configurationRestore.phase==='released'?(configurationRestore.outcome==='restored'?`网站内容已补回 ${configurationRestore.restored_rows} 条；已记录人工内容与缓存核对，商城服务已恢复。`:'网站内容恢复未提交，原数据已核验并恢复服务。'):`网站内容任务保留维护隔离；已补回 ${configurationRestore.restored_rows||0} 条。须核验提交结果，并完成内容、文章缓存及页面/CDN缓存核对后恢复服务。`;article.append(paragraph);}
    if(isConfigurationSiteContentRestoreActionAvailable(task,targetStatus))taskActionButton(article,'补回缺失业务配置（暂停商城）',()=>startConfigurationSiteContentRestoreOrReviewedRelease(task.id,'restore'),'danger');
    if(isConfigurationSiteContentRestoreActionAvailable(task,targetStatus,true))taskActionButton(article,'核验网站内容恢复任务',()=>startConfigurationSiteContentRestoreOrReviewedRelease(task.id,'recover'),'danger');
    if(isConfigurationSiteContentReviewedReleaseAvailable(task,targetStatus))taskActionButton(article,'记录内容与缓存核对并恢复服务',()=>startConfigurationSiteContentRestoreOrReviewedRelease(task.id,'release'),'danger');
    if (hasPermission('export') && ['ready_to_download','generated_with_gaps'].includes(task.state)) taskActionButton(article,'下载到电脑',()=>downloadMaintenanceBackupPackageToLocalComputer(task));
    if (task.manifest && !unresolvedStoppedDrain && !unresolvedStoppedHandoff) taskActionButton(article,'查看范围与覆盖预览',()=>showOverwritePreview(task));
    if (hasPermission('restore') && !orderRestore && !chatRestore && !profileRestore && !mediaRestore && !configurationRestore && !targetStatus.application_fenced && task.kind==='upload' && task.manifest?.scope==='selected_business_modules' && !task.work_copy_expired && !task.recovery_journal && new Date(task.expires_at).getTime()>Date.now() && (['module_package_integrity_verified','module_isolated_restore_verified','module_restore_conflicts_found','module_chat_candidate_rehearsal_verified'].includes(task.state)||(['failed','interrupted'].includes(task.state)&&task.selected_module_restore_preview))) taskActionButton(article,'验证模块恢复（暂停商城）',()=>startSelectedBusinessModuleIsolatedRestorePreview(task));
    if(isOrderModuleHistoryRestoreActionAvailable(task,targetStatus,false))taskActionButton(article,'恢复缺失订单记录（暂停商城）',()=>startOrderModuleHistoryRestoreOrInterruptedRecovery(task.id,false),'danger');
    if(isOrderModuleHistoryRestoreActionAvailable(task,targetStatus,true))taskActionButton(article,'核验并恢复订单任务',()=>startOrderModuleHistoryRestoreOrInterruptedRecovery(task.id,true),'danger');
    if(isCustomerServiceModuleFormalRestoreActionAvailable(task,targetStatus,false))taskActionButton(article,'按所选策略恢复聊天（暂停商城）',()=>startCustomerServiceModuleFormalRestoreOrInterruptedRecovery(task.id,false),'danger');
    if(isCustomerServiceModuleFormalRestoreActionAvailable(task,targetStatus,true))taskActionButton(article,'核验并恢复聊天任务',()=>startCustomerServiceModuleFormalRestoreOrInterruptedRecovery(task.id,true),'danger');
    if(profileRestore){const paragraph=document.createElement('p');paragraph.textContent=profileRestore.phase==='released'?(profileRestore.outcome==='restored'?`客户资料已恢复并恢复服务，补入 ${profileRestore.restored_rows} 条完整记录；当前账号、积分、订单和聊天保留，恢复审计已同事务保存。`:'本次客户恢复未提交，原数据已核验，商城服务已恢复。'):'客户资料恢复中断，请核验原提交结果后恢复任务。';article.append(paragraph);}
    if(isMediaModuleFormalRestoreActionAvailable(task,targetStatus,false))taskActionButton(article,'恢复图片与附件（暂停商城）',()=>startMediaModuleFormalRestoreOrInterruptedRecovery(task.id,false),'danger');
    if(isMediaModuleFormalRestoreActionAvailable(task,targetStatus,true))taskActionButton(article,'核验并恢复图片与附件任务',()=>startMediaModuleFormalRestoreOrInterruptedRecovery(task.id,true),'danger');
    if(isCustomerProfileModuleFormalRestoreActionAvailable(task,targetStatus,false))taskActionButton(article,'恢复缺失客户资料（暂停商城）',()=>startCustomerProfileModuleFormalRestoreOrInterruptedRecovery(task.id,false),'danger');
    if(isCustomerProfileModuleFormalRestoreActionAvailable(task,targetStatus,true))taskActionButton(article,'核验并恢复客户资料任务',()=>startCustomerProfileModuleFormalRestoreOrInterruptedRecovery(task.id,true),'danger');
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
    if (stoppedSwitch && !['original_released','product_released'].includes(stoppedSwitch.phase)) { const description=document.createElement('p');description.textContent='商品和图片已进入切换流程，商城继续停写。完成页面、图片和缓存核对并保存凭证后，可正式接收并恢复服务；接收中断后须核验成对回退。';article.append(description); }
    if (isStoppedProductCatalogPostSwitchInspectionAvailable(task,targetStatus)) taskActionButton(article,'核对本次商品切换结果',()=>inspectStoppedProductCatalogPostSwitch(task.id));
    const cacheProgress=task.product_catalog_stopped_detail_cache_invalidation;
    if(cacheProgress){const description=document.createElement('p');description.textContent=stoppedSwitch?.phase==='product_released'?'本次恢复前的商品缓存核对记录已保留；商城已恢复服务，缓存由正常读取重建。':stoppedSwitch?.phase==='original_released'?'本次商品切换的缓存处理记录已保留；原商城已成对回退，详情缓存由正常读取重建。':cacheProgress.phase==='verified_absent'?(task.state==='product_module_switched_still_isolated'?`商品详情缓存与读取锁已核对为空（计划详情键 ${cacheProgress.planned_detail_cache_keys}、读取锁 ${cacheProgress.planned_read_lease_keys}）。商城继续停写；页面、媒体及营业接收仍待完成。`:'旧商品详情缓存处理记录已保留；当前仍需核验，商城继续停写，请检查任务状态和失败原因。'):'商品详情缓存处理尚未完成，已保存原处理清单；可在当前连续停写窗口核验重试，重启后请核验回退。';article.append(description);}
    if(isStoppedProductDetailCacheInvalidationAvailable(task,targetStatus))taskActionButton(article,cacheProgress?.phase==='verified_absent'?'复核商品详情缓存（仍停写）':cacheProgress?'重试商品详情缓存处理（仍停写）':'处理商品详情缓存（仍停写）',()=>startStoppedProductDetailCacheInvalidation(task.id),'danger');
    if(isStoppedProductProjectionInspectionAvailable(task,targetStatus))taskActionButton(article,'核对商品页面、媒体与边缘缓存',()=>inspectStoppedProductStorefrontProjectionBaseline(task.id));
    if(task.product_catalog_stopped_projection_evidence){const note=document.createElement('p');note.textContent=stoppedSwitch?.phase==='product_released'?'商品恢复已按本次核对凭证正式接收；旧任务不能再次回退已恢复营业的数据。':'已保留本次商品的人工核对凭证；请确认实际页面和图片结果，再正式接收恢复服务。';article.append(note);}
    if(isProductCatalogFormalRestoreAcceptanceAvailable(task,targetStatus))taskActionButton(article,'正式接收商品恢复并恢复服务',()=>acceptProductCatalogRestoreAndResumeService(task.id),'danger');
    if (hasPermission('restore') && stoppedSwitch && !['completed','rolled_back','running','queued'].includes(task.state)) taskActionButton(article,'成对回退商品并恢复原商城',()=>startProductCatalogIsolatedSwitchOrVerifiedPairedRollback(task.id,true),'danger');
    if (hasPermission('upload') && !orderRestore && !chatRestore && !profileRestore && !mediaRestore && !configurationRestore && task.kind==='upload' && task.uploaded_bytes===task.size && ['uploaded','failed','interrupted'].includes(task.state) && !task.recovery_journal && !unresolvedStoppedDrain && !unresolvedStoppedHandoff) taskActionButton(article,'校验包并检查可用恢复方式',async()=>{await postMaintenanceRequest(`/api/tasks/${task.id}/verify`);pendingVerificationTaskID=task.id;showMessage('正在校验包。整站及单独商品包会演练隔离恢复，模块组合包检查记录与实际文件完整性。');});
    if (hasPermission('restore') && task.kind==='upload' && task.state==='verified' && !task.work_copy_expired) {
      for (const [module, label] of [['orders','订单'],['customer_service','客服']]) taskActionButton(article,`检查${label}单独恢复依赖`,()=>inspectModuleRecoveryDependencies(task.id,module));
    }
    if (hasPermission('restore') && task.state==='verified' && task.manifest?.scope==='whole_application') taskActionButton(article,'覆盖恢复',()=>confirmAndApplyWholeStorefrontPairedRecovery(task),'danger');
    if (hasPermission('restore') && task.state==='awaiting_reconciliation') {
      taskActionButton(article,'开启维护连接进行对账',()=>confirmAndOpenWholeStorefrontReconciliationConnections(task));
      taskActionButton(article,'填写对账凭证',()=>openWholeStorefrontReconciliationEvidencePanel(task));
    }
    if (hasPermission('restore') && task.recovery_journal && !stoppedSwitch && !['completed','rolled_back','running','resuming','releasing_rollback'].includes(task.state)) taskActionButton(article,'成对回退',()=>confirmAndRollBackWholeStorefrontDatabaseAndMedia(task),'danger');
    if (hasPermission('restore') && !orderRestore && !chatRestore && !profileRestore && !mediaRestore && !configurationRestore && targetStatus.application_fenced && !task.recovery_journal && !['running','queued'].includes(task.state) && !unresolvedStoppedDrain && !unresolvedStoppedHandoff) taskActionButton(article,stoppedDrain?'解除已重建预览的维护隔离':'解除未改数据的维护隔离',async()=>{await postMaintenanceRequest(`/api/tasks/${task.id}/release`);showMessage('已解除未发生切换的维护隔离。');});
    container.append(article);
  }
}
function showOverwritePreview(task) {
  const manifest=task.manifest;getMaintenanceConsoleElement('preview-panel').hidden=false;
  renderSelectedBusinessModuleIsolatedRestorePreview(task);
  if (manifest.scope==='selected_business_modules') {
    getMaintenanceConsoleElement('preview-description').textContent=`业务模块包：${describeSelectedBusinessBackupModules(manifest.selected_modules)} · 恢复点 ${manifest.captured_at} · ${manifest.site_id}/${manifest.environment}。先校验包，再执行隔离恢复验证；依赖或内容冲突会阻断。单独订单包可在资格通过后补回缺失终态订单及符合条件的已付款未结束订单；其他范围以任务可用入口为准。`;
  } else if (manifest.scope==='product_catalog_with_images') {
    getMaintenanceConsoleElement('preview-description').textContent=`商品与图片独立包 · 恢复点 ${manifest.captured_at} · ${manifest.site_id}/${manifest.environment}。上传后先在当前商城的隔离副本恢复商品；保留未选择的业务记录、目标新增商品及已有 SKU 的当前库存。新增 SKU 使用备份库存，结果会单独列出。SKU 或库存归属冲突会阻断。正式恢复需保持停写、切换并核对后接收。`;
  } else {
  getMaintenanceConsoleElement('preview-description').textContent=`恢复点 ${manifest.captured_at} · 包 ${manifest.package_id} · ${manifest.site_id}/${manifest.environment} · 迁移版本 ${manifest.migration_version}。目标数据库 ${targetStatus.database_name}；备份时数据库占用 ${formatBackupByteSize(manifest.database_size_bytes)}。附件目标：${Object.entries(targetStatus.storage_roots).map(([name,path])=>`${name} → ${path}`).join('；')}。覆盖将替换恢复点之后的数据。`;
  }
  const rows=getMaintenanceConsoleElement('preview-tables');rows.replaceChildren();
  const previewTableNames=manifest.scope!=='whole_application'?Object.keys(manifest.tables):Array.from(new Set([...Object.keys(task.preview_target_tables||{}),...Object.keys(manifest.tables)]));
  for(const name of previewTableNames.sort()){const row=document.createElement('tr');for(const value of [name,task.preview_target_tables?.[name]??'未预检',manifest.tables[name]??'不包含']){const cell=document.createElement('td');cell.textContent=String(value);row.append(cell);}rows.append(row);}
  getMaintenanceConsoleElement('preview-files').textContent=`媒体文件 ${manifest.media_file_count} 个 · ${formatBackupByteSize(manifest.media_file_bytes)} · 整包的加密与文件清单均须通过校验。`;
  getMaintenanceConsoleElement('preview-gaps').textContent=manifest.gaps.length?`恢复被阻断：${manifest.gaps.join('；')}`:'包内未报告附件引用缺口。隔离演练结果以任务状态为准。';getMaintenanceConsoleElement('preview-panel').scrollIntoView({behavior:'smooth'});
}
