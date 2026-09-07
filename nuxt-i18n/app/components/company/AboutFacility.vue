<template>
  <section id="facility" class="company-section">
    <h2 class="company-section__title">{{ t('companyAboutFacility.title') }}</h2>
    <p class="company-section__body">
      {{ t('companyAboutFacility.paragraphs.body1') }}
    </p>
    <p class="company-section__body">
      {{ t('companyAboutFacility.paragraphs.body2') }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('companyAboutFacility')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})
</script>
