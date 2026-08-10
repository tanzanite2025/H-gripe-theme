export type BlogCategory = 'news' | 'wheelsbuild'

export interface BlogFeaturedImage {
  url: string
  width?: number | null
  height?: number | null
  alt?: string
}

export interface BlogLocalizedRoute {
  id?: number
  locale: string
  slug: string
  path: string
}

export interface BlogTranslationsMapEntry {
  id: number
  slug: string
}

export type BlogTranslationsMap = Record<string, BlogTranslationsMapEntry>

export interface BlogPostSummary {
  id: number
  lang: string
  group: string
  slug: string
  title: string
  excerpt: string
  metaTitle?: string
  metaDescription?: string
  date: string
  featuredImage: BlogFeaturedImage | null
  categories: BlogCategory[]
  translations: BlogTranslationsMap
  localizedRoutes: BlogLocalizedRoute[]
}

export interface BlogPostDetail extends BlogPostSummary {
  contentHtml: string
  canonicalUrl: string
}
