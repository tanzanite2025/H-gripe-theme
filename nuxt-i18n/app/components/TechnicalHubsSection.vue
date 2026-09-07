<template>
  <div class="space-y-8">
    <div class="nav-pill-tabs mb-6" role="tablist">
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'technology' }"
        @click="activeTab = 'technology'"
      >
        {{ t('guidesWheelsetComponentsHubs.tabs.technology') }}
      </button>
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'choose-hubs' }"
        @click="activeTab = 'choose-hubs'"
      >
        {{ t('guidesWheelsetComponentsHubs.tabs.choose') }}
      </button>
    </div>

    <div v-if="activeTab === 'technology'" class="space-y-8">
      <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center border-t-4 border-slate-500">
        <h3 class="text-lg font-bold tz-text-secondary mb-2">
          {{ t('guidesWheelsetComponentsHubs.technology.title') }}
        </h3>
        <p class="text-sm tz-text-secondary leading-relaxed max-w-2xl mx-auto">
          {{ t('guidesWheelsetComponentsHubs.technology.intro') }}
        </p>
      </div>

      <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6">
        <div class="flex items-center gap-3 mb-6 pb-3 border-b border-emerald-200">
          <h3 class="text-lg font-bold tz-text-primary">
            {{ t('guidesWheelsetComponentsHubs.technology.spokeInterface.title') }}
          </h3>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div
            v-for="type in spokeTypes"
            :key="type.id"
            class="bg-emerald-50 rounded-xl p-4 shadow-md"
          >
            <h4 class="text-emerald-600 font-bold mb-3 flex items-center gap-2">
              {{ t(type.titleKey) }}
            </h4>
            <ul class="space-y-2 text-sm tz-text-secondary mb-4">
              <li v-for="itemKey in type.itemKeys" :key="itemKey">• {{ t(itemKey) }}</li>
            </ul>
            <div class="grid grid-cols-2 gap-3">
              <GuideImage
                :src="type.images[0].src"
                :alt="t(type.images[0].altKey)"
                :zoomOnClick="true"
                :caption="t(type.images[0].captionKey)"
                class="rounded-lg"
              />
              <GuideImage
                :src="type.images[1].src"
                :alt="t(type.images[1].altKey)"
                :zoomOnClick="true"
                :caption="t(type.images[1].captionKey)"
                class="rounded-lg"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6">
          <h3 class="text-lg font-bold text-amber-500 mb-4">{{ t('guidesWheelsetComponentsHubs.technology.axle.title') }}</h3>
          <div class="space-y-4 text-sm tz-text-secondary">
            <div class="tz-surface-panel p-3 rounded-lg">
              <strong class="block tz-text-primary mb-1">{{ t('guidesWheelsetComponentsHubs.technology.axle.frontTitle') }}</strong>
              <p>{{ t('guidesWheelsetComponentsHubs.technology.axle.frontBody') }}</p>
            </div>
            <div class="tz-surface-panel p-3 rounded-lg">
              <strong class="block tz-text-primary mb-1">{{ t('guidesWheelsetComponentsHubs.technology.axle.rearTitle') }}</strong>
              <p>{{ t('guidesWheelsetComponentsHubs.technology.axle.rearBody') }}</p>
            </div>
            <p class="text-xs tz-text-muted mt-2 italic">
              {{ t('guidesWheelsetComponentsHubs.technology.axle.note') }}
            </p>
          </div>
        </div>

        <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6">
          <h3 class="text-lg font-bold text-rose-500 mb-4">{{ t('guidesWheelsetComponentsHubs.technology.brake.title') }}</h3>
          <ul class="space-y-3 text-sm tz-text-secondary">
            <li v-for="item in brakeInterfaces" :key="item.titleKey" class="flex items-start gap-3">
              <span class="mt-1 w-1.5 h-1.5 rounded-full bg-rose-500 shrink-0"></span>
              <div>
                <strong class="tz-text-primary block">{{ t(item.titleKey) }}</strong>
                {{ t(item.bodyKey) }}
              </div>
            </li>
          </ul>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6">
          <h3 class="text-lg font-bold text-emerald-600 mb-4">{{ t('guidesWheelsetComponentsHubs.technology.holeCount.title') }}</h3>
          <div class="space-y-4">
            <div
              v-for="row in holeCountRows"
              :key="row.labelKey"
              class="flex items-center justify-between text-sm border-b tz-border-strong/5 pb-2"
            >
              <span class="tz-text-muted">{{ t(row.labelKey) }}</span>
              <span class="tz-text-primary font-bold">{{ t(row.valueKey) }}</span>
            </div>
          </div>
        </div>

        <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6">
          <h3 class="text-lg font-bold text-emerald-600 mb-4">{{ t('guidesWheelsetComponentsHubs.technology.ratchet.title') }}</h3>
          <p class="text-sm tz-text-secondary mb-4">{{ t('guidesWheelsetComponentsHubs.technology.ratchet.body') }}</p>
          <div class="grid grid-cols-3 gap-2 text-center">
            <div v-for="level in ratchetLevels" :key="level.value" class="bg-emerald-50 rounded-lg p-2 shadow-md">
              <div class="text-lg font-bold text-emerald-600">{{ level.value }}</div>
              <div class="tz-caption tz-text-muted">{{ t(level.labelKey) }}</div>
            </div>
          </div>
        </div>
      </div>

      <div class="rounded-2xl tz-surface-card border tz-border-subtle shadow-md p-6 flex flex-col md:flex-row items-center justify-between gap-6">
        <div class="text-center md:text-left">
          <h3 class="text-lg font-bold tz-text-primary mb-1">{{ t('guidesWheelsetComponentsHubs.technology.freehub.title') }}</h3>
          <p class="text-sm tz-text-secondary">{{ t('guidesWheelsetComponentsHubs.technology.freehub.body') }}</p>
        </div>
        <button
          type="button"
          class="shrink-0 inline-flex items-center justify-center rounded-full bg-[var(--tz-action-primary)] px-6 py-2.5 text-sm font-bold text-white shadow-lg hover:bg-[var(--tz-action-primary-hover)] hover:shadow-slate-900/20 transition-all"
          @click="goToChooseFreehub"
        >
          {{ t('guidesWheelsetComponentsHubs.technology.freehub.button') }}
        </button>
      </div>
    </div>

    <div v-else class="space-y-8">
      <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center border-t-4 border-emerald-500">
        <h3 class="text-xl font-bold tz-text-primary mb-2">{{ t('guidesWheelsetComponentsHubs.choose.title') }}</h3>
        <p class="text-sm tz-text-secondary max-w-2xl mx-auto">{{ t('guidesWheelsetComponentsHubs.choose.intro') }}</p>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div class="space-y-6">
          <div v-for="hub in hubChoices" :key="hub.id" class="bg-emerald-50 rounded-xl p-5 shadow-md border border-emerald-200">
            <strong class="block text-emerald-600 text-lg mb-3">{{ t(hub.titleKey) }}</strong>
            <p class="text-sm tz-text-secondary mb-2"><strong>{{ t(hub.structureLabelKey) }}:</strong> {{ t(hub.structureKey) }}</p>
            <p class="text-sm tz-text-secondary mb-2"><strong>{{ t(hub.advantagesLabelKey) }}:</strong> {{ t(hub.advantagesKey) }}</p>
            <p class="text-sm tz-text-secondary mb-4"><strong>{{ t(hub.bestForLabelKey) }}:</strong> {{ t(hub.bestForKey) }}</p>
            <div class="bg-emerald-50 px-3 py-1.5 rounded text-xs text-emerald-700 inline-block font-medium">{{ t(hub.compatibleKey) }}</div>
          </div>
        </div>
        <div class="h-full">
          <GuideImage
            src="/public/wheelsetbuyersguide/wheelcomponents/hubs/bicycle-hub-spoke-type-comparison-jbend-straightpull.webp"
            :alt="t('guidesWheelsetComponentsHubs.choose.image.alt')"
            :zoomOnClick="true"
            :caption="t('guidesWheelsetComponentsHubs.choose.image.caption')"
            class="rounded-xl overflow-hidden shadow-md h-full object-cover"
          />
        </div>
      </div>

      <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-6">
        <h3 class="text-lg font-bold text-emerald-600 mb-4">{{ t('guidesWheelsetComponentsHubs.choose.holeCount.title') }}</h3>
        <div class="grid md:grid-cols-2 gap-6">
          <div v-for="choice in holeCountChoices" :key="choice.titleKey" class="tz-surface-panel p-4 rounded-xl border tz-border-subtle">
            <strong class="block tz-text-primary mb-2">{{ t(choice.titleKey) }}</strong>
            <p class="text-sm tz-text-secondary leading-relaxed">{{ t(choice.bodyKey) }}</p>
          </div>
        </div>
        <p class="mt-4 text-xs text-emerald-600/80 bg-emerald-50 p-3 rounded-lg border border-emerald-200">
          <strong>{{ t('guidesWheelsetComponentsHubs.choose.holeCount.principleLabel') }}:</strong>
          {{ t('guidesWheelsetComponentsHubs.choose.holeCount.principle') }}
        </p>
      </div>

      <div class="bg-amber-500/10 border-l-4 border-amber-500 p-4 rounded-r-xl">
        <h4 class="text-amber-500 font-bold mb-2 flex items-center gap-2">
          <span aria-hidden="true">!</span>
          {{ t('guidesWheelsetComponentsHubs.choose.notes.title') }}
        </h4>
        <ul class="space-y-1 text-sm text-amber-500/80 list-disc pl-5 marker:text-amber-500/50">
          <li v-for="itemKey in noteItemKeys" :key="itemKey">{{ t(itemKey) }}</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n, useLocalePath, useRouter } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesWheelsetComponentsHubs')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const activeTab = ref<'technology' | 'choose-hubs'>('technology')
