<template>
  <div class="wheelset-selection-two-column-layout">
    <section
      class="wheelset-selection-two-column-layout__pane wheelset-selection-two-column-layout__pane--question"
      :class="{
        'wheelset-selection-two-column-layout__pane--mobile-expanded': isQuestionExpanded,
        'wheelset-selection-two-column-layout__pane--mobile-collapsed': !isQuestionExpanded,
      }"
    >
      <slot name="question" />
    </section>

    <section
      class="wheelset-selection-two-column-layout__pane wheelset-selection-two-column-layout__pane--results"
      :class="{
        'wheelset-selection-two-column-layout__pane--mobile-expanded': isResultsExpanded,
        'wheelset-selection-two-column-layout__pane--mobile-collapsed': !isResultsExpanded,
      }"
    >
      <slot name="results" />
    </section>
  </div>
</template>

<script setup lang="ts">
import { provideWheelsetSelectionAccordion } from '~/composables/useWheelsetSelectionAccordion'

const accordion = provideWheelsetSelectionAccordion('question')
const isQuestionExpanded = accordion.isExpanded('question')
const isResultsExpanded = accordion.isExpanded('results')
</script>

<style scoped>
.wheelset-selection-two-column-layout {
  display: grid;
  width: 100%;
  height: 100%;
  min-height: 100%;
  gap: 1rem;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  overflow: hidden;
}

.wheelset-selection-two-column-layout__pane {
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: 1rem;
  background: var(--quickbuy-panel-surface-soft, #16171d);
}

@media (max-width: 767px) {
  .wheelset-selection-two-column-layout {
    display: flex;
    width: 100%;
    height: 100%;
    flex-direction: column;
    gap: 0.75rem;
    min-height: 0;
    overflow: hidden;
  }

  .wheelset-selection-two-column-layout__pane {
    flex: 0 0 auto;
    min-height: 0;
    overflow: hidden;
  }

  .wheelset-selection-two-column-layout__pane--mobile-expanded {
    order: 1;
    flex: 1 1 0;
    min-height: 0;
  }

  .wheelset-selection-two-column-layout__pane--mobile-collapsed {
    order: 2;
    flex: 0 0 auto;
  }
}
</style>
