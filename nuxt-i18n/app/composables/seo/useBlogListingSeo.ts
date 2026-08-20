import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import {
  useHead,
  useLocalePath,
  useRequestURL,
  useRuntimeConfig,
} from '#imports'
import { useSiteTitle } from '~/composables/useSiteTitle'
import type { BlogCategory, BlogPostSummary } from '~/utils/blog/types'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import { toAbsoluteSeoUrl } from '~/utils/seo/urls'

interface UseBlogListingSeoOptions {
  category: BlogCategory
  title: MaybeRefOrGetter<string>
  description: MaybeRefOrGetter<string>
  posts: MaybeRefOrGetter<BlogPostSummary[]>
}

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, '')

const cleanText = (value: unknown) => String(value || '').replace(/\s+/g, ' ').trim()

export const useBlogListingSeo = (options: UseBlogListingSeoOptions) => {
  const config = useRuntimeConfig()
  const requestUrl = useRequestURL()
  const localePath = useLocalePath()
  const { siteTitle } = useSiteTitle()

  const siteOrigin = computed(() => {
    const configured = String((config.public as { siteUrl?: string }).siteUrl || '').trim()
    return trimTrailingSlash(configured || requestUrl.origin)
  })

  const listingPath = computed(() => `/blog/${options.category}`)
  const listingUrl = computed(() => {
    return toAbsoluteSeoUrl(siteOrigin.value, localePath(listingPath.value))
  })
  const blogPath = computed(() => localePath('/blog'))
  const blogUrl = computed(() => toAbsoluteSeoUrl(siteOrigin.value, blogPath.value))
  const title = computed(() => cleanText(toValue(options.title)))
  const description = computed(() => cleanText(toValue(options.description)))
  const posts = computed(() => toValue(options.posts) || [])

  const collectionSchema = computed(() => {
    const itemListElement: Array<Record<string, unknown>> = []
    posts.value.forEach((post, index) => {
      const slug = cleanText(post.slug)
      const headline = cleanText(post.title)
      if (!slug || !headline) return

      const articleUrl = toAbsoluteSeoUrl(
        siteOrigin.value,
        localePath(`/blog/${options.category}/${slug}`),
      )
      const imageUrl = cleanText(post.featuredImage?.url)
      const article = {
        '@type': 'Article',
        '@id': `${articleUrl}#article`,
        headline,
        url: articleUrl,
        description: cleanText(post.excerpt),
        ...(cleanText(post.date) ? { datePublished: cleanText(post.date) } : {}),
        ...(imageUrl
          ? { image: toAbsoluteSeoUrl(siteOrigin.value, imageUrl) }
          : {}),
        author: {
          '@type': 'Organization',
          name: siteTitle.value,
          url: siteOrigin.value,
        },
        publisher: {
          '@type': 'Organization',
          name: siteTitle.value,
          url: siteOrigin.value,
        },
      }

      itemListElement.push({
        '@type': 'ListItem',
        position: index + 1,
        url: articleUrl,
        name: headline,
        item: article,
      })
    })

    return {
      '@context': 'https://schema.org',
      '@type': 'CollectionPage',
      '@id': `${listingUrl.value}#collection`,
      url: listingUrl.value,
      name: title.value,
      description: description.value,
      isPartOf: {
        '@type': 'Blog',
        '@id': `${blogUrl.value}#blog`,
        name: `${siteTitle.value} Blog`,
        url: blogUrl.value,
      },
      mainEntity: {
        '@type': 'ItemList',
        itemListOrder: 'https://schema.org/ItemListOrderDescending',
        numberOfItems: itemListElement.length,
        itemListElement,
      },
    }
  })

  useHead(() => ({
    script: [createSeoJsonLdScript(collectionSchema.value)],
  }))
}
