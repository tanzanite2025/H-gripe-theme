'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

function readMaintenanceConsoleClassicScriptsInDocumentOrder() {
  const directory = path.join(__dirname, '../internal/localbackup/maintenance_console');
  const html = fs.readFileSync(path.join(directory, 'index.html'), 'utf8');
  const scripts = Array.from(html.matchAll(/<script\b([^>]*)>([\s\S]*?)<\/script>/gi), match => {
    const attributes = match[1];
    const file = /\bsrc="([^"]+)"/.exec(attributes)?.[1];
    assert(file && /^maintenance-console(?:-[a-z-]+)?\.js$/.test(file), 'Only explicit local console scripts are supported');
    assert(/\bdefer\b/.test(attributes), `${file} must wait for the HTML document`);
    assert(!/\b(?:async|type)\s*(?:=|\b)/.test(attributes), `${file} must retain ordered classic-script semantics`);
    assert.equal(match[2].trim(), '', 'Inline scripts are not allowed by the maintenance CSP');
    return { file, source: fs.readFileSync(path.join(directory, file), 'utf8') };
  });
  assert(scripts.length > 1, 'Load the responsibility scripts as well as the event entry');
  assert.equal(scripts.at(-1).file, 'maintenance-console.js', 'Bind events after all responsibilities load');
  assert.equal(new Set(scripts.map(script => script.file)).size, scripts.length, 'Each script must load exactly once');
  return scripts;
}

function loadMaintenanceConsoleScriptsInDocumentOrder(context) {
  for (const { file, source } of readMaintenanceConsoleClassicScriptsInDocumentOrder()) {
    vm.runInContext(source, context, { filename: file });
  }
}

module.exports = { readMaintenanceConsoleClassicScriptsInDocumentOrder, loadMaintenanceConsoleScriptsInDocumentOrder };
