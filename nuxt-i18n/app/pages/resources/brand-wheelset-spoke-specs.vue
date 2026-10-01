<template>
  <div class="wheelset-spoke-lookup">
    <header class="wheelset-spoke-lookup__header">
      <p class="wheelset-spoke-lookup__eyebrow">{{ t('brandWheelsetSpokeSpecs.eyebrow') }}</p>
      <h1>{{ t('brandWheelsetSpokeSpecs.title') }}</h1>
      <p class="wheelset-spoke-lookup__intro">{{ t('brandWheelsetSpokeSpecs.intro') }}</p>
    </header>

    <section class="wheelset-spoke-lookup__filters" :aria-label="t('brandWheelsetSpokeSpecs.filterTitle')">
      <h2>{{ t('brandWheelsetSpokeSpecs.filterTitle') }}</h2>

      <div class="wheelset-spoke-lookup__brand-strip" role="group" :aria-label="t('brandWheelsetSpokeSpecs.brandLabel')">
        <button
          type="button"
          class="wheelset-spoke-lookup__brand-pill"
          :class="{ 'is-active': !selectedBrand }"
          :aria-pressed="!selectedBrand"
          @click="selectedBrand = ''"
        >
          {{ t('brandWheelsetSpokeSpecs.allBrands') }}
        </button>
        <button
          v-for="brand in brandOptions"
          :key="brand.slug"
          type="button"
          class="wheelset-spoke-lookup__brand-pill"
          :class="{ 'is-active': selectedBrand === brand.slug }"
          :aria-pressed="selectedBrand === brand.slug"
          @click="selectedBrand = brand.slug"
        >
          {{ brand.name }}
        </button>
      </div>

      <label class="wheelset-spoke-lookup__search">
        <span class="wheelset-spoke-lookup__sr-only">{{ t('brandWheelsetSpokeSpecs.searchLabel') }}</span>
        <input
          v-model.trim="searchTerm"
          type="search"
          :placeholder="t('brandWheelsetSpokeSpecs.searchPlaceholder')"
          autocomplete="off"
        >
      </label>
    </section>

    <section
      id="spoke-specs-gated-content"
      class="wheelset-spoke-lookup__gated spoke-specs-gated-content"
      aria-live="polite"
      :aria-label="t('brandWheelsetSpokeSpecs.gatedTitle')"
    >
      <div class="wheelset-spoke-lookup__gated-heading">
        <div>
          <p class="wheelset-spoke-lookup__gated-eyebrow">{{ t('brandWheelsetSpokeSpecs.gatedEyebrow') }}</p>
          <h2>{{ detailModel?.model || t('brandWheelsetSpokeSpecs.gatedTitle') }}</h2>
        </div>
        <span v-if="detailModel" class="wheelset-spoke-lookup__gated-status">{{ t('brandWheelsetSpokeSpecs.specLoaded') }}</span>
      </div>

      <div v-if="detailModel" class="wheelset-spoke-lookup__detail">
        <div class="wheelset-spoke-lookup__detail-meta">
          <span>{{ t('brandWheelsetSpokeSpecs.rimDepth') }}: <strong>{{ detailModel.rim.depthMm }} mm</strong></span>
          <span v-if="detailModel.lifecycleStatus === 'legacy'" class="wheelset-spoke-lookup__legacy-badge">{{ t('brandWheelsetSpokeSpecs.legacyModel') }}</span>
        </div>
        <div class="wheelset-spoke-lookup__detail-grid">
          <article v-for="position in wheelPositions" :key="position" class="wheelset-spoke-lookup__detail-wheel">
            <h3>{{ t(position === 'front' ? 'brandWheelsetSpokeSpecs.frontWheel' : 'brandWheelsetSpokeSpecs.rearWheel') }}</h3>
            <template v-if="detailWheelForPosition(detailModel, position)">
              <p class="wheelset-spoke-lookup__detail-line">
                {{ t('brandWheelsetSpokeSpecs.spokeCount', { count: detailWheelForPosition(detailModel, position)!.spokeCount }) }}
                · {{ compactLacingPattern(detailWheelForPosition(detailModel, position)!.lacingPattern) }}
              </p>
              <p class="wheelset-spoke-lookup__detail-line">
                <strong>{{ spokeSummary(detailWheelForPosition(detailModel, position)!).models.join(' / ') }}</strong>
                · {{ spokeSummary(detailWheelForPosition(detailModel, position)!).headTypes.map(headType => t(headTypeMessageKey(headType))).join(' / ') }}
              </p>
              <p class="wheelset-spoke-lookup__detail-length">
                {{ t('brandWheelsetSpokeSpecs.spokeLength') }}: {{ spokeSummary(detailWheelForPosition(detailModel, position)!).lengthsMm.length ? spokeSummary(detailWheelForPosition(detailModel, position)!).lengthsMm.map(length => length + ' mm').join(' · ') : '—' }}
              </p>
            </template>
          </article>
        </div>
        <div class="wheelset-spoke-lookup__nipple-detail">
          <span>{{ t('brandWheelsetSpokeSpecs.nippleColumn') }}</span>
          <strong>{{ detailModel.nippleModel }}</strong>
          <b v-if="detailModel.nippleLengthMm !== null">{{ detailModel.nippleLengthMm }} mm</b>
          <b v-else>—</b>
        </div>
      </div>

      <div v-else class="wheelset-spoke-lookup__login-prompt">
        <p>{{ selectedModel ? t('brandWheelsetSpokeSpecs.loginPromptForModel', { model: selectedModel.model }) : t('brandWheelsetSpokeSpecs.loginPrompt') }}</p>
        <button type="button" class="wheelset-spoke-lookup__login-button" @click="openAuth('login')">
          {{ t('brandWheelsetSpokeSpecs.loginToView') }}
        </button>
      </div>
      <p v-if="requestError" class="wheelset-spoke-lookup__request-error" role="alert">{{ requestError }}</p>
    </section>

    <section class="wheelset-spoke-lookup__matrix" :aria-label="t('brandWheelsetSpokeSpecs.matrixTitle')">
      <div class="wheelset-spoke-lookup__matrix-heading">
        <h2>{{ t('brandWheelsetSpokeSpecs.matrixTitle') }}</h2>
        <span class="wheelset-spoke-lookup__matrix-count" aria-live="polite">
          {{ t('brandWheelsetSpokeSpecs.resultCount', { count: filteredWheelsets.length }) }}
        </span>
      </div>

      <div class="wheelset-spoke-lookup__table-scroll" tabindex="0" :aria-label="t('brandWheelsetSpokeSpecs.matrixScrollLabel')">
        <table class="wheelset-spoke-lookup__table">
          <caption class="wheelset-spoke-lookup__sr-only">{{ t('brandWheelsetSpokeSpecs.matrixTitle') }}</caption>
          <thead>
            <tr>
              <th scope="col">{{ t('brandWheelsetSpokeSpecs.wheelsetColumn') }}</th>
              <th scope="col">{{ t('brandWheelsetSpokeSpecs.rimDepth') }}</th>
              <th scope="col">{{ t('brandWheelsetSpokeSpecs.holeCrossingColumn') }}</th>
              <th scope="col">{{ t('brandWheelsetSpokeSpecs.frontWheel') }}</th>
              <th scope="col">{{ t('brandWheelsetSpokeSpecs.rearWheel') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="wheelset in paginatedWheelsets"
              :id="'wheelset-' + wheelset.brandSlug + '-' + wheelset.slug"
              :key="wheelset.brandSlug + '-' + wheelset.slug"
            >
              <th scope="row" class="wheelset-spoke-lookup__model-cell">
                <div class="wheelset-spoke-lookup__brand-cell">
                  <span class="wheelset-spoke-lookup__brand-mark" :class="{ 'is-dt-swiss': wheelset.brandSlug === 'dt-swiss' }">{{ brandMark(wheelset.brandName, wheelset.brandSlug) }}</span>
                  <span class="wheelset-spoke-lookup__model-copy">
                    <span class="wheelset-spoke-lookup__model-title">
                      <button
                        type="button"
                        class="wheelset-spoke-lookup__model-select"
                        :class="{ 'is-selected': selectedSlug === wheelset.slug }"
                        :disabled="loadingSlug === wheelset.slug"
                        :aria-busy="loadingSlug === wheelset.slug"
                        :aria-pressed="selectedSlug === wheelset.slug"
                        @click="selectModel(wheelset)"
                      >
                        <strong>{{ wheelset.model }}</strong>
                      </button>
                      <span v-if="wheelset.lifecycleStatus === 'legacy'" class="wheelset-spoke-lookup__legacy-badge">
                        {{ t('brandWheelsetSpokeSpecs.legacyModel') }}
                      </span>
                    </span>
                  </span>
                </div>
              </th>

              <td class="wheelset-spoke-lookup__rim-depth-cell">{{ wheelset.rim.depthMm }} mm</td>

              <td class="wheelset-spoke-lookup__hole-crossing-cell">
                <div v-for="position in wheelPositions" :key="position" class="wheelset-spoke-lookup__hole-crossing-row">
                  <span>{{ t(position === 'front' ? 'brandWheelsetSpokeSpecs.frontWheelShort' : 'brandWheelsetSpokeSpecs.rearWheelShort') }}</span>
                  <strong v-if="wheelForPosition(wheelset, position)">
                    {{ t('brandWheelsetSpokeSpecs.spokeCount', { count: wheelForPosition(wheelset, position)!.spokeCount }) }}
                    · {{ compactLacingPattern(wheelForPosition(wheelset, position)!.lacingPattern) }}
                  </strong>
                  <strong v-else>—</strong>
                </div>
              </td>

              <td v-for="position in wheelPositions" :key="position" class="wheelset-spoke-lookup__spoke-cell">
                <div v-if="wheelForPosition(wheelset, position)" class="wheelset-spoke-lookup__spoke-summary">
                  <strong>{{ spokeSummary(wheelForPosition(wheelset, position)!).models.join(' / ') }}</strong>
                  <span>{{ spokeSummary(wheelForPosition(wheelset, position)!).headTypes.map(headType => t(headTypeMessageKey(headType))).join(' / ') }}</span>
                </div>
                <span v-else class="wheelset-spoke-lookup__empty-value">—</span>
              </td>

            </tr>

            <tr v-if="filteredWheelsets.length === 0" class="wheelset-spoke-lookup__empty-row">
              <td colspan="5"><p role="status">{{ t('brandWheelsetSpokeSpecs.emptyState') }}</p></td>
            </tr>
          </tbody>
        </table>
      </div>

      <nav
        v-if="totalPages > 1"
        class="wheelset-spoke-lookup__pagination"
        :aria-label="t('brandWheelsetSpokeSpecs.pagination.label')"
      >
        <NuxtLink
          v-if="currentPage > 1"
          class="wheelset-spoke-lookup__page-link"
          :to="pageQuery(currentPage - 1)"
          rel="prev"
          :aria-current="undefined"
          :aria-label="t('brandWheelsetSpokeSpecs.pagination.previous')"
        >
          ‹
        </NuxtLink>
        <span
          v-else
          class="wheelset-spoke-lookup__page-link is-disabled"
          aria-disabled="true"
          :aria-label="t('brandWheelsetSpokeSpecs.pagination.previous')"
        >‹</span>

        <template v-for="(pageToken, tokenIndex) in paginationPages" :key="`page-${tokenIndex}-${pageToken}`">
          <span v-if="pageToken === 'ellipsis'" class="wheelset-spoke-lookup__page-ellipsis" aria-hidden="true">…</span>
          <NuxtLink
            v-else
            class="wheelset-spoke-lookup__page-link"
            :class="{ 'is-current': pageToken === currentPage }"
            :to="pageQuery(pageToken)"
            :aria-current="pageToken === currentPage ? 'page' : undefined"
            :aria-label="t('brandWheelsetSpokeSpecs.pagination.goToPage', { page: pageToken })"
          >
            {{ pageToken }}
          </NuxtLink>
        </template>

        <NuxtLink
          v-if="currentPage < totalPages"
          class="wheelset-spoke-lookup__page-link"
          :to="pageQuery(currentPage + 1)"
          rel="next"
          :aria-current="undefined"
          :aria-label="t('brandWheelsetSpokeSpecs.pagination.next')"
        >
          ›
        </NuxtLink>
        <span
          v-else
          class="wheelset-spoke-lookup__page-link is-disabled"
          aria-disabled="true"
          :aria-label="t('brandWheelsetSpokeSpecs.pagination.next')"
        >›</span>

        <span class="wheelset-spoke-lookup__page-summary" aria-live="polite">
          {{ t('brandWheelsetSpokeSpecs.pagination.page', { page: currentPage, totalPages }) }}
        </span>
      </nav>
    </section>

    <LazyAuthModal
      v-if="showAuthModal"
      v-model="showAuthModal"
      :default-mode="authMode"
      embedded
      @mode-change="authMode = $event"
      @success="handleAuthSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  definePageMeta,
  useHead,
  useI18n,
  useRoute,
  useRouter,
  useSwitchLocalePath,
} from '#imports'
import { ApiRequestError, useApiRequest } from '~/composables/useApiRequest'
import { useAuth } from '~/composables/useAuth'
import {
  publicBrandWheelsetSpokeCatalogs,
  type PublicBrandSpokeHeadType,
  type PublicBrandWheelPosition,
  type PublicBrandWheelPositionSpokeSpec,
  type PublicBrandWheelsetRecord,
} from '~/data/brand-wheelset-spoke-specs/brand-wheelset-spoke-specs-public-catalog'
import { usePageMessages } from '~/composables/usePageMessages'
import {
  useStorefrontSeoLinks,
  useStorefrontSeoRouteOverride,
} from '~/composables/seo/useStorefrontSeoLinks'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import localeManifest from '~/i18n/locales.manifest'

type WheelsetEntry = PublicBrandWheelsetRecord & {
  brandSlug: string
  brandName: string
}
type SpokeSide = {
  side: string
  spokeModel: string
  headType: PublicBrandSpokeHeadType
  lengthMm?: number | null
}
type DetailWheel = Omit<PublicBrandWheelPositionSpokeSpec, 'sides'> & {
  sides: SpokeSide[]
}
type WheelsetDetail = Omit<WheelsetEntry, 'wheels'> & {
  wheels: DetailWheel[]
  nippleModel: string
  nippleLengthMm: number | null
}

definePageMeta({
  layout: 'products',
  footerLabelKey: 'brandWheelsetSpokeSpecs.navLabel',
  footerLabelFallback: 'Complete Wheelset Spoke Specs',
})

const searchTerm = ref('')
const selectedBrand = ref('')
const wheelPositions: PublicBrandWheelPosition[] = ['front', 'rear']
const wheelsetPageSize = 10
const { locale, t } = useI18n()
const route = useRoute()
const router = useRouter()
const switchLocalePath = useSwitchLocalePath()
const { canonicalUrl } = useStorefrontSeoLinks()
const { loadPageMessages } = usePageMessages('brandWheelsetSpokeSpecs')
const { request } = useApiRequest()
const auth = useAuth()
const authMode = ref<'login' | 'register'>('login')
const showAuthModal = ref(false)
const pendingSlug = ref('')
const selectedSlug = ref('')
const loadingSlug = ref('')
const requestError = ref('')

await loadPageMessages(locale.value)
watch(locale, nextLocale => void loadPageMessages(nextLocale))

const fallbackWheelsets = publicBrandWheelsetSpokeCatalogs.flatMap(catalog => (
  catalog.wheelsets.map(wheelset => ({
    ...wheelset,
    brandSlug: catalog.brandSlug,
    brandName: catalog.brandName,
  }))
))
const publicWheelsets = ref<WheelsetEntry[]>(fallbackWheelsets)

const unwrapPublicModels = (payload: unknown): WheelsetEntry[] => {
  if (!payload || typeof payload !== 'object') return []
  const root = payload as { data?: { models?: unknown } | unknown; models?: unknown }
  const data = root.data && typeof root.data === 'object' ? root.data as { models?: unknown } : root
  if (!Array.isArray(data.models)) return []
  return data.models.filter((model): model is WheelsetEntry => (
    !!model && typeof model === 'object'
    && typeof (model as WheelsetEntry).slug === 'string'
    && typeof (model as WheelsetEntry).model === 'string'
    && Array.isArray((model as WheelsetEntry).wheels)
  ))
}

try {
  const payload = await request<unknown>('/wheelset-spoke-specs/models', {}, 'Failed to load wheelset model index')
  const models = unwrapPublicModels(payload)
  if (models.length > 0) publicWheelsets.value = models
} catch (_) {
  // The safe public fallback keeps the page usable during a local backend
  // restart. Exact specifications never fall back to client-side data.
}

const brandOptions = computed(() => [...new Map(publicWheelsets.value.map(wheelset => [
  wheelset.brandSlug,
  { slug: wheelset.brandSlug, name: wheelset.brandName },
])).values()])
const normalizedSearch = computed(() => searchTerm.value.toLocaleLowerCase().trim())
const wheelsetEntries = computed(() => publicWheelsets.value)
const filteredWheelsets = computed(() => wheelsetEntries.value.filter((wheelset) => {
  if (selectedBrand.value && wheelset.brandSlug !== selectedBrand.value) return false
  if (!normalizedSearch.value) return true

  const searchable = [
    wheelset.brandName,
    wheelset.model,
    ...wheelset.wheels.flatMap(wheel => wheel.sides.map(side => side.spokeModel)),
  ].join(' ').toLocaleLowerCase()
  return searchable.includes(normalizedSearch.value)
}))

type PaginationToken = number | 'ellipsis'

const requestedPage = computed(() => {
  const rawPage = Array.isArray(route.query.page) ? route.query.page[0] : route.query.page
  if (typeof rawPage !== 'string' || !/^[1-9]\d*$/.test(rawPage)) return 1

  const page = Number(rawPage)
  return Number.isSafeInteger(page) && page > 0 ? page : 1
})
const totalPages = computed(() => Math.max(1, Math.ceil(filteredWheelsets.value.length / wheelsetPageSize)))
const currentPage = computed(() => Math.min(requestedPage.value, totalPages.value))
const paginatedWheelsets = computed(() => {
  const start = (currentPage.value - 1) * wheelsetPageSize
  return filteredWheelsets.value.slice(start, start + wheelsetPageSize)
})
const paginationPages = computed<PaginationToken[]>(() => {
  if (totalPages.value <= 7) {
    return Array.from({ length: totalPages.value }, (_, index) => index + 1)
  }

  const candidates = new Set([1, totalPages.value, currentPage.value])
  if (currentPage.value > 1) candidates.add(currentPage.value - 1)
  if (currentPage.value < totalPages.value) candidates.add(currentPage.value + 1)

  const sorted = [...candidates].sort((left, right) => left - right)
  const result: PaginationToken[] = []
  sorted.forEach((page, index) => {
    const previous = sorted[index - 1]
    if (previous !== undefined && page - previous > 1) result.push('ellipsis')
    result.push(page)
  })
  return result
})
const pageQuery = (page: number) => {
  const query = { ...route.query }
  if (page <= 1) delete query.page
  else query.page = String(page)
  return { query }
}

const supportedSeoLocaleCodes = new Set(['en', 'zh_cn'])
const indexablePaginationPage = computed(() => (
  requestedPage.value > 1 && currentPage.value > 1 ? currentPage.value : null
))
const localizedSeoRoutes = computed(() => localeManifest
  .filter(entry => supportedSeoLocaleCodes.has(entry.code))
  .map(({ code }) => {
    const localizedPath = switchLocalePath(code as any) || '/resources/brand-wheelset-spoke-specs'
    const localizedUrl = new URL(localizedPath, 'https://wheelset-spoke-seo.invalid')
    localizedUrl.search = ''
    if (indexablePaginationPage.value !== null) {
      localizedUrl.searchParams.set('page', String(indexablePaginationPage.value))
    }

    return {
      code,
      path: `${localizedUrl.pathname}${localizedUrl.search}`,
    }
  }))

useStorefrontSeoRouteOverride(localizedSeoRoutes)

watch([selectedBrand, searchTerm], () => {
  if (!import.meta.client || !route.query.page) return
  void router.replace({ query: pageQuery(1).query })
})

watch([totalPages, requestedPage], ([pages, requested]) => {
  if (!import.meta.client || !route.query.page) return

  const normalizedPage = Math.min(requested, pages)
  const currentRoutePage = Array.isArray(route.query.page) ? route.query.page[0] : route.query.page
  const nextQuery = pageQuery(normalizedPage).query
  const nextRoutePage = typeof nextQuery.page === 'string' ? nextQuery.page : ''
  if (nextRoutePage === currentRoutePage) return
  void router.replace({ query: nextQuery })
}, { immediate: true })

const detailBySlug = ref<Record<string, WheelsetDetail>>({})
const selectedModel = computed(() => wheelsetEntries.value.find(wheelset => wheelset.slug === selectedSlug.value) || null)
const detailModel = computed(() => detailBySlug.value[selectedSlug.value] || null)

const wheelForPosition = (
  wheelset: WheelsetEntry,
  position: PublicBrandWheelPosition,
): PublicBrandWheelPositionSpokeSpec | undefined => wheelset.wheels.find(wheel => wheel.position === position)
const detailWheelForPosition = (
  wheelset: WheelsetDetail,
  position: PublicBrandWheelPosition,
): DetailWheel | undefined => wheelset.wheels.find(wheel => wheel.position === position)

const spokeSummary = (wheel: { spokeCount: number; lacingPattern: string; sides: SpokeSide[] }) => ({
  models: [...new Set(wheel.sides.map(side => side.spokeModel))],
  headTypes: [...new Set(wheel.sides.map(side => side.headType))],
  lengthsMm: [...new Set(wheel.sides
    .map(side => side.lengthMm)
    .filter((length): length is number => typeof length === 'number'))].sort((a, b) => a - b),
  spokeCount: wheel.spokeCount,
  lacingPattern: wheel.lacingPattern,
})

const compactLacingPattern = (pattern: string) => [...new Set(
  pattern.split('/').map(value => value.trim()).filter(Boolean),
)].join(' / ')

const headTypeMessageKey = (headType: PublicBrandSpokeHeadType) => {
  const keys: Record<PublicBrandSpokeHeadType, string> = {
    'straight-pull': 'brandWheelsetSpokeSpecs.headTypeStraightPull',
    'j-bend': 'brandWheelsetSpokeSpecs.headTypeJBend',
    unknown: 'brandWheelsetSpokeSpecs.headTypeUnknown',
  }
  return keys[headType]
}

const brandMark = (brandName: string, brandSlug: string) => (
  brandSlug === 'dt-swiss'
    ? 'DT'
    : brandSlug === 'shimano'
      ? 'SH'
      : brandName.split(/\s+/).map(part => part[0]).join('').slice(0, 2).toUpperCase()
)

const openAuth = (mode: 'login' | 'register') => {
  authMode.value = mode
  showAuthModal.value = true
}

const loadModelDetail = async (slug: string) => {
  if (!slug || detailBySlug.value[slug]) return
  loadingSlug.value = slug
  requestError.value = ''
  try {
    const payload = await auth.request<unknown>(`/wheelset-spoke-specs/models/${encodeURIComponent(slug)}`, {
      headers: { Accept: 'application/json' },
    }, 'Failed to load wheelset repair-kit specifications')
    const root = payload && typeof payload === 'object' ? payload as { data?: { model?: unknown } } : {}
    const model = root.data?.model
    if (!model || typeof model !== 'object') throw new Error('The wheelset specification response was empty.')
    detailBySlug.value = { ...detailBySlug.value, [slug]: model as WheelsetDetail }
  } catch (error) {
    if (error instanceof ApiRequestError && error.status === 401) {
      pendingSlug.value = slug
      openAuth('login')
    } else if (error instanceof ApiRequestError && error.status === 429) {
      requestError.value = t('brandWheelsetSpokeSpecs.rateLimited')
    } else {
      requestError.value = error instanceof Error ? error.message : t('brandWheelsetSpokeSpecs.requestFailed')
    }
  } finally {
    loadingSlug.value = ''
  }
}

const selectModel = async (wheelset: WheelsetEntry) => {
  selectedSlug.value = wheelset.slug
  requestError.value = ''
  if (detailBySlug.value[wheelset.slug]) return

  const session = auth.user.value || await auth.ensureSession()
  if (!session) {
    pendingSlug.value = wheelset.slug
    openAuth('login')
    return
  }
  await loadModelDetail(wheelset.slug)
}

const handleAuthSuccess = async () => {
  showAuthModal.value = false
  const slug = pendingSlug.value || selectedSlug.value
  pendingSlug.value = ''
  if (slug) await loadModelDetail(slug)
}

const itemList = computed(() => paginatedWheelsets.value.map((wheelset, index) => ({
  '@type': 'ListItem',
  position: index + 1,
  name: wheelset.brandName + ' ' + wheelset.model,
  url: canonicalUrl.value + '#wheelset-' + wheelset.brandSlug + '-' + wheelset.slug,
})))

const collectionSchema = computed(() => ({
  '@context': 'https://schema.org',
  '@type': 'CollectionPage',
  '@id': canonicalUrl.value + '#collection',
  url: canonicalUrl.value,
  name: t('brandWheelsetSpokeSpecs.seoTitle'),
  description: t('brandWheelsetSpokeSpecs.seoDescription'),
  inLanguage: localeManifest.find(entry => entry.code === locale.value)?.iso || locale.value,
  isAccessibleForFree: false,
  hasPart: {
    '@type': 'WebPageElement',
    '@id': canonicalUrl.value + '#spoke-specs-gated-content',
    cssSelector: '.spoke-specs-gated-content',
    isAccessibleForFree: false,
  },
  mainEntity: {
    '@type': 'ItemList',
    itemListOrder: 'https://schema.org/ItemListOrderAscending',
    numberOfItems: itemList.value.length,
    itemListElement: itemList.value,
  },
}))

const shouldNoIndexQuery = computed(() => {
  const queryKeys = Object.keys(route.query)
  if (queryKeys.length === 0) return false
  if (queryKeys.length !== 1 || queryKeys[0] !== 'page') return true

  const rawPage = Array.isArray(route.query.page) ? route.query.page[0] : route.query.page
  if (typeof rawPage !== 'string' || !/^[1-9]\d*$/.test(rawPage)) return true
  return !Number.isSafeInteger(Number(rawPage))
})

useHead(() => ({
  title: t('brandWheelsetSpokeSpecs.seoTitle'),
  meta: [
    { name: 'description', content: t('brandWheelsetSpokeSpecs.seoDescription'), key: 'description' },
    {
      name: 'robots',
      content: shouldNoIndexQuery.value
        ? 'noindex,follow'
        : (publicWheelsets.value.length > 0 ? 'index,follow' : 'noindex,follow'),
      key: 'robots',
    },
    { property: 'og:title', content: t('brandWheelsetSpokeSpecs.seoTitle'), key: 'og:title' },
    { property: 'og:description', content: t('brandWheelsetSpokeSpecs.seoDescription'), key: 'og:description' },
  ],
  script: [createSeoJsonLdScript(collectionSchema.value)],
}))
</script>

<style scoped>
.wheelset-spoke-lookup {
  --spoke-ink: #17212b;
  --spoke-muted: #64748b;
  --spoke-line: #dbe3ec;
  --spoke-panel: #fff;
  --spoke-wash: #f5f8fc;
  --spoke-blue: #2563eb;
  display: grid;
  gap: 1rem;
  width: 100%;
  max-width: none;
  margin: 0 auto;
  padding: 0 clamp(0.75rem, 1.5vw, 1.25rem) 2.5rem;
  color: var(--spoke-ink);
}

.wheelset-spoke-lookup *,
.wheelset-spoke-lookup *::before,
.wheelset-spoke-lookup *::after {
  box-sizing: border-box;
}

.wheelset-spoke-lookup__header,
.wheelset-spoke-lookup__filters,
.wheelset-spoke-lookup__matrix {
  min-width: 0;
  border: 1px solid var(--spoke-line);
  border-radius: 1.35rem;
  background: var(--spoke-panel);
  box-shadow: 0 8px 24px -18px rgba(15, 23, 42, 0.24);
}

.wheelset-spoke-lookup__header {
  padding: 1.4rem clamp(1rem, 2vw, 1.75rem);
  background:
    radial-gradient(circle at 8% 18%, rgba(37, 99, 235, 0.07), transparent 44%),
    #f8fafc;
}

.wheelset-spoke-lookup__eyebrow {
  margin: 0 0 0.35rem;
  color: var(--spoke-blue);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.1em;
}

.wheelset-spoke-lookup h1 {
  margin: 0;
  font-size: clamp(1.35rem, 2.1vw, 1.8rem);
  font-weight: 850;
  line-height: 1.22;
  letter-spacing: -0.025em;
}

.wheelset-spoke-lookup__intro {
  margin: 0.4rem 0 0;
  color: var(--spoke-muted);
  font-size: 0.86rem;
  line-height: 1.5;
}

.wheelset-spoke-lookup__filters {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(15rem, 25rem);
  align-items: center;
  gap: 0.9rem 1.25rem;
  padding: 1rem 1.2rem;
}

.wheelset-spoke-lookup__filters h2 {
  grid-column: 1 / -1;
  margin: 0;
  font-size: 0.78rem;
  font-weight: 800;
}

.wheelset-spoke-lookup__brand-strip {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.4rem;
  overflow-x: auto;
  padding: 0.1rem 0.1rem 0.3rem;
  scrollbar-width: thin;
}

.wheelset-spoke-lookup__brand-pill {
  flex: 0 0 auto;
  min-height: 2rem;
  padding: 0.38rem 0.75rem;
  border: 1px solid var(--spoke-line);
  border-radius: 999px;
  background: #fff;
  color: #475569;
  font: inherit;
  font-size: 0.7rem;
  font-weight: 700;
  cursor: pointer;
  transition: border-color 0.16s ease, background-color 0.16s ease, color 0.16s ease;
}

.wheelset-spoke-lookup__brand-pill:hover,
.wheelset-spoke-lookup__brand-pill.is-active {
  border-color: rgba(37, 99, 235, 0.35);
  background: rgba(37, 99, 235, 0.08);
  color: #1d4ed8;
}

.wheelset-spoke-lookup__search {
  display: block;
  min-width: 0;
}

.wheelset-spoke-lookup__search input {
  width: 100%;
  min-height: 2.5rem;
  padding: 0.5rem 0.8rem;
  border: 1px solid var(--spoke-line);
  border-radius: 0.75rem;
  background: #fff;
  color: var(--spoke-ink);
  font: inherit;
  font-size: 0.74rem;
}

.wheelset-spoke-lookup__search input:focus-visible,
.wheelset-spoke-lookup__brand-pill:focus-visible,
.wheelset-spoke-lookup__table-scroll:focus-visible {
  outline: 3px solid rgba(37, 99, 235, 0.28);
  outline-offset: 2px;
}

.wheelset-spoke-lookup__matrix {
  overflow: hidden;
}

.wheelset-spoke-lookup__matrix-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem;
  padding: 1rem 1.2rem;
  border-bottom: 1px solid var(--spoke-line);
  background: #fff;
}

