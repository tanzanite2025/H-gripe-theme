<template>
  <Teleport to="body">
    <Transition name="global-product-detail-bottom-sheet-fade">
      <div
        v-if="isGlobalProductDetailBottomSheetOpen"
        class="global-product-detail-bottom-sheet-mask tz-mobile-dialog-mask"
        @click.self="closeGlobalProductDetailBottomSheet"
      >
        <div
          class="global-product-detail-bottom-sheet-backdrop"
          aria-hidden="true"
          @click="closeGlobalProductDetailBottomSheet"
        ></div>

        <Transition name="global-product-detail-bottom-sheet-slide" appear>
          <section
            v-if="isGlobalProductDetailBottomSheetOpen"
            class="global-product-detail-bottom-sheet-shell tz-mobile-dialog-surface"
            role="dialog"
            aria-modal="true"
            :aria-label="globalProductDetailBottomSheetAriaLabel"
          >
            <header class="global-product-detail-bottom-sheet-header">
              <div class="global-product-detail-bottom-sheet-header-copy">
                <span class="global-product-detail-bottom-sheet-eyebrow">
                  {{ t('products.detail.viewDetails', 'Product details') }}
                </span>
                <h2 class="global-product-detail-bottom-sheet-title">
                  {{ product?.title || productReference?.title || t('common.loading', 'Loading...') }}
                </h2>
              </div>
              <button
                type="button"
                class="global-product-detail-bottom-sheet-close"
                :aria-label="t('common.close', 'Close')"
                :title="t('common.close', 'Close')"
                @click="closeGlobalProductDetailBottomSheet"
              >
                <Icon name="lucide:x" class="h-5 w-5" aria-hidden="true" />
              </button>
            </header>

            <div class="global-product-detail-bottom-sheet-content">
              <div v-if="isLoadingProduct" class="global-product-detail-bottom-sheet-state">
                <Icon name="lucide:loader-circle" class="h-7 w-7 animate-spin" aria-hidden="true" />
                <span>{{ t('common.loading', 'Loading...') }}</span>
              </div>

              <div v-else-if="productLoadError" class="global-product-detail-bottom-sheet-state global-product-detail-bottom-sheet-state--error" role="alert">
                <Icon name="lucide:triangle-alert" class="h-7 w-7" aria-hidden="true" />
                <span>{{ productLoadError }}</span>
                <button
                  type="button"
                  class="global-product-detail-bottom-sheet-inline-action"
                  @click="loadGlobalProductDetailBottomSheetProduct"
                >
                  <Icon name="lucide:refresh-cw" class="h-4 w-4" aria-hidden="true" />
                  <span>{{ t('common.retry', 'Retry') }}</span>
                </button>
              </div>

              <div v-else-if="product" class="global-product-detail-bottom-sheet-layout">
                <section class="global-product-detail-bottom-sheet-media-section" :aria-label="t('products.detail.media', 'Product media')">
                  <div class="global-product-detail-bottom-sheet-media-stage">
                    <StorefrontImage
                      v-if="selectedMedia?.kind === 'image'"
                      :src="selectedMedia.url"
                      :alt="selectedMedia.alt"
                      class="global-product-detail-bottom-sheet-media-image"
                      preset="detail"
                      loading="eager"
                      fetchpriority="high"
                    />
                    <video
                      v-else-if="selectedMedia?.kind === 'video'"
                      :src="selectedMedia.url"
                      :poster="selectedMedia.poster || undefined"
                      controls
                      playsinline
                      preload="metadata"
                      class="global-product-detail-bottom-sheet-media-video"
                    />
                    <div v-else class="global-product-detail-bottom-sheet-media-empty">
                      <Icon name="lucide:image-off" class="h-10 w-10" aria-hidden="true" />
                    </div>

                    <template v-if="productMediaItems.length > 1">
                      <button
                        type="button"
                        class="tz-directional-arrow global-product-detail-bottom-sheet-media-nav global-product-detail-bottom-sheet-media-nav--previous"
                        :aria-label="t('common.previous', 'Previous')"
                        :title="t('common.previous', 'Previous')"
                        @click="selectPreviousGlobalProductDetailBottomSheetMedia"
                      >
                        <Icon name="lucide:chevron-left" aria-hidden="true" />
                      </button>
                      <button
                        type="button"
                        class="tz-directional-arrow global-product-detail-bottom-sheet-media-nav global-product-detail-bottom-sheet-media-nav--next"
                        :aria-label="t('common.next', 'Next')"
                        :title="t('common.next', 'Next')"
                        @click="selectNextGlobalProductDetailBottomSheetMedia"
                      >
                        <Icon name="lucide:chevron-right" aria-hidden="true" />
                      </button>
                    </template>
                  </div>

                  <div v-if="productMediaItems.length > 1" class="global-product-detail-bottom-sheet-media-thumbnails">
                    <button
                      v-for="media in productMediaItems"
                      :key="media.id"
                      type="button"
                      class="global-product-detail-bottom-sheet-media-thumbnail"
                      :class="{ 'global-product-detail-bottom-sheet-media-thumbnail--active': media.id === selectedMediaId }"
                      :aria-label="`${media.kind === 'video' ? t('products.detail.viewVideo', 'View product video') : t('products.detail.viewImage', 'View product image')} ${media.index + 1}`"
                      :aria-pressed="media.id === selectedMediaId"
                      @click="selectGlobalProductDetailBottomSheetMedia(media.id)"
                    >
                      <StorefrontImage
                        v-if="media.thumbnailUrl"
                        :src="media.thumbnailUrl"
                        :alt="media.alt"
                        preset="thumbnail"
                      />
                      <span v-else class="global-product-detail-bottom-sheet-media-thumbnail-placeholder">
                        <Icon :name="media.kind === 'video' ? 'lucide:video' : 'lucide:image'" class="h-4 w-4" aria-hidden="true" />
                      </span>
                      <span v-if="media.kind === 'video'" class="global-product-detail-bottom-sheet-media-thumbnail-badge" aria-hidden="true">
                        <Icon name="lucide:play" class="h-3 w-3" />
                      </span>
                    </button>
                  </div>
                </section>

                <section class="global-product-detail-bottom-sheet-information">
                  <div class="global-product-detail-bottom-sheet-product-heading">
                    <div>
                      <h3 class="global-product-detail-bottom-sheet-product-title">{{ product.title }}</h3>
                      <p v-if="rawProduct?.product_specification_template?.name || product.productSpecificationTemplate?.name" class="global-product-detail-bottom-sheet-product-specification-template">
                        {{ rawProduct?.product_specification_template?.name || product.productSpecificationTemplate?.name }}
                      </p>
                    </div>
                    <strong class="global-product-detail-bottom-sheet-price">{{ formattedSelectedProductPrice }}</strong>
                  </div>

                  <div
                    v-if="rawProduct?.short_description || rawProduct?.description || product.description"
                    class="global-product-detail-bottom-sheet-description"
                  >
                    <SafeRichText
                      v-if="rawProduct?.short_description"
                      :html="rawProduct.short_description"
                    />
                    <SafeRichText
                      v-else
                      :html="rawProduct?.description || product.description"
                    />
                  </div>

                  <div v-if="variantOptionGroups.length" class="global-product-detail-bottom-sheet-variant-groups">
                    <fieldset
                      v-for="group in variantOptionGroups"
                      :key="group.slug"
                      class="global-product-detail-bottom-sheet-variant-group"
                    >
                      <legend>{{ group.name }}</legend>
                      <div class="global-product-detail-bottom-sheet-variant-options">
                        <button
                          v-for="option in group.options"
                          :key="`${group.slug}-${option.value}`"
                          type="button"
                          class="global-product-detail-bottom-sheet-variant-option"
                          :class="{
                            'global-product-detail-bottom-sheet-variant-option--selected': option.selected,
                            'global-product-detail-bottom-sheet-variant-option--unavailable': !option.available,
                          }"
                          :disabled="!option.available"
                          :aria-pressed="option.selected"
                          @click="selectGlobalProductDetailBottomSheetVariantOption(group.slug, option.value)"
                        >
                          {{ option.label }}
                        </button>
                      </div>
                    </fieldset>
                  </div>

                  <div v-else-if="activeProductVariants.length > 1" class="global-product-detail-bottom-sheet-variant-select">
                    <label for="global-product-detail-bottom-sheet-variant-select">
                      {{ t('products.detail.option', 'Option') }}
                    </label>
                    <select
                      id="global-product-detail-bottom-sheet-variant-select"
                      v-model.number="selectedVariantId"
                    >
                      <option
                        v-for="variant in activeProductVariants"
                        :key="variant.id"
                        :value="variant.id"
                      >
                        {{ variant.title || variant.id }}
                      </option>
                    </select>
                  </div>

                  <div class="global-product-detail-bottom-sheet-facts">
                    <span v-if="selectedVariantWeight" class="global-product-detail-bottom-sheet-fact">
                      <Icon name="lucide:scale" class="h-4 w-4" aria-hidden="true" />
                      <span>{{ selectedVariantWeight }}g</span>
                    </span>
                    <span v-if="selectedVariant?.sku || rawProduct?.sku || product.sku" class="global-product-detail-bottom-sheet-fact">
                      <Icon name="lucide:barcode" class="h-4 w-4" aria-hidden="true" />
                      <span>{{ selectedVariant?.sku || rawProduct?.sku || product.sku }}</span>
                    </span>
                    <span
                      class="global-product-detail-bottom-sheet-fact"
                      :class="{ 'global-product-detail-bottom-sheet-fact--unavailable': selectedAvailability !== 'in_stock' }"
                    >
                      <Icon name="lucide:circle-check" class="h-4 w-4" aria-hidden="true" />
                      <span>{{ selectedAvailability === 'in_stock' ? t('products.detail.inStock', 'In stock') : t('products.detail.outOfStock', 'Out of stock') }}</span>
                    </span>
                  </div>

                  <section v-if="productSpecificationGroups.length" class="global-product-detail-bottom-sheet-specifications">
                    <h4>{{ t('products.detail.specifications', 'Specifications') }}</h4>
                    <div
                      v-for="group in productSpecificationGroups"
                      :key="group.name"
                      class="global-product-detail-bottom-sheet-specification-group"
                    >
                      <h5 v-if="productSpecificationGroups.length > 1">{{ group.name }}</h5>
                      <dl>
                        <div v-for="item in group.items" :key="item.slug" class="global-product-detail-bottom-sheet-specification-item">
                          <dt>{{ item.name }}</dt>
                          <dd>{{ item.value }}</dd>
                        </div>
                      </dl>
                    </div>
                  </section>

                  <div class="global-product-detail-bottom-sheet-purchase">
                    <div class="global-product-detail-bottom-sheet-quantity-control">
                      <span>{{ t('products.detail.quantity', 'Quantity') }}</span>
                      <div class="global-product-detail-bottom-sheet-quantity-stepper" role="group" :aria-label="t('products.detail.quantity', 'Quantity')">
                        <button
                          type="button"
                          :disabled="selectedQuantity <= 1"
                          :aria-label="t('products.detail.decreaseQuantity', 'Decrease quantity')"
                          :title="t('products.detail.decreaseQuantity', 'Decrease quantity')"
                          @click="decreaseGlobalProductDetailBottomSheetQuantity"
                        >
                          <Icon name="lucide:minus" class="h-4 w-4" aria-hidden="true" />
                        </button>
                        <input
                          v-model.number="selectedQuantity"
                          type="number"
                          min="1"
                          max="99"
                          inputmode="numeric"
                          :aria-label="t('products.detail.quantity', 'Quantity')"
                          @change="normalizeGlobalProductDetailBottomSheetQuantity"
                        />
                        <button
                          type="button"
                          :disabled="selectedQuantity >= 99"
                          :aria-label="t('products.detail.increaseQuantity', 'Increase quantity')"
                          :title="t('products.detail.increaseQuantity', 'Increase quantity')"
                          @click="increaseGlobalProductDetailBottomSheetQuantity"
                        >
                          <Icon name="lucide:plus" class="h-4 w-4" aria-hidden="true" />
                        </button>
                      </div>
                    </div>

                    <div class="global-product-detail-bottom-sheet-purchase-actions">
                      <button
                        type="button"
                        class="global-product-detail-bottom-sheet-purchase-action global-product-detail-bottom-sheet-purchase-action--secondary"
                        :disabled="!canAddSelectedProductToCart"
                        @click="addGlobalProductDetailToCart"
                      >
                        <Icon name="lucide:shopping-cart" class="h-4 w-4" aria-hidden="true" />
                        <span>{{ t('quickBuy.actions.addToCart', 'Add to cart') }}</span>
                      </button>
                      <button
                        type="button"
                        class="global-product-detail-bottom-sheet-purchase-action global-product-detail-bottom-sheet-purchase-action--primary"
                        :disabled="!canBuySelectedProductNow"
                        @click="buyGlobalProductDetailNow"
                      >
                        <Icon name="lucide:credit-card" class="h-4 w-4" aria-hidden="true" />
                        <span>{{ t('checkout.product.buyNow', 'Buy now') }}</span>
                      </button>
                    </div>
                  </div>

                  <p v-if="purchaseFeedbackMessage" class="global-product-detail-bottom-sheet-feedback" role="status">
                    {{ purchaseFeedbackMessage }}
                  </p>
                </section>
              </div>
            </div>
          </section>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from '#imports'
