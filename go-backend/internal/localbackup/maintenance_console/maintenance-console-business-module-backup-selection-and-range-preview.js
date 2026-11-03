'use strict';
// business module backup selection and range preview. Loaded in index.html order; session credentials stay in memory.
let selectedBusinessBackupPreview = null;
let selectedBusinessBackupRevision = 0;
let businessBackupRequestInProgress = false;
let pendingAdminBusinessBackupScopeSelection = null;
let pendingAdminBusinessBackupWorkflow = '';
function clearPendingAdminBusinessBackupScopeSelectionAndWorkflow() {
  pendingAdminBusinessBackupScopeSelection = null;
  pendingAdminBusinessBackupWorkflow = '';
}
function applyPendingAdminBusinessBackupScopeSelectionAndWorkflow() {
  if(!targetStatus||businessBackupRequestInProgress||uploading)return;
  const selection=pendingAdminBusinessBackupScopeSelection;
  const workflow=pendingAdminBusinessBackupWorkflow;
  clearPendingAdminBusinessBackupScopeSelectionAndWorkflow();
  if(selection){
    const allowed=new Set((targetStatus.business_module_backup_options||[]).map(option=>option.id));
    if(selection.modules.some(id=>!allowed.has(id)))return;
    for(const radio of document.querySelectorAll('input[name="backup-export-mode"]'))radio.checked=radio.value===selection.mode;
    for(const input of document.querySelectorAll('#business-backup-module-options input'))input.checked=selection.modules.includes(input.value);
    invalidateSelectedBusinessBackupPreview();
  }
  if(workflow==='upload'){getMaintenanceConsoleElement('upload-form').scrollIntoView({behavior:'smooth',block:'start'});getMaintenanceConsoleElement('upload-file').focus();}
  if(workflow==='export'){getMaintenanceConsoleElement('business-backup-selection').scrollIntoView({behavior:'smooth',block:'start'});if(isBusinessBackupModuleModeSelected())previewCurrentSelectedBusinessBackupModules().catch(error=>showMessage(error.message,true));}
}
function receiveAdminBackupScopeSelectionOrWorkflowNavigation(event) {
  if(event.origin!==location.origin||event.source!==window.parent||window.parent===window||!targetStatus)return;
  if(event.data?.type==='local-backup-maintenance-select-scope'){
    const {mode,modules}=event.data;
    const allowed=new Set((targetStatus.business_module_backup_options||[]).map(option=>option.id));
    if(!['whole','modules'].includes(mode)||!Array.isArray(modules)||modules.some(id=>!allowed.has(id))||new Set(modules).size!==modules.length)return;
    if(mode==='whole'&&modules.length)return;
    pendingAdminBusinessBackupScopeSelection={mode,modules:[...modules]};
    invalidateSelectedBusinessBackupPreview();
    applyPendingAdminBusinessBackupScopeSelectionAndWorkflow();
  }else if(event.data?.type==='local-backup-maintenance-show-workflow'){
    const workflow=event.data.workflow;
    if(!['export','upload'].includes(workflow))return;
    pendingAdminBusinessBackupWorkflow=workflow;
    applyPendingAdminBusinessBackupScopeSelectionAndWorkflow();
  }
}
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
  const busy=businessBackupRequestInProgress||uploading;
  const allowed=Boolean(hasPermission('export')&&!targetStatus.application_fenced&&!busy);
  const uploadAllowed=Boolean(hasPermission('upload')&&!busy);
  getMaintenanceConsoleElement('business-backup-module-picker').hidden=!moduleMode;
  getMaintenanceConsoleElement('business-backup-selection-summary').textContent=modules.length?`已选择 ${modules.length} 个模块：${describeSelectedBusinessBackupModules(modules)}`:'请勾选要保存的模块。';
  getMaintenanceConsoleElement('preview-selected-business-modules').disabled=!allowed||!modules.length;
  getMaintenanceConsoleElement('generate').textContent=moduleMode?'生成所选模块备份':'生成整站备份';
  getMaintenanceConsoleElement('generate').disabled=!allowed||(moduleMode&&(!modules.length||!selectedBusinessBackupPreview));
  for(const input of document.querySelectorAll('#business-backup-selection input'))input.disabled=!allowed;
  for(const id of ['select-all-business-backup-modules','clear-business-backup-modules'])getMaintenanceConsoleElement(id).disabled=!allowed;
  getMaintenanceConsoleElement('upload-button').disabled=!uploadAllowed;
  if(window.parent!==window&&typeof window.parent?.postMessage==='function')window.parent.postMessage({type:'local-backup-maintenance-backup-controls-state',busy,exportAllowed:allowed,uploadAllowed},location.origin);
}
async function previewCurrentSelectedBusinessBackupModules() {
  if(!hasPermission('export')||businessBackupRequestInProgress||uploading||targetStatus.application_fenced)return;
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
  } finally {if(connectionRevision===maintenanceConnectionRevision){businessBackupRequestInProgress=false;applyPendingAdminBusinessBackupScopeSelectionAndWorkflow();updateBusinessBackupSelectionControls();}}
}
async function generateBackupForCurrentBusinessModuleSelection() {
  if(!hasPermission('export')||businessBackupRequestInProgress||uploading||targetStatus.application_fenced)return;
  const connectionRevision=maintenanceConnectionRevision;
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
  } finally {if(connectionRevision===maintenanceConnectionRevision){businessBackupRequestInProgress=false;applyPendingAdminBusinessBackupScopeSelectionAndWorkflow();updateBusinessBackupSelectionControls();}}
}