.wheelset-spoke-lookup__matrix-heading h2 {
  margin: 0;
  font-size: 0.86rem;
  font-weight: 850;
}

.wheelset-spoke-lookup__matrix-count {
  flex: 0 0 auto;
  color: var(--spoke-muted);
  font-size: 0.7rem;
  font-weight: 700;
}

.wheelset-spoke-lookup__table-scroll {
  max-width: 100%;
  overflow-x: auto;
  outline: none;
}

.wheelset-spoke-lookup__pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
  flex-wrap: wrap;
  padding: 0.8rem 1.2rem 0.95rem;
  border-top: 1px solid var(--spoke-line);
  background: #fff;
}

.wheelset-spoke-lookup__page-link {
  display: inline-grid;
  place-items: center;
  min-width: 2rem;
  min-height: 2rem;
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--spoke-line);
  border-radius: 0.55rem;
  background: #fff;
  color: #475569;
  font-size: 0.7rem;
  font-weight: 800;
  line-height: 1;
  text-decoration: none;
}

.wheelset-spoke-lookup__page-link:hover,
.wheelset-spoke-lookup__page-link.is-current {
  border-color: rgba(37, 99, 235, 0.45);
  background: rgba(37, 99, 235, 0.08);
  color: #1d4ed8;
}

.wheelset-spoke-lookup__page-link.is-disabled {
  color: #cbd5e1;
  cursor: default;
}

