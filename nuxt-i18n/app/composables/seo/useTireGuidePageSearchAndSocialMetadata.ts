import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import { useHead, useI18n, useRoute } from '#imports'
import { useSiteTitle } from '~/composables/useSiteTitle'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'
import { isInnerTubeGuidePublishedLocale } from '~/utils/seo/innerTubeGuidePublishedLocaleIndexingPolicy'
import { resolveInnerTubeGuideContentPanelFromTabQuery } from '~/utils/innerTubeGuideContentSectionNavigation'
import { locales as storefrontLocales } from '~/i18n/locales.manifest'
import type { TireGuideTabId } from '~/utils/pageSubNavigation'

interface TireGuidePageSearchAndSocialMetadataOptions {
  fallbackPageTitle: MaybeRefOrGetter<string>
  activeGuideTabId: MaybeRefOrGetter<TireGuideTabId | null>
}

interface TireGuideSearchAndSocialMetadataDefinition {
  messageKey: string
  socialPreviewImage?: { path: string; mimeType: string; width: number; height: number }
}

const tireGuideSearchAndSocialMetadataByTabId: Partial<Record<TireGuideTabId, TireGuideSearchAndSocialMetadataDefinition>> = {
  'choose-inner-tube': {
    messageKey: 'innerTube',
    socialPreviewImage: {
      path: '/images/guides/inner-tube/bicycle-inner-tube-size-valve-and-rim-depth-guide-social-preview.png',
      mimeType: 'image/png', width: 1200, height: 630,
    },
  },
  'schwalbe-tire-circumference': {
    messageKey: 'schwalbeCircumference',
    socialPreviewImage: {
      path: '/public/tiresizecharts/schwalbe-tire-circumference/exact-circumference-of-tire.webp',
      mimeType: 'image/webp', width: 1200, height: 675,
    },
  },
  'tire-size-markings': {
    messageKey: 'tireSizeMarkings',
    socialPreviewImage: {
      path: '/public/tiresizecharts/tiresize/schwalbe-tiresize.webp',
      mimeType: 'image/webp', width: 1280, height: 720,
    },
  },
  tubeless: {
    messageKey: 'tubeless',
    socialPreviewImage: {
      path: '/public/tiresizecharts/installation/3/Tubelesstire-installation-diagram-2.webp',
      mimeType: 'image/webp', width: 1280, height: 720,
    },
  },
  'tire-pressure': { messageKey: 'tirePressure' },
}

export function useTireGuidePageSearchAndSocialMetadata(
  options: TireGuidePageSearchAndSocialMetadataOptions,
) {
  const { t, locale } = useI18n()
  const route = useRoute()
  const { siteTitle } = useSiteTitle()
  const { canonicalUrl } = useStorefrontSeoLinks()
  const activeGuideTabId = computed(() => toValue(options.activeGuideTabId))
  const isInnerTubeGuidePage = computed(() => activeGuideTabId.value === 'choose-inner-tube')
  const activeTireGuideMetadataDefinition = computed(() => (
    activeGuideTabId.value ? tireGuideSearchAndSocialMetadataByTabId[activeGuideTabId.value] : undefined
  ))
  const activeTireGuideSearchMetadataKey = computed(() => {
    const messageKey = isInnerTubeGuidePage.value
      && resolveInnerTubeGuideContentPanelFromTabQuery(route.query) === 'calculator'
      ? 'innerTubeCalculator'
      : activeTireGuideMetadataDefinition.value?.messageKey
    return `guidesTireguides.seo.${messageKey}`
  })
  const activeTireGuideSearchTitle = computed(() => {
    const guideTitle = t(`${activeTireGuideSearchMetadataKey.value}.title`)
    const configuredBrand = siteTitle.value.trim()
    return configuredBrand ? `${guideTitle} | ${configuredBrand}` : guideTitle
  })

  // One reactive head entry owns these tags. Returning an empty list removes
  // guide metadata when Nuxt reuses the page for an unconfigured guide or landing.
  useHead(() => {
    const metadataDefinition = activeTireGuideMetadataDefinition.value
    if (!metadataDefinition) {
      return { title: toValue(options.fallbackPageTitle), meta: [] }
    }

    const title = activeTireGuideSearchTitle.value
    const description = t(`${activeTireGuideSearchMetadataKey.value}.description`)
    const image = metadataDefinition.socialPreviewImage
    const imageAlt = image ? t(`guidesTireguides.seo.${metadataDefinition.messageKey}.imageAlt`) : ''
    const imageUrl = image ? new URL(image.path, canonicalUrl.value).toString() : ''
    const openGraphLocale = isInnerTubeGuidePage.value
      ? t('guidesTireguides.seo.innerTube.openGraphLocale')
      : (storefrontLocales.find(entry => entry.code === locale.value)?.iso || locale.value).replace(/-/g, '_')

    return {
      title,
      meta: [
        ...(isInnerTubeGuidePage.value && !isInnerTubeGuidePublishedLocale(locale.value)
          ? [{ key: 'robots', name: 'robots', content: 'noindex,follow' }]
          : []),
        { key: 'description', name: 'description', content: description },
        { key: 'og:type', property: 'og:type', content: 'article' },
        { key: 'og:title', property: 'og:title', content: title },
        { key: 'og:description', property: 'og:description', content: description },
        { key: 'og:url', property: 'og:url', content: canonicalUrl.value },
        { key: 'og:locale', property: 'og:locale', content: openGraphLocale },
        ...(siteTitle.value.trim()
          ? [{ key: 'og:site_name', property: 'og:site_name', content: siteTitle.value.trim() }]
          : []),
        ...(image ? [
          { key: 'og:image', property: 'og:image', content: imageUrl },
          { key: 'og:image:type', property: 'og:image:type', content: image.mimeType },
          { key: 'og:image:width', property: 'og:image:width', content: String(image.width) },
          { key: 'og:image:height', property: 'og:image:height', content: String(image.height) },
          { key: 'og:image:alt', property: 'og:image:alt', content: imageAlt },
          { key: 'twitter:image', name: 'twitter:image', content: imageUrl },
          { key: 'twitter:image:alt', name: 'twitter:image:alt', content: imageAlt },
        ] : []),
        { key: 'twitter:card', name: 'twitter:card', content: image ? 'summary_large_image' : 'summary' },
        { key: 'twitter:title', name: 'twitter:title', content: title },
        { key: 'twitter:description', name: 'twitter:description', content: description },
      ],
    }
  })
}
