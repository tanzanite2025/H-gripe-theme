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
      <div class="schwalbe-selector__search">
        <label class="schwalbe-selector__label" :for="searchInputId">{{ tx('search.label') }}</label>
        <span class="schwalbe-selector__search-row">
          <input
            :id="searchInputId"
            v-model="searchInput"
            type="search"
            :placeholder="tx('search.placeholder')"
          >
          <button type="submit" class="schwalbe-selector__button">
            {{ tx('search.submit') }}
          </button>
          <button
            type="button"
            class="schwalbe-selector__sort-toggle"
            :aria-label="tx(sortBy === 'weight_desc' ? 'filters.sortWeightDescendingAria' : 'filters.sortWeightAscendingAria')"
            :title="tx(sortBy === 'weight_desc' ? 'filters.sortWeightDescendingAria' : 'filters.sortWeightAscendingAria')"
            @click="toggleWeightSort"
          >
            <Icon name="lucide:scale" class="schwalbe-selector__sort-icon" aria-hidden="true" />
            <span class="schwalbe-selector__sort-direction" aria-hidden="true">
              {{ sortBy === 'weight_desc' ? '↓' : '↑' }}
            </span>
          </button>
          <button
            type="button"
            class="schwalbe-selector__filter-button"
            :aria-label="tx('filters.openFilters')"
            aria-haspopup="dialog"
            :aria-expanded="filterDialogOpen"
            :aria-controls="filterDialogOpen ? filterDialogId : undefined"
            @click="openFilterDialog"
          >
            <Icon name="lucide:sliders-horizontal" class="schwalbe-selector__filter-icon" aria-hidden="true" />
            <span
              v-if="activeFilterCount > 0"
              class="schwalbe-selector__filter-count"
              :aria-label="tx('filters.activeCount', { count: activeFilterCount })"
            >
              {{ activeFilterCount }}
            </span>
          </button>
        </span>
      </div>

      <div class="schwalbe-selector__selects">
        <label class="schwalbe-selector__model-filter">
          <span class="schwalbe-selector__label">{{ tx('filters.model') }}</span>
          <select v-model="selectedModel">
            <option v-for="model in modelOptions" :key="model" :value="model">
              {{ model === 'ALL' ? tx('filters.allModels') : model }}
            </option>
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

    <SchwalbeTireWheelSizeTabs
      :label="tx('wheelSizeTabsLabel')"
      :all-label="tx('allWheelSizes')"
      :multiple-label="tx('multipleWheelSizes')"
      :dialog-title="tx('wheelSizeDialogTitle')"
      :close-label="tx('wheelSizeDialogClose')"
      :option-template="tx('wheelSizeTabOption', { diameter: '{diameter}', bsd: '{bsd}' })"
      :options="wheelSizeNavigationOptions"
      :selected-wheel-size-keys="selectedWheelSizeKeys"
      @select="selectWheelSize"
    />

    <SchwalbeTireCatalogFilterDrawer
      :id="filterDialogId"
      v-model:open="filterDialogOpen"
      :title="tx('filters.dialogTitle')"
      :close-label="tx('filters.closeFilters')"
      :show-results-label="tx('filters.showResults')"
      @apply="applyFilterDraft"
      @cancel="discardFilterDraft"
    >
      <SchwalbeTireCatalogFilterPanel
        v-model:selected-radial-only="draftFacetFilters.radialOnly"
        v-model:selected-beads="draftFacetFilters.beads"
        v-model:selected-seals="draftFacetFilters.seals"
        v-model:selected-e-bike-ratings="draftFacetFilters.eBikeRatings"
        @reset="clearDraftFacetFilters"
        :label="tx('filters.catalogFilters')"
        :radial-group-label="tx('filters.radialGroup')"
        :radial-all-label="tx('filters.radialAll')"
        :radial-label="tx('filters.radial')"
        :bead-label="tx('filters.bead')"
        :seal-label="tx('filters.seal')"
        :e-bike-rating-label="tx('filters.eBikeRating')"
        :e-bike-unrated-label="tx('filters.eBikeUnrated')"
        :reset-label="tx('filters.clearFilters')"
        :bead-options="beadOptions"
        :seal-options="sealOptions"
        :e-bike-rating-options="eBikeRatingOptions"
      />
    </SchwalbeTireCatalogFilterDrawer>

    <div class="schwalbe-selector__summary" aria-live="polite">
      <span>{{ tx('summary', { count: totalItems, page: currentPage, totalPages }) }}</span>
      <span v-if="submittedSearch">{{ tx('search.active', { term: submittedSearch }) }}</span>
      <span v-if="hasRimWidthGuidance">{{ tx('rimWidth.catalogHint') }}</span>
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
import { computed, reactive, ref, useId } from 'vue'
import { useI18n } from '#imports'
import SchwalbeTireCard from '~/components/tireguides/schwalbe/SchwalbeTireCard.vue'
import SchwalbeTireCatalogFilterPanel from '~/components/tireguides/schwalbe/SchwalbeTireCatalogFilterPanel.vue'
import SchwalbeTireCatalogFilterDrawer from '~/components/tireguides/schwalbe/SchwalbeTireCatalogFilterDrawer.vue'
import SchwalbeTelemetryGuide from '~/components/tireguides/schwalbe/SchwalbeTelemetryGuide.vue'
import SchwalbeTireWheelSizeTabs from '~/components/tireguides/schwalbe/SchwalbeTireWheelSizeTabs.vue'
import {
  useSchwalbeTireSelector,
  type SchwalbeTireCatalogFacetFilterState,
} from '~/composables/useSchwalbeTireSelector'

