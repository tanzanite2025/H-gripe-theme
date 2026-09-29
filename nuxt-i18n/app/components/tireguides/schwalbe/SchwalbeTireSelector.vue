<template>
  <section class="schwalbe-selector" aria-labelledby="schwalbe-selector-title">
    <div class="schwalbe-selector__intro">
      <p class="schwalbe-selector__kicker">{{ tx('kicker') }}</p>
      <h1 id="schwalbe-selector-title" class="schwalbe-selector__title">
        {{ tx('title') }}
      </h1>
    </div>

    <SchwalbeTelemetryGuide />

    <form class="schwalbe-selector__controls" role="search" @submit.prevent="submitSearch">
      <label class="schwalbe-selector__search">
        <span class="schwalbe-selector__label">{{ tx('search.label') }}</span>
        <span class="schwalbe-selector__search-row">
          <input
            v-model="searchInput"
            type="search"
            :placeholder="tx('search.placeholder')"
            :aria-describedby="'schwalbe-search-help'"
          >
          <button type="submit" class="schwalbe-selector__button">
            {{ tx('search.submit') }}
          </button>
        </span>
        <span id="schwalbe-search-help" class="schwalbe-selector__hint">
          {{ tx('search.hint') }}
        </span>
      </label>

      <div class="schwalbe-selector__selects">
        <label>
          <span class="schwalbe-selector__label">{{ tx('filters.model') }}</span>
          <select v-model="selectedModel">
            <option v-for="model in modelOptions" :key="model" :value="model">
              {{ model === 'ALL' ? tx('filters.allModels') : model }}
            </option>
          </select>
        </label>
        <label>
          <span class="schwalbe-selector__label">{{ tx('filters.sort') }}</span>
          <select v-model="sortBy">
            <option value="model">{{ tx('filters.sortModel') }}</option>
            <option value="etrto">{{ tx('filters.sortEtrto') }}</option>
            <option value="article">{{ tx('filters.sortArticle') }}</option>
          </select>
        </label>
        <button
          v-if="submittedSearch"
          type="button"
          class="schwalbe-selector__clear"
          @click="clearSearch"
        >
          {{ tx('search.clear') }}
        </button>
      </div>
    </form>

    <SchwalbeTireCatalogFilterPanel
      v-model:selected-tire-widths-mm="selectedTireWidthsMm"
      v-model:selected-bead-seat-diameters-mm="selectedBeadSeatDiametersMm"
      :label="tx('filters.dimensionFilters')"
      :tire-width-label="tx('filters.tireWidth')"
      :bead-seat-diameter-label="tx('filters.beadSeatDiameter')"
      :scroll-hint="tx('filters.scrollHint')"
      :reset-label="tx('filters.clearDimensions')"
      :tire-width-options="tireWidthOptions"
      :bead-seat-diameter-options="beadSeatDiameterOptions"
    />

    <div class="schwalbe-selector__summary" aria-live="polite">
      <span>{{ tx('summary', { count: totalItems, page: currentPage, totalPages }) }}</span>
      <span v-if="submittedSearch">{{ tx('search.active', { term: submittedSearch }) }}</span>
    </div>

    <div v-if="pending" class="schwalbe-selector__state" role="status">
      {{ tx('states.loading') }}
    </div>
    <div v-else-if="error" class="schwalbe-selector__state schwalbe-selector__state--error" role="alert">
      <p>{{ tx('states.error') }}</p>
      <button type="button" class="schwalbe-selector__button" @click="() => refresh()">
        {{ tx('states.retry') }}
      </button>
    </div>
    <div v-else-if="totalItems === 0" class="schwalbe-selector__state">
      {{ submittedSearch ? tx('states.noResults') : hasActiveFilters ? tx('states.noFilterResults') : tx('states.empty') }}
    </div>
    <div v-else class="schwalbe-selector__grid">
      <SchwalbeTireCard v-for="item in visibleItems" :key="item.article_no" :item="item" />
    </div>

    <nav
      v-if="!pending && !error && totalPages > 1"
      class="schwalbe-selector__pagination"
      :aria-label="tx('pagination.label')"
    >
      <NuxtLink
        v-if="currentPage > 1"
        class="schwalbe-selector__page-link"
        :to="pageQuery(currentPage - 1)"
        rel="prev"
        :aria-label="tx('pagination.previous')"
      >
        ‹
      </NuxtLink>
      <span v-else class="schwalbe-selector__page-link schwalbe-selector__page-link--disabled" aria-disabled="true">‹</span>

      <template v-for="(pageToken, tokenIndex) in paginationPages" :key="`page-${tokenIndex}-${pageToken}`">
        <span v-if="pageToken === 'ellipsis'" class="schwalbe-selector__page-ellipsis" aria-hidden="true">…</span>
        <NuxtLink
          v-else
          class="schwalbe-selector__page-link"
          :class="{ 'schwalbe-selector__page-link--current': pageToken === currentPage }"
          :to="pageQuery(pageToken)"
          :aria-current="pageToken === currentPage ? 'page' : undefined"
        >
          {{ pageToken }}
        </NuxtLink>
      </template>

      <NuxtLink
        v-if="currentPage < totalPages"
        class="schwalbe-selector__page-link"
        :to="pageQuery(currentPage + 1)"
        rel="next"
        :aria-label="tx('pagination.next')"
      >
        ›
      </NuxtLink>
      <span v-else class="schwalbe-selector__page-link schwalbe-selector__page-link--disabled" aria-disabled="true">›</span>
    </nav>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import SchwalbeTireCard from '~/components/tireguides/schwalbe/SchwalbeTireCard.vue'
