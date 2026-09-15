import axios from '@/utils/axios'
import type {
  SEOArticleResource,
  SEOCategoryResource,
  SEOProductResource,
  SEOResourceEditorValues,
  SEOResourceList,
  SEOResourceListParams,
} from './types'

const emptyPagination = (params: SEOResourceListParams) => ({
  page: params.page,
  page_size: params.page_size,
  total: 0,
  total_pages: 0,
})

const normalizeList = <T>(
  payload: any,
  params: SEOResourceListParams,
): SEOResourceList<T> => ({
  items: Array.isArray(payload?.items) ? payload.items : [],
  pagination: {
    ...emptyPagination(params),
    ...(payload?.pagination || {}),
  },
})

const normalizeUpdated = <T>(payload: any): T => payload?.data ?? payload ?? {}

export const seoArticlesApi = {
  async list(params: SEOResourceListParams): Promise<SEOResourceList<SEOArticleResource>> {
    const response = await axios.get('/api/admin/seo/articles', { params })
    return normalizeList<SEOArticleResource>(response.data, params)
  },

  async update(id: number | string, payload: SEOResourceEditorValues): Promise<SEOArticleResource> {
    const response = await axios.put(`/api/admin/seo/articles/${id}`, {
      meta_title: payload.meta_title,
      meta_description: payload.meta_description,
      canonical_url: payload.canonical_url,
    })
    return normalizeUpdated<SEOArticleResource>(response.data)
  },
}

export const seoCategoriesApi = {
  async list(params: SEOResourceListParams): Promise<SEOResourceList<SEOCategoryResource>> {
    const response = await axios.get('/api/admin/seo/categories', { params })
    return normalizeList<SEOCategoryResource>(response.data, params)
  },

  async update(
    id: number | string,
    payload: SEOResourceEditorValues,
    locale?: string,
  ): Promise<SEOCategoryResource> {
    const response = await axios.put(`/api/admin/seo/categories/${id}`, {
      locale: locale || undefined,
      meta_title: payload.meta_title,
      meta_description: payload.meta_description,
      intro: payload.intro,
    })
    return normalizeUpdated<SEOCategoryResource>(response.data)
  },
}

export const seoProductsApi = {
  async list(params: SEOResourceListParams): Promise<SEOResourceList<SEOProductResource>> {
    const response = await axios.get('/api/admin/seo/products', { params })
    return normalizeList<SEOProductResource>(response.data, params)
  },

  async update(id: number | string, payload: SEOResourceEditorValues): Promise<SEOProductResource> {
    const response = await axios.put(`/api/admin/seo/products/${id}`, {
      meta_title: payload.meta_title,
      meta_description: payload.meta_description,
    })
    return normalizeUpdated<SEOProductResource>(response.data)
  },
}