const router = useRouter()
const localePath = useLocalePath()

interface HubImage {
  src: string
  altKey: string
  captionKey: string
}

interface HubSpokeType {
  id: string
  titleKey: string
  itemKeys: string[]
  images: [HubImage, HubImage]
}

const spokeTypes: HubSpokeType[] = [
  {
    id: 'j-bend',
    titleKey: 'guidesWheelsetComponentsHubs.technology.spokeInterface.jBend.title',
    itemKeys: [
      'guidesWheelsetComponentsHubs.technology.spokeInterface.jBend.items.0',
      'guidesWheelsetComponentsHubs.technology.spokeInterface.jBend.items.1',
      'guidesWheelsetComponentsHubs.technology.spokeInterface.jBend.items.2',
    ],
    images: [
      {
        src: '/public/technical/hubs/hubs/dt240-6bolt-jbend-hub.webp',
        altKey: 'guidesWheelsetComponentsHubs.technology.spokeInterface.jBend.images.sixBoltAlt',
        captionKey: 'guidesWheelsetComponentsHubs.technology.spokeInterface.jBend.images.sixBoltCaption',
      },
      {
        src: '/public/technical/hubs/hubs/dt240-centerlock-jbend-hub.webp',
        altKey: 'guidesWheelsetComponentsHubs.technology.spokeInterface.jBend.images.centerLockAlt',
        captionKey: 'guidesWheelsetComponentsHubs.technology.spokeInterface.jBend.images.centerLockCaption',
      },
    ],
  },
  {
    id: 'straight-pull',
    titleKey: 'guidesWheelsetComponentsHubs.technology.spokeInterface.straightPull.title',
    itemKeys: [
      'guidesWheelsetComponentsHubs.technology.spokeInterface.straightPull.items.0',
      'guidesWheelsetComponentsHubs.technology.spokeInterface.straightPull.items.1',
      'guidesWheelsetComponentsHubs.technology.spokeInterface.straightPull.items.2',
    ],
    images: [
      {
        src: '/public/technical/hubs/hubs/dt240-6bolt-straightpull-hub.webp',
        altKey: 'guidesWheelsetComponentsHubs.technology.spokeInterface.straightPull.images.sixBoltAlt',
        captionKey: 'guidesWheelsetComponentsHubs.technology.spokeInterface.straightPull.images.sixBoltCaption',
      },
      {
        src: '/public/technical/hubs/hubs/dt240-centerlock-straightpull-hub.webp',
        altKey: 'guidesWheelsetComponentsHubs.technology.spokeInterface.straightPull.images.centerLockAlt',
        captionKey: 'guidesWheelsetComponentsHubs.technology.spokeInterface.straightPull.images.centerLockCaption',
      },
    ],
  },
]

