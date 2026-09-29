<template>
  <div class="schwalbe-page">
    <SchwalbeTireSelector />
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useHead, useI18n, useRoute, useSwitchLocalePath } from '#imports'
import SchwalbeTireSelector from '~/components/tireguides/schwalbe/SchwalbeTireSelector.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import { useStorefrontSeoRouteOverride } from '~/composables/seo/useStorefrontSeoLinks'
import localeManifest from '~/i18n/locales.manifest'

definePageMeta({
  // Keep this page at the Schwalbe selector URL while keeping it out of the
  // `tireguides.vue` component hierarchy. The parent page renders the default
  // Tubeless guide itself and has no `<NuxtPage />` outlet.
  path: '/guides/tireguides/schwalbe-tire-selector',
  middleware: ['schwalbe-selector-i18n'],
  breadcrumbLabelKey: 'guidesSchwalbeTireSelector.title',
  breadcrumbLabelFallback: 'Schwalbe tire selector',
  layout: 'products',
  footerLabelKey: 'products.nav.tireSizeCharts',
  footerLabelFallback: 'Tire Guides',
})

const { locale, t } = useI18n()
const route = useRoute()
const switchLocalePath = useSwitchLocalePath()
const { loadPageMessages } = usePageMessages('guidesSchwalbeTireSelector')

await loadPageMessages(locale.value)
watch(locale, (nextLocale) => void loadPageMessages(nextLocale))

const indexablePaginationPage = computed(() => {
  const queryKeys = Object.keys(route.query)
  if (queryKeys.length !== 1 || queryKeys[0] !== 'page') return null

  const rawPage = route.query.page
  if (typeof rawPage !== 'string') return null

  const normalizedPage = rawPage.trim()
  if (!/^[1-9]\d*$/.test(normalizedPage)) return null

  const page = Number(normalizedPage)
  return Number.isSafeInteger(page) && page > 1 ? page : null
})

const localizedSeoRoutes = computed(() => {
  if (Object.keys(route.query).length === 0) return null

  const page = indexablePaginationPage.value

  return localeManifest.map(({ code }) => {
    const localizedPath = switchLocalePath(code as any) || route.path
    const localizedUrl = new URL(localizedPath, 'https://selector-seo.invalid')
    localizedUrl.search = ''
    localizedUrl.hash = ''
    if (page !== null) localizedUrl.searchParams.set('page', String(page))

    return {
      code,
      path: `${localizedUrl.pathname}${localizedUrl.search}`,
    }
  })
})

useStorefrontSeoRouteOverride(localizedSeoRoutes)

const hasPageOneQuery = computed(() => (
  Object.keys(route.query).length === 1
  && typeof route.query.page === 'string'
  && route.query.page.trim() === '1'
))
const shouldNoIndexQuery = computed(() => (
  Object.keys(route.query).length > 0
  && !hasPageOneQuery.value
  && indexablePaginationPage.value === null
))

useHead(() => ({
  title: t('guidesSchwalbeTireSelector.seo.title'),
  meta: [
    {
      name: 'description',
      content: t('guidesSchwalbeTireSelector.seo.description'),
    },
    ...(shouldNoIndexQuery.value
      ? [{ name: 'robots', content: 'noindex,follow', key: 'robots' }]
      : []),
  ],
}))
</script>

<style scoped>
.schwalbe-page {
  display: grid;
  gap: 2rem;
  width: 100%;
  max-width: 96rem;
  margin: 0 auto;
}
</style>
