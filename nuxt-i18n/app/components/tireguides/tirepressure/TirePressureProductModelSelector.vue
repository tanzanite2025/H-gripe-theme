<template>
  <section
    class="tire-pressure-product-model-selector"
    aria-labelledby="tire-pressure-product-model-selector-title"
  >
    <header class="tire-pressure-product-model-selector__header">
      <h3 id="tire-pressure-product-model-selector-title">
        {{ t('guidesTirePressure.productSelector.title') }}
      </h3>
    </header>

    <div
      ref="searchShell"
      class="tire-pressure-product-model-selector__search"
      @focusin="openSuggestions"
      @focusout="handleSearchFocusOut"
    >
      <label for="tire-pressure-product-model-search">
        {{ t('guidesTirePressure.productSelector.searchLabel') }}
      </label>
      <div class="tire-pressure-product-model-selector__input-wrap">
        <input
          id="tire-pressure-product-model-search"
          v-model="searchInput"
          type="search"
          role="combobox"
          autocomplete="off"
          aria-autocomplete="list"
          aria-controls="tire-pressure-product-model-suggestions"
          :aria-expanded="isSuggestionsOpen"
          :aria-activedescendant="activeSuggestionId || undefined"
          :placeholder="t('guidesTirePressure.productSelector.searchPlaceholder')"
          @input="openSuggestions"
          @keydown="handleSearchKeydown"
        >
        <div
          v-if="isSuggestionsOpen && !pending && !error"
          id="tire-pressure-product-model-suggestions"
          class="tire-pressure-product-model-selector__popover"
          role="listbox"
          :aria-label="t('guidesTirePressure.productSelector.resultsLabel')"
        >
          <div v-if="!hasSearchInput" class="tire-pressure-product-model-selector__hint">
            {{ t('guidesTirePressure.productSelector.searchHint') }}
          </div>
          <template v-else>
            <div class="tire-pressure-product-model-selector__results-heading" aria-live="polite">
              {{ t('guidesTirePressure.productSelector.resultsSummary', { count: matchingItems.length }) }}
            </div>
            <div v-if="matchingItems.length === 0" class="tire-pressure-product-model-selector__hint">
              {{ t('guidesTirePressure.productSelector.noResults') }}
            </div>
            <div v-else class="tire-pressure-product-model-selector__results">
              <button
                v-for="(item, index) in visibleMatchingItems"
                :id="suggestionId(item.article_no)"
                :key="item.article_no"
                type="button"
                role="option"
                class="tire-pressure-product-model-selector__result"
                :class="{
                  'tire-pressure-product-model-selector__result--active': index === activeSuggestionIndex,
                  'tire-pressure-product-model-selector__result--selected': item.article_no === selectedTireModelArticleNumber,
                }"
                :aria-selected="item.article_no === selectedTireModelArticleNumber"
                @mousedown.prevent
                @click="selectTireModel(item)"
              >
                <span class="tire-pressure-product-model-selector__result-main">
                  <strong>{{ item.model_name }}</strong>
                  <small>{{ item.article_no }} · {{ item.etrto }}<template v-if="item.inch_designation"> · {{ item.inch_designation }}</template></small>
                </span>
                <span class="tire-pressure-product-model-selector__result-pressure">
                  {{ formatPressureRange(item) }}
                </span>
              </button>
            </div>
            <p v-if="matchingItems.length > visibleMatchingItems.length" class="tire-pressure-product-model-selector__result-limit">
              {{ t('guidesTirePressure.productSelector.resultLimit', { count: visibleMatchingItems.length }) }}
            </p>
          </template>
        </div>
      </div>
      <span class="tire-pressure-product-model-selector__catalog-count">
        {{ t('guidesTirePressure.productSelector.catalogCount', { count: catalogItems.length }) }}
      </span>
    </div>

    <div v-if="pending" class="tire-pressure-product-model-selector__state" role="status">
      {{ t('guidesTirePressure.productSelector.loading') }}
    </div>
    <div v-else-if="error" class="tire-pressure-product-model-selector__state tire-pressure-product-model-selector__state--error" role="alert">
      <span>{{ t('guidesTirePressure.productSelector.error') }}</span>
      <button type="button" @click="() => refresh()">
        {{ t('guidesTirePressure.productSelector.retry') }}
      </button>
    </div>
    <template v-else>
      <article v-if="selectedItem" class="tire-pressure-product-model-selector__selected">
        <div class="tire-pressure-product-model-selector__selected-heading">
          <span>{{ t('guidesTirePressure.productSelector.selectedLabel') }}</span>
          <button type="button" @click="clearSelectedTireModel">
            {{ t('guidesTirePressure.productSelector.clear') }}
          </button>
        </div>
        <strong>{{ selectedItem.model_name }}</strong>
        <span>{{ selectedItem.article_no }} · {{ selectedItem.etrto }}<template v-if="selectedItem.inch_designation"> · {{ selectedItem.inch_designation }}</template></span>
        <small>{{ formatPressureRange(selectedItem) }}</small>
      </article>

    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useAsyncData, useI18n } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import {
  fetchSchwalbeTirePressureReferenceCatalog,
  type SchwalbeTireCatalogItem,
} from '~/data/tireguides/schwalbeCatalog'

