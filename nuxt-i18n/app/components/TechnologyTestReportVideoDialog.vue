<template>
  <Teleport to="body">
    <dialog
      ref="testReportVideoDialogElement"
      class="technology-test-report-video-dialog"
      :aria-label="dialogLabel"
      @cancel.prevent="emit('close')"
      @click="closeTestReportVideoDialogWhenBackdropIsClicked"
    >
      <div class="technology-test-report-video-dialog__content">
        <button
          type="button"
          class="tz-global-close-btn technology-test-report-video-dialog__close"
          :aria-label="t('common.close')"
          autofocus
          @click="emit('close')"
        >×</button>
        <video :src="videoSourcePath" class="technology-test-report-video-dialog__video" controls playsinline preload="metadata" />
      </div>
    </dialog>
  </Teleport>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from '#imports'

defineProps<{ videoSourcePath: string; dialogLabel: string }>()
const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()
const testReportVideoDialogElement = ref<HTMLDialogElement | null>(null)

onMounted(() => { testReportVideoDialogElement.value?.showModal() })

function closeTestReportVideoDialogWhenBackdropIsClicked(event: MouseEvent): void {
  if (event.target === event.currentTarget) emit('close')
}
</script>

<style scoped>
.technology-test-report-video-dialog {
  width: min(960px, calc(100% - 2rem));
  max-height: calc(100dvh - 2rem);
  margin: auto;
  padding: 0;
  border: 0;
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
}

.technology-test-report-video-dialog::backdrop {
  background: rgba(15, 23, 42, 0.85);
}

.technology-test-report-video-dialog__content {
  position: relative;
}

.technology-test-report-video-dialog__close {
  position: absolute;
  top: 0.5rem;
  right: 1rem;
  z-index: 1;
}

.technology-test-report-video-dialog__video {
  display: block;
  width: 100%;
  max-height: calc(100dvh - 2rem);
}
</style>
