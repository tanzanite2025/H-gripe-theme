export type BlogCategory = 'news' | 'wheelsbuild'

export interface BlogFeaturedImage {
  url: string
  width?: number | null
  height?: number | null
  alt?: string
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
  date: string
  featuredImage: BlogFeaturedImage | null
  categories: BlogCategory[]
  translations: BlogTranslationsMap
}

export interface BlogPostDetail extends BlogPostSummary {
  contentHtml: string
  canonicalUrl: string
}

const isoDate = (value: string) => new Date(value).toISOString()

const buildTranslations = (items: Array<{ lang: string; id: number; slug: string }>): BlogTranslationsMap => {
  return items.reduce<BlogTranslationsMap>((acc, entry) => {
    acc[entry.lang] = { id: entry.id, slug: entry.slug }
    return acc
  }, {})
}

const makePost = (post: Omit<BlogPostDetail, 'translations'> & { translationEntries: Array<{ lang: string; id: number; slug: string }> }): BlogPostDetail => {
  return {
    id: post.id,
    lang: post.lang,
    group: post.group,
    slug: post.slug,
    title: post.title,
    excerpt: post.excerpt,
    date: post.date,
    featuredImage: post.featuredImage,
    categories: post.categories,
    translations: buildTranslations(post.translationEntries),
    contentHtml: post.contentHtml,
    canonicalUrl: post.canonicalUrl,
  }
}