import {
  useGlobalProductDetailBottomSheet,
} from '~/composables/useGlobalProductDetailBottomSheet'
import { useProductDetailLookup } from '~/composables/useProductDetailLookup'
import { useCart } from '~/composables/useCart'
import { useShopProducts, type ShopProduct, type ShopProductVariant } from '~/composables/useShopProducts'
import { useStorefrontContext } from '~/composables/useStorefrontContext'
import type { GoProduct } from '~/types/productDetail'

interface ProductMediaItem {
  id: string
  kind: 'image' | 'video'
  url: string
  thumbnailUrl: string
  poster?: string
  alt: string
  index: number
  isPrimary: boolean
}

interface ProductSpecificationGroup {
  name: string
  items: Array<{
    slug: string
    name: string
    value: string
  }>
}

interface VariantOptionGroup {
  slug: string
  name: string
  options: Array<{
    value: string
    label: string
    selected: boolean
    available: boolean
  }>
}

const { t, locale } = useI18n()
const { displayCurrency, countryCode } = useStorefrontContext()
const { addToCart, openCart, openCheckout } = useCart()
const { toCartItem } = useShopProducts()
const { fetchProductDetailSnapshot } = useProductDetailLookup()
const {
  isGlobalProductDetailBottomSheetOpen,
  globalProductDetailBottomSheetProductReference: productReference,
  globalProductDetailBottomSheetProductSlug: productSlug,
  closeGlobalProductDetailBottomSheet,
} = useGlobalProductDetailBottomSheet()