const brakeInterfaces = [
  {
    titleKey: 'guidesWheelsetComponentsHubs.technology.brake.sixBoltTitle',
    bodyKey: 'guidesWheelsetComponentsHubs.technology.brake.sixBoltBody',
  },
  {
    titleKey: 'guidesWheelsetComponentsHubs.technology.brake.centerLockTitle',
    bodyKey: 'guidesWheelsetComponentsHubs.technology.brake.centerLockBody',
  },
]

const holeCountRows = [
  {
    labelKey: 'guidesWheelsetComponentsHubs.technology.holeCount.rows.0.label',
    valueKey: 'guidesWheelsetComponentsHubs.technology.holeCount.rows.0.value',
  },
  {
    labelKey: 'guidesWheelsetComponentsHubs.technology.holeCount.rows.1.label',
    valueKey: 'guidesWheelsetComponentsHubs.technology.holeCount.rows.1.value',
  },
  {
    labelKey: 'guidesWheelsetComponentsHubs.technology.holeCount.rows.2.label',
    valueKey: 'guidesWheelsetComponentsHubs.technology.holeCount.rows.2.value',
  },
]

const ratchetLevels = [
  {
    value: '18T',
    labelKey: 'guidesWheelsetComponentsHubs.technology.ratchet.levels.standard',
  },
  {
    value: '36T',
    labelKey: 'guidesWheelsetComponentsHubs.technology.ratchet.levels.balanced',
  },
  {
    value: '54T',
    labelKey: 'guidesWheelsetComponentsHubs.technology.ratchet.levels.instant',
  },
]

