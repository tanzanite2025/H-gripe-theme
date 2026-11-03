'use strict';
// order history restore eligibility confirmation and row inspection. Loaded in index.html order; session credentials stay in memory.
function describeOrderHistoryRestoreEligibilityBlocker(reason) {
  const [code,table]=reason.split(':');
  const descriptions={only_a_separate_order_package_is_supported:'请选择单独的订单模块包',missing_attachment_requires_terminal_evidence_owner:'缺失附件未能关联到终态订单证据',missing_history_requires_a_terminal_order:'待恢复记录不属于已完成、已取消或已退款订单',missing_execution_shared_or_unsupported_records:'缺失记录尚无可执行的恢复规则',nonterminal_order_events_require_separate_reconciliation:'订单事件尚未处理完毕，需要先核对执行结果',dependencies_or_files_conflict:'共享依赖或附件文件存在差异',current_related_execution_records_not_in_package:'这些订单已有备份之后新增的执行或通知记录',current_content_or_protected_record_conflict:'当前记录存在内容差异或受保护的缺失记录',transaction_requires_recorded_completion_and_provider_identity:'交易缺少完成时间或支付平台交易编号',refund_requires_recorded_completion_and_matching_transaction:'退款缺少完成凭据，或与原交易不对应',refund_execution_requires_matching_successful_receipt:'退款执行的成功凭据、平台编号、金额或归属不对应',refund_line_requires_matching_completed_refund_and_order_item:'退款明细与已完成退款或原订单明细不对应',order_request_requires_bound_identity_and_digest:'订单请求幂等记录缺少有效订单身份或请求摘要',refund_request_requires_bound_completed_refund_and_digest:'退款请求幂等记录缺少有效退款身份或请求摘要',shipment_evidence_requires_recorded_shipping_and_order_identity:'发货证据缺少发货时间或订单号不对应',historical_event_requires_recorded_processed_state:'事件缺少已处理状态、完成时间或幂等键',historical_event_has_execution_or_reconciliation_state:'事件仍有执行租约、不明结果或对账标记',historical_event_requires_registered_order_ownership:'事件类型或业务归属尚未支持恢复',historical_payment_event_requires_completed_refund:'支付事件未关联到有完成凭据的退款',historical_event_owner_not_supported:'该事件归属范围尚未支持恢复',historical_fact_has_no_recorded_order_owner:'历史记录缺少明确订单归属',notification_requires_sent_receipt_and_processed_event:'通知缺少已发送凭据或对应事件尚未处理完毕',restored_order_domain_or_receipt_validation_failed:'副本恢复后的订单关联、事件载荷或通知凭据核对未通过'};
  Object.assign(descriptions,{global_payment_receipt_requires_completed_state_without_execution_lease:'支付请求尚未完成或仍有执行租约',global_payment_receipt_scope_not_registered:'支付请求类型尚未登记恢复规则',global_payment_receipt_request_identity_invalid:'支付请求的客户、原始幂等键或请求摘要无效',global_payment_receipt_completion_timestamps_invalid:'支付请求的创建、更新或对账时间不完整',global_payment_receipt_recorded_response_invalid:'支付请求缺少可安全返回的原始 HTTP 结果'});
  Object.assign(descriptions,{missing_attachment_requires_registered_order_evidence_owner:'缺失附件未能关联到原订单证据',order_state_requires_separate_payment_or_dispute_reconciliation:'订单仍待付款、审核或争议核对，当前不能补回',active_order_requires_recorded_payment_without_fulfillment_hold:'未结束订单缺少已付款事实或仍有履约冻结',active_order_lifecycle_facts_are_inconsistent:'未结束订单的终止时间或历史事件与当前状态不一致',active_order_requires_matching_shipping_facts:'未结束订单的发货状态、时间或原发货证据不对应',active_order_requires_one_matching_settled_transaction:'未结束订单须有唯一且金额、币种、渠道一致的已完成收款凭据',active_order_refund_dispute_or_review_workflow_requires_reconciliation:'未结束订单仍有关联退款、售后、争议或支付审核，需要单独核对',active_order_requires_complete_original_item_identities:'未结束订单缺少完整原明细、商品或 SKU 身份',active_order_component_inventory_requires_separate_dependency_contract:'未结束订单的配置组件库存分配尚需单独核对'});
  Object.assign(descriptions,{pending_order_inventory_reservation_requires_durable_release_reconciliation:'待付款订单缺少可核验的库存预留与返还状态，暂不能补回',active_order_component_inventory_configuration_invalid:'配置组件的库存分配结构、SKU 或数量无效',active_order_component_inventory_original_dependency_missing:'原包缺少配置组件 SKU 的依赖摘要，请使用带完整依赖的新订单包'});
  const tableNames={orders:'订单',transactions:'交易',refunds:'退款',refund_line_items:'退款明细',payment_refund_executions:'退款执行凭据',order_idempotencies:'订单幂等记录',refund_idempotencies:'退款幂等记录',shipment_records:'发货证据',outbox_events:'业务事件',email_delivery_records:'通知凭据',payment_operation_idempotencies:'全局支付请求凭据'};
  return `${descriptions[code]||code}${table?`（${tableNames[table]||table}）`:''}`;
}
function isOrderModuleHistoryRestoreActionAvailable(task,status,recoverInterrupted) {
  if(!status?.permissions?.includes('ops:backup:restore')||task?.kind!=='upload'||task.manifest?.scope!=='selected_business_modules'||JSON.stringify(task.manifest.selected_modules)!=='["orders"]'||task.recovery_journal||task.configuration_site_content_restore||task.media_module_restore||task.customer_service_module_restore||task.customer_profile_module_restore||['running','queued'].includes(task.state))return false;
  if(recoverInterrupted)return Boolean(task.order_module_restore&&task.order_module_restore.phase!=='released'&&['failed','interrupted'].includes(task.state));
  const report=task.selected_module_restore_preview;
  return Boolean(!status.application_fenced&&!task.order_module_restore&&!task.work_copy_expired&&new Date(task.expires_at).getTime()>Date.now()&&task.state==='module_isolated_restore_verified'&&report?.sha256&&report.binding&&report.target_unchanged&&report.disposable_stage_removed&&report.isolated_restore_verified&&report.order_history_restore_eligibility?.eligible&&report.order_history_restore_eligibility.policy==='restore_terminal_and_settled_active_orders_with_normal_constraints_and_original_inventory_dependencies_v6');
}
async function startOrderModuleHistoryRestoreOrInterruptedRecovery(taskID,recoverInterrupted) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(!isOrderModuleHistoryRestoreActionAvailable(task,targetStatus,recoverInterrupted))throw new Error('订单恢复资格、任务或权限已变化，请刷新后重新核对。');
  const taskSnapshot=JSON.stringify(task),statusSnapshot=JSON.stringify(targetStatus),credential=maintenanceCredential,revision=maintenanceConnectionRevision;
  const phrase=`${recoverInterrupted?'RECOVER ORDER RESTORE':'RESTORE ORDER HISTORY'} ${targetStatus.site_id}/${targetStatus.environment}/${taskID}`;
  const confirmation=prompt(`${recoverInterrupted?'将核验本次订单恢复的提交凭证；未提交则撤回本次新增附件，已提交则验证结果，核验通过后恢复商城服务。':'将暂停商城，补回缺失终态订单、符合条件的已付款未结束订单及附件，以及具备完成凭据的交易、退款、幂等记录、发货证据、已处理事件、已发送通知和已完成支付请求凭据。已有记录及较新备注保留；不会重新收款、退款、发货或发送通知。其他冲突会阻断。成功后恢复商城服务。'}\n请输入：${phrase}`);
  if(confirmation===null)return;
  const current=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision||JSON.stringify(current)!==taskSnapshot||JSON.stringify(targetStatus)!==statusSnapshot||!isOrderModuleHistoryRestoreActionAvailable(current,targetStatus,recoverInterrupted))throw new Error('订单恢复状态已变化，请刷新后重新核对。');
  const body=recoverInterrupted?{confirmation}:{confirmation,expected_preview_sha256:current.selected_module_restore_preview.sha256};
  await postMaintenanceRequest(`/api/tasks/${taskID}/${recoverInterrupted?'order-module-interrupted-restore-recovery':'order-module-formal-history-restore'}`,body);
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision)return;
  showMessage(recoverInterrupted?'正在核验订单提交结果与附件，请查看任务最终状态。':'正在补回历史订单与附件，请查看任务最终状态。');
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
