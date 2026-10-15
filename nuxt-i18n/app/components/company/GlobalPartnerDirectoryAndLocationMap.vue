<template>
  <section
    id="global-partner-directory-and-location-map"
    aria-labelledby="global-partner-directory-and-location-map-heading"
    class="premium-card mt-4 overflow-hidden rounded-2xl text-left"
  >
    <header class="border-b tz-border-subtle p-5 sm:px-6">
      <h2 id="global-partner-directory-and-location-map-heading" class="text-sm font-semibold tz-text-primary">
        {{ t(hasPublicGlobalPartners ? 'companyGlobalPartners.map.directoryTitle' : 'companyGlobalPartners.map.title') }}
      </h2>
      <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
        {{ t('companyGlobalPartners.map.description') }}
      </p>
    </header>

    <div class="grid lg:grid-cols-[minmax(16rem,0.3fr)_minmax(0,0.7fr)]">
      <div class="flex min-w-0 flex-col border-b tz-border-subtle p-5 sm:p-6 lg:border-b-0 lg:border-r">
        <template v-if="hasPublicGlobalPartners">
          <h3 class="text-xs font-semibold tz-text-primary">
            {{ t('companyGlobalPartners.map.directoryHeading') }}
          </h3>
          <ul class="mt-3 max-h-72 space-y-2 overflow-y-auto">
            <li v-for="partner in partners" :key="partner.id">
              <button
                type="button"
                :data-global-partner-directory-entry="partner.id"
                :aria-pressed="selectedGlobalPartnerId === partner.id"
                :class="['global-partner-directory-entry flex w-full items-start gap-3 rounded-xl border p-3 text-left', selectedGlobalPartnerId === partner.id ? 'global-partner-directory-entry--selected' : 'tz-border-subtle']"
                @click="selectGlobalPartnerDirectoryEntry(partner.id)"
              >
                <img v-if="partner.logoUrl" :src="partner.logoUrl" alt="" class="h-9 w-9 shrink-0 rounded-lg object-contain" loading="lazy" />
                <span v-else class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg tz-surface-muted tz-text-secondary" aria-hidden="true">
                  <Icon name="lucide:handshake" class="h-4 w-4" />
                </span>
                <span class="min-w-0">
                  <span class="block text-sm font-semibold tz-text-primary">{{ partner.name }}</span>
                  <span class="mt-1 block text-xs tz-text-secondary">{{ partner.cityName }}, {{ partner.countryName }}</span>
                  <span class="mt-1 block text-xs tz-text-accent">{{ t(globalPartnerTypeTranslationKeys[partner.partnershipChannelId]) }}</span>
                </span>
              </button>
            </li>
          </ul>
        </template>
        <div v-else>
          <p class="text-xs font-medium tz-text-accent">{{ t('companyGlobalPartners.map.contactEyebrow') }}</p>
          <h3 class="mt-3 text-base font-semibold leading-snug tz-text-primary">
            {{ t('companyGlobalPartners.map.contactTitle') }}
          </h3>
          <p class="mt-3 text-xs leading-relaxed tz-text-secondary">
            {{ t('companyGlobalPartners.map.contactDescription') }}
          </p>
        </div>

        <div class="mt-auto pt-5">
          <button
            type="button"
            data-global-partner-contact-location
            :aria-pressed="selectedGlobalPartnerId === null"
            class="global-partner-directory-entry flex w-full items-center gap-3 rounded-xl border tz-border-subtle p-3 text-left"
            @click="selectXiamenGlobalPartnerContactTeam"
          >
            <span class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg tz-surface-muted tz-text-accent" aria-hidden="true">
              <Icon name="lucide:map-pin" class="h-4 w-4" />
            </span>
            <span>
              <span class="block text-xs font-semibold tz-text-primary">{{ t('companyGlobalPartners.map.xiamenLocation') }}</span>
              <span class="mt-1 block text-xs tz-text-secondary">{{ t('companyGlobalPartners.map.contactTeam') }}</span>
            </span>
          </button>
          <NuxtLink
            :to="localePath('/company/contact')"
            class="premium-button premium-button--active mt-3 w-full justify-center gap-2 text-xs"
          >
            <span>{{ t('companyGlobalPartners.map.contactCta') }}</span>
            <Icon name="lucide:arrow-right" class="h-4 w-4" aria-hidden="true" />
          </NuxtLink>
        </div>
      </div>

      <div class="flex min-w-0 flex-col tz-surface-muted p-4 sm:p-6">
        <div class="flex flex-wrap items-center justify-between gap-2 text-xs tz-text-secondary">
          <span>{{ t('companyGlobalPartners.map.locationHeading') }}</span>
          <div class="flex flex-wrap items-center gap-4">
            <span class="inline-flex items-center gap-1.5">
              <Icon name="lucide:map-pin" class="h-3.5 w-3.5 tz-text-accent" aria-hidden="true" />
              {{ t('companyGlobalPartners.map.contactTeam') }}
            </span>
            <span v-if="hasPublicGlobalPartners" class="inline-flex items-center gap-1.5">
              <span class="global-partner-map-legend-dot" aria-hidden="true"></span>
              {{ t('companyGlobalPartners.map.directoryHeading') }}
            </span>
          </div>
        </div>

        <div class="global-partner-location-map relative my-auto w-full" data-global-partner-location-map>
          <svg viewBox="0 0 960 400" class="block h-auto w-full" role="img" :aria-labelledby="`${globalPartnerMapTitleId} ${globalPartnerMapDescriptionId}`">
            <title :id="globalPartnerMapTitleId">{{ t('companyGlobalPartners.map.svgTitle') }}</title>
            <desc :id="globalPartnerMapDescriptionId">
              {{ t(hasPublicGlobalPartners ? 'companyGlobalPartners.map.svgDescriptionWithPartners' : 'companyGlobalPartners.map.svgDescription') }}
            </desc>
            <defs>
              <mask :id="globalPartnerWorldLandMaskId" maskUnits="userSpaceOnUse" x="0" y="0" width="960" height="400">
                <image href="/company/globalpartners/global-partner-world-land-outline.svg" width="960" height="400" />
              </mask>
            </defs>
            <rect width="960" height="400" fill="var(--tz-border-strong)" :mask="`url(#${globalPartnerWorldLandMaskId})`" />
          </svg>

          <button
            v-for="partner in globalPartnerMapLocations"
            :key="partner.id"
            type="button"
            :data-global-partner-map-marker="partner.id"
            :aria-label="`${partner.name}, ${partner.cityName}, ${partner.countryName}`"
            :aria-pressed="selectedGlobalPartnerId === partner.id"
            :style="{ left: `${partner.mapPosition.leftPercent}%`, top: `${partner.mapPosition.topPercent}%` }"
            :class="['global-partner-map-marker global-partner-map-marker--partner', { 'global-partner-map-marker--selected': selectedGlobalPartnerId === partner.id }]"
            @click="selectGlobalPartnerDirectoryEntry(partner.id)"
          >
            <span class="global-partner-map-marker-dot" aria-hidden="true"></span>
          </button>

          <button
            type="button"
            data-global-partner-map-contact-marker
            :aria-label="t('companyGlobalPartners.map.xiamenContactMarker')"
            :aria-pressed="selectedGlobalPartnerId === null"
            :style="{ left: `${xiamenContactMapPosition.leftPercent}%`, top: `${xiamenContactMapPosition.topPercent}%` }"
            :class="['global-partner-map-marker global-partner-map-marker--contact', { 'global-partner-map-marker--selected': selectedGlobalPartnerId === null }]"
            @click="selectXiamenGlobalPartnerContactTeam"
          >
            <Icon name="lucide:map-pin" class="h-5 w-5" aria-hidden="true" />
            <span class="global-partner-map-contact-label">{{ t('companyGlobalPartners.map.xiamenContactMarker') }}</span>
          </button>
        </div>

        <div class="border-t tz-border-subtle pt-3 text-xs leading-relaxed" aria-live="polite" aria-atomic="true">
          <template v-if="selectedGlobalPartner">
            <p class="font-semibold tz-text-primary">{{ selectedGlobalPartner.name }}</p>
            <p class="mt-1 tz-text-secondary">{{ selectedGlobalPartner.cityName }}, {{ selectedGlobalPartner.countryName }} · {{ t(globalPartnerTypeTranslationKeys[selectedGlobalPartner.partnershipChannelId]) }}</p>
            <a v-if="selectedGlobalPartner.websiteUrl" :href="selectedGlobalPartner.websiteUrl" target="_blank" rel="noopener noreferrer" class="mt-2 inline-flex items-center gap-1 font-medium tz-text-accent hover:underline">
              {{ t('companyGlobalPartners.map.partnerWebsite') }}
              <Icon name="lucide:external-link" class="h-3 w-3" aria-hidden="true" />
            </a>
          </template>
          <template v-else>
            <p class="font-semibold tz-text-primary">{{ t('companyGlobalPartners.map.xiamenContactMarker') }}</p>
            <p class="mt-1 tz-text-secondary">{{ t('companyGlobalPartners.map.contactLocationDescription') }}</p>
          </template>
        </div>
      </div>
    </div>

    <p class="border-t tz-border-subtle px-5 py-3 text-xs leading-relaxed tz-text-secondary sm:px-6">
      {{ t('companyGlobalPartners.confidentiality') }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useId, watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import type { GlobalPartnerPublicDirectoryEntry } from '~/data/company/globalPartnerPublicDirectoryEntries'

const props = withDefaults(defineProps<{
  partners?: readonly GlobalPartnerPublicDirectoryEntry[]
}>(), { partners: () => [] })

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages } = usePageMessages('companyGlobalPartners')
await loadPageMessages(locale.value)
watch(locale, nextLocale => { void loadPageMessages(nextLocale) })