const { t } = useI18n()
const { request } = useApiRequest()
const props = defineProps<{
  selectedTireModel: SchwalbeTireCatalogItem | null
}>()
const emit = defineEmits<{
  (event: 'update:selectedTireModel', value: SchwalbeTireCatalogItem | null): void
}>()
const searchInput = ref('')
const searchShell = ref<HTMLElement | null>(null)
const isSuggestionsOpen = ref(false)
const activeSuggestionIndex = ref(-1)
const visibleResultLimit = 12
let closeSuggestionsTimer: ReturnType<typeof setTimeout> | null = null

const { data: catalogData, pending, error, refresh } = await useAsyncData<SchwalbeTireCatalogItem[]>(
  'tire-pressure-product-model-selector-catalog',
  () => fetchSchwalbeTirePressureReferenceCatalog(request),
  { default: () => [] },
)

const catalogItems = computed(() => catalogData.value || [])
const normalizedSearchInput = computed(() => searchInput.value.trim().toLowerCase())
const hasSearchInput = computed(() => normalizedSearchInput.value.length > 0)

const matchingItems = computed(() => {
  if (!hasSearchInput.value) return []
  const query = normalizedSearchInput.value
  return catalogItems.value
    .filter((item) => [
      item.model_name,
      item.article_no,
      item.etrto,
      item.inch_designation,
      item.ean,
      item.version_label,
    ].some(value => value?.toLowerCase().includes(query)))
    .sort((left, right) => (
      left.model_name.localeCompare(right.model_name)
      || left.etrto.localeCompare(right.etrto, undefined, { numeric: true })
      || left.article_no.localeCompare(right.article_no)
    ))
})

const visibleMatchingItems = computed(() => matchingItems.value.slice(0, visibleResultLimit))
const selectedItem = computed(() => props.selectedTireModel)
const selectedTireModelArticleNumber = computed(() => selectedItem.value?.article_no ?? null)

const suggestionId = (articleNumber: string) => (
  `tire-pressure-product-model-option-${articleNumber.replace(/[^a-zA-Z0-9_-]/g, '-')}`
)

const activeSuggestionId = computed(() => {
  const item = visibleMatchingItems.value[activeSuggestionIndex.value]
  return item ? suggestionId(item.article_no) : undefined
})

const selectTireModel = (item: SchwalbeTireCatalogItem) => {
  emit('update:selectedTireModel', item)
  searchInput.value = item.model_name
  isSuggestionsOpen.value = false
  activeSuggestionIndex.value = -1
}

const openSuggestions = () => {
  if (closeSuggestionsTimer) {
    clearTimeout(closeSuggestionsTimer)
    closeSuggestionsTimer = null
  }
  isSuggestionsOpen.value = true
}

const handleSearchFocusOut = () => {
  if (closeSuggestionsTimer) clearTimeout(closeSuggestionsTimer)
  closeSuggestionsTimer = setTimeout(() => {
    if (!searchShell.value?.contains(document.activeElement)) {
      isSuggestionsOpen.value = false
      activeSuggestionIndex.value = -1
    }
    closeSuggestionsTimer = null
  }, 0)
}

const handleSearchKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    isSuggestionsOpen.value = false
    activeSuggestionIndex.value = -1
    return
  }

  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    if (!visibleMatchingItems.value.length) return
    event.preventDefault()
    openSuggestions()
    const direction = event.key === 'ArrowDown' ? 1 : -1
    const nextIndex = activeSuggestionIndex.value + direction
    activeSuggestionIndex.value = nextIndex < 0
      ? visibleMatchingItems.value.length - 1
      : nextIndex >= visibleMatchingItems.value.length ? 0 : nextIndex
    return
  }

  if (event.key === 'Enter' && isSuggestionsOpen.value && activeSuggestionIndex.value >= 0) {
    const item = visibleMatchingItems.value[activeSuggestionIndex.value]
    if (!item) return
    event.preventDefault()
    selectTireModel(item)
  }
}

const clearSelectedTireModel = () => {
  emit('update:selectedTireModel', null)
  searchInput.value = ''
  activeSuggestionIndex.value = -1
}

const formatPressureRange = (item: SchwalbeTireCatalogItem) => {
  const barRange = item.min_pressure_bar !== undefined || item.max_pressure_bar !== undefined
    ? `${item.min_pressure_bar ?? '—'}–${item.max_pressure_bar ?? '—'} bar`
    : ''
  const psiRange = item.min_pressure_psi !== undefined || item.max_pressure_psi !== undefined
    ? `${item.min_pressure_psi ?? '—'}–${item.max_pressure_psi ?? '—'} PSI`
    : ''
  return [barRange, psiRange].filter(Boolean).join(' · ') || t('guidesTirePressure.productSelector.pressureUnavailable')
}

onBeforeUnmount(() => {
  if (closeSuggestionsTimer) clearTimeout(closeSuggestionsTimer)
})
</script>

