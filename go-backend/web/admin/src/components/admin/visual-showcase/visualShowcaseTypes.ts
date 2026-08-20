export const HOME_HERO_VISUAL_SHOWCASE_KEY = 'home-hero'
export const HOME_HERO_VISUAL_SHOWCASE_REQUIRED_ITEM_COUNT = 9
export const HOME_HERO_VISUAL_SHOWCASE_MOBILE_ITEM_COUNT = 8
export const HOME_MAIN_PRODUCT_CATEGORIES_SHOWCASE_KEY = 'home-main-product-categories'
export const HOME_MAIN_PRODUCT_CATEGORIES_REQUIRED_ITEM_COUNT = 6

export type VisualShowcaseLayoutVariant = 'standard' | 'offset' | 'wide'

export interface VisualShowcaseAdministrationItemApiRecord {
  id?: number | string
  showcase_key?: string
  locale?: string
  image_url?: string | null
  thumbnail_url?: string | null
  storage_key?: string | null
  title?: string | null
  caption?: string | null
  alt_text?: string | null
  desktop_order?: number | string | null
  mobile_pair_index?: number | string | null
  target_url?: string | null
  target_label?: string | null
  layout_variant?: string | null
  is_published?: boolean | null
  width?: number | string | null
  height?: number | string | null
  created_at?: string
  updated_at?: string
}

export interface VisualShowcaseAdministrationResponse {
  showcase_key: string
  locale: string
  items: VisualShowcaseAdministrationItemApiRecord[]
}

export interface VisualShowcaseAdministrationImageUploadResponse {
  image_url: string
  thumbnail_url: string
  storage_key: string
  width: number
  height: number
  filename?: string
  original_filename?: string
}

export interface VisualShowcaseAdministrationItemFormState {
  client_id: string
  id?: number
  image_url: string
  thumbnail_url: string
  storage_key: string
  title: string
  caption: string
  alt_text: string
  desktop_order: number
  mobile_pair_index: number
  target_url: string
  target_label: string
  layout_variant: VisualShowcaseLayoutVariant
  is_published: boolean
  width: number
  height: number
}

export interface VisualShowcaseAdministrationItemSavePayload {
  image_url: string
  thumbnail_url: string
  storage_key: string
  title: string
  caption: string
  alt_text: string
  desktop_order: number
  mobile_pair_index: number
  target_url: string
  target_label: string
  layout_variant: VisualShowcaseLayoutVariant
  is_published: boolean
  width: number
  height: number
}

export interface VisualShowcaseAdministrationReplacePayload {
  locale: string
  items: VisualShowcaseAdministrationItemSavePayload[]
}

export interface VisualShowcaseAdministrationUploadRequest {
  index: number
  file: File
}
