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
      <h3 class="support-section__title text-center mb-8">Change / Cancel</h3>
      
      <div class="max-w-3xl mx-auto text-left space-y-8">
        <ul class="space-y-6">
          <li class="flex gap-4 items-start">
              <span class="shrink-0 w-1.5 h-1.5 rounded-full bg-teal-400 mt-2 shadow-[0_0_8px_rgba(45,212,191,0.6)]"></span>
              <div class="text-slate-300 text-sm leading-relaxed">
                  <strong class="text-slate-100 block mb-1.5 text-base">Scope of Application</strong>
                  Customers may request to cancel all or part of an order before the item has shipped, regardless of whether payment has been partially or fully completed.
              </div>
          </li>
          <li class="flex gap-4 items-start">
              <span class="shrink-0 w-1.5 h-1.5 rounded-full bg-teal-400 mt-2 shadow-[0_0_8px_rgba(45,212,191,0.6)]"></span>
              <div class="text-slate-300 text-sm leading-relaxed">
                  <strong class="text-slate-100 block mb-1.5 text-base">Cancellation After Shipment</strong>
                  If the product has already shipped, any cancellation request will be treated as a "return" and subject to the return process and applicable fees. Please refer to our 
                  <a href="#" @click.prevent="setActiveTab('returns')" class="text-teal-400 hover:text-teal-300 transition-colors border-b border-teal-400/30 hover:border-teal-300 pb-0.5">Product Return Policy</a> 
                  for details.
              </div>
          </li>
          <li class="flex gap-4 items-start">
              <span class="shrink-0 w-1.5 h-1.5 rounded-full bg-teal-400 mt-2 shadow-[0_0_8px_rgba(45,212,191,0.6)]"></span>
              <div class="text-slate-300 text-sm leading-relaxed">
                  <strong class="text-slate-100 block mb-1.5 text-base">Authorization Requirement</strong>
                  Any order amendment or cancellation must be confirmed and authorized by TANZANITE.
              </div>
          </li>
          <li class="flex gap-4 items-start">
              <span class="shrink-0 w-1.5 h-1.5 rounded-full bg-teal-400 mt-2 shadow-[0_0_8px_rgba(45,212,191,0.6)]"></span>
              <div class="text-slate-300 text-sm leading-relaxed">
                  <strong class="text-slate-100 block mb-1.5 text-base">Fee Policy</strong>
                  Once authorized, a cancellation may incur a charge of <strong class="text-white bg-white/5 px-1.5 py-0.5 rounded text-xs font-bold border border-white/10">10% of the order value (minimum USD $50)</strong> to cover re-stocking and operational costs.
              </div>
          </li>
        </ul>

        <!-- Note Box -->
        <div class="relative pl-5 py-3 border-l-2 border-slate-600 bg-slate-800/20 rounded-r-lg">
            <p class="text-slate-400 text-xs italic leading-relaxed">
              <strong class="text-slate-300 not-italic uppercase tracking-wider text-[10px] mr-1">Note:</strong> 
              Please carefully review product specifications and quantities before placing an order to avoid unnecessary fees and procedures.
            </p>
        </div>
      </div>
    </section>

    <!-- 2. Damaged or Lost Goods -->
    <section
      v-show="activeTab === 'damaged-lost'"
      id="damaged-lost"
      class="support-section"
    >
      <h3 class="support-section__title text-center mb-8">Damaged or Lost Goods</h3>

      <div class="max-w-3xl mx-auto text-left space-y-8">
        <!-- Intro Alert -->
        <div class="bg-indigo-500/10 border-l-4 border-indigo-500 p-4 rounded-r-lg">
          <p class="text-indigo-200 font-medium">Please carefully inspect the package before signing for it.</p>
        </div>

        <div class="space-y-10">
          <!-- 1. Normal Acceptance -->
          <div class="relative pl-12">
            <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">1</div>
            <h4 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Normal Acceptance</h4>
            <ul class="space-y-3 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
              <li>If the package is intact, please sign for it.</li>
              <li>If you notice damage, note it on the delivery slip or refuse the package. The goods will be returned to Tanzanite.</li>
              <li>
                If you need to file a claim with the carrier, they will issue a claim number. Please provide this number to Tanzanite so we can track the shipment. Once the returned goods are received, we will arrange re-shipment within <strong class="text-white">10 business days</strong>.
              </li>
            </ul>
          </div>

          <!-- 2. Damaged but Accepted -->
          <div class="relative pl-12">
            <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">2</div>
            <h4 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Damaged but Accepted</h4>
            <!-- Warning Box for this critical section -->
            <div class="bg-rose-500/5 rounded-lg p-4 border border-rose-500/10">
                <ul class="space-y-3 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-rose-500">
                  <li>If the goods are damaged upon delivery but you still accept them without reporting, the carrier cannot be held responsible.</li>
                  <li>In this case, you will be responsible for purchasing any replacement products.</li>
                </ul>
            </div>
          </div>

          <!-- 3. Damage During Transit -->
          <div class="relative pl-12">
            <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">3</div>
            <h4 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Damage During Transit</h4>
            <ul class="space-y-3 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
              <li>If the package appears intact but the product is seriously damaged during transit, please keep the waybill, packaging, and related materials.</li>
              <li>Contact us within <strong class="text-white">7 days of receipt</strong> at <a href="mailto:support@tanzanite.site" class="text-teal-400 hover:text-teal-300 transition-colors">support@tanzanite.site</a>.</li>
            </ul>
          </div>

          <!-- 4. Lost Shipment -->
          <div class="relative pl-12">
            <div class="absolute left-0 top-0 w-8 h-8 rounded-full bg-slate-800 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm shadow-lg shadow-black/50">4</div>
            <h4 class="text-slate-100 font-bold text-lg mb-3 pt-0.5">Lost Shipment</h4>
            <ul class="space-y-3 text-slate-400 text-sm leading-relaxed list-disc pl-4 marker:text-teal-500/60">
              <li>If the goods are lost during transit, please contact us immediately at <a href="mailto:support@tanzanite.site" class="text-teal-400 hover:text-teal-300 transition-colors">support@tanzanite.site</a>.</li>
              <li>All Tanzanite shipments are insured, and we will assist you with the claim process.</li>
            </ul>
          </div>
        </div>
      </div>
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
