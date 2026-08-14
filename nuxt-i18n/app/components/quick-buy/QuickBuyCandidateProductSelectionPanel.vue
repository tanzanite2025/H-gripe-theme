<script setup lang="ts">
import QuickBuyLocalizedHelpQuestionMarkDialog from '~/components/quick-buy/QuickBuyLocalizedHelpQuestionMarkDialog.vue'
import ShopProductDisplayCard from '~/components/shop/ShopProductDisplayCard.vue'
import type { ShopProduct } from '~/composables/useShopProducts'

const props = withDefaults(defineProps<{
  title: string
  fallbackTitle: string
  helpTitle: string
  helpContent: string
  searchPlaceholder: string
  products: ShopProduct[]
  query: string
  errorMessage?: string
  loading?: boolean
  canGoToPreviousProductPage?: boolean
  canGoToNextProductPage?: boolean
  showProductPagination?: boolean
  productPage?: number
  currentStepIndex: number
  totalSteps: number
  previousLabel: string
  nextLabel: string
  productRailLabel: string
  emptyLabel: string
  loadingLabel: string
  currentPageLabel: string
  helpTriggerAriaLabel: string
  closeLabel: string
  isProductSelected?: (product: ShopProduct) => boolean
}>(), {
  errorMessage: '',
  loading: false,
  canGoToPreviousProductPage: false,
  canGoToNextProductPage: false,
  showProductPagination: false,
  productPage: 1,
})

const emit = defineEmits<{
  'update:query': [value: string]
  queryInput: []
  submitSearch: []
  previousProductPage: []
  nextProductPage: []
  selectProduct: [product: ShopProduct]
  openProductDetails: [product: ShopProduct]
  previousStep: []
  nextStep: []
}>()

const handleSearchInput = (event: Event) => {
  emit('update:query', (event.target as HTMLInputElement | null)?.value || '')
  emit('queryInput')
}
</script>

