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

export const useBlogApi = () => {
  const config = useRuntimeConfig()

  const wpApiBase = computed(() => {
    const base = (config.public as { wpApiBase?: string }).wpApiBase || '/wp-json'
    return base.replace(/\/$/, '')
  })

  const blogApiMode = computed(() => {
    return String((config.public as { blogApiMode?: string }).blogApiMode || 'auto').toLowerCase()
  })

  const useLocalBlog = computed(() => {
    if (['local', 'mock', 'disabled'].includes(blogApiMode.value)) return true
    return import.meta.server && wpApiBase.value.startsWith('/')
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

    const base = wpApiBase.value

    try {
      return await $fetch<BlogPostsResponse>(`${base}/tanzanite/v1/posts`, {
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

    const base = wpApiBase.value

    try {
      return await $fetch<BlogPostDetail>(`${base}/tanzanite/v1/post`, {
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
  }): Promise<{ group: string; translations: Record<string, { id: number; slug: string }> }> => {
    const localTranslations = () => getBlogTranslationsByGroup(params.group)
    if (useLocalBlog.value) return localTranslations()

    const base = wpApiBase.value

    try {
      return await $fetch<{ group: string; translations: Record<string, { id: number; slug: string }> }>(
        `${base}/tanzanite/v1/translations`,
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
    wpApiBase,
    listPosts,
    getPost,
    getTranslations,
  }
}
