<template>
  <section v-if="product" class="product-page" :aria-label="metaTitle">
    <div class="product-hero">
      <section class="product-media-column" aria-label="Product media">
        <div class="product-media-layout">
          <div class="product-media-stage">
            <figure v-if="previewMedia?.kind === 'image'" class="product-media-frame">
              <NuxtImg :src="previewMedia.url" :alt="previewMedia.alt" loading="eager" format="webp" />
            </figure>
            <figure v-else-if="previewMedia?.kind === 'video'" class="product-media-frame product-media-frame--video">
              <video
                :src="previewMedia.url"
                :poster="previewMedia.poster || undefined"
                controls
                playsinline
                preload="metadata"
              />
            </figure>
            <div v-else class="product-media-placeholder">
              <Icon name="lucide:image-off" class="product-media-placeholder__icon" aria-hidden="true" />
            </div>
            <template v-if="productGalleryItems.length > 1">
              <button
                type="button"
                class="product-media-nav product-media-nav--previous"
                aria-label="Previous media"
                @click="selectPreviousMedia"
              >
                <Icon name="lucide:chevron-left" aria-hidden="true" />
              </button>
              <button
                type="button"
                class="product-media-nav product-media-nav--next"
                aria-label="Next media"
                @click="selectNextMedia"
              >
                <Icon name="lucide:chevron-right" aria-hidden="true" />
              </button>
            </template>
          </div>
          <div
            ref="productMediaThumbnailsRef"
            class="product-media-thumbnails"
            :class="{ 'product-media-thumbnails--centered': !productMediaSlotsOverflowing }"
            aria-label="Media thumbnails"
          >
            <template v-for="(media, index) in productMediaSlots" :key="media?.id || `media-placeholder-${index}`">
              <button
                v-if="media"
                type="button"
                class="product-media-thumbnail"
                :data-media-id="media.id"
                :class="{ 'product-media-thumbnail--active': selectedMediaId === media.id }"
                :aria-label="`${media.kind === 'video' ? 'View product video' : 'View product image'} ${index + 1}`"
                :aria-pressed="selectedMediaId === media.id"
                @click="selectMedia(media.id)"
              >
                <NuxtImg
                  v-if="media.thumbnailUrl"
                  :src="media.thumbnailUrl"
                  :alt="media.alt"
                  loading="lazy"
                  format="webp"
                />
                <span v-else class="product-media-thumbnail__placeholder">
                  <Icon
                    :name="media.kind === 'video' ? 'lucide:video' : 'lucide:image-off'"
                    aria-hidden="true"
                  />
                </span>
                <span v-if="media.kind === 'video'" class="product-media-thumbnail__badge" aria-hidden="true">
                  <Icon name="lucide:play" />
                </span>
              </button>
              <div v-else class="product-media-thumbnail product-media-thumbnail--placeholder" aria-hidden="true">
                <span class="product-media-thumbnail__placeholder">
                  <Icon name="lucide:image" aria-hidden="true" />
                </span>
              </div>
            </template>
          </div>
        </div>
      </section>
      <div class="product-summary">
        <h1 class="product-title">{{ product.name }}</h1>
        <p v-if="product.short_description" class="product-description" v-html="product.short_description" />
        <p v-else-if="productSummaryDescription" class="product-description">{{ productSummaryDescription }}</p>
        <div class="product-meta" aria-live="polite" aria-atomic="true">
          <span v-if="formattedPrice" class="product-price">{{ formattedPrice }}</span>
          <span v-if="product.product_type?.name" class="product-type-pill">{{ product.product_type.name }}</span>
        </div>
        <div v-if="activeVariants.length" class="product-purchase-panel">
          <div v-if="variantOptionGroups.length" class="variant-option-groups">
            <fieldset
              v-for="group in variantOptionGroups"
              :key="group.slug"
              class="variant-option-group"
            >
              <legend>{{ group.name }}</legend>
              <div class="variant-option-buttons">
                <button
                  v-for="option in group.options"
                  :key="`${group.slug}-${option.value}`"
                  type="button"
                  class="variant-option-button"
                  :class="{
                    'variant-option-button--selected': option.selected,
                    'variant-option-button--out': !option.available,
                    'variant-option-button--visual': group.presentation === 'color' || group.presentation === 'image',
                  }"
                  :aria-label="`${group.name}: ${option.label}`"
                  :aria-pressed="option.selected"
                  @click="selectVariantOption(group.slug, option.value)"
                >
                  <span
                    v-if="group.presentation === 'color' || group.presentation === 'image'"
                    class="variant-option-swatch"
                    :class="{ 'variant-option-swatch--image': Boolean(option.swatchUrl) || group.presentation === 'image' }"
                    :style="option.swatchUrl ? undefined : option.colorHex ? { backgroundColor: option.colorHex } : undefined"
                    aria-hidden="true"
                  >
                    <NuxtImg v-if="option.swatchUrl" :src="option.swatchUrl" :alt="option.label" loading="lazy" format="webp" />
                  </span>
                  <span class="variant-option-button__label">{{ option.label }}</span>
                  <small v-if="!option.available" class="variant-option-button__status">Out</small>
                </button>
              </div>
            </fieldset>
          </div>
          <div v-else-if="activeVariants.length > 1" class="product-variants">
            <label for="variant-select">Choose option</label>
            <select id="variant-select" v-model.number="selectedVariantId">
              <option
                v-for="variant in activeVariants"
                :key="variant.id"
                :value="variant.id"
              >
                {{ variantLabel(variant) }}
              </option>
            </select>
          </div>

          <dl v-if="selectedVariantWeight" class="selected-sku-facts" aria-live="polite" aria-atomic="true">
            <div class="selected-sku-fact-pill">
              <dt>Weight</dt>
              <dd>{{ selectedVariantWeight }}g</dd>
            </div>
          </dl>
        </div>
        <div class="product-quantity-control">
          <label for="product-quantity-input">{{ t('products.detail.quantity', 'Quantity') }}</label>
          <div class="product-quantity-stepper" role="group" :aria-label="t('products.detail.quantity', 'Quantity')">
            <button
              type="button"
              class="product-quantity-button"
              :disabled="selectedQuantity <= 1"
              :aria-label="t('products.detail.decreaseQuantity', 'Decrease quantity')"
              @click="decreaseSelectedQuantity"
            >
              <Icon name="lucide:minus" aria-hidden="true" />
            </button>
            <input
              id="product-quantity-input"
              :value="selectedQuantity"
              class="product-quantity-input"
              type="number"
              inputmode="numeric"
              min="1"
              :max="maxProductQuantity"
              @input="onSelectedQuantityInput"
            />
            <button
              type="button"
              class="product-quantity-button"
              :disabled="selectedQuantity >= maxProductQuantity"
              :aria-label="t('products.detail.increaseQuantity', 'Increase quantity')"
              @click="increaseSelectedQuantity"
            >
              <Icon name="lucide:plus" aria-hidden="true" />
            </button>
          </div>
        </div>
        <div v-if="productPaymentOptions.length" class="product-payment-selector" aria-label="Payment method">
          <div class="product-payment-selector__header">
            <span>{{ t('checkout.steps.payment', 'Choose payment method') }}</span>
            <small v-if="paymentMethodsLoading">{{ t('common.loading', 'Loading...') }}</small>
          </div>
          <div class="product-payment-options">
            <button
              v-for="option in productPaymentOptions"
              :key="productPaymentKey(option)"
              type="button"
              class="product-payment-option"
              :class="{
                'product-payment-option--selected': selectedProductPaymentMethod === productPaymentMethod(option),
                'product-payment-option--unavailable': !isProductPaymentAvailable(option),
              }"
              :disabled="!isProductPaymentAvailable(option)"
              :aria-disabled="!isProductPaymentAvailable(option)"
              :aria-pressed="selectedProductPaymentMethod === productPaymentMethod(option)"
              @click="selectProductPaymentMethod(productPaymentMethod(option))"
            >
              <span class="product-payment-option__logos" aria-hidden="true">
                <img
                  v-for="logo in productPaymentLogos(option)"
                  :key="logo.src"
                  :src="logo.src"
                  :alt="logo.alt"
                  :class="logo.className"
                  loading="lazy"
                />
              </span>
              <span class="product-payment-option__body">
                <span class="product-payment-option__title-row">
                  <span class="product-payment-option__title">{{ productPaymentTitle(option) }}</span>
                  <small v-if="!isProductPaymentAvailable(option)" class="product-payment-option__status">
                    {{ productPaymentUnavailableLabel(option) }}
                  </small>
                </span>
                <span class="product-payment-option__description">{{ productPaymentDescription(option) }}</span>
              </span>
              <Icon
                v-if="selectedProductPaymentMethod === productPaymentMethod(option)"
                name="lucide:check"
                class="product-payment-option__check"
                aria-hidden="true"
              />
            </button>
          </div>
          <p v-if="paymentMethodsError" class="product-payment-status">
            {{ paymentMethodsError }}
          </p>
        </div>
        <div class="product-actions" aria-label="Product actions">
          <button
            type="button"
            class="product-add-button"
            :disabled="!canAddToCart"
            @click="addSelectedToCart"
          >
            {{ canAddToCart ? 'Add to cart' : 'Out of stock' }}
          </button>
          <button
            type="button"
            class="product-buy-now-button"
            :disabled="!canBuyNow"
            @click="checkoutSelectedWithPayment"
          >
            {{ canBuyNow ? t('checkout.product.buyNow', 'Buy now') : productBuyNowUnavailableLabel }}
          </button>
        </div>
      </div>
    </div>

    <section
      v-if="specGroups.length"
      key="product-specifications"
      class="product-specs"
      aria-label="Product specifications"
    >
      <h2>Specifications</h2>
      <div v-for="group in specGroups" :key="group.name" class="spec-group">
        <h3 v-if="shouldShowSpecGroupHeading(group.name, specGroups.length)">{{ group.name }}</h3>
        <dl>
          <div v-for="item in group.items" :key="item.slug" class="spec-pill">
            <dt>{{ item.name }}</dt>
            <dd>{{ item.displayValue }}</dd>
          </div>
        </dl>
      </div>
    </section>

    <div key="product-information-tabs" class="product-tabs-anchor">
      <ProductInformationTabs
        :key="product.id"
        :details-html="product.description"
        :after-sales-html="product.after_sales_template?.content"
        :packaging-html="product.packaging_template?.content"
      />
    </div>

    <ProductRecommendations
      :key="`product-recommendations-${product.id}`"
      surface="product_detail_bottom"
      :title="t('recommendations.productDetailTitle', 'You may also like')"
      :product-id="product.id"
      :category-id="product.product_type_id || product.product_type?.id || null"
      :exclude-product-ids="[product.id]"
      :limit="6"
    />
  </section>
  <section v-else-if="pending" class="product-page product-page--pending">Loading...</section>
  <section v-else class="product-page product-page--error" role="alert">Product not found.</section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  createError,
  useAsyncData,
  useHead,
  useI18n,
  useLocalePath,
  useRequestURL,
  useRoute,
  useRuntimeConfig,
} from '#imports'
import { useCart } from '~/composables/useCart'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import { normalizeShopProduct, useShopProducts } from '~/composables/useShopProducts'
import { useSiteSettings } from '~/composables/usePublicSettings'
import ProductRecommendations from '~/components/shop/ProductRecommendations.vue'
import type { CheckoutPaymentOption } from '~/types/payment'
import {
  isPaymentOptionAvailable,
  paymentMethodFromOption,
  paymentPresentation,
  storefrontPaymentMethodOrder,
  type PaymentLogoAsset,
  type StorefrontPaymentMethod,
} from '~/utils/paymentPresentation'
import {
  buildProductSeoDocument,
  resolveProductMetaDescription,
  resolveProductMetaTitle,
} from '~/utils/seo/product'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import { buildProductPath, toAbsoluteSeoUrl } from '~/utils/seo/urls'
import {
  type StorefrontSeoAlternateLinkEntry,
  useStorefrontSeoRouteOverride,
} from '~/composables/seo/useStorefrontSeoLinks'

