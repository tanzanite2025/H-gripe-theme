<template>
  <div class="w-full pt-4 md:pt-8 pb-6">
    <h1 class="sr-only">{{ t('company.nav.oemOdm') }}</h1>
    
    <div class="max-w-7xl mx-auto">
      <OemOdmServices />
    </div>

    <!-- FAQ Section -->
    <div class="max-w-[960px] mx-auto mt-4 px-0">
      <PageFaq pageId="company-oem-odm" />
    </div>

    <!-- Feedback Section -->
    <div class="max-w-4xl mx-auto px-0 mt-4">
      <UserFeedbackThread
        threadKey="company-oem-odm"
        title="Share your feedback about our OEM/ODM Services"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { definePageMeta, useHead, useI18n } from '#imports'
import OemOdmServices from '~/components/company/OemOdmServices.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import PageFaq from '~/components/PageFaq.vue'

const { t } = useI18n()

definePageMeta({
  layout: 'products',
})

useHead({
  title: t('company.nav.oemOdm'),
})
</script>
