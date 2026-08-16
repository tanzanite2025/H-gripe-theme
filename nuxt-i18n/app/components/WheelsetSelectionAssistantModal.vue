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
          class="absolute inset-0 bg-black/80 backdrop-blur-sm"
          aria-hidden="true"
          @click="handleClose"
        />

        <section
          ref="modalElement"
          class="wheelset-selection-assistant-modal-shell tz-mobile-dialog-surface relative flex h-full w-full max-w-none flex-col overflow-hidden text-white md:h-[95dvh] md:max-h-[95dvh] md:w-[95vw] md:max-w-[95vw] md:rounded-2xl"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="modalTitleId"
          tabindex="-1"
        >
          <header class="wheelset-selection-assistant-modal__header flex shrink-0 items-center justify-between gap-4 px-4 py-2 md:px-6 md:py-2.5">
            <div class="min-w-0">
              <span v-if="eyebrowLabel" class="wheelset-selection-assistant-modal__eyebrow">
                {{ eyebrowLabel }}
              </span>
              <h2 :id="modalTitleId" class="wheelset-selection-assistant-modal__title">
                {{ titleLabel }}
              </h2>
              <p v-if="descriptionLabel" class="wheelset-selection-assistant-modal__description">
                {{ descriptionLabel }}
              </p>
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

          <div
            v-if="steps.length > 0"
            class="wheelset-selection-assistant-modal__steps shrink-0 px-4 py-3 md:px-6"
          >
            <div class="wheelset-selection-assistant-modal__step-meta flex items-center justify-between gap-3">
              <span>{{ currentStepLabel }}</span>
              <span>{{ progressLabel }}</span>
            </div>
            <div class="mt-3 grid gap-1" :style="stepTrackStyle">
              <span
                v-for="step in steps"
                :key="step.key"
                class="wheelset-selection-assistant-modal__step-bar"
                :class="{ 'wheelset-selection-assistant-modal__step-bar--active': isStepActiveOrComplete(step.key) }"
                aria-hidden="true"
              />
            </div>
          </div>

          <div class="wheelset-selection-assistant-modal__body min-h-0 flex-1 overflow-y-auto px-4 py-4 md:px-6">
            <slot
              :current-step-key="normalizedCurrentStepKey"
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
              :current-step-key="normalizedCurrentStepKey"
            />
          </footer>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useSlots, watch } from 'vue'
import { useI18n } from '#imports'
import type {
  WheelsetSelectionAssistantShellStep,
  WheelsetSelectionAssistantSource,
  WheelsetSelectionAssistantStepKey,
} from '~/types/wheelsetSelectionAssistant'

const props = withDefaults(defineProps<{
  modelValue: boolean
  source?: WheelsetSelectionAssistantSource
  title?: string
  eyebrow?: string
  description?: string | null
  currentStepKey?: WheelsetSelectionAssistantStepKey | string
  steps?: WheelsetSelectionAssistantShellStep[]
  showSteps?: boolean
}>(), {
  source: 'guides/wheelset-buyers',
  title: '',
  eyebrow: '',
  currentStepKey: 'start',
  steps: () => [],
  showSteps: true,
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

const defaultSteps = computed<WheelsetSelectionAssistantShellStep[]>(() => [
  { key: 'start', label: t('wheelsetSelectionAssistant.steps.start', 'Start') },
  { key: 'knowledge', label: t('wheelsetSelectionAssistant.steps.knowledge', 'Known specs') },
  { key: 'bike-basics', label: t('wheelsetSelectionAssistant.steps.bikeBasics', 'Bike fit') },
  { key: 'summary', label: t('wheelsetSelectionAssistant.steps.summary', 'Summary') },
])
const steps = computed(() => {
  if (!props.showSteps) return []
  return props.steps.length > 0 ? props.steps : defaultSteps.value
})
const normalizedCurrentStepKey = computed(() => props.currentStepKey || steps.value[0]?.key || 'start')
const currentStepIndex = computed(() => {
  const index = steps.value.findIndex(step => step.key === normalizedCurrentStepKey.value)
  return index >= 0 ? index : 0
})
const currentStepLabel = computed(() => steps.value[currentStepIndex.value]?.label || titleLabel.value)
const progressLabel = computed(() => {
  if (steps.value.length === 0) return ''
  return `${currentStepIndex.value + 1}/${steps.value.length}`
})
const stepTrackStyle = computed(() => ({
  gridTemplateColumns: `repeat(${Math.max(steps.value.length, 1)}, minmax(0, 1fr))`,
}))

const isStepActiveOrComplete = (stepKey: string) => {
  const index = steps.value.findIndex(step => step.key === stepKey)
  return index >= 0 && index <= currentStepIndex.value
}

const handleClose = () => {
  emit('update:modelValue', false)
  emit('close')
}

const handleKeydown = (event: KeyboardEvent) => {
  if (props.modelValue && event.key === 'Escape') {
    handleClose()
  }
}

watch(
  () => props.modelValue,
  value => {
    if (!value) return
    nextTick(() => modalElement.value?.focus())
  },
)

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.wheelset-selection-assistant-modal-shell {
  --quickbuy-shell-surface: #17181f;
  --quickbuy-shell-surface-soft: #111218;
  --quickbuy-panel-surface: #1b1c23;
  --quickbuy-panel-surface-soft: #16171d;
  --quickbuy-panel-surface-raised: #22242c;
  --quickbuy-control-surface-raised: #1a1c24;
  --quickbuy-divider: rgba(255, 255, 255, 0.085);
  --quickbuy-divider-strong: rgba(255, 255, 255, 0.12);
  background:
    linear-gradient(
      180deg,
      #24262e 0%,
      #1d1f26 34%,
      var(--quickbuy-shell-surface) 68%,
      var(--quickbuy-shell-surface-soft) 100%
    );
  box-shadow:
    0 30px 90px rgba(0, 0, 0, 0.66),
    inset 0 1px 0 rgba(255, 255, 255, 0.075),
    inset 0 0 0 1px rgba(255, 255, 255, 0.045);
}

.wheelset-selection-assistant-modal__header {
  border-bottom: 1px solid var(--quickbuy-divider);
  background:
    linear-gradient(
      180deg,
      var(--quickbuy-panel-surface-raised),
      var(--tz-card-surface, #111116)
    );
}

.wheelset-selection-assistant-modal__eyebrow {
  color: rgba(181, 255, 109, 0.82);
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.24em;
  text-transform: uppercase;
}

.wheelset-selection-assistant-modal__title {
  margin-top: 0.25rem;
  color: var(--tz-text-primary, #f8fafc);
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.25;
}

.wheelset-selection-assistant-modal__description {
  margin-top: 0.25rem;
  color: var(--tz-text-muted, #94a3b8);
  font-size: 0.75rem;
  line-height: 1.45;
}

.wheelset-selection-assistant-modal__steps,
.wheelset-selection-assistant-modal__footer {
  background: var(--quickbuy-panel-surface-soft);
}

.wheelset-selection-assistant-modal__body {
  background: var(--tz-input-surface, #0b0b0e);
}

.wheelset-selection-assistant-modal__steps {
  border-bottom: 1px solid var(--quickbuy-divider);
}

.wheelset-selection-assistant-modal__step-meta {
  color: var(--tz-text-muted, #94a3b8);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.wheelset-selection-assistant-modal__step-bar {
  height: 0.25rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.14);
}

.wheelset-selection-assistant-modal__step-bar--active {
  background: var(--tz-brand-primary, #b5ff6d);
}

.wheelset-selection-assistant-modal__footer {
  border-top: 1px solid var(--quickbuy-divider);
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
