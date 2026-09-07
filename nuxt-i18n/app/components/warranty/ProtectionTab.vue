<template>
  <section id="protection" class="support-section">
    <h2 class="support-section__title text-center mb-8 tz-text-primary">
      {{ t('warrantyProtection.title') }}
    </h2>

    <div class="w-full max-w-none text-left space-y-10">
      <div class="bg-emerald-50 border-l-4 border-emerald-500 p-4 rounded-r-lg">
        <h3 class="text-emerald-200 font-bold text-lg mb-1">
          {{ t('warrantyProtection.intro.title') }}
        </h3>
        <p class="tz-text-secondary text-sm">
          {{ t('warrantyProtection.intro.bodyPrefix') }}
          <strong class="tz-text-primary">{{ t('warrantyProtection.intro.legalAgreement') }}</strong>{{ t('warrantyProtection.intro.bodySuffix') }}
        </p>
      </div>

      <div class="space-y-8">
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">1</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-3 pt-0.5">
             {{ t('warrantyProtection.sections.reasons.title') }}
           </h3>
           <p class="tz-text-secondary text-sm mb-3">
             {{ t('warrantyProtection.sections.reasons.body') }}
           </p>
           <ul class="space-y-2 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-emerald-600/60">
             <li><strong class="tz-text-primary">{{ t('warrantyProtection.sections.reasons.items.0') }}</strong></li>
             <li><strong class="tz-text-primary">{{ t('warrantyProtection.sections.reasons.items.1') }}</strong></li>
           </ul>
         </div>

         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">2</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">
             {{ t('warrantyProtection.sections.timeframe.title') }}
           </h3>
           <p class="tz-text-secondary text-sm leading-relaxed">
             {{ t('warrantyProtection.sections.timeframe.bodyPrefix') }}
             <strong class="tz-text-primary">{{ t('warrantyProtection.sections.timeframe.window') }}</strong>{{ t('warrantyProtection.sections.timeframe.bodySuffix') }}
           </p>
         </div>

         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">3</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">
             {{ t('warrantyProtection.sections.claimProcess.title') }}
           </h3>
           <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-emerald-600/60">
             <li>{{ t('warrantyProtection.sections.claimProcess.items.0') }}</li>
             <li>{{ t('warrantyProtection.sections.claimProcess.items.1') }}</li>
           </ul>
         </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('warrantyProtection')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})
</script>
