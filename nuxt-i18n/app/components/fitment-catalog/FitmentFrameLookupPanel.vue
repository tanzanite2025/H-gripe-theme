<template>
  <section class="fitment-catalog-lookup-panel" :aria-label="t('fitmentCatalog.tabs.frame')">
    <FitmentCatalogSearchForm
      :search-query="searchQuery"
      :year="year"
      :is-searching="isSearching"
      @update:search-query="searchQuery = $event"
      @update:year="year = $event"
      @submit="submit"
      @clear="clear"
    />

    <div
      v-if="isSearching"
      class="fitment-catalog-lookup-panel__state"
      aria-live="polite"
    >
      <Icon name="lucide:loader-circle" class="fitment-catalog-lookup-panel__spinner" aria-hidden="true" />
      <p>{{ t('fitmentCatalog.search.searching') }}</p>
    </div>

    <div
      v-else-if="error"
      class="fitment-catalog-lookup-panel__state fitment-catalog-lookup-panel__state--error"
      role="alert"
    >
      <Icon name="lucide:triangle-alert" aria-hidden="true" />
      <p>{{ error }}</p>
      <button type="button" class="fitment-catalog-lookup-panel__retry" @click="submit">
        <Icon name="lucide:refresh-cw" aria-hidden="true" />
        <span>{{ t('common.retry') }}</span>
      </button>
    </div>

    <div
      v-else-if="!hasSearched"
      class="fitment-catalog-lookup-panel__state"
    >
      <Icon name="lucide:search" aria-hidden="true" />
      <p>{{ t('fitmentCatalog.states.initial') }}</p>
    </div>

    <div
      v-else-if="!entries.length"
      class="fitment-catalog-lookup-panel__state"
    >
      <Icon name="lucide:search-x" aria-hidden="true" />
      <p>{{ t('fitmentCatalog.states.empty') }}</p>
    </div>

    <div v-else class="fitment-catalog-lookup-panel__results">
      <FitmentCatalogEntryCard
        v-for="entry in entries"
        :key="entry.id"
        :entry="entry"
      />
    </div>

    <nav
      v-if="hasSearched && pagination.total_pages > 1"
      class="fitment-catalog-lookup-panel__pagination"
      :aria-label="t('fitmentCatalog.pagination.label')"
    >
      <button
        type="button"
        :disabled="!canGoPrevious"
        :aria-label="t('fitmentCatalog.pagination.previous')"
        :title="t('fitmentCatalog.pagination.previous')"
        @click="goToPage(currentPage - 1)"
      >
        <Icon name="lucide:chevron-left" aria-hidden="true" />
      </button>
      <span>
        {{ t('fitmentCatalog.pagination.page', {
          page: currentPage,
          totalPages: pagination.total_pages,
        }) }}
      </span>
      <button
        type="button"
        :disabled="!canGoNext"
        :aria-label="t('fitmentCatalog.pagination.next')"
        :title="t('fitmentCatalog.pagination.next')"
        @click="goToPage(currentPage + 1)"
      >
        <Icon name="lucide:chevron-right" aria-hidden="true" />
      </button>
    </nav>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import FitmentCatalogEntryCard from '~/components/fitment-catalog/FitmentCatalogEntryCard.vue'
import FitmentCatalogSearchForm from '~/components/fitment-catalog/FitmentCatalogSearchForm.vue'
import { useFitmentFrameLookup } from '~/composables/fitment-catalog/useFitmentFrameLookup'

const { t } = useI18n()
const {
  searchQuery,
  year,
  entries,
  pagination,
  currentPage,
  isSearching,
  error,
  hasSearched,
  canGoPrevious,
  canGoNext,
  submit,
  clear,
  goToPage,
} = useFitmentFrameLookup()
</script>

<style scoped>
.fitment-catalog-lookup-panel {
  display: grid;
  gap: 1rem;
}

.fitment-catalog-lookup-panel__state {
  display: grid;
  min-height: 15rem;
  place-items: center;
  align-content: center;
  gap: 0.65rem;
  color: var(--tz-text-muted);
  text-align: center;
}

.fitment-catalog-lookup-panel__state p {
  max-width: 28rem;
  margin: 0;
  font-size: 0.85rem;
  line-height: 1.55;
}

.fitment-catalog-lookup-panel__state--error {
  color: var(--tz-status-danger-text, #b91c1c);
}

.fitment-catalog-lookup-panel__spinner {
  animation: fitment-frame-lookup-spin 900ms linear infinite;
}

.fitment-catalog-lookup-panel__retry {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  gap: 0.4rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.45rem;
  padding: 0.4rem 0.7rem;
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
  cursor: pointer;
  font: inherit;
  font-size: 0.78rem;
  font-weight: 800;
}

.fitment-catalog-lookup-panel__retry:hover {
  border-color: var(--tz-action-primary);
}

.fitment-catalog-lookup-panel__results {
  display: grid;
  gap: 0.75rem;
}

.fitment-catalog-lookup-panel__pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.85rem;
  padding-top: 0.25rem;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
}

.fitment-catalog-lookup-panel__pagination button {
  display: inline-grid;
  width: 2rem;
  height: 2rem;
  place-items: center;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.4rem;
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
  cursor: pointer;
}

.fitment-catalog-lookup-panel__pagination button:hover:not(:disabled) {
  border-color: var(--tz-action-primary);
  color: var(--tz-action-primary);
}

.fitment-catalog-lookup-panel__pagination button:disabled {
  cursor: not-allowed;
  opacity: 0.4;
}

.fitment-catalog-lookup-panel__pagination button:focus-visible,
.fitment-catalog-lookup-panel__retry:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

@keyframes fitment-frame-lookup-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
