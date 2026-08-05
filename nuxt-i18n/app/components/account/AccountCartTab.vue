<template>
  <section class="account-tab-panel">
    <div class="tab-head">
      <div>
        <p>{{ t('accountSidebar.cart.eyebrow', 'Ready to pay') }}</p>
        <h3>{{ t('accountSidebar.cart.title', 'Cart') }} · {{ cartCount }}</h3>
      </div>
      <strong>{{ formatPrice(total) }}</strong>
    </div>

    <div v-if="isLoadingCart" class="account-loading">
      <Icon name="lucide:loader-circle" />
      {{ t('cartDrawer.loading', 'Loading cart...') }}
    </div>

    <div v-else-if="!cartItems.length" class="account-empty">
      <Icon name="lucide:shopping-bag" />
      <strong>{{ t('cartDrawer.empty.title', 'Your cart is empty') }}</strong>
      <span>{{ t('cartDrawer.empty.description', 'Add products to cart and check out from here.') }}</span>
      <NuxtLink :to="localePath('/shop')" @click="$emit('close')">
        {{ t('accountSidebar.actions.goShop', 'Go to shop') }}
      </NuxtLink>
    </div>

    <template v-else>
      <div class="account-cart-list">
        <article v-for="item in cartItems" :key="item.id" class="account-cart-item">
          <div class="account-cart-item__image">
            <img v-if="cartImage(item)" :src="cartImage(item)" :alt="item.title" loading="lazy" />
            <Icon v-else name="lucide:image" />
          </div>
          <div class="account-cart-item__body">
            <div class="account-cart-item__title">{{ item.title }}</div>
            <div class="account-cart-item__meta">
              <span>{{ formatPrice(item.price) }}</span>
            </div>
            <div class="account-cart-item__actions">
              <button type="button" @click="decrementQuantity(item.id)">
                <Icon name="lucide:minus" />
              </button>
              <span>{{ item.quantity }}</span>
              <button type="button" @click="incrementQuantity(item.id)">
                <Icon name="lucide:plus" />
              </button>
              <button type="button" class="account-cart-item__remove" @click="removeFromCart(item.id)">
                {{ t('cartDrawer.actions.remove', 'Remove') }}
              </button>
            </div>
          </div>
        </article>
      </div>

      <div class="account-cart-summary">
        <div>
          <span>{{ t('cartDrawer.summary.subtotal', 'Subtotal') }}</span>
          <strong>{{ formatPrice(subtotal) }}</strong>
        </div>
        <div>
          <span>{{ t('cartDrawer.summary.tax', 'Tax') }}</span>
          <strong>{{ formatPrice(tax) }}</strong>
        </div>
        <div class="account-cart-summary__total">
          <span>{{ t('cartDrawer.summary.estimatedTotal', 'Estimated total') }}</span>
          <strong>{{ formatPrice(total) }}</strong>
        </div>
      </div>

    </template>
  </section>
</template>

<script setup lang="ts">
import { useI18n, useLocalePath } from '#imports'
import type { CartItem } from '~~/types/cart'
import { useCart } from '~/composables/useCart'

const { t } = useI18n()
const localePath = useLocalePath()

const {
  cartItems,
  cartCount,
  subtotal,
  tax,
  total,
  isLoadingCart,
  incrementQuantity,
  decrementQuantity,
  removeFromCart,
  formatPrice,
} = useCart()

const emit = defineEmits<{
  (event: 'close'): void
}>()

const cartImage = (item: CartItem) => item.thumbnail || item.image || ''

</script>

<style scoped>
.account-tab-panel {
  display: flex;
  min-height: 0;
  flex-direction: column;
  gap: 0.85rem;
}

.tab-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.tab-head p,
.tab-head h3 {
  margin: 0;
}

.tab-head p {
  color: rgba(181, 255, 109, 0.9);
  font-size: var(--tz-type-micro-label);
  font-weight: 800;
  letter-spacing: 0.15em;
  text-transform: uppercase;
}

.tab-head h3 {
  margin-top: 0.18rem;
  color: #ffffff;
  font-size: 1.05rem;
  font-weight: 850;
}

