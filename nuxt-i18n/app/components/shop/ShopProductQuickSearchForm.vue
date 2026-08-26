<template>
  <form
    class="shop-product-quick-search-form"
    :class="[
      `shop-product-quick-search-form--${density}`,
      { 'shop-product-quick-search-form--has-filter': showFilterButton },
    ]"
    @submit.prevent="submitSearch"
  >
    <label class="shop-search-input-shell">
      <span class="sr-only">{{ $t('sidebar.searchProductPlaceholder', 'Enter product name...') }}</span>
      <input
        v-model="freeTextQuery"
        type="text"
        :placeholder="$t('sidebar.searchProductPlaceholder', 'Enter product name...')"
        class="shop-search-input-inner"
      />
    </label>

    <div
      class="shop-price-range"
      :aria-label="`${$t('filter.priceRange', 'Price Range')} (${baseCurrency})`"
    >
      <span class="shop-price-range__heading">
        <span class="shop-price-range__label">{{ $t('filter.price', 'Price') }}</span>
        <span class="shop-price-range__currency" :aria-label="`Currency: ${baseCurrency}`">
          {{ baseCurrency }}
        </span>
      </span>
      <label class="shop-price-field">
        <span>{{ $t('filter.from', 'From') }}</span>
        <input
          v-model.number="quickPriceMin"
          type="number"
          min="0"
          inputmode="numeric"
          aria-label="Minimum price"
        />
      </label>
      <label class="shop-price-field">
        <span>{{ $t('filter.to', 'To') }}</span>
        <input
          v-model.number="quickPriceMax"
          type="number"
          min="0"
          inputmode="numeric"
          aria-label="Maximum price"
        />
      </label>
    </div>

    <div class="shop-search-actions">
      <button
        type="submit"
        class="shop-search-submit"
      >
        {{ $t('sidebar.search', 'Search') }}
      </button>

      <button
        v-if="showFilterButton"
        type="button"
        class="shop-filter-button"
        :aria-label="$t('filter.filters', 'Filters')"
        @click="openFilters"
      >
        <Icon name="lucide:sliders-horizontal" />
        <span>{{ $t('filter.filters', 'Filters') }}</span>
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useStorefrontContext } from '~/composables/useStorefrontContext'

type SearchDensity = 'page' | 'drawer'

interface QuickSearchFilters {
  priceRange: [number, number]
  attributes: Record<string, string[]>
}

interface QuickSearchPayload {
  query: string
  filters: QuickSearchFilters
}

const props = withDefaults(defineProps<{
  initialQuery?: string
  initialPriceRange?: [number, number]
  showFilterButton?: boolean
  density?: SearchDensity
}>(), {
  initialQuery: '',
  initialPriceRange: () => [0, 5000] as [number, number],
  showFilterButton: false,
  density: 'page',
})

const emit = defineEmits<{
  (e: 'submit', payload: QuickSearchPayload): void
  (e: 'filter-click'): void
}>()

const { baseCurrency } = useStorefrontContext()
const freeTextQuery = ref('')
const quickPriceMin = ref<number | null>(0)
const quickPriceMax = ref<number | null>(5000)

const normalizePriceRange = (range?: [number, number]): [number, number] => {
  const source = Array.isArray(range) && range.length === 2 ? range : props.initialPriceRange
  const fallback: [number, number] = [0, 5000]
  const rawMin = Number(source?.[0])
  const rawMax = Number(source?.[1])
  const min = Number.isFinite(rawMin) ? Math.max(0, rawMin) : fallback[0]
  const max = Number.isFinite(rawMax) ? Math.max(0, rawMax) : fallback[1]

  return min <= max ? [min, max] : [max, min]
}

const buildFilters = (): QuickSearchFilters => {
  return {
    priceRange: normalizePriceRange([Number(quickPriceMin.value), Number(quickPriceMax.value)]),
    attributes: {},
  }
}

const submitSearch = () => {
  emit('submit', {
    query: freeTextQuery.value.trim(),
    filters: buildFilters(),
  })
}

const openFilters = () => {
  emit('filter-click')
}

watch(() => props.initialQuery, (query) => {
  freeTextQuery.value = String(query || '')
}, { immediate: true })

watch(() => props.initialPriceRange, (range) => {
  const [min, max] = normalizePriceRange(range)
  quickPriceMin.value = min
  quickPriceMax.value = max
}, { immediate: true, deep: true })
</script>

