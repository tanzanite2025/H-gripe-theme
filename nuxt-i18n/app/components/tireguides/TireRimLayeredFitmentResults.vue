<template>
  <div class="tire-rim-layered-results" aria-live="polite">
    <section class="tire-rim-layered-results__best" :class="{ 'tire-rim-layered-results__best--empty': !suggestion.bestFitment }">
      <div class="tire-rim-layered-results__heading">
        <h4>{{ t(`guidesTireRimHelper.layered.${isKnownRim ? 'bestTireTitle' : 'bestRimTitle'}`) }}</h4>
        <span>{{ t(`guidesTireRimHelper.applicationProfiles.${resolvedProfile}`) }}</span>
      </div>
      <template v-if="suggestion.bestFitment">
        <strong class="tire-rim-layered-results__value">
          {{ isKnownRim ? suggestion.bestFitment.tireWidthMm : suggestion.bestFitment.rimInnerWidthMm }}<small>mm</small>
        </strong>
        <p>{{ t('guidesTireRimHelper.layered.pairing', {
          rimWidth: suggestion.bestFitment.rimInnerWidthMm,
          tireWidth: suggestion.bestFitment.tireWidthMm,
        }) }}</p>
        <p class="tire-rim-layered-results__note">
          {{ t(`guidesTireRimHelper.layered.${suggestion.aeroRuleApplies ? 'aeroSelectionNote' : 'wideSelectionNote'}`) }}
        </p>
        <p v-if="bestFitmentManufacturerNames.length" class="tire-rim-layered-results__note">
          {{ t('guidesTireRimHelper.layered.bestManufacturerSources', { sources: bestFitmentManufacturerNames.join(' / ') }) }}
        </p>
      </template>
      <template v-else>
        <strong class="tire-rim-layered-results__empty-title">{{ t('guidesTireRimHelper.layered.noBestTitle') }}</strong>
        <p>{{ t(`guidesTireRimHelper.layered.${resolvedProfile === 'fat_snow' ? 'noFatBestNote' : 'noBestNote'}`) }}</p>
      </template>
      <p v-if="rimOuterWidthEstimate" class="tire-rim-layered-results__note">
        {{ t('guidesTireRimHelper.reverseRimOuterWidthEstimate', {
          innerWidth: rimOuterWidthEstimate.rimInnerWidthMm,
          wallThickness: rimOuterWidthEstimate.wallThicknessPerSideEstimateMm,
          outerWidth: rimOuterWidthEstimate.rimOuterWidthEstimateMm,
        }) }}
      </p>
      <p v-if="rimOuterWidthEstimate" class="tire-rim-layered-results__note">{{ t('guidesTireRimHelper.layered.outerWidthAssumption') }}</p>
      <p v-if="suggestion.aeroRuleApplies && 'aeroTargetTireWidthMm' in suggestion && suggestion.aeroTargetTireWidthMm !== null && rimOuterWidthEstimate" class="tire-rim-layered-results__note">
        {{ t('guidesTireRimHelper.reverseAeroTargetTireWidth', { outerWidth: rimOuterWidthEstimate.rimOuterWidthEstimateMm, targetWidth: suggestion.aeroTargetTireWidthMm }) }}
      </p>
      <p v-else-if="suggestion.aeroRuleApplies && 'physical' in suggestion && suggestion.physical.aeroTargetOuterWidth" class="tire-rim-layered-results__note">
        {{ t('guidesTireRimHelper.aeroTargetOuterWidth') }} {{ formatTireRimWidthReferenceRanges([suggestion.physical.aeroTargetOuterWidth]) }} mm
      </p>
      <div class="tire-rim-layered-results__clearance">
        <h5>{{ t('guidesTireRimHelper.layered.frameClearanceTitle') }}</h5>
        <p class="tire-rim-layered-results__note">{{ t('guidesTireRimHelper.layered.frameClearanceNote') }}</p>
        <NuxtLink :to="localePath('/guides/tireguides/tire-frame-clearance')" class="tire-rim-layered-results__clearance-link">
          {{ t('guidesTireRimHelper.layered.frameClearanceLink') }}<span aria-hidden="true"> →</span>
        </NuxtLink>
      </div>
    </section>

    <section class="tire-rim-layered-results__normal">
      <h4>{{ t('guidesTireRimHelper.layered.normalTitle') }}</h4>
      <dl>
        <div>
          <dt>{{ t(`guidesTireRimHelper.layered.${isKnownRim ? 'normalTireLabel' : 'normalRimLabel'}`) }}</dt>
          <dd>{{ normalRanges.length ? formatTireRimWidthReferenceRanges(normalRanges) + ' mm' : t('guidesTireRimHelper.layered.noNormalRange') }}</dd>
        </div>
      </dl>
      <p class="tire-rim-layered-results__note">{{ t('guidesTireRimHelper.layered.normalNote') }}</p>
      <div v-if="enteredCalculation" class="tire-rim-layered-results__entered" :class="`tire-rim-layered-results__entered--${enteredCalculation.verdict}`">
        <strong>{{ t('guidesTireRimHelper.layered.enteredPairing') }}</strong>
        <p>{{ t('guidesTireRimHelper.layered.pairing', { rimWidth: enteredCalculation.rimInnerWidthMm, tireWidth: enteredCalculation.tireWidthMm }) }}</p>
        <p>{{ t(`guidesTireRimHelper.layered.enteredReasons.${enteredCalculation.verdictReason}`, {
          minimum: enteredCalculation.allowedTireWidthRange.min,
          maximum: enteredCalculation.allowedTireWidthRange.max,
        }) }}</p>
      </div>
      <div class="tire-rim-layered-results__reference">
        <h5>{{ t('guidesTireRimHelper.layered.referenceTitle') }}</h5>
        <p v-if="standardsText" class="tire-rim-layered-results__standards">{{ standardsText }}</p>
        <p v-else class="tire-rim-layered-results__note">{{ t('guidesTireRimHelper.layered.referenceUnavailable') }}</p>
        <p class="tire-rim-layered-results__note">{{ t(`guidesTireRimHelper.layered.${normalFitmentReference ? 'standardRangeRole' : 'standardRole'}`) }}</p>
        <template v-if="normalFitmentManufacturerNames.length">
          <p>{{ t('guidesTireRimHelper.layered.manufacturerSources', { sources: normalFitmentManufacturerNames.join(' / ') }) }}</p>
          <p class="tire-rim-layered-results__note">{{ t('guidesTireRimHelper.layered.manufacturerRole') }}</p>
        </template>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import type { TireRimSuggestion, TireRimReverseSuggestion } from '~/composables/useTireRimWidthReferenceRecommendation'
