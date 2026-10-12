export interface StainlessSteelSpokeMaterialSourceCitation {
  title: string
  role: string
  location?: string
  url?: string
}

interface StainlessSteelSpokeTechnicalArticleAndElongationCalculatorStructuredDataOptions {
  canonicalUrl: string
  headline: string
  description: string
  calculatorName: string
  calculatorDescription: string
  calculatorSoftwareVersion?: string
  inLanguage: string
  publisherName: string
  articleSections: readonly string[]
  topics: readonly string[]
  sourceReferences: readonly StainlessSteelSpokeMaterialSourceCitation[]
}

export function normalizeStainlessSteelSpokeMaterialSourcesForVisibleReferencesAndCitations(
  sources: readonly StainlessSteelSpokeMaterialSourceCitation[],
): StainlessSteelSpokeMaterialSourceCitation[] {
  return sources.map(source => {
    let sourceUrl: string | undefined
    try {
      const parsedSourceUrl = new URL(source.url || '')
      if (['https:', 'http:'].includes(parsedSourceUrl.protocol)) sourceUrl = parsedSourceUrl.toString()
    } catch {
      // A reference without a valid public URL can still be cited by title.
    }
    return {
      title: source.title.trim(),
      role: source.role.trim(),
      ...(source.location?.trim() ? { location: source.location.trim() } : {}),
      ...(sourceUrl ? { url: sourceUrl } : {}),
    }
  }).filter(source => source.title)
}

export function createStainlessSteelSpokeTechnicalArticleAndElongationCalculatorStructuredData(
  options: StainlessSteelSpokeTechnicalArticleAndElongationCalculatorStructuredDataOptions,
) {
  const cleanCanonicalUrl = new URL(options.canonicalUrl)
  cleanCanonicalUrl.search = ''
  cleanCanonicalUrl.hash = ''
  const canonicalUrl = cleanCanonicalUrl.toString()
  const technicalArticleId = `${canonicalUrl}#stainless-steel-spoke-technical-article`
  const elongationCalculatorId = `${canonicalUrl}#stainless-steel-spoke-axial-elongation-calculator`

  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'TechArticle',
        '@id': technicalArticleId,
        url: canonicalUrl,
        mainEntityOfPage: canonicalUrl,
        headline: options.headline,
        description: options.description,
        inLanguage: options.inLanguage,
        proficiencyLevel: 'Expert',
        articleSection: [...options.articleSections],
        about: options.topics.map(name => ({ '@type': 'Thing', name })),
        isAccessibleForFree: true,
        ...(options.publisherName.trim() ? {
          publisher: { '@type': 'Organization', name: options.publisherName.trim() },
        } : {}),
        ...(options.sourceReferences.length ? {
          citation: options.sourceReferences.map(source => ({
            '@type': 'CreativeWork',
            name: source.title,
            ...(source.url ? { url: source.url } : {}),
            description: [source.location, source.role].filter(Boolean).join('; '),
          })),
        } : {}),
        hasPart: { '@id': elongationCalculatorId },
      },
      {
        '@type': 'WebApplication',
        '@id': elongationCalculatorId,
        url: elongationCalculatorId,
        mainEntityOfPage: canonicalUrl,
        isPartOf: { '@id': technicalArticleId },
        name: options.calculatorName,
        description: options.calculatorDescription,
        inLanguage: options.inLanguage,
        applicationCategory: 'EngineeringApplication',
        operatingSystem: 'Web',
        isAccessibleForFree: true,
        offers: { '@type': 'Offer', price: '0', priceCurrency: 'USD', url: elongationCalculatorId },
        ...(options.calculatorSoftwareVersion ? { softwareVersion: options.calculatorSoftwareVersion } : {}),
      },
    ],
  }
}
