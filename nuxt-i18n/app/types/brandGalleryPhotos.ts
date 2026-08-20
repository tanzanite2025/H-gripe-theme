export interface PictureWarehouseProductLink {
  product_id: number
  name?: string
  slug?: string
  locale?: string
}

export interface BrandGalleryImage {
  id: string
  url: string
  thumbnail?: string
  title?: string
  description?: string
  alt?: string
  tags?: string
  order?: number
}

export interface BrandGalleryPhoto {
  id: string
  kind: 'brand'
  title: string
  region: string
  nickname?: string
  slug?: string
  description?: string
  coverImage: string
  galleryId: string
  galleryDetailsLoaded: boolean
  galleryDetailsLoading: boolean
  galleryImages: string[]
  galleryImageItems: BrandGalleryImage[]
  productLinks: PictureWarehouseProductLink[]
}
