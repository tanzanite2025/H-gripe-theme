'use strict';
// shared dom formatting and task state labels. Loaded in index.html order; session credentials stay in memory.
const getMaintenanceConsoleElement = (id) => document.getElementById(id);
const formatBackupByteSize = (value) => value < 1024 * 1024 ? `${(value / 1024).toFixed(1)} KB` : `${(value / 1024 / 1024).toFixed(1)} MB`;
const productCatalogPhaseNames = {capturing_product_catalog_and_images:'正在读取商品与图片',product_catalog_and_actual_images_encrypted:'商品和实际图片已加密保存',restoring_product_module_in_isolated_target_clone:'正在隔离副本恢复商品',product_rows_images_and_unselected_data_preservation_verified:'商品、图片和其他数据保留验证完成'};
const stateNames = {queued:'等待执行',running:'执行中',ready_to_download:'已生成，可下载',generated_with_gaps:'已生成，但缺口阻止恢复',uploading:'上传中',uploaded:'上传完成，尚未验证',verified:'隔离恢复验证通过',awaiting_reconciliation:'已切换，等待对账',resuming:'对账已记录，正在恢复服务',releasing_rollback:'回退已核对，正在解除隔离',completed:'已记录对账并解除隔离',rolled_back:'已成对回退',failed:'失败',interrupted:'执行中断，请检查阶段',expired:'服务器工作副本已过期'};
function showMessage(message, failed = false) { const output = getMaintenanceConsoleElement('message'); output.textContent = message; output.className = failed ? 'error' : ''; }
stateNames.product_module_isolated_verified = '商品与图片隔离恢复通过';
Object.assign(productCatalogPhaseNames,{restoring_missing_terminal_order_history:'正在恢复缺失历史订单',order_history_rows_and_audit_ready_for_atomic_commit:'正在保存订单历史与恢复记录',order_history_restore_result_verified_before_service_release:'订单结果已核验，正在恢复服务',recovering_interrupted_order_history_restore:'正在核验中断订单任务',order_history_restore_requires_verified_recovery:'订单任务中断，等待核验恢复',order_history_restored_and_service_resumed:'订单历史已恢复，商城服务已恢复',uncommitted_order_restore_cancelled_and_service_resumed:'未提交的订单恢复已撤销，商城服务已恢复'});
productCatalogPhaseNames.accepting_verified_product_restore = '正在正式接收商品恢复';
productCatalogPhaseNames.product_restore_accepted_and_service_resumed = '商品与图片已正式恢复，商城服务已恢复';
stateNames.module_package_integrity_verified = '模块内容完整性校验通过，尚未恢复';
stateNames.module_isolated_restore_verified = '模块隔离恢复验证通过，未覆盖当前商城';
stateNames.module_restore_conflicts_found = '模块恢复存在冲突，请查看结果';
stateNames.module_chat_candidate_rehearsal_verified = '所选聊天策略副本验证通过，未覆盖当前商城';
Object.assign(productCatalogPhaseNames,{preparing_selected_chat_choice_candidate_rehearsal:'准备所选聊天策略副本验证',selected_chat_choice_full_rows_and_retained_target_verified:'完整聊天记录及其他数据保留核对通过'});
Object.assign(productCatalogPhaseNames,{preparing_selected_module_isolated_restore:'准备模块隔离恢复',checking_selected_module_target_dependencies:'检查模块依赖与内容冲突',selected_module_disposable_restore_and_target_preservation_verified:'隔离恢复及当前商城数据保留验证通过',selected_module_dependency_or_content_conflicts_require_resolution:'依赖、记录或文件存在冲突，恢复已阻断'});
Object.assign(productCatalogPhaseNames,{capturing_selected_business_modules:'正在读取所选业务模块和实际附件',selected_business_modules_encrypted:'所选业务模块及实际文件已加密保存',selected_module_contents_verified_restore_unavailable:'模块内容已校验，请继续验证可用恢复方式'});
stateNames.product_module_stopped_handoff_prepared = '商品交接准备完成，商城持续停写';
stateNames.product_module_switched_still_isolated = '商品已隔离切换，商城继续停写';
Object.assign(productCatalogPhaseNames,{invalidating_stopped_product_detail_cache_and_read_leases:'正在核验并处理商品详情缓存与读取锁',product_detail_cache_and_read_leases_verified_absent_still_isolated:'商品详情缓存与读取锁已核验为空，继续停写'});
Object.assign(productCatalogPhaseNames,{switching_product_with_application_continuously_stopped:'正在持续停写窗口内切换商品',product_stopped_switch_journal_persisted:'商品切换意图已保存，失败时须成对回退',product_database_and_complete_media_switched_application_stays_stopped:'商品数据库与完整媒体已切换，继续停写',rolling_back_stopped_product_database_and_complete_media:'正在成对回退商品数据库与完整媒体',product_stopped_switch_requires_explicit_verified_rollback:'商品切换中断，需明确执行核验回退',product_original_database_and_complete_media_verified_before_release:'原商城数据库与完整媒体已核验，正在解除隔离',product_stopped_switch_original_service_restored_after_verified_paired_rollback:'商品切换已成对回退，原商城已恢复服务'});
Object.assign(productCatalogPhaseNames,{preparing_product_handoff_with_continuous_isolation:'正在持续停写窗口内准备商品交接',product_handoff_prepared_application_remains_stopped:'商品候选已重建，继续保持停写',product_stopped_handoff_requires_explicit_revalidation:'交接准备中断，需重新核验或取消',verifying_original_before_product_handoff_cancellation:'正在核对当前商城后取消准备',product_handoff_cancelled_original_service_released:'交接准备已取消，当前商城已解除隔离'});
Object.assign(productCatalogPhaseNames,{waiting_for_product_view_writers_to_stop:'等待商品浏览相关进程停止',draining_idempotent_product_view_batches:'正在幂等保存浏览增量',rebuilding_product_candidate_after_view_drain:'正在基于最新浏览数重建商品预览',product_view_drain_and_fresh_candidate_rebuild_verified:'浏览增量已保存，商品预览已重建',product_view_drain_retained_for_explicit_retry:'浏览排空中断，维护隔离保留，需检查后重试'});

stateNames.configuration_site_content_restored_still_isolated="网站内容已写入，等待内容与缓存核对";
