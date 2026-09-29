<template>
  <div class="wheelset-spoke-lookup">
    <header class="wheelset-spoke-lookup__header">
      <p class="wheelset-spoke-lookup__eyebrow">
        {{ t('brandWheelsetSpokeSpecs.eyebrow') }}
      </p>
      <div class="wheelset-spoke-lookup__title-row">
        <div>
          <h1>{{ t('brandWheelsetSpokeSpecs.title') }}</h1>
          <p class="wheelset-spoke-lookup__intro">
            {{ t('brandWheelsetSpokeSpecs.intro') }}
          </p>
        </div>
      </div>
      <aside class="wheelset-spoke-lookup__review-note" role="note">
        <strong>{{ t('brandWheelsetSpokeSpecs.reviewPendingTitle') }}</strong>
        <p>{{ t('brandWheelsetSpokeSpecs.reviewPendingBody') }}</p>
      </aside>
    </header>

    <section class="wheelset-spoke-lookup__controls" aria-labelledby="wheelset-search-label">
      <div class="wheelset-spoke-lookup__search">
        <label id="wheelset-search-label" for="wheelset-spoke-search">
          {{ t('brandWheelsetSpokeSpecs.searchLabel') }}
        </label>
        <div class="wheelset-spoke-lookup__search-row">
          <input
            id="wheelset-spoke-search"
            v-model.trim="searchTerm"
            type="search"
            :placeholder="t('brandWheelsetSpokeSpecs.searchPlaceholder')"
            autocomplete="off"
          >
          <button
            v-if="searchTerm"
            type="button"
            class="wheelset-spoke-lookup__clear"
            @click="searchTerm = ''"
          >
            {{ t('brandWheelsetSpokeSpecs.clearSearch') }}
          </button>
        </div>
      </div>
      <div v-if="brandOptions.length > 1" class="wheelset-spoke-lookup__brand-filter">
        <label for="wheelset-spoke-brand">
          {{ t('brandWheelsetSpokeSpecs.brandLabel') }}
        </label>
        <select id="wheelset-spoke-brand" v-model="selectedBrand">
          <option value="">{{ t('brandWheelsetSpokeSpecs.allBrands') }}</option>
          <option
            v-for="brand in brandOptions"
            :key="brand.slug"
            :value="brand.slug"
          >
            {{ brand.name }}
          </option>
        </select>
      </div>
      <p class="wheelset-spoke-lookup__count" aria-live="polite">
        {{ t('brandWheelsetSpokeSpecs.resultCount', { count: filteredWheelsets.length }) }}
      </p>
    </section>

    <section
      class="wheelset-spoke-lookup__grid"
      :aria-label="t('brandWheelsetSpokeSpecs.title')"
    >
      <article
        v-for="wheelset in filteredWheelsets"
        :id="`wheelset-${wheelset.brandSlug}-${wheelset.slug}`"
        :key="`${wheelset.brandSlug}-${wheelset.slug}`"
        class="wheelset-spoke-card"
      >
        <header class="wheelset-spoke-card__header">
          <div>
            <p class="wheelset-spoke-card__brand">{{ wheelset.brandName }}</p>
            <h2>{{ wheelset.model }}</h2>
            <p class="wheelset-spoke-card__years">
              {{ t('brandWheelsetSpokeSpecs.modelYears', { years: wheelset.modelYears }) }}
            </p>
          </div>
        </header>

        <div class="wheelset-spoke-card__positions">
          <section
            v-for="wheel in wheelset.wheels"
            :key="wheel.position"
            class="wheelset-spoke-card__position"
          >
            <div class="wheelset-spoke-card__position-heading">
              <div>
                <h3>{{ wheelPositionLabel(wheel.position) }}</h3>
                <p>
                  {{ t('brandWheelsetSpokeSpecs.spokeCount', { count: wheel.spokeCount }) }}
                  · {{ t('brandWheelsetSpokeSpecs.lacing') }} {{ wheel.lacingPattern }}
                </p>
              </div>
            </div>

            <dl class="wheelset-spoke-card__lengths">
              <div
                v-for="side in wheel.sides"
                :key="side.side"
                class="wheelset-spoke-card__length"
              >
                <dt>{{ spokeSideLabel(side.side) }}</dt>
                <dd>{{ side.lengthMm }} mm</dd>
              </div>
            </dl>

            <p class="wheelset-spoke-card__type">
              <span>{{ t('brandWheelsetSpokeSpecs.type') }}</span>
              <strong>{{ wheel.spokeType }}</strong>
            </p>
          </section>
        </div>
      </article>

      <p
        v-if="filteredWheelsets.length === 0"
        class="wheelset-spoke-lookup__empty"
        role="status"
      >
        {{ t('brandWheelsetSpokeSpecs.emptyState') }}
      </p>
    </section>

    <footer class="wheelset-spoke-lookup__footer">
      <span>{{ t('brandWheelsetSpokeSpecs.calculatorLink') }}</span>
      <NuxtLink :to="localePath('/resources/spoke-calculator')">
        {{ t('brandWheelsetSpokeSpecs.calculatorLinkLabel') }}
      </NuxtLink>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  definePageMeta,
  useHead,
  useI18n,
  useLocalePath,
  useSwitchLocalePath,
} from '#imports'
import { brandWheelsetSpokeCatalogs } from '~/data/brand-wheelset-spoke-specs/catalog'
import { usePageMessages } from '~/composables/usePageMessages'
import {
  useStorefrontSeoLinks,
  useStorefrontSeoRouteOverride,
} from '~/composables/seo/useStorefrontSeoLinks'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import localeManifest from '~/i18n/locales.manifest'
import type {
  BrandWheelSpokeSide,
  BrandWheelPosition,
} from '~/data/brand-wheelset-spoke-specs/catalog'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'brandWheelsetSpokeSpecs.navLabel',
  footerLabelFallback: 'Complete Wheelset Spoke Specs',
})

