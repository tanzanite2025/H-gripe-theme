<template>
  <div class="how-to-choose-guide">
    <!-- Main Premium Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] p-5 text-center shadow-md md:p-6">
      <h2 class="mb-6 flex items-center justify-center gap-2 text-xl font-bold tz-text-secondary">
        {{ t('guidesTireChoose.title') }}
      </h2>

      <!-- Interactive Helper -->
      <div class="mb-8">
        <TireRimHelper />
      </div>

      <div class="mt-8 border-t tz-border-subtle pt-8">
        <div class="mb-6 flex items-center justify-center gap-2">
          <span class="h-px w-8 tz-surface-panel"></span>
          <h3 class="text-lg font-bold uppercase tracking-wider tz-text-primary">{{ t('guidesTireChoose.standards.title') }}</h3>
          <span class="h-px w-8 tz-surface-panel"></span>
        </div>

        <p class="mx-auto mb-3 max-w-3xl text-sm tz-text-secondary">
          {{ t('guidesTireChoose.standards.description') }}
        </p>

        <div class="mb-8 flex flex-wrap items-center justify-center gap-x-5 gap-y-2 text-xs tz-text-muted">
          <span class="inline-flex items-center gap-2">
            <span class="tire-chart-legend__swatch tire-chart-legend__swatch--recommended"></span>
            {{ t('guidesTireChoose.standards.recommended') }}
          </span>
          <span class="inline-flex items-center gap-2">
            <span class="tire-chart-legend__swatch tire-chart-legend__swatch--possible"></span>
            {{ t('guidesTireChoose.standards.possible') }}
          </span>
        </div>
        <p class="mx-auto mb-6 max-w-3xl text-xs tz-text-muted">
          {{ t('guidesTireChoose.standards.legendDescription') }}
        </p>

        <section
          class="tire-chart-methodology mx-auto mb-6 max-w-3xl text-left"
          aria-labelledby="tire-chart-methodology-title"
        >
          <h4 id="tire-chart-methodology-title" class="tire-chart-methodology__title">
            {{ t('guidesTireChoose.methodology.title') }}
          </h4>
          <p v-if="tireRimWidthReferenceMetadata?.source" class="tire-chart-methodology__source">
            {{ t('guidesTireChoose.methodology.source', {
              source: tireRimWidthReferenceMetadata.source.name,
              date: tireRimWidthReferenceMetadata.knowledge_as_of,
            }) }}
          </p>
          <p v-else class="tire-chart-methodology__source">
            {{ t('guidesTireChoose.methodology.unavailable') }}
          </p>
          <p v-if="tireRimWidthReferenceMetadata?.source" class="tire-chart-methodology__provenance">
            {{ t('guidesTireChoose.methodology.provenance', {
              provenance: tireRimWidthReferenceMetadata.source.provenance,
            }) }}
          </p>
          <p class="tire-chart-methodology__body">
            {{ t('guidesTireChoose.methodology.body') }}
          </p>
          <p v-if="tireRimWidthReferenceMetadata?.model_version" class="tire-chart-methodology__version">
            {{ t('guidesTireChoose.methodology.version', {
              modelVersion: tireRimWidthReferenceMetadata.model_version,
            }) }}
          </p>
          <ul class="tire-chart-methodology__limitations">
            <li
              v-for="limitationKey in tireRimWidthReferenceLimitationKeys"
              :key="limitationKey"
            >
              {{ t(`guidesTireChoose.methodology.limitations.${limitationKey}`) }}
            </li>
          </ul>
        </section>

        <div id="tire-chart-tables" class="tire-chart-table-grid">
          <section class="tire-chart-table-panel" aria-labelledby="hookless-chart-table-title">
            <h4 id="hookless-chart-table-title" class="tire-chart-table-panel__title">
              {{ t('guidesTireChoose.standards.hooklessTitle') }}
            </h4>
            <div class="tire-chart-table-scroll">
              <table class="tire-chart-table">
                <caption>
                  {{ t('guidesTireChoose.standards.captionHookless') }}
                </caption>
                <thead>
                  <tr>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.tireWidthMm') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.tireWidthInch') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.recommendedRim') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.possibleRim') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in hooklessChartRows" :key="`hookless-${row.tire_width_mm}`">
                    <th scope="row">{{ row.tire_width_mm }}</th>
                    <td>{{ row.inch }}</td>
                    <td class="tire-chart-table__recommended">
                      {{ formatWidthList(row.recommended) }}
                    </td>
                    <td class="tire-chart-table__possible">
                      {{ formatWidthList(row.possible) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section class="tire-chart-table-panel" aria-labelledby="hooked-chart-table-title">
            <h4 id="hooked-chart-table-title" class="tire-chart-table-panel__title">
              {{ t('guidesTireChoose.standards.hookedTitle') }}
            </h4>
            <div class="tire-chart-table-scroll">
              <table class="tire-chart-table">
                <caption>
                  {{ t('guidesTireChoose.standards.captionHooked') }}
                </caption>
                <thead>
                  <tr>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.tireWidthMm') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.tireWidthInch') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.recommendedRim') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.possibleRim') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in hookedChartRows" :key="`hooked-${row.tire_width_mm}`">
                    <th scope="row">{{ row.tire_width_mm }}</th>
                    <td>{{ row.inch }}</td>
                    <td class="tire-chart-table__recommended">
                      {{ formatWidthList(row.recommended) }}
                    </td>
                    <td class="tire-chart-table__possible">
                      {{ formatWidthList(row.possible) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </div>

        <h4 class="mt-10 mb-4 text-left text-sm font-semibold uppercase tracking-wider tz-text-primary">
          {{ t('guidesTireChoose.images.sourceTitle') }}
        </h4>
        <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
          <figure class="overflow-hidden rounded-xl border tz-border-subtle shadow-md transition-colors hover:tz-border-subtle">
            <img
              src="/public/tiresizecharts/howtochoose/dtswiss-hookless-tss-rim-table.webp"
              :alt="t('guidesTireChoose.images.hooklessAlt')"
              class="block h-auto w-full"
              loading="lazy"
            />
            <figcaption class="tz-surface-panel py-2 text-xs tracking-wider tz-text-muted">
              {{ t('guidesTireChoose.images.hooklessCaption') }}
            </figcaption>
          </figure>
          <figure class="overflow-hidden rounded-xl border tz-border-subtle shadow-md transition-colors hover:tz-border-subtle">
            <img
              src="/public/tiresizecharts/howtochoose/dtswiss-hooked-tc-rim-table.webp"
              :alt="t('guidesTireChoose.images.hookedAlt')"
              class="block h-auto w-full"
              loading="lazy"
            />
            <figcaption class="tz-surface-panel py-2 text-xs tracking-wider tz-text-muted">
              {{ t('guidesTireChoose.images.hookedCaption') }}
            </figcaption>
          </figure>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useAsyncData, useHead, useI18n, useSwitchLocalePath } from '#imports'
import TireRimHelper from '~/components/TireRimHelper.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import {
  useStorefrontSeoLinks,
  useStorefrontSeoRouteOverride,
} from '~/composables/seo/useStorefrontSeoLinks'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import localeManifest from '~/i18n/locales.manifest'
import {
  formatTireRimWidthReferenceRanges,
  type TireRimWidthRange,
  type TireRimWidthReferenceMatrixRow,
  type TireRimWidthReferenceMetadata,
} from '~/data/tireguides/tireRimWidthReferencePresentation'
import { useApiRequest } from '~/composables/useApiRequest'

