<template>
  <div>
    <h1 v-if="!isGuideCategoryLandingPage" class="products-page__title products-page__title--sr-only">
      {{ activePageTitle }}
    </h1>
    <p v-if="!isGuideCategoryLandingPage" class="products-page__intro products-page__intro--sr-only">
      {{ t('guidesTireguides.intro') }}
    </p>

    <GuideCategoryChildRouteNavigationCards
      v-if="isGuideCategoryLandingPage"
      eyebrow="Tire Guides"
      :heading="t('guidesTireguides.title')"
      :description="t('guidesTireguides.intro')"
      open-label="Open guide"
      :cards="tireGuideNavigationCards"
    />

    <div v-else class="sizecharts-page">
      <!-- Tire size marking education -->
      <section
        v-show="activeTab === 'tire-size-markings'"
        id="tire-size-markings"
        class="sizecharts-section tz-text-secondary"
      >
        <TireSizeGuide v-if="activeTab === 'tire-size-markings'" @open-tire-products="openTireProductsDrawer" />
      </section>

      <!-- Tire frame clearance -->
      <section
        v-show="activeTab === 'tire-frame-clearance'"
        id="tire-frame-clearance"
        class="sizecharts-section"
      >
        <TireFrameClearanceGuide v-if="activeTab === 'tire-frame-clearance'" />
      </section>

      <!-- Schwalbe tire circumference -->
      <section
        v-show="activeTab === 'schwalbe-tire-circumference'"
        id="schwalbe-tire-circumference"
        class="sizecharts-section"
      >
        <SchwalbeTireCircumferenceGuide v-if="activeTab === 'schwalbe-tire-circumference'" />
      </section>

      <section
        v-show="activeTab === 'tubeless'"
        id="tubeless"
        class="sizecharts-section"
      >
        <div class="tubeless-installation-guide space-y-12">
          <TubelessGuide v-if="activeTab === 'tubeless'" @change-tab="setActiveTab" />
          <div id="tubeless-installation">
            <InstallationGuide v-if="activeTab === 'tubeless'" />
          </div>
        </div>
      </section>

      <!-- How to choose -->
      <section
        v-show="activeTab === 'choose'"
        id="choose"
        class="sizecharts-section"
      >
        <HowToChooseGuide v-if="activeTab === 'choose'" />
      </section>

      <!-- Tire pressure -->
      <section
        v-show="activeTab === 'tire-pressure'"
        id="tire-pressure"
        class="sizecharts-section"
      >
        <TirePressureGuide v-if="activeTab === 'tire-pressure'" />
      </section>

      <!-- How to choose an inner tube -->
      <section
        v-show="activeTab === 'choose-inner-tube'"
        id="choose-inner-tube"
        class="sizecharts-section"
      >
        <InnerTubeGuide v-if="activeTab === 'choose-inner-tube'" />
      </section>

      <div v-if="activeTab !== 'choose-inner-tube'" class="sizecharts-feedback">
      <UserFeedbackThread
        threadKey="guides-tireguides"
        :title="t('guidesTireguides.feedbackTitle')"
      />
      </div>
    </div>
  </div>

  <WhatsAppProductSearchResultDrawer
    v-model="tireProductsDrawerVisible"
    :loading="tireProductsLoading"
    :results="tireProductsResults"
    :error="tireProductsError"
    :query="tireProductsQuery"
    @close="handleTireProductsDrawerClose"
  />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useHead, useI18n, useLocalePath, useRoute } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import WhatsAppProductSearchResultDrawer from '~/components/WhatsAppProductSearchResultDrawer.vue'
import TireFrameClearanceGuide from '~/components/tireguides/TireFrameClearanceGuide.vue'
import SchwalbeTireCircumferenceGuide from '~/components/tireguides/SchwalbeTireCircumferenceGuide.vue'
import TubelessGuide from '~/components/tireguides/TubelessGuide.vue'
import HowToChooseGuide from '~/components/tireguides/HowToChooseGuide.vue'
import TirePressureGuide from '~/components/tireguides/TirePressureGuide.vue'
import InnerTubeGuide from '~/components/tireguides/InnerTubeGuide.vue'
import InstallationGuide from '~/components/tireguides/InstallationGuide.vue'
import TireSizeGuide from '~/components/tireguides/TireSizeGuide.vue'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import { normalizeShopProduct } from '~/composables/useShopProducts'
import { pageSubNavigationChildPath, tireGuideTabs } from '~/utils/pageSubNavigation'
import { usePageMessages } from '~/composables/usePageMessages'
import GuideCategoryChildRouteNavigationCards from '~/components/GuideCategoryChildRouteNavigationCards.vue'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'products.nav.tireSizeCharts',
  footerLabelFallback: 'Tire Guides',
})

const { locale, t } = useI18n()
const route = useRoute()
const localePath = useLocalePath()
const { loadPageMessages } = usePageMessages('guidesTireguides')
const { loadPageMessages: loadTireRimReferencePageMessages } = usePageMessages('guidesTireChoose')
const { loadPageMessages: loadTireSizePageMessages } = usePageMessages('guidesTireSize')

const normalizeGuideCategoryLandingRoutePath = (path: string) => path.replace(/\/+$/, '') || '/'

await Promise.all([
  loadPageMessages(locale.value),
  loadTireRimReferencePageMessages(locale.value),
  loadTireSizePageMessages(locale.value),
])

watch(locale, (nextLocale) => {
  void Promise.all([
    loadPageMessages(nextLocale),
    loadTireRimReferencePageMessages(nextLocale),
    loadTireSizePageMessages(nextLocale),
  ])
})