const globalPartnerMapInstanceId = useId()
const globalPartnerMapTitleId = `global-partner-map-title-${globalPartnerMapInstanceId}`
const globalPartnerMapDescriptionId = `global-partner-map-description-${globalPartnerMapInstanceId}`
const globalPartnerWorldLandMaskId = `global-partner-world-land-mask-${globalPartnerMapInstanceId}`
const selectedGlobalPartnerId = ref<string | null>(null)
const hasPublicGlobalPartners = computed(() => props.partners.length > 0)
const selectedGlobalPartner = computed(() => props.partners.find(partner => partner.id === selectedGlobalPartnerId.value))

const globalPartnerTypeTranslationKeys = {
  'oem-odm': 'companyGlobalPartners.map.partnerTypes.oemOdm',
  'dealer-builder': 'companyGlobalPartners.map.partnerTypes.dealerBuilder',
  'teams-ambassadors': 'companyGlobalPartners.map.partnerTypes.teamsAndAmbassadors',
} as const

// Match the land outline's equirectangular projection and 85 N / 60 S crop.
function projectGlobalPartnerCoordinatesToWorldMap(latitude: number, longitude: number) {
  return {
    leftPercent: (longitude + 180) / 360 * 100,
    topPercent: (85 - latitude) / 145 * 100,
  }
}