const searchTerm = ref('')
const selectedBrand = ref('')
const { locale, t } = useI18n()
const localePath = useLocalePath()
const switchLocalePath = useSwitchLocalePath()
const { canonicalUrl } = useStorefrontSeoLinks()
const { loadPageMessages } = usePageMessages('brandWheelsetSpokeSpecs')

await loadPageMessages(locale.value)
watch(locale, nextLocale => void loadPageMessages(nextLocale))

const supportedSeoLocaleCodes = new Set(['en', 'zh_cn'])
const localizedSeoRoutes = computed(() => localeManifest
  .filter(entry => supportedSeoLocaleCodes.has(entry.code))
  .map(({ code }) => ({
    code,
    path: switchLocalePath(code as any) || '/resources/brand-wheelset-spoke-specs',
  })))

useStorefrontSeoRouteOverride(localizedSeoRoutes)

const isIndexable = computed(() => (
  brandWheelsetSpokeCatalogs.length > 0
  && brandWheelsetSpokeCatalogs.every(catalog => (
    catalog.publicationStatus === 'published'
    && catalog.wheelsets.length > 0
    && catalog.wheelsets.every(wheelset => wheelset.verificationStatus === 'verified')
  ))
))

const brandOptions = computed(() => brandWheelsetSpokeCatalogs
  .filter(catalog => catalog.wheelsets.length > 0)
  .map(catalog => ({
  slug: catalog.brandSlug,
  name: catalog.brandName,
  })))

const wheelsetEntries = computed(() => brandWheelsetSpokeCatalogs.flatMap(catalog => (
  catalog.wheelsets.map(wheelset => ({
    ...wheelset,
    brandSlug: catalog.brandSlug,
    brandName: catalog.brandName,
  }))
)))

const normalizedSearch = computed(() => searchTerm.value.toLocaleLowerCase().trim())
const filteredWheelsets = computed(() => {
  return wheelsetEntries.value.filter((wheelset) => {
    if (selectedBrand.value && wheelset.brandSlug !== selectedBrand.value) return false
    if (!normalizedSearch.value) return true

    const searchable = [
      wheelset.brandName,
      wheelset.model,
      wheelset.modelYears,
      ...wheelset.wheels.map(wheel => wheel.spokeType),
    ].join(' ').toLocaleLowerCase()
    return searchable.includes(normalizedSearch.value)
  })
})

const wheelPositionLabel = (position: BrandWheelPosition) => t(
  position === 'front'
    ? 'brandWheelsetSpokeSpecs.frontWheel'
    : 'brandWheelsetSpokeSpecs.rearWheel',
)

const spokeSideLabel = (side: BrandWheelSpokeSide) => {
  const labels = {
    left: 'brandWheelsetSpokeSpecs.left',
    right: 'brandWheelsetSpokeSpecs.right',
    drive: 'brandWheelsetSpokeSpecs.driveSide',
    nonDrive: 'brandWheelsetSpokeSpecs.nonDriveSide',
  } as const
  return t(labels[side])
}

const itemList = computed(() => wheelsetEntries.value.map((wheelset, index) => ({
  '@type': 'ListItem',
  position: index + 1,
  name: wheelset.brandName + ' ' + wheelset.model,
  url: canonicalUrl.value + `#wheelset-${wheelset.brandSlug}-${wheelset.slug}`,
})))

