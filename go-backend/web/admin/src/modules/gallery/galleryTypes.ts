export type GalleryId = number | string
export type GalleryDialogMode = 'create' | 'edit'
export type GallerySelectionState = boolean | 'indeterminate'
export type GalleryConfirmationType = '' | 'delete-gallery' | 'delete-image' | 'batch-delete-images'

export interface GalleryImage {
  id: GalleryId
  media_asset_id?: GalleryId | null
  url?: string | null
  thumbnail?: string | null
  title?: string | null
  description?: string | null
  tags?: string | null
  order?: number | string | null
  sort_order?: number | string | null
}

export interface GalleryProductLink {
  product_id: GalleryId
  name?: string | null
  slug?: string | null
  locale?: string | null
}

export interface GalleryRecord {
  id: GalleryId
  name?: string | null
  title?: string | null
  slug?: string | null
  description?: string | null
  cover_image?: string | null
  images?: GalleryImage[] | null
  product_links?: GalleryProductLink[] | null
  image_count?: number | string | null
  images_count?: number | string | null
  created_at?: string | null
}

export interface GalleryPagination {
  page: number
  pageSize: number
  total: number
}

export interface GalleryForm {
  id: GalleryId | null
  title: string
  slug: string
  description: string
  product_links: GalleryProductLink[]
  images: GalleryImageForm[]
}

export interface GalleryImageForm {
  id: GalleryId | null
  media_asset_id: GalleryId | null
  url: string
  thumbnail: string
  title: string
  description: string
  tags: string
  order: number
}

export interface GalleryPayload {
  title: string
  slug: string
  description: string
  product_ids: number[]
  images?: GalleryImagePayload[]
}

export interface GalleryImagePayload {
  media_asset_id: number | null
  title: string
  description: string
  tags: string
  order: number
}

export interface GalleryPreviewImage {
  url: string
  title: string
}

export interface GalleryListResponse {
  galleries?: GalleryRecord[]
  pagination?: {
    total?: number
  }
  total?: number
}

export interface GalleryDetailResponse {
  gallery?: GalleryRecord
}

export interface GalleryImagesResponse {
  images?: GalleryImage[]
}

export interface GalleryConfirmation {
  open: boolean
  type: GalleryConfirmationType
  target: GalleryRecord | GalleryImage | GalleryImage[] | null
  title: string
  description: string
  confirmLabel: string
}

export type GalleryFormErrors = Record<string, string>
export type GalleryTitleResolver = (gallery?: GalleryRecord | null) => string
export type GalleryCoverResolver = (gallery?: GalleryRecord | null) => string
export type GalleryImageCountResolver = (gallery?: GalleryRecord | null) => number | string
export type GalleryDateFormatter = (value?: string | null) => string
