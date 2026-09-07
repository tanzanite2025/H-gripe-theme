<template>
  <div>
    <div class="nav-pill-tabs mb-6" role="tablist">
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'technology' }"
        @click="activeTab = 'technology'"
      >
        {{ t('guidesWheelsetComponentsRims.tabs.technology') }}
      </button>
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'how-to-choose' }"
        @click="activeTab = 'how-to-choose'"
      >
        {{ t('guidesWheelsetComponentsRims.tabs.choose') }}
      </button>
    </div>

    <div v-if="activeTab === 'technology'">
      <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center mb-8 border-t-4 border-emerald-500">
        <h3 class="text-xl font-bold tz-text-secondary mb-4 flex items-center justify-center gap-2">
          {{ t('guidesWheelsetComponentsRims.technology.title') }}
        </h3>
        <p class="tz-text-secondary text-sm leading-relaxed mb-6 max-w-2xl mx-auto">
          {{ t('guidesWheelsetComponentsRims.technology.intro') }}
        </p>
        <div class="grid grid-cols-3 gap-2 md:gap-4 max-w-lg mx-auto mb-4">
          <div v-for="metric in metrics" :key="metric.valueKey" class="bg-emerald-50 rounded-lg p-3">
            <span class="block text-xl font-bold text-emerald-600">{{ t(metric.valueKey) }}</span>
            <span class="tz-micro-label uppercase tracking-wider tz-text-muted">{{ t(metric.labelKey) }}</span>
          </div>
        </div>
      </div>

      <div class="mb-10">
        <div class="flex items-center justify-center gap-2 mb-6">
          <div class="h-px w-8 tz-surface-panel"></div>
          <h4 class="tz-text-primary font-bold text-base uppercase tracking-wider">
            {{ t('guidesWheelsetComponentsRims.technology.profilesTitle') }}
          </h4>
          <div class="h-px w-8 tz-surface-panel"></div>
        </div>
        <div class="grid md:grid-cols-2 gap-6">
          <div v-for="profile in profiles" :key="profile.titleKey" class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 transition-transform hover:-translate-y-1">
            <div class="flex justify-between items-start mb-4">
              <span class="text-sm font-bold tz-text-secondary">{{ t(profile.titleKey) }}</span>
              <span class="text-xs tz-surface-panel tz-text-muted px-2 py-1 rounded">{{ t(profile.depthKey) }}</span>
            </div>
            <p class="text-xs tz-text-secondary leading-relaxed mb-3">{{ t(profile.bodyKey) }}</p>
            <div class="text-xs text-emerald-600 font-medium uppercase tracking-wide">{{ t(profile.bestForKey) }}</div>
          </div>
        </div>
      </div>

      <div class="mb-10">
        <div class="flex items-center justify-center gap-2 mb-6">
          <div class="h-px w-8 tz-surface-panel"></div>
          <h4 class="tz-text-primary font-bold text-base uppercase tracking-wider">
            {{ t('guidesWheelsetComponentsRims.technology.widthTitle') }}
          </h4>
          <div class="h-px w-8 tz-surface-panel"></div>
        </div>
        <div class="tz-surface-panel rounded-2xl p-6 shadow-md">
          <div class="flex flex-col md:flex-row gap-8 items-center">
            <div class="flex-1 space-y-4">
              <div v-for="item in widthItems" :key="item.titleKey">
                <strong class="block tz-text-primary text-sm mb-1">{{ t(item.titleKey) }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">{{ t(item.bodyKey) }}</p>
              </div>
              <div class="p-3 bg-emerald-50 border-l-2 border-emerald-500 text-xs text-emerald-200">
                <strong>{{ t('guidesWheelsetComponentsRims.technology.trend.label') }}:</strong>
                {{ t('guidesWheelsetComponentsRims.technology.trend.body') }}
              </div>
            </div>
            <div class="w-full md:w-1/3">
              <GuideImage
                src="/public/wheelsetbuyersguide/wheelcomponents/rim/carbon-rim-hooked-vs-hookless.webp"
                :alt="t('guidesWheelsetComponentsRims.technology.widthImage.alt')"
                :zoomOnClick="true"
                :caption="t('guidesWheelsetComponentsRims.technology.widthImage.caption')"
                class="rounded-lg w-full shadow-lg"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center border-b-2 tz-border-subtle">
          <h5 class="tz-text-primary font-bold mb-3">{{ t('guidesWheelsetComponentsRims.technology.hooked.title') }}</h5>
          <p class="text-xs tz-text-secondary leading-relaxed mb-4 min-h-[3rem]">{{ t('guidesWheelsetComponentsRims.technology.hooked.body') }}</p>
          <span class="inline-block px-3 py-1 tz-surface-panel tz-text-secondary tz-micro-label uppercase font-bold rounded-full">{{ t('guidesWheelsetComponentsRims.technology.hooked.badge') }}</span>
        </div>
        <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center border-b-2 border-emerald-500">
          <h5 class="text-emerald-600 font-bold mb-3">{{ t('guidesWheelsetComponentsRims.technology.hookless.title') }}</h5>
          <p class="text-xs tz-text-secondary leading-relaxed mb-4 min-h-[3rem]">
            {{ t('guidesWheelsetComponentsRims.technology.hookless.body') }}
            <br>
            <span class="text-rose-400/80 mt-1 block">*{{ t('guidesWheelsetComponentsRims.technology.hookless.note') }}</span>
          </p>
          <span class="inline-block px-3 py-1 bg-emerald-50 text-emerald-600 tz-micro-label uppercase font-bold rounded-full">{{ t('guidesWheelsetComponentsRims.technology.hookless.badge') }}</span>
        </div>
      </div>
    </div>

    <div v-else>
      <div class="space-y-10">
        <div>
          <div class="flex items-center justify-center gap-2 mb-6">
            <div class="h-px w-8 tz-surface-panel"></div>
            <h4 class="tz-text-primary font-bold text-base uppercase tracking-wider">
              {{ t('guidesWheelsetComponentsRims.choose.bedTitle') }}
            </h4>
            <div class="h-px w-8 tz-surface-panel"></div>
          </div>
          <div class="overflow-x-auto rounded-lg shadow-md border tz-border-subtle">
            <table class="w-full text-sm text-left tz-text-secondary">
              <thead class="text-xs tz-text-primary uppercase tz-surface-panel">
                <tr>
                  <th v-for="key in bedHeaderKeys" :key="key" scope="col" class="px-6 py-4 font-bold">{{ t(key) }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200">
                <tr v-for="(row, index) in bedRows" :key="row.featureKey" class="bg-[var(--tz-card-surface)] hover:tz-surface-panel">
                  <td class="px-6 py-4 font-medium tz-text-primary">{{ t(row.featureKey) }}</td>
                  <td v-if="index < 4" class="px-6 py-4">{{ t(row.undrilledKey) }}</td>
                  <td v-if="index < 4" class="px-6 py-4">{{ t(row.drilledKey) }}</td>
                </tr>
                <tr class="bg-[var(--tz-card-surface)] hover:tz-surface-panel">
                  <td class="px-6 py-4 font-medium text-rose-400">{{ t('guidesWheelsetComponentsRims.choose.cautionLabel') }}</td>
                  <td class="px-6 py-4 text-xs italic" colspan="2">
                    {{ t('guidesWheelsetComponentsRims.choose.cautionBody') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="grid md:grid-cols-2 gap-6">
          <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5">
            <h5 class="flex items-center gap-2 tz-text-primary font-bold mb-3">
              <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
              {{ t('guidesWheelsetComponentsRims.choose.symmetryTitle') }}
            </h5>
            <div class="space-y-4">
              <div v-for="item in symmetryItems" :key="item.titleKey" class="p-3 rounded-lg" :class="item.surfaceClass">
                <strong class="block text-sm mb-1" :class="item.titleClass">{{ t(item.titleKey) }}</strong>
                <p class="text-xs tz-text-secondary">{{ t(item.bodyKey) }}</p>
              </div>
            </div>
          </div>

          <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5">
            <h5 class="flex items-center gap-2 tz-text-primary font-bold mb-3">
              <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
              {{ t('guidesWheelsetComponentsRims.choose.drainTitle') }}
            </h5>
            <div class="space-y-3">
              <div v-for="item in drainItems" :key="item.titleKey" class="flex items-start gap-3">
                <span class="mt-1 w-1.5 h-1.5 rounded-full shrink-0" :class="item.dotClass"></span>
                <div>
                  <strong class="tz-text-primary text-sm block">{{ t(item.titleKey) }}</strong>
                  <p class="text-xs tz-text-secondary">{{ t(item.bodyKey) }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div>
          <div class="flex items-center justify-center gap-2 mb-6">
            <div class="h-px w-8 tz-surface-panel"></div>
            <h4 class="tz-text-primary font-bold text-base uppercase tracking-wider">
              {{ t('guidesWheelsetComponentsRims.choose.valveTitle') }}
            </h4>
            <div class="h-px w-8 tz-surface-panel"></div>
          </div>
          <div class="grid md:grid-cols-2 gap-6">
            <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 border-l-4 border-emerald-500">
              <div class="flex justify-between items-start mb-3">
                <h5 class="text-lg font-bold tz-text-primary">{{ t('guidesWheelsetComponentsRims.choose.presta.title') }}</h5>
                <span class="bg-emerald-100 text-emerald-600 px-2 py-1 rounded tz-micro-label uppercase font-bold">{{ t('guidesWheelsetComponentsRims.choose.presta.badge') }}</span>
              </div>
              <ul class="space-y-2 text-sm tz-text-secondary mb-4">
                <li v-for="key in prestaItemKeys" :key="key">• {{ t(key) }}</li>
              </ul>
              <div class="bg-rose-900/20 p-3 rounded border border-rose-500/20 text-xs text-rose-300">
                <strong>{{ t('guidesWheelsetComponentsRims.choose.presta.noteLabel') }}:</strong>
                {{ t('guidesWheelsetComponentsRims.choose.presta.note') }}
              </div>
            </div>
            <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 opacity-75 grayscale hover:grayscale-0 transition-all duration-300 border-l-4 tz-border-subtle">
              <div class="flex justify-between items-start mb-3">
                <h5 class="text-lg font-bold tz-text-primary">{{ t('guidesWheelsetComponentsRims.choose.schrader.title') }}</h5>
                <span class="tz-surface-panel tz-text-muted px-2 py-1 rounded tz-micro-label uppercase font-bold">{{ t('guidesWheelsetComponentsRims.choose.schrader.badge') }}</span>
              </div>
              <ul class="space-y-2 text-sm tz-text-secondary">
                <li v-for="key in schraderItemKeys" :key="key">• {{ t(key) }}</li>
              </ul>
            </div>
          </div>
        </div>

        <div class="bg-amber-900/10 border border-amber-500/20 rounded-2xl p-6">
          <h5 class="text-amber-500 font-bold mb-4 flex items-center gap-2">
            <span aria-hidden="true">!</span>
            {{ t('guidesWheelsetComponentsRims.choose.risksTitle') }}
          </h5>
          <div class="grid md:grid-cols-2 gap-6 text-sm">
            <div>
              <strong class="block tz-text-primary mb-2">{{ t('guidesWheelsetComponentsRims.choose.drillingRisk.title') }}</strong>
              <p class="tz-text-secondary leading-relaxed mb-2">{{ t('guidesWheelsetComponentsRims.choose.drillingRisk.body') }}</p>
              <p class="text-emerald-600 text-xs">✓ {{ t('guidesWheelsetComponentsRims.choose.drillingRisk.confirmation') }}</p>
            </div>
            <div>
              <strong class="block tz-text-primary mb-2">{{ t('guidesWheelsetComponentsRims.choose.closedBed.title') }}</strong>
              <p class="tz-text-secondary leading-relaxed">{{ t('guidesWheelsetComponentsRims.choose.closedBed.body') }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesWheelsetComponentsRims')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const activeTab = ref<'technology' | 'how-to-choose'>('technology')

const metrics = [
  {
    valueKey: 'guidesWheelsetComponentsRims.technology.metrics.torayValue',
    labelKey: 'guidesWheelsetComponentsRims.technology.metrics.torayLabel',
  },
  {
    valueKey: 'guidesWheelsetComponentsRims.technology.metrics.uciValue',
    labelKey: 'guidesWheelsetComponentsRims.technology.metrics.uciLabel',
  },
  {
    valueKey: 'guidesWheelsetComponentsRims.technology.metrics.warrantyValue',
    labelKey: 'guidesWheelsetComponentsRims.technology.metrics.warrantyLabel',
  },
]

const profiles = [
  {
    titleKey: 'guidesWheelsetComponentsRims.technology.shallow.title',
    depthKey: 'guidesWheelsetComponentsRims.technology.shallow.depth',
    bodyKey: 'guidesWheelsetComponentsRims.technology.shallow.body',
    bestForKey: 'guidesWheelsetComponentsRims.technology.shallow.bestFor',
  },
  {
    titleKey: 'guidesWheelsetComponentsRims.technology.aero.title',
    depthKey: 'guidesWheelsetComponentsRims.technology.aero.depth',
    bodyKey: 'guidesWheelsetComponentsRims.technology.aero.body',
    bestForKey: 'guidesWheelsetComponentsRims.technology.aero.bestFor',
  },
]

const widthItems = [
  {
    titleKey: 'guidesWheelsetComponentsRims.technology.internalWidth.title',
    bodyKey: 'guidesWheelsetComponentsRims.technology.internalWidth.body',
  },
  {
    titleKey: 'guidesWheelsetComponentsRims.technology.externalWidth.title',
    bodyKey: 'guidesWheelsetComponentsRims.technology.externalWidth.body',
  },
]

const bedHeaderKeys = [
  'guidesWheelsetComponentsRims.choose.bedHeaders.feature',
  'guidesWheelsetComponentsRims.choose.bedHeaders.undrilled',
  'guidesWheelsetComponentsRims.choose.bedHeaders.drilled',
]

const bedRows = [
  {
    featureKey: 'guidesWheelsetComponentsRims.choose.bedRows.0.feature',
    undrilledKey: 'guidesWheelsetComponentsRims.choose.bedRows.0.undrilled',
    drilledKey: 'guidesWheelsetComponentsRims.choose.bedRows.0.drilled',
  },
  {
    featureKey: 'guidesWheelsetComponentsRims.choose.bedRows.1.feature',
    undrilledKey: 'guidesWheelsetComponentsRims.choose.bedRows.1.undrilled',
    drilledKey: 'guidesWheelsetComponentsRims.choose.bedRows.1.drilled',
  },
  {
    featureKey: 'guidesWheelsetComponentsRims.choose.bedRows.2.feature',
    undrilledKey: 'guidesWheelsetComponentsRims.choose.bedRows.2.undrilled',
    drilledKey: 'guidesWheelsetComponentsRims.choose.bedRows.2.drilled',
  },
  {
    featureKey: 'guidesWheelsetComponentsRims.choose.bedRows.3.feature',
    undrilledKey: 'guidesWheelsetComponentsRims.choose.bedRows.3.undrilled',
    drilledKey: 'guidesWheelsetComponentsRims.choose.bedRows.3.drilled',
  },
]

const symmetryItems = [
  {
    titleKey: 'guidesWheelsetComponentsRims.choose.symmetrical.title',
    bodyKey: 'guidesWheelsetComponentsRims.choose.symmetrical.body',
    surfaceClass: 'bg-emerald-50',
    titleClass: 'text-emerald-600',
  },
  {
    titleKey: 'guidesWheelsetComponentsRims.choose.asymmetrical.title',
    bodyKey: 'guidesWheelsetComponentsRims.choose.asymmetrical.body',
    surfaceClass: 'tz-surface-panel border tz-border-subtle',
    titleClass: 'tz-text-primary',
  },
]

const drainItems = [
  {
    titleKey: 'guidesWheelsetComponentsRims.choose.withDrain.title',
    bodyKey: 'guidesWheelsetComponentsRims.choose.withDrain.body',
    dotClass: 'bg-emerald-500',
  },
  {
    titleKey: 'guidesWheelsetComponentsRims.choose.withoutDrain.title',
    bodyKey: 'guidesWheelsetComponentsRims.choose.withoutDrain.body',
    dotClass: 'bg-amber-500',
  },
]

const prestaItemKeys = [
  'guidesWheelsetComponentsRims.choose.presta.items.0',
  'guidesWheelsetComponentsRims.choose.presta.items.1',
  'guidesWheelsetComponentsRims.choose.presta.items.2',
]

const schraderItemKeys = [
  'guidesWheelsetComponentsRims.choose.schrader.items.0',
  'guidesWheelsetComponentsRims.choose.schrader.items.1',
  'guidesWheelsetComponentsRims.choose.schrader.items.2',
]
</script>
