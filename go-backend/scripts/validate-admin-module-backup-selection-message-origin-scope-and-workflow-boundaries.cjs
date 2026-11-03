'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const { test } = require('node:test');

function createAdminModuleBackupSelectionMessageHarness() {
  const parent = {};
  const radios = [{value:'whole',checked:true},{value:'modules',checked:false}];
  const checkboxes = ['orders','customers'].map(value=>({value,checked:false}));
  const actions = [];
  const browserWindow = {parent};
  const context = vm.createContext({window:browserWindow,location:{origin:'http://127.0.0.1:10242'},
    document:{querySelectorAll:selector=>selector.includes('backup-export-mode')?radios:checkboxes},
    targetStatus:{business_module_backup_options:[{id:'orders'},{id:'customers'}]}, uploading:false,
    getMaintenanceConsoleElement:id=>({scrollIntoView(){actions.push('scroll:'+id);},focus(){actions.push('focus:'+id);}}),
    showMessage:message=>actions.push(message)});
  vm.runInContext(fs.readFileSync(path.join(__dirname,'../internal/localbackup/maintenance_console/maintenance-console-business-module-backup-selection-and-range-preview.js'),'utf8'),context);
  vm.runInContext('invalidateSelectedBusinessBackupPreview=()=>{};isBusinessBackupModuleModeSelected=()=>true;previewCurrentSelectedBusinessBackupModules=async()=>{previewCalls++;};',context);
  context.previewCalls=0;
  return {context,radios,checkboxes,actions,receive:data=>context.receiveAdminBackupScopeSelectionOrWorkflowNavigation({origin:context.location.origin,source:parent,data})};
}

test('admin selection synchronizes exact modules and whole scope without starting an export',()=>{
  const h=createAdminModuleBackupSelectionMessageHarness();
  h.receive({type:'local-backup-maintenance-select-scope',mode:'modules',modules:['customers']});
  assert.deepEqual(h.radios.map(r=>r.checked),[false,true]);
  assert.deepEqual(h.checkboxes.map(r=>r.checked),[false,true]);
  assert.equal(h.context.previewCalls,0);
  h.receive({type:'local-backup-maintenance-select-scope',mode:'whole',modules:[]});
  assert.deepEqual(h.radios.map(r=>r.checked),[true,false]);
  assert.deepEqual(h.checkboxes.map(r=>r.checked),[false,false]);
});

for (const invalid of ['foreign_origin','foreign_window','standalone','no_session','unknown_module','duplicate_module','invalid_mode','uploading','exporting']) {
  test('reject unsafe or stale admin selection: '+invalid,()=>{
    const h=createAdminModuleBackupSelectionMessageHarness();
    const event={origin:h.context.location.origin,source:h.context.window.parent,data:{type:'local-backup-maintenance-select-scope',mode:'modules',modules:['customers']}};
    if(invalid==='foreign_origin')event.origin='https://foreign.invalid';
    if(invalid==='foreign_window')event.source={};
    if(invalid==='standalone')h.context.window.parent=h.context.window;
    if(invalid==='no_session')h.context.targetStatus=null;
    if(invalid==='unknown_module')event.data.modules=['unregistered'];
    if(invalid==='duplicate_module')event.data.modules=['customers','customers'];
    if(invalid==='invalid_mode')event.data.mode='restore';
    if(invalid==='uploading')h.context.uploading=true;
    if(invalid==='exporting')vm.runInContext('businessBackupRequestInProgress=true',h.context);
    h.context.receiveAdminBackupScopeSelectionOrWorkflowNavigation(event);
    assert.deepEqual(h.radios.map(r=>r.checked),[true,false]);
    assert.deepEqual(h.checkboxes.map(r=>r.checked),[false,false]);
  });
}

test('workflow navigation only previews or focuses upload; it cannot restore or generate',()=>{
  const h=createAdminModuleBackupSelectionMessageHarness();
  h.receive({type:'local-backup-maintenance-show-workflow',workflow:'export'});
  h.receive({type:'local-backup-maintenance-show-workflow',workflow:'upload'});
  h.receive({type:'local-backup-maintenance-show-workflow',workflow:'restore'});
  assert.equal(h.context.previewCalls,1);
  assert.deepEqual(h.actions,['scroll:business-backup-selection','scroll:upload-form','focus:upload-file']);
});