const tabs = tireGuideTabs
const isGuideCategoryLandingPage = computed(() => (
  normalizeGuideCategoryLandingRoutePath(route.path) ===
  normalizeGuideCategoryLandingRoutePath(localePath('/guides/tireguides'))
))
const getTireGuideCardRoute = (tab: (typeof tabs)[number]) => {
  const explicitRoute = (tab as { to?: string }).to
  return explicitRoute || pageSubNavigationChildPath('/guides/tireguides', tab.id)
}
const tireGuideNavigationCards = computed(() => tabs.map(tab => ({
  id: tab.id,
  label: t(tab.labelKey, tab.fallback),
  description: t(tab.descriptionKey, tab.description),
  to: getTireGuideCardRoute(tab),
})))
const { activeTab, setActiveTab } = usePageSubNavigationTab({
  tabs,
  basePath: '/guides/tireguides',
  defaultValue: 'tubeless',
})

const activePageTitle = computed(() => {
  if (isGuideCategoryLandingPage.value) {
    return t('guidesTireguides.title')
  }

  if (activeTab.value === 'tire-size-markings') {
    return t('guidesTireSize.title')
  }
  if (activeTab.value === 'tubeless') {
    return t('guidesTireguides.tabs.tubeless.label')
  }
  if (activeTab.value === 'tire-frame-clearance') {
    return t('guidesTireguides.tabs.tireFrameClearance.label')
  }
  if (activeTab.value === 'schwalbe-tire-circumference') {
    return t('guidesTireguides.tabs.schwalbeCircumference.label')
  }
  if (activeTab.value === 'choose') {
    return t('guidesTireChoose.seo.title')
  }
  if (activeTab.value === 'choose-inner-tube') {
    return t('guidesTireguides.tabs.innerTube.label')
  }
  return t('guidesTireguides.title')
})

useHead(() => ({
  title: activePageTitle.value,
}))

const { request } = useApiRequest()

// Tire products drawer
const tireProductsDrawerVisible = ref(false)
const tireProductsLoading = ref(false)
const tireProductsResults = ref<any[]>([])
const tireProductsError = ref<string | null>(null)
const tireProductsQuery = ref('')

const openTireProductsDrawer = async () => {
  const keyword = 'tire'

  tireProductsQuery.value = t('guidesTireguides.drawer.query')
  tireProductsError.value = null
  tireProductsDrawerVisible.value = true
  tireProductsLoading.value = true

  try {
    const response = await request<any>('/customer-service/products', {
      params: {
        keyword,
        per_page: 20,
        status: 'active',
      },
      credentials: 'include',
    }, t('guidesTireguides.drawer.error'))

    const products = Array.isArray(response?.items) ? response.items : []
    if (products.length > 0) {
      tireProductsResults.value = products.map((item: any) => {
        const normalized = normalizeShopProduct(item)
        return {
          ...normalized,
          thumbnail: normalized.thumbnail || item.thumbnail,
          price: normalized.priceLabel ||
            (Number(item.prices?.sale) > 0
              ? `$${item.prices.sale}`
              : Number(item.prices?.regular) > 0
                ? `$${item.prices.regular}`
                : ''),
        }
      })
    } else {
      tireProductsResults.value = []
    }
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to load tire products', error)
    tireProductsError.value = t('guidesTireguides.drawer.error')
    tireProductsResults.value = []
  } finally {
    tireProductsLoading.value = false
  }
}

const handleTireProductsDrawerClose = () => {
  tireProductsDrawerVisible.value = false
  tireProductsError.value = null
  tireProductsQuery.value = ''
  tireProductsResults.value = []
  tireProductsLoading.value = false
}

</script>

<style src="~/assets/css/guide-sections.css"></style>

<style scoped>
.products-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.products-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

.products-page__title--sr-only {
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

.products-page__intro--sr-only {
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

.sizecharts-section__title--sr-only {
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

.sizecharts-page {
  margin: 0 auto;
  width: 100%;
  max-width: none;
}

/* Page-level tab entry points are rendered by the header/mobile mega menu. */


.sizecharts-brand-button {
  margin-left: 0.5rem;
  margin-top: 0.5rem;
  padding: 0.25rem 0.8rem;
  border-radius: 9999px;
  border: 1px solid var(--tz-action-primary);
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.sizecharts-brand-button:hover {
  border-color: var(--tz-action-primary-hover);
  background: var(--tz-action-primary-hover);
}

.sizecharts-feedback {
  margin-top: 2.5rem;
}

.sizecharts-installation-images {
  margin-top: 0.75rem;
  display: flex;
  gap: 0.5rem;
}

.sizecharts-installation-images__item {
  flex: 1 1 0;
}

.sizecharts-installation-images__img {
  width: 100%;
  height: 140px;
  object-fit: cover;
  border-radius: 0.5rem;
  border: none;
  box-shadow: 2px 2px 4px rgba(0, 0, 0, 0.9);
}

.sizecharts-installation-images--tubeless {
  /* AUTO-FIT GRID for GuideImage rows: 1 image = full row, 2 images = two equal columns */
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 0.5rem;
}

.sizecharts-installation-images--installation {
  /* AUTO-FIT GRID for GuideImage rows: 1–3 images share the row evenly */
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 0.5rem;
}

.sizecharts-installation-images--tubeless .sizecharts-installation-images__item {
  flex: 1 1 0;
}

.sizecharts-installation-images--tubeless .sizecharts-installation-images__img {
  height: auto;
  aspect-ratio: 16 / 9;
  object-fit: contain;
}

@media (max-width: 768px) {
  .sizecharts-tabs {
    justify-content: flex-start;
  }
}
</style>
