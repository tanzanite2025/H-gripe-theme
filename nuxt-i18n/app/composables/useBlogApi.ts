import { computed } from 'vue'
import { useRuntimeConfig } from '#imports'
import {
  getBlogPostBySlug,
  getBlogTranslationsByGroup,
  listBlogPosts,
  type BlogCategory,
  type BlogPostDetail,
  type BlogPostSummary,
} from '~/utils/blogMock'

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

const getGoOriginBase = (apiBase: string) => {
  const normalized = trimTrailingSlash(apiBase || '/api/v1')
  return normalized.endsWith('/api/v1')
    ? normalized.slice(0, -'/api/v1'.length)
    : normalized
}

export const useBlogApi = () => {
  const config = useRuntimeConfig()

  const apiBase = computed(() => {
    return (config.public as { apiBase?: string }).apiBase || '/api/v1'
  })

  const blogApiMode = computed(() => {
    return String((config.public as { blogApiMode?: string }).blogApiMode || 'auto').toLowerCase()
  })

  const useLocalBlog = computed(() => {
    if (['local', 'mock', 'disabled'].includes(blogApiMode.value)) return true
    return import.meta.server && apiBase.value.startsWith('/')
  })

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
    if (useLocalBlog.value) return localResponse()

    try {
      const response = await $fetch<{ data: BlogPostSummary[], total: number }>(`${trimTrailingSlash(apiBase.value)}/content/posts`, {
        params: {
          locale: params.lang,
          category: params.category,
          page: params.page,
          page_size: params.perPage,
          status: 'published'
        }
      })
      
      if (!response.data) throw new Error("[CRITICAL] response.data missing");
      return {
        page: params.page,
        per_page: params.perPage,
        total: response.total || 0,
        items: response.data.map((item: any) => ({
          id: item.id,
          lang: item.locale || params.lang,
          group: item.translation_group_id ? `grp-${item.translation_group_id}` : '',
          slug: item.slug,
          title: item.title,
          excerpt: item.excerpt,
          date: item.published_at || item.created_at,
          featuredImage: item.featured_image ? { url: item.featured_image } : null,
          categories: item.tags ? item.tags.split(',') : [],
          translations: {}
        } as BlogPostSummary))
      }
    } catch {
      return localResponse()
    }
  }

  const getPost = async (params: { lang: string; slug: string }): Promise<BlogPostDetail> => {
    const localPost = () => getBlogPostBySlug(params)
    if (useLocalBlog.value) {
      const post = localPost()
      if (post) return post
      throw new Error('Blog post not found')
    }

    try {
      const response = await $fetch<{ data: BlogPostDetail } | BlogPostDetail>(`${trimTrailingSlash(apiBase.value)}/content/posts/${encodeURIComponent(params.slug)}`, {
        params: {
          locale: params.lang,
        }
      })
      // Backend might wrap in data or return directly
      const post = (response as any).data || response
      if (!post) throw new Error('Post not found in response')
        
      return {
        id: post.id,
        lang: post.locale || params.lang,
        group: post.translation_group_id ? `grp-${post.translation_group_id}` : '',
        slug: post.slug,
        title: post.title,
        excerpt: post.excerpt,
        date: post.published_at || post.created_at,
        featuredImage: post.featured_image ? { url: post.featured_image } : null,
        categories: post.tags ? post.tags.split(',') : [],
        translations: {},
        contentHtml: post.content || '',
        canonicalUrl: post.canonical_url || ''
      } as BlogPostDetail
    } catch {
      const post = localPost()
      if (post) return post
      throw new Error('Blog post not found')
    }
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
    const apiBase = (config.public as { apiBase?: string }).apiBase || '/api/v1'
    try {
      const response = await $fetch<{ translations: Record<string, any> }>(
        `${trimTrailingSlash(apiBase)}/api/v1/i18n/translations/${postId}`
      )
      if (!response || !response.translations) {
        throw new Error(`[CRITICAL] Post translations response invalid for postId ${postId}`)
      }
      return response.translations
    } catch (error) {
      console.error('Failed to fetch post translations from Go backend:', error)
      throw error // FAIL LOUDLY: Never return {} on critical fetch error
    }
  }

  return {
    apiBase,
    listPosts,
    getPost,
    getTranslations,
    getPostTranslations,
  }
}
