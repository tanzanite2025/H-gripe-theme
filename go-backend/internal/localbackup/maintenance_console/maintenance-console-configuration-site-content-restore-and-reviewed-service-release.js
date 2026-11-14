'use strict';
// Configuration content has its own confirmation and post-commit service release.
function isConfigurationSiteContentRestoreActionAvailable(task,status,recoverInterrupted=false) {
  if(!status?.permissions?.includes('ops:backup:restore')||task?.kind!=='upload'||task.manifest?.scope!=='selected_business_modules'||JSON.stringify(task.manifest.selected_modules)!=='["configuration"]'||task.recovery_journal||task.order_module_restore||task.customer_service_module_restore||task.customer_profile_module_restore||task.media_module_restore||['running','queued'].includes(task.state))return false;
  if(recoverInterrupted)return Boolean(task.configuration_site_content_restore&&task.configuration_site_content_restore.phase!=='released'&&['failed','interrupted'].includes(task.state));
  const report=task.selected_module_restore_preview;
  return Boolean(!status.application_fenced&&!task.configuration_site_content_restore&&!task.work_copy_expired&&!task.stage_database&&!task.selected_module_disposable_stage&&new Date(task.expires_at).getTime()>Date.now()&&task.state==='module_isolated_restore_verified'&&report?.sha256&&report.binding&&JSON.stringify(report.modules)==='["configuration"]'&&report.target_unchanged&&report.disposable_stage_removed&&report.isolated_restore_verified&&report.online_overwrite_available===false&&report.configuration_site_content_restore_eligibility?.eligible&&report.configuration_site_content_restore_eligibility.policy==='restore_missing_business_configuration_including_payment_methods_and_original_time_bounded_protection_preserve_existing_rows_and_registered_files_v14');
}
function isConfigurationSiteContentReviewedReleaseAvailable(task,status) {
  return Boolean(isConfigurationSiteContentRestoreActionAvailable({...task,state:'interrupted'},status,true)&&status.application_fenced&&['configuration_site_content_restored_still_isolated','failed','interrupted'].includes(task.state)&&task.configuration_site_content_restore.phase==='rows_committed'&&task.configuration_site_content_restore.sha256);
}
async function startConfigurationSiteContentRestoreOrReviewedRelease(taskID,action) {
  const task=displayedMaintenanceTasks.find(item=>item.id===taskID);
  const available=current=>action==='release'?isConfigurationSiteContentReviewedReleaseAvailable(current,targetStatus):isConfigurationSiteContentRestoreActionAvailable(current,targetStatus,action==='recover');
  if(!['restore','recover','release'].includes(action)||!available(task))throw new Error('网站内容恢复状态或权限已变化，请刷新后核对。');
  const taskSnapshot=JSON.stringify(task),statusSnapshot=JSON.stringify(targetStatus),credential=maintenanceCredential,revision=maintenanceConnectionRevision;
  let reference='';
  if(action==='release') {
    reference=prompt('请先核对恢复后的文章、图库、FAQ、展示、跳转、适配/辐条/轮胎目录、选购流程/问卷及市场/汇率/税率、邮件模板/历史版本和会员等级、承运商/配送区域、包装规格/商品SKU绑定及报关分类、运费模板/计费规则和线路聚合、承运服务/4PX渠道及燕文发布渠道和产品/国家/仓库参考目录、积分计划和推荐计划原版本、优惠券原规则与使用次数、追踪服务商/承运映射和税率采集配置、4PX/燕文物流账号及邮件发送方、支付方式/临时支付保护配置，核对目录与计算结果、选购步骤和问卷选项、市场语言/币种及汇率有效期、结算税率与参考快照、邮件模板预览和版本历史、会员等级区间及折扣、包装商品/SKU归属和报关风险/来源、客户运费币种/首续重/附加费、承运服务/线路成员、渠道环境/启停状态及燕文产品参考、积分/推荐计划原版本及启停和注册奖励规则，确认原积分余额与已发奖励保持，核对追踪原凭据、生产/沙箱环境、承运映射、自动注册/回调/轮询开关和税率采集间隔/原同步时间；放行后后台按原开关继续运行。核对4PX/燕文原账号密钥、网关环境及启停，核对邮件原SMTP、密钥优先级、默认/启停及原测试记录。核对支付方式原精确费率/金额、启停和保护规则范围/到期时间，确认过期或撤销规则未被重新启用。核对优惠券原有效期/状态、商品/分类范围、私人接收人、用户/游客额度及原使用/撤销记录，处理文章缓存与页面/CDN缓存。填写核对记录位置或工单号（4–2000 字）；本入口记录人工核对，不会自动清理缓存：');
    if(reference===null)return;
    if(reference.trim().length<4||reference.length>2000)throw new Error('请填写有效的内容与缓存核对记录。');
  }
  const prefix=action==='release'?'RELEASE CONFIGURATION CONTENT':action==='recover'?'RECOVER CONFIGURATION CONTENT':'RESTORE CONFIGURATION CONTENT';
  const phrase=`${prefix} ${targetStatus?.site_id}/${targetStatus?.environment}/${taskID}`;
  const description=action==='restore'?'将暂停商城，只补回缺失的网站内容及车架/前叉/花鼓、辐条计算和轮胎适配目录、选购流程/问卷及市场/汇率/税率、邮件模板/历史版本和会员等级、承运商/配送区域、包装规格/商品SKU绑定及报关分类、运费模板/计费规则和线路聚合、承运服务/4PX渠道及燕文发布渠道和产品/国家/仓库参考目录、积分计划和推荐计划原版本、优惠券原规则与使用次数、追踪服务商/承运映射和税率采集配置、4PX/燕文物流账号及邮件发送方、支付方式/临时支付保护配置。恢复后的包装规则和报关分类用于之后的业务查询；只恢复原包中的物流数据，不联网同步或发布渠道，不购买面单或调用外部物流。恢复的会员等级会用于之后的等级匹配与折扣。积分/推荐计划按原版本恢复，恢复不新建或发布版本、不签到、不发放积分或重放注册推荐奖励；后续业务读取原启停和规则。优惠券保留原日期、商品/分类范围、私人接收人和使用额度；原使用记录须完整且一致，不补造或重放使用记录。追踪服务商保留原 API/Webhook 凭据、环境及自动注册/轮询/回调开关，税率采集保留原开关、间隔及同步记录；恢复过程中不注册物流、不轮询、不采集，放行后后台按原开关继续运行。4PX/燕文账号保留原密文、生产/test/fat环境、启停与同步记录；原业务密钥须在维护服务与商城一致，不联网自检、同步或购买面单。放行后原后台任务可继续使用这些账号。邮件发送方保留原SMTP地址/端口/加密方式、密码密文、启停/默认标记及原测试记录；按原业务密钥优先级本地验证，不发送测试邮件或重置发送记录，不清除当前默认发送方；放行后按原配置处理后续邮件。支付方式保留原精确费率/金额、启停及软删除；临时支付保护保留原范围、原因、操作员和到期时间，不续期或重新启用。网关密钥仍属于通用配置；不发起支付/退款、不重放支付事件。不发送邮件、不修改原客户积分。不联网同步汇率/税率，不刷新商品价格或重算旧订单；原汇率有效期保留。现有内容、支付交易、客户、库存、订单和聊天保留；相关图片须已存在且一致。写入后继续维护，完成内容、目录/计算结果与缓存核对后另行恢复服务。':action==='recover'?'将核对原提交收据。未提交则核验原数据并恢复服务；已提交则保留维护，继续内容与缓存核对。':'将记录你对本次内容与缓存的核对结果，复核提交数据和图片后恢复商城服务。';
  const confirmation=prompt(`${description}\n请输入：${phrase}`);
  if(confirmation===null)return;
  const current=displayedMaintenanceTasks.find(item=>item.id===taskID);
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision||JSON.stringify(current)!==taskSnapshot||JSON.stringify(targetStatus)!==statusSnapshot||!available(current))throw new Error('网站内容恢复状态已变化，请刷新后核对。');
  const suffix=action==='release'?'reviewed-service-release':action==='recover'?'interrupted-restore-recovery':'formal-restore';
  const body=action==='release'?{confirmation,expected_journal_sha256:current.configuration_site_content_restore.sha256,content_and_cache_review_reference:reference.trim()}:action==='restore'?{confirmation,expected_preview_sha256:current.selected_module_restore_preview.sha256}:{confirmation};
  await postMaintenanceRequest(`/api/tasks/${taskID}/configuration-site-content-${suffix}`,body);
  if(credential!==maintenanceCredential||revision!==maintenanceConnectionRevision)return;
  showMessage('网站内容恢复操作已提交，请查看任务最终状态。');
}