const collectionSchema = computed(() => ({
  '@context': 'https://schema.org',
  '@type': 'CollectionPage',
  '@id': canonicalUrl.value + '#collection',
  url: canonicalUrl.value,
  name: t('brandWheelsetSpokeSpecs.seoTitle'),
  description: t('brandWheelsetSpokeSpecs.seoDescription'),
  inLanguage: localeManifest.find(entry => entry.code === locale.value)?.iso || locale.value,
  mainEntity: {
    '@type': 'ItemList',
    itemListOrder: 'https://schema.org/ItemListOrderAscending',
    numberOfItems: itemList.value.length,
    itemListElement: itemList.value,
  },
}))

useHead(() => ({
  title: t('brandWheelsetSpokeSpecs.seoTitle'),
  meta: [
    {
      name: 'description',
      content: t('brandWheelsetSpokeSpecs.seoDescription'),
      key: 'description',
    },
    {
      name: 'robots',
      content: isIndexable.value ? 'index,follow' : 'noindex,follow',
      key: 'robots',
    },
    {
      property: 'og:title',
      content: t('brandWheelsetSpokeSpecs.seoTitle'),
      key: 'og:title',
    },
    {
      property: 'og:description',
      content: t('brandWheelsetSpokeSpecs.seoDescription'),
      key: 'og:description',
    },
  ],
  script: [createSeoJsonLdScript(collectionSchema.value)],
}))
</script>

