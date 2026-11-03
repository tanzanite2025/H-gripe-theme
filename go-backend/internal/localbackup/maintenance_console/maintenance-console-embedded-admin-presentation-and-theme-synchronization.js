'use strict';

function synchronizeEmbeddedMaintenanceConsolePresentationWithAdminTheme() {
  const maintenanceRoot = document.documentElement;
  if (!maintenanceRoot || window.parent === window) return;
  maintenanceRoot.classList.add('maintenance-console-embedded');
  try {
    const adminRoot = window.parent.document.documentElement;
    const synchronizeTheme = () => maintenanceRoot.classList.toggle('dark', adminRoot.classList.contains('dark'));
    synchronizeTheme();
    const themeObserver = new MutationObserver(synchronizeTheme);
    themeObserver.observe(adminRoot, {attributes: true, attributeFilter: ['class']});
    window.addEventListener('pagehide', () => themeObserver.disconnect(), {once: true});
  } catch (_) {
    // Keep the independent page's default theme when its parent is unavailable.
  }
}

synchronizeEmbeddedMaintenanceConsolePresentationWithAdminTheme();