.wheelset-spoke-lookup__page-link:focus-visible {
  outline: 3px solid rgba(37, 99, 235, 0.28);
  outline-offset: 2px;
}

.wheelset-spoke-lookup__page-ellipsis {
  min-width: 1.5rem;
  color: var(--spoke-muted);
  text-align: center;
}

.wheelset-spoke-lookup__page-summary {
  margin-left: 0.35rem;
  color: var(--spoke-muted);
  font-size: 0.68rem;
  font-weight: 700;
  white-space: nowrap;
}

.wheelset-spoke-lookup__table {
  width: 100%;
  min-width: 1150px;
  border-collapse: collapse;
  table-layout: fixed;
  text-align: left;
  font-size: 0.72rem;
}

.wheelset-spoke-lookup__table th,
.wheelset-spoke-lookup__table td {
  padding: 0.55rem 0.75rem;
  border-bottom: 1px solid #e8edf3;
  text-align: left;
  vertical-align: middle;
}

.wheelset-spoke-lookup__table thead th {
  padding-top: 0.75rem;
  padding-bottom: 0.75rem;
  background: var(--spoke-wash);
  color: #64748b;
  font-size: 0.64rem;
  font-weight: 800;
  white-space: nowrap;
}

.wheelset-spoke-lookup__table thead th:nth-child(1) { width: 24%; }
.wheelset-spoke-lookup__table thead th:nth-child(2) { width: 10%; }
.wheelset-spoke-lookup__table thead th:nth-child(3) { width: 20%; }
.wheelset-spoke-lookup__table thead th:nth-child(4) { width: 23%; }
.wheelset-spoke-lookup__table thead th:nth-child(5) { width: 23%; }

