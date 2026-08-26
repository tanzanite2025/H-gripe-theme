<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { QuickBuySelectedProduct } from '~/utils/quickBuy/selection'

const props = defineProps<{
  product: QuickBuySelectedProduct | null
  stepLabel: string
  quantityLabel: string
  viewDetailsLabel: string
  closeLabel: string
  doneLabel: string
  decreaseLabel: string
  increaseLabel: string
  removeLabel: string
}>()

const emit = defineEmits<{
  close: []
  removeProduct: [product: QuickBuySelectedProduct]
  changeQuantity: [product: QuickBuySelectedProduct, delta: number]
  updateQuantity: [product: QuickBuySelectedProduct, value: string]
}>()

const draftQuantity = ref(1)

const normalizeQuantity = (value: unknown) => {
  const parsed = Math.floor(Number(value))
  if (!Number.isFinite(parsed)) return 1
  return Math.min(999, Math.max(1, parsed))
}

watch(
  () => props.product?.quantity,
  (quantity) => {
    draftQuantity.value = normalizeQuantity(quantity)
  },
  { immediate: true },
)

const editorAriaLabel = computed(() =>
  props.product
    ? `${props.viewDetailsLabel}: ${props.product.title}`
    : props.viewDetailsLabel,
)

const formatAmount = (value: number, currency: string) => {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency: currency || 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(Number(value || 0))
  } catch {
    return `${currency || 'USD'} ${Number(value || 0).toFixed(2)}`
  }
}

const changeDraftQuantity = (delta: number) => {
  if (!props.product) return
  const nextQuantity = normalizeQuantity(draftQuantity.value + delta)
  if (nextQuantity === draftQuantity.value) return
  draftQuantity.value = nextQuantity
  emit('changeQuantity', props.product, delta)
}

const handleQuantityInput = (event: Event) => {
  if (!props.product) return
  const value = (event.target as HTMLInputElement | null)?.value || '1'
  draftQuantity.value = normalizeQuantity(value)
  emit('updateQuantity', props.product, String(draftQuantity.value))
}

const handleQuantityBlur = () => {
  if (!props.product) return
  draftQuantity.value = normalizeQuantity(draftQuantity.value)
  emit('updateQuantity', props.product, String(draftQuantity.value))
}
</script>