const rawProduct = ref<GoProduct | null>(null)
const product = ref<ShopProduct | null>(null)
const isLoadingProduct = ref(false)
const productLoadError = ref('')
const selectedMediaId = ref('')
const selectedVariantId = ref<number | null>(null)
const selectedQuantity = ref(1)
const purchaseFeedbackMessage = ref('')
let fetchRequestSequence = 0

const globalProductDetailBottomSheetAriaLabel = computed(() =>
  product.value?.title
    || productReference.value?.title
    || t('products.detail.viewDetails', 'Product details'),
)

const loadGlobalProductDetailBottomSheetProduct = async () => {
  const slug = productSlug.value
  if (!slug) return

  const requestSequence = fetchRequestSequence + 1
  fetchRequestSequence = requestSequence
  isLoadingProduct.value = true
  productLoadError.value = ''
  purchaseFeedbackMessage.value = ''

  try {
    const snapshot = await fetchProductDetailSnapshot(slug)
    if (requestSequence !== fetchRequestSequence) return
    if (!snapshot) {
      throw new Error(t('products.detail.notFound', 'Product not found'))
    }

    rawProduct.value = snapshot.rawProduct
    product.value = snapshot.shopProduct
    selectedQuantity.value = 1
    selectedVariantId.value = resolveInitialGlobalProductDetailBottomSheetVariantId(product.value)
    selectedMediaId.value = resolveInitialGlobalProductDetailBottomSheetMediaId(snapshot.rawProduct)
  } catch (error) {
    if (requestSequence !== fetchRequestSequence) return
    rawProduct.value = null
    product.value = null
    productLoadError.value = error instanceof Error
      ? error.message
      : t('products.detail.notFound', 'Product not found')
  } finally {
    if (requestSequence === fetchRequestSequence) {
      isLoadingProduct.value = false
    }
  }
}

