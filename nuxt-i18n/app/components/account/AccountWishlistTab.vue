<template>
  <section class="account-tab-panel">
    <div class="tab-head">
      <div>
        <p>{{ t('accountSidebar.wishlist.eyebrow', 'Saved products') }}</p>
        <h3>{{ t('accountSidebar.wishlist.title', 'Wishlist') }}</h3>
      </div>
      <button
        type="button"
        class="tab-head__action"
        :disabled="loading || !items.length"
        @click="addAllVisibleToCart"
      >
        {{ t('accountSidebar.wishlist.addAll', 'Add all') }}
      </button>
    </div>

    <div v-if="loading" class="account-loading">
      <Icon name="lucide:loader-circle" />
      {{ t('wishlistDrawer.loading', 'Loading wishlist...') }}
    </div>

    <div v-else-if="error" class="account-message account-message--error">
      {{ error }}
    </div>

    <div v-else-if="!items.length" class="account-empty">
      <Icon name="lucide:heart" />
      <strong>{{ t('wishlistDrawer.empty.title', 'No saved products yet') }}</strong>
      <span>{{ t('wishlistDrawer.empty.description', 'Save products from the shop and they will appear here.') }}</span>
      <NuxtLink :to="localePath('/shop')" @click="$emit('close')">
        {{ t('accountSidebar.actions.goShop', 'Go to shop') }}
      </NuxtLink>
    </div>

    <div v-else class="account-item-list">
      <article v-for="item in items" :key="item.id" class="account-product-card">
        <NuxtLink :to="productPath(item)" class="account-product-card__image" @click="$emit('close')">
          <img v-if="productImage(item)" :src="productImage(item)" :alt="productTitle(item)" loading="lazy" />
          <Icon v-else name="lucide:image" />
        </NuxtLink>
        <div class="account-product-card__body">
          <NuxtLink :to="productPath(item)" class="account-product-card__title" @click="$emit('close')">
            {{ productTitle(item) }}
          </NuxtLink>
          <span v-if="displayPrice(item)" class="account-product-card__price">{{ displayPrice(item) }}</span>
          <div class="account-product-card__actions">
            <button type="button" @click="addWishlistItemToCart(item)">
              {{ t('accountSidebar.wishlist.addToCart', 'Add to cart') }}
            </button>
            <button type="button" class="account-product-card__ghost" @click="removeFromWishlist(item.id)">
              {{ t('wishlistDrawer.actions.remove', 'Remove') }}
            </button>
          </div>
        </div>
      </article>
    </div>

    <p v-if="feedback" class="account-toast">{{ feedback }}</p>
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { useWishlist, type WishlistItem } from '~/composables/useWishlist'

const props = defineProps<{
  active: boolean
}>()

defineEmits<{
  (event: 'close'): void
}>()

interface ProductLike {
  id?: number | string
  product_id?: number | string
  variant_id?: number | string | null
  title?: string
  name?: string
  slug?: string
  sku?: string
  thumbnail?: string
  featured_image?: string
  image?: string
  price?: number | string
  sale_price?: number | string | null
  regular_price?: number | string
  prices?: {
    sale?: number | string
    regular?: number | string
  }
}

const { t } = useI18n()
const localePath = useLocalePath()
const auth = useAuth()
const { items, loading, error, loadWishlist, removeFromWishlist } = useWishlist()
const { addToCart, formatPrice } = useCart()

const feedback = ref('')

const productOf = (item: WishlistItem): ProductLike => {
  return item.product && typeof item.product === 'object' ? item.product as ProductLike : {}
}

const toNumber = (value: unknown): number | null => {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : null
}

const productId = (item: WishlistItem): number | null => {
  const product = productOf(item)
  return toNumber(product.id ?? product.product_id ?? item.product_id)
}

const variantId = (item: WishlistItem): number | null => {
  const product = productOf(item)
  return toNumber(product.variant_id)
}

const productTitle = (item: WishlistItem) => {
  const product = productOf(item)
  return product.title || product.name || t('wishlistDrawer.productFallback', 'Product')
}

const productImage = (item: WishlistItem) => {
  const product = productOf(item)
  return product.thumbnail || product.featured_image || product.image || ''
}

const productSlug = (item: WishlistItem) => {
  const product = productOf(item)
  return product.slug || String(productId(item) || '')
}

const productPath = (item: WishlistItem) => {
  const slug = productSlug(item)
  return localePath(slug ? `/shop/${slug}` : '/shop')
}

const rawPrice = (item: WishlistItem) => {
  const product = productOf(item)
  return toNumber(product.prices?.sale ?? product.sale_price ?? product.price ?? product.prices?.regular ?? product.regular_price)
}

const displayPrice = (item: WishlistItem) => {
  const price = rawPrice(item)
  return price !== null ? formatPrice(price) : ''
}

