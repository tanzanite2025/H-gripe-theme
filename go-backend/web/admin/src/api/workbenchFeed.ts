import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiBooleanField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiPagination,
  requireApiStringField,
  readApiBody,
} from '@/utils/apiResponse'

export type WorkbenchFeedStatus = 'draft' | 'published' | 'archived'

export interface WorkbenchFeedMedia {
  id: number
  file_path: string
  width: number
  height: number
  file_size_kb: number
  caption: string
  sort_order: number
}

export interface WorkbenchFeedTaggedProduct {
  id: number
  product_id: number
  variant_id?: number | null
  product_slug: string
  display_title: string
  price: number
  currency: string
  direct_action: 'detail_drawer'
  available: boolean
  sort_order: number
}

export interface WorkbenchFeedEntry {
  id: number
  entry_number: string
  published_at: string
  locale: string
  content: string
  tags: string[]
  status: WorkbenchFeedStatus
  media: WorkbenchFeedMedia[]
  tagged_products: WorkbenchFeedTaggedProduct[]
  created_at?: string
  updated_at?: string
}

export interface WorkbenchFeedPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface WorkbenchFeedListResult {
  entries: WorkbenchFeedEntry[]
  pagination: WorkbenchFeedPagination
}

export interface WorkbenchFeedProductOption {
  product_id: number
  variant_id?: number | null
  product_name: string
  product_slug: string
  variant_title: string
  price: number
  currency: string
  available: boolean
}

export interface WorkbenchFeedProductOptionResult {
  options: WorkbenchFeedProductOption[]
  pagination: WorkbenchFeedPagination
}

export interface WorkbenchFeedMediaPayload {
  file_path: string
  width: number
  height: number
  file_size_kb: number
  caption: string
  sort_order: number
}

export interface WorkbenchFeedTaggedProductPayload {
  product_id: number
  variant_id?: number | null
  direct_action: 'detail_drawer'
  sort_order: number
}

export interface WorkbenchFeedEntryPayload {
  published_at?: string
  locale: string
  content: string
  tags: string[]
  status: WorkbenchFeedStatus
  media: WorkbenchFeedMediaPayload[]
  tagged_products: WorkbenchFeedTaggedProductPayload[]
}

const readObject = (response: unknown, endpoint: string): Record<string, any> => {
  const body = requireApiObject(readApiBody(response, endpoint), endpoint, 'response body')
  const nested = body.data
  if (nested && typeof nested === 'object' && !Array.isArray(nested)) {
    return requireApiObject(nested, endpoint, 'response data')
  }
  return body
}

const readPagination = (body: Record<string, any>, endpoint: string) => (
  requireApiPagination(body, body, endpoint)
)

const readMedia = (value: unknown, endpoint: string): WorkbenchFeedMedia => {
  const media = requireApiObject(value, endpoint, 'media') as WorkbenchFeedMedia
  requireApiNumberField(media, 'id', endpoint)
  requireApiStringField(media, 'file_path', endpoint)
  requireApiNumberField(media, 'width', endpoint)
  requireApiNumberField(media, 'height', endpoint)
  requireApiNumberField(media, 'file_size_kb', endpoint)
  requireApiStringField(media, 'caption', endpoint)
  requireApiNumberField(media, 'sort_order', endpoint)
  return media
}

const readTaggedProduct = (value: unknown, endpoint: string): WorkbenchFeedTaggedProduct => {
  const product = requireApiObject(value, endpoint, 'tagged product') as WorkbenchFeedTaggedProduct
  requireApiNumberField(product, 'id', endpoint)
  requireApiNumberField(product, 'product_id', endpoint)
  requireApiStringField(product, 'product_slug', endpoint)
  requireApiStringField(product, 'display_title', endpoint)
  requireApiNumberField(product, 'price', endpoint)
  requireApiStringField(product, 'currency', endpoint)
  requireApiStringField(product, 'direct_action', endpoint)
  requireApiBooleanField(product, 'available', endpoint)
  requireApiNumberField(product, 'sort_order', endpoint)
  return product
}