definePageMeta({
  layout: 'products',
})

interface ProductMediaImage {
  id?: number | string
  url: string
  alt?: string
}

interface ProductMedia {
  id?: number | string
  url: string
  media_type?: 'image' | 'video' | string
  role?: string
  variant_id?: number | null
  variant_option_value_id?: number | null
  thumbnail_url?: string
  poster_url?: string
  alt?: string
  title?: string
  is_primary?: boolean
  is_visible?: boolean
}

type ProductPreviewMedia =
  | {
      kind: 'image'
      url: string
      alt: string
    }
  | {
      kind: 'video'
      url: string
      poster?: string
    }

interface ProductGalleryItem {
  id: string
  kind: 'image' | 'video'
  url: string
  thumbnailUrl: string
  poster?: string
  alt: string
  isPrimary: boolean
  sourceIndex: number
}

interface ProductType {
  id?: number
  name: string
  slug: string
  spec_definitions?: SpecDefinition[]
}

interface SpecDefinition {
  id?: number
  name: string
  slug: string
  group?: string
  field_type: string
  presentation?: 'text' | 'color' | 'image' | string
  unit?: string
  is_visible?: boolean
  is_variant_option?: boolean
  sort_order?: number
}

interface ProductSpecValue {
  id?: number
  value: string
  definition?: SpecDefinition
}

interface ProductVariant {
  id: number
  sku?: string
  title?: string
  option_values?: string | Record<string, string>
  currency?: string
  price: number
  sale_price?: number | null
  display_price?: ProductDisplayPrice
  display_prices?: ProductDisplayPrice[]
  weight_grams?: number | null
  availability: 'in_stock' | 'out_of_stock'
  is_default?: boolean
  is_active?: boolean
}

interface ProductVariantOptionValue {
  id: number
  spec_definition_id?: number
  spec_slug: string
  value_key: string
  label: string
  color_hex?: string
  swatch_url?: string
  sort_order?: number
  is_enabled?: boolean
}

interface ProductInformationTemplate {
  id: number
  kind: 'after_sales' | 'packaging' | string
  name: string
  content: string
  locale?: string
}

interface ProductLocalizedRoute {
  locale: string
  slug: string
}

interface ProductDisplayPrice {
  amount: number
  currency: string
  rate?: number
  source?: string
  converted?: boolean
  fallback_reason?: string
}

interface GoProduct {
  id: number
  product_type_id?: number
  product_type?: ProductType
  name: string
  slug: string
  short_description?: string
  description?: string
  sku?: string
  currency?: string
  price: number
  sale_price?: number
  display_price?: ProductDisplayPrice
  display_prices?: ProductDisplayPrice[]
  availability?: 'in_stock' | 'out_of_stock'
  media?: ProductMedia[]
  thumbnail?: string
  meta_title?: string
  meta_description?: string
  localized_routes?: ProductLocalizedRoute[]
  after_sales_template?: ProductInformationTemplate | null
  packaging_template?: ProductInformationTemplate | null
  spec_values?: ProductSpecValue[]
  variants?: ProductVariant[]
  variant_option_values?: ProductVariantOptionValue[]
}

const route = useRoute()
const config = useRuntimeConfig()
const requestUrl = useRequestURL()
const { siteSettings } = useSiteSettings()
const { locale, t } = useI18n()
const localePath = useLocalePath()
const selectedVariantId = ref<number | null>(null)
const selectedProductPaymentMethod = ref<StorefrontPaymentMethod>('card')
const maxProductQuantity = 99
const selectedQuantity = ref(1)
const { addToCart, openCart, openCheckout } = useCart()
const { toCartItem } = useShopProducts()
const { displayCurrency, countryCode } = useStorefrontContext()
const { addToHistory } = useBrowsingHistory()
const { track: trackBehaviorEvent } = useBehaviorEvents()
const {
  paymentMethodOptions,
  paymentMethodsLoading,
  paymentMethodsError,
  loadPaymentMethods,
} = usePaymentMethods()

let activeTrackedProductID = 0
let productVisibleSince = 0

const slug = computed(() => String(route.params.slug || ''))

const siteOrigin = computed(() => {
  const value = (config.public as { siteUrl?: string }).siteUrl
  if (value && value.trim().length) {
    return value.replace(/\/$/, '')
  }
  return requestUrl.origin.replace(/\/$/, '')
})

