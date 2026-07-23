<template>
  <section v-if="product" class="product-page" :aria-label="metaTitle">
    <div class="product-hero">
      <figure v-if="primaryImage" class="product-image">
        <NuxtImg :src="primaryImage" :alt="product.name || metaTitle" loading="lazy" format="webp" />
      </figure>
      <div class="product-summary">
        <h1 class="product-title">{{ product.name }}</h1>
        <p v-if="product.short_description" class="product-description" v-html="product.short_description" />
        <p v-else-if="product.description" class="product-description" v-html="product.description" />
        <div class="product-meta" aria-live="polite" aria-atomic="true">
          <span v-if="formattedPrice" class="product-price">{{ formattedPrice }}</span>
          <span v-if="displaySKU" class="product-sku">SKU: {{ displaySKU }}</span>
          <span v-if="product.product_type?.name" class="product-sku">{{ product.product_type.name }}</span>
        </div>
        <div v-if="activeVariants.length" class="product-variants">
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
          <p class="variant-stock" role="status" aria-live="polite">Stock: {{ selectedVariant?.stock ?? 0 }}</p>
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
import { computed, ref, watch } from 'vue'
import { useRoute, useRuntimeConfig, useAsyncData, useHead } from '#imports'
import { useCart } from '~/composables/useCart'
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
  id: number
  name: string
  slug: string
}

interface SpecDefinition {
  id: number
  name: string
  slug: string
  group?: string
  field_type: string
  unit?: string
  is_visible?: boolean
}

interface ProductSpecValue {
  id: number
  value: string
  definition?: SpecDefinition
}

interface ProductVariant {
  id: number
  sku: string
  title?: string
  option_values?: string | Record<string, string>
  price: number
  sale_price?: number | null
  stock: number
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
  stock?: number
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

watch(product, (currentProduct) => {
  const variants = (currentProduct?.variants || []).filter((variant) => variant.is_active !== false)
  if (variants.length === 0) {
    selectedVariantId.value = null
    return
  }
  const defaultVariant = variants.find((variant) => variant.is_default) || variants[0]
  selectedVariantId.value = defaultVariant.id
}, { immediate: true })

const selectedVariant = computed(() => {
  if (!selectedVariantId.value) return null
  return activeVariants.value.find((variant) => variant.id === selectedVariantId.value) || null
})

const parseVariantOptions = (variant: ProductVariant) => {
  if (!variant.option_values) return {}
  if (typeof variant.option_values === 'object') return variant.option_values
  try {
    const parsed = JSON.parse(variant.option_values)
    return parsed && typeof parsed === 'object' ? parsed as Record<string, string> : {}
  } catch {
    return {}
  }
}

const variantLabel = (variant: ProductVariant) => {
  const options = Object.values(parseVariantOptions(variant)).filter(Boolean)
  const optionLabel = options.length ? ` · ${options.join(' / ')}` : ''
  return `${variant.title || variant.sku}${optionLabel}`
}

const displaySKU = computed(() => selectedVariant.value?.sku || product.value?.sku || '')

const effectivePrice = computed(() => {
  return selectedVariant.value?.sale_price
    ?? selectedVariant.value?.price
    ?? product.value?.sale_price
    ?? product.value?.price
    ?? 0
})

const selectedStock = computed(() => {
  if (selectedVariant.value) return selectedVariant.value.stock
  return product.value && activeVariants.value.length === 0 ? product.value.stock ?? 0 : 0
})

const canAddToCart = computed(() => {
  return Boolean(product.value && Number(effectivePrice.value) > 0 && selectedStock.value > 0)
})

const formattedPrice = computed(() => {
  const raw = effectivePrice.value
  if (raw == null) return ''
  const numeric = Number(raw)
  if (!Number.isFinite(numeric)) return ''
  try {
    return new Intl.NumberFormat(locale.value.replace('_', '-'), {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(numeric)
  } catch (err) {
    return `$${numeric.toFixed(2)}`
  }
})

watch(product, (currentProduct) => {
  if (!import.meta.client || !currentProduct) return

  const productID = Number(currentProduct.id)
  if (!Number.isInteger(productID) || productID <= 0) return

  addToHistory({
    id: productID,
    title: currentProduct.name,
    thumbnail: primaryImage.value || '',
    price: formattedPrice.value,
    url: route.path,
  })
}, { immediate: true })

const addSelectedToCart = () => {
  if (!product.value || !shopProduct.value || !canAddToCart.value) return

  const variant = selectedVariant.value
  const result = addToCart(toCartItem(shopProduct.value, {
    variantId: variant?.id || null,
    price: Number(effectivePrice.value),
    salePrice: variant?.sale_price ?? product.value.sale_price ?? null,
    sku: variant?.sku || product.value.sku || '',
    thumbnail: primaryImage.value || undefined,
    stockQuantity: selectedStock.value,
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
    return {
      '@type': 'Offer',
      price: numeric,
      priceCurrency: 'USD',
      availability: 'https://schema.org/InStock',
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

.product-variants {
  display: grid;
  gap: 0.5rem;
  margin-top: 1rem;
  max-width: 24rem;
}

.product-variants label {
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
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

.variant-stock {
  color: var(--tz-text-muted);
  font-size: 0.9rem;
}

.product-add-button {
  width: fit-content;
  border: 0;
  border-radius: 999px;
  background: linear-gradient(135deg, #6b73ff, #40ffaa);
  color: #06111f;
  cursor: pointer;
  font-weight: 800;
  padding: 0.85rem 1.35rem;
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.product-add-button:hover:not(:disabled) {
  transform: translateY(-1px);
}

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
  letter-spacing: 0.08em;
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
