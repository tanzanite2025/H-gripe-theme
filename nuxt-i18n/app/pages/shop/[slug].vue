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
              <p class="product-media-placeholder__title">No media preview yet</p>
              <p class="product-media-placeholder__text">Upload images or videos to reserve this space.</p>
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
          <div v-if="productGalleryItems.length > 1" class="product-media-thumbnails" aria-label="Media thumbnails">
            <button
              v-for="(media, index) in productGalleryItems"
              :key="media.id"
              type="button"
              class="product-media-thumbnail"
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
          </div>
        </div>
      </section>
      <div class="product-summary">
        <h1 class="product-title">{{ product.name }}</h1>
        <p v-if="product.short_description" class="product-description" v-html="product.short_description" />
        <p v-else-if="productSummaryDescription" class="product-description">{{ productSummaryDescription }}</p>
        <div class="product-meta" aria-live="polite" aria-atomic="true">
          <span v-if="formattedPrice" class="product-price">{{ formattedPrice }}</span>
          <span v-if="displaySKU" class="product-sku">SKU: {{ displaySKU }}</span>
          <span v-if="product.product_type?.name" class="product-sku">{{ product.product_type.name }}</span>
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
                  <span>{{ option.label }}</span>
                  <small v-if="!option.available">Out</small>
                </button>
              </div>
            </fieldset>
          </div>
          <div v-else-if="activeVariants.length > 1" class="product-variants">
            <label for="variant-select">Choose SKU</label>
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

          <dl class="selected-sku-facts" aria-live="polite" aria-atomic="true">
            <div v-if="selectedVariantWeight">
              <dt>Weight</dt>
              <dd>{{ selectedVariantWeight }}g</dd>
            </div>
            <div>
              <dt>Availability</dt>
              <dd>{{ selectedAvailability === 'in_stock' ? 'Available' : 'Out of stock' }}</dd>
            </div>
          </dl>
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

          <div v-if="productDirectPaymentOptions.length" class="product-direct-pay" aria-label="Express checkout">
            <button
              v-for="option in productDirectPaymentOptions"
              :key="productPaymentKey(option)"
              type="button"
              class="product-direct-pay-button"
              :class="{ 'product-direct-pay-button--unavailable': !isProductPaymentAvailable(option) }"
              :disabled="!canAddToCart"
              :aria-disabled="!canAddToCart"
              :title="productPaymentButtonTitle(option)"
              @click="checkoutSelectedWithPayment(productPaymentMethod(option))"
            >
              <Icon :name="productPaymentIcon(option)" aria-hidden="true" />
              <span>{{ productPaymentTitle(option) }}</span>
              <small v-if="!isProductPaymentAvailable(option)">
                {{ productPaymentUnavailableLabel(option) }}
              </small>
            </button>
          </div>

          <p v-if="paymentMethodsError" class="product-payment-status">
            {{ paymentMethodsError }}
          </p>
          <p v-else-if="paymentMethodsLoading" class="product-payment-status">
            Loading payment methods...
          </p>
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
        <h3>{{ group.name }}</h3>
        <dl>
          <template v-for="item in group.items" :key="item.slug">
            <dt>{{ item.name }}</dt>
            <dd>{{ item.displayValue }}</dd>
          </template>
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
  </section>
  <section v-else-if="pending" class="product-page product-page--pending">Loading...</section>
  <section v-else class="product-page product-page--error" role="alert">Product not found.</section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRuntimeConfig, useAsyncData, useHead } from '#imports'
import { useCart } from '~/composables/useCart'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import { normalizeShopProduct, useShopProducts } from '~/composables/useShopProducts'
import type { CheckoutPaymentOption } from '~/types/payment'

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
  after_sales_template?: ProductInformationTemplate | null
  packaging_template?: ProductInformationTemplate | null
  spec_values?: ProductSpecValue[]
  variants?: ProductVariant[]
  variant_option_values?: ProductVariantOptionValue[]
}

const route = useRoute()
const config = useRuntimeConfig()
const { locale, t } = useI18n()
const selectedVariantId = ref<number | null>(null)
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
  return 'https://example.com'
})

