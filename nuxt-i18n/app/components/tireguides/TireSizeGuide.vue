<template>
  <div class="tz-text-secondary">
    <h2 class="sizecharts-section__title">{{ t('guidesTireSize.title') }}</h2>
    <div class="mt-1">
      <TireSizeSection @openTireProducts="$emit('openTireProducts')" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import TireSizeSection from '~/components/TireSizeSection.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTireSize')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

defineEmits<{
  (e: 'openTireProducts'): void
}>()
</script>
