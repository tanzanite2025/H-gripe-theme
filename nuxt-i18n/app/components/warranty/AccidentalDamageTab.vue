<template>
  <section id="accidental-damage" class="support-section">
    <h2 class="support-section__title text-center mb-8 tz-text-primary">
      {{ t('warrantyAccidentalDamage.title') }}
    </h2>

    <div class="w-full max-w-none text-left space-y-10">
      <div class="space-y-8">
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">1</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">
             {{ t('warrantyAccidentalDamage.sections.scope.title') }}
           </h3>
           <p class="tz-text-secondary text-sm leading-relaxed">
             {{ t('warrantyAccidentalDamage.sections.scope.body') }}
           </p>
         </div>

         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">2</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">
             {{ t('warrantyAccidentalDamage.sections.productGroup.title') }}
           </h3>
           <p class="tz-text-secondary text-sm leading-relaxed">
             {{ t('warrantyAccidentalDamage.sections.productGroup.body') }}
           </p>
         </div>

         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">3</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">
             {{ t('warrantyAccidentalDamage.sections.coverage.title') }}
           </h3>
           <ul class="space-y-2 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-emerald-600/60">
             <li>{{ t('warrantyAccidentalDamage.sections.coverage.items.0') }}</li>
             <li>{{ t('warrantyAccidentalDamage.sections.coverage.items.1') }}</li>
           </ul>
         </div>

         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">4</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">
             {{ t('warrantyAccidentalDamage.sections.conditions.title') }}
           </h3>
           <ul class="space-y-4 tz-text-secondary text-sm leading-relaxed list-decimal pl-4 marker:tz-text-muted">
             <li>{{ t('warrantyAccidentalDamage.sections.conditions.items.0') }}</li>
             <li>
               {{ t('warrantyAccidentalDamage.sections.conditions.items.1') }}
               <p class="mt-2 text-xs tz-surface-subtle p-2 rounded border tz-border-strong/5">
                 {{ t('warrantyAccidentalDamage.sections.conditions.note') }}
               </p>
             </li>
             <li>{{ t('warrantyAccidentalDamage.sections.conditions.items.2') }}</li>
           </ul>
         </div>

         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">5</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">
             {{ t('warrantyAccidentalDamage.sections.claimProcess.title') }}
           </h3>
           <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-emerald-600/60">
             <li>{{ t('warrantyAccidentalDamage.sections.claimProcess.items.0') }}</li>
             <li>{{ t('warrantyAccidentalDamage.sections.claimProcess.items.1') }}</li>
             <li class="list-none -ml-4 mt-2">
               <a href="#" @click.prevent="$emit('change-tab', 'submit-warranty')" class="inline-flex items-center gap-1 text-emerald-600 hover:text-emerald-700 font-bold transition-colors">
                 {{ t('warrantyAccidentalDamage.sections.claimProcess.submitClaim') }}
                 <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" /></svg>
               </a>
             </li>
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
const { loadPageMessages } = usePageMessages('warrantyAccidentalDamage')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

defineEmits<{
  (e: 'change-tab', id: string): void
}>()
</script>
