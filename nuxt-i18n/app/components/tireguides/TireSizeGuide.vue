<template>
  <div class="tz-text-secondary">
    <InteractiveToolPageHeader
      :title="t('guidesTireSize.title')"
    />
    <div class="mt-1">
      <TireSizeSection />
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import TireSizeSection from '~/components/TireSizeSection.vue'
import InteractiveToolPageHeader from '~/components/InteractiveToolPageHeader.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTireSize')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

</script>