const { data: product, pending } = await useAsyncData<GoProduct>(
  () => [
    'go-product',
    locale.value || 'default',
    slug.value,
    displayCurrency.value || 'default',
    countryCode.value || 'ZZ',
  ].map((part) => encodeURIComponent(String(part))).join(':'),
  async () => {
    if (!slug.value) {
      throw createError({
        statusCode: 404,
        statusMessage: 'Product not found',
      })
    }

    const base = ((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
    const response = await $fetch<any>(
      `${base}/products/${encodeURIComponent(slug.value)}`,
      {
        headers: {
          accept: 'application/json',
          ...(locale.value ? { 'Accept-Language': String(locale.value) } : {}),
          ...(displayCurrency.value ? { 'X-Display-Currency': displayCurrency.value } : {}),
          ...(countryCode.value && countryCode.value !== 'ZZ' ? { 'X-Market-Country': countryCode.value } : {}),
        },
        params: {
          locale: locale.value || undefined,
          currency: displayCurrency.value || undefined,
          country: countryCode.value !== 'ZZ' ? countryCode.value : undefined,
        },
      }
    )
    const data = response?.data || response
    if (!data || typeof data !== 'object') {
      throw createError({
        statusCode: 404,
        statusMessage: 'Product not found',
      })
    }
    return data as GoProduct
  },
  {
    server: true,
    watch: [() => slug.value, () => locale.value, () => displayCurrency.value, () => countryCode.value]
  }
)

if (!pending.value && !product.value) {
  throw createError({
    statusCode: 404,
    statusMessage: 'Product not found',
  })
}

const stripHtml = (value: string | null | undefined): string => {
  if (!value) return ''
  return value.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim()
}

const normalizeCurrencyCode = (value: unknown) => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : ''
}

const validDisplayPrice = (displayPrice?: ProductDisplayPrice | null) => {
  const amount = Number(displayPrice?.amount)
  const currency = normalizeCurrencyCode(displayPrice?.currency)
  if (!Number.isFinite(amount) || amount <= 0 || !currency) return null
  return { amount, currency }
}

const displayPriceSnapshotForCurrency = (displayPrices?: ProductDisplayPrice[]) => {
  const requestedCurrency = normalizeCurrencyCode(displayCurrency.value)
  if (!requestedCurrency || !Array.isArray(displayPrices)) return null
  const snapshot = displayPrices.find((price) => normalizeCurrencyCode(price?.currency) === requestedCurrency)
  return validDisplayPrice(snapshot)
}

const metaTitle = computed(() => resolveProductMetaTitle(
  product.value?.meta_title,
  product.value?.name,
))

const metaDescription = computed(() => resolveProductMetaDescription({
  metaDescription: product.value?.meta_description,
  shortDescription: product.value?.short_description,
  description: product.value?.description,
}))

const productSummaryDescription = computed(() => {
  const text = stripHtml(product.value?.description || '')
  if (text.length <= 220) return text
  return `${text.slice(0, 217)}...`
})

const productMediaImages = computed<ProductMedia[]>(() => {
  return (product.value?.media || []).filter((item) => {
    return item.media_type === 'image' && item.url && item.is_visible !== false
  })
})

const productImages = computed<ProductMediaImage[]>(() => {
  return productMediaImages.value.map((item) => ({
    id: item.id,
    url: item.url,
    alt: item.alt || item.title,
  }))
})

const parseVariantOptions = (variant: ProductVariant | null | undefined): Record<string, string> => {
  if (!variant?.option_values) return {}
  if (typeof variant.option_values === 'object') {
    return Object.entries(variant.option_values).reduce<Record<string, string>>((acc, [key, raw]) => {
      const value = String(raw ?? '').trim()
      if (key && value) {
        acc[key] = value
      }
      return acc
    }, {})
  }
  try {
    const parsed = JSON.parse(variant.option_values)
    return parsed && typeof parsed === 'object' ? parsed as Record<string, string> : {}
  } catch {
    return {}
  }
}

const selectedMediaId = ref<string | null>(null)
const productMediaSlotCount = 5
const productMediaThumbnailsRef = ref<HTMLElement | null>(null)

const productGalleryItems = computed<ProductGalleryItem[]>(() => {
  const currentProduct = product.value
  if (!currentProduct) return []

  const productThumbnail = String(currentProduct.thumbnail || '').trim()
  const currentVariant = (currentProduct.variants || []).find((variant) => variant.id === selectedVariantId.value) || null
  const currentOptions = parseVariantOptions(currentVariant)
  const optionValueIds = new Set(
    (currentProduct.variant_option_values || [])
      .filter((option) => option.is_enabled !== false && currentOptions[option.spec_slug] === option.value_key)
      .map((option) => Number(option.id))
      .filter((id) => Number.isFinite(id) && id > 0)
  )
  const visibleMedia = (currentProduct.media || []).filter((item) => item.url && item.is_visible !== false)
  const exactVariantMedia = selectedVariantId.value
    ? visibleMedia.filter((item) => Number(item.variant_id || 0) === Number(selectedVariantId.value))
    : []
  const optionMedia = visibleMedia.filter((item) => optionValueIds.has(Number(item.variant_option_value_id || 0)))
  const productMedia = visibleMedia.filter((item) => !item.variant_id && !item.variant_option_value_id)
  const scopedMedia = exactVariantMedia.length ? exactVariantMedia : optionMedia.length ? optionMedia : productMedia.length ? productMedia : visibleMedia
  const shouldIncludeProductThumbnail = exactVariantMedia.length === 0 && optionMedia.length === 0
  const mediaItems = scopedMedia
    .map((item, index): ProductGalleryItem | null => {
      const kind = String(item.media_type || '').toLowerCase() === 'video' ? 'video' : 'image'
      const url = String(item.url || '').trim()
      if (!url) return null

      const poster = String(item.poster_url || item.thumbnail_url || '').trim()
      const thumbnailUrl = kind === 'video'
        ? poster
        : String(item.thumbnail_url || item.url || '').trim()

      return {
        id: String(item.id ?? `product-media-${kind}-${index}-${url}`),
        kind,
        url,
        thumbnailUrl,
        poster: kind === 'video' ? poster || undefined : undefined,
        alt: item.alt || item.title || currentProduct.name,
        isPrimary: Boolean(item.is_primary || item.role === 'primary' || url === productThumbnail),
        sourceIndex: index,
      }
    })
    .filter((item): item is ProductGalleryItem => Boolean(item))
    .sort((left, right) => {
      if (left.kind !== right.kind) return left.kind === 'image' ? -1 : 1
      if (left.isPrimary !== right.isPrimary) return left.isPrimary ? -1 : 1
      return left.sourceIndex - right.sourceIndex
    })

  if (!shouldIncludeProductThumbnail || !productThumbnail || mediaItems.some((item) => item.url === productThumbnail)) {
    return mediaItems
  }

  return [
    {
      id: `product-thumbnail-${currentProduct.id}`,
      kind: 'image',
      url: productThumbnail,
      thumbnailUrl: productThumbnail,
      alt: currentProduct.name,
      isPrimary: true,
      sourceIndex: -1,
    },
    ...mediaItems,
  ]
})

const productMediaSlots = computed<Array<ProductGalleryItem | null>>(() => {
  const slotCount = Math.max(productMediaSlotCount, productGalleryItems.value.length)
  return Array.from({ length: slotCount }, (_, index) => productGalleryItems.value[index] || null)
})

const productMediaSlotsOverflowing = computed(() => productGalleryItems.value.length > productMediaSlotCount)

const centerSelectedMediaThumbnail = async () => {
  const mediaId = selectedMediaId.value
  if (!mediaId) return

  await nextTick()
  const container = productMediaThumbnailsRef.value
  if (!container) return

  const selected = Array.from(container.querySelectorAll<HTMLElement>('[data-media-id]'))
    .find((element) => element.dataset.mediaId === mediaId)
  if (!selected) return

  const containerRect = container.getBoundingClientRect()
  const selectedRect = selected.getBoundingClientRect()
  const scrollOptions: ScrollToOptions = {
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
  }

  if (container.scrollHeight > container.clientHeight + 1) {
    scrollOptions.top = container.scrollTop
      + selectedRect.top
      - containerRect.top
      - ((container.clientHeight - selectedRect.height) / 2)
  }

  if (container.scrollWidth > container.clientWidth + 1) {
    scrollOptions.left = container.scrollLeft
      + selectedRect.left
      - containerRect.left
      - ((container.clientWidth - selectedRect.width) / 2)
  }

  container.scrollTo(scrollOptions)
}

const primaryGalleryItem = computed(() => {
  return productGalleryItems.value.find((item) => item.isPrimary) || productGalleryItems.value[0] || null
})

const selectedMediaIndex = computed(() => {
  if (!productGalleryItems.value.length) return -1
  const index = productGalleryItems.value.findIndex((item) => item.id === selectedMediaId.value)
  return index >= 0 ? index : 0
})

const selectedMedia = computed(() => {
  return productGalleryItems.value[selectedMediaIndex.value] || null
})

const selectMedia = (mediaId: string) => {
  if (productGalleryItems.value.some((item) => item.id === mediaId)) {
    selectedMediaId.value = mediaId
  }
}

const selectPreviousMedia = () => {
  const items = productGalleryItems.value
  if (items.length < 2) return
  const currentIndex = selectedMediaIndex.value >= 0 ? selectedMediaIndex.value : 0
  const item = items[(currentIndex - 1 + items.length) % items.length]
  if (item) selectMedia(item.id)
}

const selectNextMedia = () => {
  const items = productGalleryItems.value
  if (items.length < 2) return
  const currentIndex = selectedMediaIndex.value >= 0 ? selectedMediaIndex.value : 0
  const item = items[(currentIndex + 1) % items.length]
  if (item) selectMedia(item.id)
}

watch(productGalleryItems, (items) => {
  if (!items.length) {
    selectedMediaId.value = null
    return
  }

  if (selectedMediaId.value && items.some((item) => item.id === selectedMediaId.value)) {
    return
  }

  const nextItem = items.find((item) => item.isPrimary) || items[0]
  if (nextItem) selectedMediaId.value = nextItem.id
}, { immediate: true })

watch(selectedMediaId, () => {
  void centerSelectedMediaThumbnail()
}, { flush: 'post' })

onMounted(() => {
  void centerSelectedMediaThumbnail()
})

const primaryImage = computed(() => {
  if (product.value?.thumbnail) {
    return product.value.thumbnail
  }
  const primaryMediaImage = productMediaImages.value.find((img) => img.is_primary || img.role === 'primary')
  if (primaryMediaImage?.url) {
    return primaryMediaImage.url
  }
  const firstProductMediaImage = productImages.value.find((img) => img.url)
  return firstProductMediaImage?.url || null
})

const previewMedia = computed<ProductPreviewMedia | null>(() => {
  const media = selectedMedia.value || primaryGalleryItem.value
  if (media?.kind === 'image') {
    return {
      kind: 'image',
      url: media.url,
      alt: media.alt || product.value?.name || metaTitle.value,
    }
  }

  if (media?.kind === 'video') {
    return {
      kind: 'video',
      url: media.url,
      poster: media.poster,
    }
  }

  return null
})

const primaryMediaThumbnail = computed(() => {
  const media = primaryGalleryItem.value
  if (!media) return ''
  if (media.kind === 'video') {
    return media.poster || media.thumbnailUrl || ''
  }
  return media.thumbnailUrl || media.url
})

const localizedProductPath = computed(() => localePath(
  buildProductPath(product.value?.slug || slug.value),
))

const localizedProductSeoRoutes = computed<StorefrontSeoAlternateLinkEntry[] | null>(() => {
  if (!product.value || !Array.isArray(product.value.localized_routes)) {
    return null
  }

  const routes = product.value.localized_routes
    .map((entry) => {
      const code = String(entry?.locale || '').trim()
      const translatedSlug = String(entry?.slug || '').trim()
      if (!code || !translatedSlug) return null
      return {
        code,
        path: localePath(buildProductPath(translatedSlug), code as any),
      }
    })
    .filter((entry): entry is StorefrontSeoAlternateLinkEntry => Boolean(entry))

  return routes.length ? routes : null
})

useStorefrontSeoRouteOverride(localizedProductSeoRoutes)

const canonicalUrl = computed(() => toAbsoluteSeoUrl(
  siteOrigin.value,
  localizedProductPath.value,
))

const shopProduct = computed(() => {
  return product.value ? normalizeShopProduct(product.value) : null
})

const activeVariants = computed(() => {
  return (product.value?.variants || []).filter((variant) => variant.is_active !== false)
})

const isVariantInStock = (variant: ProductVariant) => variant.availability === 'in_stock'
const requestedVariantId = computed(() => {
  const value = Number(route.query.variant || 0)
  return Number.isFinite(value) && value > 0 ? value : 0
})

watch([product, requestedVariantId], ([currentProduct, variantId]) => {
  const variants = (currentProduct?.variants || []).filter((variant) => variant.is_active !== false)
  if (variants.length === 0) {
    selectedVariantId.value = null
    return
  }
  const requestedVariant = variants.find((variant) => variant.id === variantId)
  if (requestedVariant) {
    selectedVariantId.value = requestedVariant.id
    return
  }
  const defaultVariant = variants.find((variant) => variant.is_default && isVariantInStock(variant))
    || variants.find(isVariantInStock)
    || variants.find((variant) => variant.is_default)
    || variants[0]
  if (!defaultVariant) return
  selectedVariantId.value = defaultVariant.id
}, { immediate: true })

const selectedVariant = computed(() => {
  if (!selectedVariantId.value) return null
  return activeVariants.value.find((variant) => variant.id === selectedVariantId.value) || null
})

type VariantOptionGroup = {
  slug: string
  name: string
  presentation: 'text' | 'color' | 'image'
  options: Array<{
    value: string
    label: string
    colorHex: string
    swatchUrl: string
    selected: boolean
    available: boolean
  }>
}

const humanizeSpecSlug = (slug: string) => {
  return slug
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/\b\w/g, char => char.toUpperCase())
}