const { locale, t } = useI18n()
const switchLocalePath = useSwitchLocalePath()
const { canonicalUrl } = useStorefrontSeoLinks()
const { loadPageMessages } = usePageMessages('guidesTireChoose')

const tireRimReferencePublishedLocaleCodes = ['en', 'zh_cn'] as const
const tireRimWidthReferenceLimitationKeys = [
  'modelSpecificCertification',
  'possibleReferenceStatus',
  'interpolationProjection',
  'displayOnlyMetrics',
] as const
const localizedTireRimReferenceSeoRoutes = computed(() => (
  tireRimReferencePublishedLocaleCodes.map((code) => {
    const localizedPath = switchLocalePath(code as any)
    return {
      code,
      path: localizedPath || (code === 'zh_cn'
        ? '/zh_cn/guides/tireguides/choose'
        : '/guides/tireguides/choose'),
    }
  })
))

// Only the locales with dedicated page copy are indexable. The route remains
// reachable in other locale shells, but those fallback copies must not create
// duplicate search results or asymmetric hreflang declarations.
useStorefrontSeoRouteOverride(localizedTireRimReferenceSeoRoutes)
const isPublishedTireRimReferenceLocale = computed(() => (
  tireRimReferencePublishedLocaleCodes.includes(locale.value as (typeof tireRimReferencePublishedLocaleCodes)[number])
))

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const { request } = useApiRequest()
const { data: tireRimWidthReferenceMetadata } = await useAsyncData<TireRimWidthReferenceMetadata>(
  'tire-rim-width-reference-metadata',
  async () => {
    const response = await request<{ data?: TireRimWidthReferenceMetadata }>(
      '/engineering/tire-rim/matrix',
      {},
      'Tire/rim reference matrix is temporarily unavailable',
    )
    if (!response.data) throw new Error('Tire/rim reference matrix response is missing data')
    return response.data
  },
  { default: () => ({ rows: [] } as unknown as TireRimWidthReferenceMetadata) },
)

