<template>
  <section id="warranty" class="support-section">
    <h2 class="support-section__title text-center mb-8 tz-text-primary">
      {{ t('warrantyPolicy.title') }}
    </h2>

    <div class="w-full max-w-none text-left space-y-10">
      <div class="space-y-8">
         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">1</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">
             {{ t('warrantyPolicy.sections.applicableProducts.title') }}
           </h3>
           <ul class="space-y-2 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-emerald-600/60">
              <li>{{ t('warrantyPolicy.sections.applicableProducts.items.0') }}</li>
              <li>{{ t('warrantyPolicy.sections.applicableProducts.items.1') }}</li>
           </ul>
         </div>

         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">2</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">
             {{ t('warrantyPolicy.sections.coverageOptions.title') }}
           </h3>
           <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-decimal pl-4 marker:tz-text-muted">
              <li>{{ t('warrantyPolicy.sections.coverageOptions.items.0') }}</li>
              <li>{{ t('warrantyPolicy.sections.coverageOptions.items.1') }}</li>
           </ul>
         </div>

         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">3</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-4 pt-0.5">
             {{ t('warrantyPolicy.sections.policyConditions.title') }}
           </h3>
           <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="tz-surface-panel rounded-lg p-5 border tz-border-subtle md:col-span-2">
                  <h4 class="text-emerald-600 font-bold text-sm mb-3">
                    {{ t('warrantyPolicy.sections.policyConditions.manufacturingDefects.title') }}
                  </h4>
                  <ul class="space-y-2 tz-text-secondary text-xs leading-relaxed list-disc pl-4 marker:text-emerald-600/40">
                      <li>{{ t('warrantyPolicy.sections.policyConditions.manufacturingDefects.items.0') }}</li>
                      <li>{{ t('warrantyPolicy.sections.policyConditions.manufacturingDefects.items.1') }}</li>
                  </ul>
              </div>
               <div class="tz-surface-panel rounded-lg p-5 border tz-border-subtle">
                  <h4 class="text-emerald-600 font-bold text-sm mb-3">
                    {{ t('warrantyPolicy.sections.policyConditions.productUpgrade.title') }}
                  </h4>
                  <ul class="space-y-2 tz-text-secondary text-xs leading-relaxed list-disc pl-4 marker:text-emerald-600/40">
                      <li>{{ t('warrantyPolicy.sections.policyConditions.productUpgrade.items.0') }}</li>
                      <li>{{ t('warrantyPolicy.sections.policyConditions.productUpgrade.items.1') }}</li>
                  </ul>
              </div>
              <div class="tz-surface-panel rounded-lg p-5 border tz-border-subtle">
                  <h4 class="text-emerald-600 font-bold text-sm mb-3">
                    {{ t('warrantyPolicy.sections.policyConditions.nonTransferable.title') }}
                  </h4>
                  <p class="tz-text-secondary text-xs leading-relaxed">
                      {{ t('warrantyPolicy.sections.policyConditions.nonTransferable.body') }}
                  </p>
              </div>
              <div class="tz-surface-panel rounded-lg p-5 border tz-border-subtle md:col-span-2">
                  <h4 class="text-emerald-600 font-bold text-sm mb-3">
                    {{ t('warrantyPolicy.sections.policyConditions.claimTimeframe.title') }}
                  </h4>
                  <ul class="space-y-2 tz-text-secondary text-xs leading-relaxed list-disc pl-4 marker:text-emerald-600/40">
                       <li>{{ t('warrantyPolicy.sections.policyConditions.claimTimeframe.items.0') }}</li>
                       <li>{{ t('warrantyPolicy.sections.policyConditions.claimTimeframe.items.1') }}</li>
                  </ul>
              </div>
           </div>
         </div>

         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">4</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">
             {{ t('warrantyPolicy.sections.exclusions.title') }}
           </h3>
           <p class="tz-text-secondary text-sm mb-3">
             {{ t('warrantyPolicy.sections.exclusions.intro') }}
           </p>
           <ul class="space-y-2 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-red-400/50">
              <li>{{ t('warrantyPolicy.sections.exclusions.items.0') }}</li>
              <li>{{ t('warrantyPolicy.sections.exclusions.items.1') }}</li>
              <li>{{ t('warrantyPolicy.sections.exclusions.items.2') }}</li>
              <li>{{ t('warrantyPolicy.sections.exclusions.items.3') }}</li>
              <li>{{ t('warrantyPolicy.sections.exclusions.items.4') }}</li>
           </ul>
         </div>

         <div class="relative pl-12">
           <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">5</div>
           <h3 class="tz-text-secondary font-bold text-lg mb-2 pt-0.5">
             {{ t('warrantyPolicy.sections.contact.title') }}
           </h3>
           <div class="space-y-3 tz-text-secondary text-sm">
              <p v-if="contactEmail">
                {{ t('warrantyPolicy.sections.contact.emailLabel') }}
                <a :href="contactEmailHref" class="text-emerald-600 hover:text-emerald-700 transition-colors">{{ contactEmail }}</a>
              </p>
              <p>
                  <a href="#" @click.prevent="$emit('change-tab', 'submit-warranty')" class="inline-flex items-center gap-1 text-emerald-600 hover:text-emerald-700 font-bold transition-colors">
                 {{ t('warrantyPolicy.sections.contact.submitClaim') }}
                 <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" /></svg>
               </a>
              </p>
           </div>
         </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

const props = defineProps<{
  contactEmail?: string
}>()

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('warrantyPolicy')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

defineEmits<{
  (e: 'change-tab', id: string): void
}>()

const contactEmail = computed(() => props.contactEmail?.trim() || '')
const contactEmailHref = computed(() => `mailto:${contactEmail.value}`)
</script>
