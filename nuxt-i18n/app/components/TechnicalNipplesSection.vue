<template>
  <div>
    <div class="nav-pill-tabs mb-6" role="tablist">
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'brass-vs-alloy' }"
        @click="activeTab = 'brass-vs-alloy'"
      >
        {{ t('guidesWheelsetComponentsNipples.tabs.comparison') }}
      </button>
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'choose-nipple' }"
        @click="activeTab = 'choose-nipple'"
      >
        {{ t('guidesWheelsetComponentsNipples.tabs.choose') }}
      </button>
    </div>

    <div v-if="activeTab === 'brass-vs-alloy'" class="space-y-8">
      <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center border-t-4 border-amber-500">
        <h3 class="text-xl font-bold tz-text-secondary mb-2">{{ t('guidesWheelsetComponentsNipples.comparison.title') }}</h3>
        <p class="text-sm tz-text-secondary max-w-2xl mx-auto">{{ t('guidesWheelsetComponentsNipples.comparison.intro') }}</p>
      </div>

      <div class="grid gap-6 md:grid-cols-2">
        <div v-for="section in comparisonSections" :key="section.titleKey" class="tz-surface-panel p-5 rounded-xl border tz-border-subtle">
          <div class="flex items-center gap-3 mb-4">
            <span class="flex h-8 w-8 items-center justify-center rounded-full tz-surface-panel tz-text-secondary font-bold text-sm">{{ section.number }}</span>
            <h4 class="text-lg font-bold tz-text-primary">{{ t(section.titleKey) }}</h4>
          </div>

          <template v-if="section.id === 'weight'">
            <ul class="space-y-3 text-sm tz-text-secondary">
              <li class="flex justify-between border-b tz-border-subtle pb-2">
                <span>{{ t('guidesWheelsetComponentsNipples.comparison.weight.brass') }}</span>
                <span class="font-mono tz-text-primary">{{ t('guidesWheelsetComponentsNipples.comparison.weight.brassValue') }}</span>
              </li>
              <li class="flex justify-between border-b tz-border-subtle pb-2">
                <span>{{ t('guidesWheelsetComponentsNipples.comparison.weight.alloy') }}</span>
                <span class="font-mono text-emerald-600">{{ t('guidesWheelsetComponentsNipples.comparison.weight.alloyValue') }}</span>
              </li>
              <li class="text-xs tz-text-muted italic pt-1">{{ t('guidesWheelsetComponentsNipples.comparison.weight.impact') }}</li>
            </ul>
          </template>
          <template v-else-if="section.id === 'color'">
            <ul class="space-y-3 text-sm tz-text-secondary">
              <li class="flex items-start gap-2">
                <strong class="text-amber-500 shrink-0">{{ t('guidesWheelsetComponentsNipples.comparison.color.brass') }}:</strong>
                <span>{{ t('guidesWheelsetComponentsNipples.comparison.color.brassBody') }}</span>
              </li>
              <li class="flex items-start gap-2">
                <strong class="text-emerald-600 shrink-0">{{ t('guidesWheelsetComponentsNipples.comparison.color.alloy') }}:</strong>
                <span>{{ t('guidesWheelsetComponentsNipples.comparison.color.alloyBody') }}</span>
              </li>
            </ul>
          </template>
          <template v-else-if="section.id === 'strength'">
            <p class="text-sm tz-text-secondary mb-3"><strong>{{ t('guidesWheelsetComponentsNipples.comparison.strength.standardLabel') }}:</strong> {{ t('guidesWheelsetComponentsNipples.comparison.strength.standardBody') }}</p>
            <p class="text-sm tz-text-secondary"><strong>{{ t('guidesWheelsetComponentsNipples.comparison.strength.extendedLabel') }}:</strong> {{ t('guidesWheelsetComponentsNipples.comparison.strength.extendedBody') }}</p>
          </template>
          <template v-else-if="section.id === 'assembly'">
            <ul class="space-y-2 text-sm tz-text-secondary list-disc list-inside marker:tz-text-muted">
              <li v-for="key in assemblyItemKeys" :key="key">{{ t(key) }}</li>
            </ul>
          </template>
          <template v-else>
            <div class="grid md:grid-cols-2 gap-4">
              <div class="bg-amber-500/5 p-3 rounded-lg">
                <strong class="block text-amber-500 mb-1 text-sm">{{ t('guidesWheelsetComponentsNipples.comparison.corrosion.brass') }}</strong>
                <p class="text-xs tz-text-secondary">{{ t('guidesWheelsetComponentsNipples.comparison.corrosion.brassBody') }}</p>
              </div>
              <div class="bg-emerald-50 p-3 rounded-lg">
                <strong class="block text-emerald-600 mb-1 text-sm">{{ t('guidesWheelsetComponentsNipples.comparison.corrosion.alloy') }}</strong>
                <p class="text-xs tz-text-secondary">{{ t('guidesWheelsetComponentsNipples.comparison.corrosion.alloyBody') }}</p>
              </div>
            </div>
          </template>
        </div>
      </div>

      <div class="rounded-2xl bg-emerald-50 border border-emerald-200 p-6">
        <h4 class="text-emerald-600 font-bold text-lg mb-4 flex items-center gap-2">
          {{ t('guidesWheelsetComponentsNipples.comparison.summary.title') }}
        </h4>
        <div class="space-y-3">
          <div v-for="item in summaryItems" :key="item.labelKey" class="flex items-center gap-3 tz-surface-panel p-3 rounded-lg">
            <span class="tz-text-secondary text-sm">{{ t(item.labelKey) }}</span>
            <span class="flex-1 border-b border-dashed tz-border-strong mx-2"></span>
            <strong :class="item.choiceClass" class="text-sm">{{ t(item.choiceKey) }}</strong>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="space-y-10">
      <section>
        <div class="mb-6">
          <h3 class="text-xl font-bold tz-text-secondary mb-3">{{ t('guidesWheelsetComponentsNipples.choose.depth.title') }}</h3>
          <p class="text-sm tz-text-secondary leading-relaxed max-w-3xl">{{ t('guidesWheelsetComponentsNipples.choose.depth.body') }}</p>
        </div>
        <div class="grid md:grid-cols-2 gap-6">
          <div class="space-y-4">
            <div class="bg-emerald-50 p-4 rounded-xl border border-emerald-200">
              <strong class="block text-emerald-600 text-sm mb-2">{{ t('guidesWheelsetComponentsNipples.choose.depth.challengeTitle') }}</strong>
              <p class="text-xs tz-text-secondary leading-relaxed">{{ t('guidesWheelsetComponentsNipples.choose.depth.challengeBody') }}</p>
            </div>
            <div class="tz-surface-panel p-4 rounded-xl">
              <strong class="block tz-text-primary text-sm mb-2">{{ t('guidesWheelsetComponentsNipples.choose.depth.matchedTitle') }}</strong>
              <p class="text-xs tz-text-secondary leading-relaxed">{{ t('guidesWheelsetComponentsNipples.choose.depth.matchedBody') }}</p>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <GuideImage
              src="/public/wheelsetbuyersguide/wheelcomponents/nipple/choosenipple/sapim-doublesquare-nipple.webp"
              :alt="t('guidesWheelsetComponentsNipples.choose.depth.doubleSquareAlt')"
              :zoomOnClick="true"
              :caption="t('guidesWheelsetComponentsNipples.choose.depth.doubleSquareCaption')"
              class="rounded-lg shadow-lg"
            />
            <GuideImage
              src="/public/wheelsetbuyersguide/wheelcomponents/nipple/choosenipple/sapim-upsidedown-nipple.webp"
              :alt="t('guidesWheelsetComponentsNipples.choose.depth.upsideDownAlt')"
              :zoomOnClick="true"
              :caption="t('guidesWheelsetComponentsNipples.choose.depth.upsideDownCaption')"
              class="rounded-lg shadow-lg"
            />
          </div>
        </div>
      </section>

      <section>
        <h3 class="text-lg font-bold tz-text-secondary mb-6 flex items-center gap-2">
          <span class="w-1 h-6 bg-emerald-500 rounded-full"></span>
          {{ t('guidesWheelsetComponentsNipples.choose.advanced.title') }}
        </h3>
        <div class="grid md:grid-cols-3 gap-6">
          <div v-for="option in advancedOptions" :key="option.id" class="bg-[var(--tz-card-surface)] rounded-xl overflow-hidden shadow-md flex flex-col">
            <div v-if="option.image" class="h-32 overflow-hidden tz-surface-subtle">
              <GuideImage
                :src="option.image.src"
                :alt="t(option.image.altKey)"
                :zoomOnClick="true"
                class="w-full h-full object-cover"
              />
            </div>
            <div class="p-4 flex-1">
              <strong class="block mb-2" :class="option.titleClass">{{ t(option.titleKey) }}</strong>
              <p class="text-xs tz-text-secondary mb-3">{{ t(option.bodyKey) }}</p>
              <ul v-if="option.itemKeys" class="text-xs tz-text-muted space-y-1 list-disc list-inside">
                <li v-for="key in option.itemKeys" :key="key">{{ t(key) }}</li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section class="tz-surface-panel rounded-2xl p-6 border tz-border-subtle">
        <div class="flex flex-col md:flex-row gap-8 items-center">
          <div class="flex-1">
            <h3 class="text-lg font-bold tz-text-secondary mb-3">{{ t('guidesWheelsetComponentsNipples.choose.washers.title') }}</h3>
            <p class="text-sm tz-text-secondary leading-relaxed mb-4">{{ t('guidesWheelsetComponentsNipples.choose.washers.body') }}</p>
            <div class="flex flex-wrap gap-2">
              <span v-for="key in washerBenefits" :key="key" class="px-3 py-1 rounded-full tz-surface-subtle text-xs tz-text-secondary">{{ t(key) }}</span>
            </div>
          </div>
          <div class="flex gap-4 shrink-0">
            <GuideImage
              src="/public/wheelsetbuyersguide/wheelcomponents/nipple/choosenipple/sapim-hm-washer.webp"
              :alt="t('guidesWheelsetComponentsNipples.choose.washers.hmAlt')"
              :zoomOnClick="true"
              :caption="t('guidesWheelsetComponentsNipples.choose.washers.hmCaption')"
              class="w-24 h-24 rounded-lg object-cover shadow-lg"
            />
            <GuideImage
              src="/public/wheelsetbuyersguide/wheelcomponents/nipple/choosenipple/sapim-nipple-washer.webp"
              :alt="t('guidesWheelsetComponentsNipples.choose.washers.nippleAlt')"
              :zoomOnClick="true"
              :caption="t('guidesWheelsetComponentsNipples.choose.washers.nippleCaption')"
              class="w-24 h-24 rounded-lg object-cover shadow-lg"
            />
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesWheelsetComponentsNipples')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const activeTab = ref<'brass-vs-alloy' | 'choose-nipple'>('brass-vs-alloy')

