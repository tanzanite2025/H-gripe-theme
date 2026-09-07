<template>
  <section id="factory" class="company-section">
    <!-- Intro Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-12 lg:mb-16">
      <div v-for="item in factoryIntro" :key="item.title" class="premium-card p-6 flex flex-col gap-3">

        <div>
          <h3 class="text-lg font-bold tz-text-primary mb-2">{{ item.title }}</h3>
          <p class="text-sm tz-text-secondary leading-relaxed">{{ item.desc }}</p>
        </div>
      </div>
    </div>

    <!-- Process Section Header -->
    <div class="text-center mb-8 lg:mb-12 relative">
        <div class="inline-flex items-center justify-center w-12 h-1 mb-4 rounded-full bg-emerald-500"></div>
        <h2 class="text-2xl sm:text-3xl font-bold tz-text-primary mb-2">
          {{ t('companyAboutFactory.process.title') }}
        </h2>
        <p class="tz-text-secondary max-w-2xl mx-auto">
          {{ t('companyAboutFactory.process.description') }}
        </p>
    </div>

    <!-- Process Steps Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
      <article 
        v-for="step in factorySteps" 
        :key="step.step" 
        class="premium-card overflow-hidden group flex flex-col h-full"
      >
          <!-- Image -->
          <div class="relative aspect-[4/3] overflow-hidden tz-surface-panel shrink-0">
            <img 
              :src="step.img" 
              :alt="step.alt"
              class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105" 
              loading="lazy"
              decoding="async"
            >
            <div class="absolute inset-0 bg-black/20"></div>
            <!-- Step Badge -->
            <div class="absolute top-3 left-3 tz-surface-panel backdrop-blur border tz-border-subtle tz-text-primary text-xs font-mono font-bold px-2 py-1 rounded">
              {{ t('companyAboutFactory.process.stepLabel') }} {{ step.step }}
            </div>
          </div>
          
          <!-- Content -->
          <div class="p-5 flex flex-col flex-1">
            <h3 class="text-base font-bold tz-text-primary mb-2 group-hover:text-emerald-700 transition-colors">
              {{ step.title }}
            </h3>
            <p class="text-sm tz-text-secondary leading-relaxed">
              {{ step.description }}
            </p>
          </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('companyAboutFactory')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const factoryIntroKeys = ['history', 'scale', 'innovation', 'leader'] as const
const factoryIntro = computed(() => factoryIntroKeys.map((key) => ({
  title: t(`companyAboutFactory.intro.${key}.title`),
  desc: t(`companyAboutFactory.intro.${key}.description`),
})))

const factoryStepSources = [
  { step: '01', key: 'step01', img: '/company/ourstory/factory/factory-EpoxyResinWorkshop1.webp' },
  { step: '02', key: 'step02', img: '/company/ourstory/factory/factory-carbonprepregsworkshop2.webp' },
  { step: '03', key: 'step03', img: '/company/ourstory/factory/factory-carbonprepregsstoreroom3.webp' },
  { step: '04', key: 'step04', img: '/company/ourstory/factory/factory-cuttingworkshop4.webp' },
  { step: '05', key: 'step05', img: '/company/ourstory/factory/factory-laminationpreparationworkshop5.webp' },
  { step: '06', key: 'step06', img: '/company/ourstory/factory/factory-premoldlayupworkshop6.webp' },
  { step: '07', key: 'step07', img: '/company/ourstory/factory/factory-moldingworkshop7.webp' },
  { step: '08', key: 'step08', img: '/company/ourstory/factory/factory-Appearance%20inspection8.webp' },
  { step: '09', key: 'step09', img: '/company/ourstory/factory/factory-cncmachiningworkshop9.webp' },
  { step: '10', key: 'step10', img: '/company/ourstory/factory/factory-Weightdetection10.webp' },
  { step: '11', key: 'step11', img: '/company/ourstory/factory/factory-sandingworshop11.webp' },
  { step: '12', key: 'step12', img: '/company/ourstory/factory/factory-grinding12.webp' },
  { step: '13', key: 'step13', img: '/company/ourstory/factory/factory-detailstreatment13.webp' },
  { step: '14', key: 'step14', img: '/company/ourstory/factory/factory-otherdetection14.webp' },
  { step: '15', key: 'step15', img: '/company/ourstory/factory/factory-painting15.webp' },
  { step: '16', key: 'step16', img: '/company/ourstory/factory/factory-logolidedecals16.webp' },
  { step: '17', key: 'step17', img: '/company/ourstory/factory/factory-lasercarving17.webp' },
  { step: '18', key: 'step18', img: '/company/ourstory/factory/factory-inspectionpacking18.webp' },
  { step: '19', key: 'step19', img: '/company/ourstory/factory/factory-ourwarehouse19.webp' },
] as const

const factorySteps = computed(() => factoryStepSources.map((step) => ({
  ...step,
  title: t(`companyAboutFactory.steps.${step.key}.title`),
  description: t(`companyAboutFactory.steps.${step.key}.description`),
  alt: t(`companyAboutFactory.steps.${step.key}.alt`),
})))
</script>
