'use strict';
// recovery durable event and side effect receipt reports. Loaded in index.html order; session credentials stay in memory.
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
