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
import { computed } from 'vue'
import WarrantyChangeCancelTab from '~/components/warranty/ChangeCancelTab.vue'
import WarrantyDamagedLostTab from '~/components/warranty/DamagedLostTab.vue'
import WarrantyReturnsTab from '~/components/warranty/ReturnsTab.vue'
import WarrantyWarrantyPolicyTab from '~/components/warranty/WarrantyPolicyTab.vue'
import WarrantyAccidentalDamageTab from '~/components/warranty/AccidentalDamageTab.vue'
import WarrantyProtectionTab from '~/components/warranty/ProtectionTab.vue'
import WarrantySubmitClaimTab from '~/components/warranty/SubmitClaimTab.vue'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import { warrantyTabs } from '~/utils/pageSubNavigation'
import { useSiteSettings } from '~/composables/usePublicSettings'

definePageMeta({
  layout: 'support',
  footerLabelKey: 'support.nav.warranty',
  footerLabelFallback: 'Warranty',
})

useHead({
  title: 'Warranty',
})

const tabs = warrantyTabs
const { siteSettings } = useSiteSettings()

const { activeTab, setActiveTab } = usePageSubNavigationTab({
  tabs,
  basePath: '/support/warranty',
  defaultValue: 'warranty',
})
const supportEmail = computed(() => siteSettings.value.contactEmail?.trim() || '')
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
  color: var(--tz-text-primary);
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