<style scoped>
.wheelset-spoke-lookup {
  display: grid;
  gap: 1.25rem;
  width: min(100%, 88rem);
  margin: 0 auto;
  padding: 0 1rem 3rem;
  color: var(--tz-text-primary, #17212b);
}

.wheelset-spoke-lookup__header,
.wheelset-spoke-lookup__controls,
.wheelset-spoke-card,
.wheelset-spoke-lookup__footer {
  min-width: 0;
  border: 1px solid var(--tz-border-subtle, #d8dee5);
  border-radius: 1.1rem;
  background: var(--tz-card-surface, #fff);
}

.wheelset-spoke-lookup__header {
  display: grid;
  gap: 1rem;
  padding: clamp(1.25rem, 3vw, 2rem);
}

.wheelset-spoke-lookup__eyebrow,
.wheelset-spoke-card__brand {
  margin: 0;
  color: var(--tz-text-accent, #16745a);
  font-size: 0.75rem;
  font-weight: 750;
  letter-spacing: 0.11em;
  text-transform: uppercase;
}

.wheelset-spoke-lookup__title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1.25rem;
}

.wheelset-spoke-lookup h1 {
  margin: 0;
  max-width: 18ch;
  font-size: clamp(1.8rem, 3vw, 2.7rem);
  line-height: 1.12;
  letter-spacing: -0.035em;
  overflow-wrap: anywhere;
}

.wheelset-spoke-lookup__intro {
  max-width: 66ch;
  margin: 0.65rem 0 0;
  color: var(--tz-text-secondary, #56616c);
  line-height: 1.6;
}

.wheelset-spoke-lookup__review-note {
  padding: 0.9rem 1rem;
  border-left: 3px solid #b9781c;
  border-radius: 0.35rem 0.8rem 0.8rem 0.35rem;
  background: #fff8ed;
  color: #593f1f;
}

.wheelset-spoke-lookup__review-note strong {
  display: block;
  margin-bottom: 0.2rem;
  font-size: 0.9rem;
}

.wheelset-spoke-lookup__review-note p {
  margin: 0;
  font-size: 0.85rem;
  line-height: 1.55;
}

.wheelset-spoke-lookup__controls {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.25rem;
}

.wheelset-spoke-lookup__search {
  display: grid;
  gap: 0.4rem;
  width: min(100%, 34rem);
  min-width: 0;
}

.wheelset-spoke-lookup__search label {
  font-size: 0.8rem;
  font-weight: 700;
}

.wheelset-spoke-lookup__brand-filter {
  display: grid;
  gap: 0.4rem;
  min-width: min(100%, 12rem);
}

.wheelset-spoke-lookup__brand-filter label {
  font-size: 0.8rem;
  font-weight: 700;
}

.wheelset-spoke-lookup__brand-filter select {
  min-height: 2.75rem;
  padding: 0 0.75rem;
  border: 1px solid var(--tz-border-subtle, #cbd5df);
  border-radius: 0.7rem;
  background: var(--tz-surface-page, #f8fafb);
  color: inherit;
  font: inherit;
}

.wheelset-spoke-lookup__search-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.wheelset-spoke-lookup__search input {
  width: 100%;
  min-width: 0;
  height: 2.75rem;
  padding: 0 0.85rem;
  border: 1px solid var(--tz-border-subtle, #cbd5df);
  border-radius: 0.7rem;
  background: var(--tz-surface-page, #f8fafb);
  color: inherit;
  font: inherit;
}

.wheelset-spoke-lookup__search input:focus-visible,
.wheelset-spoke-lookup__brand-filter select:focus-visible,
.wheelset-spoke-lookup__clear:focus-visible,
.wheelset-spoke-lookup__footer a:focus-visible {
  outline: 3px solid color-mix(in srgb, var(--tz-text-accent, #16745a) 35%, transparent);
  outline-offset: 2px;
}

.wheelset-spoke-lookup__clear {
  flex: 0 0 auto;
  min-height: 2.75rem;
  padding: 0 0.75rem;
  border: 1px solid var(--tz-border-subtle, #cbd5df);
  border-radius: 0.7rem;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.wheelset-spoke-lookup__count {
  margin: 0;
  padding-bottom: 0.7rem;
  color: var(--tz-text-secondary, #56616c);
  font-size: 0.85rem;
  white-space: nowrap;
}

.wheelset-spoke-lookup__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}

.wheelset-spoke-card {
  display: grid;
  align-content: start;
  gap: 0.9rem;
  padding: 1.15rem;
}

.wheelset-spoke-card__header h2 {
  margin: 0.2rem 0 0;
  font-size: 1.1rem;
  line-height: 1.3;
  overflow-wrap: anywhere;
}

.wheelset-spoke-card__years {
  margin: 0.25rem 0 0;
  color: var(--tz-text-secondary, #56616c);
  font-size: 0.78rem;
}

.wheelset-spoke-card__positions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.7rem;
}

.wheelset-spoke-card__position {
  min-width: 0;
  padding: 0.8rem;
  border-radius: 0.8rem;
  background: var(--tz-surface-page, #f6f8f9);
}

.wheelset-spoke-card__position-heading h3 {
  margin: 0;
  font-size: 0.85rem;
}

.wheelset-spoke-card__position-heading p {
  margin: 0.2rem 0 0;
  color: var(--tz-text-secondary, #56616c);
  font-size: 0.7rem;
}

.wheelset-spoke-card__lengths {
  display: grid;
  gap: 0.35rem;
  margin: 0.7rem 0;
}

.wheelset-spoke-card__length {
  display: flex;
  justify-content: space-between;
  gap: 0.4rem;
  align-items: baseline;
}

.wheelset-spoke-card__length dt {
  color: var(--tz-text-secondary, #56616c);
  font-size: 0.72rem;
}

.wheelset-spoke-card__length dd {
  margin: 0;
  font-size: 0.9rem;
  font-variant-numeric: tabular-nums;
  font-weight: 750;
  white-space: nowrap;
}

.wheelset-spoke-card__type {
  display: grid;
  gap: 0.15rem;
  margin: 0;
  padding-top: 0.55rem;
  border-top: 1px solid var(--tz-border-subtle, #d8dee5);
  font-size: 0.7rem;
}

.wheelset-spoke-card__type span {
  color: var(--tz-text-secondary, #56616c);
}

.wheelset-spoke-card__type strong {
  overflow-wrap: anywhere;
  font-weight: 650;
}

.wheelset-spoke-lookup__empty {
  grid-column: 1 / -1;
  margin: 0;
  padding: 2rem 1rem;
  border: 1px dashed var(--tz-border-subtle, #cbd5df);
  border-radius: 1rem;
  color: var(--tz-text-secondary, #56616c);
  text-align: center;
}

.wheelset-spoke-lookup__footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  color: var(--tz-text-secondary, #56616c);
  font-size: 0.85rem;
}

.wheelset-spoke-lookup__footer a {
  color: var(--tz-text-accent, #16745a);
  font-weight: 700;
  text-decoration-thickness: 1px;
  text-underline-offset: 0.18em;
}

@media (max-width: 900px) {
  .wheelset-spoke-lookup__grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 620px) {
  .wheelset-spoke-lookup__title-row,
  .wheelset-spoke-lookup__controls,
  .wheelset-spoke-lookup__footer {
    align-items: stretch;
    flex-direction: column;
  }

  .wheelset-spoke-lookup__count {
    padding: 0;
  }
}

@media (max-width: 400px) {
  .wheelset-spoke-card {
    padding: 0.85rem;
  }

  .wheelset-spoke-card__positions {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