const hubChoices = [
  {
    id: 'j-bend',
    titleKey: 'guidesWheelsetComponentsHubs.choose.jBend.title',
    structureLabelKey: 'guidesWheelsetComponentsHubs.choose.jBend.structureLabel',
    structureKey: 'guidesWheelsetComponentsHubs.choose.jBend.structure',
    advantagesLabelKey: 'guidesWheelsetComponentsHubs.choose.jBend.advantagesLabel',
    advantagesKey: 'guidesWheelsetComponentsHubs.choose.jBend.advantages',
    bestForLabelKey: 'guidesWheelsetComponentsHubs.choose.jBend.bestForLabel',
    bestForKey: 'guidesWheelsetComponentsHubs.choose.jBend.bestFor',
    compatibleKey: 'guidesWheelsetComponentsHubs.choose.jBend.compatible',
  },
  {
    id: 'straight-pull',
    titleKey: 'guidesWheelsetComponentsHubs.choose.straightPull.title',
    structureLabelKey: 'guidesWheelsetComponentsHubs.choose.straightPull.structureLabel',
    structureKey: 'guidesWheelsetComponentsHubs.choose.straightPull.structure',
    advantagesLabelKey: 'guidesWheelsetComponentsHubs.choose.straightPull.advantagesLabel',
    advantagesKey: 'guidesWheelsetComponentsHubs.choose.straightPull.advantages',
    bestForLabelKey: 'guidesWheelsetComponentsHubs.choose.straightPull.bestForLabel',
    bestForKey: 'guidesWheelsetComponentsHubs.choose.straightPull.bestFor',
    compatibleKey: 'guidesWheelsetComponentsHubs.choose.straightPull.compatible',
  },
]

const holeCountChoices = [
  {
    titleKey: 'guidesWheelsetComponentsHubs.choose.holeCount.lightweightTitle',
    bodyKey: 'guidesWheelsetComponentsHubs.choose.holeCount.lightweightBody',
  },
  {
    titleKey: 'guidesWheelsetComponentsHubs.choose.holeCount.durabilityTitle',
    bodyKey: 'guidesWheelsetComponentsHubs.choose.holeCount.durabilityBody',
  },
]

const noteItemKeys = [
  'guidesWheelsetComponentsHubs.choose.notes.items.0',
  'guidesWheelsetComponentsHubs.choose.notes.items.1',
  'guidesWheelsetComponentsHubs.choose.notes.items.2',
]

const goToChooseFreehub = async () => {
  await router.push(localePath('/guides/wheelset-buyers/choose-freehub'))
}
</script>
