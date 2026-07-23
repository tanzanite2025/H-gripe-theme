<template>
  <div class="w-full pb-6">
    <h1 class="sr-only">{{ t('company.nav.oemOdm') }}</h1>
    
    <div class="w-full max-w-none">
      <OemOdmServices />
    </div>

    <!-- Feedback Section -->
    <div class="w-full max-w-none px-0 mt-4">
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

const { t } = useI18n()

definePageMeta({
  layout: 'products',
})

useHead({
  title: t('company.nav.oemOdm'),
})
</script>
