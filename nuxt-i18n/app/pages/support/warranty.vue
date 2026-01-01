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

    <!-- 1. Change / Cancel -->
    <section
      v-show="activeTab === 'change-cancel'"
      id="change-cancel"
      class="support-section"
    >
      <h3 class="support-section__title text-center">Change / Cancel</h3>
      <p class="support-section__body text-center">
        Content placeholder for Change / Cancel.
      </p>
    </section>

    <!-- 2. Damaged or Lost Goods -->
    <section
      v-show="activeTab === 'damaged-lost'"
      id="damaged-lost"
      class="support-section"
    >
      <h3 class="support-section__title text-center">Damaged or Lost Goods</h3>
      <p class="support-section__body text-center">
        Content placeholder for Damaged or Lost Goods.
      </p>
    </section>

    <!-- 3. Returns -->
    <section
      v-show="activeTab === 'returns'"
      id="returns"
      class="support-section"
    >
      <h3 class="support-section__title text-center">Returns</h3>
      <p class="support-section__body text-center">
        Content placeholder for Returns.
      </p>
    </section>

    <!-- 4. Warranty -->
    <section
      v-show="activeTab === 'warranty'"
      id="warranty"
      class="support-section"
    >
      <h3 class="support-section__title text-center">Warranty</h3>
      <p class="support-section__body text-center">
        Content placeholder for Warranty terms and coverage.
      </p>
    </section>

    <!-- 5. Accidental Damage -->
    <section
      v-show="activeTab === 'accidental-damage'"
      id="accidental-damage"
      class="support-section"
    >
      <h3 class="support-section__title text-center">Accidental Damage</h3>
      <p class="support-section__body text-center">
        Content placeholder for Accidental Damage policy.
      </p>
    </section>

    <!-- 6. Protection -->
    <section
      v-show="activeTab === 'protection'"
      id="protection"
      class="support-section"
    >
      <h3 class="support-section__title text-center">Protection</h3>
      <p class="support-section__body text-center">
        Content placeholder for Protection plans.
      </p>
    </section>

    <!-- 7. Submit Warranty -->
    <section
      v-show="activeTab === 'submit-warranty'"
      id="submit-warranty"
      class="support-section"
    >
      <h3 class="support-section__title text-center">Submit Warranty</h3>
      <p class="support-section__body text-center">
        Content placeholder for Submit Warranty form.
      </p>
    </section>

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

const activeTab = ref<WarrantyTabId>('warranty') // Default to Warranty tab as it's the main topic? Or 'change-cancel' as first? 
// User listed tabs in order: Change/Cancel... but page title is Warranty.
// Usually default is first tab unless deep linked. I'll set default to 'change-cancel' (first tab) or 'warranty' if contextually appropriate.
// User listed "Change / Cancel" first. I will default to 'change-cancel' to match list order.
// But wait, the previous page was "Warranty". It might be confusing if "Warranty" tab is hidden.
// I'll stick to the first tab in the list: 'change-cancel'.

const route = useRoute()

const setActiveTab = (id: WarrantyTabId) => {
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
  /* margin-top: -1rem; similar to test-report if needed, but let's see */
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
  border: 0;
}

.support-section {
  margin-top: 2rem;
  /* text-align: center; Optional, keeping typical flow */
}

.support-section__title {
  margin: 0 0 0.5rem;
  font-size: 1.25rem;
  font-weight: 600;
  color: #e5e7eb;
}

.support-section__body {
  margin: 0 auto;
  max-width: 65ch; /* Readable line length */
  font-size: 0.95rem;
  line-height: 1.6;
  color: rgba(148, 163, 184, 0.9);
}
</style>
