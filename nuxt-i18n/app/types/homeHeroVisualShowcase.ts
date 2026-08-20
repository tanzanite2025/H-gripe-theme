export type HomeHeroVisualShowcaseLayoutVariant = 'standard' | 'offset' | 'wide'

export interface HomeHeroVisualShowcaseItem {
  id: string
  showcaseKey: string
  locale: string
  src: string
  altText: string
  title: string
  caption: string
  width: number
  height: number
  desktopOrder: number
  mobilePairIndex: number
  targetUrl?: string
  targetLabel?: string
  layoutVariant: HomeHeroVisualShowcaseLayoutVariant
}

export interface HomeHeroVisualShowcaseApiItem {
  id?: number | string
  showcase_key?: string
  locale?: string
  image_url?: string
  thumbnail_url?: string
  alt_text?: string
  title?: string
  caption?: string
  width?: number | string
  height?: number | string
  desktop_order?: number | string
  mobile_pair_index?: number | string
  target_url?: string
  target_label?: string
  layout_variant?: HomeHeroVisualShowcaseLayoutVariant
}

export interface HomeHeroVisualShowcaseApiEnvelope {
  data?: {
    showcase_key?: string
    locale?: string
    requested_locale?: string
    fallback?: boolean
    configured_count?: number | string
    items?: HomeHeroVisualShowcaseApiItem[]
  }
}
