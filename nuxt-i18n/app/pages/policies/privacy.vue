<template>
  <div class="company-page">
    <h1 class="sr-only">{{ t('privacy.title') }}</h1>

    <div class="policies-content">
      <p class="text-sm tz-text-secondary mb-6">
        {{ t('privacy.pageIntro') }}
      </p>
      <div>
        <PrivacyStatementContent />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import PrivacyStatementContent from '~/components/PrivacyStatementContent.vue'
import { useHead, useI18n, definePageMeta } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'policyTabs.privacy',
  footerLabelFallback: 'Privacy',
})

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('privacy')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

useHead(() => ({
  title: t('privacy.title'),
}))
</script>

<style scoped>
.company-page {
  width: 100%;
  max-width: none;
  margin: 0 auto;
  padding: 0 16px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.policies-content {
  margin-top: 0;
}

@media (max-width: 767px) {
  .company-page {
    max-width: 100%;
    padding-inline: 0;
  }
}
</style>
