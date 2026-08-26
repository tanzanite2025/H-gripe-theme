<template>
  <form class="fitment-catalog-search-form" @submit.prevent="emit('submit')">
    <div class="fitment-catalog-search-form__field fitment-catalog-search-form__field--search">
      <label :for="searchInputId">
        {{ t('fitmentCatalog.search.label') }}
      </label>
      <div class="fitment-catalog-search-form__input-wrap">
        <Icon
          name="lucide:search"
          class="fitment-catalog-search-form__input-icon"
          aria-hidden="true"
        />
        <input
          :id="searchInputId"
          :value="searchQuery"
          type="search"
          :placeholder="t('fitmentCatalog.search.placeholder')"
          autocomplete="off"
          @input="handleSearchInput"
        >
        <button
          v-if="searchQuery || year !== null"
          type="button"
          class="fitment-catalog-search-form__clear"
          :aria-label="t('fitmentCatalog.search.clear')"
          :title="t('fitmentCatalog.search.clear')"
          @click="emit('clear')"
        >
          <Icon name="lucide:x-circle" aria-hidden="true" />
        </button>
      </div>
    </div>

    <div class="fitment-catalog-search-form__field fitment-catalog-search-form__field--year">
      <label :for="yearInputId">
        {{ t('fitmentCatalog.search.yearLabel') }}
      </label>
      <input
        :id="yearInputId"
        :value="year ?? ''"
        type="number"
        min="1800"
        max="2200"
        inputmode="numeric"
        :placeholder="t('fitmentCatalog.search.yearPlaceholder')"
        @input="handleYearInput"
      >
    </div>

    <button
      type="submit"
      class="fitment-catalog-search-form__submit"
      :disabled="isSearching"
    >
      <Icon
        :name="isSearching ? 'lucide:loader-circle' : 'lucide:search'"
        :class="{ 'fitment-catalog-search-form__spinner': isSearching }"
        aria-hidden="true"
      />
      <span>
        {{ isSearching ? t('fitmentCatalog.search.searching') : t('fitmentCatalog.search.submit') }}
      </span>
    </button>
  </form>
</template>

<script setup lang="ts">
import { useId } from 'vue'
import { useI18n } from '#imports'

defineProps<{
  searchQuery: string
  year: number | null
  isSearching: boolean
}>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:year': [value: number | null]
  submit: []
  clear: []
}>()

const { t } = useI18n()
const id = useId()
const searchInputId = `fitment-catalog-search-${id}`
const yearInputId = `fitment-catalog-year-${id}`

const handleSearchInput = (event: Event) => {
  emit('update:searchQuery', (event.target as HTMLInputElement).value)
}

const handleYearInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value
  const year = value ? Number(value) : null
  emit('update:year', year !== null && Number.isFinite(year) ? year : null)
}
</script>

<style scoped>
.fitment-catalog-search-form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(8rem, 10rem) auto;
  gap: 0.75rem;
  align-items: end;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.fitment-catalog-search-form__field {
  display: grid;
  min-width: 0;
  gap: 0.35rem;
}

.fitment-catalog-search-form__field label {
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
  font-weight: 800;
}

.fitment-catalog-search-form input {
  width: 100%;
  min-height: 2.65rem;
  box-sizing: border-box;
  border: 1px solid var(--tz-form-control-border, var(--tz-border-subtle));
  border-radius: 0.5rem;
  color: var(--tz-text-primary);
  background: var(--tz-form-control-surface, var(--tz-surface-subtle));
  font: inherit;
  font-size: 0.875rem;
  outline: none;
}

.fitment-catalog-search-form input:focus {
  border-color: var(--tz-action-primary);
  box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.12);
}

.fitment-catalog-search-form__input-wrap {
  position: relative;
}

.fitment-catalog-search-form__input-wrap input {
  padding: 0 2.4rem 0 2.35rem;
}

.fitment-catalog-search-form__input-icon {
  position: absolute;
  top: 50%;
  left: 0.75rem;
  width: 1rem;
  height: 1rem;
  color: var(--tz-text-muted);
  pointer-events: none;
  transform: translateY(-50%);
}

.fitment-catalog-search-form__clear {
  position: absolute;
  top: 50%;
  right: 0.45rem;
  display: inline-grid;
  width: 1.8rem;
  height: 1.8rem;
  place-items: center;
  border: 0;
  color: var(--tz-text-muted);
  background: transparent;
  cursor: pointer;
  transform: translateY(-50%);
}

.fitment-catalog-search-form__clear:hover {
  color: var(--tz-text-primary);
}

.fitment-catalog-search-form__submit {
  display: inline-flex;
  min-height: 2.65rem;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  box-sizing: border-box;
  border: 1px solid var(--tz-action-primary);
  border-radius: 0.5rem;
  padding: 0.5rem 0.9rem;
  color: var(--tz-action-primary-foreground, #fff);
  background: var(--tz-action-primary);
  cursor: pointer;
  font: inherit;
  font-size: 0.82rem;
  font-weight: 800;
  white-space: nowrap;
}

.fitment-catalog-search-form__submit:hover:not(:disabled) {
  border-color: var(--tz-action-primary-hover);
  background: var(--tz-action-primary-hover);
}

.fitment-catalog-search-form__submit:disabled {
  cursor: wait;
  opacity: 0.65;
}

.fitment-catalog-search-form__submit:focus-visible,
.fitment-catalog-search-form__clear:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.fitment-catalog-search-form__spinner {
  animation: fitment-catalog-search-spin 900ms linear infinite;
}

@keyframes fitment-catalog-search-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 640px) {
  .fitment-catalog-search-form {
    grid-template-columns: minmax(0, 1fr) minmax(7rem, 8rem);
  }

  .fitment-catalog-search-form__submit {
    grid-column: 1 / -1;
  }
}
</style>
