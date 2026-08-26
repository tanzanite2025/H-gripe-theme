<script setup lang="ts">
import { computed, nextTick, ref, toRefs, watch } from 'vue'
import type {
  QuickBuySelectedProduct,
  QuickBuySelectedProductStepSlot,
} from '~/utils/quickBuy/selection'
import QuickBuySelectedProductEditorDialog from '~/components/quick-buy/QuickBuySelectedProductEditorDialog.vue'

const props = withDefaults(defineProps<{
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
  showHeader?: boolean
  mobileStepList?: boolean
}>(), {
  showHeader: true,
  mobileStepList: false,
})

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
  showHeader,
  mobileStepList,
} = toRefs(props)

const emit = defineEmits<{
  removeProduct: [product: QuickBuySelectedProduct]
  changeQuantity: [product: QuickBuySelectedProduct, delta: number]
  updateQuantity: [product: QuickBuySelectedProduct, value: string]
  addToCart: []
  directPayment: []
  selectStep: [stepIndex: number]
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

const handleStepSlotClick = (stepIndex: number) => {
  if (!mobileStepList.value) return
  emit('selectStep', stepIndex)
}

const handleStepSlotKeydown = (event: KeyboardEvent, stepIndex: number) => {
  if (!mobileStepList.value || (event.key !== 'Enter' && event.key !== ' ')) return
  event.preventDefault()
  emit('selectStep', stepIndex)
}
</script>

<template>
  <aside
    class="quickbuy-summary-panel"
    :class="{ 'quickbuy-summary-panel--mobile-step-list': mobileStepList }"
    :aria-label="title"
  >
    <div v-if="showHeader" class="quickbuy-summary-panel-header">
      <div class="quickbuy-panel-heading-row quickbuy-panel-heading-row--summary">
        <h3 class="my-0 text-base font-semibold tz-text-primary">
          {{ title }}
        </h3>
      </div>
      <Icon name="lucide:shopping-bag" class="h-5 w-5 tz-text-secondary" aria-hidden="true" />
    </div>

    <div class="quickbuy-selected-list-shell">
      <button
        class="tz-directional-arrow tz-directional-arrow--small quickbuy-selected-list-arrow quickbuy-selected-list-arrow--previous"
        type="button"
        :disabled="!canMoveSelectedRailBackward"
        :aria-label="previousProductLabel"
        :title="previousProductLabel"
        @click="moveSelectedRail(-1)"
      >
        <Icon name="lucide:chevron-left" aria-hidden="true" />
      </button>

      <div class="quickbuy-selected-list">
        <template v-if="mobileStepList">
          <div
            v-for="slot in slots"
            :key="slot.slotKey"
            class="quickbuy-mobile-step-group"
          >
            <button
              class="quickbuy-mobile-step-trigger"
              type="button"
              :aria-label="`${slot.stepLabel}: ${slot.item?.title || title}`"
              @click="handleStepSlotClick(slot.index)"
              @keydown="handleStepSlotKeydown($event, slot.index)"
            >
              <span class="quickbuy-mobile-step-trigger__index">{{ slot.index }}</span>
              <span class="quickbuy-mobile-step-trigger__label">{{ slot.stepLabel }}</span>
              <Icon
                name="lucide:chevron-right"
                class="quickbuy-mobile-step-chevron"
                aria-hidden="true"
              />
            </button>

            <div
              v-if="slot.item"
              class="quickbuy-mobile-product-row quickbuy-mobile-product-row--filled"
            >
              <StorefrontImage
                v-if="slot.item.thumbnail"
                :src="slot.item.thumbnail"
                :alt="slot.item.title"
                class="quickbuy-selected-card-image"
                preset="thumbnail"
              />
              <span v-else class="quickbuy-selected-card-image quickbuy-selected-card-image--empty">
                <Icon name="lucide:image" class="h-4 w-4" aria-hidden="true" />
              </span>
              <span class="quickbuy-selected-card-content">
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
                class="quickbuy-selected-card-details quickbuy-mobile-product-details"
                type="button"
                :aria-label="`${viewDetailsLabel}: ${slot.item.title}`"
                :title="viewDetailsLabel"
                @click="openSelectedProductEditor(slot.item)"
              >
                <Icon name="lucide:sliders-horizontal" class="h-3.5 w-3.5" aria-hidden="true" />
                <span>{{ viewDetailsLabel }}</span>
              </button>
            </div>
            <button
              v-else
              class="quickbuy-mobile-product-row quickbuy-mobile-product-row--empty"
              type="button"
              :aria-label="`${slot.stepLabel}: ${title}`"
              :title="`${slot.stepLabel}: ${title}`"
              @click="handleStepSlotClick(slot.index)"
            >
              <Icon name="lucide:plus" class="quickbuy-mobile-product-plus" aria-hidden="true" />
            </button>
          </div>
        </template>
        <template v-else>
          <div
            v-for="slot in slots"
            :key="slot.slotKey"
            ref="selectedRailItems"
            class="quickbuy-selected-step-slot"
            :class="{ 'quickbuy-selected-step-slot--filled': slot.item }"
          >
            <template v-if="slot.item">
              <StorefrontImage
                v-if="slot.item.thumbnail"
                :src="slot.item.thumbnail"
                :alt="slot.item.title"
                class="quickbuy-selected-card-image"
                preset="thumbnail"
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
        </template>
      </div>

      <button
        class="tz-directional-arrow tz-directional-arrow--small quickbuy-selected-list-arrow quickbuy-selected-list-arrow--next"
        type="button"
        :disabled="!canMoveSelectedRailForward"
        :aria-label="nextProductLabel"
        :title="nextProductLabel"
        @click="moveSelectedRail(1)"
      >
        <Icon name="lucide:chevron-right" aria-hidden="true" />
      </button>
    </div>

    <div
      class="quickbuy-summary-footer"
      :class="{ 'quickbuy-summary-footer--mobile-pinned': mobileStepList }"
    >
      <!-- Summary stat cards: items / weight / price. Labels are intentionally not rendered visually; title/aria-label keep the names available for QA and accessibility. -->
      <div class="quickbuy-summary-stats">
        <span
          class="quickbuy-summary-stat quickbuy-summary-stat--items"
          :title="`${itemsLabel}: ${totalQty}`"
          :aria-label="`${itemsLabel}: ${totalQty}`"
        >
          <span class="quickbuy-summary-stat__icon">
            <Icon name="lucide:package" class="h-4 w-4" aria-hidden="true" />
          </span>
          <span class="quickbuy-summary-stat__content">
            <span class="quickbuy-summary-stat__value">{{ totalQty }}</span>
          </span>
        </span>
        <span
          class="quickbuy-summary-stat quickbuy-summary-stat--weight"
          :title="`${weightLabel}: ${totalWeightG}g`"
          :aria-label="`${weightLabel}: ${totalWeightG}g`"
        >
          <span class="quickbuy-summary-stat__icon">
            <Icon name="lucide:scale" class="h-4 w-4" aria-hidden="true" />
          </span>
          <span class="quickbuy-summary-stat__content">
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
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--quickbuy-panel-surface, var(--tz-card-surface));
  box-shadow:
    0 16px 42px rgba(20, 32, 43, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 -1px 0 var(--quickbuy-dark-edge, rgba(20, 32, 43, 0.12));
}

