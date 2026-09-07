<template>
  <section id="qualitycontrol" class="company-section">
    <h2 class="company-section__title">{{ t('companyAboutQualityControl.title') }}</h2>
    <p class="company-section__body">
      {{ t('companyAboutQualityControl.paragraphs.body1') }}
    </p>
    <p class="company-section__body">
      {{ t('companyAboutQualityControl.paragraphs.body2') }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('companyAboutQualityControl')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})
</script>