<style scoped>
.tire-pressure-product-model-selector {
  display: grid;
  gap: 0.6rem;
  min-width: 0;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.85rem;
  background: var(--tz-card-surface);
  padding: 0.65rem;
  color: var(--tz-text-primary);
  text-align: left;
}

.tire-pressure-product-model-selector__header {
  display: grid;
  gap: 0.1rem;
}

.tire-pressure-product-model-selector__header h3 {
  margin: 0;
  font-size: 1rem;
  line-height: 1.3;
}

.tire-pressure-product-model-selector__header p {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.76rem;
  line-height: 1.45;
}

.tire-pressure-product-model-selector__search {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.35rem 0.65rem;
  align-items: center;
}

.tire-pressure-product-model-selector__input-wrap {
  position: relative;
  min-width: 0;
}

.tire-pressure-product-model-selector__search label {
  grid-column: 1 / -1;
  color: var(--tz-text-primary);
  font-size: 0.75rem;
  font-weight: 700;
}

.tire-pressure-product-model-selector__search input {
  box-sizing: border-box;
  width: 100%;
  min-height: 2.45rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.6rem;
  background: var(--tz-surface-page);
  color: var(--tz-text-primary);
  padding: 0.5rem 0.65rem;
  font: inherit;
  font-size: 0.82rem;
}

.tire-pressure-product-model-selector__popover {
  position: absolute;
  z-index: 20;
  top: calc(100% + 0.35rem);
  left: 0;
  right: 0;
  display: grid;
  gap: 0.45rem;
  max-height: 20rem;
  overflow-y: auto;
  padding: 0.45rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.7rem;
  background: var(--tz-card-surface);
  box-shadow: 0 0.7rem 1.5rem rgba(15, 23, 42, 0.18);
}

.tire-pressure-product-model-selector__catalog-count {
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  white-space: nowrap;
}

.tire-pressure-product-model-selector__state,
.tire-pressure-product-model-selector__hint {
  border-radius: 0.6rem;
  background: var(--tz-surface-muted);
  color: var(--tz-text-secondary);
  padding: 0.7rem;
  font-size: 0.75rem;
  line-height: 1.4;
}

.tire-pressure-product-model-selector__state--error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.7rem;
  color: var(--tz-status-error-text);
}

.tire-pressure-product-model-selector__state button,
.tire-pressure-product-model-selector__selected-heading button {
  border: 0;
  background: transparent;
  color: var(--tz-action-primary);
  padding: 0;
  font: inherit;
  font-size: 0.72rem;
  font-weight: 700;
  cursor: pointer;
}

.tire-pressure-product-model-selector__selected {
  display: grid;
  gap: 0.18rem;
  border: 1px solid var(--tz-action-primary);
  border-radius: 0.65rem;
  background: color-mix(in srgb, var(--tz-action-primary) 7%, var(--tz-card-surface));
  padding: 0.65rem 0.7rem;
}

.tire-pressure-product-model-selector__selected-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
}

.tire-pressure-product-model-selector__selected strong {
  font-size: 0.86rem;
}

.tire-pressure-product-model-selector__selected > span,
.tire-pressure-product-model-selector__selected small {
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  line-height: 1.35;
}

.tire-pressure-product-model-selector__results-heading {
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
}

.tire-pressure-product-model-selector__results {
  display: grid;
  gap: 0.35rem;
}

.tire-pressure-product-model-selector__result {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.7rem;
  align-items: center;
  min-width: 0;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.6rem;
  background: var(--tz-surface-page);
  color: var(--tz-text-primary);
  padding: 0.55rem 0.65rem;
  font: inherit;
  text-align: left;
  cursor: pointer;
}

.tire-pressure-product-model-selector__result:hover,
.tire-pressure-product-model-selector__result:focus-visible,
.tire-pressure-product-model-selector__result--active,
.tire-pressure-product-model-selector__result--selected {
  border-color: var(--tz-action-primary);
  background: color-mix(in srgb, var(--tz-action-primary) 7%, var(--tz-surface-page));
  outline: none;
}

.tire-pressure-product-model-selector__result-main {
  display: grid;
  min-width: 0;
  gap: 0.15rem;
}

.tire-pressure-product-model-selector__result-main strong,
.tire-pressure-product-model-selector__result-main small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tire-pressure-product-model-selector__result-main strong {
  font-size: 0.78rem;
}

.tire-pressure-product-model-selector__result-main small,
.tire-pressure-product-model-selector__result-pressure,
.tire-pressure-product-model-selector__result-limit {
  color: var(--tz-text-secondary);
  font-size: 0.66rem;
  line-height: 1.3;
}

.tire-pressure-product-model-selector__result-pressure {
  white-space: nowrap;
}

.tire-pressure-product-model-selector__result-limit {
  margin: 0;
}

@media (max-width: 620px) {
  .tire-pressure-product-model-selector__search {
    grid-template-columns: 1fr;
  }

  .tire-pressure-product-model-selector__catalog-count {
    white-space: normal;
  }

  .tire-pressure-product-model-selector__result {
    grid-template-columns: 1fr;
    gap: 0.25rem;
  }
}
</style>