.quickbuy-summary-panel--mobile-step-list {
  min-height: 0;
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
    var(--quickbuy-control-surface-raised, var(--tz-surface-subtle));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 -1px 0 var(--tz-border-subtle);
}

.quickbuy-selected-step-slot--filled {
  background:
    var(--quickbuy-panel-surface-raised, var(--tz-surface-muted));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 -1px 0 var(--tz-border-subtle),
    0 8px 24px rgba(20, 32, 43, 0.08);
}

.quickbuy-selected-step-slot-placeholder {
  display: flex;
  min-width: 0;
  width: 100%;
  align-items: center;
  gap: 0.5rem;
  color: var(--tz-text-disabled);
}

.quickbuy-selected-step-slot-placeholder__index {
  display: grid;
  width: 1.75rem;
  height: 1.75rem;
  flex: 0 0 auto;
  place-items: center;
  border: 0;
  border-radius: 0.375rem;
  color: var(--tz-text-muted);
  background: var(--quickbuy-control-surface, var(--tz-input-surface));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 -1px 0 var(--tz-border-subtle);
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
  background: var(--quickbuy-control-surface, var(--tz-input-surface));
}

.quickbuy-selected-card-image--empty {
  color: var(--tz-text-muted);
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
  color: var(--tz-text-primary);
  font-size: 0.75rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quickbuy-selected-card-step {
  overflow: hidden;
  color: rgba(5, 150, 105, 0.82);
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
  color: var(--tz-text-muted);
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
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
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
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 -1px 0 var(--tz-border-subtle);
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
    var(--tz-surface-muted);
  transform: translateY(-1px);
}

.quickbuy-mobile-step-chevron {
  display: none;
}

.quickbuy-selected-card-extra-count {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  padding: 0.1rem 0.3rem;
  border-radius: 999px;
  color: rgba(5, 150, 105, 0.9);
  background: rgba(5, 150, 105, 0.08);
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
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
  box-shadow: inset 0 0 0 1px var(--tz-border-subtle);
  transition: background-color 160ms ease, opacity 160ms ease;
}

.quickbuy-quantity-button:hover:not(:disabled) {
  background:
    var(--tz-surface-muted);
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
  color: var(--tz-text-primary);
  background: var(--tz-input-surface) !important;
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
    0 0 0 3px var(--quickbuy-focus-ring, rgba(5, 150, 105, 0.12));
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
  color: var(--tz-text-muted);
  transition: background-color 160ms ease, color 160ms ease;
}

.quickbuy-selected-card-remove:hover {
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
}

.quickbuy-summary-stats {
  display: grid;
  flex: 0 0 auto;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.5rem;
  padding-top: 0.75rem;
  box-shadow: inset 0 1px 0 var(--quickbuy-divider, rgba(255, 255, 255, 0.045));
  color: var(--tz-text-primary);
}

/* Stat card selectors kept explicit for QA: --items, --weight, --price map to the three label-free summary cards. */
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
    var(--quickbuy-panel-surface-raised, var(--tz-surface-muted));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 -1px 0 var(--tz-border-subtle),
    0 8px 22px rgba(20, 32, 43, 0.08);
}