import SchwalbeTireCatalogFilterPanel from '~/components/tireguides/schwalbe/SchwalbeTireCatalogFilterPanel.vue'
import SchwalbeTelemetryGuide from '~/components/tireguides/schwalbe/SchwalbeTelemetryGuide.vue'
import { useSchwalbeTireSelector } from '~/composables/useSchwalbeTireSelector'

const { t: translate } = useI18n()
const tx = (key: string, params?: Record<string, unknown>) => translate(`guidesSchwalbeTireSelector.${key}`, params || {})
const {
  searchInput,
  submittedSearch,
  filteredItems,
  totalItems,
  selectedModel,
  selectedTireWidthsMm,
  selectedBeadSeatDiametersMm,
  sortBy,
  visibleItems,
  modelOptions,
  tireWidthOptions,
  beadSeatDiameterOptions,
  hasActiveFilters,
  currentPage,
  totalPages,
  paginationPages,
  pageQuery,
  pending,
  error,
  refresh,
  submitSearch,
  clearSearch,
} = await useSchwalbeTireSelector()
</script>

<style scoped>
.schwalbe-selector {
  display: grid;
  gap: 1.25rem;
  width: 100%;
}

.schwalbe-selector__intro {
  max-width: 60rem;
  margin: 0 auto;
  text-align: center;
}

.schwalbe-selector__kicker {
  margin: 0 0 0.35rem;
  color: var(--tz-text-accent);
  font-size: 0.75rem;
  font-weight: 750;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.schwalbe-selector__title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: clamp(1.65rem, 3vw, 2.4rem);
  font-weight: 800;
  line-height: 1.15;
}

.schwalbe-selector__controls {
  display: grid;
  gap: 0.85rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: var(--tz-card-surface);
  padding: 1rem;
}

.schwalbe-selector__search,
.schwalbe-selector__selects label {
  display: grid;
  gap: 0.35rem;
  min-width: 0;
}

.schwalbe-selector__label {
  color: var(--tz-text-primary);
  font-size: 0.78rem;
  font-weight: 700;
}

.schwalbe-selector__search-row {
  display: flex;
  gap: 0.55rem;
}

.schwalbe-selector__search input,
.schwalbe-selector__selects select {
  min-height: 2.5rem;
  min-width: 0;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.65rem;
  background: var(--tz-surface-page);
  color: var(--tz-text-primary);
  padding: 0.55rem 0.7rem;
  font: inherit;
  font-size: 0.86rem;
}

.schwalbe-selector__search input {
  flex: 1 1 auto;
}

.schwalbe-selector__button,
.schwalbe-selector__clear {
  min-height: 2.5rem;
  border: 1px solid var(--tz-action-primary);
  border-radius: 0.65rem;
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  padding: 0.55rem 0.85rem;
  font: inherit;
  font-size: 0.82rem;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
}

.schwalbe-selector__button:hover,
.schwalbe-selector__clear:hover {
  border-color: var(--tz-action-primary-hover);
  background: var(--tz-action-primary-hover);
}

.schwalbe-selector__clear {
  align-self: end;
  border-color: var(--tz-border-strong);
  background: transparent;
  color: var(--tz-text-primary);
}

.schwalbe-selector__hint {
  color: var(--tz-text-secondary);
  font-size: 0.74rem;
  line-height: 1.45;
}

.schwalbe-selector__selects {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
  gap: 0.75rem;
  align-items: end;
}

.schwalbe-selector__summary {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem 1rem;
  justify-content: space-between;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
}

.schwalbe-selector__state {
  border: 1px dashed var(--tz-border-strong);
  border-radius: 1rem;
  background: var(--tz-surface-subtle);
  color: var(--tz-text-secondary);
  padding: 2rem 1rem;
  text-align: center;
}

.schwalbe-selector__state p {
  margin: 0 0 0.8rem;
}

.schwalbe-selector__state--error {
  border-color: color-mix(in srgb, var(--tz-status-danger-text) 38%, transparent);
  color: var(--tz-status-danger-text);
}

.schwalbe-selector__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
}

.schwalbe-selector__pagination {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  align-items: center;
  justify-content: center;
  padding-top: 0.25rem;
}

.schwalbe-selector__page-link {
  display: inline-grid;
  min-width: 2rem;
  min-height: 2rem;
  place-items: center;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.55rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-primary);
  padding: 0.25rem 0.5rem;
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1;
  text-decoration: none;
}

.schwalbe-selector__page-link:hover {
  border-color: var(--tz-action-primary);
  color: var(--tz-action-primary);
}

.schwalbe-selector__page-link--current {
  border-color: var(--tz-action-primary);
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
}

.schwalbe-selector__page-link--disabled {
  color: var(--tz-text-disabled);
  cursor: not-allowed;
}

.schwalbe-selector__page-ellipsis {
  min-width: 1.25rem;
  color: var(--tz-text-muted);
  text-align: center;
}

@media (max-width: 1024px) {
  .schwalbe-selector__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .schwalbe-selector__search-row,
  .schwalbe-selector__selects {
    display: grid;
    grid-template-columns: 1fr;
  }

  .schwalbe-selector__grid {
    grid-template-columns: 1fr;
  }

  .schwalbe-selector__button,
  .schwalbe-selector__clear {
    width: 100%;
  }

  .schwalbe-selector__pagination {
    justify-content: flex-start;
  }
}
</style>