.wheelset-spoke-lookup__table tbody tr:last-child > * {
  border-bottom: 0;
}

.wheelset-spoke-lookup__table tbody tr:not(.wheelset-spoke-lookup__empty-row):hover {
  background: #f8fafc;
}

.wheelset-spoke-lookup__model-cell {
  font-weight: inherit;
}

.wheelset-spoke-lookup__brand-cell {
  display: flex;
  align-items: flex-start;
  gap: 0.65rem;
  min-width: 0;
}

.wheelset-spoke-lookup__brand-mark {
  flex: 0 0 auto;
  display: grid;
  place-items: center;
  width: 2rem;
  height: 2rem;
  border-radius: 0.6rem;
  background: #1d4ed8;
  color: #fff;
  font-size: 0.58rem;
  font-weight: 850;
}

.wheelset-spoke-lookup__brand-mark.is-dt-swiss {
  background: #e30613;
}

.wheelset-spoke-lookup__model-copy {
  display: grid;
  min-width: 0;
  gap: 0.2rem;
}

.wheelset-spoke-lookup__model-title {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.4rem;
  min-width: 0;
}

.wheelset-spoke-lookup__model-select {
  min-width: 0;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
}

.wheelset-spoke-lookup__model-select:hover strong,
.wheelset-spoke-lookup__model-select.is-selected strong {
  color: var(--spoke-blue);
  text-decoration: underline;
  text-underline-offset: 0.14em;
}

