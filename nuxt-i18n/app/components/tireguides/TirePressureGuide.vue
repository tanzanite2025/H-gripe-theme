<template>
  <div class="tire-pressure-guide">
    <h2 class="sizecharts-section__title">{{ t('guidesTirePressure.title') }}</h2>
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center">
      <TirePressureSection @openTireProducts="$emit('openTireProducts')" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import TirePressureSection from '~/components/TirePressureSection.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTirePressure')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

defineEmits<{
  (e: 'openTireProducts'): void
}>()
</script>