const resolveInitialGlobalProductDetailBottomSheetVariantId = (shopProduct: ShopProduct) => {
  const defaultVariant = shopProduct.variants.find(variant => variant.isDefault)
    || shopProduct.variants[0]
  return defaultVariant?.id || null
}

const resolveInitialGlobalProductDetailBottomSheetMediaId = (detail: GoProduct) => {
  const media = Array.isArray(detail.media) ? detail.media : []
  const visibleMedia = media.filter(item => item.url && item.is_visible !== false)
  const primaryMedia = visibleMedia.find(item => item.is_primary) || visibleMedia[0]
  return primaryMedia?.id !== undefined ? String(primaryMedia.id) : ''
}

const activeProductVariants = computed(() =>
  product.value?.variants || [],
)

const selectedVariant = computed<ShopProductVariant | null>(() => {
  if (!product.value || !selectedVariantId.value) return null
  return product.value.variants.find(variant => variant.id === selectedVariantId.value) || null
})

const selectedAvailability = computed(() =>
  selectedVariant.value?.availability || product.value?.availability || 'out_of_stock',
)

const selectedVariantWeight = computed(() => {
  const weight = Number(selectedVariant.value?.weightGrams || 0)
  return Number.isFinite(weight) && weight > 0 ? Math.round(weight) : null
})

const productMediaItems = computed<ProductMediaItem[]>(() => {
  const media = Array.isArray(rawProduct.value?.media) ? rawProduct.value.media : []
  return media
    .filter(item => item.url && item.is_visible !== false)
    .map((item, index) => {
      const kind: ProductMediaItem['kind'] = item.media_type === 'video' ? 'video' : 'image'
      const sourceUrl = String(item.url)
      const largeVariantUrl = kind === 'image'
        ? String(item.image_variants?.large?.url || '').trim()
        : ''
      return {
        id: item.id !== undefined ? String(item.id) : `${kind}-${index}`,
        kind,
        url: largeVariantUrl || sourceUrl,
        thumbnailUrl: String(item.thumbnail_url || item.image_variants?.thumbnail?.url || (kind === 'image' ? sourceUrl : '')),
        poster: item.poster_url ? String(item.poster_url) : undefined,
        alt: String(item.alt || item.title || product.value?.title || ''),
        index,
        isPrimary: Boolean(item.is_primary),
      }
    })
})

const selectedMedia = computed(() =>
  productMediaItems.value.find(media => media.id === selectedMediaId.value)
    || productMediaItems.value.find(media => media.isPrimary)
    || productMediaItems.value[0]
    || null,
)

const selectGlobalProductDetailBottomSheetMedia = (mediaId: string) => {
  if (productMediaItems.value.some(media => media.id === mediaId)) {
    selectedMediaId.value = mediaId
  }
}

const selectPreviousGlobalProductDetailBottomSheetMedia = () => {
  const items = productMediaItems.value
  if (items.length < 2) return
  const currentIndex = Math.max(0, items.findIndex(media => media.id === selectedMedia.value?.id))
  const nextMedia = items[(currentIndex - 1 + items.length) % items.length]
  if (nextMedia) selectedMediaId.value = nextMedia.id
}

const selectNextGlobalProductDetailBottomSheetMedia = () => {
  const items = productMediaItems.value
  if (items.length < 2) return
  const currentIndex = Math.max(0, items.findIndex(media => media.id === selectedMedia.value?.id))
  const nextMedia = items[(currentIndex + 1) % items.length]
  if (nextMedia) selectedMediaId.value = nextMedia.id
}

