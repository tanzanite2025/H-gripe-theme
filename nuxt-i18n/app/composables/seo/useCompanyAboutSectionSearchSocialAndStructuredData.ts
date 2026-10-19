import { toValue, type MaybeRefOrGetter } from 'vue'
import { useHead, useI18n, useLocalePath } from '#imports'
import { useSiteTitle } from '~/composables/useSiteTitle'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'
import { getStorefrontLocaleLanguageTag } from '~/utils/storefrontLocales'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import type { CompanyAboutTabId } from '~/utils/pageSubNavigation'

interface CompanyAboutSectionGalleryImage {
  id: string
  src: string
  alt: string
  caption: string
  mimeType: string
  width: number
  height: number
}

interface CompanyAboutSectionSearchSocialAndStructuredDataOptions {
  sectionId: CompanyAboutTabId
  messageNamespace: 'companyAboutFactory' | 'companyAboutAppearance' | 'companyAboutHolePatterns'
  resolvedContentLocale: MaybeRefOrGetter<string>
  socialPreviewImage: { path: string; mimeType: string; width: number; height: number }
  video?: { contentPath: string; duration: string; uploadDate?: string }
  imageGallery?: {
    elementId: string
    title: MaybeRefOrGetter<string>
    images: MaybeRefOrGetter<readonly CompanyAboutSectionGalleryImage[]>
  }
}

// Called by the mounted section after its translations load. Unmounting removes
// its head entry, including JSON-LD, when the shared About route changes sections.
export function useCompanyAboutSectionSearchSocialAndStructuredData(
  options: CompanyAboutSectionSearchSocialAndStructuredDataOptions,
) {
  const { locale, t } = useI18n()
  const localePath = useLocalePath()
  const { siteTitle } = useSiteTitle()
  const { canonicalUrl } = useStorefrontSeoLinks()

  useHead(() => {
    const namespace = options.messageNamespace
    const pageName = t(`${namespace}.title`)
    const searchTitle = t(`${namespace}.seo.title`)
    const brandName = siteTitle.value.trim()
    const title = brandName ? `${searchTitle} | ${brandName}` : searchTitle
    const description = t(`${namespace}.seo.description`)
    const imageAlt = t(`${namespace}.seo.imageAlt`)
    const image = options.socialPreviewImage
    const imageUrl = new URL(image.path, canonicalUrl.value).toString()
    const renderedLanguage = getStorefrontLocaleLanguageTag(toValue(options.resolvedContentLocale) || locale.value, 'en')
    const pageId = `${canonicalUrl.value}#${options.sectionId}-page`
    const breadcrumbId = `${canonicalUrl.value}#${options.sectionId}-breadcrumbs`
    const videoId = `${canonicalUrl.value}#${options.sectionId}-process-video`
    const gallery = options.imageGallery
    const galleryId = gallery ? `${canonicalUrl.value}#${gallery.elementId}` : ''
    const galleryImages = gallery ? toValue(gallery.images).map(photo => ({
      '@type': 'ImageObject',
      '@id': `${canonicalUrl.value}#${options.sectionId}-photo-${photo.id}`,
      url: new URL(photo.src, canonicalUrl.value).toString(),
      contentUrl: new URL(photo.src, canonicalUrl.value).toString(),
      name: photo.caption,
      caption: photo.caption,
      description: photo.alt,
      encodingFormat: photo.mimeType,
      width: photo.width,
      height: photo.height,
      inLanguage: renderedLanguage,
      isPartOf: { '@id': galleryId },
    })) : []
    const galleryImageReferences = galleryImages.map(photo => ({ '@id': photo['@id'] }))

    return {
      title,
      titleTemplate: null,
      meta: [
        { key: 'description', name: 'description', content: description },
        { key: 'og:type', property: 'og:type', content: 'website' },
        { key: 'og:title', property: 'og:title', content: title },
        { key: 'og:description', property: 'og:description', content: description },
        { key: 'og:url', property: 'og:url', content: canonicalUrl.value },
        { key: 'og:locale', property: 'og:locale', content: renderedLanguage.replace(/-/g, '_') },
        ...(brandName ? [{ key: 'og:site_name', property: 'og:site_name', content: brandName }] : []),
        { key: 'og:image', property: 'og:image', content: imageUrl },
        { key: 'og:image:type', property: 'og:image:type', content: image.mimeType },
        { key: 'og:image:width', property: 'og:image:width', content: String(image.width) },
        { key: 'og:image:height', property: 'og:image:height', content: String(image.height) },
        { key: 'og:image:alt', property: 'og:image:alt', content: imageAlt },
        { key: 'twitter:card', name: 'twitter:card', content: 'summary_large_image' },
        { key: 'twitter:title', name: 'twitter:title', content: title },
        { key: 'twitter:description', name: 'twitter:description', content: description },
        { key: 'twitter:image', name: 'twitter:image', content: imageUrl },
        { key: 'twitter:image:alt', name: 'twitter:image:alt', content: imageAlt },
      ],
      script: [{
        key: `company-about-${options.sectionId}-structured-data`,
        ...createSeoJsonLdScript({
          '@context': 'https://schema.org',
          '@graph': [
            {
              '@type': 'WebPage',
              '@id': pageId,
              url: canonicalUrl.value,
              name: pageName,
              description,
              inLanguage: renderedLanguage,
              primaryImageOfPage: {
                '@type': 'ImageObject', contentUrl: imageUrl, caption: imageAlt,
                width: image.width, height: image.height,
              },
              breadcrumb: { '@id': breadcrumbId },
              ...(options.video ? { video: { '@id': videoId } } : {}),
              ...(galleryImages.length ? {
                image: galleryImageReferences,
                hasPart: { '@id': galleryId },
              } : {}),
            },
            {
              '@type': 'BreadcrumbList',
              '@id': breadcrumbId,
              itemListElement: [
                { name: t(`${namespace}.seo.breadcrumbHome`), url: new URL(localePath('/'), canonicalUrl.value).toString() },
                { name: t('company.nav.about'), url: new URL(localePath('/company/about'), canonicalUrl.value).toString() },
                ...(options.sectionId === 'factory' && canonicalUrl.value.endsWith('/company/about')
                  ? [] : [{ name: pageName, url: canonicalUrl.value }]),
              ].map((item, index) => ({
                '@type': 'ListItem', position: index + 1, name: item.name, item: item.url,
              })),
            },
            ...(options.video ? [{
              '@type': 'VideoObject',
              '@id': videoId,
              name: t(`${namespace}.seo.videoName`),
              description: t(`${namespace}.seo.videoDescription`),
              thumbnailUrl: [imageUrl],
              contentUrl: new URL(options.video.contentPath, canonicalUrl.value).toString(),
              duration: options.video.duration,
              ...(options.video.uploadDate ? { uploadDate: options.video.uploadDate } : {}),
              isPartOf: { '@id': pageId },
            }] : []),
            ...(gallery && galleryImages.length ? [{
              '@type': 'ImageGallery',
              '@id': galleryId,
              url: galleryId,
              name: toValue(gallery.title),
              inLanguage: renderedLanguage,
              isPartOf: { '@id': pageId },
              associatedMedia: galleryImageReferences,
            }, ...galleryImages] : []),
          ],
        }),
      }],
    }
  })
}
