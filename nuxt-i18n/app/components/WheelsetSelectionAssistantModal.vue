<template>
  <Teleport to="body">
    <Transition name="wheelset-selection-assistant-modal" appear>
      <div
        v-if="modelValue"
        class="quickbuy-modal-mask fixed inset-0 z-[10040] flex items-center justify-center p-0 md:p-4 tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
        role="presentation"
        @click.self="handleClose"
      >
        <div
          class="absolute inset-0 bg-slate-900/20 backdrop-blur-sm"
          aria-hidden="true"
          @click="handleClose"
        />

        <section
          ref="modalElement"
          class="wheelset-selection-assistant-modal-shell tz-mobile-dialog-surface relative flex h-full w-full max-w-none flex-col overflow-hidden tz-text-primary md:h-[95dvh] md:max-h-[95dvh] md:w-[95vw] md:max-w-[95vw] md:rounded-2xl"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="modalTitleId"
          tabindex="-1"
        >
          <header class="wheelset-selection-assistant-modal__header shrink-0 px-4 py-2 md:px-6 md:py-2.5">
            <div class="wheelset-selection-assistant-modal__header-row">
              <div class="min-w-0">
                <span v-if="eyebrowLabel" class="wheelset-selection-assistant-modal__eyebrow">
                  {{ eyebrowLabel }}
                </span>
                <div class="wheelset-selection-assistant-modal__title-row">
                  <h2 :id="modalTitleId" class="wheelset-selection-assistant-modal__title">
                    {{ titleLabel }}
                  </h2>
                  <QuickBuyLocalizedHelpQuestionMarkDialog
                    v-if="assistantHelp.content.value"
                    :key="assistantHelp.content.value"
                    :title="assistantHelp.title.value || t('quickBuy.help.title', 'Help')"
                    :content="assistantHelp.content.value"
                    :trigger-aria-label="assistantHelp.title.value || t('quickBuy.help.title', 'Help')"
                    :close-label="t('common.close', 'Close')"
                  />
                </div>
                <p v-if="descriptionLabel" class="wheelset-selection-assistant-modal__description">
                  {{ descriptionLabel }}
                </p>
              </div>

              <div class="wheelset-selection-assistant-modal__header-actions">
                <button
                  type="button"
                  class="tz-global-close-btn shrink-0"
                  :aria-label="t('common.close', 'Close')"
                  :title="t('common.close', 'Close')"
                  @click="handleClose"
                >
                  <Icon name="lucide:x" class="h-3.5 w-3.5" />
                </button>
              </div>
            </div>
          </header>

          <div class="wheelset-selection-assistant-modal__body min-h-0 flex-1 overflow-hidden px-0 py-0 md:px-6">
            <slot
              :source="source"
            />
          </div>

          <footer
            v-if="hasFooterSlot"
            class="wheelset-selection-assistant-modal__footer flex shrink-0 items-center justify-between gap-3 px-4 py-3 md:px-6"
          >
            <slot
              name="footer"
              :close="handleClose"
            />
          </footer>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, provide, ref, useSlots, watch } from 'vue'
import { useI18n } from '#imports'
import QuickBuyLocalizedHelpQuestionMarkDialog from '~/components/quick-buy/QuickBuyLocalizedHelpQuestionMarkDialog.vue'
import { provideWheelsetSelectionAssistantHelp } from '~/composables/useWheelsetSelectionAssistantHelp'
import { wheelsetSelectionAssistantQuestionPaginationKey } from '~/composables/useWheelsetSelectionAssistantQuestionPagination'
import { useWheelsetSelectionAssistantModalState } from '~/composables/useWheelsetSelectionAssistantModalState'
import type {
  WheelsetSelectionAssistantSource,
} from '~/types/wheelsetSelectionAssistant'

