<template>
  <div>
    <h1 class="products-page__title products-page__title--sr-only">Tire Guides</h1>
    <p class="products-page__intro products-page__intro--sr-only">
      Reference charts for common tire and rim sizes. Detailed data will be added here later.
    </p>

    <div class="sizecharts-page">
      <PageTabBar
        :tabs="tabs"
        :active-id="activeTab"
        aria-label="Tire guide sections"
        @select="setActiveTab"
      />

      <!-- Tire size (new top-level tab) -->
      <section
        v-show="activeTab === 'size'"
        id="size"
        class="sizecharts-section text-slate-100"
      >
        <TireSizeGuide @open-tire-products="openTireProductsDrawer" />
      </section>

      <!-- Match (tire & rim matching helpers) -->
      <section
        v-show="activeTab === 'match'"
        id="match"
        class="sizecharts-section"
      >
        <MatchGuide />
      </section>

      <section
        v-show="activeTab === 'tubeless'"
        id="tubeless"
        class="sizecharts-section"
      >
        <TubelessGuide @change-tab="setActiveTab" />
      </section>

      <!-- Installation -->
      <section
        v-show="activeTab === 'installation'"
        id="installation"
        class="sizecharts-section"
      >
        <InstallationGuide @change-tab="setActiveTab" />
      </section>

      <!-- How to choose -->
      <section
        v-show="activeTab === 'choose'"
        id="choose"
        class="sizecharts-section"
      >
        <HowToChooseGuide />
      </section>

      <!-- Tire pressure -->
      <section
        v-show="activeTab === 'rims'"
        id="rims"
        class="sizecharts-section"
      >
        <TirePressureGuide @open-tire-products="openTireProductsDrawer" />
      </section>

      <!-- Inner Tube -->
      <section
        v-show="activeTab === 'tube'"
        id="tube"
        class="sizecharts-section"
      >
        <InnerTubeGuide />
      </section>

      <section class="mt-6">
        <h2 class="products-page__title products-page__title--sr-only">Tire Guides FAQ</h2>
        <PageFaq
          page-id="guides-tireguides"
          theme="dark"
          :show-categories="true"
        />
      </section>

      <div class="sizecharts-feedback">
        <UserFeedbackThread
          threadKey="guides-tireguides"
          title="Share your feedback about this Tire Guides guide"
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
import { ref, watch } from 'vue'
import { useRoute, useRuntimeConfig } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import GuideImage from '~/components/GuideImage.vue'
import TireSizeSection from '~/components/TireSizeSection.vue'
import WhatsAppProductSearchResultDrawer from '~/components/WhatsAppProductSearchResultDrawer.vue'
import MatchGuide from '~/components/tireguides/MatchGuide.vue'
import TubelessGuide from '~/components/tireguides/TubelessGuide.vue'
import HowToChooseGuide from '~/components/tireguides/HowToChooseGuide.vue'
import TirePressureGuide from '~/components/tireguides/TirePressureGuide.vue'
import InnerTubeGuide from '~/components/tireguides/InnerTubeGuide.vue'
import InstallationGuide from '~/components/tireguides/InstallationGuide.vue'
import TireSizeGuide from '~/components/tireguides/TireSizeGuide.vue'

type SizeChartsTabId = 'size' | 'match' | 'tubeless' | 'installation' | 'choose' | 'rims' | 'tube'

definePageMeta({
  layout: 'products',
  path: '/guides/tireguides',
})

useHead({
  title: 'Tire Guides',
})

const tabs: { id: SizeChartsTabId; label: string }[] = [
  { id: 'size', label: 'Tire size' },
  { id: 'match', label: 'Match' },
  { id: 'tubeless', label: 'Tubeless tires' },
  { id: 'installation', label: 'Installation' },
  { id: 'choose', label: 'How to choose' },
  { id: 'rims', label: 'Tire pressure' },
  { id: 'tube', label: 'Inner tube' },
]

const activeTab = ref<SizeChartsTabId>('tubeless')

const route = useRoute()
const config = useRuntimeConfig()

const getTabFromHash = (hash: string): SizeChartsTabId | null => {
  const raw = String(hash || '').replace(/^#/, '')
  const allowed: SizeChartsTabId[] = ['size', 'match', 'tubeless', 'installation', 'choose', 'rims', 'tube']
  return (allowed as string[]).includes(raw) ? (raw as SizeChartsTabId) : null
}

// Tire products drawer
const tireProductsDrawerVisible = ref(false)
const tireProductsLoading = ref(false)
const tireProductsResults = ref<any[]>([])
const tireProductsError = ref<string | null>(null)
const tireProductsQuery = ref('')

const openTireProductsDrawer = async () => {
  const keyword = 'tire'

  tireProductsQuery.value = 'Tire products'
  tireProductsError.value = null
  tireProductsDrawerVisible.value = true
  tireProductsLoading.value = true

  try {
    const apiBase = String((config.public as { apiBase?: string }).apiBase || '/api/v1').replace(/\/$/, '')
    const response = await $fetch<any>(`${apiBase}/customer-service/products`, {
      params: {
        keyword,
        per_page: 20,
        status: 'active',
      },
      credentials: 'include',
    })

    const products = Array.isArray(response?.items) ? response.items : []
    if (products.length > 0) {
      tireProductsResults.value = products.map((item: any) => ({
        id: item.id,
        title: item.title,
        url: item.preview_url || `/shop/${item.slug || item.id}`,
        thumbnail: item.thumbnail,
        price:
          Number(item.prices?.sale) > 0
            ? `$${item.prices.sale}`
            : Number(item.prices?.regular) > 0
              ? `$${item.prices.regular}`
              : '',
      }))
    } else {
      tireProductsResults.value = []
    }
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to load tire products', error)
    tireProductsError.value = 'Failed to load tire products. Please try again.'
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

watch(
  () => route.hash,
  (hash) => {
    const next = getTabFromHash(hash)
    if (next) activeTab.value = next
  },
  { immediate: true }
)

const setActiveTab = (id: SizeChartsTabId | string) => {
  if (!tabs.some((tab) => tab.id === id)) return
  const next = id as SizeChartsTabId
  activeTab.value = next
  if (typeof window !== 'undefined') {
    const url = new URL(window.location.href)
    url.hash = `#${next}`
    window.history.replaceState(null, '', url.toString())
  }
}
</script>

<style scoped>
.products-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: #f9fafb;
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
  margin: 0.25rem auto 0;
  width: 100%;
  max-width: none;
}

/* Page-level tabs are handled by PageTabBar. */


.sizecharts-brand-button {
  margin-left: 0.5rem;
  margin-top: 0.5rem;
  padding: 0.25rem 0.8rem;
  border-radius: 9999px;
  border: 1px solid rgba(56, 189, 248, 0.9);
  background-image: linear-gradient(
    135deg,
    rgba(56, 189, 248, 0.9),
    rgba(59, 130, 246, 0.95)
  );
  color: #0b1020;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-align: center;
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