<template>
  <Teleport to="body">
    <Transition name="quickbuy-selected-product-editor-fade">
      <div
        v-if="product"
        class="quickbuy-selected-product-editor-mask tz-mobile-dialog-mask"
        tabindex="-1"
        @click.self="emit('close')"
        @keydown.esc="emit('close')"
      >
        <div class="quickbuy-selected-product-editor-backdrop" aria-hidden="true"></div>

        <section
          class="quickbuy-selected-product-editor-shell tz-mobile-dialog-surface"
          role="dialog"
          aria-modal="true"
          :aria-label="editorAriaLabel"
        >
          <header class="quickbuy-selected-product-editor-header">
            <div class="quickbuy-selected-product-editor-heading">
              <span class="quickbuy-selected-product-editor-eyebrow">
                {{ stepLabel || viewDetailsLabel }}
              </span>
              <h2>{{ product.title }}</h2>
            </div>
            <button
              type="button"
              class="quickbuy-selected-product-editor-close"
              :aria-label="closeLabel"
              :title="closeLabel"
              @click="emit('close')"
            >
              <Icon name="lucide:x" class="h-5 w-5" aria-hidden="true" />
            </button>
          </header>

          <div class="quickbuy-selected-product-editor-content">
            <div class="quickbuy-selected-product-editor-product">
              <StorefrontImage
                v-if="product.thumbnail"
                :src="product.thumbnail"
                :alt="product.title"
                class="quickbuy-selected-product-editor-image"
                preset="thumbnail"
              />
              <span v-else class="quickbuy-selected-product-editor-image quickbuy-selected-product-editor-image--empty">
                <Icon name="lucide:image" class="h-6 w-6" aria-hidden="true" />
              </span>
              <div class="quickbuy-selected-product-editor-product-copy">
                <strong>{{ product.title }}</strong>
                <span v-if="product.sku">{{ product.sku }}</span>
              </div>
              <strong class="quickbuy-selected-product-editor-price">
                {{ formatAmount(product.unitPrice, product.currency) }}
              </strong>
            </div>

            <div class="quickbuy-selected-product-editor-facts">
              <span v-if="product.sku">
                <small>SKU</small>
                <strong>{{ product.sku }}</strong>
              </span>
              <span>
                <small>Weight</small>
                <strong>{{ product.weightGrams }}g</strong>
              </span>
              <span>
                <small>Unit price</small>
                <strong>{{ formatAmount(product.unitPrice, product.currency) }}</strong>
              </span>
            </div>

            <div class="quickbuy-selected-product-editor-quantity">
              <span>{{ quantityLabel }}</span>
              <div class="quickbuy-selected-product-editor-stepper" role="group" :aria-label="quantityLabel">
                <button
                  type="button"
                  :disabled="draftQuantity <= 1"
                  :aria-label="`${decreaseLabel} ${product.title}`"
                  :title="decreaseLabel"
                  @click="changeDraftQuantity(-1)"
                >
                  <Icon name="lucide:minus" class="h-4 w-4" aria-hidden="true" />
                </button>
                <input
                  v-model="draftQuantity"
                  type="number"
                  min="1"
                  max="999"
                  step="1"
                  inputmode="numeric"
                  :aria-label="quantityLabel"
                  @change="handleQuantityInput"
                  @blur="handleQuantityBlur"
                />
                <button
                  type="button"
                  :disabled="draftQuantity >= 999"
                  :aria-label="`${increaseLabel} ${product.title}`"
                  :title="increaseLabel"
                  @click="changeDraftQuantity(1)"
                >
                  <Icon name="lucide:plus" class="h-4 w-4" aria-hidden="true" />
                </button>
              </div>
            </div>
          </div>

          <footer class="quickbuy-selected-product-editor-footer">
            <button
              type="button"
              class="quickbuy-selected-product-editor-action quickbuy-selected-product-editor-action--danger"
              @click="emit('removeProduct', product)"
            >
              <Icon name="lucide:trash-2" class="h-4 w-4" aria-hidden="true" />
              <span>{{ removeLabel }}</span>
            </button>
            <button
              type="button"
              class="quickbuy-selected-product-editor-action quickbuy-selected-product-editor-action--primary"
              @click="emit('close')"
            >
              <Icon name="lucide:check" class="h-4 w-4" aria-hidden="true" />
              <span>{{ doneLabel }}</span>
            </button>
          </footer>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.quickbuy-selected-product-editor-mask {
  position: fixed;
  inset: 0;
  z-index: 10003;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  padding: 1rem;
}

.quickbuy-selected-product-editor-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(3, 4, 6, 0.68);
  backdrop-filter: blur(5px);
}