const readEntry = (value: unknown, endpoint: string): WorkbenchFeedEntry => {
  const entry = requireApiObject(value, endpoint, 'entry') as WorkbenchFeedEntry
  requireApiNumberField(entry, 'id', endpoint)
  requireApiStringField(entry, 'entry_number', endpoint)
  requireApiStringField(entry, 'published_at', endpoint)
  requireApiStringField(entry, 'locale', endpoint)
  requireApiStringField(entry, 'content', endpoint)
  requireApiStringField(entry, 'status', endpoint)
  const tags = requireApiArrayField<unknown>(entry, 'tags', endpoint).map(item => String(item))
  const media = requireApiArrayField(entry, 'media', endpoint).map(item => readMedia(item, endpoint))
  const taggedProducts = requireApiArrayField(entry, 'tagged_products', endpoint)
    .map(item => readTaggedProduct(item, endpoint))
  return { ...entry, tags, media, tagged_products: taggedProducts }
}

const readProductOption = (value: unknown, endpoint: string): WorkbenchFeedProductOption => {
  const option = requireApiObject(value, endpoint, 'product option') as WorkbenchFeedProductOption
  requireApiNumberField(option, 'product_id', endpoint)
  requireApiStringField(option, 'product_name', endpoint)
  requireApiStringField(option, 'product_slug', endpoint)
  requireApiStringField(option, 'variant_title', endpoint)
  requireApiNumberField(option, 'price', endpoint)
  requireApiStringField(option, 'currency', endpoint)
  requireApiBooleanField(option, 'available', endpoint)
  return option
}

export const workbenchFeedApi = {
  async list(params: Record<string, unknown> = {}): Promise<WorkbenchFeedListResult> {
    const endpoint = '/api/admin/workbench-feed/entries'
    const body = readObject(await axios.get(endpoint, { params }), endpoint)
    return {
      entries: requireApiArrayField(body, 'entries', endpoint).map(entry => readEntry(entry, endpoint)),
      pagination: readPagination(body, endpoint),
    }
  },

  async get(id: number | string): Promise<WorkbenchFeedEntry> {
    const endpoint = `/api/admin/workbench-feed/entries/${id}`
    const body = readObject(await axios.get(endpoint), endpoint)
    return readEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async create(payload: WorkbenchFeedEntryPayload): Promise<WorkbenchFeedEntry> {
    const endpoint = '/api/admin/workbench-feed/entries'
    const body = readObject(await axios.post(endpoint, payload), endpoint)
    return readEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async update(id: number | string, payload: WorkbenchFeedEntryPayload): Promise<WorkbenchFeedEntry> {
    const endpoint = `/api/admin/workbench-feed/entries/${id}`
    const body = readObject(await axios.put(endpoint, payload), endpoint)
    return readEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async updateStatus(id: number | string, status: WorkbenchFeedStatus): Promise<WorkbenchFeedEntry> {
    const endpoint = `/api/admin/workbench-feed/entries/${id}/status`
    const body = readObject(await axios.patch(endpoint, { status }), endpoint)
    return readEntry(requireApiObjectField(body, 'entry', endpoint), endpoint)
  },

  async remove(id: number | string) {
    const endpoint = `/api/admin/workbench-feed/entries/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },

  async uploadMedia(file: File, caption = ''): Promise<WorkbenchFeedMedia> {
    const endpoint = '/api/admin/workbench-feed/media'
    const formData = new FormData()
    formData.append('file', file)
    formData.append('caption', caption)
    const body = readObject(await axios.post(endpoint, formData), endpoint)
    return readMedia(requireApiObjectField(body, 'media', endpoint), endpoint)
  },

  async listProductOptions(params: Record<string, unknown> = {}): Promise<WorkbenchFeedProductOptionResult> {
    const endpoint = '/api/admin/workbench-feed/product-options'
    const body = readObject(await axios.get(endpoint, { params }), endpoint)
    return {
      options: requireApiArrayField(body, 'options', endpoint).map(option => readProductOption(option, endpoint)),
      pagination: readPagination(body, endpoint),
    }
  },
}

export default workbenchFeedApi
