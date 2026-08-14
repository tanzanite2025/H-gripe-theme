<script setup lang="ts">
import { computed, nextTick, ref, toRefs, watch } from 'vue'
import type {
  QuickBuySelectedProduct,
  QuickBuySelectedProductStepSlot,
} from '~/utils/quickBuy/selection'
import QuickBuySelectedProductEditorDialog from '~/components/quick-buy/QuickBuySelectedProductEditorDialog.vue'

const props = defineProps<{
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
  viewDetailsLabel: string
  closeLabel: string
  doneLabel: string
  previousProductLabel: string
  nextProductLabel: string
}>()

const {
  title,
  slots,
  totalQty,
  totalWeightG,
  formattedTotalPrice,
  hasSelectedProducts,
  itemsLabel,
  weightLabel,
  priceLabel,
  addToCartLabel,
  directPaymentLabel,
  decreaseLabel,
  increaseLabel,
  removeLabel,
  viewDetailsLabel,
  closeLabel,
  doneLabel,
  previousProductLabel,
  nextProductLabel,
} = toRefs(props)

const emit = defineEmits<{
  removeProduct: [product: QuickBuySelectedProduct]
  changeQuantity: [product: QuickBuySelectedProduct, delta: number]
  updateQuantity: [product: QuickBuySelectedProduct, value: string]
  addToCart: []
  directPayment: []
}>()

const selectedRailIndex = ref(0)
const selectedRailItems = ref<HTMLElement[]>([])

const canMoveSelectedRailBackward = computed(() => selectedRailIndex.value > 0)
const canMoveSelectedRailForward = computed(() =>
  selectedRailIndex.value < Math.max(0, slots.value.length - 1),
)

const scrollToSelectedRailItem = async () => {
  await nextTick()
  selectedRailItems.value[selectedRailIndex.value]?.scrollIntoView({
    behavior: 'smooth',
    block: 'nearest',
    inline: 'center',
  })
}

const moveSelectedRail = (delta: number) => {
  const lastIndex = Math.max(0, slots.value.length - 1)
  const nextIndex = Math.min(lastIndex, Math.max(0, selectedRailIndex.value + delta))
  if (nextIndex === selectedRailIndex.value) return
  selectedRailIndex.value = nextIndex
  scrollToSelectedRailItem()
}

watch(() => slots.value.length, (nextLength) => {
  selectedRailIndex.value = Math.min(selectedRailIndex.value, Math.max(0, nextLength - 1))
})

const editingProduct = ref<QuickBuySelectedProduct | null>(null)

const editingStepLabel = computed(() =>
  slots.value.find(slot =>
    slot.item?.stepKey === editingProduct.value?.stepKey
    && slot.item?.productId === editingProduct.value?.productId
    && slot.item?.variantId === editingProduct.value?.variantId,
  )?.stepLabel || '',
)

const openSelectedProductEditor = (product: QuickBuySelectedProduct) => {
  editingProduct.value = product
}

const closeSelectedProductEditor = () => {
  editingProduct.value = null
}

const removeEditingProduct = (product: QuickBuySelectedProduct) => {
  emit('removeProduct', product)
  closeSelectedProductEditor()
}

const handleEditorChangeQuantity = (
  product: QuickBuySelectedProduct,
  delta: number,
) => {
  emit('changeQuantity', product, delta)
}