.quickbuy-summary-stat--price {
  background:
    var(--quickbuy-panel-surface-raised, var(--tz-surface-muted));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 -1px 0 var(--tz-border-subtle),
    0 12px 30px rgba(20, 32, 43, 0.08);
}

.quickbuy-summary-stat__icon {
  display: grid;
  width: 1.85rem;
  height: 1.85rem;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 999px;
  background: var(--tz-surface-muted);
  color: var(--tz-text-secondary);
}

.quickbuy-summary-stat--price .quickbuy-summary-stat__icon {
  background: #ffffff;
  color: var(--tz-text-primary);
}

.quickbuy-summary-stat__content {
  display: grid;
  min-width: 0;
  gap: 0.125rem;
}

.quickbuy-summary-stat__label {
  overflow: hidden;
  color: var(--tz-text-muted);
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1.1;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.quickbuy-summary-stat__value {
  overflow: hidden;
  color: var(--tz-text-primary);
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
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
  box-shadow:
    0 6px 18px rgba(20, 32, 43, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 -1px 0 var(--tz-border-subtle);
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
    var(--tz-surface-muted);
}

.quickbuy-summary-action:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

.quickbuy-summary-action--primary {
  color: var(--tz-action-primary-foreground);
  background: var(--tz-action-primary);
  border: 1px solid var(--tz-action-primary);
  box-shadow:
    0 8px 20px rgba(0, 0, 0, 0.32),
    inset 0 0 0 1px rgba(0, 0, 0, 0.08);
}

.quickbuy-summary-action--primary:hover:not(:disabled) {
  background: var(--tz-action-primary-hover);
  border-color: var(--tz-action-primary-hover);
}

@media (min-width: 768px) and (max-width: 1100px) {
  .quickbuy-summary-stat {
    min-height: 4.75rem;
    flex-direction: column;
    justify-content: center;
    gap: 0.25rem;
    padding: 0.45rem 0.25rem;
  }

  .quickbuy-summary-stat__icon {
    width: 1.5rem;
    height: 1.5rem;
  }

  .quickbuy-summary-stat__icon :deep(svg) {
    width: 0.8rem;
    height: 0.8rem;
  }

  .quickbuy-summary-stat__content {
    width: 100%;
    justify-items: center;
  }

  .quickbuy-summary-stat__label {
    max-width: 100%;
    overflow: visible;
    font-size: 0.6rem;
    text-align: center;
    text-overflow: clip;
  }

  .quickbuy-summary-stat__value,
  .quickbuy-summary-stat--price .quickbuy-summary-stat__value {
    max-width: 100%;
    overflow: visible;
    font-size: 0.95rem;
    text-align: center;
    text-overflow: clip;
  }
}

@media (max-width: 767px) {
  .quickbuy-summary-panel {
    min-height: 15rem;
    padding: 0.5rem;
  }

  .quickbuy-selected-list-shell {
    display: grid;
    grid-template-columns: 1.75rem minmax(0, 1fr) 1.75rem;
    align-items: center;
    gap: 0.25rem;
    flex: 0 1 auto;
    min-height: 0;
  }

  .quickbuy-selected-list-arrow {
    display: inline-grid;
  }

  .quickbuy-summary-panel--mobile-step-list {
    height: 100%;
    overflow: hidden;
  }

  .quickbuy-summary-panel--mobile-step-list .quickbuy-selected-list-shell {
    display: flex;
    flex: 0 0 auto;
    width: 100%;
    align-items: stretch;
    min-height: 0;
  }

  .quickbuy-summary-panel--mobile-step-list .quickbuy-selected-list-arrow {
    display: none;
  }

  .quickbuy-summary-panel--mobile-step-list .quickbuy-selected-list {
    display: flex;
    flex: 0 0 auto;
    width: 100%;
    flex-direction: column;
    gap: 0.3rem;
    justify-content: flex-start;
    min-height: 0;
    padding: 0;
    overflow: visible;
  }

  .quickbuy-mobile-step-group {
    display: grid;
    width: 100%;
    min-width: 0;
    gap: 0.2rem;
  }

  .quickbuy-mobile-step-trigger {
    display: flex;
    width: 100%;
    min-width: 0;
    min-height: 2.25rem;
    align-items: center;
    gap: 0.375rem;
    padding: 0.25rem 0.4rem;
    border: 0;
    border-radius: 0.45rem;
    color: var(--tz-text-secondary);
    background:
      var(--quickbuy-control-surface-raised, var(--tz-surface-subtle));
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.8),
      inset 0 -1px 0 var(--tz-border-subtle);
    cursor: pointer;
    text-align: left;
    transition: background-color 160ms ease, color 160ms ease, transform 160ms ease;
  }

  .quickbuy-mobile-step-trigger:hover {
    color: var(--tz-text-primary);
    background: var(--tz-surface-muted);
    transform: translateY(-1px);
  }

  .quickbuy-mobile-step-trigger:focus-visible {
    outline: 2px solid rgba(5, 150, 105, 0.72);
    outline-offset: 2px;
  }

  .quickbuy-mobile-step-trigger__index {
    display: grid;
    width: 1.35rem;
    height: 1.35rem;
    flex: 0 0 auto;
    place-items: center;
    border-radius: 0.35rem;
    color: var(--tz-text-muted);
    background: var(--quickbuy-control-surface, var(--tz-input-surface));
    font-size: 0.625rem;
    font-weight: 800;
  }

  .quickbuy-mobile-step-trigger__label {
    min-width: 0;
    flex: 1;
    overflow: hidden;
    font-size: 0.6875rem;
    font-weight: 750;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .quickbuy-mobile-step-trigger .quickbuy-mobile-step-chevron {
    display: block;
    width: 0.9rem;
    height: 0.9rem;
    flex: 0 0 auto;
    color: var(--tz-text-muted);
  }

  .quickbuy-mobile-product-row {
    display: flex;
    width: 100%;
    min-width: 0;
    min-height: 4.25rem;
    box-sizing: border-box;
    align-items: center;
    gap: 0.375rem;
    padding: 0.45rem 0.55rem;
    border-radius: 0.45rem;
  }

  .quickbuy-mobile-product-row--filled {
    background:
      var(--quickbuy-panel-surface-raised, var(--tz-surface-muted));
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.8),
      inset 0 -1px 0 var(--tz-border-subtle);
  }

  .quickbuy-mobile-product-row--empty {
    justify-content: center;
    border: 1px dashed rgba(5, 150, 105, 0.3);
    color: rgba(5, 150, 105, 0.9);
    background: rgba(5, 150, 105, 0.035);
    cursor: pointer;
    transition: background-color 160ms ease, border-color 160ms ease, transform 160ms ease;
  }

  .quickbuy-mobile-product-row--empty:hover {
    border-color: rgba(5, 150, 105, 0.62);
    background: rgba(5, 150, 105, 0.09);
    transform: translateY(-1px);
  }

  .quickbuy-mobile-product-row--empty:focus-visible {
    outline: 2px solid rgba(5, 150, 105, 0.72);
    outline-offset: 2px;
  }

  .quickbuy-mobile-product-plus {
    width: 1.35rem;
    height: 1.35rem;
  }

  .quickbuy-mobile-product-row .quickbuy-selected-card-image {
    width: 3rem;
    height: 3rem;
  }

  .quickbuy-mobile-product-row .quickbuy-selected-card-content {
    gap: 0.0625rem;
  }

  .quickbuy-mobile-product-row .quickbuy-selected-card-title {
    font-size: 0.75rem;
  }

  .quickbuy-mobile-product-row .quickbuy-selected-card-meta {
    font-size: 0.625rem;
  }

  .quickbuy-mobile-product-row .quickbuy-selected-card-quantity {
    min-width: 2rem;
    height: 2rem;
    padding: 0 0.25rem;
    font-size: 0.75rem;
  }

  .quickbuy-mobile-product-details {
    width: 2rem;
    min-width: 2rem;
    min-height: 2rem;
    padding: 0.35rem;
  }

  .quickbuy-mobile-product-details > span {
    display: none;
  }

  .quickbuy-summary-panel--mobile-step-list .quickbuy-selected-step-slot {
    width: 100%;
    min-width: 0;
    flex: 0 0 auto;
  }

  .quickbuy-summary-footer--mobile-pinned {
    position: sticky;
    bottom: 0;
    z-index: 3;
    display: block;
    flex: 0 0 auto;
    width: 100%;
    box-sizing: border-box;
    padding-top: 0.5rem;
    padding-bottom: max(0.25rem, var(--tz-safe-area-bottom, 0px));
    margin-top: auto;
    border-top: 1px solid var(--quickbuy-divider, rgba(255, 255, 255, 0.045));
    background: rgba(11, 13, 18, 0.96);
    backdrop-filter: blur(10px);
  }

  .quickbuy-summary-footer--mobile-pinned .quickbuy-summary-stats {
    padding-top: 0;
    box-shadow: none;
  }

  .quickbuy-summary-footer--mobile-pinned .quickbuy-summary-actions {
    margin-top: 0.5rem;
    padding-top: 0.5rem;
  }

  .quickbuy-selected-list {
    display: flex;
    flex: 0 1 auto;
    max-height: none;
    gap: 0.25rem;
    min-height: 6rem;
    padding: 0.25rem 0;
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
    gap: 0.25rem;
    padding-top: 0.5rem;
    margin-top: 0.5rem;
  }

  .quickbuy-summary-stats {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.25rem;
    padding-top: 0.5rem;
  }

  .quickbuy-summary-stat {
    min-height: 3.25rem;
    justify-content: center;
    gap: 0.3rem;
    padding: 0.35rem 0.25rem;
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
    min-height: 2.5rem;
    gap: 0.25rem;
    padding: 0.45rem 0.3rem;
    font-size: 0.6875rem;
  }

  .quickbuy-summary-action :deep(svg) {
    width: 0.95rem;
    height: 0.95rem;
    flex: 0 0 auto;
  }
}
</style>