const referenceRows = computed<TireRimWidthReferenceMatrixRow[]>(() => (
  tireRimWidthReferenceMetadata.value?.rows || []
))
const hooklessChartRows = computed(() => referenceRows.value.filter(row => row.rim_system === 'hookless'))
const hookedChartRows = computed(() => referenceRows.value.filter(row => row.rim_system === 'hooked'))

const formatWidthList = (widths: TireRimWidthRange[]) =>
  widths.length > 0
    ? `${formatTireRimWidthReferenceRanges(widths)} mm`
    : t('guidesTireChoose.helper.noneShown')
const tireRimReferenceSchema = computed(() => {
  const metadata = tireRimWidthReferenceMetadata.value
  const language = localeManifest.find(entry => entry.code === locale.value)?.iso || locale.value
  const datasetId = canonicalUrl.value + '#tire-rim-width-reference-dataset'
  const measurementTechnique = metadata?.methodology
    ? [
        metadata.methodology.exact_rows,
        metadata.methodology.interpolated_rows,
        metadata.methodology.possible_rows,
        metadata.methodology.derived_metrics,
      ].filter(Boolean).join(' ')
    : t('guidesTireChoose.seo.measurementTechnique')

  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'TechArticle',
        '@id': canonicalUrl.value + '#tire-rim-width-reference',
        url: canonicalUrl.value,
        mainEntityOfPage: canonicalUrl.value,
        headline: t('guidesTireChoose.seo.title'),
        description: t('guidesTireChoose.seo.description'),
        articleSection: t('guidesTireChoose.seo.articleSection'),
        articleBody: t('guidesTireChoose.seo.articleBody'),
        proficiencyLevel: 'Expert',
        author: {
          '@type': 'Organization',
          name: 'Tanzanite Engineering Laboratory',
        },
        inLanguage: language,
        ...(metadata?.model_version ? { version: metadata.model_version } : {}),
        ...(metadata?.knowledge_as_of ? { dateModified: metadata.knowledge_as_of } : {}),
        hasPart: { '@id': datasetId },
      },
      {
        '@type': 'Dataset',
        '@id': datasetId,
        url: canonicalUrl.value + '#tire-chart-tables',
        name: t('guidesTireChoose.seo.datasetName'),
        description: t('guidesTireChoose.seo.datasetDescription'),
        ...(metadata?.model_version ? { version: metadata.model_version } : {}),
        inLanguage: language,
        isAccessibleForFree: true,
        ...(metadata?.knowledge_as_of ? { dateModified: metadata.knowledge_as_of } : {}),
        ...(metadata?.rows?.length ? { numberOfItems: metadata.rows.length } : {}),
        measurementTechnique,
        variableMeasured: [
          t('guidesTireChoose.seo.tireWidthVariable'),
          t('guidesTireChoose.seo.rimWidthVariable'),
        ],
        ...(metadata?.source
          ? {
              additionalProperty: [
                {
                  '@type': 'PropertyValue',
                  name: t('guidesTireChoose.seo.sourceProperty'),
                  value: metadata.source.name,
                },
                {
                  '@type': 'PropertyValue',
                  name: t('guidesTireChoose.seo.sourceDateProperty'),
                  value: metadata.knowledge_as_of,
                },
                {
                  '@type': 'PropertyValue',
                  name: 'source_provenance',
                  value: metadata.source.provenance,
                },
              ],
            }
          : {}),
      },
    ],
  }
})