const hiddenStorefrontSpecSlugs = new Set(['availability', 'sku'])

const variantOptionDefinitions = computed(() => {
  const definitions = product.value?.product_type?.spec_definitions || []
  return definitions
    .filter((definition) => (
      definition.is_visible !== false
      && definition.is_variant_option
      && !hiddenStorefrontSpecSlugs.has(String(definition.slug || '').trim().toLowerCase())
    ))
    .sort((left, right) => {
      const leftOrder = Number(left.sort_order || 0)
      const rightOrder = Number(right.sort_order || 0)
      if (leftOrder !== rightOrder) return leftOrder - rightOrder
      return String(left.name || left.slug).localeCompare(String(right.name || right.slug))
    })
})

const specDefinitionsBySlug = computed(() => {
  const entries = (product.value?.product_type?.spec_definitions || [])
    .filter((definition) => definition.slug)
    .map((definition) => [definition.slug, definition] as const)
  return new Map(entries)
})

const variantOptionSlugs = computed(() => {
  const slugs = variantOptionDefinitions.value.map((definition) => definition.slug)
  const seen = new Set(slugs)

  activeVariants.value.forEach((variant) => {
    Object.keys(parseVariantOptions(variant)).forEach((slug) => {
      if (!slug || seen.has(slug)) return
      if (hiddenStorefrontSpecSlugs.has(String(slug).trim().toLowerCase())) return
      const definition = specDefinitionsBySlug.value.get(slug)
      if (definition?.is_visible === false) return
      seen.add(slug)
      slugs.push(slug)
    })
  })

  return slugs
})

const currentVariantOptions = computed(() => {
  return selectedVariant.value ? parseVariantOptions(selectedVariant.value) : {}
})

const variantOptionGroups = computed<VariantOptionGroup[]>(() => {
  return variantOptionSlugs.value
    .map((slug) => {
      const definition = specDefinitionsBySlug.value.get(slug)
      const presentation: VariantOptionGroup['presentation'] = definition?.presentation === 'color'
        ? 'color'
        : definition?.presentation === 'image'
          ? 'image'
          : 'text'
      const optionsByValue = new Map<string, VariantOptionGroup['options'][number]>()

      activeVariants.value.forEach((variant) => {
        const value = String(parseVariantOptions(variant)[slug] || '').trim()
        if (!value) return

        const existing = optionsByValue.get(value)
        const available = variant.availability === 'in_stock'
        const metadata = variantOptionMetadata(slug, value)
        if (existing) {
          existing.available = existing.available || available
          return
        }

        optionsByValue.set(value, {
          value,
          label: metadata?.label || value,
          colorHex: metadata?.color_hex || '',
          swatchUrl: metadata?.swatch_url || '',
          selected: currentVariantOptions.value[slug] === value,
          available,
        })
      })

      return {
        slug,
        name: definition?.name || humanizeSpecSlug(slug),
        presentation,
        options: [...optionsByValue.values()],
      }
    })
    .filter((group) => group.options.length > 0)
})