const flash = (message: string) => {
  feedback.value = message
  if (typeof window === 'undefined') return
  window.setTimeout(() => {
    feedback.value = ''
  }, 2200)
}

const addWishlistItemToCart = (item: WishlistItem) => {
  const id = productId(item)
  if (!id) {
    flash(t('accountSidebar.wishlist.missingProduct', 'This saved product needs product data before adding to cart.'))
    return
  }

  const product = productOf(item)
  const price = rawPrice(item) ?? 0
  const result = addToCart({
    id,
    product_id: id,
    variant_id: variantId(item),
    title: productTitle(item),
    name: productTitle(item),
    slug: product.slug || '',
    sku: product.sku || '',
    price,
    sale_price: toNumber(product.sale_price),
    thumbnail: productImage(item),
    image: productImage(item),
  })

  flash(result.message || t('accountSidebar.wishlist.added', 'Added to cart'))
}

const addAllVisibleToCart = () => {
  let added = 0
  items.value.forEach((item) => {
    if (productId(item)) {
      addWishlistItemToCart(item)
      added += 1
    }
  })
  if (added > 0) {
    flash(t('accountSidebar.wishlist.addedAll', 'Wishlist products added to cart'))
  }
}

watch(
  [() => props.active, () => auth.isAuthenticated.value],
  ([active, logged]) => {
    if (active && logged) {
      loadWishlist()
    }
  },
  { immediate: true },
)
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
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.tab-head p,
.tab-head h3 {
  margin: 0;
}

.tab-head p {
  color: rgba(103, 232, 249, 0.9);
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

.tab-head__action {
  min-height: 2.1rem;
  flex: 0 0 auto;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  color: #ffffff;
  padding: 0 0.85rem;
  font-size: var(--tz-type-micro-label);
  font-weight: 800;
}

.tab-head__action:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.account-loading,
.account-message,
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
  color: rgba(226, 232, 240, 0.78);
  font-size: 0.82rem;
}

.account-loading :deep(svg) {
  width: 1.4rem;
  height: 1.4rem;
  animation: wishlist-spin 0.85s linear infinite;
}

.account-message--error {
  color: #fca5a5;
  font-size: 0.82rem;
}

.account-empty :deep(svg) {
  width: 2.2rem;
  height: 2.2rem;
  color: rgba(64, 255, 170, 0.72);
}

.account-empty strong {
  color: #ffffff;
  font-size: 0.95rem;
}

.account-empty a {
  margin-top: 0.3rem;
  border-radius: 999px;
  background: linear-gradient(135deg, #4efce7, #60a5fa);
  color: #020617;
  padding: 0.65rem 1rem;
  font-size: 0.8rem;
  font-weight: 850;
  text-decoration: none;
}

.account-item-list {
  display: flex;
  min-height: 0;
  max-height: min(54vh, 34rem);
  flex-direction: column;
  gap: 0.7rem;
  overflow-y: auto;
  padding-right: 0.25rem;
}

.account-product-card {
  display: grid;
  grid-template-columns: 4.4rem 1fr;
  gap: 0.78rem;
  border-radius: 1.15rem;
  background: rgba(255, 255, 255, 0.055);
  padding: 0.75rem;
}

.account-product-card__image {
  display: flex;
  width: 4.4rem;
  height: 4.4rem;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 0.95rem;
  background: rgba(255, 255, 255, 0.07);
  color: rgba(226, 232, 240, 0.72);
}

.account-product-card__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.account-product-card__image :deep(svg) {
  width: 1.5rem;
  height: 1.5rem;
}

.account-product-card__body {
  min-width: 0;
}

.account-product-card__title {
  display: block;
  overflow: hidden;
  color: #ffffff;
  font-size: 0.84rem;
  font-weight: 800;
  line-height: 1.35;
  text-decoration: none;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-product-card__price {
  display: block;
  margin-top: 0.2rem;
  color: #40ffaa;
  font-size: 0.79rem;
  font-weight: 850;
}

.account-product-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  margin-top: 0.65rem;
}

.account-product-card__actions button {
  min-height: 1.9rem;
  border-radius: 999px;
  background: rgba(64, 255, 170, 0.16);
  color: #ffffff;
  padding: 0 0.72rem;
  font-size: var(--tz-type-micro-label);
  font-weight: 800;
}

.account-product-card__actions .account-product-card__ghost {
  background: rgba(255, 255, 255, 0.08);
  color: rgba(226, 232, 240, 0.82);
}

.account-toast {
  margin: 0;
  border-radius: 999px;
  background: rgba(64, 255, 170, 0.12);
  color: #bfffe3;
  padding: 0.6rem 0.8rem;
  text-align: center;
  font-size: 0.75rem;
  font-weight: 800;
}

@keyframes wishlist-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>

