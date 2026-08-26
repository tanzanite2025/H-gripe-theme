<template>
  <div
    class="wheelset-selection-question-panel"
    :class="{ 'wheelset-selection-question-panel--mobile-expanded': isMobileQuestionExpanded }"
  >
    <div class="wheelset-selection-question-panel__header">
      <button
        v-if="canGoBack"
        type="button"
        class="wheelset-selection-question-panel__back"
        :disabled="isSelecting"
        @click="emit('back')"
      >
        <Icon name="lucide:arrow-left" class="h-4 w-4" />
        <span>{{ t('wheelsetSelectionAssistant.question.back') }}</span>
      </button>
      <div class="wheelset-selection-question-panel__title-row">
        <button
          type="button"
          class="wheelset-selection-question-panel__mobile-toggle"
          :aria-expanded="isMobileQuestionExpanded"
          :aria-controls="mobileQuestionContentId"
          @click="toggleMobileQuestion"
        >
          <span>{{ question.prompt }}</span>
          <Icon
            name="lucide:chevron-down"
            class="wheelset-selection-question-panel__toggle-icon"
            aria-hidden="true"
          />
        </button>
        <h3 class="wheelset-selection-question-panel__desktop-title">{{ question.prompt }}</h3>
      </div>
    </div>

    <div
      :id="mobileQuestionContentId"
      class="wheelset-selection-question-panel__content"
      :class="{ 'wheelset-selection-question-panel__content--collapsed': !isMobileQuestionExpanded }"
    >
      <div class="wheelset-selection-question-panel__options">
        <button
          v-for="option in question.options"
          :key="option.value"
          type="button"
          class="wheelset-selection-question-panel__option"
          :class="{ 'wheelset-selection-question-panel__option--selected': option.value === selectedValue }"
          :disabled="isSelecting"
          @click="emit('select', option.value)"
        >
          <span class="wheelset-selection-question-panel__option-label">
            <span class="wheelset-selection-question-panel__option-indicator" aria-hidden="true">
              <span class="wheelset-selection-question-panel__option-indicator-dot" />
            </span>
            <span>{{ option.label }}</span>
          </span>
          <small v-if="option.description">{{ option.description }}</small>
        </button>
      </div>

      <slot name="after-options" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from '#imports'
import { useWheelsetSelectionAccordion } from '~/composables/useWheelsetSelectionAccordion'
import type { WheelsetSelectionQuestion } from '~/types/wheelsetSelectionAssistant'

withDefaults(defineProps<{
  question: WheelsetSelectionQuestion
  selectedValue?: string
  canGoBack?: boolean
  isSelecting?: boolean
}>(), {
  canGoBack: false,
  isSelecting: false,
})

const { t } = useI18n()
const accordion = useWheelsetSelectionAccordion()
const localMobileQuestionExpanded = ref(true)
const isMobileQuestionExpanded = computed(() => (
  accordion?.isExpanded('question').value ?? localMobileQuestionExpanded.value
))
const mobileQuestionContentId = 'wheelset-selection-question-content'

const toggleMobileQuestion = () => {
  if (accordion) {
    accordion.toggle('question')
    return
  }

  localMobileQuestionExpanded.value = !localMobileQuestionExpanded.value
}

const emit = defineEmits<{
  select: [value: string]
  back: []
}>()
</script>

<style scoped>
.wheelset-selection-question-panel {
  display: flex;
  width: 100%;
  height: auto;
  min-height: 0;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 1rem;
  box-sizing: border-box;
  padding: 1rem;
  overflow: hidden;
}

.wheelset-selection-question-panel__header {
  display: grid;
  gap: 0.35rem;
}

.wheelset-selection-question-panel__title-row {
  display: flex;
  align-items: center;
  gap: 0.55rem;
}

