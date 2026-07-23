<template>
  <div class="w-full pb-6">
    <h1 class="sr-only">{{ t('company.nav.certificates') }}</h1>
    
    <div class="w-full max-w-none">
      <CertificatesGallery />
    </div>

    <!-- Feedback Section -->
    <div class="w-full max-w-none px-0 mt-4">
      <UserFeedbackThread
        threadKey="company-certificates"
        title="Share your feedback about our Certificates"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { definePageMeta, useHead, useI18n } from '#imports'
import CertificatesGallery from '~/components/company/CertificatesGallery.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

const { t } = useI18n()

definePageMeta({
  layout: 'products',
})

useHead({
  title: t('company.nav.certificates'),
})
</script>
