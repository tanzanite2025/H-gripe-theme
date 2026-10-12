export const bicycleInnerTubeGuideManufacturerReferenceUrls = [
  'https://www.schwalbe.com/en/technology-faq/tire-dimensions/',
  'https://www.tubolito.com/faq/',
] as const

interface BicycleInnerTubeGuideArticleAndBreadcrumbOptions {
  canonicalUrl: string
  headline: string
  description: string
  inLanguage: string
  publisherName: string
  imageCaption: string
  breadcrumbs: readonly { name: string; url: string }[]
}

export function createBicycleInnerTubeGuideArticleAndBreadcrumbStructuredData(options: BicycleInnerTubeGuideArticleAndBreadcrumbOptions) {
  const cleanCanonicalUrl = new URL(options.canonicalUrl)
  cleanCanonicalUrl.search = cleanCanonicalUrl.searchParams.get('tab') === 'calculator' ? '?tab=calculator' : ''
  cleanCanonicalUrl.hash = ''
  const canonicalUrl = cleanCanonicalUrl.toString()
  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'Article',
        '@id': `${canonicalUrl}#inner-tube-guide`,
        url: canonicalUrl,
        mainEntityOfPage: canonicalUrl,
        headline: options.headline,
        description: options.description,
        inLanguage: options.inLanguage,
        ...(options.publisherName.trim() ? {
          publisher: { '@type': 'Organization', name: options.publisherName.trim() },
        } : {}),
        image: {
          '@type': 'ImageObject',
          url: new URL('/images/guides/inner-tube/schwalbe-innertube-valve.webp', canonicalUrl).toString(),
          width: 600,
          height: 300,
          caption: options.imageCaption,
        },
        citation: [...bicycleInnerTubeGuideManufacturerReferenceUrls],
        hasPart: ['inner-tube-fitment-calculator', 'inner-tube-fitment-how-to', 'inner-tube-fitment-matrix']
          .map(sectionId => ({ '@id': `${canonicalUrl}#${sectionId}` })),
      },
      {
        '@type': 'BreadcrumbList',
        '@id': `${canonicalUrl}#inner-tube-guide-breadcrumbs`,
        itemListElement: options.breadcrumbs.map((breadcrumb, index) => {
          const cleanBreadcrumbUrl = new URL(breadcrumb.url, canonicalUrl)
          cleanBreadcrumbUrl.search = ''
          cleanBreadcrumbUrl.hash = ''
          return { '@type': 'ListItem', position: index + 1, name: breadcrumb.name, item: cleanBreadcrumbUrl.toString() }
        }),
      },
    ],
  }
}
