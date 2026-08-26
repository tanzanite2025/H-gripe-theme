<template>
  <div class="product-quantity-control">
    <label for="product-quantity-input">{{ quantityLabel }}</label>
    <div class="product-quantity-stepper" role="group" :aria-label="quantityLabel">
      <button
        type="button"
        class="product-quantity-button"
        :disabled="quantity <= 1"
        :aria-label="decreaseQuantityLabel"
        @click="emit('decrease')"
      >
        <Icon name="lucide:minus" aria-hidden="true" />
      </button>
      <input
        id="product-quantity-input"
        :value="quantity"
        class="product-quantity-input"
        type="number"
        inputmode="numeric"
        min="1"
        :max="maxQuantity"
        @input="handleQuantityInput"
      />
      <button
        type="button"
        class="product-quantity-button"
        :disabled="quantity >= maxQuantity"
        :aria-label="increaseQuantityLabel"
        @click="emit('increase')"
      >
        <Icon name="lucide:plus" aria-hidden="true" />
      </button>
    </div>
  </div>
  <div class="product-actions" aria-label="Product actions">
    <button
      type="button"
      class="product-add-button"
      :disabled="!canAddToCart"
      @click="emit('add-to-cart')"
    >
      {{ canAddToCart ? addToCartLabel : outOfStockLabel }}
    </button>
    <button
      type="button"
      class="product-buy-now-button"
      :disabled="!canBuyNow"
      @click="emit('buy-now')"
    >
      {{ canBuyNow ? buyNowLabel : buyNowUnavailableLabel }}
    </button>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  quantity: number
  maxQuantity: number
  canAddToCart: boolean
  canBuyNow: boolean
  addToCartLabel?: string
  outOfStockLabel?: string
  buyNowLabel?: string
  buyNowUnavailableLabel: string
  quantityLabel?: string
  decreaseQuantityLabel?: string
  increaseQuantityLabel?: string
}>(), {
  addToCartLabel: 'Add to cart',
  outOfStockLabel: 'Out of stock',
  buyNowLabel: 'Buy now',
  quantityLabel: 'Quantity',
  decreaseQuantityLabel: 'Decrease quantity',
  increaseQuantityLabel: 'Increase quantity',
})

const emit = defineEmits<{
  (event: 'decrease'): void
  (event: 'increase'): void
  (event: 'input', value: string): void
  (event: 'add-to-cart'): void
  (event: 'buy-now'): void
}>()

const handleQuantityInput = (event: Event) => {
  emit('input', (event.target as HTMLInputElement).value)
}
</script>

<style scoped>
.product-quantity-control {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.product-quantity-control label {
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
  font-weight: 700;
  text-transform: uppercase;
}

.product-quantity-stepper {
  display: inline-grid;
  grid-template-columns: var(--product-control-pill-height) 3.5rem var(--product-control-pill-height);
  height: var(--product-control-pill-height);
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: var(--product-control-pill-radius);
  background: var(--tz-surface-subtle);
}

.product-quantity-button,
.product-quantity-input {
  display: inline-flex;
  height: 100%;
  min-width: 0;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  border: 0;
  background: transparent;
  color: var(--tz-text-primary);
  font: inherit;
  font-size: 0.86rem;
  font-weight: 800;
  line-height: 1;
}

.product-quantity-button {
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.product-quantity-button:hover:not(:disabled) {
  background: rgba(5, 150, 105, 0.14);
  color: #059669;
}

.product-quantity-button:disabled {
  cursor: not-allowed;
  color: var(--tz-text-disabled);
}

.product-quantity-button svg {
  width: 0.9rem;
  height: 0.9rem;
}

.product-quantity-input {
  width: 100%;
  border-inline: 1px solid var(--tz-border-subtle);
  color-scheme: light;
  text-align: center;
  -moz-appearance: textfield;
}

.product-quantity-input::-webkit-inner-spin-button,
.product-quantity-input::-webkit-outer-spin-button {
  margin: 0;
  -webkit-appearance: none;
}

.product-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.55rem;
}

.product-add-button,
.product-buy-now-button {
  display: inline-flex;
  height: var(--product-control-pill-height);
  min-height: 0;
  align-items: center;
  justify-content: center;
  width: fit-content;
  box-sizing: border-box;
  border: 0;
  border-radius: var(--product-control-pill-radius);
  cursor: pointer;
  font-weight: 800;
  line-height: 1;
  padding: 0 1.05rem;
  transition: background-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
}

.product-add-button {
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
}

.product-buy-now-button {
  border: 1px solid var(--tz-border-strong);
  background: var(--tz-surface-inset);
  color: var(--tz-text-primary);
}

.product-add-button:hover:not(:disabled) {
  background: var(--tz-action-primary-hover);
  box-shadow: 0 0 0 3px rgb(15 23 42 / 0.12);
  transform: translateY(-1px);
}

.product-buy-now-button:hover:not(:disabled) {
  border-color: var(--tz-action-primary);
  background: var(--tz-surface-card);
  transform: translateY(-1px);
}

.product-add-button:active:not(:disabled),
.product-buy-now-button:active:not(:disabled) {
  transform: translateY(1px);
}

.product-quantity-button:focus-visible,
.product-quantity-input:focus-visible,
.product-add-button:focus-visible,
.product-buy-now-button:focus-visible {
  outline: 2px solid var(--tz-site-accent);
  outline-offset: 3px;
}

.product-add-button:disabled,
.product-buy-now-button:disabled {
  border: 1px solid var(--tz-border-subtle);
  background: var(--tz-surface-muted);
  color: var(--tz-text-secondary);
  cursor: not-allowed;
  opacity: 1;
}
</style>
