<template>
  <div class="wheelset-selection-question-panel">
    <div class="wheelset-selection-question-panel__copy">
      <button
        v-if="canGoBack"
        type="button"
        class="wheelset-selection-question-panel__back"
        @click="emit('back')"
      >
        <Icon name="lucide:arrow-left" class="h-4 w-4" />
        <span>Back</span>
      </button>
      <h3>{{ question.prompt }}</h3>
      <p v-if="question.helper">{{ question.helper }}</p>
    </div>

    <div class="wheelset-selection-question-panel__options">
      <button
        v-for="option in question.options"
        :key="option.value"
        type="button"
        class="wheelset-selection-question-panel__option"
        :class="{ 'wheelset-selection-question-panel__option--selected': option.value === selectedValue }"
        @click="emit('select', option.value)"
      >
        <span>{{ option.label }}</span>
        <small v-if="option.description">{{ option.description }}</small>
      </button>
    </div>

    <slot name="after-options" />
  </div>
</template>

<script setup lang="ts">
import type { WheelsetSelectionQuestion } from '~/types/wheelsetSelectionAssistant'

withDefaults(defineProps<{
  question: WheelsetSelectionQuestion
  selectedValue?: string
  canGoBack?: boolean
}>(), {
  canGoBack: false,
})

const emit = defineEmits<{
  select: [value: string]
  back: []
}>()
</script>

<style scoped>
.wheelset-selection-question-panel {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
}

.wheelset-selection-question-panel__copy {
  display: grid;
  gap: 0.35rem;
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

.wheelset-selection-question-panel__copy h3 {
  color: var(--tz-text-primary, #f8fafc);
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.35;
}

.wheelset-selection-question-panel__copy p {
  color: var(--tz-text-muted, #94a3b8);
  font-size: 0.875rem;
  line-height: 1.5;
}

.wheelset-selection-question-panel__options {
  display: grid;
  gap: 0.65rem;
}

.wheelset-selection-question-panel__option {
  display: grid;
  gap: 0.2rem;
  min-height: 3.5rem;
  border: 1px solid var(--quickbuy-divider-strong, rgba(255, 255, 255, 0.12));
  border-radius: 0.75rem;
  background: var(--quickbuy-panel-surface-raised, #22242c);
  padding: 0.75rem;
  text-align: left;
  transition: background 160ms ease, border-color 160ms ease, transform 160ms ease;
}

.wheelset-selection-question-panel__option:hover {
  border-color: rgba(181, 255, 109, 0.45);
  transform: translateY(-1px);
}

.wheelset-selection-question-panel__option--selected {
  border-color: rgba(181, 255, 109, 0.75);
  background: color-mix(in srgb, var(--quickbuy-panel-surface-raised, #22242c) 82%, var(--tz-brand-primary, #b5ff6d));
}

.wheelset-selection-question-panel__option span {
  color: var(--tz-text-primary, #f8fafc);
  font-size: 0.95rem;
  font-weight: 750;
}

.wheelset-selection-question-panel__option small {
  color: var(--tz-text-muted, #94a3b8);
  font-size: 0.75rem;
  line-height: 1.35;
}
</style>
