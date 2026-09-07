<template>
  <section id="manufacture" class="company-section">
    <SmartAccordion default-id="rim-build">
      <AccordionItem
        id="rim-build"
        :title="t('companyAboutManufacture.sections.rimBuild.title')"
      >
        <div class="p-4 tz-text-secondary text-sm">
          {{ t('companyAboutManufacture.sections.rimBuild.body') }}
        </div>
      </AccordionItem>

      <AccordionItem
        id="wheelset-build"
        :title="t('companyAboutManufacture.sections.wheelsetBuild.title')"
      >
        <div class="p-4 tz-text-secondary text-sm">
          {{ t('companyAboutManufacture.sections.wheelsetBuild.body') }}
        </div>
      </AccordionItem>

      <AccordionItem
        id="carbon-spoke-build"
        :title="t('companyAboutManufacture.sections.carbonSpokeBuild.title')"
      >
        <div class="p-4 tz-text-secondary text-sm">
          {{ t('companyAboutManufacture.sections.carbonSpokeBuild.body') }}
        </div>
      </AccordionItem>
    </SmartAccordion>
  </section>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import AccordionItem from '~/components/ui/AccordionItem.vue'
import SmartAccordion from '~/components/ui/SmartAccordion.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('companyAboutManufacture')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})
</script>
