<template>
  <div class="w-full pb-6">
    <h1 class="sr-only">{{ t('companyOemOdm.title') }}</h1>
    
    <div class="w-full max-w-none">
      <OemOdmServices />
    </div>

    <!-- Feedback Section -->
    <div class="w-full max-w-none px-0 mt-4">
      <UserFeedbackThread
        threadKey="company-oem-odm"
        :title="t('companyOemOdm.feedbackTitle')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { definePageMeta, useHead, useI18n } from '#imports'
import OemOdmServices from '~/components/company/OemOdmServices.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('companyOemOdm')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

definePageMeta({
  layout: 'products',
  footerLabelKey: 'company.nav.oemOdm',
  footerLabelFallback: 'OEM & ODM',
})

useHead({
  title: t('companyOemOdm.title'),
})
</script>
