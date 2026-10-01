<template>
  <nav class="schwalbe-wheel-size-tabs" :aria-label="label">
    <div class="schwalbe-wheel-size-tabs__scroller">
      <button
        type="button"
        class="schwalbe-wheel-size-tabs__tab"
        :class="{ 'schwalbe-wheel-size-tabs__tab--active': selectedWheelSizeKeys.length === 0 }"
        :aria-pressed="selectedWheelSizeKeys.length === 0"
        @click="emit('select', null)"
      >
        {{ allLabel }}
      </button>
      <button
        v-for="option in options"
        :key="option.value"
        type="button"
        class="schwalbe-wheel-size-tabs__tab"
        :class="{ 'schwalbe-wheel-size-tabs__tab--active': selectedWheelSizeKeys.includes(option.value) }"
        :aria-pressed="selectedWheelSizeKeys.includes(option.value)"
        @click="emit('select', option.value)"
      >
        {{ optionLabel(option) }}
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
import type { SchwalbeTireCatalogWheelSizeOption } from '~/data/tireguides/schwalbeTireCatalogFilterModel'

const props = defineProps<{
  label: string
  allLabel: string
  optionTemplate: string
  options: readonly SchwalbeTireCatalogWheelSizeOption[]
  selectedWheelSizeKeys: readonly string[]
}>()

const emit = defineEmits<{
  select: [wheelSizeKey: string | null]
}>()

const optionLabel = (option: SchwalbeTireCatalogWheelSizeOption) => (
  // The template is localized by the page. The value itself remains the
  // stable wheel-diameter + BSD key used by the URL and API.
  props.optionTemplate
    .replace('{diameter}', option.wheelDiameterIn)
    .replace('{bsd}', String(option.beadSeatDiameterMm))
)
</script>

<style scoped>
.schwalbe-wheel-size-tabs {
  width: 100%;
  min-width: 0;
  overflow: hidden;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.schwalbe-wheel-size-tabs__scroller {
  display: flex;
  min-width: max-content;
  gap: 0.35rem;
  overflow-x: auto;
  padding: 0 0.1rem 0.45rem;
  scrollbar-width: thin;
  scrollbar-color: var(--tz-border-strong) transparent;
}

.schwalbe-wheel-size-tabs__tab {
  min-height: 2.35rem;
  flex: 0 0 auto;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 999px;
  background: var(--tz-card-surface);
  color: var(--tz-text-secondary);
  padding: 0.45rem 0.8rem;
  font: inherit;
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
  cursor: pointer;
}

.schwalbe-wheel-size-tabs__tab:hover {
  border-color: var(--tz-border-strong);
  color: var(--tz-text-primary);
}

.schwalbe-wheel-size-tabs__tab:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.schwalbe-wheel-size-tabs__tab--active {
  border-color: var(--tz-action-primary);
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
}

@media (max-width: 480px) {
  .schwalbe-wheel-size-tabs__scroller {
    gap: 0.3rem;
    padding-bottom: 0.4rem;
  }

  .schwalbe-wheel-size-tabs__tab {
    min-height: 2.2rem;
    padding-inline: 0.7rem;
    font-size: 0.74rem;
  }
}
</style>
