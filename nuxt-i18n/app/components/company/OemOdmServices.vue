<template>
  <div class="flex flex-col gap-8 md:gap-12">
    
    <!-- Hero Section -->
    <section class="relative text-center">
      <h2 class="text-2xl md:text-3xl font-bold tz-text-primary mb-4 tracking-tight">
        {{ t('companyOemOdm.hero.title') }}
      </h2>
      <p class="tz-text-secondary max-w-3xl mx-auto text-base md:text-lg leading-relaxed font-light">
        {{ t('companyOemOdm.hero.description') }}
      </p>
    </section>

    <!-- Core Services (OEM vs ODM) -->
    <section class="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-none w-full">
      <!-- OEM Card -->
      <div class="relative group tz-surface-panel backdrop-blur-sm border tz-border-subtle hover:border-amber-500/50 rounded-3xl p-4 md:p-6 transition-all duration-300 shadow-lg hover:shadow-amber-500/10 hover:-translate-y-1">
        <div class="absolute top-0 right-0 p-5 opacity-10 group-hover:opacity-20 transition-opacity">
          <svg class="w-20 h-20 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg>
        </div>
        <h3 class="text-xl font-bold tz-text-primary mb-3 group-hover:text-amber-400 transition-colors">{{ t('companyOemOdm.services.oem.title') }}</h3>
        <p class="tz-text-secondary leading-relaxed mb-4 text-sm">
          {{ t('companyOemOdm.services.oem.description') }}
        </p>
        <ul class="space-y-2">
          <li class="flex items-center gap-3 tz-text-secondary text-xs">
            <span class="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
            <span>{{ t('companyOemOdm.services.oem.features.consistency') }}</span>
          </li>
          <li class="flex items-center gap-3 tz-text-secondary text-xs">
            <span class="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
            <span>{{ t('companyOemOdm.services.oem.features.capacity') }}</span>
          </li>
          <li class="flex items-center gap-3 tz-text-secondary text-xs">
            <span class="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
            <span>{{ t('companyOemOdm.services.oem.features.confidentiality') }}</span>
          </li>
        </ul>
      </div>

      <!-- ODM Card -->
      <div class="relative group tz-surface-panel backdrop-blur-sm border tz-border-subtle hover:border-amber-500/50 rounded-3xl p-4 md:p-6 transition-all duration-300 shadow-lg hover:shadow-amber-500/10 hover:-translate-y-1">
        <div class="absolute top-0 right-0 p-5 opacity-10 group-hover:opacity-20 transition-opacity">
          <svg class="w-20 h-20 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" /></svg>
        </div>
        <h3 class="text-xl font-bold tz-text-primary mb-3 group-hover:text-amber-400 transition-colors">{{ t('companyOemOdm.services.odm.title') }}</h3>
        <p class="tz-text-secondary leading-relaxed mb-4 text-sm">
          {{ t('companyOemOdm.services.odm.description') }}
        </p>
        <ul class="space-y-2">
          <li class="flex items-center gap-3 tz-text-secondary text-xs">
            <span class="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
            <span>{{ t('companyOemOdm.services.odm.features.modeling') }}</span>
          </li>
          <li class="flex items-center gap-3 tz-text-secondary text-xs">
            <span class="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
            <span>{{ t('companyOemOdm.services.odm.features.layup') }}</span>
          </li>
          <li class="flex items-center gap-3 tz-text-secondary text-xs">
            <span class="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
            <span>{{ t('companyOemOdm.services.odm.features.aerodynamics') }}</span>
          </li>
        </ul>
      </div>
    </section>

    <!-- Why Choose Us -->
    <section class="max-w-none w-full">
      <h3 class="text-xl font-bold tz-text-primary text-center mb-6">{{ t('companyOemOdm.advantages.title') }}</h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
         <!-- Advantage 1 -->
         <div class="bg-[var(--tz-card-surface)] p-5 rounded-2xl border tz-border-subtle text-center">
            <div class="w-10 h-10 tz-surface-panel rounded-full flex items-center justify-center mx-auto mb-3 text-amber-400">
               <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" /></svg>
            </div>
            <h4 class="tz-text-primary font-semibold mb-1 text-sm">{{ t('companyOemOdm.advantages.drilling.title') }}</h4>
            <p class="text-xs tz-text-secondary">{{ t('companyOemOdm.advantages.drilling.description') }}</p>
         </div>
         <!-- Advantage 2 -->
         <div class="bg-[var(--tz-card-surface)] p-5 rounded-2xl border tz-border-subtle text-center">
            <div class="w-10 h-10 tz-surface-panel rounded-full flex items-center justify-center mx-auto mb-3 text-amber-400">
               <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" /></svg>
            </div>
            <h4 class="tz-text-primary font-semibold mb-1 text-sm">{{ t('companyOemOdm.advantages.molds.title') }}</h4>
            <p class="text-xs tz-text-secondary">{{ t('companyOemOdm.advantages.molds.description') }}</p>
         </div>
         <!-- Advantage 3 -->
         <div class="bg-[var(--tz-card-surface)] p-5 rounded-2xl border tz-border-subtle text-center">
             <div class="w-10 h-10 tz-surface-panel rounded-full flex items-center justify-center mx-auto mb-3 text-amber-400">
               <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            </div>
            <h4 class="tz-text-primary font-semibold mb-1 text-sm">{{ t('companyOemOdm.advantages.quality.title') }}</h4>
            <p class="text-xs tz-text-secondary">{{ t('companyOemOdm.advantages.quality.description') }}</p>
         </div>
      </div>
    </section>

    <!-- Workflow Process -->
    <section class="max-w-none w-full">
      <h3 class="text-xl font-bold tz-text-primary text-center mb-6">{{ t('companyOemOdm.process.title') }}</h3>
      <div class="grid grid-cols-2 md:grid-cols-5 gap-3">
        <div v-for="(step, index) in steps" :key="index" class="relative group">
           <div class="aspect-square rounded-xl tz-surface-panel border tz-border-subtle flex flex-col items-center justify-center p-3 text-center group-hover:tz-surface-panel transition-colors pointer-cursor">
              <span class="tz-micro-label text-amber-500 font-mono mb-1">0{{ index + 1}}</span>
              <span class="text-xs font-semibold tz-text-primary">{{ t(`companyOemOdm.process.steps.${step}`) }}</span>
           </div>
           <!-- Arrow for desktop -->
           <div v-if="index < steps.length - 1" class="hidden md:block absolute top-1/2 -right-2.5 -translate-y-1/2 tz-text-muted z-10">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
           </div>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="text-center mb-6">
      <h3 class="text-xl font-bold tz-text-primary mb-3">{{ t('companyOemOdm.cta.title') }}</h3>
      <p class="tz-text-secondary mb-6 max-w-2xl mx-auto text-sm">{{ t('companyOemOdm.cta.description') }}</p>
      <NuxtLink :to="localePath('/company/contact')" class="inline-block bg-amber-500 hover:bg-amber-400 text-slate-900 font-bold py-2.5 px-6 rounded-full shadow-lg hover:shadow-amber-500/25 transition-all transform hover:-translate-y-0.5 text-sm">
        {{ t('companyOemOdm.cta.button') }}
      </NuxtLink>
    </section>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages } = usePageMessages('companyOemOdm')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const steps = [
  'consultation',
  'engineering',
  'prototyping',
  'testing',
  'production',
]
</script>
