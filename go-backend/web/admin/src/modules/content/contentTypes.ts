export type ContentPostId = number | string
export type ContentDialogMode = 'create' | 'edit'
export type ContentStatus = 'draft' | 'published' | 'archived'
export type ContentBadgeTone = 'gray' | 'green' | 'amber'
export type ContentSelectionState = boolean | 'indeterminate'
export type ContentConfirmationType = '' | 'status' | 'delete' | 'batch-status' | 'batch-delete'

export type ContentLabelResolver = (value?: string | null) => string
export type ContentToneResolver = (status?: string | null) => ContentBadgeTone
export type ContentDateFormatter = (value?: string | null) => string

export interface ContentFilters {
  search: string
  status: string
  locale: string
}

export interface ContentPagination {
  page: number
  pageSize: number
  total: number
}

export interface ContentPost {
  id: ContentPostId
  title?: string | null
  slug?: string | null
  content?: string | null
  excerpt?: string | null
  status?: string | null
  locale?: string | null
  featured_image?: string | null
  tags?: string | null
  translation_group_id?: number | string | null
  view_count?: number | string | null
  created_at?: string | null
}

export interface ContentPostForm {
  id: ContentPostId | null
  title: string
  slug: string
  content: string
  excerpt: string
  status: ContentStatus
  locale: string
  featured_image: string
  tags: string
  translation_group_id: number | string | null
}

export interface ContentPostPayload {
  title: string
  slug: string
  content: string
  excerpt: string
  status: ContentStatus
  locale: string
  featured_image: string
  tags: string
  translation_group_id: number | string | null
}

export interface ContentStats {
  total?: number
  published?: number
  draft?: number
  total_views?: number | string
}

export interface ContentListResponse {
  posts?: ContentPost[]
  pagination?: {
    total?: number
  }
}

export interface ContentTranslationsResponse {
  translations?: ContentPost[]
}

export interface ContentConfirmation {
  open: boolean
  type: ContentConfirmationType
  target: ContentPost | ContentPost[] | null
  status: ContentStatus | ''
  title: string
  description: string
  confirmLabel: string
  destructive: boolean
}

export type ContentFormErrors = Record<string, string>
