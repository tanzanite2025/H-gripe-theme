import { computed } from 'vue'
import { useRuntimeConfig } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import {
  getBlogPostBySlug,
  getBlogTranslationsByGroup,
  listBlogPosts,
} from '~/utils/blogMock'
import {
  normalizeBlogLocalizedRoutes,
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

type BlogCategoriesResponse = {
  data: BlogCategory[]
}

export const useBlogApi = () => {
  const config = useRuntimeConfig()
  const { baseURL, request } = useApiRequest()
  const mediaContext = createStorefrontMediaContext(config)

  const apiBase = computed(() => {
    return baseURL
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

  const mapPost = (item: any, fallbackLocale: string): BlogPostSummary => {
    const categories = Array.isArray(item.categories)
      ? item.categories
        .map((category: any): BlogCategory | null => {
          if (!category || typeof category !== 'object') return null
          const slug = String(category.slug || '').trim()
          const name = String(category.name || '').trim()
          if (!slug || !name) return null
          return {
            id: Number(category.id) || undefined,
            name,
            slug,
            description: String(category.description || ''),
            locale: String(category.locale || fallbackLocale),
            sortOrder: Number(category.sort_order) || 0,
          }
        })
        .filter((category: BlogCategory | null): category is BlogCategory => Boolean(category))
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
      categories,
      translations: {},
      localizedRoutes: normalizeBlogLocalizedRoutes(item.localized_routes),
    }
  }

  const buildLocalPostsResponse = (params: {
    lang: string
    category?: string
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
    category?: string
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

    const response = await request<{ data: BlogPostSummary[], total: number }>(
      '/content/posts',
      {
        params: {
          locale: params.lang,
          category: params.category,
          page: params.page,
          page_size: params.perPage,
          status: 'published',
        },
      },
      'Failed to load blog posts',
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

  const listCategories = async (params: { lang: string }): Promise<BlogCategory[]> => {
    if (useLocalBlog.value) {
      const categories = new Map<string, BlogCategory>()
      for (const post of listBlogPosts({ lang: params.lang })) {
        for (const category of post.categories) {
          categories.set(category.slug, category)
        }
      }
      return Array.from(categories.values()).sort((a, b) => (
        (a.sortOrder || 0) - (b.sortOrder || 0) || a.name.localeCompare(b.name)
      ))
    }

    const response = await request<BlogCategoriesResponse>(
      '/content/blog-categories',
      { params: { locale: params.lang } },
      'Failed to load blog categories',
    )
    if (!Array.isArray(response.data)) {
      throw new Error('Blog categories response data is invalid')
    }
    return response.data
      .map((category: any): BlogCategory | null => {
        if (!category || typeof category !== 'object') return null
        const slug = String(category.slug || '').trim()
        const name = String(category.name || '').trim()
        if (!slug || !name) return null
        return {
          id: Number(category.id) || undefined,
          name,
          slug,
          description: String(category.description || ''),
          locale: String(category.locale || params.lang),
          sortOrder: Number(category.sort_order) || 0,
        }
      })
      .filter((category: BlogCategory | null): category is BlogCategory => Boolean(category))
  }

  const getPost = async (params: { lang: string; slug: string }): Promise<BlogPostDetail> => {
    const localPost = () => getBlogPostBySlug(params)
    if (useLocalBlog.value) {
      const post = localPost()
      if (post) return normalizePostMedia(post)
      throw new Error('Blog post not found')
    }

    const response = await request<{ data: BlogPostDetail } | BlogPostDetail>(
      `/content/posts/${encodeURIComponent(params.slug)}`,
      {
        params: {
          locale: params.lang,
        },
      },
      'Failed to load blog post',
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
    const response = await request<{ translations: Record<string, any> }>(
      `/i18n/translations/${postId}`,
      {},
      'Failed to load blog translations',
    )
    if (!response || !response.translations) {
      throw new Error(`Post translations response invalid for postId ${postId}`)
    }
    return response.translations
  }

  return {
    apiBase,
    listPosts,
    listCategories,
    getPost,
    getTranslations,
    getPostTranslations,
  }
}