const xiamenContactMapPosition = projectGlobalPartnerCoordinatesToWorldMap(24.4798, 118.0895)
const globalPartnerMapLocations = computed(() => props.partners.map(partner => ({
  ...partner,
  mapPosition: projectGlobalPartnerCoordinatesToWorldMap(partner.latitude, partner.longitude),
})))

function selectGlobalPartnerDirectoryEntry(partnerId: string) {
  selectedGlobalPartnerId.value = partnerId
}

function selectXiamenGlobalPartnerContactTeam() {
  selectedGlobalPartnerId.value = null
}

watch(() => props.partners, () => {
  if (!selectedGlobalPartner.value) selectedGlobalPartnerId.value = null
})
</script>

<style scoped>
.global-partner-directory-entry:hover {
  background: var(--tz-surface-muted);
}

.global-partner-directory-entry--selected {
  border-color: var(--tz-site-accent);
  background: var(--tz-site-accent-soft-surface);
}

.global-partner-directory-entry:focus-visible,
.global-partner-map-marker:focus-visible {
  outline: 2px solid var(--tz-form-control-focus-border);
  outline-offset: 3px;
}

.global-partner-location-map {
  aspect-ratio: 12 / 5;
  margin-top: 1.5rem;
  margin-bottom: 1.5rem;
}

.global-partner-map-marker {
  position: absolute;
  display: flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  transform: translate(-50%, -50%);
  color: var(--tz-text-accent);
}

.global-partner-map-marker--contact {
  z-index: 2;
  background: var(--tz-surface-card);
  border: 1px solid var(--tz-border-strong);
}

.global-partner-map-marker--selected {
  z-index: 3;
  outline: 2px solid var(--tz-site-accent);
  outline-offset: 2px;
}

.global-partner-map-marker-dot,
.global-partner-map-legend-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 9999px;
  background: var(--tz-text-accent);
}

.global-partner-map-marker-dot {
  width: 13px;
  height: 13px;
  border: 2px solid var(--tz-surface-card);
}

.global-partner-map-contact-label {
  position: absolute;
  right: calc(100% + 8px);
  padding: 4px 7px;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 6px;
  background: var(--tz-surface-card);
  color: var(--tz-text-primary);
  white-space: nowrap;
  font-size: var(--tz-type-micro-label);
  line-height: 1.5;
  pointer-events: none;
}
</style>