.wheelset-selection-question-panel__back {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  gap: 0.35rem;
  color: var(--tz-text-muted, #94a3b8);
  font-size: 0.75rem;
  font-weight: 800;
}

.wheelset-selection-question-panel__desktop-title {
  min-width: 0;
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.wheelset-selection-question-panel__mobile-toggle {
  display: none;
}

.wheelset-selection-question-panel__toggle-icon {
  color: var(--tz-text-secondary, #cbd5e1);
  transition: transform 160ms ease;
}

.wheelset-selection-question-panel--mobile-expanded
  .wheelset-selection-question-panel__toggle-icon {
  transform: rotate(180deg);
}

.wheelset-selection-question-panel__content {
  display: flex;
  width: 100%;
  min-height: 0;
  flex: 1 1 0;
  flex-direction: column;
  gap: 1rem;
  overflow: hidden;
}

.wheelset-selection-question-panel__options {
  display: grid;
  width: 100%;
  min-height: 0;
  flex: 1 1 0;
  gap: 0.65rem;
  align-content: start;
  grid-auto-rows: max-content;
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.wheelset-selection-question-panel__option {
  display: grid;
  gap: 0.2rem;
  min-height: 0;
  box-sizing: border-box;
  border: 1px solid var(--quickbuy-divider-strong, var(--tz-border-strong));
  border-radius: 0.75rem;
  background: var(--quickbuy-panel-surface-raised, var(--tz-surface-muted));
  padding: 5px 0.75rem;
  text-align: left;
  transition: background 160ms ease, border-color 160ms ease, transform 160ms ease;
}

.wheelset-selection-question-panel__option-label {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.65rem;
}

.wheelset-selection-question-panel__option-indicator {
  display: grid;
  width: 1rem;
  height: 1rem;
  flex: 0 0 auto;
  place-items: center;
  border: 2px solid var(--tz-border-strong);
  border-radius: 999px;
  background: transparent;
  transition: border-color 160ms ease, background-color 160ms ease;
}

.wheelset-selection-question-panel__option-indicator-dot {
  width: 0.45rem;
  height: 0.45rem;
  border-radius: 999px;
  background: var(--tz-site-accent, #059669);
  opacity: 0;
  transform: scale(0.55);
  transition: opacity 160ms ease, transform 160ms ease;
}

.wheelset-selection-question-panel__option--selected
  .wheelset-selection-question-panel__option-indicator {
  border-color: var(--tz-site-accent, #059669);
  background: rgba(5, 150, 105, 0.08);
}

.wheelset-selection-question-panel__option--selected
  .wheelset-selection-question-panel__option-indicator-dot {
  opacity: 1;
  transform: scale(1);
}

.wheelset-selection-question-panel__option:hover {
  border-color: rgba(5, 150, 105, 0.45);
}

.wheelset-selection-question-panel__option--selected {
  border-color: rgba(5, 150, 105, 0.75);
  background: color-mix(in srgb, var(--quickbuy-panel-surface-raised, var(--tz-surface-muted)) 82%, var(--tz-site-accent, #059669));
}

.wheelset-selection-question-panel__option-label > span:last-child {
  min-width: 0;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 750;
  line-height: 1.3;
  overflow-wrap: anywhere;
  white-space: normal;
}

.wheelset-selection-question-panel__option small {
  min-width: 0;
  color: var(--tz-text-muted, #94a3b8);
  font-size: 0.75rem;
  line-height: 1.35;
  overflow-wrap: anywhere;
  white-space: normal;
}

@media (max-width: 767px) {
  .wheelset-selection-question-panel {
    min-height: 0;
    gap: 0.75rem;
    padding: 0.75rem;
  }

  .wheelset-selection-question-panel__back {
    display: none;
  }

  .wheelset-selection-question-panel--mobile-expanded
    .wheelset-selection-question-panel__back {
    display: inline-flex;
  }

  .wheelset-selection-question-panel__title-row {
    gap: 0.5rem;
  }

  .wheelset-selection-question-panel__mobile-toggle {
    display: flex;
    min-width: 0;
    flex: 1 1 auto;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.75rem;
    border: 1px solid var(--quickbuy-divider-strong, var(--tz-border-strong));
    border-radius: 0.75rem;
    background: var(--quickbuy-panel-surface-raised, var(--tz-surface-muted));
    padding: 0.65rem 0.75rem;
    text-align: left;
  }

  .wheelset-selection-question-panel__mobile-toggle span {
    min-width: 0;
    color: var(--tz-text-primary);
    font-size: 0.95rem;
    font-weight: 800;
    line-height: 1.3;
    overflow-wrap: anywhere;
    white-space: normal;
  }

  .wheelset-selection-question-panel__mobile-toggle .wheelset-selection-question-panel__toggle-icon {
    flex: 0 0 auto;
    margin-top: 0.1rem;
  }

  .wheelset-selection-question-panel__desktop-title {
    display: none;
  }

  .wheelset-selection-question-panel__content--collapsed {
    display: none;
  }
}
</style>