const parseGlobalProductDetailBottomSheetVariantOptions = (variant: ShopProductVariant | null) =>
  variant?.optionValues || {}

const variantOptionGroups = computed<VariantOptionGroup[]>(() => {
  const currentProduct = product.value
  if (!currentProduct || activeProductVariants.value.length < 2) return []

  const definitionMap = new Map(
    (currentProduct.productSpecificationTemplate?.specDefinitions || [])
      .map(definition => [definition.slug, definition]),
  )
  const optionSlugs = new Set<string>()
  activeProductVariants.value.forEach(variant => {
    Object.keys(variant.optionValues).forEach(slug => optionSlugs.add(slug))
  })

  return Array.from(optionSlugs).map(slug => {
    const definition = definitionMap.get(slug)
    const options = new Map<string, VariantOptionGroup['options'][number]>()
    activeProductVariants.value.forEach(variant => {
      const value = String(parseGlobalProductDetailBottomSheetVariantOptions(variant)[slug] || '').trim()
      if (!value || options.has(value)) return
      options.set(value, {
        value,
        label: value,
        selected: String(parseGlobalProductDetailBottomSheetVariantOptions(selectedVariant.value)[slug] || '') === value,
        available: activeProductVariants.value.some(candidateVariant =>
          String(candidateVariant.optionValues[slug] || '') === value
          && candidateVariant.availability === 'in_stock',
        ),
      })
    })

    return {
      slug,
      name: definition?.name || slug,
      options: Array.from(options.values()),
    }
  }).filter(group => group.options.length > 0)
})

const selectGlobalProductDetailBottomSheetVariantOption = (slug: string, value: string) => {
  const matchingVariant = activeProductVariants.value.find(variant =>
    String(variant.optionValues[slug] || '') === value
    && Object.entries(parseGlobalProductDetailBottomSheetVariantOptions(selectedVariant.value))
      .every(([optionSlug, optionValue]) => optionSlug === slug || !optionValue || variant.optionValues[optionSlug] === optionValue),
  ) || activeProductVariants.value.find(variant => String(variant.optionValues[slug] || '') === value)

  if (matchingVariant) selectedVariantId.value = matchingVariant.id
}

const normalizeGlobalProductDetailBottomSheetCurrency = (value: unknown) => {
  const currency = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(currency) ? currency : 'USD'
}

const selectedProductDisplayPrice = computed(() => {
  const variant = selectedVariant.value
  const requestedCurrency = normalizeGlobalProductDetailBottomSheetCurrency(displayCurrency.value)
  const variantDisplayPrice = variant?.displayPrices.find(price => price.currency === requestedCurrency)
    || variant?.displayPrices[0]
  if (variantDisplayPrice && variantDisplayPrice.amount > 0) {
    return { amount: variantDisplayPrice.amount, currency: variantDisplayPrice.currency }
  }
  return {
    amount: variant?.priceNumber || product.value?.displayPriceNumber || product.value?.priceNumber || 0,
    currency: variant?.currency || product.value?.displayPriceCurrency || product.value?.currency || 'USD',
  }
})

const selectedProductSourcePrice = computed(() => {
  const variant = selectedVariant.value
  return {
    amount: variant?.priceNumber ?? product.value?.priceNumber ?? 0,
    currency: variant?.currency || product.value?.currency || 'USD',
  }
})

const formattedSelectedProductPrice = computed(() => {
  const { amount, currency } = selectedProductDisplayPrice.value
  try {
    return new Intl.NumberFormat(locale.value || 'en-US', {
      style: 'currency',
      currency: normalizeGlobalProductDetailBottomSheetCurrency(currency),
    }).format(amount)
  } catch {
    return `${currency} ${Number(amount || 0).toFixed(2)}`
  }
})

const canAddSelectedProductToCart = computed(() =>
  Boolean(product.value && selectedProductDisplayPrice.value.amount > 0 && selectedAvailability.value === 'in_stock'),
)

const canBuySelectedProductNow = computed(() => canAddSelectedProductToCart.value)

const normalizeGlobalProductDetailBottomSheetQuantity = () => {
  const numericQuantity = Math.floor(Number(selectedQuantity.value))
  selectedQuantity.value = Number.isFinite(numericQuantity)
    ? Math.min(99, Math.max(1, numericQuantity))
    : 1
}

const decreaseGlobalProductDetailBottomSheetQuantity = () => {
  selectedQuantity.value = Math.max(1, selectedQuantity.value - 1)
}

const increaseGlobalProductDetailBottomSheetQuantity = () => {
  selectedQuantity.value = Math.min(99, selectedQuantity.value + 1)
}

const createGlobalProductDetailBottomSheetCartItem = () => {
  if (!product.value) return null
  const variant = selectedVariant.value
  const sourcePrice = selectedProductSourcePrice.value
  return toCartItem(product.value, {
    variantId: variant?.id || null,
    price: sourcePrice.amount,
    currency: sourcePrice.currency,
    title: product.value.title,
    thumbnail: product.value.thumbnail,
    weightGrams: variant?.weightGrams || null,
  })
}

