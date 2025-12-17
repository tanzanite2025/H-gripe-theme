import { computed } from 'vue'
import { useRuntimeConfig } from '#imports'
import type { BlogCategory, BlogPostDetail, BlogPostSummary } from '~/utils/blogMock'

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

  const listPosts = async (params: {
    lang: string
    category?: BlogCategory
    page: number
    perPage: number
  }): Promise<BlogPostsResponse> => {
    const base = wpApiBase.value

    return await $fetch<BlogPostsResponse>(`${base}/tanzanite/v1/posts`, {
      params: {
        lang: params.lang,
        category: params.category,
        page: params.page,
        per_page: params.perPage,
      },
      credentials: 'include',
    })
  }

  const getPost = async (params: { lang: string; slug: string }): Promise<BlogPostDetail> => {
    const base = wpApiBase.value

    return await $fetch<BlogPostDetail>(`${base}/tanzanite/v1/post`, {
      params: {
        lang: params.lang,
        slug: params.slug,
      },
      credentials: 'include',
    })
  }

  const getTranslations = async (params: {
    group: string
  }): Promise<{ group: string; translations: Record<string, { id: number; slug: string }> }> => {
    const base = wpApiBase.value

    return await $fetch<{ group: string; translations: Record<string, { id: number; slug: string }> }>(
      `${base}/tanzanite/v1/translations`,
      {
        params: {
          group: params.group,
        },
        credentials: 'include',
      }
    )
  }

  return {
    wpApiBase,
    listPosts,
    getPost,
    getTranslations,
  }
}
