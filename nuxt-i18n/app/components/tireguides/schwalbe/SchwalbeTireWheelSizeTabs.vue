<template>
  <nav class="schwalbe-wheel-size-tabs" :aria-label="label">
    <div class="schwalbe-wheel-size-tabs__desktop">
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

    <label class="schwalbe-wheel-size-tabs__mobile">
      <span class="schwalbe-wheel-size-tabs__sr-only">{{ label }}</span>
      <select
        :value="mobileSelectedValue"
        :aria-label="label"
        @change="onMobileSelectionChange"
      >
        <option v-if="selectedWheelSizeKeys.length > 1" :value="MULTI_SELECTION_VALUE" disabled>
          {{ multipleLabel }}
        </option>
        <option value="">{{ allLabel }}</option>
        <option v-for="option in options" :key="option.value" :value="option.value">
          {{ optionLabel(option) }}
        </option>
      </select>
    </label>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { SchwalbeTireCatalogWheelSizeOption } from '~/data/tireguides/schwalbeTireCatalogFilterModel'

const MULTI_SELECTION_VALUE = '__multiple_wheel_sizes__'

const props = defineProps<{
  label: string
  allLabel: string
  multipleLabel: string
  optionTemplate: string
  options: readonly SchwalbeTireCatalogWheelSizeOption[]
  selectedWheelSizeKeys: readonly string[]
}>()

const emit = defineEmits<{
  select: [wheelSizeKey: string | null]
}>()

const mobileSelectedValue = computed(() => (
  props.selectedWheelSizeKeys.length > 1
    ? MULTI_SELECTION_VALUE
    : props.selectedWheelSizeKeys[0] || ''
))

const optionLabel = (option: SchwalbeTireCatalogWheelSizeOption) => (
  // The template is localized by the page. The value itself remains the
  // stable wheel-diameter + BSD key used by the URL and API.
  props.optionTemplate
    .replace('{diameter}', option.wheelDiameterIn)
    .replace('{bsd}', String(option.beadSeatDiameterMm))
)

const onMobileSelectionChange = (event: Event) => {
  const value = (event.target as HTMLSelectElement | null)?.value
  if (value === undefined || value === MULTI_SELECTION_VALUE) return
  emit('select', value || null)
}
</script>

<style scoped>
.schwalbe-wheel-size-tabs {
  width: 100%;
  min-width: 0;
  overflow: hidden;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.schwalbe-wheel-size-tabs__desktop {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  gap: 0.35rem;
  padding: 0 0.1rem 0.45rem;
}

.schwalbe-wheel-size-tabs__mobile {
  display: none;
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

@media (max-width: 760.5px) {
  .schwalbe-wheel-size-tabs__desktop {
    display: none;
  }

  .schwalbe-wheel-size-tabs__mobile {
    display: block;
    padding: 0.1rem 0 0.45rem;
  }

  .schwalbe-wheel-size-tabs__mobile select {
    width: 100%;
    min-height: 2.5rem;
    border: 1px solid var(--tz-border-strong);
    border-radius: 0.7rem;
    background: var(--tz-card-surface);
    color: var(--tz-text-primary);
    padding: 0.55rem 0.75rem;
    font: inherit;
    font-size: 0.84rem;
    font-weight: 700;
  }

  .schwalbe-wheel-size-tabs__mobile select:focus-visible {
    outline: 2px solid var(--tz-action-primary);
    outline-offset: 2px;
  }
}

.schwalbe-wheel-size-tabs__sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  clip-path: inset(50%);
}
</style>
