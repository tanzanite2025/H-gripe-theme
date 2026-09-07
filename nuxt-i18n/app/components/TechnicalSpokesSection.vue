<template>
  <div>
    <div class="nav-pill-tabs mb-6" role="tablist">
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'basic-concepts' }"
        @click="activeTab = 'basic-concepts'"
      >
        {{ t('guidesWheelsetComponentsSpokes.tabs.basic') }}
      </button>
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'choose-spoke' }"
        @click="activeTab = 'choose-spoke'"
      >
        {{ t('guidesWheelsetComponentsSpokes.tabs.choose') }}
      </button>
    </div>

    <div v-if="activeTab === 'basic-concepts'">
      <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center mb-8">
        <h3 class="text-xl font-bold tz-text-secondary mb-4 flex items-center justify-center gap-2">
          {{ t('guidesWheelsetComponentsSpokes.basic.title') }}
        </h3>
        <p class="tz-text-secondary text-sm leading-relaxed mb-6 max-w-2xl mx-auto hidden">
          {{ t('guidesWheelsetComponentsSpokes.basic.intro') }}
        </p>
        <div class="w-full mb-8">
          <p class="text-xs font-bold uppercase tracking-wider tz-text-muted mb-3">
            {{ t('guidesWheelsetComponentsSpokes.basic.profilesTitle') }}
          </p>
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <GuideImage
              src="/public/technical/j-bend-spokeandstraightpull-spoke.webp"
              :alt="t('guidesWheelsetComponentsSpokes.basic.profiles.comparisonAlt')"
              :zoomOnClick="true"
              :caption="t('guidesWheelsetComponentsSpokes.basic.profiles.comparisonCaption')"
              class="rounded-xl overflow-hidden shadow-md"
            />
            <GuideImage
              src="/public/technical/Sapim-black-stainless-steel-straight-gauge-2.0mm-spoke-double-butted-2.0-1.8-2.0mm-spoke-aero-flat-bladed-spoke.webp"
              :alt="t('guidesWheelsetComponentsSpokes.basic.profiles.shapesAlt')"
              :zoomOnClick="true"
              :caption="t('guidesWheelsetComponentsSpokes.basic.profiles.shapesCaption')"
              class="rounded-xl overflow-hidden shadow-md"
            />
          </div>
        </div>

        <div class="w-full mb-8 text-center">
          <div class="flex items-center gap-3 mb-4 justify-center">
            <span class="h-px w-8 tz-surface-panel"></span>
            <h4 class="tz-text-primary font-bold text-base uppercase tracking-wider">{{ t('guidesWheelsetComponentsSpokes.basic.brandsTitle') }}</h4>
            <span class="h-px w-8 tz-surface-panel"></span>
          </div>
          <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <div v-for="brand in brands" :key="brand.id" class="tz-surface-panel p-3 rounded-lg shadow-md flex flex-col items-center text-center">
              <strong class="block text-sm mb-1" :class="brand.nameClass">{{ t(brand.nameKey) }}</strong>
              <p class="text-xs tz-text-secondary leading-relaxed mb-3">{{ t(brand.bodyKey) }}</p>
              <button class="mt-auto px-4 py-1.5 rounded-full border tz-caption font-bold uppercase tracking-wider transition-colors" :class="brand.buttonClass">
                {{ t(brand.buttonKey) }}
              </button>
            </div>
          </div>
        </div>

        <div class="w-full mb-8 text-center">
          <div class="flex items-center gap-3 mb-4 justify-center">
            <span class="h-px w-8 tz-surface-panel"></span>
            <h4 class="tz-text-primary font-bold text-base uppercase tracking-wider">{{ t('guidesWheelsetComponentsSpokes.basic.typesTitle') }}</h4>
            <span class="h-px w-8 tz-surface-panel"></span>
          </div>
          <div class="space-y-4">
            <div v-for="type in spokeTypes" :key="type.id" class="tz-surface-panel rounded-xl p-4 shadow-md flex flex-col items-center text-center">
              <h5 class="tz-text-secondary font-bold mb-2 flex items-center justify-center gap-2">
                <span class="w-2 h-2 rounded-full" :class="type.dotClass"></span>{{ t(type.titleKey) }}
              </h5>
              <div class="grid md:grid-cols-2 gap-x-6 gap-y-2 text-sm tz-text-secondary w-full">
                <p><span class="tz-text-muted font-mono text-xs uppercase mr-1">{{ type.labels.specs }}:</span> {{ t(type.specsKey) }}</p>
                <p><span class="tz-text-muted font-mono text-xs uppercase mr-1">{{ type.labels.features }}:</span> {{ t(type.featuresKey) }}</p>
                <p><span class="tz-text-muted font-mono text-xs uppercase mr-1">{{ type.labels.examples }}:</span> {{ t(type.examplesKey) }}</p>
                <p><span class="tz-text-muted font-mono text-xs uppercase mr-1">{{ type.labels.use }}:</span> {{ t(type.useKey) }}</p>
              </div>
            </div>
          </div>
        </div>

        <div class="w-full text-center">
          <div class="flex items-center gap-3 mb-4 justify-center">
            <span class="h-px w-8 tz-surface-panel"></span>
            <h4 class="tz-text-primary font-bold text-base uppercase tracking-wider">{{ t('guidesWheelsetComponentsSpokes.basic.interfacesTitle') }}</h4>
            <span class="h-px w-8 tz-surface-panel"></span>
          </div>
          <div class="grid gap-3 sm:grid-cols-3">
            <div v-for="item in interfaces" :key="item.titleKey" class="tz-surface-panel p-3 rounded-lg border tz-border-subtle text-center">
              <strong class="block tz-text-primary text-sm mb-1">{{ t(item.titleKey) }}</strong>
              <p class="text-xs tz-text-muted">{{ t(item.bodyKey) }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="mt-4">
        <div class="flex items-center justify-center gap-2 mb-4">
          <div class="h-px w-12 tz-surface-panel"></div>
          <p class="text-xs font-bold uppercase tracking-widest tz-text-muted">{{ t('guidesWheelsetComponentsSpokes.basic.materialTitle') }}</p>
          <div class="h-px w-12 tz-surface-panel"></div>
        </div>

        <div class="bg-amber-900/10 border-l-2 border-amber-600/50 p-3 mb-6 mx-auto max-w-3xl rounded-r-lg">
          <div class="flex gap-3">
            <span class="text-amber-500 shrink-0 mt-0.5" aria-hidden="true">i</span>
            <p class="text-xs text-amber-500/80 leading-relaxed text-left">
              <strong class="font-bold text-amber-500">{{ t('guidesWheelsetComponentsSpokes.basic.disclaimerLabel') }}:</strong>
              {{ t('guidesWheelsetComponentsSpokes.basic.disclaimer') }}
            </p>
          </div>
        </div>

        <ul class="technical-spokes-list mt-8 space-y-8 w-full">
          <li
            v-for="material in materials"
            :key="material.id"
            class="rounded-2xl shadow-md p-5 md:p-6 transition-transform hover:-translate-y-1 duration-300 text-center"
            :class="material.cardClass"
          >
            <div class="flex items-center justify-center gap-3 mb-4">
              <span class="flex h-8 w-8 items-center justify-center rounded-lg text-lg font-bold" :class="material.numberClass">{{ material.number }}</span>
              <strong class="text-xl tz-text-secondary">{{ t(material.titleKey) }}</strong>
              <span v-if="material.badgeKey" class="tz-micro-label font-bold uppercase tracking-wider px-2 py-1 rounded-full" :class="material.badgeClass">{{ t(material.badgeKey) }}</span>
            </div>
            <div class="w-full tz-text-secondary text-sm leading-relaxed flex flex-col items-center">
              <p class="mb-4 text-center">{{ t(material.bodyKey) }}</p>
              <div v-if="material.marketTitleKey" class="mb-6 w-full text-center">
                <strong class="text-xs uppercase tracking-wider tz-text-muted mb-2 block">{{ t(material.marketTitleKey) }}</strong>
                <p class="text-sm tz-text-secondary max-w-2xl mx-auto">{{ t(material.marketBodyKey) }}</p>
              </div>
              <div class="grid gap-4 md:grid-cols-2 w-full">
                <div class="bg-emerald-50 rounded-lg p-3 shadow-md w-full">
                  <strong class="block text-xs uppercase tracking-wider text-emerald-600 mb-2 text-center">{{ t('guidesWheelsetComponentsSpokes.basic.pros') }}</strong>
                  <ul class="space-y-1.5 text-sm list-none pl-0 my-0 flex flex-col items-center">
                    <li v-for="key in material.prosKeys" :key="key" class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                      <span class="text-emerald-600 shrink-0 text-center w-4">✓</span><span class="text-center">{{ t(key) }}</span>
                    </li>
                  </ul>
                </div>
                <div class="bg-rose-500/5 rounded-lg p-3 shadow-md w-full">
                  <strong class="block text-xs uppercase tracking-wider text-rose-400 mb-2 text-center">{{ t('guidesWheelsetComponentsSpokes.basic.cons') }}</strong>
                  <ul class="space-y-1.5 text-sm list-none pl-0 my-0 flex flex-col items-center">
                    <li v-for="key in material.consKeys" :key="key" class="flex items-center justify-center gap-2 pl-0 before:content-none w-full">
                      <span class="text-rose-500 shrink-0 text-center w-4">×</span><span class="text-center">{{ t(key) }}</span>
                    </li>
                  </ul>
                </div>
              </div>
              <p class="mt-4 text-xs font-medium tz-text-muted uppercase tracking-widest text-center">{{ t('guidesWheelsetComponentsSpokes.basic.bestForLabel') }}</p>
              <p class="mt-1 tz-text-primary text-center">{{ t(material.bestForKey) }}</p>
            </div>
          </li>
        </ul>
        <p class="sizecharts-section__intro mt-3">{{ t('guidesWheelsetComponentsSpokes.basic.conclusion') }}</p>
      </div>
    </div>

    <div v-else>
      <div class="space-y-8">
        <div class="bg-[var(--tz-card-surface)] rounded-2xl shadow-md border-t-4 border-emerald-500 p-6 text-center">
          <h3 class="text-xl font-bold tz-text-secondary mb-2">{{ t('guidesWheelsetComponentsSpokes.choose.title') }}</h3>
          <p class="tz-text-secondary text-sm max-w-2xl mx-auto">{{ t('guidesWheelsetComponentsSpokes.choose.intro') }}</p>
        </div>

        <div class="overflow-x-auto rounded-lg shadow-md border tz-border-subtle">
          <table class="w-full text-sm text-left tz-text-secondary">
            <thead class="text-xs tz-text-primary uppercase tz-surface-panel">
              <tr>
                <th v-for="key in spokeHeaderKeys" :key="key" scope="col" class="px-6 py-4 font-bold">{{ t(key) }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200">
              <tr
                v-for="row in spokeRows"
                :key="row.nameKey"
                class="bg-[var(--tz-card-surface)] hover:tz-surface-panel transition-colors"
                :class="{ 'border-t-4 tz-border-subtle': row.separator, 'tz-surface-card relative overflow-hidden group border-t border-emerald-200': row.carbon }"
              >
                <td class="px-6 py-4 relative z-10">
                  <div v-if="row.carbon" class="flex items-center gap-3">
                    <div class="w-1 h-8 rounded-full bg-emerald-500"></div>
                    <strong class="flex items-center gap-2 text-emerald-600 text-base uppercase tracking-wide">{{ t(row.nameKey) }}</strong>
                  </div>
                  <div v-else class="flex items-center gap-2">
                    <strong class="block tz-text-primary text-base">{{ t(row.nameKey) }}</strong>
                    <span v-if="row.top" class="tz-micro-label bg-amber-500 text-slate-900 px-1.5 py-0.5 rounded font-bold">{{ t('guidesWheelsetComponentsSpokes.choose.top') }}</span>
                  </div>
                </td>
                <td class="px-6 py-4 text-center relative z-10">
                  <span class="inline-flex px-2 py-1 rounded text-xs font-bold border" :class="row.typeClass">{{ t(row.typeKey) }}</span>
                </td>
                <td class="px-6 py-4 text-center font-mono relative z-10" :class="row.comingSoon ? 'tz-text-muted italic' : 'tz-text-primary'">{{ t(row.weightKey) }}</td>
                <td class="px-6 py-4 text-center font-mono text-xs relative z-10" :class="{ 'tz-text-muted italic': row.comingSoon }">{{ t(row.sectionKey) }}</td>
                <td class="px-6 py-4 text-xs leading-relaxed min-w-[200px] relative z-10" :class="{ 'tz-text-muted italic': row.comingSoon && !row.carbon, 'text-emerald-200/80 italic': row.carbon }">{{ t(row.detailsKey) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="grid md:grid-cols-2 gap-4">
          <div class="tz-surface-panel border tz-border-subtle p-4 rounded-xl flex gap-4 items-start">
            <span class="tz-surface-panel p-2 rounded-lg tz-text-secondary shrink-0" aria-hidden="true">→</span>
            <div>
              <h5 class="tz-text-primary font-bold text-sm mb-1">{{ t('guidesWheelsetComponentsSpokes.choose.weightTitle') }}</h5>
              <p class="text-xs tz-text-secondary">{{ t('guidesWheelsetComponentsSpokes.choose.weightBody') }}</p>
            </div>
          </div>
          <div class="bg-emerald-50 border border-emerald-200 p-4 rounded-xl flex gap-4 items-start">
            <span class="bg-emerald-100 p-2 rounded-lg text-emerald-600 shrink-0" aria-hidden="true">↗</span>
            <div>
              <h5 class="text-emerald-700 font-bold text-sm mb-1">{{ t('guidesWheelsetComponentsSpokes.choose.customTitle') }}</h5>
              <p class="text-xs text-emerald-600/70">{{ t('guidesWheelsetComponentsSpokes.choose.customBody') }}</p>
            </div>
          </div>
        </div>

        <div class="mt-8 rounded-2xl tz-surface-card p-6 shadow-md border tz-border-subtle">
          <div class="flex flex-col md:flex-row items-center justify-between gap-6 text-center md:text-left">
            <div class="space-y-2">
              <h4 class="text-lg font-bold tz-text-secondary">{{ t('guidesWheelsetComponentsSpokes.choose.helpTitle') }}</h4>
              <p class="text-sm tz-text-secondary max-w-xl leading-relaxed">{{ t('guidesWheelsetComponentsSpokes.choose.helpBody') }}</p>
            </div>
            <div class="flex flex-col sm:flex-row gap-3 shrink-0">
              <button type="button" class="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-full bg-[var(--tz-action-primary)] hover:bg-[var(--tz-action-primary-hover)] text-white text-sm font-bold shadow-lg shadow-slate-900/20 transition-all" @click="goToSpokeCalculator">
                {{ t('guidesWheelsetComponentsSpokes.choose.calculator') }}
              </button>
              <button type="button" class="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-full bg-emerald-500 hover:bg-emerald-400 tz-text-primary text-sm font-bold shadow-lg shadow-emerald-500/20 transition-all" @click="openSupportChat">
                {{ t('guidesWheelsetComponentsSpokes.choose.liveChat') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n, useLocalePath, useRouter } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { useChatWidget } from '~/composables/useChatWidget'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesWheelsetComponentsSpokes')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const router = useRouter()
const localePath = useLocalePath()
const { openChat } = useChatWidget()
const activeTab = ref<'basic-concepts' | 'choose-spoke'>('basic-concepts')

const brands = [
  {
    id: 'dt-swiss',
    nameKey: 'guidesWheelsetComponentsSpokes.basic.brands.dtSwiss.name',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.brands.dtSwiss.body',
    buttonKey: 'guidesWheelsetComponentsSpokes.basic.brands.dtSwiss.button',
    nameClass: 'text-emerald-600',
    buttonClass: 'border-emerald-300 text-emerald-600 hover:bg-emerald-50',
  },
  {
    id: 'sapim',
    nameKey: 'guidesWheelsetComponentsSpokes.basic.brands.sapim.name',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.brands.sapim.body',
    buttonKey: 'guidesWheelsetComponentsSpokes.basic.brands.sapim.button',
    nameClass: 'text-emerald-600',
    buttonClass: 'border-emerald-300 text-emerald-600 hover:bg-emerald-50',
  },
  {
    id: 'pillar',
    nameKey: 'guidesWheelsetComponentsSpokes.basic.brands.pillar.name',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.brands.pillar.body',
    buttonKey: 'guidesWheelsetComponentsSpokes.basic.brands.pillar.button',
    nameClass: 'text-amber-400',
    buttonClass: 'border-amber-500/30 text-amber-400 hover:bg-amber-500/10',
  },
  {
    id: 'mac',
    nameKey: 'guidesWheelsetComponentsSpokes.basic.brands.mac.name',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.brands.mac.body',
    buttonKey: 'guidesWheelsetComponentsSpokes.basic.brands.mac.button',
    nameClass: 'tz-text-primary',
    buttonClass: 'border-slate-500/30 tz-text-secondary hover:bg-slate-500/10',
  },
  {
    id: 'yt',
    nameKey: 'guidesWheelsetComponentsSpokes.basic.brands.yt.name',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.brands.yt.body',
    buttonKey: 'guidesWheelsetComponentsSpokes.basic.brands.yt.button',
    nameClass: 'tz-text-primary',
    buttonClass: 'border-slate-500/30 tz-text-secondary hover:bg-slate-500/10',
  },
  {
    id: 'carbon',
    nameKey: 'guidesWheelsetComponentsSpokes.basic.brands.carbon.name',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.brands.carbon.body',
    buttonKey: 'guidesWheelsetComponentsSpokes.basic.brands.carbon.button',
    nameClass: 'text-emerald-600',
    buttonClass: 'border-emerald-300 text-emerald-600 hover:bg-emerald-50',
  },
]

const spokeTypes = [
  {
    id: 'straight',
    titleKey: 'guidesWheelsetComponentsSpokes.basic.types.straight.title',
    specsKey: 'guidesWheelsetComponentsSpokes.basic.types.straight.specs',
    featuresKey: 'guidesWheelsetComponentsSpokes.basic.types.straight.features',
    examplesKey: 'guidesWheelsetComponentsSpokes.basic.types.straight.examples',
    useKey: 'guidesWheelsetComponentsSpokes.basic.types.straight.use',
    dotClass: 'bg-slate-500',
    labels: { specs: 'Specs', features: 'Features', examples: 'Examples', use: 'Use' },
  },
  {
    id: 'butted',
    titleKey: 'guidesWheelsetComponentsSpokes.basic.types.butted.title',
    specsKey: 'guidesWheelsetComponentsSpokes.basic.types.butted.specs',
    featuresKey: 'guidesWheelsetComponentsSpokes.basic.types.butted.features',
    examplesKey: 'guidesWheelsetComponentsSpokes.basic.types.butted.examples',
    useKey: 'guidesWheelsetComponentsSpokes.basic.types.butted.use',
    dotClass: 'bg-emerald-500',
    labels: { specs: 'Specs', features: 'Features', examples: 'Examples', use: 'Use' },
  },
  {
    id: 'aero',
    titleKey: 'guidesWheelsetComponentsSpokes.basic.types.aero.title',
    specsKey: 'guidesWheelsetComponentsSpokes.basic.types.aero.specs',
    featuresKey: 'guidesWheelsetComponentsSpokes.basic.types.aero.features',
    examplesKey: 'guidesWheelsetComponentsSpokes.basic.types.aero.examples',
    useKey: 'guidesWheelsetComponentsSpokes.basic.types.aero.use',
    dotClass: 'bg-emerald-500',
    labels: { specs: 'Specs', features: 'Features', examples: 'Examples', use: 'Use' },
  },
]

const interfaces = [
  {
    titleKey: 'guidesWheelsetComponentsSpokes.basic.interfaces.jBend.title',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.interfaces.jBend.body',
  },
  {
    titleKey: 'guidesWheelsetComponentsSpokes.basic.interfaces.straightPull.title',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.interfaces.straightPull.body',
  },
  {
    titleKey: 'guidesWheelsetComponentsSpokes.basic.interfaces.nipples.title',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.interfaces.nipples.body',
  },
]

const materialConfig = [
  {
    id: 'stainless',
    number: 1,
    titleKey: 'guidesWheelsetComponentsSpokes.basic.materials.stainless.title',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.materials.stainless.body',
    marketTitleKey: 'guidesWheelsetComponentsSpokes.basic.materials.stainless.marketTitle',
    marketBodyKey: 'guidesWheelsetComponentsSpokes.basic.materials.stainless.marketBody',
    prosKeys: [
      'guidesWheelsetComponentsSpokes.basic.materials.stainless.pros.0',
      'guidesWheelsetComponentsSpokes.basic.materials.stainless.pros.1',
      'guidesWheelsetComponentsSpokes.basic.materials.stainless.pros.2',
      'guidesWheelsetComponentsSpokes.basic.materials.stainless.pros.3',
    ],
    consKeys: ['guidesWheelsetComponentsSpokes.basic.materials.stainless.cons.0'],
    bestForKey: 'guidesWheelsetComponentsSpokes.basic.materials.stainless.bestFor',
    cardClass: 'bg-[var(--tz-card-surface)]',
    numberClass: 'bg-amber-500/10 text-amber-500',
  },
  {
    id: 'carbon',
    number: 2,
    titleKey: 'guidesWheelsetComponentsSpokes.basic.materials.carbon.title',
    badgeKey: 'guidesWheelsetComponentsSpokes.basic.materials.carbon.badge',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.materials.carbon.body',
    prosKeys: [
      'guidesWheelsetComponentsSpokes.basic.materials.carbon.pros.0',
      'guidesWheelsetComponentsSpokes.basic.materials.carbon.pros.1',
      'guidesWheelsetComponentsSpokes.basic.materials.carbon.pros.2',
    ],
    consKeys: [
      'guidesWheelsetComponentsSpokes.basic.materials.carbon.cons.0',
      'guidesWheelsetComponentsSpokes.basic.materials.carbon.cons.1',
      'guidesWheelsetComponentsSpokes.basic.materials.carbon.cons.2',
    ],
    bestForKey: 'guidesWheelsetComponentsSpokes.basic.materials.carbon.bestFor',
    cardClass: 'tz-surface-card border-l-4 border-emerald-500',
    numberClass: 'bg-emerald-50 text-emerald-600',
    badgeClass: 'bg-emerald-500 tz-text-primary',
  },
  {
    id: 'titanium',
    number: 3,
    titleKey: 'guidesWheelsetComponentsSpokes.basic.materials.titanium.title',
    badgeKey: 'guidesWheelsetComponentsSpokes.basic.materials.titanium.badge',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.materials.titanium.body',
    prosKeys: [
      'guidesWheelsetComponentsSpokes.basic.materials.titanium.pros.0',
      'guidesWheelsetComponentsSpokes.basic.materials.titanium.pros.1',
      'guidesWheelsetComponentsSpokes.basic.materials.titanium.pros.2',
    ],
    consKeys: [
      'guidesWheelsetComponentsSpokes.basic.materials.titanium.cons.0',
      'guidesWheelsetComponentsSpokes.basic.materials.titanium.cons.1',
    ],
    bestForKey: 'guidesWheelsetComponentsSpokes.basic.materials.titanium.bestFor',
    cardClass: 'bg-[var(--tz-card-surface)] opacity-90',
    numberClass: 'tz-surface-panel tz-text-secondary',
    badgeClass: 'border tz-border-subtle tz-text-muted',
  },
  {
    id: 'aluminum',
    number: 4,
    titleKey: 'guidesWheelsetComponentsSpokes.basic.materials.aluminum.title',
    badgeKey: 'guidesWheelsetComponentsSpokes.basic.materials.aluminum.badge',
    bodyKey: 'guidesWheelsetComponentsSpokes.basic.materials.aluminum.body',
    prosKeys: [
      'guidesWheelsetComponentsSpokes.basic.materials.aluminum.pros.0',
      'guidesWheelsetComponentsSpokes.basic.materials.aluminum.pros.1',
    ],
    consKeys: [
      'guidesWheelsetComponentsSpokes.basic.materials.aluminum.cons.0',
      'guidesWheelsetComponentsSpokes.basic.materials.aluminum.cons.1',
      'guidesWheelsetComponentsSpokes.basic.materials.aluminum.cons.2',
    ],
    bestForKey: 'guidesWheelsetComponentsSpokes.basic.materials.aluminum.bestFor',
    cardClass: 'bg-[var(--tz-card-surface)] opacity-90',
    numberClass: 'tz-surface-panel tz-text-secondary',
    badgeClass: 'border tz-border-subtle tz-text-muted',
  },
]

const materials = materialConfig.map((material) => ({
  ...material,
  titleKey: material.titleKey,
}))

const spokeHeaderKeys = [
  'guidesWheelsetComponentsSpokes.choose.headers.model',
  'guidesWheelsetComponentsSpokes.choose.headers.type',
  'guidesWheelsetComponentsSpokes.choose.headers.weight',
  'guidesWheelsetComponentsSpokes.choose.headers.section',
  'guidesWheelsetComponentsSpokes.choose.headers.details',
]

const spokeRows = [
  {
    nameKey: 'guidesWheelsetComponentsSpokes.choose.rows.0.name',
    typeKey: 'guidesWheelsetComponentsSpokes.choose.type.straight',
    weightKey: 'guidesWheelsetComponentsSpokes.choose.rows.0.weight',
    sectionKey: 'guidesWheelsetComponentsSpokes.choose.rows.0.section',
    detailsKey: 'guidesWheelsetComponentsSpokes.choose.rows.0.details',
    typeClass: 'tz-surface-panel tz-text-secondary border tz-border-subtle',
  },
  {
    nameKey: 'guidesWheelsetComponentsSpokes.choose.rows.1.name',
    typeKey: 'guidesWheelsetComponentsSpokes.choose.type.butted',
    weightKey: 'guidesWheelsetComponentsSpokes.choose.rows.1.weight',
    sectionKey: 'guidesWheelsetComponentsSpokes.choose.rows.1.section',
    detailsKey: 'guidesWheelsetComponentsSpokes.choose.rows.1.details',
    typeClass: 'bg-emerald-50 text-emerald-600 border-emerald-200',
  },
  {
    nameKey: 'guidesWheelsetComponentsSpokes.choose.rows.2.name',
    typeKey: 'guidesWheelsetComponentsSpokes.choose.type.butted',
    weightKey: 'guidesWheelsetComponentsSpokes.choose.rows.2.weight',
    sectionKey: 'guidesWheelsetComponentsSpokes.choose.rows.2.section',
    detailsKey: 'guidesWheelsetComponentsSpokes.choose.rows.2.details',
    typeClass: 'bg-emerald-50 text-emerald-600 border-emerald-200',
  },
  {
    nameKey: 'guidesWheelsetComponentsSpokes.choose.rows.3.name',
    typeKey: 'guidesWheelsetComponentsSpokes.choose.type.aero',
    weightKey: 'guidesWheelsetComponentsSpokes.choose.rows.3.weight',
    sectionKey: 'guidesWheelsetComponentsSpokes.choose.rows.3.section',
    detailsKey: 'guidesWheelsetComponentsSpokes.choose.rows.3.details',
    typeClass: 'bg-emerald-50 text-emerald-600 border-emerald-200',
  },
  {
    nameKey: 'guidesWheelsetComponentsSpokes.choose.rows.4.name',
    typeKey: 'guidesWheelsetComponentsSpokes.choose.type.aero',
    weightKey: 'guidesWheelsetComponentsSpokes.choose.rows.4.weight',
    sectionKey: 'guidesWheelsetComponentsSpokes.choose.rows.4.section',
    detailsKey: 'guidesWheelsetComponentsSpokes.choose.rows.4.details',
    typeClass: 'bg-emerald-50 text-emerald-600 border-emerald-200',
    top: true,
  },
  {
    nameKey: 'guidesWheelsetComponentsSpokes.choose.rows.5.name',
    typeKey: 'guidesWheelsetComponentsSpokes.choose.type.aero',
    weightKey: 'guidesWheelsetComponentsSpokes.choose.rows.5.weight',
    sectionKey: 'guidesWheelsetComponentsSpokes.choose.rows.5.section',
    detailsKey: 'guidesWheelsetComponentsSpokes.choose.rows.5.details',
    typeClass: 'bg-emerald-50 text-emerald-600 border-emerald-200',
  },
  {
    nameKey: 'guidesWheelsetComponentsSpokes.choose.rows.6.name',
    typeKey: 'guidesWheelsetComponentsSpokes.choose.type.butted',
    weightKey: 'guidesWheelsetComponentsSpokes.choose.rows.6.weight',
    sectionKey: 'guidesWheelsetComponentsSpokes.choose.rows.6.section',
    detailsKey: 'guidesWheelsetComponentsSpokes.choose.rows.6.details',
    typeClass: 'bg-emerald-50 text-emerald-600 border-emerald-200',
    comingSoon: true,
    separator: true,
  },
  {
    nameKey: 'guidesWheelsetComponentsSpokes.choose.rows.7.name',
    typeKey: 'guidesWheelsetComponentsSpokes.choose.type.aero',
    weightKey: 'guidesWheelsetComponentsSpokes.choose.rows.7.weight',
    sectionKey: 'guidesWheelsetComponentsSpokes.choose.rows.7.section',
    detailsKey: 'guidesWheelsetComponentsSpokes.choose.rows.7.details',
    typeClass: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
    comingSoon: true,
  },
  {
    nameKey: 'guidesWheelsetComponentsSpokes.choose.rows.8.name',
    typeKey: 'guidesWheelsetComponentsSpokes.choose.type.aero',
    weightKey: 'guidesWheelsetComponentsSpokes.choose.rows.8.weight',
    sectionKey: 'guidesWheelsetComponentsSpokes.choose.rows.8.section',
    detailsKey: 'guidesWheelsetComponentsSpokes.choose.rows.8.details',
    typeClass: 'bg-emerald-500 tz-text-primary border-emerald-500',
    comingSoon: true,
    carbon: true,
  },
]

const goToSpokeCalculator = async () => {
  await router.push(localePath('/resources/spoke-calculator'))
}

const openSupportChat = () => {
  openChat({ showAgentList: true })
}
</script>

<style src="~/assets/css/guide-sections.css"></style>