const handleEditorUpdateQuantity = (
  product: QuickBuySelectedProduct,
  value: string,
) => {
  emit('updateQuantity', product, value)
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

    <div class="quickbuy-selected-list-shell">
      <button
        class="quickbuy-selected-list-arrow quickbuy-selected-list-arrow--previous"
        type="button"
        :disabled="!canMoveSelectedRailBackward"
        :aria-label="previousProductLabel"
        :title="previousProductLabel"
        @click="moveSelectedRail(-1)"
      >
        <Icon name="lucide:chevron-left" class="h-4 w-4" aria-hidden="true" />
      </button>

      <div class="quickbuy-selected-list">
      <div
        v-for="slot in slots"
        :key="slot.slotKey"
        ref="selectedRailItems"
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
              <span>{{ itemsLabel }}: {{ slot.item.quantity }}</span>
              <span v-if="slot.additionalItemCount" class="quickbuy-selected-card-extra-count">
                +{{ slot.additionalItemCount }}
              </span>
            </span>
          </span>
          <span
            class="quickbuy-selected-card-quantity"
            :aria-label="`${itemsLabel} ${slot.item.title}: ${slot.item.quantity}`"
          >
            {{ slot.item.quantity }}
          </span>
          <button
            class="quickbuy-selected-card-details"
            type="button"
            :aria-label="`${viewDetailsLabel}: ${slot.item.title}`"
            :title="viewDetailsLabel"
            @click="openSelectedProductEditor(slot.item)"
          >
            <Icon name="lucide:sliders-horizontal" class="h-3.5 w-3.5" aria-hidden="true" />
            <span>{{ viewDetailsLabel }}</span>
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

      <button
        class="quickbuy-selected-list-arrow quickbuy-selected-list-arrow--next"
        type="button"
        :disabled="!canMoveSelectedRailForward"
        :aria-label="nextProductLabel"
        :title="nextProductLabel"
        @click="moveSelectedRail(1)"
      >
        <Icon name="lucide:chevron-right" class="h-4 w-4" aria-hidden="true" />
      </button>
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

    <QuickBuySelectedProductEditorDialog
      :product="editingProduct"
      :step-label="editingStepLabel"
      :quantity-label="itemsLabel"
      :view-details-label="viewDetailsLabel"
      :close-label="closeLabel"
      :done-label="doneLabel"
      :decrease-label="decreaseLabel"
      :increase-label="increaseLabel"
      :remove-label="removeLabel"
      @close="closeSelectedProductEditor"
      @remove-product="removeEditingProduct"
      @change-quantity="handleEditorChangeQuantity"
      @update-quantity="handleEditorUpdateQuantity"
    />
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
  border: 1px solid rgba(255, 255, 255, 0.035);
  border-radius: 0.75rem;
  background: #0b0d12;
  box-shadow:
    0 16px 42px rgba(0, 0, 0, 0.34),
    inset 0 1px 0 rgba(255, 255, 255, 0.045),
    inset 0 -1px 0 var(--quickbuy-dark-edge, rgba(0, 0, 0, 0.5));
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

.quickbuy-selected-list-shell {
  display: flex;
  flex: 1;
  min-width: 0;
  min-height: 0;
}

.quickbuy-selected-list-arrow {
  display: none;
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
  min-width: 0;
  gap: 0.5rem;
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
    linear-gradient(180deg, var(--quickbuy-control-surface-raised, #171920), var(--quickbuy-control-surface, #0d0f14));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.035),
    inset 0 -1px 0 rgba(0, 0, 0, 0.34);
}

.quickbuy-selected-step-slot--filled {
  background:
    linear-gradient(180deg, var(--quickbuy-panel-surface-raised, #1c1e25), var(--quickbuy-panel-surface, #15171d));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.045),
    inset 0 -1px 0 rgba(0, 0, 0, 0.36),
    0 8px 24px rgba(0, 0, 0, 0.22);
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
  background: var(--quickbuy-control-surface, #0d0f14);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.03),
    inset 0 -1px 0 rgba(0, 0, 0, 0.38);
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
  background: var(--quickbuy-control-surface, #0d0f14);
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

.quickbuy-selected-card-quantity {
  display: inline-grid;
  min-width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  place-items: center;
  padding: 0 0.4rem;
  border-radius: 0.5rem;
  color: #fff;
  background: rgba(255, 255, 255, 0.08);
  font-size: 0.8rem;
  font-weight: 800;
}

.quickbuy-selected-card-details {
  display: inline-flex;
  min-width: 0;
  min-height: 2rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
  padding: 0.35rem 0.55rem;
  border: 0;
  border-radius: 0.45rem;
  color: #fff;
  background:
    linear-gradient(180deg, var(--quickbuy-control-surface-raised, #171920), var(--quickbuy-control-surface, #0d0f14));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.035),
    inset 0 -1px 0 rgba(0, 0, 0, 0.34);
  font-size: 0.68rem;
  font-weight: 700;
  transition: background-color 160ms ease, transform 160ms ease;
}

.quickbuy-selected-card-details > span {
  max-width: 7rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quickbuy-selected-card-details:hover {
  background:
    linear-gradient(180deg, #32343d, #24262e);
  transform: translateY(-1px);
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
    linear-gradient(180deg, var(--quickbuy-panel-surface-raised, #1c1e25), var(--quickbuy-panel-surface, #15171d));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.045),
    inset 0 -1px 0 rgba(0, 0, 0, 0.36),
    0 8px 22px rgba(0, 0, 0, 0.2);
}

.quickbuy-summary-stat--price {
  background:
    linear-gradient(180deg, #252831, var(--quickbuy-panel-surface-raised, #1c1e25));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    inset 0 -1px 0 rgba(0, 0, 0, 0.32),
    0 12px 30px rgba(0, 0, 0, 0.24);
}

.quickbuy-summary-stat__icon {
  display: grid;
  width: 1.85rem;
  height: 1.85rem;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 999px;
  background: #2b2e38;
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
    linear-gradient(180deg, var(--quickbuy-control-surface-raised, #171920), var(--quickbuy-control-surface, #0d0f14));
  box-shadow:
    0 6px 18px rgba(0, 0, 0, 0.24),
    inset 0 1px 0 rgba(255, 255, 255, 0.035),
    inset 0 -1px 0 rgba(0, 0, 0, 0.36);
  font-size: 0.75rem;
  font-weight: 700;
  text-align: center;
  transition: background-color 160ms ease, opacity 160ms ease;
}

.quickbuy-summary-action > span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quickbuy-summary-action:hover:not(:disabled) {
  background:
    linear-gradient(180deg, #32343d, #24262e);
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

  .quickbuy-selected-list-shell {
    display: grid;
    grid-template-columns: 1.75rem minmax(0, 1fr) 1.75rem;
    align-items: center;
    gap: 0.375rem;
    flex: 0 1 auto;
    min-height: 0;
  }

  .quickbuy-selected-list-arrow {
    display: inline-grid;
    width: 1.75rem;
    height: 2rem;
    place-items: center;
    border: 0;
    border-radius: 999px;
    color: rgba(255, 255, 255, 0.86);
    background:
      linear-gradient(180deg, var(--quickbuy-control-surface-raised, #25272f), var(--quickbuy-control-surface, #1b1c23));
    box-shadow:
      0 5px 14px rgba(0, 0, 0, 0.2),
      inset 0 0 0 1px rgba(255, 255, 255, 0.04);
    transition: background-color 160ms ease, opacity 160ms ease, transform 160ms ease;
  }

  .quickbuy-selected-list-arrow:hover:not(:disabled) {
    background:
      linear-gradient(180deg, #32343d, #24262e);
    transform: translateY(-1px);
  }

  .quickbuy-selected-list-arrow:disabled {
    cursor: not-allowed;
    opacity: 0.28;
  }

  .quickbuy-selected-list {
    display: flex;
    flex: 0 1 auto;
    max-height: none;
    gap: 0.5rem;
    min-height: 7.25rem;
    padding: 0.625rem 0;
    overflow: hidden;
    scroll-behavior: smooth;
    scroll-snap-type: x mandatory;
  }

  .quickbuy-selected-step-slot {
    width: 100%;
    min-width: 100%;
    flex: 0 0 100%;
    scroll-snap-align: center;
  }

  .quickbuy-summary-actions {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.375rem;
    padding-top: 0.625rem;
    margin-top: 0.625rem;
  }

  .quickbuy-summary-stats {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.375rem;
    padding-top: 0.625rem;
  }

  .quickbuy-summary-stat {
    min-height: 3.25rem;
    justify-content: center;
    gap: 0.375rem;
    padding: 0.45rem 0.375rem;
    border-radius: 0.5rem;
  }

  .quickbuy-summary-stat__icon {
    width: 1.35rem;
    height: 1.35rem;
  }

  .quickbuy-summary-stat__icon :deep(svg) {
    width: 0.85rem;
    height: 0.85rem;
  }

  .quickbuy-summary-stat__content {
    justify-items: start;
  }

  .quickbuy-summary-stat__label {
    max-width: 100%;
    font-size: 0.5625rem;
    line-height: 1;
  }

  .quickbuy-summary-stat__value,
  .quickbuy-summary-stat--price .quickbuy-summary-stat__value {
    max-width: 100%;
    font-size: 0.8125rem;
    line-height: 1.1;
  }

  .quickbuy-summary-action {
    min-height: 2.625rem;
    gap: 0.3rem;
    padding: 0.5rem 0.375rem;
    font-size: 0.6875rem;
  }

  .quickbuy-summary-action :deep(svg) {
    width: 0.95rem;
    height: 0.95rem;
    flex: 0 0 auto;
  }
}
</style>