.wheelset-spoke-lookup__model-select:disabled {
  cursor: wait;
  opacity: 0.68;
}

.wheelset-spoke-lookup__legacy-badge {
  flex: 0 0 auto;
  padding: 0.16rem 0.38rem;
  border: 1px solid #fed7aa;
  border-radius: 999px;
  background: #fff7ed;
  color: #9a3412;
  font-size: 0.57rem;
  font-weight: 800;
  line-height: 1.25;
  white-space: nowrap;
}

.wheelset-spoke-lookup__model-copy strong,
.wheelset-spoke-lookup__nipple-cell strong,
.wheelset-spoke-lookup__spoke-summary strong {
  overflow-wrap: anywhere;
  color: var(--spoke-ink);
  font-size: 0.76rem;
  font-weight: 800;
  line-height: 1.4;
}

.wheelset-spoke-lookup__model-copy small {
  color: var(--spoke-muted);
  font-size: 0.65rem;
  line-height: 1.4;
}

.wheelset-spoke-lookup__rim-depth-cell {
  color: #334155;
  font-size: 0.78rem;
  font-weight: 800;
  white-space: nowrap;
}

.wheelset-spoke-lookup__hole-crossing-cell {
  display: grid;
  gap: 0.26rem;
}

.wheelset-spoke-lookup__hole-crossing-row {
  display: flex;
  align-items: baseline;
  gap: 0.45rem;
  white-space: nowrap;
}