const addGlobalProductDetailToCart = () => {
  normalizeGlobalProductDetailBottomSheetQuantity()
  const cartItem = createGlobalProductDetailBottomSheetCartItem()
  if (!cartItem || !canAddSelectedProductToCart.value) return
  addToCart(cartItem, selectedQuantity.value)
  purchaseFeedbackMessage.value = t('cart.added', 'Added to cart')
  closeGlobalProductDetailBottomSheet()
  openCart()
}

const buyGlobalProductDetailNow = () => {
  normalizeGlobalProductDetailBottomSheetQuantity()
  const cartItem = createGlobalProductDetailBottomSheetCartItem()
  if (!cartItem || !canBuySelectedProductNow.value) return
  addToCart(cartItem, selectedQuantity.value)
  closeGlobalProductDetailBottomSheet()
  openCheckout()
}

type ProductSpecValueItem = NonNullable<GoProduct['spec_values']>[number]

const formatGlobalProductDetailBottomSheetSpecificationValue = (item: ProductSpecValueItem) => {
  const definition = item.definition
  const value = String(item.value ?? '').trim()
  if (!value) return ''
  if (definition?.field_type === 'boolean') {
    return ['true', '1', 'yes'].includes(value.toLowerCase())
      ? t('common.yes', 'Yes')
      : t('common.no', 'No')
  }
  return definition?.unit ? `${value} ${definition.unit}` : value
}

const productSpecificationGroups = computed<ProductSpecificationGroup[]>(() => {
  const groups = new Map<string, ProductSpecificationGroup['items']>()
  for (const item of rawProduct.value?.spec_values || []) {
    const definition = item.definition
    if (!definition || definition.is_visible === false) continue
    const value = formatGlobalProductDetailBottomSheetSpecificationValue(item)
    if (!value) continue
    const groupName = String(definition.group || t('products.detail.specifications', 'Specifications'))
    const groupItems = groups.get(groupName) || []
    groupItems.push({
      slug: String(definition.slug || definition.name || groupItems.length),
      name: String(definition.name || definition.slug || ''),
      value,
    })
    groups.set(groupName, groupItems)
  }
  return Array.from(groups.entries()).map(([name, items]) => ({ name, items }))
})

watch(
  [isGlobalProductDetailBottomSheetOpen, productSlug, locale, displayCurrency, countryCode],
  ([isOpen]) => {
    if (isOpen) void loadGlobalProductDetailBottomSheetProduct()
  },
  { immediate: true },
)

watch(selectedVariantId, () => {
  purchaseFeedbackMessage.value = ''
})
</script>

<style scoped>
.global-product-detail-bottom-sheet-mask {
  position: fixed;
  inset: 0;
  z-index: 11000;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  box-sizing: border-box;
  padding: 1rem;
}

.global-product-detail-bottom-sheet-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.2);
  -webkit-backdrop-filter: blur(5px);
  backdrop-filter: blur(5px);
}

.global-product-detail-bottom-sheet-shell {
  position: relative;
  z-index: 1;
  display: flex;
  width: 90vw;
  height: 80vh;
  max-height: 80vh;
  flex-direction: column;
  overflow: hidden;
  box-sizing: border-box;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem 1rem 0 0;
  background: var(--tz-card-surface);
  box-shadow: 0 -20px 60px rgba(15, 23, 42, 0.16);
}

