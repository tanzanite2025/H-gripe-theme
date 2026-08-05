<template>
  <section v-if="product" class="product-page" :aria-label="metaTitle">
    <div class="product-hero">
      <figure v-if="primaryImage" class="product-image">
        <NuxtImg :src="primaryImage" :alt="product.name || metaTitle" loading="lazy" format="webp" />
      </figure>
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
                  }"
                  :aria-pressed="option.selected"
                  @click="selectVariantOption(group.slug, option.value)"
                >
                  <span>{{ option.value }}</span>
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
            <div v-if="displaySKU">
              <dt>SKU</dt>
              <dd>{{ displaySKU }}</dd>
            </div>
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
        <button
          type="button"
          class="product-add-button"
          :disabled="!canAddToCart"
          @click="addSelectedToCart"
        >
          {{ canAddToCart ? 'Add to cart' : 'Out of stock' }}
        </button>
      </div>
    </div>

    <section v-if="productImages.length || productVideos.length" class="product-gallery" aria-label="Product gallery">
      <h2>Gallery</h2>
      <ul class="gallery-list">
        <li v-for="image in productImages" :key="image.id || image.url" class="gallery-item">
          <NuxtImg :src="image.url" :alt="image.alt || product.name || 'Product image'" loading="lazy" format="webp" />
        </li>
        <li v-for="video in productVideos" :key="video.id || video.url" class="gallery-item gallery-item--video">
          <video
            :src="video.url"
            :poster="video.poster_url || video.thumbnail_url"
            controls
            preload="metadata"
          />
        </li>
      </ul>
    </section>

    <ProductInformationTabs :key="product.id" :details-html="product.description" />

    <section v-if="specGroups.length" class="product-specs" aria-label="Product specifications">
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
  </section>
  <section v-else-if="pending" class="product-page product-page--pending">Loading...</section>
  <section v-else class="product-page product-page--error" role="alert">Product not found.</section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRuntimeConfig, useAsyncData, useHead } from '#imports'
import { useCart } from '~/composables/useCart'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import { usePaymentCurrencies } from '~/composables/usePaymentCurrencies'
import { normalizeShopProduct, useShopProducts } from '~/composables/useShopProducts'

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
  thumbnail_url?: string
  poster_url?: string
  alt?: string
  title?: string
  is_primary?: boolean
  is_visible?: boolean
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
  price: number
  sale_price?: number | null
  weight_grams?: number | null
  availability: 'in_stock' | 'out_of_stock'
  is_default?: boolean
  is_active?: boolean
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
  price: number
  sale_price?: number
  availability?: 'in_stock' | 'out_of_stock'
  media?: ProductMedia[]
  thumbnail?: string
  meta_title?: string
  meta_description?: string
  spec_values?: ProductSpecValue[]
  variants?: ProductVariant[]
}

const route = useRoute()
const config = useRuntimeConfig()
const { locale } = useI18n()
const selectedVariantId = ref<number | null>(null)
const { addToCart, openCart } = useCart()
const { toCartItem } = useShopProducts()
const { addToHistory } = useBrowsingHistory()
const { track: trackBehaviorEvent } = useBehaviorEvents()
const { defaultOrderCurrency, loadCurrencies } = usePaymentCurrencies()

await loadCurrencies()

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
        { headers: { accept: 'application/json' } }
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
    watch: [() => slug.value]
  }
)

const stripHtml = (value: string | null | undefined): string => {
  if (!value) return ''
  return value.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim()
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

const productVideos = computed<ProductMedia[]>(() => {
  return (product.value?.media || []).filter((item) => {
    return item.media_type === 'video' && item.url && item.is_visible !== false
  })
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

const parseVariantOptions = (variant: ProductVariant): Record<string, string> => {
  if (!variant.option_values) return {}
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

type VariantOptionGroup = {
  slug: string
  name: string
  options: Array<{
    value: string
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
      const optionsByValue = new Map<string, VariantOptionGroup['options'][number]>()

      activeVariants.value.forEach((variant) => {
        const value = String(parseVariantOptions(variant)[slug] || '').trim()
        if (!value) return

        const existing = optionsByValue.get(value)
        const available = variant.availability === 'in_stock'
        if (existing) {
          existing.available = existing.available || available
          return
        }

        optionsByValue.set(value, {
          value,
          selected: currentVariantOptions.value[slug] === value,
          available,
        })
      })

      return {
        slug,
        name: definition?.name || humanizeSpecSlug(slug),
        options: [...optionsByValue.values()],
      }
    })
    .filter((group) => group.options.length > 0)
})

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

const formattedPrice = computed(() => {
  const raw = effectivePrice.value
  if (raw == null) return ''
  const numeric = Number(raw)
  if (!Number.isFinite(numeric)) return ''
  const currencyCode = defaultOrderCurrency.value
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
    thumbnail: primaryImage.value || '',
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

onMounted(() => {
  document.addEventListener('visibilitychange', handleProductVisibilityChange)
})

onBeforeUnmount(() => {
  trackProductDwell('unmount')
  document.removeEventListener('visibilitychange', handleProductVisibilityChange)
})

const addSelectedToCart = () => {
  if (!product.value || !shopProduct.value || !canAddToCart.value) return

  const variant = selectedVariant.value
  const result = addToCart(toCartItem(shopProduct.value, {
    variantId: variant?.id || null,
    price: Number(effectivePrice.value),
    salePrice: variant?.sale_price ?? product.value.sale_price ?? null,
    sku: variant?.sku || product.value.sku || '',
    title: selectedCartTitle.value,
    thumbnail: primaryImage.value || undefined,
    weightGrams: selectedVariantWeight.value,
  }))

  if (result?.success) {
    openCart()
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
    const currencyCode = defaultOrderCurrency.value
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
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  }
}

.product-image {
  aspect-ratio: 1 / 1;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0;
  border-radius: 1rem;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.96);
}

.product-image img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
}

.product-summary {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  min-width: 0;
}

.product-title {
  margin: 0;
  color: #f8fafc;
  font-size: clamp(1.8rem, 2.4vw + 1rem, 2.8rem);
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

.product-add-button {
  width: fit-content;
  border: 0;
  border-radius: 999px;
  background: linear-gradient(135deg, #6b73ff, #B5FF6D);
  color: #06111f;
  cursor: pointer;
  font-weight: 800;
  padding: 0.85rem 1.35rem;
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.product-add-button:hover:not(:disabled) {
  transform: translateY(-1px);
}

.variant-option-button:focus-visible,
.product-variants select:focus-visible,
.product-add-button:focus-visible {
  outline: 2px solid #38bdf8;
  outline-offset: 3px;
}

.product-add-button:disabled {
  border: 1px solid rgba(148, 163, 184, 0.24);
  background: rgba(148, 163, 184, 0.16);
  color: var(--tz-text-secondary);
  cursor: not-allowed;
  opacity: 1;
}

.product-gallery h2,
.product-specs h2 {
  margin-bottom: 0.75rem;
  color: #f8fafc;
  font-size: 1.5rem;
}

.gallery-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
}

.gallery-item {
  aspect-ratio: 1 / 1;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.75rem;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.96);
}

.gallery-item img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
}

.gallery-item video {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
  background: #020617;
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
</style>