.wheelset-spoke-lookup__hole-crossing-row > span {
  min-width: 2.5rem;
  color: var(--spoke-muted);
  font-size: 0.63rem;
}

.wheelset-spoke-lookup__hole-crossing-row > strong {
  color: #334155;
  font-size: 0.68rem;
  font-weight: 800;
}

.wheelset-spoke-lookup__spoke-summary {
  display: grid;
  gap: 0.12rem;
  min-width: 0;
}

.wheelset-spoke-lookup__spoke-summary > span,
.wheelset-spoke-lookup__nipple-cell > span {
  color: #475569;
  font-size: 0.69rem;
  font-weight: 700;
  line-height: 1.45;
}

.wheelset-spoke-lookup__nipple-cell strong,
.wheelset-spoke-lookup__nipple-cell > span {
  display: block;
}

.wheelset-spoke-lookup__nipple-cell > span {
  margin-top: 0.18rem;
  color: var(--spoke-blue);
  font-size: 0.72rem;
  font-weight: 800;
}

.wheelset-spoke-lookup__login-button {
  min-height: 2.15rem;
  padding: 0.42rem 0.7rem;
  border: 1px solid rgba(37, 99, 235, 0.35);
  border-radius: 0.65rem;
  background: rgba(37, 99, 235, 0.08);
  color: #1d4ed8;
  font: inherit;
  font-size: 0.67rem;
  font-weight: 800;
  line-height: 1.3;
  cursor: pointer;
  transition: border-color 0.16s ease, background-color 0.16s ease, color 0.16s ease;
}

