<template>
  <main v-if="product" class="product-page" :aria-label="metaTitle">
    <div class="product-hero">
      <figure v-if="primaryImage" class="product-image">
        <img :src="primaryImage" :alt="product.name || metaTitle" loading="eager" />
      </figure>
      <div class="product-summary">
        <h1 class="product-title">{{ product.name }}</h1>
        <p v-if="product.short_description" class="product-description" v-html="product.short_description" />
        <p v-else-if="product.description" class="product-description" v-html="product.description" />
        <div class="product-meta">
          <span v-if="formattedPrice" class="product-price">{{ formattedPrice }}</span>
          <span v-if="product.sku" class="product-sku">SKU: {{ product.sku }}</span>
        </div>
      </div>
    </div>

    <section v-if="product.images?.length" class="product-gallery" aria-label="Product gallery">
      <h2>Gallery</h2>
      <ul class="gallery-list">
        <li v-for="image in product.images" :key="image.id || image.src" class="gallery-item">
          <img :src="image.src" :alt="image.alt || image.name || product.name || 'Product image'" loading="lazy" />
        </li>
      </ul>
    </section>

    <section v-if="product.description" class="product-content" aria-label="Product details">
      <h2>Details</h2>
      <article v-html="product.description" />
    </section>
  </main>
  <section v-else-if="pending" class="product-page product-page--pending">Loading…</section>
  <section v-else class="product-page product-page--error" role="alert">Product not found.</section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRuntimeConfig, useAsyncData, useHead } from '#imports'

interface WooImage {
  id?: number | string
  src: string
  name?: string
  alt?: string
}

interface WooPriceInfo {
  price?: string | number
  regular_price?: string | number
  sale_price?: string | number
  currency_code?: string
  currency_symbol?: string
}

interface WooProduct {
  id: number
  title: string
  name?: string
  slug: string
  permalink?: string
  preview_url?: string
  excerpt?: string
  short_description?: string
  description?: string
  content?: string
  sku?: string
  price?: string | number
  price_html?: string
  prices?: WooPriceInfo & { regular?: number; sale?: number; member?: number }
  images?: WooImage[]
  thumbnail?: string
  categories?: Array<{ id: number; name: string; slug: string }>
}

interface SeoImageEntry {
  url?: string
  alt?: string
  title?: string
  caption?: string
  focus_keyword?: string
  license?: string
  creator?: string
  credit_url?: string
  active?: boolean
}

interface SeoVideoEntry {
  title?: string
  description?: string
  focus_keyword?: string
  url?: string
  embed_url?: string
  content_url?: string
  thumbnail_url?: string
  upload_date?: string
  duration?: string
  type?: string
  active?: boolean
}

interface MyThemeSeoEntry {
  title?: string
  description?: string
  focus_keyword?: string
  images?: SeoImageEntry[]
  video?: SeoVideoEntry[]
  og?: {
    title?: string
    description?: string
    image?: string
  }
  twitter?: {
    card?: string
    title?: string
    description?: string
    image?: string
  }
}

interface ProductPayload {
  product: WooProduct | null
  seo: { id: number; payload?: MyThemeSeoEntry } | null
}

const route = useRoute()
const config = useRuntimeConfig()

const slug = computed(() => String(route.params.slug || ''))

const apiBase = computed(() => {
  const base = (config.public as { wpApiBase?: string }).wpApiBase || '/wp-json'
  return base.replace(/\/$/, '')
})

const wpApiBase = computed(() => {
  const base = (config.public as { wpApiBase?: string }).wpApiBase || ''
  return base.replace(/\/$/, '')
})

const siteOrigin = computed(() => {
  const value = (config.public as { siteUrl?: string }).siteUrl
  if (value && value.trim().length) {
    return value.replace(/\/$/, '')
  }
  return 'https://example.com'
})

const { data: productBundle, pending, error } = await useAsyncData<ProductPayload>(
  () => `tanz-product:${slug.value}`,
  async () => {
    if (!slug.value || !apiBase.value) {
      return { product: null, seo: null }
    }

    try {
      // 使用 Tanzanite API 通过 slug 搜索商品
      const response = await $fetch<{ items: WooProduct[] }>(
        `${apiBase.value}/tanzanite/v1/products?keyword=${encodeURIComponent(slug.value)}&per_page=1`,
        { headers: { accept: 'application/json' } }
      )
      // 从返回的商品列表中找到精确匹配 slug 的商品
      const product = response.items?.find(p => p.slug === slug.value) || null

      if (!product) {
        return { product: null, seo: null }
      }

      let seo: ProductPayload['seo'] = null
      if (apiBase.value) {
        try {
          seo = await $fetch<{ id: number; payload?: MyThemeSeoEntry }>(
            `${apiBase.value}/mytheme/v1/seo/product/${product.id}`,
            { headers: { accept: 'application/json' } }
          )
        } catch (err) {
          console.warn('Failed to load product SEO', err)
        }
      }

      return { product, seo }
    } catch (err) {
      console.warn('Failed to load Tanzanite product', err)
      return { product: null, seo: null }
    }
  },
  {
    server: true,
    default: () => ({ product: null, seo: null }),
    watch: [() => slug.value, () => apiBase.value]
  }
)

