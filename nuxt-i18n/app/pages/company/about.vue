<template>
  <div class="company-page">

    <ReusablePageSectionUnderlineTabs
      :tabs="companyAboutSectionNavigationTabs"
      :active-id="activeTab"
      :tab-list-label="t('companyAboutAppearance.navigation.ariaLabel')"
    />

    <h1 v-if="activeTab !== 'appearance'" class="company-page__title company-page__title--sr-only">{{ t('company.nav.about') }}</h1>

    <AboutFactory
      v-if="activeTab === 'factory'"
    />

    <AboutAppearance
      v-else-if="activeTab === 'appearance'"
    />

    <AboutHolePatterns
      v-else-if="activeTab === 'hole-patterns'"
    />

    <div class="company-feedback">
      <UserFeedbackThread
        threadKey="company-ourstory"
      />
    </div>


  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useHead, definePageMeta, useI18n, useLocalePath } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import ReusablePageSectionUnderlineTabs from '~/components/ReusablePageSectionUnderlineTabs.vue'
import AboutFactory from '~/components/company/AboutFactory.vue'
import AboutAppearance from '~/components/company/AboutAppearance.vue'
import AboutHolePatterns from '~/components/company/AboutHolePatterns.vue'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import { companyAboutTabs } from '~/utils/pageSubNavigation'
import { usePageMessages } from '~/composables/usePageMessages'
import { useSiteTitle } from '~/composables/useSiteTitle'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'
import { getStorefrontLocaleLanguageTag } from '~/utils/storefrontLocales'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'

const { locale, t } = useI18n()
const tabs = companyAboutTabs
const { activeTab, localizedTabPath } = usePageSubNavigationTab({
  tabs,
  basePath: '/company/about',
  defaultValue: 'factory',
})

const { loadPageMessages, resolvedLocale } = usePageMessages('companyAboutAppearance')
const { siteTitle } = useSiteTitle()
const { canonicalUrl } = useStorefrontSeoLinks()
const localePath = useLocalePath()

// All About Us sections use the same translated route navigation.
await loadPageMessages(locale.value)

watch(locale, (requestedLocale) => {
  void loadPageMessages(requestedLocale)
})

const companyAboutSectionNavigationTabs = computed(() => tabs.map(tab => ({
  id: tab.id,
  label: t(`companyAboutAppearance.navigation.tabs.${tab.id}`),
  to: localizedTabPath(tab.id),
})))

const appearanceSearchTitle = computed(() => {
  const title = t('companyAboutAppearance.seo.title')
  const brandName = siteTitle.value.trim()
  return brandName ? `${title} | ${brandName}` : title
})
const appearanceSearchDescription = computed(() => t('companyAboutAppearance.seo.description'))
const appearanceSocialPreviewImageUrl = computed(() => new URL(
  '/company/aboutus/appearance/carbon-rim-laser-engraving-LOGO.webp',
  canonicalUrl.value,
).toString())

definePageMeta({
  layout: 'products',
  footerLabelKey: 'company.nav.about',
  footerLabelFallback: 'About Us',
})

useHead(() => {
  if (activeTab.value !== 'appearance') {
    return { title: t('company.nav.about'), meta: [], script: [] }
  }

  const title = appearanceSearchTitle.value
  const description = appearanceSearchDescription.value
  const imageUrl = appearanceSocialPreviewImageUrl.value
  const imageAlt = t('companyAboutAppearance.seo.imageAlt')
  const renderedLanguage = getStorefrontLocaleLanguageTag(resolvedLocale.value || locale.value, 'en')

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
      ...(siteTitle.value.trim()
        ? [{ key: 'og:site_name', property: 'og:site_name', content: siteTitle.value.trim() }]
        : []),
      { key: 'og:image', property: 'og:image', content: imageUrl },
      { key: 'og:image:type', property: 'og:image:type', content: 'image/webp' },
      { key: 'og:image:width', property: 'og:image:width', content: '1280' },
      { key: 'og:image:height', property: 'og:image:height', content: '720' },
      { key: 'og:image:alt', property: 'og:image:alt', content: imageAlt },
      { key: 'twitter:card', name: 'twitter:card', content: 'summary_large_image' },
      { key: 'twitter:title', name: 'twitter:title', content: title },
      { key: 'twitter:description', name: 'twitter:description', content: description },
      { key: 'twitter:image', name: 'twitter:image', content: imageUrl },
      { key: 'twitter:image:alt', name: 'twitter:image:alt', content: imageAlt },
    ],
    script: [{
      key: 'company-about-appearance-web-page-and-breadcrumb-structured-data',
      ...createSeoJsonLdScript({
        '@context': 'https://schema.org',
        '@graph': [
          {
            '@type': 'WebPage',
            '@id': `${canonicalUrl.value}#appearance-page`,
            url: canonicalUrl.value,
            name: t('companyAboutAppearance.title'),
            description,
            inLanguage: renderedLanguage,
            primaryImageOfPage: {
              '@type': 'ImageObject',
              contentUrl: imageUrl,
              caption: imageAlt,
              width: 1280,
              height: 720,
            },
            breadcrumb: { '@id': `${canonicalUrl.value}#appearance-breadcrumbs` },
          },
          {
            '@type': 'BreadcrumbList',
            '@id': `${canonicalUrl.value}#appearance-breadcrumbs`,
            itemListElement: [
              { name: t('companyAboutAppearance.seo.breadcrumbHome'), path: localePath('/') },
              { name: t('company.nav.about'), path: localePath('/company/about') },
              { name: t('companyAboutAppearance.title'), path: localePath('/company/about/appearance') },
            ].map((item, index) => ({
              '@type': 'ListItem',
              position: index + 1,
              name: item.name,
              item: new URL(item.path, canonicalUrl.value).toString(),
            })),
          },
        ],
      }),
    }],
  }
})
</script>

<style scoped>
.company-page {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.company-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.company-page__title--sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.company-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

.company-section {
  margin-top: 0;
}

.company-section__title {
  margin: 0 0 0.5rem;
  font-size: var(--tz-type-section-title);
  line-height: 1.35;
  font-weight: 600;
  color: var(--tz-text-primary);
  text-align: center;
}

.company-section__body {
  margin: 0 0 0.5rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
  text-align: center;
}

.company-section__list {
  margin: 0.25rem auto 0;
  padding-left: 0;
  list-style-type: none;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
  text-align: center;
}

.company-section__list li + li {
  margin-top: 0.25rem;
}

.company-video-button {
  margin-top: 0.5rem;
  padding: 0.35rem 0.85rem;
  border-radius: 9999px;
  border: 1px solid var(--tz-action-primary);
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
}





.company-section--values {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-section--timeline {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-section--cta {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-values {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
}

.company-values__item {
  padding: 0.75rem 0.9rem;
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
  border: 1px solid var(--tz-border-subtle);
}

.company-values__title {
  margin: 0 0 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.company-values__body {
  margin: 0;
  font-size: 0.9rem;
  color: var(--tz-text-secondary);
}

.company-timeline {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.company-timeline__item {
  display: grid;
  grid-template-columns: auto 1fr;
  column-gap: 0.75rem;
  row-gap: 0.25rem;
  align-items: flex-start;
}

.company-timeline__year {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--tz-text-primary);
  min-width: 3.5rem;
}

.company-timeline__content {
  font-size: 0.9rem;
  color: var(--tz-text-secondary);
}

.company-feedback {
  margin-top: 2.5rem;
}

@media (min-width: 768px) {
  .company-values {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

}



</style>