.wheelset-spoke-lookup__login-button:hover {
  border-color: #1d4ed8;
  background: #1d4ed8;
  color: #fff;
}

.wheelset-spoke-lookup__gated {
  min-width: 0;
  border: 1px solid var(--spoke-line);
  border-radius: 1.35rem;
  background: var(--spoke-panel);
  box-shadow: 0 8px 24px -18px rgba(15, 23, 42, 0.24);
  overflow: hidden;
}

.wheelset-spoke-lookup__gated-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem;
  padding: 1rem 1.2rem;
  border-bottom: 1px solid var(--spoke-line);
  background: var(--spoke-wash);
}

.wheelset-spoke-lookup__gated-heading h2 {
  margin: 0;
  color: var(--spoke-ink);
  font-size: 0.95rem;
  font-weight: 850;
}

.wheelset-spoke-lookup__gated-eyebrow {
  margin: 0 0 0.2rem;
  color: var(--spoke-blue);
  font-size: 0.62rem;
  font-weight: 850;
  letter-spacing: 0.08em;
}

.wheelset-spoke-lookup__gated-status {
  flex: 0 0 auto;
  color: #166534;
  font-size: 0.68rem;
  font-weight: 800;
}

.wheelset-spoke-lookup__login-prompt {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.2rem 1.15rem;
}