.quickbuy-selected-product-editor-shell {
  position: relative;
  z-index: 1;
  display: flex;
  width: min(34rem, calc(100vw - 2rem));
  max-height: min(32rem, calc(100vh - 2rem));
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: var(--tz-card-surface);
  box-shadow:
    0 24px 72px rgba(20, 32, 43, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 -1px 0 var(--tz-border-subtle);
}

.quickbuy-selected-product-editor-header {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.9rem 1rem;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.quickbuy-selected-product-editor-heading {
  min-width: 0;
}

.quickbuy-selected-product-editor-eyebrow {
  display: block;
  margin-bottom: 0.2rem;
  color: rgba(5, 150, 105, 0.78);
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.quickbuy-selected-product-editor-heading h2 {
  margin: 0;
  overflow: hidden;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quickbuy-selected-product-editor-close {
  display: inline-grid;
  width: 2.25rem;
  height: 2.25rem;
  flex: 0 0 auto;
  place-items: center;
  border: 0;
  border-radius: 999px;
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
  transition: background-color 160ms ease;
}

.quickbuy-selected-product-editor-close:hover {
  background: var(--tz-surface-muted);
}

.quickbuy-selected-product-editor-content {
  display: grid;
  gap: 0.85rem;
  min-height: 0;
  overflow-y: auto;
  padding: 1rem;
}

.quickbuy-selected-product-editor-product {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.75rem;
  padding: 0.7rem;
  border-radius: 0.7rem;
  background: var(--tz-surface-subtle);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    inset 0 -1px 0 var(--tz-border-subtle);
}

.quickbuy-selected-product-editor-image {
  display: grid;
  width: 3.5rem;
  height: 3.5rem;
  flex: 0 0 auto;
  place-items: center;
  object-fit: cover;
  border-radius: 0.55rem;
  background: var(--tz-image-loading-surface);
}

.quickbuy-selected-product-editor-image--empty {
  color: var(--tz-text-muted);
}

.quickbuy-selected-product-editor-product-copy {
  display: grid;
  min-width: 0;
  flex: 1;
  gap: 0.2rem;
}

.quickbuy-selected-product-editor-product-copy strong,
.quickbuy-selected-product-editor-product-copy span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quickbuy-selected-product-editor-product-copy strong {
  color: var(--tz-text-primary);
  font-size: 0.85rem;
}

.quickbuy-selected-product-editor-product-copy span {
  color: var(--tz-text-muted);
  font-size: 0.7rem;
}

.quickbuy-selected-product-editor-price {
  flex: 0 0 auto;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
}

.quickbuy-selected-product-editor-facts {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.5rem;
}

.quickbuy-selected-product-editor-facts > span {
  display: grid;
  min-width: 0;
  gap: 0.2rem;
  padding: 0.65rem;
  border-radius: 0.6rem;
  background: var(--tz-surface-subtle);
}

.quickbuy-selected-product-editor-facts small {
  overflow: hidden;
  color: var(--tz-text-muted);
  font-size: 0.62rem;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.quickbuy-selected-product-editor-facts strong {
  overflow: hidden;
  color: var(--tz-text-primary);
  font-size: 0.78rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quickbuy-selected-product-editor-quantity {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding-top: 0.25rem;
  color: var(--tz-text-primary);
  font-size: 0.85rem;
  font-weight: 700;
}

.quickbuy-selected-product-editor-stepper {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.quickbuy-selected-product-editor-stepper button,
.quickbuy-selected-product-editor-stepper input {
  width: 2.15rem;
  height: 2.15rem;
  box-sizing: border-box;
  border: 0;
  border-radius: 0.45rem;
  color: var(--tz-text-primary);
  background: var(--tz-input-surface);
  box-shadow: inset 0 0 0 1px var(--tz-border-subtle);
}

.quickbuy-selected-product-editor-stepper button {
  display: inline-grid;
  place-items: center;
  transition: background-color 160ms ease, opacity 160ms ease;
}

.quickbuy-selected-product-editor-stepper button:hover:not(:disabled) {
  background: var(--tz-surface-muted);
}

.quickbuy-selected-product-editor-stepper button:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

.quickbuy-selected-product-editor-stepper input {
  padding: 0 0.25rem;
  background: var(--tz-input-surface);
  text-align: center;
}

.quickbuy-selected-product-editor-footer {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 0.5rem;
  padding: 0.85rem 1rem 1rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.quickbuy-selected-product-editor-action {
  display: inline-flex;
  min-width: 0;
  min-height: 2.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.5rem 0.75rem;
  border: 0;
  border-radius: 0.5rem;
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
  font-size: 0.75rem;
  font-weight: 700;
}

.quickbuy-selected-product-editor-action--danger {
  color: #fecaca;
}

.quickbuy-selected-product-editor-action--primary {
  color: var(--tz-action-primary-foreground);
  background: var(--tz-action-primary);
  border: 1px solid var(--tz-action-primary);
}

.quickbuy-selected-product-editor-action--primary:hover:not(:disabled) {
  background: var(--tz-action-primary-hover);
  border-color: var(--tz-action-primary-hover);
}

@media (max-width: 767px) {
  .quickbuy-selected-product-editor-mask {
    align-items: flex-end;
    padding: 0;
  }

  .quickbuy-selected-product-editor-shell {
    width: 100%;
    max-height: min(48vh, 25rem);
    border-radius: 1rem 1rem 0 0;
  }

  .quickbuy-selected-product-editor-content {
    padding: 0.85rem;
  }

  .quickbuy-selected-product-editor-header {
    padding: 0.75rem 0.85rem;
  }

  .quickbuy-selected-product-editor-footer {
    padding: 0.7rem 0.85rem max(0.7rem, calc(0.7rem + var(--tz-safe-area-bottom, 0px)));
  }

  .quickbuy-selected-product-editor-image {
    width: 3rem;
    height: 3rem;
  }

  .quickbuy-selected-product-editor-facts {
    gap: 0.375rem;
  }

  .quickbuy-selected-product-editor-facts > span {
    padding: 0.55rem 0.45rem;
  }
}
</style>
