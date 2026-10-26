<template>
  <footer class="admin-bottom-utility-dock">
    <nav class="admin-bottom-utility-dock__menu" :aria-label="t('layout.utilityDock')">
      <slot />
    </nav>
  </footer>
</template>

<script setup lang="ts">
import { useAdminI18n } from '@/i18n'

const { t } = useAdminI18n()
</script>

<style scoped>
.admin-bottom-utility-dock {
  z-index: 30;
  display: flex;
  flex-shrink: 0;
  justify-content: center;
  padding: 0.625rem 0.75rem max(0.75rem, env(safe-area-inset-bottom));
  border-top: 1px dashed var(--border);
  background: var(--card);
}

.admin-bottom-utility-dock__menu {
  display: flex;
  max-width: 100%;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
}
</style>