<template>
  <div class="quickbuy-selection-panel">
    <div class="quickbuy-selection-panel-header">
      <div class="quickbuy-panel-heading-row">
        <h2 class="my-0 min-w-0 text-lg font-semibold text-white">
          {{ title || fallbackTitle }}
        </h2>
        <QuickBuyLocalizedHelpQuestionMarkDialog
          :title="helpTitle"
          :content="helpContent"
          :trigger-aria-label="helpTriggerAriaLabel"
          :close-label="closeLabel"
        />
      </div>
      <input
        :value="query"
        type="text"
        :placeholder="searchPlaceholder"
        class="quickbuy-search-input mt-3 w-full px-3 py-2.5 rounded-lg text-white box-border max-w-full focus:outline-none transition-colors"
        @keydown.enter.prevent="emit('submitSearch')"
        @input="handleSearchInput"
      />
    </div>

    <div class="quickbuy-candidate-area">
      <div class="quickbuy-product-grid-shell">
        <button
          class="quickbuy-product-grid-arrow quickbuy-product-grid-arrow--previous"
          type="button"
          :disabled="!canGoToPreviousProductPage || loading"
          :aria-label="previousLabel"
          :title="previousLabel"
          @click="emit('previousProductPage')"
        >
          <Icon name="lucide:chevron-left" class="h-4 w-4" aria-hidden="true" />
        </button>

        <div class="quickbuy-product-grid-stage">
          <div
            class="quickbuy-product-grid"
            :aria-label="productRailLabel"
          >
            <ShopProductDisplayCard
              v-for="product in products"
              :key="product.id"
              :product="product"
              density="quick-buy"
              selectable
              :selected="props.isProductSelected?.(product) || false"
              show-details-action
              :show-view-action="false"
              @select="emit('selectProduct', product)"
              @details="emit('openProductDetails', product)"
            />
          </div>
          <div
            v-if="errorMessage && !products.length"
            class="quickbuy-product-grid-empty"
          >
            {{ errorMessage }}
          </div>
          <div
            v-else-if="loading && !products.length"
            class="quickbuy-product-grid-empty"
          >
            {{ loadingLabel }}
          </div>
          <div
            v-else-if="!products.length"
            class="quickbuy-product-grid-empty"
          >
            {{ emptyLabel }}
          </div>
        </div>

        <button
          class="quickbuy-product-grid-arrow quickbuy-product-grid-arrow--next"
          type="button"
          :disabled="!canGoToNextProductPage || loading"
          :aria-label="nextLabel"
          :title="nextLabel"
          @click="emit('nextProductPage')"
        >
          <Icon name="lucide:chevron-right" class="h-4 w-4" aria-hidden="true" />
        </button>
      </div>
      <div v-if="showProductPagination" class="quickbuy-product-pagination">
        <span class="quickbuy-product-pagination__page" :aria-label="currentPageLabel">
          {{ productPage }}
        </span>
      </div>
    </div>

    <div class="quickbuy-step-actions">
      <button
        class="quickbuy-step-action quickbuy-step-action--secondary"
        type="button"
        :disabled="currentStepIndex <= 1"
        :title="previousLabel"
        :aria-label="previousLabel"
        @click="emit('previousStep')"
      >
        <Icon name="lucide:arrow-left" class="h-5 w-5" aria-hidden="true" />
      </button>
      <button
        v-if="currentStepIndex < totalSteps"
        class="quickbuy-step-action quickbuy-step-action--primary"
        type="button"
        :title="nextLabel"
        :aria-label="nextLabel"
        @click="emit('nextStep')"
      >
        <Icon name="lucide:arrow-right" class="h-5 w-5" aria-hidden="true" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.quickbuy-selection-panel {
  display: flex;
  min-width: 0;
  min-height: 0;
  box-sizing: border-box;
  flex-direction: column;
  padding: 0.75rem;
  border: 0;
  border-radius: 0.75rem;
  background:
    linear-gradient(180deg, var(--quickbuy-panel-surface, #111116), var(--quickbuy-panel-surface-soft, #0c0c0e));
  box-shadow:
    0 16px 42px rgba(0, 0, 0, 0.42),
    inset 0 1px 0 rgba(255, 255, 255, 0.026),
    inset 0 0 0 1px var(--quickbuy-dark-edge, rgba(0, 0, 0, 0.68));
}

.quickbuy-selection-panel-header {
  flex: 0 0 auto;
}

.quickbuy-search-input {
  border: 0 !important;
  background-color: #080809 !important;
  background-image:
    linear-gradient(180deg, var(--quickbuy-control-surface, #0a0a0c), #080809) !important;
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.82);
}

.quickbuy-search-input:focus {
  border: 0 !important;
  background-color: #09090a !important;
  background-image: none !important;
  box-shadow:
    inset 0 1px 2px rgba(0, 0, 0, 0.74),
    0 0 0 3px var(--quickbuy-focus-ring, rgba(181, 255, 109, 0.12));
}

.quickbuy-panel-heading-row {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.quickbuy-candidate-area {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  margin-top: 0.75rem;
}

.quickbuy-step-actions {
  display: flex;
  flex: 0 0 auto;
  justify-content: center;
  gap: 0.5rem;
  padding-top: 0.75rem;
  margin-top: 0.75rem;
  box-shadow: inset 0 1px 0 var(--quickbuy-divider, rgba(255, 255, 255, 0.045));
}

.quickbuy-step-action {
  display: inline-grid;
  width: 2.75rem;
  height: 2.75rem;
  place-items: center;
  border: 0;
  border-radius: 999px;
  color: white;
  background:
    linear-gradient(180deg, var(--quickbuy-control-surface-raised, #151519), #0f0f12);
  box-shadow:
    0 6px 18px rgba(0, 0, 0, 0.24),
    inset 0 0 0 1px rgba(0, 0, 0, 0.7);
  transition: background-color 160ms ease, opacity 160ms ease, transform 160ms ease;
}

.quickbuy-step-action:hover:not(:disabled) {
  background:
    linear-gradient(180deg, #1b1b20, #121216);
  transform: translateY(-1px);
}

.quickbuy-step-action:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

.quickbuy-step-action--primary {
  color: black;
  background: white;
  box-shadow:
    0 8px 20px rgba(0, 0, 0, 0.32),
    inset 0 0 0 1px rgba(0, 0, 0, 0.08);
}

.quickbuy-step-action--primary:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.88);
}

.quickbuy-product-grid-shell {
  display: grid;
  grid-template-columns: 2rem minmax(0, 1fr) 2rem;
  flex: 1;
  align-items: center;
  gap: 0.5rem;
  box-sizing: border-box;
  min-height: 22rem;
  padding: 0.5rem;
  border: 0;
  border-radius: 0.75rem;
  background:
    linear-gradient(180deg, #080809, var(--quickbuy-control-surface, #0a0a0c));
  box-shadow:
    inset 0 0 0 1px rgba(0, 0, 0, 0.72),
    inset 0 -18px 36px rgba(0, 0, 0, 0.26);
}

.quickbuy-product-grid-stage {
  position: relative;
  min-width: 0;
  min-height: 0;
  height: 100%;
}

.quickbuy-product-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  grid-template-rows: repeat(2, minmax(0, 1fr));
  gap: 0.625rem;
  min-width: 0;
  min-height: 0;
  height: 100%;
}

.quickbuy-product-grid-stage > .quickbuy-product-grid:empty {
  display: block;
}

.quickbuy-product-grid-arrow {
  width: 2rem;
  height: 2rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 999px;
  color: white;
  background:
    linear-gradient(180deg, var(--quickbuy-control-surface-raised, #151519), #101013);
  box-shadow:
    0 6px 16px rgba(0, 0, 0, 0.25),
    inset 0 0 0 1px rgba(0, 0, 0, 0.7);
  transition: background-color 160ms ease, opacity 160ms ease, transform 160ms ease;
}

.quickbuy-product-grid-arrow:hover:not(:disabled) {
  background:
    linear-gradient(180deg, #1e1e24, #131318);
  transform: translateY(-1px);
}

.quickbuy-product-grid-arrow:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.quickbuy-product-grid-empty {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 1rem;
  color: rgba(255, 255, 255, 0.62);
  font-size: 0.8125rem;
  text-align: center;
}

.quickbuy-product-pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 1.75rem;
  padding-top: 0.35rem;
}

.quickbuy-product-pagination__page {
  display: inline-grid;
  min-width: 1.75rem;
  height: 1.75rem;
  place-items: center;
  border: 0;
  border-radius: 999px;
  color: rgba(255, 255, 255, 0.78);
  background: var(--quickbuy-control-surface-raised, #151519);
  box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.68);
  font-size: 0.6875rem;
  font-weight: 700;
}

@media (max-width: 767px) {
  .quickbuy-selection-panel {
    min-height: 15rem;
  }

  .quickbuy-product-grid-shell {
    grid-template-columns: 1.75rem minmax(0, 1fr) 1.75rem;
    gap: 0.35rem;
    min-height: 20rem;
    padding: 0.375rem;
  }

  .quickbuy-product-grid-arrow {
    width: 1.75rem;
    height: 1.75rem;
  }

  .quickbuy-product-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    grid-template-rows: repeat(3, minmax(0, 1fr));
    gap: 0.5rem;
  }

  .quickbuy-product-grid-stage {
    min-height: 0;
  }
}
</style>
