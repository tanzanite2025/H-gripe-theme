<template>
  <div class="freehub-groupset-helper mt-5 rounded-2xl tz-surface-panel p-4 shadow-md">
    <h3 class="mb-2 text-sm font-semibold tz-text-secondary">
      {{ title || t('wheelsetFreehubHelper.title') }}
    </h3>
    <p class="mb-3 text-xs tz-text-secondary">
      {{ description || t('wheelsetFreehubHelper.description') }}
    </p>

    <div class="freehub-groupset-helper__fields">
      <div
        class="freehub-groupset-helper__field"
        :class="{ 'freehub-groupset-helper__field--current': !selectedBrand }"
      >
        <label class="freehub-groupset-helper__label text-xs font-medium tz-text-secondary" :for="brandSelectId">
          <span class="freehub-groupset-helper__step-number" aria-hidden="true">1</span>
          {{ t('wheelsetFreehubHelper.drivetrainBrand') }}
        </label>
        <SpokeCalculatorSelect
          :id="brandSelectId"
          density="compact"
          :model-value="selectedBrand"
          :options="freehubBrandSelectOptions"
          :placeholder="t('wheelsetFreehubHelper.selectBrand')"
          :aria-label="t('wheelsetFreehubHelper.drivetrainBrand')"
          :keep-menu-within-viewport="true"
          class="freehub-groupset-helper__select mt-1"
          @update:model-value="updateSelectedFreehubBrand"
        />
      </div>

      <div
        class="freehub-groupset-helper__field"
        :class="{ 'freehub-groupset-helper__field--current': selectedBrand && !selectedCassetteSpec }"
      >
        <label class="freehub-groupset-helper__label text-xs font-medium tz-text-secondary" :for="groupsetSelectId">
          <span class="freehub-groupset-helper__step-number" aria-hidden="true">2</span>
          {{ t('wheelsetFreehubHelper.groupset') }}
        </label>
        <SpokeCalculatorSelect
          :id="groupsetSelectId"
          density="compact"
          :model-value="selectedCassetteSpec"
          :options="freehubCassetteSelectOptions"
          :disabled="!selectedBrand || isLoading"
          :placeholder="isLoading
            ? t('wheelsetFreehubHelper.loading')
            : selectedBrand
              ? t('wheelsetFreehubHelper.selectGroupset')
              : t('wheelsetFreehubHelper.chooseBrandFirst')"
          :aria-label="t('wheelsetFreehubHelper.groupset')"
          :keep-menu-within-viewport="true"
          class="freehub-groupset-helper__select mt-1"
          @update:model-value="updateSelectedFreehubCassetteSpecification"
        />
      </div>
    </div>

    <p v-if="matrixError" class="mt-3 text-xs text-amber-700">
      {{ t('wheelsetFreehubHelper.loadError') }}
    </p>

    <div class="freehub-groupset-helper__status">
      <p v-if="!activeRule" class="text-xs tz-text-muted">
        {{ t('wheelsetFreehubHelper.empty') }}
      </p>

      <div v-else class="text-xs tz-text-secondary">
        <p class="font-semibold text-emerald-700">
          {{ t('wheelsetFreehubHelper.recommended') }}
          <span class="ml-1">{{ recommendedOption ? optionDisplayName(recommendedOption) : activeRule.recommended_freehub }}</span>
        </p>
        <p class="mt-0.5 tz-caption tz-text-muted">
          {{ ruleDisplayName(activeRule) }} · {{ ruleHintGroupsets(activeRule) }}
        </p>
      </div>
    </div>

    <section v-if="activeRule" class="freehub-groupset-helper__results" :aria-label="t('wheelsetFreehubHelper.resultHeading')">
      <article
        v-for="option in activeRule.fitment_options"
        :key="`${activeRule.cassette_spec}-${option.standard}`"
        class="freehub-groupset-helper__result"
        :class="{ 'is-recommended': option.standard === activeRule.recommended_freehub }"
      >
        <GuideImage
          v-if="option.image_src"
          :src="option.image_src"
          :alt="optionDisplayName(option)"
          :zoomOnClick="true"
          :caption="optionDisplayName(option)"
          class="freehub-groupset-helper__result-image rounded-lg"
        />
        <div class="freehub-groupset-helper__result-copy">
          <p class="font-semibold tz-text-primary">
            {{ optionDisplayName(option) }}
            <span v-if="option.standard === activeRule.recommended_freehub" class="text-emerald-700">
              ({{ t('wheelsetFreehubHelper.recommendedShort') }})
            </span>
          </p>
          <p class="mt-1 tz-caption tz-text-secondary">
            {{ spacerDescription(option.spacer) }}
          </p>
          <p v-if="optionNotes(option)" class="mt-1 tz-caption tz-text-muted">
            {{ optionNotes(option) }}
          </p>
        </div>
      </article>
    </section>

    <section
      v-if="activeRule"
      class="freehub-groupset-helper__facts"
      :aria-label="t('wheelsetFreehubHelper.facts.heading')"
    >
      <div
        ref="factRail"
        class="freehub-groupset-helper__fact-rail"
        role="region"
        :aria-label="t('wheelsetFreehubHelper.facts.heading')"
        @scroll.passive="updateActiveFact"
      >
        <article
          v-for="(fact, index) in compatibilityFacts"
          :key="fact.id"
          :id="`${helperId}-fact-panel-${index}`"
          :data-freehub-fact-index="index"
          :aria-labelledby="`${helperId}-fact-tab-${index}`"
          class="freehub-groupset-helper__fact"
          :class="[
            `freehub-groupset-helper__fact--${fact.id}`,
            { 'is-relevant': isFactRelevant(fact.id) },
          ]"
          role="tabpanel"
        >
          <div class="freehub-groupset-helper__fact-heading">
            <span class="freehub-groupset-helper__fact-label">{{ t(fact.labelKey) }}</span>
            <h4>{{ t(fact.titleKey) }}</h4>
          </div>
          <p class="freehub-groupset-helper__fact-callout">{{ t(fact.calloutKey) }}</p>
          <p class="freehub-groupset-helper__fact-copy">{{ t(fact.bodyKey) }}</p>
        </article>
      </div>

      <div
        class="tz-carousel-pagination freehub-groupset-helper__pagination"
        role="tablist"
        :aria-label="t('wheelsetFreehubHelper.facts.pagination')"
      >
        <button
          v-for="(fact, index) in compatibilityFacts"
          :key="`${fact.id}-dot`"
          :id="`${helperId}-fact-tab-${index}`"
          type="button"
          class="tz-carousel-pagination__dot"
          :class="{ 'is-active': activeFactIndex === index }"
          :aria-label="t('wheelsetFreehubHelper.facts.showFact', { fact: index + 1 })"
          :aria-controls="`${helperId}-fact-panel-${index}`"
          :aria-selected="activeFactIndex === index"
          role="tab"
          @click="scrollToFact(index)"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, useId, watch } from 'vue'
