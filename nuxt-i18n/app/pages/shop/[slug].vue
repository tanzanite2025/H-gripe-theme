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
        <li v-for="image in product.images" :key="image.id || image.url" class="gallery-item">
          <img :src="image.url" :alt="image.alt || product.name || 'Product image'" loading="lazy" />
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

interface ProductImage {
  id?: number | string
  url: string
  alt?: string
}

interface GoProduct {
  id: number
  name: string
  slug: string
  short_description?: string
  description?: string
  sku?: string
  price: number
  sale_price?: number
  images?: ProductImage[]
  thumbnail?: string
  meta_title?: string
  meta_description?: string
}

const route = useRoute()
const config = useRuntimeConfig()

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
      const response = await $fetch<GoProduct>(
        `${base}/products/${encodeURIComponent(slug.value)}`,
        { headers: { accept: 'application/json' } }
      )
      return response || null
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

const productImages = computed(() => {
  if (!product.value?.images) throw new Error("[CRITICAL] product images missing")
  return product.value.images
})

const primaryImage = computed(() => {
  if (product.value?.thumbnail) {
    return product.value.thumbnail
  }
  const firstProductImage = productImages.value.find((img) => img.url)
  return firstProductImage?.url || null
})

const canonicalUrl = computed(() => `${siteOrigin.value}/shop/${product.value?.slug || slug.value}`)

const formattedPrice = computed(() => {
  const raw = product.value?.sale_price || product.value?.price
  if (raw == null) return ''
  const numeric = Number(raw)
  if (!Number.isFinite(numeric)) return ''
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(numeric)
  } catch (err) {
    return `$${numeric.toFixed(2)}`
  }
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
    const raw = product.value?.sale_price || product.value?.price
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
