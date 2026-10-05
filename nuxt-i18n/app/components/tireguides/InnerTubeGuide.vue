<template>
  <div class="inner-tube-guide space-y-5 md:space-y-6">
    <div class="rounded-2xl bg-[var(--tz-card-surface)] p-2 shadow-md md:p-3">
      <div
        class="grid grid-cols-2 gap-1 rounded-xl bg-[var(--tz-page-surface)] p-1"
        role="tablist"
        :aria-label="t('guidesTireInnerTube.tabs.ariaLabel')"
      >
        <button
          id="inner-tube-introduction-tab"
          type="button"
          role="tab"
          :aria-selected="activeInnerTubeGuideTab === 'introduction'"
          aria-controls="inner-tube-introduction-panel"
          class="rounded-lg px-3 py-2.5 text-sm font-bold transition-colors"
          :class="activeInnerTubeGuideTab === 'introduction'
            ? 'bg-[var(--tz-action-primary)] text-[var(--tz-action-primary-foreground)] shadow-sm'
            : 'tz-text-secondary hover:tz-text-primary'"
          @click="activeInnerTubeGuideTab = 'introduction'"
        >
          {{ t('guidesTireInnerTube.tabs.introduction') }}
        </button>
        <button
          id="inner-tube-calculator-tab"
          type="button"
          role="tab"
          :aria-selected="activeInnerTubeGuideTab === 'calculator'"
          aria-controls="inner-tube-calculator-panel"
          class="rounded-lg px-3 py-2.5 text-sm font-bold transition-colors"
          :class="activeInnerTubeGuideTab === 'calculator'
            ? 'bg-[var(--tz-action-primary)] text-[var(--tz-action-primary-foreground)] shadow-sm'
            : 'tz-text-secondary hover:tz-text-primary'"
          @click="activeInnerTubeGuideTab = 'calculator'"
        >
          {{ t('guidesTireInnerTube.tabs.calculator') }}
        </button>
      </div>
      <p class="mt-2 px-2 text-center text-xs leading-relaxed tz-text-secondary">
        {{ activeInnerTubeGuideTab === 'introduction'
          ? t('guidesTireInnerTube.tabs.introductionDescription')
          : t('guidesTireInnerTube.tabs.calculatorDescription') }}
      </p>
    </div>

    <div
      id="inner-tube-introduction-panel"
      v-show="activeInnerTubeGuideTab === 'introduction'"
      role="tabpanel"
      aria-labelledby="inner-tube-introduction-tab"
      class="space-y-5 md:space-y-6"
    >
    <!-- Step 1: Search / Intro Card -->
     <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-3 md:p-4">
       <div class="flex items-center justify-between gap-3">
        <div class="min-w-0 text-left">
          <div class="sizecharts-section__step-header sizecharts-section__step-header--compact !mb-1 justify-start">
            <span class="sizecharts-section__step-badge">1</span>
            <h3 class="sizecharts-section__step-title">
              {{ t('guidesTireInnerTube.steps.size.title') }}
            </h3>
          </div>

          <p class="tz-text-secondary text-xs leading-snug">
            {{ t('guidesTireInnerTube.steps.size.description') }}
          </p>
        </div>

       <div class="shrink-0">
         <button
          type="button"
          class="inline-flex items-center justify-center rounded-full bg-[var(--tz-card-surface)] border tz-border-subtle px-5 py-2 text-xs font-bold uppercase tracking-wider tz-text-primary shadow-md hover:border-[rgba(5, 150, 105,0.28)] hover:tz-surface-subtle hover:text-[var(--tz-site-accent)] transition-colors"
          @click="openInnerTubeSearch"
        >
          {{ t('guidesTireInnerTube.actions.findTube') }}
        </button>
       </div>
       </div>
    </div>


    <!-- Step 2: Valve Selection Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] p-4 text-left shadow-md md:p-5">
      <div class="sizecharts-section__step-header sizecharts-section__step-header--compact justify-start">
        <span class="sizecharts-section__step-badge">2</span>
        <div>
          <h3 class="sizecharts-section__step-title">
            {{ t('guidesTireInnerTube.steps.valve.title') }}
          </h3>
          <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
            {{ t('guidesTireInnerTube.valves.introduction') }}
          </p>
        </div>
      </div>

      <div class="mx-auto mt-3 w-full max-w-[600px]">
        <GuideImage
          class="overflow-hidden rounded-xl shadow-sm"
          src="/images/guides/inner-tube/schwalbe-innertube-valve.webp"
          :alt="t('guidesTireInnerTube.valves.imageOverviewAlt')"
          :zoomOnClick="true"
          aspectRatio="2 / 1"
        />
      </div>

      <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <article
          v-for="valveType in valveTypes"
          :key="valveType"
          class="min-w-0 rounded-xl border tz-border-subtle p-3 shadow-sm"
        >
          <div class="flex items-baseline justify-between gap-2">
            <strong class="sizecharts-section__card-title tz-text-primary">
              {{ valveType.toUpperCase() }}
            </strong>
            <span class="sizecharts-section__card-meta tz-text-muted">
              {{ t(`guidesTireInnerTube.valves.${valveType}.type`) }}
            </span>
          </div>
          <p class="mt-2 text-xs leading-relaxed tz-text-secondary">
            {{ t(`guidesTireInnerTube.valves.${valveType}.description`) }}
          </p>
          <p class="mt-2 text-[11px] font-semibold leading-relaxed tz-text-primary">
            {{ t(`guidesTireInnerTube.valves.${valveType}.fit`) }}
          </p>
        </article>
      </div>

      <div class="mt-4 grid grid-cols-1 gap-3 lg:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]">
        <section class="rounded-xl border tz-border-subtle p-3">
          <h4 class="text-sm font-bold tz-text-primary">
            {{ t('guidesTireInnerTube.valves.measure.title') }}
          </h4>
          <ol class="mt-2 space-y-2 text-xs leading-relaxed tz-text-secondary">
            <li class="flex gap-2">
              <span class="font-bold tz-text-primary">1.</span>
              <span>{{ t('guidesTireInnerTube.valves.measure.stepOne') }}</span>
            </li>
            <li class="flex gap-2">
              <span class="font-bold tz-text-primary">2.</span>
              <span>{{ t('guidesTireInnerTube.valves.measure.stepTwo') }}</span>
            </li>
            <li class="flex gap-2">
              <span class="font-bold tz-text-primary">3.</span>
              <span>{{ t('guidesTireInnerTube.valves.measure.stepThree') }}</span>
            </li>
          </ol>
          <p class="mt-3 rounded-lg bg-amber-50/70 p-2 text-[11px] font-semibold leading-relaxed text-amber-900">
            {{ t('guidesTireInnerTube.valves.measure.warning') }}
          </p>
        </section>

        <section class="rounded-xl border tz-border-subtle p-3">
          <h4 class="text-sm font-bold tz-text-primary">
            {{ t('guidesTireInnerTube.valves.holeReference.title') }}
          </h4>
          <div class="mt-2 overflow-hidden rounded-lg border tz-border-subtle">
            <table class="min-w-full text-left text-[11px] leading-relaxed">
              <thead class="tz-surface-muted">
                <tr>
                  <th class="px-2 py-2 font-semibold tz-text-primary">
                    {{ t('guidesTireInnerTube.valves.holeReference.headers.valve') }}
                  </th>
                  <th class="px-2 py-2 font-semibold tz-text-primary">
                    {{ t('guidesTireInnerTube.valves.holeReference.headers.diameter') }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y tz-border-subtle">
                <tr v-for="reference in valveHoleReferences" :key="reference.type">
                  <td class="px-2 py-2 font-semibold tz-text-primary">
                    {{ t(`guidesTireInnerTube.valves.holeReference.rows.${reference.type}.valve`) }}
                  </td>
                  <td class="px-2 py-2 tz-text-secondary">
                    {{ t(`guidesTireInnerTube.valves.holeReference.rows.${reference.type}.diameter`) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <p class="mt-2 text-[11px] leading-relaxed tz-text-secondary">
            {{ t('guidesTireInnerTube.valves.holeReference.note') }}
          </p>
        </section>
      </div>

      <div class="mt-3 rounded-xl border border-sky-200/70 bg-sky-50/70 p-3">
        <h4 class="text-sm font-bold text-sky-950">
          {{ t('guidesTireInnerTube.valves.tubeless.title') }}
        </h4>
        <p class="mt-1 text-xs leading-relaxed text-sky-950/80">
          {{ t('guidesTireInnerTube.valves.tubeless.description') }}
        </p>
      </div>
    </div>

    <!-- Step 3: Tube Type Selection -->
     <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-4 md:p-5 text-center">
       <div class="sizecharts-section__step-header sizecharts-section__step-header--compact">
         <span class="sizecharts-section__step-badge">3</span>
           <h3 class="sizecharts-section__step-title">
             {{ t('guidesTireInnerTube.steps.model.title') }}
           </h3>
       </div>

       <div class="mb-5">
         <h4 class="text-sm font-semibold tz-text-primary">
           {{ t('guidesTireInnerTube.materials.title') }}
         </h4>
         <p class="mx-auto mt-1 max-w-3xl text-xs leading-relaxed tz-text-secondary">
           {{ t('guidesTireInnerTube.materials.description') }}
         </p>
         <div class="mt-3 grid grid-cols-1 gap-3 md:grid-cols-3">
           <div
             v-for="material in ['butyl', 'latex', 'tpu']"
             :key="material"
             class="rounded-xl border tz-border-subtle p-3 text-left"
           >
             <strong class="block text-sm font-bold tz-text-primary">
               {{ t(`guidesTireInnerTube.materials.${material}.title`) }}
             </strong>
             <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
               {{ t(`guidesTireInnerTube.materials.${material}.description`) }}
             </p>
             <p class="mt-2 text-[11px] font-semibold leading-relaxed tz-text-primary">
               {{ t(`guidesTireInnerTube.materials.${material}.fit`) }}
             </p>
           </div>
         </div>
       </div>

       <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-5">
          <!-- Standard -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md border tz-border-subtle hover:tz-surface-subtle transition-colors">
                <strong class="block tz-text-primary text-sm font-bold mb-2">{{ t('guidesTireInnerTube.models.standard.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ t('guidesTireInnerTube.models.standard.description') }}
              </p>
          </div>

          <!-- Air Plus -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md border tz-border-subtle hover:tz-surface-subtle transition-colors">
              <strong class="block tz-text-primary text-sm font-bold mb-2">{{ t('guidesTireInnerTube.models.airPlus.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ t('guidesTireInnerTube.models.airPlus.description') }}
              </p>
          </div>

          <!-- Extralight -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md border tz-border-subtle hover:tz-surface-subtle transition-colors">
              <strong class="block tz-text-primary text-sm font-bold mb-2">{{ t('guidesTireInnerTube.models.extralight.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ t('guidesTireInnerTube.models.extralight.description') }}
              </p>
          </div>

           <!-- Freeride -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md border tz-border-subtle hover:tz-surface-subtle transition-colors">
                <strong class="block tz-text-primary text-sm font-bold mb-2">{{ t('guidesTireInnerTube.models.freeride.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ t('guidesTireInnerTube.models.freeride.description') }}
              </p>
          </div>

           <!-- Downhill -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md border tz-border-subtle hover:tz-surface-subtle transition-colors">
                <strong class="block tz-text-primary text-sm font-bold mb-2">{{ t('guidesTireInnerTube.models.downhill.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ t('guidesTireInnerTube.models.downhill.description') }}
              </p>
          </div>
       </div>
    </div>

    <div class="rounded-2xl border border-amber-200/70 bg-amber-50/60 p-4 text-left shadow-md md:p-5">
      <div class="sizecharts-section__step-header sizecharts-section__step-header--compact justify-start">
        <span class="sizecharts-section__step-badge !bg-amber-100 !text-amber-700">!</span>
        <div>
          <h3 class="sizecharts-section__step-title">
            {{ t('guidesTireInnerTube.advanced.title') }}
          </h3>
          <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
            {{ t('guidesTireInnerTube.advanced.description') }}
          </p>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
        <article class="rounded-xl border border-amber-200/70 bg-[var(--tz-card-surface)] p-3">
          <h4 class="text-sm font-bold tz-text-primary">
            {{ t('guidesTireInnerTube.advanced.carbonValveHole.title') }}
          </h4>
          <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
            {{ t('guidesTireInnerTube.advanced.carbonValveHole.description') }}
          </p>
          <p class="mt-2 text-[11px] font-semibold leading-relaxed tz-text-primary">
            {{ t('guidesTireInnerTube.advanced.carbonValveHole.rule') }}
          </p>
        </article>

        <article class="rounded-xl border border-amber-200/70 bg-[var(--tz-card-surface)] p-3">
          <h4 class="text-sm font-bold tz-text-primary">
            {{ t('guidesTireInnerTube.advanced.co2Seal.title') }}
          </h4>
          <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
            {{ t('guidesTireInnerTube.advanced.co2Seal.description') }}
          </p>
          <p class="mt-2 text-[11px] font-semibold leading-relaxed tz-text-primary">
            {{ t('guidesTireInnerTube.advanced.co2Seal.rule') }}
          </p>
        </article>
      </div>
    </div>

    </div>

    <div
      id="inner-tube-calculator-panel"
      v-show="activeInnerTubeGuideTab === 'calculator'"
      role="tabpanel"
      aria-labelledby="inner-tube-calculator-tab"
      class="space-y-5 md:space-y-6"
    >
      <InnerTubeValveLengthAndExtenderFitmentGuide />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'
import { usePageMessages } from '~/composables/usePageMessages'
import InnerTubeValveLengthAndExtenderFitmentGuide from '~/components/tireguides/InnerTubeValveLengthAndExtenderFitmentGuide.vue'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTireInnerTube')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const { open: openShopSearchSheet } = useShopSearchSheet()
const activeInnerTubeGuideTab = ref<'introduction' | 'calculator'>('introduction')

type InnerTubeValveType = 'av' | 'dv' | 'sv'

const valveTypes: InnerTubeValveType[] = ['av', 'dv', 'sv']
const valveHoleReferences: Array<{ type: InnerTubeValveType }> = [
  { type: 'sv' },
  { type: 'av' },
  { type: 'dv' },
]

const openInnerTubeSearch = () => {
  openShopSearchSheet({
    presetCategorySlug: 'inner-tube',
    presetKeywords: ['Inner tube'],
  })
}
</script>
