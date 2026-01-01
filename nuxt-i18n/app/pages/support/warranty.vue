<template>
  <div class="support-warranty">
    <!-- SEO-friendly hidden H1 -->
    <h1 class="support-page__title support-page__title--sr-only">Warranty</h1>

    <!-- Tabs header -->
    <div class="nav-pill-tabs" role="tablist">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === tab.id }"
        @click="setActiveTab(tab.id)"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Tab Components -->
    <WarrantyChangeCancelTab 
      v-show="activeTab === 'change-cancel'" 
      @change-tab="setActiveTab" 
    />
    
    <WarrantyDamagedLostTab 
      v-show="activeTab === 'damaged-lost'" 
    />
    
    <WarrantyReturnsTab 
      v-show="activeTab === 'returns'" 
    />
    
    <WarrantyWarrantyPolicyTab 
      v-show="activeTab === 'warranty'" 
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

    <!-- FAQ Section -->
    <section class="support-section mt-8">
      <PageFaq 
        page-id="support-warranty"
        theme="dark"
        :show-categories="true"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from '#imports'
import PageFaq from '~/components/PageFaq.vue'

definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Warranty',
})

type WarrantyTabId = 
  | 'change-cancel' 
  | 'damaged-lost' 
  | 'returns' 
  | 'warranty' 
  | 'accidental-damage' 
  | 'protection' 
  | 'submit-warranty'

const tabs: { id: WarrantyTabId; label: string }[] = [
  { id: 'change-cancel', label: 'Change / Cancel' },
  { id: 'damaged-lost', label: 'Damaged or Lost Goods' },
  { id: 'returns', label: 'Returns' },
  { id: 'warranty', label: 'Warranty' },
  { id: 'accidental-damage', label: 'Accidental Damage' },
  { id: 'protection', label: 'Protection' },
  { id: 'submit-warranty', label: 'Submit Warranty' },
]

const activeTab = ref<WarrantyTabId>('warranty')
const route = useRoute()

const setActiveTab = (id: WarrantyTabId | string) => {
  // Ensure the id passed is a valid WarrantyTabId, otherwise ignore or cast
  if (tabs.some(t => t.id === id)) {
    activeTab.value = id as WarrantyTabId
    if (typeof window !== 'undefined') {
      const url = new URL(window.location.href)
      url.hash = `#${id}`
      window.history.replaceState(null, '', url.toString())
    }
  }
}

const syncTabWithHash = (hash: string | null | undefined) => {
  if (!hash) return
  const clean = hash.startsWith('#') ? hash.slice(1) : hash
  if (tabs.some((tab) => tab.id === clean)) {
    activeTab.value = clean as WarrantyTabId
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
  margin-top: -1rem;
}

.support-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
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