const variantOptionMetadata = (slug: string, value: string) => {
  return (product.value?.variant_option_values || []).find((option) => (
    option.is_enabled !== false
      && option.spec_slug === slug
      && option.value_key === value
  ))
}

const selectVariantOption = (slug: string, value: string) => {
  const requestedOptions = {
    ...currentVariantOptions.value,
    [slug]: value,
  }

  const isExactMatch = (variant: ProductVariant) => {
    const options = parseVariantOptions(variant)
    return Object.entries(requestedOptions).every(([key, expectedValue]) => {
      return !expectedValue || options[key] === expectedValue
    })
  }

  const isFallbackMatch = (variant: ProductVariant) => {
    return parseVariantOptions(variant)[slug] === value
  }

  const exactVariant = activeVariants.value.find((variant) => isExactMatch(variant) && isVariantInStock(variant))
    || activeVariants.value.find(isExactMatch)
  const fallbackVariant = activeVariants.value.find((variant) => isFallbackMatch(variant) && isVariantInStock(variant))
    || activeVariants.value.find(isFallbackMatch)

  const nextVariant = exactVariant || fallbackVariant
  if (nextVariant) {
    selectedVariantId.value = nextVariant.id
  }
}

const variantLabel = (variant: ProductVariant) => {
  const options = Object.values(parseVariantOptions(variant)).filter(Boolean)
  const optionText = options.join(' / ')
  const title = variant.title || optionText || 'Option'
  const optionLabel = optionText && title !== optionText ? ` · ${optionText}` : ''
  const weightLabel = variant.weight_grams ? ` · ${variant.weight_grams}g` : ''
  return `${title}${optionLabel}${weightLabel}`
}

const selectedVariantWeight = computed(() => {
  const value = Number(selectedVariant.value?.weight_grams || 0)
  return Number.isFinite(value) && value > 0 ? Math.round(value) : null
})

const selectedCartTitle = computed(() => {
  const productName = product.value?.name || ''
  const variant = selectedVariant.value
  if (!variant) return productName

  const optionText = Object.values(parseVariantOptions(variant)).filter(Boolean).join(' / ')
  if (optionText) {
    return `${productName} - ${optionText}`
  }

  const variantTitle = String(variant.title || '').trim()
  if (variantTitle && variantTitle.toLowerCase() !== 'default') {
    return `${productName} - ${variantTitle}`
  }

  return productName
})

const effectivePrice = computed(() => {
  return selectedVariant.value?.sale_price
    ?? selectedVariant.value?.price
    ?? product.value?.sale_price
    ?? product.value?.price
    ?? 0
})

const currentCurrency = computed(() => {
  return normalizeCurrencyCode(selectedVariant.value?.currency || product.value?.currency) || 'USD'
})

const currentDisplayPrice = computed(() => {
  const selectedVariantDisplayPrice =
    validDisplayPrice(selectedVariant.value?.display_price) ||
    displayPriceSnapshotForCurrency(selectedVariant.value?.display_prices)
  if (selectedVariantDisplayPrice) return selectedVariantDisplayPrice

  const productDisplayPrice =
    validDisplayPrice(product.value?.display_price) ||
    displayPriceSnapshotForCurrency(product.value?.display_prices)
  if (productDisplayPrice) return productDisplayPrice

  return { amount: Number(effectivePrice.value || 0), currency: currentCurrency.value }
})

const selectedAvailability = computed(() => {
  if (selectedVariant.value) return selectedVariant.value.availability || 'out_of_stock'
  if (product.value && activeVariants.value.length === 0) {
    return product.value.availability || 'out_of_stock'
  }
  return 'out_of_stock'
})

const canAddToCart = computed(() => {
  return Boolean(
    product.value
      && Number(effectivePrice.value) > 0
      && selectedAvailability.value === 'in_stock'
  )
})

const normalizeSelectedQuantity = (value: unknown) => {
  const numeric = Math.floor(Number(value))
  if (!Number.isFinite(numeric)) return 1
  return Math.min(maxProductQuantity, Math.max(1, numeric))
}

const setSelectedQuantity = (value: unknown) => {
  selectedQuantity.value = normalizeSelectedQuantity(value)
}

const decreaseSelectedQuantity = () => {
  setSelectedQuantity(selectedQuantity.value - 1)
}

const increaseSelectedQuantity = () => {
  setSelectedQuantity(selectedQuantity.value + 1)
}

const onSelectedQuantityInput = (event: Event) => {
  const target = event.target as HTMLInputElement | null
  setSelectedQuantity(target?.value)
}

const fallbackProductPaymentOptions = computed<CheckoutPaymentOption[]>(() => [
  {
    id: 'card',
    code: 'card',
    provider: 'stripe',
    title: 'Credit / Debit cards',
    subtitle: '',
    description: '',
    enabled: true,
    available: false,
    unavailableReason: 'gateway_not_configured',
  },
  {
    id: 'paypal',
    code: 'paypal',
    provider: 'paypal',
    title: 'PayPal',
    subtitle: '',
    description: '',
    enabled: true,
    available: false,
    unavailableReason: 'gateway_not_configured',
  },
  {
    id: 'alipay',
    code: 'alipay',
    provider: 'alipay',
    title: 'Alipay',
    subtitle: '',
    description: '',
    enabled: true,
    available: false,
    unavailableReason: 'gateway_not_configured',
  },
  {
    id: 'wechat',
    code: 'wechat',
    provider: 'wechat',
    title: 'WeChat Pay',
    subtitle: '',
    description: '',
    enabled: true,
    available: false,
    unavailableReason: 'gateway_not_configured',
  },
])

const productPaymentMethod = (option: CheckoutPaymentOption) => paymentMethodFromOption(option)

const productPaymentOptions = computed(() => {
  const optionsByMethod = new Map<string, CheckoutPaymentOption>(
    fallbackProductPaymentOptions.value.map(option => [productPaymentMethod(option), option]),
  )

  paymentMethodOptions.value.forEach(option => {
    const method = productPaymentMethod(option)
    if (method) optionsByMethod.set(method, option)
  })

  return storefrontPaymentMethodOrder
    .map(method => optionsByMethod.get(method))
    .filter((option): option is CheckoutPaymentOption => Boolean(option))
})

const productPaymentKey = (option: CheckoutPaymentOption) =>
  `${productPaymentMethod(option)}-${option.id || option.code || option.provider || 'payment'}`

const productPaymentTitle = (option: CheckoutPaymentOption) => {
  const method = productPaymentMethod(option)
  if (!method) return option.title || option.code || option.id
  const presentation = paymentPresentation(method)
  return t(presentation.titleKey, presentation.title)
}

const productPaymentDescription = (option: CheckoutPaymentOption) => {
  if (option.description) return option.description
  const method = productPaymentMethod(option)
  if (!method) return option.subtitle || ''
  const presentation = paymentPresentation(method)
  return t(presentation.descriptionKey, presentation.description)
}

const productPaymentLogos = (option: CheckoutPaymentOption): PaymentLogoAsset[] => {
  const method = productPaymentMethod(option)
  return method ? paymentPresentation(method).logos : [{ src: '/icons/payment/default.svg', alt: productPaymentTitle(option) }]
}

const isProductPaymentAvailable = isPaymentOptionAvailable

const productPaymentUnavailableLabel = (option: CheckoutPaymentOption) => {
  const reason = String(option.unavailableReason || option.unavailable_reason || '').trim()
  if (reason === 'gateway_not_configured') {
    return t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
  }
  if (reason === 'gateway_config_invalid') {
    return t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
  }
  if (reason === 'disabled') {
    return t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
  }
  return reason ? reason.replace(/_/g, ' ') : t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
}

const selectedProductPaymentOption = computed(() => {
  return productPaymentOptions.value.find(option => productPaymentMethod(option) === selectedProductPaymentMethod.value) || null
})

const canBuyNow = computed(() => Boolean(
  canAddToCart.value
    && selectedProductPaymentOption.value
    && isProductPaymentAvailable(selectedProductPaymentOption.value),
))

const productBuyNowUnavailableLabel = computed(() => {
  if (!canAddToCart.value) return t('products.detail.outOfStock', 'Out of stock')
  if (paymentMethodsLoading.value) return t('common.loading', 'Loading...')
  return t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
})

const selectProductPaymentMethod = (method: StorefrontPaymentMethod | '') => {
  if (!method) return
  selectedProductPaymentMethod.value = method
}

