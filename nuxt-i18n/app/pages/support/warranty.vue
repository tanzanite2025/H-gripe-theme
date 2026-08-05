<template>
  <div class="support-warranty">
    <!-- SEO-friendly hidden H1 -->
    <h1 class="support-page__title support-page__title--sr-only">Warranty</h1>

    <!-- Tab Components -->
    <WarrantyChangeCancelTab 
      v-show="activeTab === 'change-cancel'" 
      @change-tab="setActiveTab" 
    />
    
    <WarrantyDamagedLostTab 
      v-show="activeTab === 'damaged-lost'" 
      :contact-email="supportEmail"
    />
    
    <WarrantyReturnsTab 
      v-show="activeTab === 'returns'" 
      :contact-email="supportEmail"
    />
    
    <WarrantyWarrantyPolicyTab 
      v-show="activeTab === 'warranty'" 
      :contact-email="supportEmail"
      @change-tab="setActiveTab" 
    />
    
    <WarrantyAccidentalDamageTab 
      v-show="activeTab === 'accidental-damage'" 
      @change-tab="setActiveTab" 
    />
    
    <WarrantyProtectionTab 
      v-show="activeTab === 'protection'" 
    />
    
    <WarrantySubmitClaimTab 
      v-show="activeTab === 'submit-warranty'" 
    />

  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from '#imports'
import WarrantyChangeCancelTab from '~/components/warranty/ChangeCancelTab.vue'
import WarrantyDamagedLostTab from '~/components/warranty/DamagedLostTab.vue'
import WarrantyReturnsTab from '~/components/warranty/ReturnsTab.vue'
import WarrantyWarrantyPolicyTab from '~/components/warranty/WarrantyPolicyTab.vue'
import WarrantyAccidentalDamageTab from '~/components/warranty/AccidentalDamageTab.vue'
import WarrantyProtectionTab from '~/components/warranty/ProtectionTab.vue'
import WarrantySubmitClaimTab from '~/components/warranty/SubmitClaimTab.vue'
import {
  isPageSubNavigationTabId,
  warrantyTabs,
  type WarrantyTabId,
} from '~/utils/pageSubNavigation'
import { useSiteSettings } from '~/composables/usePublicSettings'

definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Warranty',
})

const tabs = warrantyTabs
const { siteSettings } = useSiteSettings()

const activeTab = ref<WarrantyTabId>('warranty')
const route = useRoute()
const supportEmail = computed(() => siteSettings.value.contactEmail?.trim() || '')

const setActiveTab = (id: WarrantyTabId | string) => {
  if (!isPageSubNavigationTabId(tabs, id)) return
  activeTab.value = id
  if (typeof window !== 'undefined') {
    const url = new URL(window.location.href)
    url.hash = `#${id}`
    window.history.replaceState(null, '', url.toString())
  }
}

const syncTabWithHash = (hash: string | null | undefined) => {
  if (!hash) return
  const clean = hash.startsWith('#') ? hash.slice(1) : hash
  if (isPageSubNavigationTabId(tabs, clean)) {
    activeTab.value = clean
  }
}

onMounted(() => {
  syncTabWithHash(route.hash)
})

watch(
  () => route.hash,
  (newHash) => {
    syncTabWithHash(newHash)
  }
)
</script>

<style scoped>
.support-warranty {
  margin-top: 0;
}

.support-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: #f9fafb;
}

.support-page__title--sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}
</style>
