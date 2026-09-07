import { buildBlogPath, buildLocalizedBlogPath } from '~/utils/seo/blog'
import type {
  BlogCategory,
  BlogFeaturedImage,
  BlogLocalizedRoute,
  BlogPostDetail,
  BlogPostSummary,
  BlogTranslationsMap,
} from '~/utils/blog/types'

export type {
  BlogCategory,
  BlogFeaturedImage,
  BlogLocalizedRoute,
  BlogPostDetail,
  BlogPostSummary,
  BlogTranslationsMap,
} from '~/utils/blog/types'

const posts: BlogPostDetail[] = []

const buildTranslations = (): BlogTranslationsMap => ({})

const buildLocalizedRoutes = (
  items: Array<{ lang: string; id: number; slug: string }>,
): BlogLocalizedRoute[] => items.map((entry) => ({
  id: entry.id,
  locale: entry.lang,
  slug: entry.slug,
  path: buildLocalizedBlogPath(entry.lang, entry.slug),
}))

export const listBlogPosts = (params: {
  lang: string
  category?: string
}): BlogPostSummary[] => {
  return posts
    .filter((post) => post.lang === params.lang)
    .filter((post) => !params.category || post.categories.some((category) => category.slug === params.category))
    .map(({ contentHtml, canonicalUrl, ...summary }) => summary)
}

export const getBlogPostBySlug = (params: {
  lang: string
  slug: string
}): BlogPostDetail | null => {
  return posts.find((post) => post.lang === params.lang && post.slug === params.slug) || null
}

export const getBlogTranslationsByGroup = (
  group: string,
): { group: string; translations: BlogTranslationsMap } => ({
  group,
  translations: buildTranslations(),
})

export const buildBlogDetailPath = (post: BlogPostSummary): string => {
  return buildBlogPath(post.slug)
}

export const getBlogCategories = (lang: string): BlogCategory[] => {
  const categories = new Map<string, BlogCategory>()
  for (const post of posts.filter((item) => item.lang === lang)) {
    for (const category of post.categories) {
      categories.set(category.slug, category)
    }
  }
  return Array.from(categories.values())
}
