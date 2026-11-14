<template>
  <div
    class="test-report-content-loading-status"
    :role="failed ? 'alert' : 'status'"
    :data-test-report-loading-state="failed ? 'error' : 'loading'"
  >
    <p>{{ t(failed ? 'testReportDialog.loadError' : 'testReportDialog.loading') }}</p>
    <button v-if="failed" type="button" @click="retryLoadingFailedTestReportContent">{{ t('testReportDialog.retry') }}</button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import { useIndependentTestReportMessages } from '~/composables/useIndependentTestReportMessages'

defineProps<{ failed?: boolean }>()
const emit = defineEmits<{ retry: [] }>()
const { t } = useI18n()
useIndependentTestReportMessages('testReportDialog')

function retryLoadingFailedTestReportContent(event: MouseEvent): void {
  // 重试会移除当前按钮；先把焦点交还给弹窗，避免落到其外部的 body。
  const retryButton = event.currentTarget as HTMLButtonElement
  retryButton.closest('dialog')?.querySelector<HTMLElement>('[autofocus]')?.focus({ preventScroll: true })
  emit('retry')
}
</script>

<style scoped>
.test-report-content-loading-status {
  padding: 2rem;
  color: var(--tz-text-secondary);
  text-align: center;
}

.test-report-content-loading-status p {
  margin: 0;
}

.test-report-content-loading-status button {
  min-height: 2.75rem;
  margin-top: 1rem;
  padding: 0.5rem 1rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-accent);
  cursor: pointer;
}

.test-report-content-loading-status button:focus-visible {
  outline: 2px solid var(--tz-site-accent);
  outline-offset: 3px;
}
</style>
