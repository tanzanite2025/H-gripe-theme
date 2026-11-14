<template>
  <Teleport to="body">
    <dialog
      ref="testReportDialogElement"
      class="test-report-block-dialog"
      :aria-labelledby="`${testReportDialogInstanceId}-title`"
      :data-test-report-dialog="testId"
      @cancel.prevent="closeTestReportDialog"
      @close="closeTestReportDialog"
      @click="closeTestReportDialogWhenBackdropIsClicked"
    >
      <div class="test-report-block-dialog__shell">
        <header class="test-report-block-dialog__header">
          <h2 :id="`${testReportDialogInstanceId}-title`">{{ areTestReportMessagesReady ? t('testReportDialog.title') : t('common.loading') }}</h2>
          <button
            type="button"
            class="tz-global-close-btn test-report-block-dialog__close"
            :aria-label="t('common.close')"
            autofocus
            @click="closeTestReportDialog"
          >×</button>
        </header>
        <div class="test-report-block-dialog__body">
          <NuxtErrorBoundary :key="testId">
            <Suspense>
              <component :is="testReportBlockRegistrationsById[testId].component" heading-tag="h3" presentation="dialog" />
              <template #fallback>
                <TestReportContentLoadingAndRetryStatus />
              </template>
            </Suspense>
            <template #error="{ clearError }">
              <TestReportContentLoadingAndRetryStatus failed @retry="clearError" />
            </template>
          </NuxtErrorBoundary>
        </div>
      </div>
    </dialog>
  </Teleport>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, useId } from 'vue'
import { useI18n } from '#imports'
import { NuxtErrorBoundary } from '#components'
import TestReportContentLoadingAndRetryStatus from '~/components/technology/test-reports/TestReportContentLoadingAndRetryStatus.vue'
import { useIndependentTestReportMessages } from '~/composables/useIndependentTestReportMessages'
import { useDialogStack } from '~/composables/useDialogStack'
import { useOverlayBackStack, type OverlayCloseReason } from '~/composables/useOverlayBackStack'
import { testReportBlockRegistrationsById } from '~/utils/testReportBlockComponentRegistry'
import type { TestReportBlockId } from '~/types/testReportBlock'

const { returnFocusElement } = defineProps<{ testId: TestReportBlockId; returnFocusElement?: HTMLElement | null }>()
const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()
const { areTestReportMessagesReady } = useIndependentTestReportMessages('testReportDialog')

const testReportDialogElement = ref<HTMLDialogElement | null>(null)
const testReportDialogInstanceId = `test-report-dialog-${useId()}`
const dialogStack = useDialogStack()
const overlayBackStack = useOverlayBackStack()
let unregisterTestReportDialogFromKeyboardStack: (() => void) | null = null
let openingTriggerElement: HTMLElement | null = null
let previousDocumentBodyOverflow: string | null = null
let isTestReportDialogClosing = false
let isTestReportDialogUnmounting = false

function releaseTestReportDialogPlaybackScrollAndKeyboardResources(): void {
  testReportDialogElement.value?.querySelectorAll('video').forEach(video => video.pause())
  unregisterTestReportDialogFromKeyboardStack?.()
  unregisterTestReportDialogFromKeyboardStack = null
  if (previousDocumentBodyOverflow !== null) {
    document.body.style.overflow = previousDocumentBodyOverflow
    previousDocumentBodyOverflow = null
  }
  testReportDialogElement.value?.close()
}

function finishClosingTestReportDialog(reason: OverlayCloseReason): void {
  if (isTestReportDialogClosing) return
  isTestReportDialogClosing = true
  releaseTestReportDialogPlaybackScrollAndKeyboardResources()
  if (isTestReportDialogUnmounting) return
  emit('close')
  if (reason !== 'navigate' && openingTriggerElement?.isConnected) {
    openingTriggerElement.focus({ preventScroll: true })
  }
}

function closeTestReportDialog(): void {
  if (isTestReportDialogClosing) return
  if (overlayBackStack.isActive(testReportDialogInstanceId)) {
    void overlayBackStack.close(testReportDialogInstanceId)
  } else {
    finishClosingTestReportDialog('user')
  }
}

function closeTestReportDialogWhenBackdropIsClicked(event: MouseEvent): void {
  if (event.target === event.currentTarget) closeTestReportDialog()
}

onMounted(() => {
  const dialogElement = testReportDialogElement.value
  if (!dialogElement) return
  openingTriggerElement = returnFocusElement
    || (document.activeElement instanceof HTMLElement ? document.activeElement : null)
  previousDocumentBodyOverflow = document.body.style.overflow
  document.body.style.overflow = 'hidden'
  dialogElement.showModal()
  unregisterTestReportDialogFromKeyboardStack = dialogStack.register(
    testReportDialogInstanceId,
    closeTestReportDialog,
    { priority: 14000 },
  )
  overlayBackStack.open(testReportDialogInstanceId, finishClosingTestReportDialog, { mode: 'push' })
})

onBeforeUnmount(() => {
  isTestReportDialogUnmounting = true
  if (overlayBackStack.isActive(testReportDialogInstanceId)) {
    // 路由跳转由全局弹层栈提前清理；仍然活跃说明是外部 v-if 卸载，
    // 必须按正常关闭消耗弹窗历史记录，避免浏览器返回时多停留一次。
    void overlayBackStack.close(testReportDialogInstanceId)
  } else {
    releaseTestReportDialogPlaybackScrollAndKeyboardResources()
  }
})
</script>

<style scoped>
.test-report-block-dialog {
  --test-report-dialog-max-height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - 2rem);
  position: fixed;
  inset: 0;
  width: min(1400px, calc(100% - 2rem));
  max-width: none;
  max-height: var(--test-report-dialog-max-height);
  margin: auto;
  padding: 0;
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-primary);
}

.test-report-block-dialog::backdrop {
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(4px);
}

.test-report-block-dialog__shell {
  display: flex;
  max-height: var(--test-report-dialog-max-height);
  flex-direction: column;
}

.test-report-block-dialog__header {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.8rem 1.25rem;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.test-report-block-dialog__header h2 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.test-report-block-dialog__close {
  position: static;
  flex: 0 0 auto;
}

.test-report-block-dialog__body {
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
}

@media (max-width: 767px) {
  .test-report-block-dialog {
    --test-report-dialog-max-height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - env(safe-area-inset-top, 0px) - env(safe-area-inset-bottom, 0px) - 0.5rem);
    width: calc(100% - env(safe-area-inset-left, 0px) - env(safe-area-inset-right, 0px) - 0.5rem);
    border-radius: 0.75rem;
  }
}
</style>