const comparisonSections = [
  {
    id: 'weight',
    number: 1,
    titleKey: 'guidesWheelsetComponentsNipples.comparison.weight.title',
  },
  {
    id: 'color',
    number: 2,
    titleKey: 'guidesWheelsetComponentsNipples.comparison.color.title',
  },
  {
    id: 'strength',
    number: 3,
    titleKey: 'guidesWheelsetComponentsNipples.comparison.strength.title',
  },
  {
    id: 'assembly',
    number: 4,
    titleKey: 'guidesWheelsetComponentsNipples.comparison.assembly.title',
  },
  {
    id: 'corrosion',
    number: 5,
    titleKey: 'guidesWheelsetComponentsNipples.comparison.corrosion.title',
  },
]

const assemblyItemKeys = [
  'guidesWheelsetComponentsNipples.comparison.assembly.items.0',
  'guidesWheelsetComponentsNipples.comparison.assembly.items.1',
  'guidesWheelsetComponentsNipples.comparison.assembly.items.2',
  'guidesWheelsetComponentsNipples.comparison.assembly.items.3',
]

const summaryItems = [
  {
    labelKey: 'guidesWheelsetComponentsNipples.comparison.summary.lightweight',
    choiceKey: 'guidesWheelsetComponentsNipples.comparison.summary.lightweightChoice',
    choiceClass: 'text-emerald-600',
  },
  {
    labelKey: 'guidesWheelsetComponentsNipples.comparison.summary.durability',
    choiceKey: 'guidesWheelsetComponentsNipples.comparison.summary.durabilityChoice',
    choiceClass: 'text-amber-500',
  },
  {
    labelKey: 'guidesWheelsetComponentsNipples.comparison.summary.carbon',
    choiceKey: 'guidesWheelsetComponentsNipples.comparison.summary.carbonChoice',
    choiceClass: 'tz-text-primary',
  },
]

