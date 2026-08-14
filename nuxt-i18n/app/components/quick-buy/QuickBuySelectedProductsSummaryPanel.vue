<script setup lang="ts">
import type {
  QuickBuySelectedProduct,
  QuickBuySelectedProductStepSlot,
} from '~/utils/quickBuy/selection'

defineProps<{
  title: string
  slots: QuickBuySelectedProductStepSlot[]
  totalQty: number
  totalWeightG: number
  formattedTotalPrice: string
  hasSelectedProducts: boolean
  itemsLabel: string
  weightLabel: string
  priceLabel: string
  addToCartLabel: string
  directPaymentLabel: string
  decreaseLabel: string
  increaseLabel: string
  removeLabel: string
}>()

const emit = defineEmits<{
  removeProduct: [product: QuickBuySelectedProduct]
  changeQuantity: [product: QuickBuySelectedProduct, delta: number]
  updateQuantity: [product: QuickBuySelectedProduct, value: string]
  addToCart: []
  directPayment: []
}>()

const selectedProductSpecificationSummary = (item: QuickBuySelectedProduct) => {
  const summaryParts = [
    item.sku,
    `${item.weightGrams}g`,
    `$${formatAmount(item.unitPrice)}`,
  ].filter(Boolean)

  return summaryParts.join(' · ')
}