const { data: product, pending, error } = await useAsyncData<GoProduct | null>(
  () => `go-product:${slug.value}`,
  async () => {
    if (!slug.value) {
      return null
    }

    try {
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
      return response?.data || response || null
    } catch (err) {
      console.warn('Failed to load product', err)
      return null
    }
  },
  {
    server: true,
    default: () => null,
    watch: [() => slug.value, () => locale.value, () => displayCurrency.value, () => countryCode.value]
  }
)

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

const metaTitle = computed(() => product.value?.meta_title || product.value?.name || 'Product')

const rawDescription = computed(() => {
  if (product.value?.meta_description) {
    return product.value.meta_description
  }
  return stripHtml(product.value?.short_description || product.value?.description || '')
})

const metaDescription = computed(() => {
  const text = rawDescription.value
  if (text.length <= 160) return text
  return `${text.slice(0, 157)}...`
})

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

const canonicalUrl = computed(() => `${siteOrigin.value}/shop/${product.value?.slug || slug.value}`)

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

const variantOptionDefinitions = computed(() => {
  const definitions = product.value?.product_type?.spec_definitions || []
  return definitions
    .filter((definition) => definition.is_visible !== false && definition.is_variant_option)
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
  const title = variant.title || optionText || variant.sku || 'Option'
  const optionLabel = optionText && title !== optionText ? ` · ${optionText}` : ''
  const skuLabel = variant.sku ? ` · ${variant.sku}` : ''
  const weightLabel = variant.weight_grams ? ` · ${variant.weight_grams}g` : ''
  return `${title}${optionLabel}${skuLabel}${weightLabel}`
}

const displaySKU = computed(() => selectedVariant.value?.sku || product.value?.sku || '')

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

const fallbackProductPaymentOptions = computed<CheckoutPaymentOption[]>(() => [
  {
    id: 'card',
    code: 'card',
    provider: 'stripe',
    title: 'Stripe',
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

const productPaymentMethod = (option: CheckoutPaymentOption) => {
  const keys = [option.id, option.code, option.provider]
    .map(value => String(value || '').trim().toLowerCase())
    .filter(Boolean)

  if (keys.some(key => ['stripe', 'card', 'credit_card', 'credit-card'].includes(key))) return 'card'
  if (keys.includes('paypal')) return 'paypal'
  if (keys.includes('alipay')) return 'alipay'
  if (keys.includes('wechat')) return 'wechat'
  return keys[0] || ''
}

const productDirectPaymentOptions = computed(() => {
  const optionsByMethod = new Map<string, CheckoutPaymentOption>(
    fallbackProductPaymentOptions.value.map(option => [productPaymentMethod(option), option]),
  )

  paymentMethodOptions.value.forEach(option => {
    const method = productPaymentMethod(option)
    if (method) optionsByMethod.set(method, option)
  })

  return ['card', 'paypal', 'alipay', 'wechat']
    .map(method => optionsByMethod.get(method))
    .filter((option): option is CheckoutPaymentOption => Boolean(option))
})

const productPaymentKey = (option: CheckoutPaymentOption) =>
  `${productPaymentMethod(option)}-${option.id || option.code || option.provider || 'payment'}`

const productPaymentTitle = (option: CheckoutPaymentOption) => {
  switch (productPaymentMethod(option)) {
    case 'card': return 'Stripe'
    case 'paypal': return 'PayPal'
    case 'alipay': return 'Alipay'
    case 'wechat': return 'WeChat Pay'
    default: return option.title || option.code || option.id
  }
}

const productPaymentIcon = (option: CheckoutPaymentOption) => {
  switch (productPaymentMethod(option)) {
    case 'paypal': return 'lucide:wallet-cards'
    case 'alipay': return 'lucide:scan-line'
    case 'wechat': return 'lucide:qr-code'
    default: return 'lucide:credit-card'
  }
}

const isProductPaymentAvailable = (option: CheckoutPaymentOption) =>
  option.enabled !== false && option.available === true

const productPaymentUnavailableLabel = (option: CheckoutPaymentOption) => {
  const reason = String(option.unavailableReason || option.unavailable_reason || '').trim()
  if (reason === 'gateway_not_configured') {
    return t('checkout.payment.unconfigured', 'Not configured')
  }
  if (reason === 'gateway_config_invalid') {
    return t('checkout.payment.configInvalid', 'Configuration error')
  }
  if (reason === 'disabled') {
    return t('checkout.payment.disabled', 'Unavailable')
  }
  return reason ? reason.replace(/_/g, ' ') : t('checkout.payment.unavailable', 'Unavailable')
}

const productPaymentButtonTitle = (option: CheckoutPaymentOption) => {
  const title = productPaymentTitle(option)
  return isProductPaymentAvailable(option)
    ? `Pay with ${title}`
    : `${title}: ${productPaymentUnavailableLabel(option)}`
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
  }))
}

const addSelectedToCart = () => {
  const result = addSelectedProductToCart()

  if (result?.success) {
    openCart()
  }
}

const checkoutSelectedWithPayment = (paymentMethod: string) => {
  const result = addSelectedProductToCart()
  if (result?.success && paymentMethod) {
    openCheckout(paymentMethod)
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

const specGroups = computed(() => {
  const groups = new Map<string, Array<{ slug: string; name: string; displayValue: string }>>()

  ;(product.value?.spec_values || []).forEach((item) => {
    const definition = item.definition
    if (!definition || definition.is_visible === false) return

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

const productSchema = computed(() => {
  if (!product.value) return null

  const images: string[] = []
  if (product.value.thumbnail) {
    images.push(product.value.thumbnail)
  }
  productImages.value.forEach((img) => {
    if (img.url) images.push(img.url)
  })

  const offers = (() => {
    const raw = effectivePrice.value
    if (raw == null) return null
    const numeric = Number(raw)
    if (!Number.isFinite(numeric)) return null
    const currencyCode = currentCurrency.value
    if (!currencyCode) return null
    return {
      '@type': 'Offer',
      price: numeric,
      priceCurrency: currencyCode,
      availability: selectedAvailability.value === 'in_stock'
        ? 'https://schema.org/InStock'
        : 'https://schema.org/OutOfStock',
      url: canonicalUrl.value
    }
  })()

  return {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: metaTitle.value,
    description: metaDescription.value,
    sku: displaySKU.value,
    image: images,
    offers: offers || undefined
  }
})

useHead(() => {
  const metaEntries = [
    { name: 'description', content: metaDescription.value },
    { property: 'og:title', content: metaTitle.value },
    { property: 'og:description', content: metaDescription.value },
    { property: 'og:type', content: 'product' },
    { property: 'og:url', content: canonicalUrl.value },
    { name: 'twitter:card', content: 'summary_large_image' },
    { name: 'twitter:title', content: metaTitle.value },
    { name: 'twitter:description', content: metaDescription.value }
  ]

  if (primaryImage.value) {
    metaEntries.push({ property: 'og:image', content: primaryImage.value })
    metaEntries.push({ name: 'twitter:image', content: primaryImage.value })
  }

  if (formattedPrice.value) {
    metaEntries.push({ property: 'product:price:amount', content: formattedPrice.value.replace(/[^0-9.]/g, '') })
  }

  return {
    title: metaTitle.value,
    meta: metaEntries.filter((entry) => Object.values(entry).every((value) => {
      if (typeof value !== 'string') return true
      return value.trim().length > 0
    })),
    link: [
      {
        rel: 'canonical',
        href: canonicalUrl.value
      }
    ],
    script: productSchema.value
      ? [
          {
            type: 'application/ld+json',
            children: JSON.stringify(productSchema.value)
          }
        ]
      : []
  }
})
</script>

<style scoped>
.product-page {
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
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.85rem;
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
  gap: 0.65rem;
  min-width: 0;
  overflow-x: auto;
  padding: 0.1rem 0.1rem 0.35rem;
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

.product-media-thumbnail:hover {
  border-color: rgba(255, 255, 255, 0.46);
  transform: translateY(-1px);
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
    grid-template-columns: minmax(0, 1fr) clamp(5rem, 7vw, 6.75rem);
    width: 100%;
    align-items: stretch;
    gap: 1rem;
  }

  .product-media-thumbnails {
    flex-direction: column;
    height: 100%;
    max-height: 100%;
    overflow-x: hidden;
    overflow-y: auto;
    padding: 0.1rem 0.25rem 0.1rem 0.1rem;
  }

  .product-media-thumbnail {
    flex: 0 0 auto;
    width: 100%;
    height: auto;
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
  color: #f8fafc;
  font-weight: 600;
  font-size: 1.15rem;
}

.product-sku {
  color: var(--tz-text-secondary);
}

@media (max-width: 767px) {
  .product-page {
    padding-inline: 1rem;
  }
}

.product-purchase-panel {
  display: grid;
  gap: 1rem;
  max-width: 34rem;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  padding: 1rem;
}

.variant-option-groups {
  display: grid;
  gap: 0.9rem;
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
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.variant-option-button {
  display: inline-flex;
  min-height: 2.5rem;
  align-items: center;
  gap: 0.4rem;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 0.55rem;
  background: rgba(255, 255, 255, 0.07);
  color: #f8fafc;
  cursor: pointer;
  font: inherit;
  font-size: 0.92rem;
  font-weight: 700;
  padding: 0.55rem 0.75rem;
  transition: background 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
}

.variant-option-button--visual {
  min-width: 5.5rem;
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
  border-color: rgba(181, 255, 109, 0.9);
  background: rgba(181, 255, 109, 0.18);
  color: #f8fafc;
}

.variant-option-button--out:not(.variant-option-button--selected) {
  color: rgba(226, 232, 240, 0.68);
}

.variant-option-button small {
  color: #fca5a5;
  font-size: 0.72rem;
  font-weight: 800;
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
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(8rem, 1fr));
  gap: 0.65rem;
  margin: 0;
  border-top: 1px solid rgba(255, 255, 255, 0.12);
  padding-top: 0.85rem;
}

.selected-sku-facts div {
  min-width: 0;
}

.selected-sku-facts dt {
  color: rgba(226, 232, 240, 0.62);
  font-size: 0.76rem;
  font-weight: 700;
}

.selected-sku-facts dd {
  margin: 0.2rem 0 0;
  color: #f8fafc;
  font-size: 0.92rem;
  font-weight: 700;
  overflow-wrap: anywhere;
}

.product-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.product-add-button {
  min-height: 2.25rem;
  width: fit-content;
  border: 0;
  border-radius: 999px;
  background: #b5ff6d;
  color: #06111f;
  cursor: pointer;
  font-weight: 800;
  padding: 0.42rem 1rem;
  transition: background-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
}

.product-add-button:hover:not(:disabled) {
  background: #c7ff8c;
  box-shadow: 0 0 0 3px rgba(181, 255, 109, 0.12);
  transform: translateY(-1px);
}

.product-add-button:active:not(:disabled),
.product-direct-pay-button:active:not(:disabled) {
  transform: translateY(1px);
}

.product-direct-pay {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.product-direct-pay-button {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  gap: 0.4rem;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.055);
  color: #f8fafc;
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 700;
  padding: 0.42rem 0.75rem;
  transition: background-color 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
}

.product-direct-pay-button:hover:not(:disabled) {
  border-color: rgba(181, 255, 109, 0.7);
  background: rgba(181, 255, 109, 0.1);
}

.product-direct-pay-button > svg {
  width: 0.95rem;
  height: 0.95rem;
}

.product-direct-pay-button small {
  color: rgba(255, 255, 255, 0.48);
  font-size: 0.64rem;
  font-weight: 600;
}

.product-direct-pay-button--unavailable {
  border-color: rgba(255, 255, 255, 0.1);
  color: rgba(248, 250, 252, 0.72);
}

.product-direct-pay-button--unavailable:hover:not(:disabled) {
  border-color: rgba(245, 158, 11, 0.55);
  background: rgba(245, 158, 11, 0.08);
}

.product-payment-status {
  flex-basis: 100%;
  margin: 0;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
}

.variant-option-button:focus-visible,
.product-variants select:focus-visible,
.product-add-button:focus-visible,
.product-direct-pay-button:focus-visible {
  outline: 2px solid #b5ff6d;
  outline-offset: 3px;
}

.product-add-button:disabled {
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: rgba(148, 163, 184, 0.16);
  color: var(--tz-text-secondary);
  cursor: not-allowed;
  opacity: 1;
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
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 0.65rem 1rem;
}

.spec-group dt {
  color: rgba(255, 255, 255, 0.56);
}

.spec-group dd {
  color: #fff;
  font-weight: 600;
  text-align: right;
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