const props = withDefaults(defineProps<{
  modelValue: boolean
  source?: WheelsetSelectionAssistantSource
  title?: string
  eyebrow?: string
  description?: string | null
}>(), {
  source: 'guides/wheelset-buyers',
  title: '',
  eyebrow: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
}>()

const { t } = useI18n()
const slots = useSlots()
const modalElement = ref<HTMLElement | null>(null)
const modalTitleId = 'wheelset-selection-assistant-modal-title'
const hasFooterSlot = computed(() => Boolean(slots.footer))
const assistantHelp = provideWheelsetSelectionAssistantHelp()
const assistantModalState = useWheelsetSelectionAssistantModalState()
const questionPaginationTotal = ref(0)
const questionPaginationActiveIndex = ref(0)
const questionPaginationReachableIndex = ref(0)
let registeredOpenState = false

const isQuickBuyWheelsetSelectionAssistantSource = computed(() => props.source === 'quick-buy/wheelset-selection-assistant')
const titleLabel = computed(() => props.title || (
  isQuickBuyWheelsetSelectionAssistantSource.value
    ? t('quickBuy.entry.wheelsetSelectionAssistant.title', 'Find my bike specs')
    : t('wheelsetSelectionAssistant.title', 'Wheelset fit helper')
))
const eyebrowLabel = computed(() => props.eyebrow || (
  isQuickBuyWheelsetSelectionAssistantSource.value
    ? t('quickBuy.entry.eyebrow', 'QUICKBUY')
    : t('wheelsetSelectionAssistant.eyebrow', 'Wheelset guide')
))
const descriptionLabel = computed(() => {
  if (props.description !== undefined && props.description !== null) {
    return props.description
  }
  return t(
    isQuickBuyWheelsetSelectionAssistantSource.value
      ? 'quickBuy.entry.wheelsetSelectionAssistant.modalDescription'
      : 'wheelsetSelectionAssistant.description',
    'Choose what you know now. We only ask for missing fit details.',
  )
})

provide(wheelsetSelectionAssistantQuestionPaginationKey, {
  total: questionPaginationTotal,
  activeIndex: questionPaginationActiveIndex,
  reachableIndex: questionPaginationReachableIndex,
})

const handleClose = () => {
  emit('update:modelValue', false)
  emit('close')
}

const handleKeydown = (event: KeyboardEvent) => {
  if (props.modelValue && event.key === 'Escape') {
    handleClose()
  }
}

const syncOpenState = (isOpen: boolean) => {
  if (isOpen && !registeredOpenState) {
    assistantModalState.register()
    registeredOpenState = true
    return
  }

  if (!isOpen && registeredOpenState) {
    assistantModalState.unregister()
    registeredOpenState = false
  }
}

watch(
  () => props.modelValue,
  value => {
    syncOpenState(value)
    if (!value) return
    nextTick(() => modalElement.value?.focus())
  },
)

onMounted(() => {
  syncOpenState(props.modelValue)
  document.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  syncOpenState(false)
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.wheelset-selection-assistant-modal-shell {
  --quickbuy-shell-surface: var(--tz-card-surface);
  --quickbuy-shell-surface-soft: var(--tz-form-panel-surface);
  --quickbuy-panel-surface: var(--tz-card-surface);
  --quickbuy-panel-surface-soft: var(--tz-surface-subtle);
  --quickbuy-panel-surface-raised: var(--tz-surface-muted);
  --quickbuy-control-surface-raised: var(--tz-surface-subtle);
  --quickbuy-divider: var(--tz-border-subtle);
  --quickbuy-divider-strong: var(--tz-border-strong);
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  min-height: 0;
  background: var(--quickbuy-shell-surface);
  box-shadow:
    0 30px 90px rgba(20, 32, 43, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 0 0 1px var(--tz-border-subtle);
}

.wheelset-selection-assistant-modal__header {
  background: var(--quickbuy-panel-surface-raised);
}

.wheelset-selection-assistant-modal__header-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.wheelset-selection-assistant-modal__header-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.5rem;
}

.wheelset-selection-assistant-modal__title-row {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.5rem;
}

.wheelset-selection-assistant-modal__title-row .wheelset-selection-assistant-modal__title {
  min-width: 0;
}

.wheelset-selection-assistant-modal__eyebrow {
  color: rgba(5, 150, 105, 0.82);
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.24em;
  text-transform: uppercase;
}

.wheelset-selection-assistant-modal__title {
  margin-top: 0.25rem;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.25;
}

.wheelset-selection-assistant-modal__description {
  margin-top: 0.25rem;
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  line-height: 1.45;
}

.wheelset-selection-assistant-modal__footer {
  background: var(--quickbuy-panel-surface-soft);
}

.wheelset-selection-assistant-modal__body {
  display: flex;
  min-height: 0;
  box-sizing: border-box;
  overflow: hidden;
  padding-block: 2px;
  background: var(--tz-input-surface);
}

.wheelset-selection-assistant-modal__footer {
  border-top: 1px solid var(--quickbuy-divider);
}

@media (max-width: 767px) {
  .wheelset-selection-assistant-modal__footer {
    padding: 0;
  }
}

.wheelset-selection-assistant-modal-enter-active,
.wheelset-selection-assistant-modal-leave-active {
  transition: opacity 0.3s ease-out;
}

.wheelset-selection-assistant-modal-enter-active section,
.wheelset-selection-assistant-modal-leave-active section {
  transition:
    transform 0.3s ease-out,
    opacity 0.3s ease-out;
}

.wheelset-selection-assistant-modal-enter-from,
.wheelset-selection-assistant-modal-leave-to {
  opacity: 0;
}

.wheelset-selection-assistant-modal-enter-from section,
.wheelset-selection-assistant-modal-leave-to section {
  opacity: 0;
  transform: translateY(100%);
}

.wheelset-selection-assistant-modal-enter-to section,
.wheelset-selection-assistant-modal-leave-from section {
  opacity: 1;
  transform: translateY(0%);
}

@media (max-width: 767px) {
  .wheelset-selection-assistant-modal-shell {
    width: 100%;
    max-width: none;
    height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
    max-height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
  }
}

@media (min-width: 768px) {
  .wheelset-selection-assistant-modal__title {
    font-size: 1.125rem;
  }

  .wheelset-selection-assistant-modal__description {
    font-size: 0.875rem;
  }
}
</style>