import { useAsyncData, useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import SpokeCalculatorSelect from '~/components/SpokeCalculatorSelect.vue'
import { useDrivetrainFitmentApi } from '~/composables/useDrivetrainFitmentApi'
import { usePageMessages } from '~/composables/usePageMessages'
import type { DrivetrainCassetteRule, DrivetrainSpacerRequirement } from '~/types/drivetrainFitment'
import {
  localizedFreehubName,
  localizedOptionNotes,
  localizedRuleDisplayName,
  localizedRuleHintGroupsets,
  localizedSpacerDescription,
} from '~/utils/drivetrainFitmentLocalization'

withDefaults(defineProps<{ title?: string; description?: string }>(), { title: '', description: '' })

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('wheelsetFreehubHelper')
const { fetchMatrix } = useDrivetrainFitmentApi()

await loadPageMessages(locale.value)
watch(locale, (nextLocale) => void loadPageMessages(nextLocale))

const { data: matrixResponse, error: matrixError, pending: isLoading } = await useAsyncData(
  'drivetrain-fitment-matrix',
  fetchMatrix,
)

const rules = computed<DrivetrainCassetteRule[]>(() => matrixResponse.value?.data?.rules || [])
const selectedBrand = ref('')
const selectedCassetteSpec = ref('')
const helperId = useId()
const factRail = ref<HTMLElement | null>(null)
const activeFactIndex = ref(0)
const brandSelectId = computed(() => `${helperId}-freehub-brand`)
const groupsetSelectId = computed(() => `${helperId}-freehub-groupset`)

const compatibilityFacts = [
  {
    id: 'xdr-xd',
    labelKey: 'wheelsetFreehubHelper.facts.xdrXd.label',
    titleKey: 'wheelsetFreehubHelper.facts.xdrXd.title',
    calloutKey: 'wheelsetFreehubHelper.facts.xdrXd.callout',
    bodyKey: 'wheelsetFreehubHelper.facts.xdrXd.body',
  },
  {
    id: 'shimano-road-12-speed',
    labelKey: 'wheelsetFreehubHelper.facts.shimano12.label',
    titleKey: 'wheelsetFreehubHelper.facts.shimano12.title',
    calloutKey: 'wheelsetFreehubHelper.facts.shimano12.callout',
    bodyKey: 'wheelsetFreehubHelper.facts.shimano12.body',
  },
] as const

type CompatibilityFactId = (typeof compatibilityFacts)[number]['id']

const brands = computed(() => Array.from(new Set(rules.value.map(rule => rule.brand))))
const filteredRules = computed(() => selectedBrand.value
  ? rules.value.filter(rule => rule.brand === selectedBrand.value)
  : [])
const activeRule = computed(() => rules.value.find(rule => rule.cassette_spec === selectedCassetteSpec.value) || null)
const recommendedOption = computed(() => activeRule.value?.fitment_options.find(
  option => option.standard === activeRule.value?.recommended_freehub,
) || activeRule.value?.fitment_options[0] || null)

watch(selectedBrand, () => { selectedCassetteSpec.value = '' })

const optionLabel = (rule: DrivetrainCassetteRule) => (
  rule.hint_groupsets ? `${ruleDisplayName(rule)} (${ruleHintGroupsets(rule)})` : ruleDisplayName(rule)
)

const freehubBrandSelectOptions = computed(() => brands.value.map(brand => ({
  label: brand,
  value: brand,
})))

const freehubCassetteSelectOptions = computed(() => filteredRules.value.map(rule => ({
  label: optionLabel(rule),
  value: rule.cassette_spec,
})))

const updateSelectedFreehubBrand = (brand: string | number | null) => {
  selectedBrand.value = typeof brand === 'string' ? brand : ''
}

const updateSelectedFreehubCassetteSpecification = (cassetteSpecification: string | number | null) => {
  selectedCassetteSpec.value = typeof cassetteSpecification === 'string' ? cassetteSpecification : ''
}

const ruleDisplayName = (rule: DrivetrainCassetteRule) => localizedRuleDisplayName(rule, locale.value)
const ruleHintGroupsets = (rule: DrivetrainCassetteRule) => localizedRuleHintGroupsets(rule, locale.value)
const optionDisplayName = (option: DrivetrainCassetteRule['fitment_options'][number]) => localizedFreehubName(option, locale.value)
const optionNotes = (option: DrivetrainCassetteRule['fitment_options'][number]) => localizedOptionNotes(option, locale.value)
const spacerDescription = (spacer: DrivetrainSpacerRequirement) => {
  const localized = localizedSpacerDescription(spacer, locale.value)
  if (localized) return localized
  if (!spacer.required) return t('wheelsetFreehubHelper.directInstall')
  return t('wheelsetFreehubHelper.spacerRequired', { thickness: spacer.thickness_mm })
}

const relevantFactId = computed<CompatibilityFactId | null>(() => {
  if (activeRule.value?.brand === 'SRAM' && activeRule.value.min_cog_teeth <= 10) return 'xdr-xd'
  if (activeRule.value?.brand === 'Shimano' && activeRule.value.speed === 12 && activeRule.value.min_cog_teeth === 11) return 'shimano-road-12-speed'
  return null
})

const isFactRelevant = (factId: CompatibilityFactId) => relevantFactId.value === factId

watch(activeRule, async () => {
  const relevantIndex = compatibilityFacts.findIndex(fact => isFactRelevant(fact.id))
  const index = relevantIndex >= 0 ? relevantIndex : 0
  activeFactIndex.value = index
  await nextTick()
  scrollToFact(index)
})

const scrollToFact = (index: number) => {
  const currentRail = factRail.value
  const card = currentRail?.querySelector<HTMLElement>(`[data-freehub-fact-index="${index}"]`)
  if (!currentRail || !card) return
  currentRail.scrollTo({ left: card.offsetLeft, behavior: 'smooth' })
  activeFactIndex.value = index
}

const updateActiveFact = () => {
  const currentRail = factRail.value
  if (!currentRail) return
  const center = currentRail.scrollLeft + currentRail.clientWidth / 2
  const cards = Array.from(currentRail.querySelectorAll<HTMLElement>('[data-freehub-fact-index]'))
  if (!cards.length) return
  const nearestCard = cards.reduce((nearest, card) => (
    Math.abs(card.offsetLeft + card.offsetWidth / 2 - center)
      < Math.abs(nearest.offsetLeft + nearest.offsetWidth / 2 - center) ? card : nearest
  ))
  activeFactIndex.value = Number(nearestCard.dataset.freehubFactIndex || 0)
}
</script>

<style scoped>
.freehub-groupset-helper__fields { display: grid; grid-template-columns: minmax(0, 1fr); gap: 0.75rem; align-items: end; }
.freehub-groupset-helper__field {
  min-width: 0;
  --spoke-border: color-mix(in srgb, var(--tz-text-secondary) 55%, var(--tz-border-strong));
  --spoke-border-strong: var(--tz-site-accent-hover);
  --spoke-focus-ring: var(--tz-site-accent);
}
.freehub-groupset-helper__select { min-width: 0; }
.freehub-groupset-helper__label { display: flex; align-items: center; gap: 0.375rem; }
.freehub-groupset-helper__step-number { display: inline-flex; width: 1.125rem; height: 1.125rem; flex-shrink: 0; align-items: center; justify-content: center; border: 1px solid var(--tz-border-strong); border-radius: 50%; background: var(--tz-surface-muted); color: var(--tz-text-secondary); font-size: 0.6875rem; line-height: 1; }
.freehub-groupset-helper__field--current {
  --spoke-border: var(--tz-site-accent-hover);
  --spoke-border-strong: var(--tz-text-accent-hover);
  --spoke-control-surface: var(--tz-site-accent-hover);
  --spoke-result-surface: var(--tz-text-accent-hover);
}
.freehub-groupset-helper__field--current .freehub-groupset-helper__step-number { border-color: var(--tz-site-accent-hover); background: var(--tz-site-accent-hover); color: #fff; }
.freehub-groupset-helper__field :deep(.spoke-calculator-select__chevron) { color: var(--tz-text-secondary); }
.freehub-groupset-helper__field--current :deep(.spoke-calculator-select__button),
.freehub-groupset-helper__field--current :deep(.spoke-calculator-select__chevron) { color: #fff; }
.freehub-groupset-helper__field--current :deep(.spoke-calculator-select__menu) { background: var(--tz-input-surface); }
.freehub-groupset-helper__field :deep(.spoke-calculator-select__button:focus-visible) { box-shadow: 0 0 0 3px color-mix(in srgb, var(--tz-site-accent) 25%, transparent); }
.freehub-groupset-helper__status { margin-top: 0.75rem; }
.freehub-groupset-helper__results { display: grid; gap: 0.625rem; margin-top: 0.875rem; }
.freehub-groupset-helper__result { display: grid; grid-template-columns: minmax(5rem, 7rem) minmax(0, 1fr); gap: 0.75rem; padding: 0.75rem; border: 1px solid var(--tz-border-subtle); border-radius: 0.5rem; background: var(--tz-surface-panel); }
.freehub-groupset-helper__result.is-recommended { border-color: #5aa88e; }
.freehub-groupset-helper__result-image { width: 100%; aspect-ratio: 1; object-fit: cover; }
.freehub-groupset-helper__result-copy { align-self: center; }
.freehub-groupset-helper__facts { margin-top: 0.875rem; }
.freehub-groupset-helper__fact-rail { display: grid; grid-auto-columns: 100%; grid-auto-flow: column; gap: 0.625rem; overflow-x: auto; overscroll-behavior-inline: contain; scroll-snap-type: x mandatory; scrollbar-width: none; }
.freehub-groupset-helper__fact-rail::-webkit-scrollbar { display: none; }
.freehub-groupset-helper__fact { min-width: 0; padding: 0.75rem; border: 1px solid var(--tz-border-subtle); border-radius: 0.5rem; background: var(--tz-surface-panel); scroll-snap-align: start; }
.freehub-groupset-helper__fact--xdr-xd { border-color: #d6a23d; background: #fffaf0; }
.freehub-groupset-helper__fact--shimano-road-12-speed { border-color: #5aa88e; background: #f2fbf7; }
.freehub-groupset-helper__fact-heading h4 { margin-top: 0.25rem; color: var(--tz-text-primary); font-size: 0.75rem; font-weight: 650; line-height: 1.35; }
.freehub-groupset-helper__fact-label { color: var(--tz-text-muted); font-size: 0.625rem; font-weight: 600; }
.freehub-groupset-helper__fact-callout { margin-top: 0.5rem; color: var(--tz-text-primary); font-size: 0.7rem; font-weight: 650; line-height: 1.45; }
.freehub-groupset-helper__fact--xdr-xd .freehub-groupset-helper__fact-callout { color: #a16207; }
.freehub-groupset-helper__fact--shimano-road-12-speed .freehub-groupset-helper__fact-callout { color: #047857; }
.freehub-groupset-helper__fact-copy { margin-top: 0.25rem; color: var(--tz-text-secondary); font-size: 0.675rem; line-height: 1.5; }
.freehub-groupset-helper__pagination { margin-top: 0.5rem; }
@media (min-width: 640px) {
  .freehub-groupset-helper__fields { width: min(100%, 37rem); grid-template-columns: 10rem minmax(0, 1fr); }
}
@media (min-width: 768px) {
  .freehub-groupset-helper__fact-rail { grid-auto-columns: minmax(0, 1fr); grid-auto-flow: initial; grid-template-columns: repeat(auto-fit, minmax(min(100%, 22rem), 1fr)); overflow: visible; }
  .freehub-groupset-helper__pagination { display: none; }
}
</style>
