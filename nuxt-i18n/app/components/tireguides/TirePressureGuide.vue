<template>
  <div class="tire-pressure-guide">
    <InteractiveToolPageHeader
      :title="t('guidesTirePressure.title')"
      :description="t('guidesTirePressure.description')"
    />
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center">
      <TirePressureSection />
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import TirePressureSection from '~/components/TirePressureSection.vue'
import InteractiveToolPageHeader from '~/components/InteractiveToolPageHeader.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTirePressure')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

</script>
