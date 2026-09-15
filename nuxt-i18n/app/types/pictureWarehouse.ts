import type {
  BrandGalleryPhoto,
  PictureWarehouseProductLink,
} from '~/types/brandGalleryPhotos'

export interface UploadOrderOption {
  id: number
  order_number: string
  status: string
  shipping_status: string
  completed_at?: string
  total_amount: number
  currency: string
  eligible: boolean
}

export interface UploadPreview {
  key: string
  file: File
  url: string
}

export interface PhotoComment {
  id: number
  author: string
  content: string
  dateGmt: string
  dateGmtFormatted: string
  location?: string
}

export interface RiderPhoto {
  id: string
  kind: 'user'
  title: string
  region: string
  nickname?: string
  galleryImages?: string[]
  productLinks?: PictureWarehouseProductLink[]
}

export type PictureWarehousePhoto = RiderPhoto | BrandGalleryPhoto