import { formatTireRimWidthReferenceRanges } from '~/data/tireguides/tireRimWidthReferencePresentation'
import type { TireRimFitmentRuleReference } from '~/data/tireguides/tireRimWidthReferencePresentation'

const props = defineProps<{ suggestion: TireRimSuggestion | TireRimReverseSuggestion }>()
const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages } = usePageMessages('guidesTireRimHelper')
await loadPageMessages(locale.value)
watch(locale, nextLocale => { void loadPageMessages(nextLocale) })

const isKnownRim = computed(() => 'normalFitmentTireWidthRange' in props.suggestion || ('engineering' in props.suggestion && props.suggestion.engineering !== null))
const resolvedProfile = computed(() => 'resolvedApplicationProfile' in props.suggestion
  ? props.suggestion.resolvedApplicationProfile
  : props.suggestion.applicationProfile)
const normalRanges = computed(() => {
  if ('normalFitmentTireWidthRange' in props.suggestion) return [props.suggestion.normalFitmentTireWidthRange]
  if (props.suggestion.engineering) return [props.suggestion.engineering.allowedTireWidthRange]
  return props.suggestion.normalFitmentRimWidthRanges
})
const enteredCalculation = computed(() => 'engineering' in props.suggestion ? props.suggestion.engineering : null)
const rimOuterWidthEstimate = computed(() => 'rimOuterWidthEstimate' in props.suggestion
  ? props.suggestion.rimOuterWidthEstimate
  : props.suggestion.physical.rimOuterWidthEstimates[0] ?? null)
const normalFitmentRuleReference = computed(() => {
  if ('fitmentRuleReference' in props.suggestion) return props.suggestion.fitmentRuleReference
  return props.suggestion.engineering?.fitmentRuleReference ?? null
})
const standardsText = computed(() => {
  if (normalFitmentReference.value) {
    return t('guidesTireRimHelper.layered.normalRangeSource', {
      standard: normalFitmentReference.value.standard,
      clause: normalFitmentReference.value.clause,
      table: normalFitmentReference.value.table,
    })
  }
  const reference = props.suggestion.standardReference
  return [reference?.iso_5775_part_1, reference?.iso_5775_part_2, reference?.etrto_edition].filter(Boolean).join(' / ')
})
const normalFitmentReference = computed(() => props.suggestion.normalFitmentReference
  ?? ('engineering' in props.suggestion ? props.suggestion.engineering?.normalFitmentReference : undefined))