<style scoped>
.shop-product-quick-search-form {
  display: grid;
  grid-template-columns: minmax(14rem, 1fr) minmax(18rem, 24rem) auto;
  gap: 8px;
  align-items: stretch;
}

.shop-product-quick-search-form--drawer {
  grid-template-columns: minmax(16rem, 1fr) minmax(20rem, 28rem) auto;
  gap: 10px;
}

.shop-search-input-shell {
  min-width: 0;
  display: flex;
  align-items: center;
  padding: 0 14px;
  background: var(--tz-card-surface);
  border-radius: 10px;
  box-shadow:
    0 4px 12px -6px rgba(20, 32, 43, 0.24),
    0 0 0 1px rgba(20, 32, 43, 0.08);
}

.shop-search-input-inner {
  width: 100%;
  min-width: 0;
  height: 38px;
  border: none;
  background: transparent;
  color: var(--tz-text-primary);
  font-size: 13px;
  outline: none;
}

.shop-search-input-inner::placeholder {
  color: var(--tz-text-muted);
}

.shop-price-range {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  min-width: 0;
  min-height: 38px;
  border-radius: 10px;
  border: 1px solid rgba(20, 32, 43, 0.1);
  background: var(--tz-card-surface);
  padding: 4px 8px;
}

.shop-price-range__heading {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 6px;
}

.shop-price-range__label {
  color: var(--tz-text-secondary);
  font-size: var(--tz-type-micro-label);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  white-space: nowrap;
}

.shop-price-range__currency {
  flex: 0 0 auto;
  border: 1px solid rgba(5, 150, 105, 0.28);
  border-radius: 999px;
  padding: 2px 5px;
  color: var(--tz-text-primary);
  background: rgba(5, 150, 105, 0.08);
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.04em;
  line-height: 1.1;
  white-space: nowrap;
}

.shop-price-field {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 5px;
  border-radius: 8px;
  background: var(--tz-input-surface);
  padding: 4px 7px;
}

.shop-price-field span {
  color: var(--tz-text-secondary);
  font-size: var(--tz-type-micro-label);
  font-weight: 750;
  white-space: nowrap;
}

.shop-price-field input {
  width: 100%;
  min-width: 0;
  border: none;
  background: transparent;
  color: var(--tz-text-primary);
  font-size: 13px;
  outline: none;
}

.shop-price-field input::-webkit-outer-spin-button,
.shop-price-field input::-webkit-inner-spin-button {
  margin: 0;
  appearance: none;
}

.shop-search-actions {
  display: flex;
  gap: 8px;
}

.shop-search-submit,
.shop-filter-button {
  min-height: 38px;
  border-radius: 10px;
  padding: 0 16px;
  font-weight: 800;
  transition:
    background-color 0.18s ease,
    border-color 0.18s ease,
    color 0.18s ease,
    box-shadow 0.18s ease;
}

.shop-search-submit {
  border: 1px solid rgba(20, 32, 43, 0.14);
  background: var(--tz-card-surface);
  color: var(--tz-text-primary);
  box-shadow: 0 8px 18px rgba(20, 32, 43, 0.14);
}

.shop-search-submit:hover {
  background: var(--tz-form-panel-surface);
  box-shadow: 0 10px 22px rgba(20, 32, 43, 0.18);
}

.shop-filter-button {
  display: none;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: 1px solid rgba(20, 32, 43, 0.16);
  background: var(--tz-form-panel-surface);
  color: var(--tz-text-primary);
}

.shop-filter-button:hover {
  border-color: rgba(20, 32, 43, 0.24);
  background: var(--tz-card-surface);
}

.shop-filter-button :deep(svg) {
  width: 16px;
  height: 16px;
}

@media (max-width: 980px) {
  .shop-product-quick-search-form--drawer {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .shop-product-quick-search-form {
    grid-template-columns: 1fr;
  }

  .shop-price-range {
    grid-template-columns: auto minmax(0, 1fr) minmax(0, 1fr);
  }

  .shop-search-actions {
    display: grid;
    grid-template-columns: 1fr;
  }

  .shop-product-quick-search-form--has-filter .shop-search-actions {
    grid-template-columns: 1fr 1fr;
  }

  .shop-filter-button {
    display: inline-flex;
  }
}

@media (max-width: 430px) {
  .shop-price-range {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  }

  .shop-price-range__label {
    grid-column: 1 / -1;
  }

  .shop-price-range__heading {
    grid-column: 1 / -1;
  }
}
</style>