const advancedOptions = [
  {
    id: 'double-square',
    titleKey: 'guidesWheelsetComponentsNipples.choose.advanced.doubleSquare.title',
    bodyKey: 'guidesWheelsetComponentsNipples.choose.advanced.doubleSquare.body',
    itemKeys: [
      'guidesWheelsetComponentsNipples.choose.advanced.doubleSquare.items.0',
      'guidesWheelsetComponentsNipples.choose.advanced.doubleSquare.items.1',
      'guidesWheelsetComponentsNipples.choose.advanced.doubleSquare.items.2',
    ],
    titleClass: 'text-emerald-600',
  },
  {
    id: 'internal',
    titleKey: 'guidesWheelsetComponentsNipples.choose.advanced.internal.title',
    bodyKey: 'guidesWheelsetComponentsNipples.choose.advanced.internal.body',
    itemKeys: [
      'guidesWheelsetComponentsNipples.choose.advanced.internal.items.0',
      'guidesWheelsetComponentsNipples.choose.advanced.internal.items.1',
      'guidesWheelsetComponentsNipples.choose.advanced.internal.items.2',
    ],
    titleClass: 'tz-text-primary',
  },
  {
    id: 'secure-lock',
    titleKey: 'guidesWheelsetComponentsNipples.choose.advanced.secureLock.title',
    bodyKey: 'guidesWheelsetComponentsNipples.choose.advanced.secureLock.body',
    titleClass: 'text-amber-500',
    image: {
      src: '/public/wheelsetbuyersguide/wheelcomponents/nipple/choosenipple/sapim-securelock-nipple.webp',
      altKey: 'guidesWheelsetComponentsNipples.choose.advanced.secureLock.alt',
    },
  },
]

const washerBenefits = [
  'guidesWheelsetComponentsNipples.choose.washers.load',
  'guidesWheelsetComponentsNipples.choose.washers.angle',
  'guidesWheelsetComponentsNipples.choose.washers.protection',
]
</script>
