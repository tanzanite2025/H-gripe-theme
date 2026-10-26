<template>
  <LazyAppFooter
    :key="footerInstanceLocale"
    :hydrate-on-visible="STOREFRONT_FOOTER_HYDRATION_OPTIONS"
    @hydrated="markFooterAsHydrated"
  />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '#imports'
import { STOREFRONT_FOOTER_HYDRATION_OPTIONS } from '~/utils/storefrontLoadingPolicy'

const { locale } = useI18n()
const footerInstanceLocale = ref(locale.value)
const hasFooterHydrated = ref(false)

watch(locale, (nextLocale) => {
  // An untouched SSR footer still contains the original language. Replace that
  // pending instance on a locale change; an active footer updates normally and
  // keeps subscription input and other local state.
  if (!hasFooterHydrated.value) footerInstanceLocale.value = nextLocale
}, { flush: 'sync' })

function markFooterAsHydrated() {
  hasFooterHydrated.value = true
}
</script>
