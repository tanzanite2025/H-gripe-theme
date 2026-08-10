import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import {
  useHead,
  useI18n,
  useLocalePath,
  useRequestURL,
  useRuntimeConfig,
} from '#imports'
import {
  type StorefrontSeoAlternateLinkEntry,
  useStorefrontSeoRouteOverride,
} from '~/composables/seo/useStorefrontSeoLinks'
import type { BlogCategory, BlogPostDetail } from '~/utils/blog/types'
import { buildBlogPath } from '~/utils/seo/blog'
import { toAbsoluteSeoUrl } from '~/utils/seo/urls'

interface UseBlogPostSeoOptions {
  post: MaybeRefOrGetter<BlogPostDetail | null | undefined>
  category: MaybeRefOrGetter<BlogCategory | null>
}

const trimTrailingSlash = (value: string): string => value.replace(/\/+$/, '')

const cleanText = (value: unknown): string => String(value || '').replace(/\s+/g, ' ').trim()

const stripHtml = (value: unknown): string => cleanText(
  String(value || '').replace(/<[^>]*>/g, ' '),
)

export const useBlogPostSeo = (options: UseBlogPostSeoOptions) => {
  const config = useRuntimeConfig()
  const requestUrl = useRequestURL()
  const { locale, t } = useI18n()
  const localePath = useLocalePath()

  const siteOrigin = computed(() => {
    const configured = String((config.public as { siteUrl?: string }).siteUrl || '').trim()
    return trimTrailingSlash(configured || requestUrl.origin)
  })

  const post = computed(() => toValue(options.post) || null)
  const category = computed(() => toValue(options.category))
  const currentSlug = computed(() => post.value?.slug || '')
  const currentPath = computed(() => localePath(buildBlogPath(category.value, currentSlug.value)))
  const canonicalUrl = computed(() => {
    const explicitCanonical = cleanText(post.value?.canonicalUrl)
    return explicitCanonical
      ? toAbsoluteSeoUrl(siteOrigin.value, explicitCanonical)
      : toAbsoluteSeoUrl(siteOrigin.value, currentPath.value)
  })

  const alternateRoutes = computed<StorefrontSeoAlternateLinkEntry[]>(() => {
    const routes = (post.value?.localizedRoutes || [])
      .map((route) => {
        const code = cleanText(route.locale)
        const path = cleanText(route.path)
        if (!code || !path) return null
        return { code, path }
      })
      .filter((route): route is StorefrontSeoAlternateLinkEntry => Boolean(route))

    if (routes.length) return routes
    if (!currentSlug.value) return []

    return [{
      code: cleanText(locale.value) || 'en',
      path: currentPath.value,
    }]
  })

  useStorefrontSeoRouteOverride(alternateRoutes)

  const title = computed(() => {
    return cleanText(post.value?.metaTitle) || cleanText(post.value?.title) || t('blog.pages.detail.metaTitle')
  })
  const description = computed(() => {
    return stripHtml(post.value?.metaDescription)
      || stripHtml(post.value?.excerpt)
      || t('blog.pages.detail.metaDescription')
  })

  useHead(() => ({
    title: title.value,
    meta: [
      { name: 'description', content: description.value, key: 'description' },
      { property: 'og:type', content: 'article', key: 'og:type' },
      { property: 'og:title', content: title.value, key: 'og:title' },
      { property: 'og:description', content: description.value, key: 'og:description' },
      { property: 'og:url', content: canonicalUrl.value, key: 'og:url' },
    ],
    link: [
      { rel: 'canonical', href: canonicalUrl.value, key: 'canonical' },
    ],
  }))

  return {
    canonicalUrl,
    alternateRoutes,
  }
}