const formattedPrice = computed(() => {
  const raw = currentDisplayPrice.value.amount
  if (raw == null) return ''
  const numeric = Number(raw)
  if (!Number.isFinite(numeric)) return ''
  const currencyCode = currentDisplayPrice.value.currency
  if (!currencyCode) return numeric.toFixed(2)
  try {
    return new Intl.NumberFormat(locale.value.replace('_', '-'), {
      style: 'currency',
      currency: currencyCode,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(numeric)
  } catch (err) {
    return numeric.toFixed(2)
  }
})

watch(product, (currentProduct) => {
  if (!import.meta.client || !currentProduct) return

  const productID = Number(currentProduct.id)
  if (!Number.isInteger(productID) || productID <= 0) return

  if (activeTrackedProductID !== productID) {
    trackProductDwell('product_change')
    activeTrackedProductID = productID
    productVisibleSince = Date.now()
    trackBehaviorEvent({
      eventType: 'product_view',
      productId: productID,
      metadata: {
        surface: 'product_page',
        product_type: currentProduct.product_type?.slug || '',
      },
    })
  }

  addToHistory({
    id: productID,
    title: currentProduct.name,
    thumbnail: primaryMediaThumbnail.value || '',
    price: formattedPrice.value,
    url: route.path,
  })
}, { immediate: true })

function trackProductDwell(reason: 'product_change' | 'visibility_hidden' | 'unmount') {
  if (!import.meta.client || !activeTrackedProductID || !productVisibleSince) return

  const durationSeconds = Math.min(
    1800,
    Math.max(0, Math.round((Date.now() - productVisibleSince) / 1000))
  )
  if (durationSeconds < 1) {
    productVisibleSince = reason === 'visibility_hidden' ? 0 : productVisibleSince
    return
  }

  trackBehaviorEvent({
    eventType: 'product_dwell',
    productId: activeTrackedProductID,
    metadata: {
      surface: 'product_page',
      duration_seconds: durationSeconds,
      end_reason: reason,
    },
  })
  productVisibleSince = reason === 'visibility_hidden' ? 0 : Date.now()
}

const handleProductVisibilityChange = () => {
  if (!import.meta.client) return
  if (document.visibilityState === 'hidden') {
    trackProductDwell('visibility_hidden')
    return
  }
  if (activeTrackedProductID && !productVisibleSince) {
    productVisibleSince = Date.now()
  }
}

const loadProductPaymentMethods = () => {
  const marketCountry = countryCode.value && countryCode.value !== 'ZZ'
    ? countryCode.value
    : undefined
  void loadPaymentMethods(marketCountry)
}

watch(productPaymentOptions, (options) => {
  const selectedExists = options.some(option => productPaymentMethod(option) === selectedProductPaymentMethod.value)
  if (selectedExists) return

  const firstAvailable = options.find(isProductPaymentAvailable) || options[0]
  const method = firstAvailable ? productPaymentMethod(firstAvailable) : ''
  if (method) selectedProductPaymentMethod.value = method
}, { immediate: true })

onMounted(() => {
  document.addEventListener('visibilitychange', handleProductVisibilityChange)
  loadProductPaymentMethods()
})

watch(countryCode, (value, previousValue) => {
  if (!import.meta.client || value === previousValue) return
  loadProductPaymentMethods()
})

onBeforeUnmount(() => {
  trackProductDwell('unmount')
  document.removeEventListener('visibilitychange', handleProductVisibilityChange)
})

const addSelectedProductToCart = () => {
  if (!product.value || !shopProduct.value || !canAddToCart.value) return

  const variant = selectedVariant.value
  return addToCart(toCartItem(shopProduct.value, {
    variantId: variant?.id || null,
    price: Number(effectivePrice.value),
    salePrice: variant?.sale_price ?? product.value.sale_price ?? null,
    sku: variant?.sku || product.value.sku || '',
    currency: currentCurrency.value,
    title: selectedCartTitle.value,
    thumbnail: primaryMediaThumbnail.value || undefined,
    weightGrams: selectedVariantWeight.value,
  }), selectedQuantity.value)
}

const addSelectedToCart = () => {
  const result = addSelectedProductToCart()

  if (result?.success) {
    openCart()
  }
}

const checkoutSelectedWithPayment = () => {
  const result = addSelectedProductToCart()
  if (result?.success && selectedProductPaymentMethod.value) {
    openCheckout(selectedProductPaymentMethod.value)
  }
}

const formatSpecValue = (item: ProductSpecValue) => {
  const definition = item.definition
  const value = String(item.value || '').trim()
  if (!definition) return value

  if (definition.field_type === 'boolean') {
    return value === 'true' ? 'Yes' : 'No'
  }
  if (definition.unit && value) {
    return `${value} ${definition.unit}`
  }
  return value
}

const genericSpecGroupNames = new Set(['specification', 'specifications', 'specs', '规格', '規格'])

const shouldShowSpecGroupHeading = (name: string, groupCount: number) => {
  const normalizedName = String(name || '').trim().toLowerCase()
  if (!normalizedName || genericSpecGroupNames.has(normalizedName)) return false
  return groupCount > 1
}

const specGroups = computed(() => {
  const groups = new Map<string, Array<{ slug: string; name: string; displayValue: string }>>()

  ;(product.value?.spec_values || []).forEach((item) => {
    const definition = item.definition
    if (!definition || definition.is_visible === false) return
    if (hiddenStorefrontSpecSlugs.has(String(definition.slug || '').trim().toLowerCase())) return

    const displayValue = formatSpecValue(item)
    if (!displayValue) return

    const groupName = definition.group || 'Specifications'
    const current = groups.get(groupName) || []
    current.push({
      slug: definition.slug,
      name: definition.name,
      displayValue,
    })
    groups.set(groupName, current)
  })

  return [...groups.entries()].map(([name, items]) => ({ name, items }))
})

const productSeoDocument = computed(() => {
  if (!product.value) return null

  const variantSeoPath = (variantId: number) => (
    `${localizedProductPath.value}?variant=${encodeURIComponent(String(variantId))}`
  )
  const variantSeoPrice = (variant: ProductVariant) => {
    const displayPrice = validDisplayPrice(variant.display_price)
      || displayPriceSnapshotForCurrency(variant.display_prices)
    return {
      amount: displayPrice?.amount ?? Number(variant.sale_price ?? variant.price ?? 0),
      currency: displayPrice?.currency
        || normalizeCurrencyCode(variant.currency || product.value?.currency)
        || 'USD',
    }
  }
  const seoVariants = activeVariants.value.map((variant) => {
    const price = variantSeoPrice(variant)
    return {
      id: variant.id,
      name: variantLabel(variant),
      sku: variant.sku,
      price: price.amount,
      currency: price.currency,
      availability: variant.availability,
      localizedPath: variantSeoPath(variant.id),
      imageUrls: productImages.value.map((image) => image.url),
    }
  })

  return buildProductSeoDocument(
    {
      name: product.value.name,
      brand: siteSettings.value.brandTitle,
      metaTitle: product.value.meta_title,
      metaDescription: product.value.meta_description,
      shortDescription: product.value.short_description,
      description: product.value.description,
      sku: product.value.sku,
      imageUrls: [
        product.value.thumbnail,
        ...productImages.value.map((image) => image.url),
      ],
      offer: {
        price: currentDisplayPrice.value.amount,
        currency: currentDisplayPrice.value.currency,
        availability: selectedAvailability.value,
        sku: selectedVariant.value?.sku || product.value.sku,
      },
      productGroupId: `product-${product.value.id}`,
      variesBy: variantOptionDefinitions.value.map((definition) => (
        `https://schema.org/${String(definition.slug || '').trim()}`
      )),
      variants: seoVariants,
    },
    {
      siteOrigin: siteOrigin.value,
      localizedPath: localizedProductPath.value,
    },
  )
})

useHead(() => {
  const seo = productSeoDocument.value
  const seoTitle = seo?.title || metaTitle.value
  const seoDescription = seo?.description || metaDescription.value
  const seoCanonicalUrl = seo?.canonicalUrl || canonicalUrl.value
  const metaEntries = [
    { name: 'description', content: seoDescription },
    { property: 'og:title', content: seoTitle },
    { property: 'og:description', content: seoDescription },
    { property: 'og:type', content: 'product' },
    { property: 'og:url', content: seoCanonicalUrl },
    { name: 'twitter:card', content: 'summary_large_image' },
    { name: 'twitter:title', content: seoTitle },
    { name: 'twitter:description', content: seoDescription }
  ]

  const seoImage = seo?.images[0] || ''
  if (seoImage) {
    metaEntries.push({ property: 'og:image', content: seoImage })
    metaEntries.push({ name: 'twitter:image', content: seoImage })
  }

  const seoOffer = seo?.schema?.['@type'] === 'Product' ? seo.schema.offers : null
  if (seoOffer) {
    metaEntries.push({ property: 'product:price:amount', content: seoOffer.price.toFixed(2) })
    metaEntries.push({ property: 'product:price:currency', content: seoOffer.priceCurrency })
  }

  return {
    title: seoTitle,
    meta: metaEntries.filter((entry) => Object.values(entry).every((value) => {
      if (typeof value !== 'string') return true
      return value.trim().length > 0
    })),
    script: seo?.schema ? [createSeoJsonLdScript(seo.schema)] : []
  }
})
</script>

<style scoped>
.product-page {
  --product-control-pill-height: 2.125rem;
  --product-control-pill-radius: 999px;
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
  color: #f8fafc;
  padding: 2rem 1rem 4rem;
}

.product-page--pending,
.product-page--error {
  padding: 4rem 1rem;
  color: #e2e8f0;
  text-align: center;
  font-size: 1.1rem;
}

.product-hero {
  display: grid;
  gap: 2rem;
  align-items: start;
}

.product-hero > * {
  min-width: 0;
}

@media (min-width: 900px) {
  .product-hero {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: clamp(2rem, 4vw, 4rem);
  }
}

.product-media-column {
  display: grid;
  gap: 1rem;
  min-width: 0;
}

.product-media-layout {
  --product-media-thumb-gap: 0.65rem;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--product-media-thumb-gap);
  min-width: 0;
}

.product-media-stage {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  min-width: 0;
}

.product-media-frame,
.product-media-placeholder {
  width: 100%;
  height: 100%;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0;
  border-radius: 0.75rem;
  overflow: hidden;
  background: #111;
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.product-media-frame img,
.product-media-frame video {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
}

.product-media-frame--video {
  background: #000;
}

.product-media-nav {
  position: absolute;
  top: 50%;
  display: inline-flex;
  width: 2.5rem;
  height: 2.5rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 50%;
  background: rgba(17, 17, 17, 0.82);
  color: #f8fafc;
  cursor: pointer;
  transform: translateY(-50%);
  transition: background 0.2s ease, border-color 0.2s ease, opacity 0.2s ease;
  z-index: 1;
}

.product-media-nav:hover {
  border-color: rgba(255, 255, 255, 0.42);
  background: rgba(17, 17, 17, 0.98);
}

.product-media-nav svg {
  width: 1.1rem;
  height: 1.1rem;
}

.product-media-nav--previous {
  left: 0.75rem;
}

.product-media-nav--next {
  right: 0.75rem;
}

.product-media-thumbnails {
  display: flex;
  flex-wrap: nowrap;
  gap: var(--product-media-thumb-gap);
  min-width: 0;
  min-height: 0;
  overflow-x: auto;
  padding: 0.1rem 0.1rem 0.35rem;
  overscroll-behavior: contain;
  scroll-behavior: smooth;
  scroll-padding-inline: 50%;
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.24) transparent;
}

.product-media-thumbnail {
  position: relative;
  flex: 0 0 4.75rem;
  width: 4.75rem;
  height: 4.75rem;
  aspect-ratio: 1 / 1;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 0.5rem;
  background: #171717;
  color: #f8fafc;
  cursor: pointer;
  padding: 0;
  transition: border-color 0.2s ease, opacity 0.2s ease, transform 0.2s ease;
}

.product-media-thumbnail--placeholder {
  cursor: default;
  border-style: dashed;
  background: rgba(255, 255, 255, 0.04);
}

.product-media-thumbnail:hover {
  border-color: rgba(255, 255, 255, 0.46);
  transform: translateY(-1px);
}

.product-media-thumbnail--placeholder:hover {
  border-color: rgba(255, 255, 255, 0.16);
  transform: none;
}

.product-media-thumbnail--active {
  border-color: rgba(181, 255, 109, 0.88);
  box-shadow: 0 0 0 2px rgba(181, 255, 109, 0.16);
}

.product-media-thumbnail img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
}

