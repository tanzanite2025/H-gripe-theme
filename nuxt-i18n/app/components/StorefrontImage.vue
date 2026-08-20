<template>
  <NuxtImg
    v-if="props.optimize && canOptimize && !optimizationFailed"
    class="tz-storefront-image"
    :src="transformSource"
    :alt="alt"
    :width="width"
    :height="height"
    :sizes="resolvedSizes"
    :densities="densities"
    :format="format"
    :quality="quality"
    :loading="loading"
    :fetchpriority="fetchpriority"
    :decoding="decoding"
    v-bind="$attrs"
    @error="handleOptimizationError"
  />
  <img
    v-else
    class="tz-storefront-image"
    :src="fallbackSource"
    :alt="alt"
    :width="width"
    :height="height"
    :loading="loading"
    :fetchpriority="fetchpriority"
    :decoding="decoding"
    v-bind="$attrs"
    @error="handleFallbackError"
  >
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRuntimeConfig } from '#imports'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'

defineOptions({
  inheritAttrs: false,
})

const emit = defineEmits(['error'])

type StorefrontImagePreset =
  | 'avatar'
  | 'card'
  | 'content'
  | 'gallery'
  | 'hero'
  | 'logo'
  | 'swatch'
  | 'thumbnail'

const PRESET_SIZES: Record<StorefrontImagePreset, string> = {
  avatar: 'xs:56px sm:64px',
  card: 'xs:50vw sm:33vw md:280px lg:320px',
  content: 'xs:100vw sm:100vw md:768px lg:1024px',
  gallery: 'xs:100vw sm:100vw md:50vw lg:640px xl:960px',
  hero: 'xs:100vw sm:100vw md:100vw lg:1280px',
  logo: 'xs:120px sm:160px md:280px',
  swatch: 'xs:40px sm:40px',
  thumbnail: 'xs:72px sm:96px',
}

const props = withDefaults(defineProps<{
  src: string
  alt?: string
  preset?: StorefrontImagePreset
  sizes?: string
  densities?: string
  format?: string
  quality?: number | string
  optimize?: boolean
  loading?: 'eager' | 'lazy'
  fetchpriority?: 'high' | 'low' | 'auto'
  decoding?: 'async' | 'sync' | 'auto'
  width?: number | string
  height?: number | string
}>(), {
  alt: '',
  preset: 'content',
  sizes: '',
  densities: '1x 2x',
  format: 'webp',
  quality: 84,
  optimize: true,
  loading: 'lazy',
  fetchpriority: 'auto',
  decoding: 'async',
})

const runtimeConfig = useRuntimeConfig()

const mediaContext = computed(() => createStorefrontMediaContext(
  runtimeConfig,
  import.meta.client ? window.location.origin : '',
))
const normalizedSource = computed(() => normalizeStorefrontMediaUrl(props.src, mediaContext.value))
const optimizationFailed = ref(false)

const parseUrl = (value: string, base?: string) => {
  try {
    return new URL(value, base)
  } catch {
    return null
  }
}

const truthyConfig = (value: unknown) => (
  value === true ||
  ['1', 'true', 'yes', 'on'].includes(String(value || '').trim().toLowerCase())
)

const siteOrigin = computed(() => {
  const configuredSiteUrl = String((runtimeConfig.public as { siteUrl?: string }).siteUrl || '').trim()
  const siteUrl = parseUrl(configuredSiteUrl)
  if (siteUrl) return siteUrl.origin

  const configuredApiBase = String((runtimeConfig.public as { apiBase?: string }).apiBase || '').trim()
  const apiUrl = parseUrl(configuredApiBase)
  if (apiUrl) return apiUrl.origin

  if (import.meta.client) return window.location.origin
  return ''
})

const fallbackSource = computed(() => {
  const source = normalizedSource.value
  if (!source) return ''

  if (/^(?:data|blob):/i.test(source)) return source
  if (!/^\/?uploads\//i.test(source)) return source

  const origin = siteOrigin.value
  if (!origin) return source

  return new URL(source.replace(/^\/?/, '/'), origin).toString()
})

const configuredOrigins = computed(() => {
  const config = runtimeConfig.public as {
    apiBase?: string
    siteUrl?: string
  }

  return new Set(
    [config.siteUrl, config.apiBase, siteOrigin.value]
      .map(value => parseUrl(String(value || ''))?.origin || '')
      .filter(Boolean)
  )
})

const uploadPath = computed(() => {
  const source = normalizedSource.value
  if (/^\/?uploads\//i.test(source)) return `/${source.replace(/^\/+/, '')}`

  const url = parseUrl(source)
  if (!url || !/^\/uploads\//i.test(url.pathname) || !configuredOrigins.value.has(url.origin)) {
    return ''
  }

  return `${url.pathname}${url.search}`
})

const uploadOptimizationEnabled = computed(() => {
  const config = runtimeConfig.public as {
    imageUploadAliasEnabled?: boolean
    imageUploadOptimizationEnabled?: boolean
  }

  return truthyConfig(config.imageUploadOptimizationEnabled ?? config.imageUploadAliasEnabled)
})

const transformSource = computed(() => {
  if (uploadOptimizationEnabled.value && uploadPath.value) return uploadPath.value
  return fallbackSource.value
})

const allowedHosts = computed(() => {
  const config = runtimeConfig.public as {
    apiBase?: string
    imageDomains?: string
    siteUrl?: string
  }
  const configuredDomains = String(config.imageDomains || '')
    .split(',')
    .map(value => value.trim())
    .filter(Boolean)
    .map((value) => {
      const url = parseUrl(value.includes('://') ? value : `https://${value}`)
      return url?.host || value.replace(/^https?:\/\//i, '').replace(/\/.*$/, '')
    })

  const configuredHosts = [config.siteUrl, config.apiBase]
    .map(value => parseUrl(String(value || ''))?.host || '')

  return new Set([...configuredDomains, ...configuredHosts].filter(Boolean))
})

const canOptimize = computed(() => {
  const source = transformSource.value
  if (!source || /^(?:data|blob):/i.test(source)) return false
  if (uploadPath.value && !uploadOptimizationEnabled.value) return false
  if (source.startsWith('/')) return true

  const url = parseUrl(source)
  return Boolean(url && /^https?:$/i.test(url.protocol) && allowedHosts.value.has(url.host))
})

const resolvedSizes = computed(() => props.sizes || PRESET_SIZES[props.preset])

const handleOptimizationError = (event: unknown) => {
  optimizationFailed.value = true
  emit('error', event)
}

const handleFallbackError = (event: unknown) => {
  emit('error', event)
}

watch(transformSource, () => {
  optimizationFailed.value = false
})
</script>

<style scoped>
.tz-storefront-image {
  display: block;
  background-color: var(--tz-image-loading-surface, #0b0b0e);
}
</style>
