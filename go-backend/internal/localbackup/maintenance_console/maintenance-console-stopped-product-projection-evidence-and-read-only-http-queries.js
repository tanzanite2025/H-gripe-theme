'use strict';
// stopped product projection evidence and read only http queries. Loaded in index.html order; session credentials stay in memory.
let stoppedProductProjectionEvidenceContext = null;
let stoppedProductProjectionInspectionRevision = 0;
let stoppedProductReadOnlyHTTPQueryRevision = 0;
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
async function handleStoppedProductReadOnlyStorefrontHTTPQuerySubmission(event) {
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
}
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
async function handleStoppedProductReadOnlyMediaHTTPQuerySubmission(event) {
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
}
async function handleStoppedProductProjectionEvidenceSubmission(event) {
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
}