const posts: BlogPostDetail[] = [
  makePost({
    id: 101,
    lang: 'en',
    group: 'grp-wheelbuild-process-1',
    slug: 'wheelbuild-process-overview',
    title: 'Wheelbuild Process Overview',
    excerpt: 'A high-level look at how we validate rim layup, drilling, and final wheel tension.',
    date: isoDate('2025-11-18T10:00:00Z'),
    featuredImage: null,
    categories: ['wheelsbuild'],
    translationEntries: [
      { lang: 'en', id: 101, slug: 'wheelbuild-process-overview' },
      { lang: 'fr', id: 201, slug: 'processus-montage-roues' },
      { lang: 'de', id: 301, slug: 'laufradbau-prozess-ueberblick' },
    ],
    contentHtml:
      '<h2>What this covers</h2><p>This article explains our baseline workflow from rim inspection to final truing.</p><h3>Key steps</h3><ul><li>Rim QC before build</li><li>Spoke prep and lacing</li><li>Tension targets</li><li>Final inspection</li></ul>',
    canonicalUrl: 'https://example.com/blog/wheelbuild-process-overview',
  }),
  makePost({
    id: 102,
    lang: 'en',
    group: 'grp-hookless-tires-1',
    slug: 'hookless-tire-compatibility-quick-guide',
    title: 'Hookless Tire Compatibility: Quick Guide',
    excerpt: 'A practical checklist for tire fit, pressure limits, and safe setup on hookless rims.',
    date: isoDate('2025-10-29T08:30:00Z'),
    featuredImage: null,
    categories: ['news'],
    translationEntries: [
      { lang: 'en', id: 102, slug: 'hookless-tire-compatibility-quick-guide' },
      { lang: 'fr', id: 202, slug: 'compatibilite-pneu-hookless-guide' },
    ],
    contentHtml:
      '<h2>Checklist</h2><ol><li>Confirm tire is hookless-rated</li><li>Verify rim internal width</li><li>Follow pressure limits</li><li>Use correct tape & valves</li></ol>',
    canonicalUrl: 'https://example.com/blog/hookless-tire-compatibility-quick-guide',
  }),
  makePost({
    id: 103,
    lang: 'en',
    group: 'grp-spoke-tension-1',
    slug: 'spoke-tension-basics-for-carbon-rims',
    title: 'Spoke Tension Basics for Carbon Rims',
    excerpt: 'How to approach tension targets and balance left/right on modern wheelsets.',
    date: isoDate('2025-10-12T12:10:00Z'),
    featuredImage: null,
    categories: ['wheelsbuild'],
    translationEntries: [{ lang: 'en', id: 103, slug: 'spoke-tension-basics-for-carbon-rims' }],
    contentHtml:
      '<h2>Targets</h2><p>Always start with the rim manufacturer target range and validate with a calibrated gauge.</p>',
    canonicalUrl: 'https://example.com/blog/spoke-tension-basics-for-carbon-rims',
  }),
  makePost({
    id: 104,
    lang: 'en',
    group: 'grp-new-finish-1',
    slug: 'new-matte-finish-now-available',
    title: 'New Matte Finish Now Available',
    excerpt: 'We have added a matte finish option for select rim models. Learn what changes.',
    date: isoDate('2025-09-30T09:00:00Z'),
    featuredImage: null,
    categories: ['news'],
    translationEntries: [{ lang: 'en', id: 104, slug: 'new-matte-finish-now-available' }],
    contentHtml:
      '<p>The matte finish is designed to reduce glare and fingerprints while keeping a clean look.</p>',
    canonicalUrl: 'https://example.com/blog/new-matte-finish-now-available',
  }),
  makePost({
    id: 105,
    lang: 'en',
    group: 'grp-hub-engagement-1',
    slug: 'hub-engagement-explained',
    title: 'Hub Engagement Explained',
    excerpt: 'What engagement angle means and how it affects feel on technical climbs.',
    date: isoDate('2025-09-15T06:45:00Z'),
    featuredImage: null,
    categories: ['news'],
    translationEntries: [{ lang: 'en', id: 105, slug: 'hub-engagement-explained' }],
    contentHtml:
      '<h2>Engagement angle</h2><p>Smaller engagement angles can feel more immediate under load.</p>',
    canonicalUrl: 'https://example.com/blog/hub-engagement-explained',
  }),
  makePost({
    id: 106,
    lang: 'en',
    group: 'grp-truing-1',
    slug: 'truing-after-first-ride',
    title: 'Truing After the First Ride',
    excerpt: 'A short routine to check spoke tension and lateral true after a new wheelset settles.',
    date: isoDate('2025-08-27T15:20:00Z'),
    featuredImage: null,
    categories: ['wheelsbuild'],
    translationEntries: [{ lang: 'en', id: 106, slug: 'truing-after-first-ride' }],
    contentHtml:
      '<p>After your first ride, re-check tension balance and verify no nipples have backed off.</p>',
    canonicalUrl: 'https://example.com/blog/truing-after-first-ride',
  }),
  makePost({
    id: 107,
    lang: 'en',
    group: 'grp-warranty-update-1',
    slug: 'warranty-policy-update-2025',
    title: 'Warranty Policy Update (2025)',
    excerpt: 'A concise summary of what is covered and what information you need for support.',
    date: isoDate('2025-08-10T11:00:00Z'),
    featuredImage: null,
    categories: ['news'],
    translationEntries: [{ lang: 'en', id: 107, slug: 'warranty-policy-update-2025' }],
    contentHtml:
      '<p>We have clarified claim timelines and improved the documentation checklist.</p>',
    canonicalUrl: 'https://example.com/blog/warranty-policy-update-2025',
  }),
  makePost({
    id: 108,
    lang: 'en',
    group: 'grp-nipple-prep-1',
    slug: 'nipple-prep-compound-or-oil',
    title: 'Nipple Prep: Compound or Oil?',
    excerpt: 'How to choose between spoke prep compound and light oil for consistent builds.',
    date: isoDate('2025-07-22T07:05:00Z'),
    featuredImage: null,
    categories: ['wheelsbuild'],
    translationEntries: [{ lang: 'en', id: 108, slug: 'nipple-prep-compound-or-oil' }],
    contentHtml:
      '<p>Use a consistent approach so tension changes are predictable and repeatable.</p>',
    canonicalUrl: 'https://example.com/blog/nipple-prep-compound-or-oil',
  }),
  makePost({
    id: 201,
    lang: 'fr',
    group: 'grp-wheelbuild-process-1',
    slug: 'processus-montage-roues',
    title: 'Processus de montage des roues',
    excerpt: 'Vue d’ensemble de notre flux de travail: contrôle, rayonnage, tension et inspection finale.',
    date: isoDate('2025-11-18T10:00:00Z'),
    featuredImage: null,
    categories: ['wheelsbuild'],
    translationEntries: [
      { lang: 'en', id: 101, slug: 'wheelbuild-process-overview' },
      { lang: 'fr', id: 201, slug: 'processus-montage-roues' },
      { lang: 'de', id: 301, slug: 'laufradbau-prozess-ueberblick' },
    ],
    contentHtml:
      '<h2>Ce que vous allez apprendre</h2><p>Du contrôle de la jante au dévoilage final.</p>',
    canonicalUrl: 'https://example.com/fr/blog/processus-montage-roues',
  }),
  makePost({
    id: 202,
    lang: 'fr',
    group: 'grp-hookless-tires-1',
    slug: 'compatibilite-pneu-hookless-guide',
    title: 'Compatibilité pneu hookless: guide rapide',
    excerpt: 'Une checklist pratique: montage, pression et sécurité sur jantes hookless.',
    date: isoDate('2025-10-29T08:30:00Z'),
    featuredImage: null,
    categories: ['news'],
    translationEntries: [
      { lang: 'en', id: 102, slug: 'hookless-tire-compatibility-quick-guide' },
      { lang: 'fr', id: 202, slug: 'compatibilite-pneu-hookless-guide' },
    ],
    contentHtml:
      '<h2>Checklist</h2><ol><li>Pneu compatible hookless</li><li>Largeur interne</li><li>Pression max</li></ol>',
    canonicalUrl: 'https://example.com/fr/blog/compatibilite-pneu-hookless-guide',
  }),
  makePost({
    id: 301,
    lang: 'de',
    group: 'grp-wheelbuild-process-1',
    slug: 'laufradbau-prozess-ueberblick',
    title: 'Laufradbau-Prozess: Überblick',
    excerpt: 'Ein Überblick über Prüfung, Einspeichen, Speichenspannung und Endkontrolle.',
    date: isoDate('2025-11-18T10:00:00Z'),
    featuredImage: null,
    categories: ['wheelsbuild'],
    translationEntries: [
      { lang: 'en', id: 101, slug: 'wheelbuild-process-overview' },
      { lang: 'fr', id: 201, slug: 'processus-montage-roues' },
      { lang: 'de', id: 301, slug: 'laufradbau-prozess-ueberblick' },
    ],
    contentHtml:
      '<h2>Worum es geht</h2><p>Von der Felgenprüfung bis zur Endkontrolle.</p>',
    canonicalUrl: 'https://example.com/de/blog/laufradbau-prozess-ueberblick',
  }),
]

export const listBlogPosts = (params: {
  lang: string
  category?: BlogCategory
}): BlogPostSummary[] => {
  const items = posts
    .filter((post) => post.lang === params.lang)
    .filter((post) => {
      if (!params.category) return true
      return post.categories.includes(params.category)
    })
    .sort((a, b) => (a.date > b.date ? -1 : a.date < b.date ? 1 : 0))

  return items.map(({ contentHtml, canonicalUrl, ...summary }) => summary)
}

export const getBlogPostBySlug = (params: {
  lang: string
  slug: string
  category?: BlogCategory
}): BlogPostDetail | null => {
  const slug = params.slug.trim()
  if (!slug) return null

  const post = posts.find((item) => {
    if (item.lang !== params.lang) return false
    if (item.slug !== slug) return false
    if (params.category && !item.categories.includes(params.category)) return false
    return true
  })

  return post || null
}

export const getBlogCategoryFromPost = (post: BlogPostSummary): BlogCategory => {
  if (post.categories.includes('news')) return 'news'
  return 'wheelsbuild'
}

export const buildBlogDetailPath = (post: BlogPostSummary): string => {
  const category = getBlogCategoryFromPost(post)
  return `/blog/${category}/${post.slug}`
}