.global-product-detail-bottom-sheet-header {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.85rem 1rem;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.global-product-detail-bottom-sheet-header-copy {
  min-width: 0;
}

.global-product-detail-bottom-sheet-eyebrow {
  display: block;
  color: var(--tz-text-muted);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.global-product-detail-bottom-sheet-title,
.global-product-detail-bottom-sheet-product-title {
  overflow: hidden;
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.global-product-detail-bottom-sheet-close {
  display: inline-grid;
  width: 2.25rem;
  height: 2.25rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid var(--tz-border-strong);
  border-radius: 999px;
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
}

.global-product-detail-bottom-sheet-content {
  min-height: 0;
  flex: 1 1 auto;
  overflow: auto;
  padding: 1rem;
}

.global-product-detail-bottom-sheet-state {
  display: grid;
  min-height: 18rem;
  place-items: center;
  align-content: center;
  gap: 0.75rem;
  color: var(--tz-text-secondary);
  text-align: center;
}

.global-product-detail-bottom-sheet-state--error {
  color: var(--tz-status-danger-text);
}

.global-product-detail-bottom-sheet-inline-action {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  min-height: 2.25rem;
  padding: 0.45rem 0.7rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.5rem;
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
}

.global-product-detail-bottom-sheet-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.05fr) minmax(20rem, 0.95fr);
  gap: 1rem;
  min-height: 100%;
}

.global-product-detail-bottom-sheet-media-section,
.global-product-detail-bottom-sheet-information {
  min-width: 0;
}

.global-product-detail-bottom-sheet-media-section {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.global-product-detail-bottom-sheet-media-stage {
  position: relative;
  display: grid;
  min-height: 20rem;
  aspect-ratio: 1 / 1;
  place-items: center;
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-surface-subtle);
}

.global-product-detail-bottom-sheet-media-image,
.global-product-detail-bottom-sheet-media-video {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.global-product-detail-bottom-sheet-media-video {
  background: var(--tz-surface-subtle);
}

.global-product-detail-bottom-sheet-media-empty {
  display: grid;
  place-items: center;
  color: var(--tz-text-muted);
}

.global-product-detail-bottom-sheet-media-nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
}

.global-product-detail-bottom-sheet-media-nav--previous {
  left: 0.65rem;
}

.global-product-detail-bottom-sheet-media-nav--next {
  right: 0.65rem;
}

.global-product-detail-bottom-sheet-media-thumbnails {
  display: flex;
  gap: 0.5rem;
  overflow-x: auto;
  padding-bottom: 0.2rem;
}

.global-product-detail-bottom-sheet-media-thumbnail {
  position: relative;
  display: grid;
  width: 3.5rem;
  height: 3.5rem;
  flex: 0 0 auto;
  place-items: center;
  overflow: hidden;
  padding: 0;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  background: var(--tz-surface-subtle);
}

.global-product-detail-bottom-sheet-media-thumbnail--active {
  border-color: #059669;
  box-shadow: 0 0 0 2px rgba(5, 150, 105, 0.16);
}

.global-product-detail-bottom-sheet-media-thumbnail img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.global-product-detail-bottom-sheet-media-thumbnail-placeholder {
  color: var(--tz-text-muted);
}

.global-product-detail-bottom-sheet-media-thumbnail-badge {
  position: absolute;
  right: 0.25rem;
  bottom: 0.25rem;
  display: grid;
  width: 1.1rem;
  height: 1.1rem;
  place-items: center;
  border-radius: 999px;
  color: var(--tz-text-primary);
  background: var(--tz-surface-muted);
}

.global-product-detail-bottom-sheet-information {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
}

.global-product-detail-bottom-sheet-product-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.global-product-detail-bottom-sheet-product-title {
  font-size: 1.35rem;
  white-space: normal;
}

.global-product-detail-bottom-sheet-product-specification-template {
  margin: 0.35rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
}

.global-product-detail-bottom-sheet-price {
  flex: 0 0 auto;
  color: #059669;
  font-size: 1.15rem;
}

.global-product-detail-bottom-sheet-description {
  color: var(--tz-text-secondary);
  font-size: 0.86rem;
  line-height: 1.55;
}

.global-product-detail-bottom-sheet-description :deep(p) {
  margin: 0 0 0.45rem;
}

.global-product-detail-bottom-sheet-description :deep(p:last-child) {
  margin-bottom: 0;
}

.global-product-detail-bottom-sheet-variant-groups {
  display: grid;
  gap: 0.7rem;
}

.global-product-detail-bottom-sheet-variant-group {
  min-width: 0;
  margin: 0;
  border: 0;
  padding: 0;
}

.global-product-detail-bottom-sheet-variant-group legend,
.global-product-detail-bottom-sheet-variant-select label,
.global-product-detail-bottom-sheet-quantity-control > span {
  display: block;
  margin-bottom: 0.4rem;
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
  font-weight: 800;
  text-transform: uppercase;
}

.global-product-detail-bottom-sheet-variant-options {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.global-product-detail-bottom-sheet-variant-option {
  min-width: 4rem;
  min-height: 2.1rem;
  padding: 0.35rem 0.65rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.45rem;
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
}

.global-product-detail-bottom-sheet-variant-option--selected {
  border-color: #059669;
  color: #06111f;
  background: #059669;
}

.global-product-detail-bottom-sheet-variant-select {
  display: grid;
}

.global-product-detail-bottom-sheet-variant-select select {
  min-height: 2.4rem;
  padding: 0.45rem 0.65rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.5rem;
  color: var(--tz-text-primary);
  background: var(--tz-form-control-surface);
}

.global-product-detail-bottom-sheet-facts {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.global-product-detail-bottom-sheet-fact {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  min-height: 2rem;
  padding: 0.35rem 0.6rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.45rem;
  color: var(--tz-text-secondary);
  background: var(--tz-surface-subtle);
  font-size: 0.74rem;
}

.global-product-detail-bottom-sheet-fact--unavailable {
  color: var(--tz-status-danger-text);
  border-color: rgba(220, 38, 38, 0.3);
}

.global-product-detail-bottom-sheet-specifications {
  display: grid;
  gap: 0.6rem;
  padding-top: 0.85rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.global-product-detail-bottom-sheet-specifications h4,
.global-product-detail-bottom-sheet-specification-group h5 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 800;
}

.global-product-detail-bottom-sheet-specification-group {
  display: grid;
  gap: 0.4rem;
}

.global-product-detail-bottom-sheet-specification-group h5 {
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  text-transform: uppercase;
}

.global-product-detail-bottom-sheet-specification-group dl {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
  gap: 0.4rem;
  margin: 0;
}

.global-product-detail-bottom-sheet-specification-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  min-width: 0;
  padding: 0.45rem 0.55rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.4rem;
  background: var(--tz-surface-subtle);
  font-size: 0.72rem;
}

.global-product-detail-bottom-sheet-specification-item dt {
  overflow: hidden;
  color: var(--tz-text-muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.global-product-detail-bottom-sheet-specification-item dd {
  margin: 0;
  color: var(--tz-text-primary);
  font-weight: 700;
  text-align: right;
}

.global-product-detail-bottom-sheet-purchase {
  display: grid;
  gap: 0.75rem;
  margin-top: auto;
  padding-top: 0.85rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.global-product-detail-bottom-sheet-quantity-control {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.global-product-detail-bottom-sheet-quantity-control > span {
  margin: 0;
}

.global-product-detail-bottom-sheet-quantity-stepper {
  display: grid;
  grid-template-columns: 2.1rem 3rem 2.1rem;
  height: 2.1rem;
  overflow: hidden;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.45rem;
  background: var(--tz-surface-subtle);
}

.global-product-detail-bottom-sheet-quantity-stepper button,
.global-product-detail-bottom-sheet-quantity-stepper input {
  display: grid;
  min-width: 0;
  place-items: center;
  border: 0;
  color: var(--tz-text-primary);
  background: transparent;
  font: inherit;
  text-align: center;
}

.global-product-detail-bottom-sheet-quantity-stepper button:hover:not(:disabled) {
  background: rgba(5, 150, 105, 0.14);
  color: #059669;
}

.global-product-detail-bottom-sheet-quantity-stepper button:disabled {
  color: var(--tz-text-disabled);
}

.global-product-detail-bottom-sheet-quantity-stepper input {
  width: 100%;
  border-inline: 1px solid var(--tz-border-subtle);
  -moz-appearance: textfield;
}

.global-product-detail-bottom-sheet-quantity-stepper input::-webkit-inner-spin-button,
.global-product-detail-bottom-sheet-quantity-stepper input::-webkit-outer-spin-button {
  margin: 0;
  appearance: none;
}

.global-product-detail-bottom-sheet-purchase-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
}

.global-product-detail-bottom-sheet-purchase-action {
  display: inline-flex;
  min-height: 2.6rem;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.5rem 0.65rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.5rem;
  color: var(--tz-text-primary);
  font-size: 0.78rem;
  font-weight: 800;
}

.global-product-detail-bottom-sheet-purchase-action--secondary {
  background: var(--tz-surface-subtle);
}

.global-product-detail-bottom-sheet-purchase-action--primary {
  border-color: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  background: var(--tz-action-primary);
}

.global-product-detail-bottom-sheet-purchase-action--primary:hover:not(:disabled) {
  border-color: var(--tz-action-primary-hover);
  background: var(--tz-action-primary-hover);
}

.global-product-detail-bottom-sheet-purchase-action:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

.global-product-detail-bottom-sheet-feedback {
  margin: 0;
  color: #059669;
  font-size: 0.78rem;
}

.global-product-detail-bottom-sheet-fade-enter-active,
.global-product-detail-bottom-sheet-fade-leave-active {
  transition: opacity 180ms ease;
}

.global-product-detail-bottom-sheet-fade-enter-from,
.global-product-detail-bottom-sheet-fade-leave-to {
  opacity: 0;
}

.global-product-detail-bottom-sheet-slide-enter-active,
.global-product-detail-bottom-sheet-slide-leave-active {
  transition: transform 220ms ease, opacity 220ms ease;
}

.global-product-detail-bottom-sheet-slide-enter-from,
.global-product-detail-bottom-sheet-slide-leave-to {
  opacity: 0;
  transform: translateY(100%);
}

@media (max-width: 767px) {
  .global-product-detail-bottom-sheet-mask {
    width: 100vw;
    min-width: 100vw;
    height: 100vh;
    min-height: 100vh;
    padding-top: calc(var(--tz-safe-area-top, 0px) + var(--tz-mobile-dialog-inset, 2px));
    padding-right: calc(var(--tz-safe-area-right, 0px) + var(--tz-mobile-dialog-inset, 2px));
    padding-bottom: calc(var(--tz-safe-area-bottom, 0px) + var(--tz-mobile-dialog-inset, 2px));
    padding-left: calc(var(--tz-safe-area-left, 0px) + var(--tz-mobile-dialog-inset, 2px));
  }

  @supports (width: 100dvw) {
    .global-product-detail-bottom-sheet-mask {
      width: 100dvw;
      min-width: 100dvw;
    }
  }

  @supports (height: 100svh) {
    .global-product-detail-bottom-sheet-mask {
      height: 100svh;
      min-height: 100svh;
    }
  }

  @supports (height: 100dvh) {
    .global-product-detail-bottom-sheet-mask {
      height: 100dvh;
      min-height: 100dvh;
    }
  }

  .global-product-detail-bottom-sheet-shell {
    width: 100%;
    max-width: 100%;
    height: 100%;
    max-height: 100%;
  }

  .global-product-detail-bottom-sheet-content {
    padding: 0.75rem;
  }

  .global-product-detail-bottom-sheet-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .global-product-detail-bottom-sheet-media-stage {
    min-height: 14rem;
  }

  .global-product-detail-bottom-sheet-product-title {
    font-size: 1.08rem;
  }

  .global-product-detail-bottom-sheet-purchase-actions {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
