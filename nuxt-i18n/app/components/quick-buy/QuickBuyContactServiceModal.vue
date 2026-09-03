<template>
  <Teleport to="body">
    <Transition name="quickbuy-contact-service-modal" appear>
      <div
        class="quickbuy-contact-service-modal-mask fixed inset-0 z-[10040] flex items-center justify-center p-2 md:p-4 tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
        role="presentation"
        @click.self="handleClose"
      >
        <div
          class="absolute inset-0 bg-[rgba(20,32,43,0.22)] backdrop-blur-sm"
          aria-hidden="true"
          @click="handleClose"
        />

        <section
          ref="modalElement"
          class="quickbuy-contact-service-modal-shell tz-mobile-dialog-surface tz-surface-card relative flex max-h-[calc(100dvh-1rem)] w-full max-w-6xl flex-col overflow-hidden rounded-2xl tz-text-primary shadow-[0_30px_90px_rgba(20,32,43,0.16)]"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="modalTitleId"
          tabindex="-1"
          @click.stop
        >
          <header class="quickbuy-contact-service-modal__header flex shrink-0 items-center justify-between gap-3 px-4 py-3 md:px-5">
            <div class="min-w-0">
              <span class="quickbuy-contact-service-modal__eyebrow">
                {{ t('quickBuy.entry.eyebrow', 'QUICKBUY') }}
              </span>
              <h2 :id="modalTitleId" class="quickbuy-contact-service-modal__title">
                {{ t('quickBuy.entry.contactService.modalTitle', 'Email or chat with support') }}
              </h2>
            </div>

            <button
              type="button"
              class="tz-global-close-btn shrink-0"
              :aria-label="t('common.close', 'Close')"
              :title="t('common.close', 'Close')"
              @click="handleClose"
            >
              <Icon name="lucide:x" class="h-3.5 w-3.5" />
            </button>
          </header>

          <div class="quickbuy-contact-service-modal__body min-h-0 flex-1 overflow-y-auto p-3 md:p-4">
            <ContactServiceEntry
              chat-source="quick-buy/contact-service"
              @email-open="handleClose"
              @open-chat="handleClose"
            />
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from '#imports'
import ContactServiceEntry from '~/components/company/ContactServiceEntry.vue'
import { createDialogStackId, useDialogStack } from '~/composables/useDialogStack'

const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()
const dialogStack = useDialogStack()
const dialogStackId = createDialogStackId('quickbuy-contact-service-modal')
const modalElement = ref<HTMLElement | null>(null)
const modalTitleId = 'quickbuy-contact-service-modal-title'
let unregisterDialogStack: (() => void) | null = null

const handleClose = () => {
  emit('close')
}

onMounted(() => {
  unregisterDialogStack = dialogStack.register(dialogStackId, () => {
    handleClose()
  }, {
    priority: 10040,
  })
  void nextTick(() => modalElement.value?.focus())
})

onBeforeUnmount(() => {
  unregisterDialogStack?.()
  unregisterDialogStack = null
})
</script>

<style scoped>
.quickbuy-contact-service-modal-enter-active,
.quickbuy-contact-service-modal-leave-active {
  transition: opacity 0.24s ease-out;
}

.quickbuy-contact-service-modal-enter-active section,
.quickbuy-contact-service-modal-leave-active section {
  transition:
    transform 0.24s ease-out,
    opacity 0.24s ease-out;
}

.quickbuy-contact-service-modal-enter-from,
.quickbuy-contact-service-modal-leave-to {
  opacity: 0;
}

.quickbuy-contact-service-modal-enter-from section,
.quickbuy-contact-service-modal-leave-to section {
  opacity: 0;
  transform: translateY(1rem) scale(0.98);
}

.quickbuy-contact-service-modal__header {
  border-bottom: 1px solid var(--tz-border-subtle);
  background: var(--tz-surface-subtle);
}

.quickbuy-contact-service-modal__eyebrow {
  color: rgba(5, 150, 105, 0.82);
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.18em;
  line-height: 1.25;
  text-transform: uppercase;
}

.quickbuy-contact-service-modal__title {
  margin: 0.25rem 0 0;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 750;
  letter-spacing: 0;
  line-height: 1.25;
}

.quickbuy-contact-service-modal__body {
  background: var(--tz-form-panel-surface);
}

.quickbuy-contact-service-modal__body :deep(.contact-service-entry) {
  min-height: min(34rem, calc(100dvh - 8rem));
  box-shadow: none;
}

@media (max-width: 767px) {
  .quickbuy-contact-service-modal-shell {
    max-height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
  }

  .quickbuy-contact-service-modal__body :deep(.contact-service-entry) {
    min-height: 0;
  }
}
</style>
