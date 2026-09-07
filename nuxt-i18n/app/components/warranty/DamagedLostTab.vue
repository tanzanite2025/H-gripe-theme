<template>
  <section id="damaged-lost" class="support-section">
    <h2 class="support-section__title text-center mb-8 tz-text-primary">
      {{ t('warrantyDamagedLost.title') }}
    </h2>

    <div class="w-full max-w-none text-left space-y-8">
      <div class="bg-emerald-50 border-l-4 border-emerald-500 p-4 rounded-r-lg">
        <p class="text-emerald-200 font-medium">{{ t('warrantyDamagedLost.alert') }}</p>
      </div>

      <div class="space-y-10">
        <div class="relative pl-12">
          <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">1</div>
          <h3 class="tz-text-secondary font-bold text-lg mb-3 pt-0.5">
            {{ t('warrantyDamagedLost.sections.normalAcceptance.title') }}
          </h3>
          <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-emerald-600/60">
            <li>{{ t('warrantyDamagedLost.sections.normalAcceptance.items.0') }}</li>
            <li>{{ t('warrantyDamagedLost.sections.normalAcceptance.items.1') }}</li>
            <li>
              {{ t('warrantyDamagedLost.sections.normalAcceptance.items.2') }}
              <strong class="tz-text-primary">
                {{ t('warrantyDamagedLost.sections.normalAcceptance.reshipmentWindow') }}
              </strong>{{ t('warrantyDamagedLost.sections.normalAcceptance.reshipmentSuffix') }}
            </li>
          </ul>
        </div>

        <div class="relative pl-12">
          <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">2</div>
          <h3 class="tz-text-secondary font-bold text-lg mb-3 pt-0.5">
            {{ t('warrantyDamagedLost.sections.damagedAccepted.title') }}
          </h3>
          <div class="bg-rose-500/5 rounded-lg p-4 border border-rose-500/10">
              <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-rose-500">
                <li>{{ t('warrantyDamagedLost.sections.damagedAccepted.items.0') }}</li>
                <li>{{ t('warrantyDamagedLost.sections.damagedAccepted.items.1') }}</li>
              </ul>
          </div>
        </div>

        <div class="relative pl-12">
          <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">3</div>
          <h3 class="tz-text-secondary font-bold text-lg mb-3 pt-0.5">
            {{ t('warrantyDamagedLost.sections.transit.title') }}
          </h3>
          <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-emerald-600/60">
            <li>{{ t('warrantyDamagedLost.sections.transit.items.packaging') }}</li>
            <li>
              {{ t('warrantyDamagedLost.sections.transit.items.contactPrefix') }}
              <strong class="tz-text-primary">
                {{ t('warrantyDamagedLost.sections.transit.items.contactWindow') }}
              </strong><template v-if="contactEmail">{{ t('warrantyDamagedLost.sections.transit.items.contactEmailPrefix') }}<a :href="contactEmailHref" class="text-emerald-600 hover:text-emerald-700 transition-colors">{{ contactEmail }}</a></template>{{ t('warrantyDamagedLost.sections.transit.items.contactSuffix') }}
            </li>
          </ul>
        </div>

        <div class="relative pl-12">
          <div class="absolute left-0 top-0 w-8 h-8 rounded-full tz-surface-panel border tz-border-subtle flex items-center justify-center tz-text-secondary font-bold text-sm shadow-md">4</div>
          <h3 class="tz-text-secondary font-bold text-lg mb-3 pt-0.5">
            {{ t('warrantyDamagedLost.sections.lost.title') }}
          </h3>
          <ul class="space-y-3 tz-text-secondary text-sm leading-relaxed list-disc pl-4 marker:text-emerald-600/60">
            <li>
              {{ t('warrantyDamagedLost.sections.lost.items.contactPrefix') }}<template v-if="contactEmail">{{ t('warrantyDamagedLost.sections.lost.items.contactEmailPrefix') }}<a :href="contactEmailHref" class="text-emerald-600 hover:text-emerald-700 transition-colors">{{ contactEmail }}</a></template>{{ t('warrantyDamagedLost.sections.lost.items.contactSuffix') }}
            </li>
            <li>{{ t('warrantyDamagedLost.sections.lost.items.insured') }}</li>
          </ul>
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
const { loadPageMessages } = usePageMessages('warrantyDamagedLost')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const contactEmail = computed(() => props.contactEmail?.trim() || '')
const contactEmailHref = computed(() => `mailto:${contactEmail.value}`)
</script>
