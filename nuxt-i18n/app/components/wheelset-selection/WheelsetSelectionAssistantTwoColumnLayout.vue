<template>
  <div class="wheelset-selection-two-column-layout">
    <section class="wheelset-selection-two-column-layout__pane wheelset-selection-two-column-layout__pane--question">
      <slot name="question" />
    </section>

    <section class="wheelset-selection-two-column-layout__pane wheelset-selection-two-column-layout__pane--results">
      <slot name="results" />
    </section>
  </div>
</template>

<style scoped>
.wheelset-selection-two-column-layout {
  display: grid;
  min-height: 100%;
  gap: 1rem;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
}

.wheelset-selection-two-column-layout__pane {
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  border-radius: 1rem;
  background: var(--quickbuy-panel-surface-soft, #16171d);
}

.wheelset-selection-two-column-layout__pane--question,
.wheelset-selection-two-column-layout__pane--results {
  display: flex;
  flex-direction: column;
}

@media (max-width: 767px) {
  .wheelset-selection-two-column-layout {
    gap: 0.75rem;
    grid-template-columns: 1fr;
  }

  .wheelset-selection-two-column-layout__pane {
    min-height: 16rem;
  }
}
</style>
