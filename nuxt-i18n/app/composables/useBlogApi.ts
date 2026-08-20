import { computed } from 'vue'
import { useRuntimeConfig } from '#imports'
import {
  getBlogPostBySlug,
  getBlogTranslationsByGroup,
  listBlogPosts,
} from '~/utils/blogMock'
import {
  normalizeBlogLocalizedRoutes,
  resolveBlogCategory,
} from '~/utils/seo/blog'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'
import type {
  BlogCategory,
  BlogPostDetail,
  BlogPostSummary,
} from '~/utils/blog/types'

type BlogPostsResponse = {
  page: number
  per_page: number
  total: number
  items: BlogPostSummary[]
}

type BlogTranslationsResponse = {
  group: string
  translations: Record<string, { id: number; slug: string }>
}

const trimTrailingSlash = (value: string) => value.replace(/\/$/, '')

const resolveApiV1Base = (apiBase: string, internalOrigin: string, isServer: boolean) => {
  const normalized = trimTrailingSlash(apiBase || '/api/v1')
  if (isServer && normalized.startsWith('/')) {
    return `${trimTrailingSlash(internalOrigin || 'http://localhost:9200')}${normalized}`
  }
  if (normalized === '/') return '/api/v1'
  return normalized.endsWith('/api/v1') ? normalized : `${normalized}/api/v1`
}

export const useBlogApi = () => {
  const config = useRuntimeConfig()
  const mediaContext = createStorefrontMediaContext(config)
  const internalApiOrigin = import.meta.server
    ? String((config as { apiInternalOrigin?: string }).apiInternalOrigin || '')
    : ''

  const apiBase = computed(() => {
    return (config.public as { apiBase?: string }).apiBase || '/api/v1'
  })

  const blogApiMode = computed(() => {
    return String((config.public as { blogApiMode?: string }).blogApiMode || 'auto').toLowerCase()
  })

  const useLocalBlog = computed(() => {
    return ['local', 'mock', 'disabled'].includes(blogApiMode.value)
  })

  const normalizePostMedia = <T extends BlogPostSummary | BlogPostDetail>(post: T): T => {
    if (!post.featuredImage?.url) return post

    return {
      ...post,
      featuredImage: {
        ...post.featuredImage,
        url: normalizeStorefrontMediaUrl(post.featuredImage.url, mediaContext),
      },
    }
  }

  const apiRoot = computed(() => resolveApiV1Base(
    apiBase.value,
    internalApiOrigin,
    import.meta.server,
  ))

  const mapPost = (item: any, fallbackLocale: string): BlogPostSummary => {
    const categories = item.tags
      ? String(item.tags).split(',').map((tag: string) => tag.trim()).filter(Boolean)
      : []

    return {
      id: item.id,
      lang: item.locale || fallbackLocale,
      group: item.translation_group_id ? `grp-${item.translation_group_id}` : '',
      slug: item.slug,
      title: item.title,
      excerpt: item.excerpt,
      metaTitle: item.meta_title || '',
      metaDescription: item.meta_description || '',
      date: item.published_at || item.created_at,
      featuredImage: item.featured_image ? { url: item.featured_image } : null,
      categories: categories as BlogCategory[],
      translations: {},
      localizedRoutes: normalizeBlogLocalizedRoutes(item.localized_routes),
    }
  }

  const buildLocalPostsResponse = (params: {
    lang: string
    category?: BlogCategory
    page: number
    perPage: number
  }): BlogPostsResponse => {
    const allItems = listBlogPosts({ lang: params.lang, category: params.category })
    const start = Math.max(params.page - 1, 0) * params.perPage

    return {
      page: params.page,
      per_page: params.perPage,
      total: allItems.length,
      items: allItems.slice(start, start + params.perPage),
    }
  }

  const listPosts = async (params: {
    lang: string
    category?: BlogCategory
    page: number
    perPage: number
  }): Promise<BlogPostsResponse> => {
    const localResponse = () => buildLocalPostsResponse(params)
    if (useLocalBlog.value) {
      const local = localResponse()
      return {
        ...local,
        items: local.items.map(normalizePostMedia),
      }
    }

    const response = await $fetch<{ data: BlogPostSummary[], total: number }>(
      `${apiRoot.value}/content/posts`,
      {
        params: {
          locale: params.lang,
          category: params.category,
          page: params.page,
          page_size: params.perPage,
          status: 'published',
        },
      },
    )

    if (!Array.isArray(response.data)) {
      throw new Error('Blog posts response data is invalid')
    }

    return {
      page: params.page,
      per_page: params.perPage,
      total: response.total || 0,
      items: response.data.map((item: any) => normalizePostMedia(mapPost(item, params.lang))),
    }
  }

  const getPost = async (params: { lang: string; slug: string }): Promise<BlogPostDetail> => {
    const localPost = () => getBlogPostBySlug(params)
    if (useLocalBlog.value) {
      const post = localPost()
      if (post) return normalizePostMedia(post)
      throw new Error('Blog post not found')
    }

    const response = await $fetch<{ data: BlogPostDetail } | BlogPostDetail>(
      `${apiRoot.value}/content/posts/${encodeURIComponent(params.slug)}`,
      {
        params: {
          locale: params.lang,
        },
      },
    )
    const post = (response as any).data || response
    if (!post || typeof post !== 'object') {
      throw new Error('Blog post response is invalid')
    }

    return normalizePostMedia({
      ...mapPost(post, params.lang),
      contentHtml: post.content || '',
      canonicalUrl: post.canonical_url || '',
    } as BlogPostDetail)
  }

  const getTranslations = async (params: {
    group: string
  }): Promise<BlogTranslationsResponse> => {
    const localTranslations = () => getBlogTranslationsByGroup(params.group)
    if (useLocalBlog.value) return localTranslations()

    // Group translations are not fully supported in Go API yet, returning fallback mock
    return localTranslations()
  }

  const getPostTranslations = async (postId: number): Promise<Record<string, any>> => {
    const response = await $fetch<{ translations: Record<string, any> }>(
      `${apiRoot.value}/i18n/translations/${postId}`,
    )
    if (!response || !response.translations) {
      throw new Error(`Post translations response invalid for postId ${postId}`)
    }
    return response.translations
  }

  return {
    apiBase,
    listPosts,
    getPost,
    getTranslations,
    getPostTranslations,
  }
}
