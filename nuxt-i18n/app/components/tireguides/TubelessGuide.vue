<template>
  <div class="tubeless-guide space-y-8">
    <!-- 1. Intro Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center">
      <h2 class="text-xl font-bold tz-text-secondary mb-6 flex items-center justify-center gap-2">
        {{ t('guidesTireTubeless.title') }}
      </h2>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 w-full max-w-5xl mx-auto">
         <GuideImage
            class="rounded-xl overflow-hidden shadow-md"
            src="/public/tiresizecharts/tubelesstires/tubelesstires-innertubes-tubulartires1.webp"
            :alt="t('guidesTireTubeless.images.constructionAlt')"
            :zoomOnClick="true"
            :caption="t('guidesTireTubeless.images.constructionCaption')"
          />
          <GuideImage
            class="rounded-xl overflow-hidden shadow-md"
            src="/public/tiresizecharts/tubelesstires/tubelesstires-innertubes-tubulartires2.webp"
            :alt="t('guidesTireTubeless.images.performanceAlt')"
            :zoomOnClick="true"
            :caption="t('guidesTireTubeless.images.performanceCaption')"
          />
      </div>
    </div>

    <!-- 2. System Comparison (Definitions) -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center">
       <div class="flex items-center justify-center gap-3 mb-6">
         <span class="h-px w-8 tz-surface-panel"></span>
         <h3 class="text-lg font-bold tz-text-secondary uppercase tracking-wider">{{ t('guidesTireTubeless.definitions.title') }}</h3>
         <span class="h-px w-8 tz-surface-panel"></span>
       </div>

       <div class="grid md:grid-cols-2 gap-8 items-stretch">
          <!-- Inner Tube Card -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-5 shadow-md flex flex-col items-center border tz-border-subtle">
             <strong class="block tz-text-primary text-lg font-bold mb-3">{{ t('guidesTireTubeless.definitions.innerTube.title') }}</strong>
          <p class="tz-text-secondary text-sm leading-relaxed mb-4">
               {{ t('guidesTireTubeless.definitions.innerTube.description') }}
             </p>
             <div class="mt-auto pt-4 border-t tz-border-subtle w-full text-center">
                <span class="text-xs font-mono tz-text-muted uppercase">{{ t('guidesTireTubeless.definitions.innerTube.label') }}</span>
             </div>
          </div>

           <!-- Tubeless Card -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-5 shadow-md flex flex-col items-center border tz-border-subtle ring-1 ring-[rgba(255,255,255,0.04)]">
             <strong class="block tz-text-primary text-lg font-bold mb-3">{{ t('guidesTireTubeless.definitions.tubeless.title') }}</strong>
          <p class="tz-text-secondary text-sm leading-relaxed mb-4">
               {{ t('guidesTireTubeless.definitions.tubeless.description') }}
             </p>
             <div class="mt-auto pt-4 border-t tz-border-subtle w-full text-center">
                <span class="text-xs font-mono tz-text-muted uppercase">{{ t('guidesTireTubeless.definitions.tubeless.label') }}</span>
             </div>
          </div>
       </div>

       <div class="mt-8">
          <button
            type="button"
              class="inline-flex items-center justify-center rounded-full bg-[var(--tz-card-surface)] border tz-border-subtle px-6 py-2 text-xs font-bold uppercase tracking-wider tz-text-primary shadow-md transition-colors hover:border-[rgba(5, 150, 105,0.28)] hover:tz-surface-subtle hover:text-[var(--tz-site-accent)]"
            @click="$emit('changeTab', 'choose')"
          >
             {{ t('guidesTireTubeless.definitions.button') }}
          </button>
       </div>
    </div>

    <!-- 3. Suitability Card -->
     <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center border tz-border-subtle">
        <h3 class="text-xl font-bold tz-text-secondary mb-4">{{ t('guidesTireTubeless.suitability.title') }}</h3>
      <p class="tz-text-secondary text-sm leading-relaxed max-w-3xl mx-auto mb-6">
           {{ t('guidesTireTubeless.suitability.innerTube') }}
           <br class="mb-2 block content-['']" />
           {{ t('guidesTireTubeless.suitability.tubeless') }}
        </p>

        <!-- Compatibility Note -->
        <div class="mt-6 p-4 rounded-xl bg-[var(--tz-card-surface)] border border-amber-500/20 text-left max-w-3xl mx-auto shadow-md">
          <strong class="block text-amber-400 text-sm mb-2">{{ t('guidesTireTubeless.suitability.noteTitle') }}:</strong>
              <p class="tz-text-secondary text-xs leading-relaxed">
            {{ t('guidesTireTubeless.suitability.note') }}
          </p>
        </div>
     </div>

    <!-- 4. Key Differences Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center">
       <h3 class="text-lg font-bold tz-text-secondary mb-6 uppercase tracking-wider">{{ t('guidesTireTubeless.advantages.title') }}</h3>
       
       <div class="grid gap-4 md:grid-cols-3">
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-4 shadow-md border tz-border-subtle">
              <strong class="block tz-text-primary text-sm mb-2">{{ t('guidesTireTubeless.advantages.selfSealing.title') }}</strong>
              <p class="text-xs tz-text-secondary leading-relaxed">{{ t('guidesTireTubeless.advantages.selfSealing.description') }}</p>
          </div>
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-4 shadow-md border tz-border-subtle">
              <strong class="block tz-text-primary text-sm mb-2">{{ t('guidesTireTubeless.advantages.lowerPressure.title') }}</strong>
              <p class="text-xs tz-text-secondary leading-relaxed">{{ t('guidesTireTubeless.advantages.lowerPressure.description') }}</p>
          </div>
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-4 shadow-md border tz-border-subtle">
              <strong class="block tz-text-primary text-sm mb-2">{{ t('guidesTireTubeless.advantages.rollingResistance.title') }}</strong>
              <p class="text-xs tz-text-secondary leading-relaxed">{{ t('guidesTireTubeless.advantages.rollingResistance.description') }}</p>
          </div>
       </div>
    </div>

    <!-- Navigation CTAs -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-6 text-center flex flex-col justify-center items-center hover:-translate-y-1 transition-transform">
           <h4 class="tz-text-secondary font-bold mb-2">{{ t('guidesTireTubeless.navigation.installation.title') }}</h4>
           <p class="text-xs tz-text-muted mb-4">{{ t('guidesTireTubeless.navigation.installation.description') }}</p>
           <button
             type="button"
              class="rounded-full bg-[var(--tz-card-surface)] border tz-border-subtle px-6 py-2 text-xs font-bold uppercase tracking-wider tz-text-primary shadow-md transition-colors hover:border-[rgba(5, 150, 105,0.28)] hover:tz-surface-subtle hover:text-[var(--tz-site-accent)]"
             @click="$emit('changeTab', 'installation')"
           >
             {{ t('guidesTireTubeless.navigation.installation.button') }}
           </button>
        </div>

        <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-6 text-center flex flex-col justify-center items-center hover:-translate-y-1 transition-transform">
           <h4 class="tz-text-secondary font-bold mb-2">{{ t('guidesTireTubeless.navigation.pressure.title') }}</h4>
           <p class="text-xs tz-text-muted mb-4">{{ t('guidesTireTubeless.navigation.pressure.description') }}</p>
           <button
             type="button"
              class="rounded-full bg-[var(--tz-card-surface)] border tz-border-subtle px-6 py-2 text-xs font-bold uppercase tracking-wider tz-text-primary shadow-md transition-colors hover:border-[rgba(5, 150, 105,0.28)] hover:tz-surface-subtle hover:text-[var(--tz-site-accent)]"
             @click="$emit('changeTab', 'choose')"
           >
             {{ t('guidesTireTubeless.navigation.pressure.button') }}
           </button>
        </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTireTubeless')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

// Match SizeChartsTabId from parent
type GuideTabId = 'size' | 'match' | 'tubeless' | 'installation' | 'choose' | 'rims' | 'tube'

defineEmits<{
  (e: 'changeTab', tab: GuideTabId): void
}>()
</script>
