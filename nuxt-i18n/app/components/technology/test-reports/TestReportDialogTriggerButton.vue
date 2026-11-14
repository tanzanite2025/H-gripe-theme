<template>
  <button
    ref="testReportTriggerButtonElement"
    v-bind="$attrs"
    type="button"
    class="test-report-dialog-trigger-button"
    aria-haspopup="dialog"
    :aria-label="typeof $attrs['aria-label'] === 'string' ? $attrs['aria-label'] : ($slots.default ? undefined : testReportTriggerAccessibleLabel)"
    :aria-expanded="isTestReportDialogOpen"
    :data-test-report-trigger="testId"
    @click="openSelectedTestReportDialog"
  >
    <Icon name="lucide:expand" aria-hidden="true" />
    <slot>{{ label || (areDialogMessagesReady ? t('testReportDialog.open') : t('common.loading')) }}</slot>
  </button>
  <TestReportBlockDialog
    v-if="isTestReportDialogOpen"
    :test-id="testId"
    :return-focus-element="testReportTriggerButtonElement"
    @close="isTestReportDialogOpen = false"
  />
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from '#imports'
import TestReportBlockDialog from '~/components/technology/test-reports/TestReportBlockDialog.vue'
import { testReportBlockRegistrationsById } from '~/utils/testReportBlockComponentRegistry'
import { useIndependentTestReportMessages } from '~/composables/useIndependentTestReportMessages'
import type { TestReportBlockId } from '~/types/testReportBlock'

defineOptions({ inheritAttrs: false })
const { testId, label } = defineProps<{ testId: TestReportBlockId; label?: string }>()
const { t } = useI18n()
const { areTestReportMessagesReady: areDialogMessagesReady } = useIndependentTestReportMessages('testReportDialog')
const selectedTestReportRegistration = computed(() => testReportBlockRegistrationsById[testId])
const { areTestReportMessagesReady: areSelectedTestMessagesReady } = useIndependentTestReportMessages(
  () => selectedTestReportRegistration.value.messageNamespace,
)

const testReportTriggerButtonElement = ref<HTMLButtonElement | null>(null)
const isTestReportDialogOpen = ref(false)
const testReportTriggerAccessibleLabel = computed(() => label || (areDialogMessagesReady.value && areSelectedTestMessagesReady.value
  ? t('testReportDialog.openTest', { test: t(selectedTestReportRegistration.value.titleMessageKey) })
  : t('common.loading')))

function openSelectedTestReportDialog(): void {
  isTestReportDialogOpen.value = true
}
</script>

<style scoped>
.test-report-dialog-trigger-button {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-accent);
  font-size: 0.8rem;
  line-height: 1.4;
  cursor: pointer;
}

.test-report-dialog-trigger-button:hover {
  border-color: var(--tz-site-accent);
  background: var(--tz-surface-inset);
}

.test-report-dialog-trigger-button:focus-visible {
  outline: 2px solid var(--tz-site-accent);
  outline-offset: 3px;
}
</style>