useHead(() => ({
  title: t('guidesTireChoose.seo.title'),
  meta: [
    {
      name: 'description',
      content: t('guidesTireChoose.seo.description'),
      key: 'description',
    },
    ...(!isPublishedTireRimReferenceLocale.value
      ? [{ name: 'robots', content: 'noindex,follow', key: 'robots' }]
      : []),
  ],
  script: [createSeoJsonLdScript(tireRimReferenceSchema.value)],
}))
</script>

<style scoped>
.tire-chart-legend__swatch {
  display: inline-block;
  height: 0.75rem;
  width: 0.75rem;
  border-radius: 0.15rem;
}

.tire-chart-legend__swatch--recommended {
  background: var(--tz-text-accent);
}

.tire-chart-legend__swatch--possible {
  background: var(--tz-text-secondary);
}

.tire-chart-table-panel {
  min-width: 0;
}

.tire-chart-methodology {
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 0.75rem;
  padding: 0.85rem 1rem;
  background: var(--tz-form-panel-surface);
}

.tire-chart-methodology__title {
  margin: 0 0 0.35rem;
  color: var(--tz-text-primary);
  font-size: 0.8rem;
  font-weight: 700;
}

.tire-chart-methodology__source,
.tire-chart-methodology__provenance,
.tire-chart-methodology__body,
.tire-chart-methodology__version {
  margin: 0;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  line-height: 1.55;
}

.tire-chart-methodology__limitations {
  display: grid;
  gap: 0.2rem;
  margin: 0.45rem 0 0;
  padding-left: 1rem;
  color: var(--tz-text-muted);
  font-size: 0.7rem;
  line-height: 1.5;
}

.tire-chart-table-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1.5rem;
  align-items: start;
  text-align: left;
}

.tire-chart-table-panel__title {
  margin: 0 0 0.65rem;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.tire-chart-table-scroll {
  max-width: 100%;
  overflow-x: auto;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 0.5rem;
}

.tire-chart-table {
  min-width: 34rem;
  width: 100%;
  border-collapse: collapse;
  font-size: 0.78rem;
  table-layout: fixed;
}

.tire-chart-table caption {
  padding: 0.75rem 0.85rem;
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  line-height: 1.45;
  text-align: left;
}

.tire-chart-table th,
.tire-chart-table td {
  border-top: 1px solid rgba(148, 163, 184, 0.14);
  padding: 0.5rem 0.6rem;
  text-align: left;
  vertical-align: top;
  overflow-wrap: anywhere;
}

.tire-chart-table thead th {
  background: rgba(15, 23, 42, 0.65);
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  font-weight: 700;
  line-height: 1.35;
}

.tire-chart-table th:nth-child(1),
.tire-chart-table td:nth-child(1) {
  width: 18%;
}

.tire-chart-table th:nth-child(2),
.tire-chart-table td:nth-child(2) {
  width: 17%;
}

.tire-chart-table th:nth-child(3),
.tire-chart-table td:nth-child(3) {
  width: 31%;
}

.tire-chart-table th:nth-child(4),
.tire-chart-table td:nth-child(4) {
  width: 34%;
}

.tire-chart-table tbody th {
  color: var(--tz-text-primary);
  font-weight: 700;
  white-space: nowrap;
}

.tire-chart-table tbody td {
  color: var(--tz-text-secondary);
}

.tire-chart-table__recommended {
  color: var(--tz-text-accent) !important;
  font-weight: 600;
}

.tire-chart-table__possible {
  color: var(--tz-text-secondary) !important;
}

@media (min-width: 1024px) {
  .tire-chart-table-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .tire-chart-table {
    min-width: 0;
  }

  .tire-chart-table caption {
    min-height: 3.35rem;
  }
}
</style>
