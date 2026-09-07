<template>
  <div class="text-sm tz-text-secondary leading-relaxed space-y-4">
    <p class="font-semibold tz-text-primary">
      {{ t('privacy.policyHeading') }}
    </p>
    <p class="tz-text-muted text-xs">
      {{ t('privacy.effectiveDate') }}
    </p>

    <p class="font-semibold tz-text-primary mt-4">
      {{ t('privacy.section1Title') }}
    </p>
    <p>
      {{ introductionText }}
    </p>

    <p class="font-semibold tz-text-primary mt-4">
      {{ t('privacy.section2Title') }}
    </p>
    <p>
      {{ t('privacy.section2Intro') }}
    </p>
    <ul class="list-disc list-inside space-y-1">
      <li>
        <span class="font-semibold">
          {{ t('privacy.identityLabel') }}:
        </span>
        {{ t('privacy.identityDesc') }}
      </li>
      <li>
        <span class="font-semibold">
          {{ t('privacy.contactLabel') }}:
        </span>
        {{ t('privacy.contactDesc') }}
      </li>
      <li>
        <span class="font-semibold">
          {{ t('privacy.transactionLabel') }}:
        </span>
        {{ t('privacy.transactionDesc') }}
      </li>
      <li>
        <span class="font-semibold">
          {{ t('privacy.technicalLabel') }}:
        </span>
        {{ t('privacy.technicalDesc') }}
      </li>
      <li>
        <span class="font-semibold">
          {{ t('privacy.marketingLabel') }}:
        </span>
        {{ t('privacy.marketingDesc') }}
      </li>
    </ul>

    <p class="font-semibold tz-text-primary mt-4">
      {{ t('privacy.section3Title') }}
    </p>
    <p>
      {{ t('privacy.section3Intro') }}
    </p>
    <ul class="list-disc list-inside space-y-1">
      <li>
        {{ t('privacy.useProcessOrders') }}
      </li>
      <li>
        {{ t('privacy.useManageAccount') }}
      </li>
      <li>
        {{ t('privacy.useImproveSite') }}
      </li>
      <li>
        {{ t('privacy.usePromotions') }}
      </li>
      <li>
        {{ t('privacy.useLegal') }}
      </li>
    </ul>

    <p class="font-semibold tz-text-primary mt-4">
      {{ t('privacy.section4Title') }}
    </p>
    <p>
      {{ t('privacy.section4Intro') }}
    </p>
    <ul class="list-disc list-inside space-y-1">
      <li>
        {{ t('privacy.shareServiceProviders') }}
      </li>
      <li>
        {{ t('privacy.shareLegalAuthorities') }}
      </li>
      <li>
        {{ t('privacy.shareBusinessTransfers') }}
      </li>
    </ul>
    <p>
      {{ t('privacy.shareNoSale') }}
    </p>

    <p class="font-semibold tz-text-primary mt-4">
      {{ t('privacy.section5Title') }}
    </p>
    <p>
      {{ t('privacy.section5Intro') }}
    </p>
    <ul class="list-disc list-inside space-y-1">
      <li>
        {{ t('privacy.cookiesRemember') }}
      </li>
      <li>
        {{ t('privacy.cookiesAnalytics') }}
      </li>
      <li>
        {{ t('privacy.cookiesPersonalize') }}
      </li>
    </ul>
    <p>
      {{ t('privacy.cookiesSettings') }}
    </p>

    <p class="font-semibold tz-text-primary mt-4">
      {{ t('privacy.section6Title') }}
    </p>
    <p>
      {{ t('privacy.section6Body') }}
    </p>

    <p class="font-semibold tz-text-primary mt-4">
      {{ t('privacy.section7Title') }}
    </p>
    <p>
      {{ t('privacy.section7Intro') }}
    </p>
    <ul class="list-disc list-inside space-y-1">
      <li>
        {{ t('privacy.rightsAccess') }}
      </li>
      <li>
        {{ t('privacy.rightsWithdrawConsent') }}
      </li>
      <li>
        {{ t('privacy.rightsObjectProcessing') }}
      </li>
      <li>
        {{ t('privacy.rightsPortability') }}
      </li>
    </ul>
    <p>
      {{ t('privacy.rightsContact') }}
    </p>

    <p class="font-semibold tz-text-primary mt-4">
      {{ t('privacy.section8Title') }}
    </p>
    <p>
      {{ t('privacy.section8Body') }}
    </p>

    <p class="font-semibold tz-text-primary mt-4">
      {{ t('privacy.section9Title') }}
    </p>
    <p>
      {{ t('privacy.section9Body') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n, useRequestURL } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('privacy')
const requestUrl = useRequestURL()
const siteLabel = computed(() => requestUrl.hostname || requestUrl.host || 'this website')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const introductionText = computed(() => (
  t('privacy.section1Body', { site: siteLabel.value })
))
</script>