.wheelset-spoke-lookup__login-prompt p {
  max-width: 48rem;
  margin: 0;
  color: var(--spoke-muted);
  font-size: 0.78rem;
  line-height: 1.55;
}

.wheelset-spoke-lookup__login-button {
  flex: 0 0 auto;
  background: #1d4ed8;
  color: #fff;
}

.wheelset-spoke-lookup__detail {
  padding: 1rem 1.2rem 1.2rem;
}

.wheelset-spoke-lookup__detail-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.55rem;
  color: var(--spoke-muted);
  font-size: 0.72rem;
}

.wheelset-spoke-lookup__detail-meta strong {
  color: var(--spoke-ink);
}

.wheelset-spoke-lookup__detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  margin-top: 0.85rem;
}

.wheelset-spoke-lookup__detail-wheel {
  min-width: 0;
  padding: 0.8rem;
  border: 1px solid var(--spoke-line);
  border-radius: 0.8rem;
  background: #fbfdff;
}

.wheelset-spoke-lookup__detail-wheel h3 {
  margin: 0 0 0.35rem;
  color: var(--spoke-ink);
  font-size: 0.74rem;
  font-weight: 850;
}

.wheelset-spoke-lookup__detail-line,
.wheelset-spoke-lookup__detail-length {
  margin: 0.2rem 0 0;
  color: #475569;
  font-size: 0.7rem;
  line-height: 1.45;
}

.wheelset-spoke-lookup__detail-line strong {
  color: var(--spoke-ink);
}

.wheelset-spoke-lookup__detail-length {
  color: var(--spoke-blue);
  font-weight: 800;
}

.wheelset-spoke-lookup__nipple-detail {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 0.35rem 0.55rem;
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--spoke-line);
  color: var(--spoke-muted);
  font-size: 0.7rem;
}

.wheelset-spoke-lookup__nipple-detail strong {
  color: var(--spoke-ink);
}

.wheelset-spoke-lookup__nipple-detail b {
  color: var(--spoke-blue);
}

.wheelset-spoke-lookup__request-error {
  margin: 0;
  padding: 0 1.2rem 1rem;
  color: #b91c1c;
  font-size: 0.72rem;
}

.wheelset-spoke-lookup__empty-value {
  color: #94a3b8;
}

.wheelset-spoke-lookup__empty-row td {
  padding: 2.5rem 1rem;
  color: var(--spoke-muted);
  text-align: center;
}

.wheelset-spoke-lookup__empty-row p {
  margin: 0;
}

.wheelset-spoke-lookup__sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  clip-path: inset(50%);
}

@media (max-width: 920px) {
  .wheelset-spoke-lookup__filters {
    grid-template-columns: minmax(0, 1fr);
  }

  .wheelset-spoke-lookup__filters h2 {
    grid-column: auto;
  }

  .wheelset-spoke-lookup__search {
    grid-row: 2;
  }
}

@media (max-width: 680px) {
  .wheelset-spoke-lookup {
    gap: 0.8rem;
    padding: 0 0.55rem 1.8rem;
  }

  .wheelset-spoke-lookup__header,
  .wheelset-spoke-lookup__filters {
    padding: 0.9rem;
    border-radius: 1rem;
  }

  .wheelset-spoke-lookup__matrix {
    border-radius: 1rem;
  }

  .wheelset-spoke-lookup__matrix-heading {
    padding: 0.85rem 0.9rem;
  }

  .wheelset-spoke-lookup__gated {
    border-radius: 1rem;
  }

  .wheelset-spoke-lookup__gated-heading,
  .wheelset-spoke-lookup__detail,
  .wheelset-spoke-lookup__login-prompt {
    padding-left: 0.9rem;
    padding-right: 0.9rem;
  }

  .wheelset-spoke-lookup__login-prompt {
    align-items: stretch;
    flex-direction: column;
  }

  .wheelset-spoke-lookup__detail-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .wheelset-spoke-lookup__matrix-heading h2 {
    font-size: 0.76rem;
  }

  .wheelset-spoke-lookup__table {
    min-width: 1150px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .wheelset-spoke-lookup__brand-pill {
    transition-duration: 0.01ms;
  }
}
</style>
