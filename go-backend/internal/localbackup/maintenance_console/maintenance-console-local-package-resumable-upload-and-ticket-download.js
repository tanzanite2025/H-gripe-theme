'use strict';
// local package resumable upload. Loaded in index.html order; session credentials stay in memory.
let uploading = false;
async function uploadLocalBackupFile() {
  if(!hasPermission('upload')||uploading||businessBackupRequestInProgress)return;
  const connectionRevision=maintenanceConnectionRevision;
  const file=getMaintenanceConsoleElement('upload-file').files[0];if(!file)return;
  if(file.size>targetStatus.max_package_bytes)throw new Error('文件超过已配置的备份包上限。');
  uploading=true;updateBusinessBackupSelectionControls();
  try{
    const sample=new Uint8Array(await new Blob([file.slice(0,65536),file.slice(Math.max(0,file.size-65536))]).arrayBuffer());
    const fingerprint=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',sample)),byte=>byte.toString(16).padStart(2,'0')).join('');
    requireUnchangedMaintenanceConnectionForPendingOperation(connectionRevision);
    const storageKey=`local-backup-upload:${targetStatus.site_id}:${targetStatus.environment}:${file.size}:${fingerprint}`;
    let task=null;const previousID=localStorage.getItem(storageKey);
    if(previousID){try{const previous=await maintenanceRequest(`/api/tasks/${previousID}`);if(previous.task.state==='uploading')task=previous.task;}catch(error){requireUnchangedMaintenanceConnectionForPendingOperation(connectionRevision);if(![404,410].includes(error.httpStatus))throw error;}}
    requireUnchangedMaintenanceConnectionForPendingOperation(connectionRevision);
    if(!task){const result=await postMaintenanceRequest('/api/uploads',{size:file.size});task=result.task;localStorage.setItem(storageKey,task.id);}
    let offset=task.uploaded_bytes;
    while(offset<file.size){const chunk=await file.slice(offset,Math.min(offset+targetStatus.chunk_bytes,file.size)).arrayBuffer();const digest=Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',chunk)),byte=>byte.toString(16).padStart(2,'0')).join('');requireUnchangedMaintenanceConnectionForPendingOperation(connectionRevision);const result=await maintenanceRequest(`/api/uploads/${task.id}/chunks`,{method:'PUT',headers:{'Upload-Offset':String(offset),'X-Chunk-SHA256':digest,'Content-Type':'application/octet-stream'},body:chunk});offset=result.task.uploaded_bytes;getMaintenanceConsoleElement('upload-progress').value=offset/file.size*100;getMaintenanceConsoleElement('upload-status').textContent=`已上传 ${formatBackupByteSize(offset)} / ${formatBackupByteSize(file.size)}；任务 ${task.id}`;}
    requireUnchangedMaintenanceConnectionForPendingOperation(connectionRevision);
    await postMaintenanceRequest(`/api/uploads/${task.id}/commit`);localStorage.removeItem(storageKey);showMessage('本地包已上传。请执行“校验包并检查可用恢复方式”，查看范围与校验结果。');
  }finally{if(connectionRevision===maintenanceConnectionRevision){uploading=false;applyPendingAdminBusinessBackupScopeSelectionAndWorkflow();updateBusinessBackupSelectionControls();await refreshMaintenanceTasks();}}
}
async function downloadMaintenanceBackupPackageToLocalComputer(task) {const ticket=await postMaintenanceRequest(`/api/tasks/${task.id}/download-ticket`);const link=document.createElement('a');link.href=ticket.download_path;link.download='';document.body.append(link);link.click();link.remove();showMessage('已发起浏览器下载，请检查电脑中的保存结果。');}