.product-media-thumbnail__placeholder {
  display: flex;
  width: 100%;
  height: 100%;
  align-items: center;
  justify-content: center;
  color: rgba(226, 232, 240, 0.66);
}

.product-media-thumbnail__placeholder svg {
  width: 1.15rem;
  height: 1.15rem;
}

.product-media-thumbnail__badge {
  position: absolute;
  right: 0.3rem;
  bottom: 0.3rem;
  display: inline-flex;
  width: 1.35rem;
  height: 1.35rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.4);
  border-radius: 50%;
  background: rgba(17, 17, 17, 0.84);
  color: #fff;
}

.product-media-thumbnail__badge svg {
  width: 0.72rem;
  height: 0.72rem;
}

.product-media-placeholder {
  flex-direction: column;
  padding: 1.5rem;
  text-align: center;
  color: var(--tz-text-secondary);
  border-style: dashed;
}

.product-media-placeholder__icon {
  width: 3rem;
  height: 3rem;
  margin-bottom: 0.85rem;
  color: rgba(181, 255, 109, 0.78);
}

.product-media-placeholder__title {
  margin: 0;
  color: #f8fafc;
  font-size: 1rem;
  font-weight: 800;
}

.product-media-placeholder__text {
  max-width: 22rem;
  margin: 0.4rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.92rem;
  line-height: 1.55;
}

@media (min-width: 900px) {
  .product-media-layout {
    grid-template-columns: clamp(3.5rem, 5vw, 5.25rem) minmax(0, 1fr);
    width: 100%;
    align-items: stretch;
    gap: 1rem;
  }

  .product-media-stage {
    grid-column: 2;
    grid-row: 1;
  }

  .product-media-thumbnails {
    grid-column: 1;
    grid-row: 1;
    align-self: stretch;
    flex-direction: column;
    justify-content: flex-start;
    height: 100%;
    max-height: 100%;
    min-height: 0;
    overflow-x: hidden;
    overflow-y: auto;
    padding: 0.1rem;
    scroll-padding-block: 50%;
  }

  .product-media-thumbnails--centered {
    justify-content: center;
  }

  .product-media-thumbnail {
    flex: 0 0 auto;
    width: 100%;
    height: auto;
    aspect-ratio: 1 / 1;
  }
}

.product-summary {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  min-width: 0;
  width: 100%;
}

.product-title {
  margin: 0;
  color: #f8fafc;
  font-size: clamp(1.6rem, 2vw + 0.9rem, 2.4rem);
  font-weight: 600;
}

.product-description {
  color: var(--tz-text-secondary);
  line-height: 1.65;
}

.product-description :deep(p) {
  margin-bottom: 0.5rem;
}

.product-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  color: var(--tz-text-secondary);
  font-size: 1rem;
}

.product-price {
  color: #b5ff6d;
  font-weight: 600;
  font-size: 1.15rem;
}

.product-type-pill {
  display: inline-flex;
  height: var(--product-control-pill-height);
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: var(--product-control-pill-radius);
  background: rgba(255, 255, 255, 0.06);
  color: var(--tz-text-secondary);
  font-size: 0.88rem;
  font-weight: 700;
  line-height: 1;
  padding: 0 0.72rem;
  white-space: nowrap;
}

@media (max-width: 767px) {
  .product-page {
    padding-inline: 1rem;
  }
}

.product-purchase-panel {
  display: grid;
  gap: 1.15rem;
  max-width: none;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 1rem;
  background: rgba(255, 255, 255, 0.045);
  padding: 1.15rem;
}

.variant-option-groups {
  display: grid;
  gap: 1rem;
}

.variant-option-group {
  min-width: 0;
  margin: 0;
  border: 0;
  padding: 0;
}

.product-variants {
  display: grid;
  gap: 0.5rem;
  max-width: 100%;
}

.variant-option-group legend {
  margin-bottom: 0.5rem;
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
  font-weight: 700;
  letter-spacing: 0;
  text-transform: uppercase;
}

.variant-option-buttons {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(6.5rem, 6.5rem));
  gap: 0.45rem;
}