const readManufacturerNamesFromFitmentRuleReference = (reference: TireRimFitmentRuleReference | null | undefined): string[] => [...new Set((reference?.manufacturer_source_urls ?? []).map(source => {
  try {
    const hostname = new URL(source).hostname.replace(/^www\./, '')
    if (hostname === 'enve.com') return 'ENVE'
    if (hostname === 'stans.com') return "Stan’s"
    if (hostname === 'wtb.com') return 'WTB'
    return hostname
  } catch { return '' }
}).filter(Boolean))]
const bestFitmentManufacturerNames = computed(() => readManufacturerNamesFromFitmentRuleReference(props.suggestion.bestFitment?.fitmentRuleReference))
const normalFitmentManufacturerNames = computed(() => readManufacturerNamesFromFitmentRuleReference(normalFitmentRuleReference.value))
</script>

<style scoped>
.tire-rim-layered-results { display: grid; gap: 0.75rem; text-align: left; color: var(--tz-text-secondary); font-size: 0.8rem; }
.tire-rim-layered-results section { min-width: 0; padding: 1rem; border: 1px solid var(--tz-border-subtle); border-radius: 0.85rem; background: var(--tz-form-panel-surface); }
.tire-rim-layered-results .tire-rim-layered-results__best { border-color: color-mix(in srgb, var(--tz-site-accent) 40%, transparent); background: color-mix(in srgb, var(--tz-site-accent) 7%, var(--tz-form-panel-surface)); }
.tire-rim-layered-results .tire-rim-layered-results__best--empty { border-color: var(--tz-border-subtle); background: var(--tz-form-panel-surface); }
.tire-rim-layered-results h4 { margin: 0; font-size: 0.84rem; font-weight: 700; color: var(--tz-text-primary); }
.tire-rim-layered-results__heading { display: flex; align-items: baseline; justify-content: space-between; gap: 0.5rem; flex-wrap: wrap; }
.tire-rim-layered-results__heading span { font-size: 0.72rem; color: var(--tz-text-muted); }
.tire-rim-layered-results__value { display: block; margin: 0.45rem 0; color: var(--tz-site-accent); font-size: 2rem; line-height: 1.1; }
.tire-rim-layered-results__value small { margin-left: 0.35rem; font-size: 0.85rem; font-weight: 500; }
.tire-rim-layered-results__empty-title { display: block; margin-top: 0.5rem; color: var(--tz-text-primary); font-size: 0.95rem; }
.tire-rim-layered-results p { margin: 0.4rem 0 0; line-height: 1.6; overflow-wrap: anywhere; text-align: left; }
.tire-rim-layered-results .tire-rim-layered-results__note { color: var(--tz-text-muted); font-size: 0.74rem; }
.tire-rim-layered-results dl { margin: 0.55rem 0 0; }
.tire-rim-layered-results dt { font-size: 0.74rem; color: var(--tz-text-muted); }
.tire-rim-layered-results dd { margin: 0.2rem 0 0; color: var(--tz-text-primary); font-size: 0.9rem; font-weight: 600; }
.tire-rim-layered-results__standards { color: var(--tz-text-primary); font-size: 0.72rem; }
.tire-rim-layered-results__reference { margin-top: 0.75rem; padding-top: 0.7rem; border-top: 1px solid var(--tz-border-subtle); }
.tire-rim-layered-results__reference h5 { margin: 0; color: var(--tz-text-primary); font-size: 0.76rem; font-weight: 600; }
.tire-rim-layered-results__clearance { margin-top: 0.75rem; padding-top: 0.7rem; border-top: 1px solid var(--tz-border-subtle); }
.tire-rim-layered-results__clearance h5 { margin: 0; color: var(--tz-text-primary); font-size: 0.76rem; font-weight: 600; }
.tire-rim-layered-results__clearance-link { display: inline-block; margin-top: 0.45rem; color: var(--tz-text-accent); font-size: 0.76rem; font-weight: 600; text-decoration: underline; text-underline-offset: 0.2em; }
.tire-rim-layered-results__entered { margin-top: 0.75rem; padding-top: 0.7rem; border-top: 1px solid var(--tz-border-subtle); }
.tire-rim-layered-results__entered--critical { color: var(--tz-error, #dc2626); }
@media (max-width: 767px) { .tire-rim-layered-results section { padding: 0.75rem; } }
</style>