const formatAmount = (value: number) => {
  try {
    return new Intl.NumberFormat(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(Number(value || 0))
  } catch {
    return String(value)
  }
}

const handleQuantityChange = (product: QuickBuySelectedProduct, event: Event) => {
  const input = event.target as HTMLInputElement | null
  emit('updateQuantity', product, input?.value || '1')
}
</script>

<template>
  <aside class="quickbuy-summary-panel" :aria-label="title">
    <div class="quickbuy-summary-panel-header">
      <div class="quickbuy-panel-heading-row quickbuy-panel-heading-row--summary">
        <h3 class="my-0 text-base font-semibold text-white">
          {{ title }}
        </h3>
      </div>
      <Icon name="lucide:shopping-bag" class="h-5 w-5 tz-text-secondary" aria-hidden="true" />
    </div>

    <div class="quickbuy-selected-list">
      <div
        v-for="slot in slots"
        :key="slot.slotKey"
        class="quickbuy-selected-step-slot"
        :class="{ 'quickbuy-selected-step-slot--filled': slot.item }"
      >
        <template v-if="slot.item">
          <img
            v-if="slot.item.thumbnail"
            :src="slot.item.thumbnail"
            :alt="slot.item.title"
            class="quickbuy-selected-card-image"
          />
          <span v-else class="quickbuy-selected-card-image quickbuy-selected-card-image--empty">
            <Icon name="lucide:image" class="h-4 w-4" aria-hidden="true" />
          </span>
          <span class="quickbuy-selected-card-content">
            <span class="quickbuy-selected-card-step">
              {{ slot.stepLabel }}
            </span>
            <span class="quickbuy-selected-card-title">{{ slot.item.title }}</span>
            <span class="quickbuy-selected-card-meta">
              {{ selectedProductSpecificationSummary(slot.item) }}
              <span v-if="slot.additionalItemCount" class="quickbuy-selected-card-extra-count">
                +{{ slot.additionalItemCount }}
              </span>
            </span>
            <span class="quickbuy-selected-card-controls">
              <button
                class="quickbuy-quantity-button"
                type="button"
                :disabled="slot.item.quantity <= 1"
                :aria-label="`${decreaseLabel} ${slot.item.title}`"
                :title="decreaseLabel"
                @click="emit('changeQuantity', slot.item, -1)"
              >
                <Icon name="lucide:minus" class="h-3.5 w-3.5" aria-hidden="true" />
              </button>
              <input
                class="quickbuy-quantity-input"
                type="number"
                min="1"
                max="999"
                step="1"
                inputmode="numeric"
                :value="slot.item.quantity"
                :aria-label="`${itemsLabel} ${slot.item.title}`"
                @change="handleQuantityChange(slot.item, $event)"
              />
              <button
                class="quickbuy-quantity-button"
                type="button"
                :disabled="slot.item.quantity >= 999"
                :aria-label="`${increaseLabel} ${slot.item.title}`"
                :title="increaseLabel"
                @click="emit('changeQuantity', slot.item, 1)"
              >
                <Icon name="lucide:plus" class="h-3.5 w-3.5" aria-hidden="true" />
              </button>
            </span>
          </span>
          <button
            class="quickbuy-selected-card-remove"
            type="button"
            :aria-label="`${removeLabel} ${slot.item.title}`"
            :title="removeLabel"
            @click="emit('removeProduct', slot.item)"
          >
            <Icon name="lucide:x" class="h-3.5 w-3.5" aria-hidden="true" />
          </button>
        </template>
        <template v-else>
          <span class="quickbuy-selected-step-slot-placeholder">
            <span class="quickbuy-selected-step-slot-placeholder__index">
              {{ slot.index }}
            </span>
            <span class="quickbuy-selected-step-slot-placeholder__label">
              {{ slot.stepLabel }}
            </span>
          </span>
        </template>
      </div>
    </div>

    <div class="quickbuy-summary-stats">
      <span
        class="quickbuy-summary-stat"
        :title="`${itemsLabel}: ${totalQty}`"
        :aria-label="`${itemsLabel}: ${totalQty}`"
      >
        <span class="quickbuy-summary-stat__icon">
          <Icon name="lucide:package" class="h-4 w-4" aria-hidden="true" />
        </span>
        <span class="quickbuy-summary-stat__content">
          <span class="quickbuy-summary-stat__label">{{ itemsLabel }}</span>
          <span class="quickbuy-summary-stat__value">{{ totalQty }}</span>
        </span>
      </span>
      <span
        class="quickbuy-summary-stat"
        :title="`${weightLabel}: ${totalWeightG}g`"
        :aria-label="`${weightLabel}: ${totalWeightG}g`"
      >
        <span class="quickbuy-summary-stat__icon">
          <Icon name="lucide:scale" class="h-4 w-4" aria-hidden="true" />
        </span>
        <span class="quickbuy-summary-stat__content">
          <span class="quickbuy-summary-stat__label">{{ weightLabel }}</span>
          <span class="quickbuy-summary-stat__value">{{ totalWeightG }}g</span>
        </span>
      </span>
      <span
        class="quickbuy-summary-stat quickbuy-summary-stat--price"
        :title="`${priceLabel}: $${formattedTotalPrice}`"
        :aria-label="`${priceLabel}: $${formattedTotalPrice}`"
      >
        <span class="quickbuy-summary-stat__icon">
          <Icon name="lucide:circle-dollar-sign" class="h-4 w-4" aria-hidden="true" />
        </span>
        <span class="quickbuy-summary-stat__content">
          <span class="quickbuy-summary-stat__label">{{ priceLabel }}</span>
          <span class="quickbuy-summary-stat__value">${{ formattedTotalPrice }}</span>
        </span>
      </span>
    </div>

    <div class="quickbuy-summary-actions">
      <button
        class="quickbuy-summary-action quickbuy-summary-action--secondary"
        type="button"
        :disabled="!hasSelectedProducts"
        @click="emit('addToCart')"
      >
        <Icon name="lucide:shopping-cart" class="h-4 w-4" aria-hidden="true" />
        <span>{{ addToCartLabel }}</span>
      </button>
      <button
        class="quickbuy-summary-action quickbuy-summary-action--primary"
        type="button"
        :disabled="!hasSelectedProducts"
        @click="emit('directPayment')"
      >
        <Icon name="lucide:credit-card" class="h-4 w-4" aria-hidden="true" />
        <span>{{ directPaymentLabel }}</span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.quickbuy-summary-panel {
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

.quickbuy-summary-panel-header {
  display: flex;
  flex: 0 0 auto;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding-bottom: 0.75rem;
  box-shadow: inset 0 -1px 0 var(--quickbuy-divider, rgba(255, 255, 255, 0.045));
}

.quickbuy-panel-heading-row {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.quickbuy-panel-heading-row--summary {
  justify-content: flex-start;
}

.quickbuy-selected-list {
  display: grid;
  grid-template-rows: repeat(5, minmax(0, 1fr));
  flex: 1;
  gap: 0.5rem;
  min-height: 0;
  overflow: hidden;
  padding: 0.75rem 0.125rem;
}

.quickbuy-selected-step-slot {
  display: flex;
  min-width: 0;
  min-height: 0;
  align-items: flex-start;
  gap: 0.5rem;
  box-sizing: border-box;
  padding: 0.5rem;
  border: 0;
  border-radius: 0.5rem;
  background:
    linear-gradient(180deg, #0b0b0d, var(--quickbuy-control-surface, #0a0a0c));
  box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.68);
}

.quickbuy-selected-step-slot--filled {
  background:
    linear-gradient(180deg, var(--quickbuy-panel-surface-raised, #17171b), #101014);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.03),
    inset 0 0 0 1px rgba(0, 0, 0, 0.56),
    0 8px 24px rgba(0, 0, 0, 0.26);
}

.quickbuy-selected-step-slot-placeholder {
  display: flex;
  min-width: 0;
  width: 100%;
  align-items: center;
  gap: 0.5rem;
  color: rgba(255, 255, 255, 0.3);
}

.quickbuy-selected-step-slot-placeholder__index {
  display: grid;
  width: 1.75rem;
  height: 1.75rem;
  flex: 0 0 auto;
  place-items: center;
  border: 0;
  border-radius: 0.375rem;
  color: rgba(255, 255, 255, 0.42);
  background: #080809;
  box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.72);
  font-size: 0.6875rem;
  font-weight: 700;
}

.quickbuy-selected-step-slot-placeholder__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.6875rem;
}

.quickbuy-selected-card-image {
  display: grid;
  width: 2.25rem;
  height: 2.25rem;
  flex: 0 0 auto;
  place-items: center;
  object-fit: cover;
  border-radius: 0.45rem;
  background: #070708;
}

.quickbuy-selected-card-image--empty {
  color: rgba(255, 255, 255, 0.45);
}

.quickbuy-selected-card-content {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 0.125rem;
}

.quickbuy-selected-card-title {
  overflow: hidden;
  color: white;
  font-size: 0.75rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quickbuy-selected-card-step {
  overflow: hidden;
  color: rgba(181, 255, 109, 0.82);
  font-size: 0.625rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quickbuy-selected-card-meta {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.35rem;
  overflow: hidden;
  color: rgba(255, 255, 255, 0.62);
  font-size: 0.625rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quickbuy-selected-card-extra-count {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  padding: 0.1rem 0.3rem;
  border-radius: 999px;
  color: rgba(181, 255, 109, 0.9);
  background: rgba(181, 255, 109, 0.08);
  font-size: 0.5625rem;
  line-height: 1;
}

.quickbuy-selected-card-controls {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin-top: 0.25rem;
}

.quickbuy-quantity-button {
  display: inline-grid;
  width: 1.5rem;
  height: 1.5rem;
  place-items: center;
  border: 0;
  border-radius: 0.375rem;
  color: white;
  background:
    linear-gradient(180deg, var(--quickbuy-control-surface-raised, #151519), #101013);
  box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.68);
  transition: background-color 160ms ease, opacity 160ms ease;
}

.quickbuy-quantity-button:hover:not(:disabled) {
  background:
    linear-gradient(180deg, #202026, #151519);
}

.quickbuy-quantity-button:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

.quickbuy-quantity-input {
  width: 2.75rem;
  height: 1.5rem;
  box-sizing: border-box;
  border: 0 !important;
  border-radius: 0.375rem;
  color: white;
  background: #070708 !important;
  background-image: none !important;
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.82);
  font-size: 0.75rem;
  text-align: center;
}

.quickbuy-quantity-input:focus {
  border: 0 !important;
  outline: none !important;
  box-shadow:
    inset 0 1px 2px rgba(0, 0, 0, 0.7),
    0 0 0 3px var(--quickbuy-focus-ring, rgba(181, 255, 109, 0.12));
}

.quickbuy-quantity-input::-webkit-inner-spin-button,
.quickbuy-quantity-input::-webkit-outer-spin-button {
  margin: 0;
  appearance: none;
}

.quickbuy-selected-card-remove {
  display: grid;
  width: 1.5rem;
  height: 1.5rem;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 999px;
  color: rgba(255, 255, 255, 0.72);
  transition: background-color 160ms ease, color 160ms ease;
}

.quickbuy-selected-card-remove:hover {
  color: white;
  background: rgba(255, 255, 255, 0.12);
}

.quickbuy-summary-stats {
  display: grid;
  flex: 0 0 auto;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.5rem;
  padding-top: 0.75rem;
  box-shadow: inset 0 1px 0 var(--quickbuy-divider, rgba(255, 255, 255, 0.045));
  color: white;
}

.quickbuy-summary-stat {
  display: flex;
  min-width: 0;
  min-height: 4.25rem;
  align-items: center;
  gap: 0.55rem;
  padding: 0.6rem 0.55rem;
  border: 0;
  border-radius: 0.65rem;
  background:
    linear-gradient(180deg, var(--quickbuy-panel-surface-raised, #17171b), #101014);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.032),
    inset 0 0 0 1px rgba(0, 0, 0, 0.56),
    0 8px 22px rgba(0, 0, 0, 0.22);
}

.quickbuy-summary-stat--price {
  background:
    linear-gradient(180deg, #232329, var(--quickbuy-panel-surface-raised, #17171b));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.052),
    inset 0 0 0 1px rgba(0, 0, 0, 0.46),
    0 12px 30px rgba(0, 0, 0, 0.3);
}

.quickbuy-summary-stat__icon {
  display: grid;
  width: 1.85rem;
  height: 1.85rem;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 999px;
  background: #24242a;
  color: rgba(255, 255, 255, 0.86);
}

.quickbuy-summary-stat--price .quickbuy-summary-stat__icon {
  background: #ffffff;
  color: #050505;
}

.quickbuy-summary-stat__content {
  display: grid;
  min-width: 0;
  gap: 0.125rem;
}

.quickbuy-summary-stat__label {
  overflow: hidden;
  color: rgba(255, 255, 255, 0.62);
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1.1;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.quickbuy-summary-stat__value {
  overflow: hidden;
  color: #ffffff;
  font-size: 1.05rem;
  font-weight: 850;
  letter-spacing: 0;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quickbuy-summary-stat--price .quickbuy-summary-stat__value {
  font-size: 1.18rem;
}

.quickbuy-summary-actions {
  display: grid;
  flex: 0 0 auto;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
  padding-top: 0.75rem;
  margin-top: 0.75rem;
  box-shadow: inset 0 1px 0 var(--quickbuy-divider, rgba(255, 255, 255, 0.045));
}

.quickbuy-summary-action {
  display: inline-flex;
  min-width: 0;
  min-height: 2.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.5rem 0.625rem;
  border: 0;
  border-radius: 0.5rem;
  color: white;
  background:
    linear-gradient(180deg, var(--quickbuy-control-surface-raised, #151519), #101013);
  box-shadow:
    0 6px 18px rgba(0, 0, 0, 0.24),
    inset 0 0 0 1px rgba(0, 0, 0, 0.68);
  font-size: 0.75rem;
  font-weight: 700;
  text-align: center;
  transition: background-color 160ms ease, opacity 160ms ease;
}

.quickbuy-summary-action:hover:not(:disabled) {
  background:
    linear-gradient(180deg, #202026, #151519);
}

.quickbuy-summary-action:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

.quickbuy-summary-action--primary {
  color: black;
  background: white;
  box-shadow:
    0 8px 20px rgba(0, 0, 0, 0.32),
    inset 0 0 0 1px rgba(0, 0, 0, 0.08);
}

.quickbuy-summary-action--primary:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.88);
}

@media (max-width: 767px) {
  .quickbuy-summary-panel {
    min-height: 15rem;
  }

  .quickbuy-selected-list {
    flex: 0 1 auto;
    max-height: 19rem;
    padding-top: 0.625rem;
    padding-bottom: 0.625rem;
  }

  .quickbuy-summary-actions {
    grid-template-columns: minmax(0, 1fr);
  }

  .quickbuy-summary-stats {
    grid-template-columns: 1fr;
  }

  .quickbuy-summary-stat {
    min-height: 3.75rem;
  }

  .quickbuy-summary-action {
    min-height: 2.75rem;
  }
}
</style>