.variant-option-button {
  display: inline-flex;
  width: 100%;
  height: var(--product-control-pill-height);
  min-height: 0;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: var(--product-control-pill-radius);
  background: rgba(255, 255, 255, 0.07);
  color: #f8fafc;
  cursor: pointer;
  font: inherit;
  font-size: 0.92rem;
  font-weight: 700;
  line-height: 1.2;
  box-sizing: border-box;
  padding: 0 0.7rem;
  text-align: center;
  transition: background 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
  overflow-wrap: anywhere;
}

.variant-option-button__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.variant-option-button__status {
  display: inline-flex;
  height: 1rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  color: #fca5a5;
  font-size: 0.66rem;
  font-weight: 800;
  line-height: 1;
}

.variant-option-button--visual {
  min-width: 0;
}

.variant-option-swatch {
  display: inline-flex;
  width: 1.35rem;
  height: 1.35rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 0.35rem;
  background: rgba(255, 255, 255, 0.12);
}

.variant-option-swatch--image {
  background: rgba(255, 255, 255, 0.08);
}

.variant-option-swatch img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.variant-option-button:hover {
  border-color: rgba(181, 255, 109, 0.65);
  background: rgba(181, 255, 109, 0.12);
  transform: translateY(-1px);
}

.variant-option-button--selected {
  border-color: #fff;
  background: #fff;
  color: #06111f;
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.16);
}

.variant-option-button--out:not(.variant-option-button--selected) {
  color: rgba(226, 232, 240, 0.68);
}

.product-variants label {
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0;
}

.product-variants select {
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 0.8rem;
  background: rgba(255, 255, 255, 0.08);
  color-scheme: dark;
  color: #fff;
  padding: 0.7rem 0.9rem;
}

.product-variants option {
  background: #111827;
  color: #f8fafc;
}

.selected-sku-facts {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  margin: 0;
  border-top: 1px solid rgba(255, 255, 255, 0.12);
  padding-top: 0.75rem;
}

.selected-sku-fact-pill {
  display: inline-flex;
  height: var(--product-control-pill-height);
  min-width: 0;
  align-items: center;
  justify-content: center;
  gap: 0.42rem;
  box-sizing: border-box;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: var(--product-control-pill-radius);
  background: rgba(255, 255, 255, 0.055);
  padding: 0 0.78rem;
}

.selected-sku-facts dt {
  color: rgba(226, 232, 240, 0.62);
  font-size: 0.7rem;
  font-weight: 700;
  line-height: 1;
  white-space: nowrap;
}

.selected-sku-facts dd {
  margin: 0;
  color: #f8fafc;
  font-size: 0.86rem;
  font-weight: 700;
  line-height: 1;
  overflow-wrap: anywhere;
}

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
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: var(--product-control-pill-radius);
  background: rgba(255, 255, 255, 0.055);
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
  color: #f8fafc;
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
  background: rgba(181, 255, 109, 0.14);
  color: #b5ff6d;
}

.product-quantity-button:disabled {
  cursor: not-allowed;
  color: rgba(226, 232, 240, 0.34);
}

.product-quantity-button svg {
  width: 0.9rem;
  height: 0.9rem;
}

.product-quantity-input {
  width: 100%;
  border-inline: 1px solid rgba(255, 255, 255, 0.12);
  color-scheme: dark;
  text-align: center;
  -moz-appearance: textfield;
}

.product-quantity-input::-webkit-inner-spin-button,
.product-quantity-input::-webkit-outer-spin-button {
  margin: 0;
  -webkit-appearance: none;
}

.product-payment-selector {
  display: grid;
  gap: 0.65rem;
}

.product-payment-selector__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
  font-weight: 700;
  text-transform: uppercase;
}

.product-payment-selector__header small {
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: none;
}

.product-payment-options {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.55rem;
}

.product-payment-option {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  min-height: 4.35rem;
  align-items: center;
  gap: 0.75rem;
  box-sizing: border-box;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 0.85rem;
  background: rgba(255, 255, 255, 0.045);
  color: #f8fafc;
  cursor: pointer;
  padding: 0.75rem;
  text-align: left;
  transition: background-color 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
}

.product-payment-option:hover:not(:disabled) {
  border-color: rgba(181, 255, 109, 0.62);
  background: rgba(181, 255, 109, 0.1);
  transform: translateY(-1px);
}

.product-payment-option--selected {
  border-color: rgba(181, 255, 109, 0.82);
  background: rgba(181, 255, 109, 0.14);
}

.product-payment-option--unavailable {
  cursor: not-allowed;
  opacity: 0.52;
}

.product-payment-option__logos {
  display: inline-flex;
  min-width: 3.1rem;
  max-width: 4.9rem;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.18rem;
}

.product-payment-option__logos img {
  display: block;
  width: auto;
  max-width: 2.35rem;
  height: 1rem;
  object-fit: contain;
}

.product-payment-option__logos img.payment-logo--alipay {
  max-width: 2.75rem;
}

.product-payment-option__body {
  display: grid;
  min-width: 0;
  gap: 0.24rem;
}

.product-payment-option__title-row {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
}

.product-payment-option__title {
  min-width: 0;
  overflow: hidden;
  font-size: 0.86rem;
  font-weight: 800;
  line-height: 1.1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-payment-option__status {
  display: inline-flex;
  height: 1rem;
  align-items: center;
  border-radius: 999px;
  background: rgba(245, 158, 11, 0.14);
  color: #fde68a;
  font-size: 0.62rem;
  font-weight: 700;
  line-height: 1;
  padding: 0 0.34rem;
  white-space: nowrap;
}

.product-payment-option__description {
  display: -webkit-box;
  overflow: hidden;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  font-weight: 600;
  line-height: 1.35;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.product-payment-option__check {
  width: 0.95rem;
  height: 0.95rem;
  color: #b5ff6d;
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
  background: #b5ff6d;
  color: #06111f;
}

.product-buy-now-button {
  border: 1px solid rgba(181, 255, 109, 0.55);
  background: rgba(181, 255, 109, 0.12);
  color: #eaffd0;
}

.product-add-button:hover:not(:disabled) {
  background: #c7ff8c;
  box-shadow: 0 0 0 3px rgba(181, 255, 109, 0.12);
  transform: translateY(-1px);
}

.product-buy-now-button:hover:not(:disabled) {
  border-color: rgba(181, 255, 109, 0.82);
  background: rgba(181, 255, 109, 0.18);
  transform: translateY(-1px);
}

.product-add-button:active:not(:disabled),
.product-buy-now-button:active:not(:disabled) {
  transform: translateY(1px);
}

.product-payment-status {
  flex-basis: 100%;
  margin: 0;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
}

.variant-option-button:focus-visible,
.product-variants select:focus-visible,
.product-quantity-button:focus-visible,
.product-quantity-input:focus-visible,
.product-add-button:focus-visible,
.product-buy-now-button:focus-visible,
.product-payment-option:focus-visible {
  outline: 2px solid #b5ff6d;
  outline-offset: 3px;
}

.product-add-button:disabled,
.product-buy-now-button:disabled {
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: rgba(148, 163, 184, 0.16);
  color: var(--tz-text-secondary);
  cursor: not-allowed;
  opacity: 1;
}

@media (max-width: 640px) {
  .product-payment-options {
    grid-template-columns: 1fr;
  }
}

.product-specs h2 {
  margin-bottom: 0.75rem;
  color: #f8fafc;
  font-size: 1.5rem;
}

.product-specs {
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.05);
  padding: 1.25rem;
}

.spec-group + .spec-group {
  margin-top: 1.25rem;
}

.spec-group h3 {
  margin-bottom: 0.75rem;
  color: rgba(255, 255, 255, 0.72);
  font-size: 0.9rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0;
}

.spec-group dl {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: 0.55rem;
}

.spec-pill {
  display: flex;
  height: var(--product-control-pill-height);
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.55rem;
  box-sizing: border-box;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: var(--product-control-pill-radius);
  background: rgba(255, 255, 255, 0.055);
  padding: 0 0.78rem;
}

.spec-pill dt {
  min-width: 0;
  overflow: hidden;
  color: rgba(255, 255, 255, 0.56);
  font-size: 0.74rem;
  font-weight: 700;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spec-pill dd {
  min-width: 0;
  overflow: hidden;
  color: #fff;
  font-size: 0.82rem;
  font-weight: 600;
  text-align: right;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 767px) {
  .product-media-thumbnail {
    flex-basis: 4.5rem;
    width: 4.5rem;
    height: 4.5rem;
  }

  .product-media-nav {
    width: 2.25rem;
    height: 2.25rem;
  }
}
</style>
