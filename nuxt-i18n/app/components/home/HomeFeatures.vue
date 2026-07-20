<template>
  <section class="bg-transparent text-white py-8 sm:py-12 lg:py-20">
    <div class="page-content-shell px-0 md:px-6">
      
      <div class="grid lg:grid-cols-12 gap-5 sm:gap-10 lg:gap-16 items-start">
        
        <!-- Left Column: Header & Context -->
        <div class="lg:col-span-5 lg:sticky lg:top-32 self-start space-y-3 sm:space-y-6">
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-teal-500/10 border border-teal-500/20 text-teal-300 text-xs font-medium uppercase tracking-wider">
            <span>Result Driven</span>
          </div>
          
          <h2 class="text-2xl sm:text-3xl font-bold text-white leading-tight">
            {{ t('home.features.title') }}
          </h2>
          
          <div class="h-1 w-20 rounded-full bg-gradient-to-r from-teal-400 to-blue-500 shadow-[0_0_12px_rgba(45,212,191,0.5)]"></div>
          
          <p class="text-lg text-slate-400 leading-relaxed max-w-md">
            {{ t('home.features.subtitle') }}
          </p>

          <!-- Decorative Visual (Abstract) -->
          <div class="hidden lg:block mt-12 p-8 rounded-3xl bg-slate-900/50 border border-white/5 relative overflow-hidden">
             <div class="absolute inset-0 bg-gradient-to-br from-teal-500/5 to-blue-600/5"></div>
             <div class="relative z-10 flex items-center gap-4 text-slate-400 text-sm">
                <svg class="w-5 h-5 text-teal-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
                <span>Engineered for efficiency</span>
             </div>
             <div class="relative z-10 flex items-center gap-4 text-slate-400 text-sm mt-3">
                <svg class="w-5 h-5 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                <span>Validated Quality Control</span>
             </div>
          </div>
        </div>

        <!-- Right Column: Feature Grid -->
        <div class="lg:col-span-7 grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6">
          <div
            v-for="item in items"
            :key="item.titleKey"
            class="group relative rounded-2xl premium-card p-6 hover:bg-slate-800/60 transition-colors duration-300"
          >


            <h3 class="text-lg font-semibold text-white mb-2 group-hover:text-teal-50 transition-colors">{{ t(item.titleKey) }}</h3>
            <p class="text-sm leading-relaxed text-slate-400 group-hover:text-slate-300 transition-colors">{{ t(item.descriptionKey) }}</p>
          </div>
        </div>

      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'

const { t } = useI18n()

const items = computed(() => {
  return [
    {
      titleKey: 'home.features.items.qc.title',
      descriptionKey: 'home.features.items.qc.description',
    },
    {
      titleKey: 'home.features.items.options.title',
      descriptionKey: 'home.features.items.options.description',
    },
    {
      titleKey: 'home.features.items.oem.title',
      descriptionKey: 'home.features.items.oem.description',
    },
    {
      titleKey: 'home.features.items.lightweight.title',
      descriptionKey: 'home.features.items.lightweight.description',
    },
  ]
})
</script>
