import { computed } from 'vue'
import {
  createError,
  useAsyncData,
  useI18n,
  useRoute,
} from '#imports'
import { useBlogApi } from '~/composables/useBlogApi'
import { resolveBlogCategory } from '~/utils/seo/blog'
import type { BlogCategory, BlogPostDetail } from '~/utils/blog/types'

interface UseBlogPostDetailOptions {
  category: BlogCategory | null
  keyPrefix: string
}

const notFoundError = () => createError({
  statusCode: 404,
  statusMessage: 'Blog post not found',
})

const isNotFoundError = (error: unknown): boolean => {
  if (!error || typeof error !== 'object') return false

  const value = error as {
    statusCode?: number
    status?: number
    response?: { status?: number }
  }

  return [value.statusCode, value.status, value.response?.status].includes(404)
}

export const useBlogPostDetail = async (
  options: UseBlogPostDetailOptions,
) => {
  const route = useRoute()
  const { locale } = useI18n()
  const blogApi = useBlogApi()

  const slug = computed(() => String(route.params.slug || '').trim())
  const lang = computed(() => String(locale.value || 'en'))
  const dataKey = computed(() => [
    options.keyPrefix,
    lang.value,
    slug.value,
  ].map((value) => encodeURIComponent(value)).join(':'))

  const { data: postData } = await useAsyncData<BlogPostDetail>(
    dataKey,
    async () => {
      if (!slug.value) throw notFoundError()

      let post: BlogPostDetail
      try {
        post = await blogApi.getPost({
          lang: lang.value,
          slug: slug.value,
        })
      } catch (error) {
        if (isNotFoundError(error) || error instanceof Error && error.message === 'Blog post not found') {
          throw notFoundError()
        }
        throw error
      }

      const actualCategory = resolveBlogCategory(post.categories)
      if (actualCategory !== options.category) {
        throw notFoundError()
      }

      return post
    },
    {
      server: true,
      watch: [lang, slug],
    },
  )

  if (!postData.value) throw notFoundError()

  return {
    lang,
    slug,
    post: computed(() => postData.value || null),
  }
}