.tab-head > strong {
  color: #B5FF6D;
  font-size: 1rem;
  font-weight: 900;
  white-space: nowrap;
}

.account-loading,
.account-empty {
  display: flex;
  min-height: 12rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.55rem;
  border-radius: 1.25rem;
  background: rgba(255, 255, 255, 0.05);
  padding: 1rem;
  text-align: center;
}

.account-loading,
.account-empty span {
  color: rgba(232, 232, 232, 0.78);
  font-size: 0.82rem;
}

.account-loading :deep(svg) {
  width: 1.4rem;
  height: 1.4rem;
  animation: cart-spin 0.85s linear infinite;
}

.account-empty :deep(svg) {
  width: 2.2rem;
  height: 2.2rem;
  color: rgba(181, 255, 109, 0.72);
}

.account-empty strong {
  color: #ffffff;
  font-size: 0.95rem;
}

.account-empty a {
  margin-top: 0.3rem;
  border-radius: 999px;
  background: #B5FF6D;
  color: #050505;
  padding: 0.65rem 1rem;
  font-size: 0.8rem;
  font-weight: 850;
  text-decoration: none;
}

.account-cart-list {
  display: flex;
  min-height: 0;
  max-height: min(48vh, 28rem);
  flex-direction: column;
  gap: 0.65rem;
  overflow-y: auto;
  padding-right: 0.25rem;
}

.account-cart-item {
  display: grid;
  grid-template-columns: 4.2rem 1fr;
  gap: 0.75rem;
  border-radius: 1.15rem;
  background: rgba(255, 255, 255, 0.055);
  padding: 0.72rem;
}

.account-cart-item__image {
  display: flex;
  width: 4.2rem;
  height: 4.2rem;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 0.95rem;
  background: rgba(255, 255, 255, 0.07);
  color: rgba(232, 232, 232, 0.72);
}

.account-cart-item__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.account-cart-item__image :deep(svg) {
  width: 1.45rem;
  height: 1.45rem;
}

.account-cart-item__body {
  min-width: 0;
}

.account-cart-item__title {
  overflow: hidden;
  color: #ffffff;
  font-size: 0.84rem;
  font-weight: 800;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-cart-item__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  margin-top: 0.25rem;
  color: rgba(232, 232, 232, 0.72);
  font-size: var(--tz-type-micro-label);
}

.account-cart-item__meta span:first-child {
  color: #B5FF6D;
  font-weight: 850;
}

.account-cart-item__actions {
  display: flex;
  align-items: center;
  gap: 0.42rem;
  margin-top: 0.6rem;
}

.account-cart-item__actions button {
  display: inline-flex;
  min-width: 1.82rem;
  height: 1.82rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  color: #ffffff;
  font-size: var(--tz-type-micro-label);
  font-weight: 800;
}

.account-cart-item__actions svg {
  width: 0.9rem;
  height: 0.9rem;
}

.account-cart-item__actions span {
  min-width: 1.5rem;
  color: #ffffff;
  text-align: center;
  font-size: 0.8rem;
  font-weight: 850;
}

.account-cart-item__remove {
  margin-left: auto;
  padding-inline: 0.7rem;
  color: #fecaca !important;
}

.account-cart-summary {
  display: grid;
  gap: 0.55rem;
  border-radius: 1.15rem;
  background: rgba(255, 255, 255, 0.055);
  padding: 0.85rem;
}

.account-cart-summary div {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  color: rgba(232, 232, 232, 0.78);
  font-size: 0.78rem;
}

.account-cart-summary strong {
  color: #ffffff;
  font-weight: 850;
}

.account-cart-summary__total {
  border-top: 1px solid rgba(255, 255, 255, 0.11);
  padding-top: 0.55rem;
}

.account-cart-summary__total span,
.account-cart-summary__total strong {
  color: #ffffff;
}

.account-cart-checkout {
  min-height: 2.7rem;
  border-radius: 999px;
  background: #B5FF6D;
  color: #050505;
  font-size: 0.86rem;
  font-weight: 900;
}

@keyframes cart-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>