const product = computed(() => {
  const p = productBundle.value?.product ?? null
  if (!p) return null
  // 兼容处理：将 title 映射到 name，preview_url 映射到 permalink
  return {
    ...p,
    name: p.name || p.title,
    permalink: p.permalink || p.preview_url,
    short_description: p.short_description || p.excerpt,
    description: p.description || p.content,
  }
})
const seoPayload = computed<MyThemeSeoEntry | null>(() => productBundle.value?.seo?.payload ?? null)

const stripHtml = (value: string | null | undefined): string => {
  if (!value) return ''
  return value.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim()
}

const metaTitle = computed(() => seoPayload.value?.title || product.value?.name || 'Product')

const rawDescription = computed(() => {
  if (seoPayload.value?.description) {
    return seoPayload.value.description
  }
  return stripHtml(product.value?.short_description || product.value?.description || '')
})

const metaDescription = computed(() => {
  const text = rawDescription.value
  if (text.length <= 160) return text
  return `${text.slice(0, 157)}...`
})

const seoImages = computed(() => {
  if (!seoPayload.value?.images) return [] as SeoImageEntry[]
  return seoPayload.value.images.filter((image) => image && image.active !== false)
})

const productImages = computed(() => product.value?.images || [])

const primaryImage = computed(() => {
  const seoImage = seoImages.value.find((img) => img.url)
  if (seoImage?.url) {
    return seoImage.url
  }
  // Tanzanite API 提供 thumbnail 字段
  if (product.value?.thumbnail) {
    return product.value.thumbnail
  }
  const firstProductImage = productImages.value.find((img) => img.src)
  return firstProductImage?.src || null
})

const keywords = computed(() => {
  const terms = new Set<string>()
  if (seoPayload.value?.focus_keyword) terms.add(seoPayload.value.focus_keyword)
  seoImages.value.forEach((img) => {
    if (img.focus_keyword) terms.add(img.focus_keyword)
  })
  return Array.from(terms).join(', ')
})

const canonicalUrl = computed(() => product.value?.permalink || new URL(route.fullPath || '/', `${siteOrigin.value}/`).toString())

const formattedPrice = computed(() => {
  const prices = product.value?.prices
  if (!prices) return ''
  // Tanzanite API 返回 regular, sale, member
  const raw = prices.sale || prices.regular || prices.price || prices.regular_price || product.value?.price
  if (raw == null) return ''
  const numeric = Number(raw)
  if (!Number.isFinite(numeric)) return ''
  try {
    const currency = prices.currency_code || 'USD'
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(numeric)
  } catch (err) {
    const symbol = prices.currency_symbol || '$'
    return `${symbol}${numeric.toFixed(2)}`
  }
})

const productSchema = computed(() => {
  if (!product.value) return null

  const images: string[] = []
  if (seoImages.value.length) {
    seoImages.value.forEach((img) => {
      if (img.url) images.push(img.url)
    })
  }
  if (!images.length) {
    // Tanzanite API 提供 thumbnail
    if (product.value.thumbnail) {
      images.push(product.value.thumbnail)
    }
    productImages.value.forEach((img) => {
      if (img.src) images.push(img.src)
    })
  }

  const offers = (() => {
    const prices = product.value?.prices
    if (!prices) return null
    const raw = prices.sale || prices.regular || prices.price || prices.regular_price || product.value?.price
    if (raw == null) return null
    const numeric = Number(raw)
    if (!Number.isFinite(numeric)) return null
    const currency = prices.currency_code || 'USD'
    return {
      '@type': 'Offer',
      price: numeric,
      priceCurrency: currency,
      availability: 'https://schema.org/InStock',
      url: canonicalUrl.value
    }
  })()

  return {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: metaTitle.value,
    description: metaDescription.value,
    sku: product.value?.sku,
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

  if (keywords.value) {
    metaEntries.push({ name: 'keywords', content: keywords.value })
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
  padding: 2rem 1rem 4rem;
}

.product-page--pending,
.product-page--error {
  padding: 4rem 1rem;
  text-align: center;
  font-size: 1.1rem;
}

.product-hero {
  display: grid;
  gap: 2rem;
  align-items: start;
}

@media (min-width: 900px) {
  .product-hero {
    grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr);
  }
}

.product-image {
  margin: 0;
  border-radius: 1rem;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.04);
}

.product-image img {
  width: 100%;
  display: block;
  object-fit: cover;
}

.product-summary {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.product-title {
  margin: 0;
  font-size: clamp(1.8rem, 2.4vw + 1rem, 2.8rem);
  font-weight: 600;
}

.product-description :deep(p) {
  margin-bottom: 0.5rem;
}

.product-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  font-size: 1rem;
}

.product-price {
  font-weight: 600;
  font-size: 1.15rem;
}

.product-gallery h2,
.product-content h2 {
  margin-bottom: 0.75rem;
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
  border-radius: 0.75rem;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.05);
}

.gallery-item img {
  width: 100%;
  display: block;
  object-fit: cover;
}

.product-content article :deep(p) {
  margin-bottom: 1rem;
  line-height: 1.6;
}
</style>