const { t: translate } = useI18n()
const tx = (key: string, params?: Record<string, unknown>) => translate(`guidesSchwalbeTireSelector.${key}`, params || {})
const filterDialogOpen = ref(false)
const searchInputId = `schwalbe-tire-search-${useId()}`
const filterDialogId = `schwalbe-tire-catalog-filter-${useId()}`
const {
  searchInput,
  submittedSearch,
  totalItems,
  selectedModel,
  selectedWheelSizeKeys,
  selectWheelSize,
  selectedRadialOnly,
  selectedBeads,
  selectedSeals,
  selectedEBikeRatings,
  getFacetFilterState,
  applyFacetFilterState,
  sortBy,
  visibleItems,
  modelOptions,
  wheelSizeNavigationOptions,
  beadOptions,
  sealOptions,
  eBikeRatingOptions,
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

const toggleWeightSort = () => {
  sortBy.value = sortBy.value === 'weight_desc' ? 'weight_asc' : 'weight_desc'
}

const draftFacetFilters = reactive<SchwalbeTireCatalogFacetFilterState>({
  radialOnly: false,
  beads: [],
  seals: [],
  eBikeRatings: [],
})

const syncDraftFacetFilters = () => {
  const committed = getFacetFilterState()
  draftFacetFilters.radialOnly = committed.radialOnly
  draftFacetFilters.beads = [...committed.beads]
  draftFacetFilters.seals = [...committed.seals]
  draftFacetFilters.eBikeRatings = [...committed.eBikeRatings]
}

const openFilterDialog = () => {
  syncDraftFacetFilters()
  filterDialogOpen.value = true
}

const discardFilterDraft = () => {
  // Closing, Escape, and clicking the backdrop intentionally leave the route
  // and catalog results untouched. The next open starts from committed state.
  filterDialogOpen.value = false
}

const applyFilterDraft = () => {
  applyFacetFilterState({
    radialOnly: draftFacetFilters.radialOnly,
    beads: [...draftFacetFilters.beads],
    seals: [...draftFacetFilters.seals],
    eBikeRatings: [...draftFacetFilters.eBikeRatings],
  })
  filterDialogOpen.value = false
}

const clearDraftFacetFilters = () => {
  draftFacetFilters.radialOnly = false
  draftFacetFilters.beads = []
  draftFacetFilters.seals = []
  draftFacetFilters.eBikeRatings = []
}

const activeFilterCount = computed(() => [
  selectedModel.value !== 'ALL',
  selectedRadialOnly.value,
  selectedBeads.value.length > 0,
  selectedSeals.value.length > 0,
  selectedEBikeRatings.value.length > 0,
].filter(Boolean).length)

const hasRimWidthGuidance = computed(() => visibleItems.value.some(item => (
  Boolean(item.rim_width_guidance?.length)
)))
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

.schwalbe-selector__sort-toggle {
  display: inline-flex;
  width: 2.7rem;
  min-width: 2.7rem;
  min-height: 2.5rem;
  height: 2.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  justify-self: start;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.65rem;
  background: var(--tz-surface-page);
  color: var(--tz-text-primary);
  padding: 0;
  font: inherit;
  font-size: 0.82rem;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
}

.schwalbe-selector__sort-toggle:hover {
  border-color: var(--tz-action-primary);
  color: var(--tz-action-primary);
}

.schwalbe-selector__sort-toggle:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.schwalbe-selector__sort-toggle :deep(svg) {
  width: 1.1rem;
  height: 1.1rem;
  stroke-width: 2;
}

.schwalbe-selector__sort-direction {
  font-size: 0.78rem;
  font-weight: 800;
  line-height: 1;
}

.schwalbe-selector__filter-button {
  position: relative;
  display: inline-grid;
  width: 2.7rem;
  min-width: 2.7rem;
  min-height: 2.5rem;
  height: 2.7rem;
  align-items: center;
  justify-content: center;
  place-items: center;
  justify-self: end;
  border: 1px solid #0b0b0b;
  border-radius: 0.65rem;
  background: #0b0b0b;
  color: #fff;
  padding: 0;
  font: inherit;
  font-weight: 750;
  cursor: pointer;
  box-shadow: 0 0.35rem 0.8rem rgb(0 0 0 / 0.16);
}

.schwalbe-selector__filter-button:hover {
  border-color: #000;
  background: #000;
  color: #fff;
  box-shadow: 0 0.45rem 1rem rgb(0 0 0 / 0.24);
}

.schwalbe-selector__filter-button :deep(svg) {
  width: 1.15rem;
  height: 1.15rem;
  stroke-width: 2.2;
}

.schwalbe-selector__filter-button:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.schwalbe-selector__search-row > .schwalbe-selector__button,
.schwalbe-selector__search-row > .schwalbe-selector__sort-toggle,
.schwalbe-selector__search-row > .schwalbe-selector__filter-button {
  width: 4.5rem;
  min-width: 4.5rem;
  height: 2.5rem;
  min-height: 2.5rem;
  border-radius: 999px;
}

.schwalbe-selector__filter-count {
  position: absolute;
  top: -0.45rem;
  right: -0.45rem;
  display: inline-grid;
  min-width: 1.15rem;
  height: 1.15rem;
  place-items: center;
  border-radius: 999px;
  border: 2px solid var(--tz-card-surface);
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  padding: 0 0.2rem;
  font-size: 0.68rem;
  font-weight: 800;
  line-height: 1;
}

.schwalbe-selector__selects {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
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

.schwalbe-selector__state--conflict {
  border-color: color-mix(in srgb, var(--tz-status-warning-text, #a16207) 38%, transparent);
  background: color-mix(in srgb, var(--tz-status-warning-text, #a16207) 7%, var(--tz-card-surface));
}

.schwalbe-selector__state--conflict .schwalbe-selector__button {
  display: inline-block;
  width: auto;
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

@media (max-width: 760.5px) {
  .schwalbe-selector__search-row {
    display: flex;
    flex-wrap: wrap;
  }

  .schwalbe-selector__search-row > input {
    width: 100%;
    flex: 1 0 100%;
  }

  .schwalbe-selector__selects {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.65rem;
  }

  .schwalbe-selector__selects .schwalbe-selector__model-filter {
    display: none;
  }

  .schwalbe-selector__grid {
    grid-template-columns: 1fr;
  }

  .schwalbe-selector__button,
  .schwalbe-selector__clear {
    width: 100%;
  }

  .schwalbe-selector__clear {
    flex: 1 1 100%;
    width: auto;
  }

  .schwalbe-selector__filter-button {
    justify-self: start;
  }

  .schwalbe-selector__pagination {
    justify-content: flex-start;
  }
}
</style>
