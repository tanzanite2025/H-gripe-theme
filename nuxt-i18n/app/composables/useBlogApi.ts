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

  const blogCompatBase = computed(() => {
    const apiBase = (config.public as { apiBase?: string }).apiBase || '/api/v1'
    return `${getGoOriginBase(apiBase)}/wp-json/tanzanite/v1`
  })

  const blogApiMode = computed(() => {
    return String((config.public as { blogApiMode?: string }).blogApiMode || 'auto').toLowerCase()
  })

  const useLocalBlog = computed(() => {
    if (['local', 'mock', 'disabled'].includes(blogApiMode.value)) return true
    return import.meta.server && blogCompatBase.value.startsWith('/')
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
      return await $fetch<BlogPostsResponse>(`${blogCompatBase.value}/posts`, {
        params: {
          lang: params.lang,
          category: params.category,
          page: params.page,
          per_page: params.perPage,
        },
        credentials: 'include',
      })
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
      return await $fetch<BlogPostDetail>(`${blogCompatBase.value}/post`, {
        params: {
          lang: params.lang,
          slug: params.slug,
        },
        credentials: 'include',
      })
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

    try {
      return await $fetch<BlogTranslationsResponse>(
        `${blogCompatBase.value}/translations`,
        {
          params: {
            group: params.group,
          },
          credentials: 'include',
        }
      )
    } catch {
      return localTranslations()
    }
  }

  return {
    blogCompatBase,
    listPosts,
    getPost,
    getTranslations,
  }
}
