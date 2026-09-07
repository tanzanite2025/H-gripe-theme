<template>
  <div class="w-full pb-6">
    <h1 class="sr-only">{{ t('companyCertificates.title') }}</h1>
    
    <div class="w-full max-w-none">
      <CertificatesGallery />
    </div>

    <!-- Feedback Section -->
    <div class="w-full max-w-none px-0 mt-4">
      <UserFeedbackThread
        threadKey="company-certificates"
        :title="t('companyCertificates.feedbackTitle')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { definePageMeta, useHead, useI18n } from '#imports'
import CertificatesGallery from '~/components/company/CertificatesGallery.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('companyCertificates')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

definePageMeta({
  layout: 'products',
  footerLabelKey: 'company.nav.certificates',
  footerLabelFallback: 'Certificates',
})

useHead({
  title: t('companyCertificates.title'),
})
</script>
