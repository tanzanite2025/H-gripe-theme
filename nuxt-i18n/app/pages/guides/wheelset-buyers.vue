<template>
  <div>
    <h2 class="products-page__title products-page__title--sr-only">Wheelset Buyers Guide</h2>

    <div class="wheelset-page">
      <PageTabBar
        :tabs="tabs"
        :active-id="activeTab"
        aria-label="Wheelset buyer guide sections"
        @select="setActiveTab"
      />

      <!-- Safety instructions -->
      <section
        v-show="activeTab === 'safety-instructions'"
        id="safety-instructions"
        class="wheelset-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Safety instructions</h2>
        <WheelsetSafetyInstructionsSection
          :openWhatsAppChat="openWhatsAppChat"
          :goToTubelessInstallation="goToTubelessInstallation"
          :goToTechnicalTension="goToTechnicalTension"
          :goToWarranty="goToWarranty"

        />
      </section>

      <!-- Sample assembly -->
      <section
        v-show="activeTab === 'sample-assembly'"
        id="sample-assembly"
        class="wheelset-section sizecharts-section wheelset-section--sample"
      >
        <h2 class="sizecharts-section__title">Sample assembly</h2>
        <WheelsetSampleAssemblySection
          :openWhatsAppChat="openWhatsAppChat"
          :goToTechnicalSpokePattern="goToTechnicalSpokePattern"
          :goToHolePatterns="goToHolePatterns"
        />
      </section>

      <!-- Special order (Mullet, Custom Front & Rear, Mixed rim) -->
      <section
        v-show="activeTab === 'special-order'"
        id="special-order"
        class="wheelset-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Special order</h2>

        <div class="space-y-6">
           <!-- Mullet wheelsets (Amber) -->
           <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
               <div class="flex items-center gap-3 mb-4 pb-3 border-b border-amber-500/10">
          <h3 class="text-lg font-bold tz-text-primary">Mullet Wheelsets (Mixed Size)</h3>
               </div>
               
          <p class="text-sm tz-text-secondary leading-relaxed mb-4">
                 The perfect setup for modern MTB riding: <strong>29" Front + 27.5" Rear</strong> (or 27.5" + 26"). 
                 This "Mullet" configuration offers the best of both worlds:
               </p>
               
               <div class="grid grid-cols-2 gap-4 mb-4">
          <div class="bg-amber-500/5 rounded-lg p-3 text-xs tz-text-secondary shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
                     <strong class="block text-amber-500 uppercase tracking-wider mb-1">Front Wheel</strong>
                     Provides rollover capability, grip, and stability.
                  </div>
          <div class="bg-amber-500/5 rounded-lg p-3 text-xs tz-text-secondary shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
                     <strong class="block text-amber-500 uppercase tracking-wider mb-1">Rear Wheel</strong>
                     Delivers agility, acceleration, and clearance.
                  </div>
               </div>

               <button
                  type="button"
                  class="w-full md:w-auto px-6 py-2 rounded-full border border-amber-500/30 text-amber-500 text-xs font-bold uppercase tracking-wider hover:bg-amber-500/10 transition-colors"
                  @click="openQuickBuy"
                >
                  Customize Mullet Setup
                </button>
           </div>

           <!-- Custom Front & Rear Wheels (Sky) -->
           <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
               <div class="flex items-center gap-3 mb-4 pb-3 border-b border-sky-500/10">
          <h3 class="text-lg font-bold tz-text-primary">Single Wheel Customization</h3>
               </div>
               
          <p class="text-sm tz-text-secondary leading-relaxed mb-4">
                 Need just a front or rear replacement? We specialize in single-wheel builds tailored to your specific needs.
                 While our website showcases complete sets, we fully support individual custom orders.
               </p>

               <div class="flex flex-wrap gap-3">
                  <a class="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-sky-500/30 text-sky-400 text-xs font-bold uppercase tracking-wider hover:bg-sky-500/10 transition-colors" href="mailto:support@tanzanite.site">
                    Email Support
                  </a>
                   <button
                    type="button"
                    class="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-slate-700 hover:bg-slate-600 text-white text-xs font-bold uppercase tracking-wider transition-colors shadow-lg"
                    @click="openWhatsAppChat"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2-2z"/></svg>
                    Chat for Single Wheel
                  </button>
               </div>
           </div>

           <!-- Mixed rim subsection -->
           <div class="mt-8">
              <WheelsetMixedRimSection
                :openQuickBuy="openQuickBuy"
                :openWhatsAppChat="openWhatsAppChat"
              />
           </div>

        </div>
      </section>

      <!-- Appearance Logo -->
      <section
        v-show="activeTab === 'appearance-logo'"
        id="appearance-logo"
        class="wheelset-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Appearance Logo</h2>
        <WheelsetAppearanceLogoSection :goToAboutAppearance="goToAboutAppearance" />
      </section>

      <!-- Choose freehub -->
      <section
        v-show="activeTab === 'choose-freehub'"
        id="choose-freehub"
        class="wheelset-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Choose freehub</h2>
        <WheelsetChooseFreehubSection />
      </section>

      <!-- Wheel Components -->
      <section
        v-show="activeTab === 'wheel-components'"
        id="wheel-components"
        class="wheelset-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Wheel Components</h2>
        <SmartAccordion default-id="hubs">
          <AccordionItem id="hubs" title="1. Hubs">
             <TechnicalHubsSection />
          </AccordionItem>
          
          <AccordionItem id="rims" title="2. Rims">
             <TechnicalRimsSection />
          </AccordionItem>

          <AccordionItem id="spokes" title="3. Spokes">
             <TechnicalSpokesSection />
          </AccordionItem>
          
          <AccordionItem id="nipples" title="4. Nipples">
             <TechnicalNipplesSection />
          </AccordionItem>
        </SmartAccordion>
      </section>

      <!-- Optional -->
      <section
        v-show="activeTab === 'optional'"
        id="optional"
        class="wheelset-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Optional</h2>
        <p class="sizecharts-section__intro">Here is some new content.</p>
      </section>

      <!-- FAQ Section - 放在所有 tab 内容之后 -->
      <section class="wheelset-section wheelset-faq">
        <PageFaq
          page-id="guides-wheelset-buyers"
          theme="dark"
          :show-categories="true"
        />
      </section>

      <!-- Feedback / Leave a message -->
      <section class="wheelset-feedback">
        <UserFeedbackThread
          threadKey="guides-wheelset-buyers"
          title="Share your feedback or leave a message about the Wheelset Buyers Guide"
        />
      </section>
    </div>
  </div>
  <QuickBuyModal v-if="quickOpen" :config="null" @close="quickOpen = false" />
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useLocalePath, useRouter, useRoute } from '#imports'
import PageFaq from '~/components/PageFaq.vue'
import WheelsetSafetyInstructionsSection from '~/components/WheelsetSafetyInstructionsSection.vue'
import WheelsetSampleAssemblySection from '~/components/WheelsetSampleAssemblySection.vue'
import WheelsetAppearanceLogoSection from '~/components/WheelsetAppearanceLogoSection.vue'
import WheelsetChooseFreehubSection from '~/components/WheelsetChooseFreehubSection.vue'
import WheelsetMixedRimSection from '~/components/WheelsetMixedRimSection.vue'
import { useChatWidget } from '~/composables/useChatWidget'
import QuickBuyModal from '@/components/QuickBuy.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import SmartAccordion from '~/components/ui/SmartAccordion.vue'
import AccordionItem from '~/components/ui/AccordionItem.vue'
import TechnicalHubsSection from '~/components/TechnicalHubsSection.vue'
import TechnicalRimsSection from '~/components/TechnicalRimsSection.vue'
import TechnicalSpokesSection from '~/components/TechnicalSpokesSection.vue'
import TechnicalNipplesSection from '~/components/TechnicalNipplesSection.vue'


definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Wheelset Buyers Guide',
})

type WheelsetTabId =
  | 'safety-instructions'
  | 'sample-assembly'
  | 'special-order'
  | 'appearance-logo'
  | 'choose-freehub'
  | 'wheel-components'
  | 'optional'

const tabs: { id: WheelsetTabId; label: string }[] = [
  { id: 'safety-instructions', label: 'Safety instructions' },
  { id: 'sample-assembly', label: 'Sample assembly' },
  { id: 'special-order', label: 'Special order' },
  { id: 'appearance-logo', label: 'Appearance Logo' },
  { id: 'choose-freehub', label: 'Choose freehub' },
  { id: 'wheel-components', label: 'Wheel Components' },
  { id: 'optional', label: 'Optional' },
]

const activeTab = ref<WheelsetTabId>('safety-instructions')
const quickOpen = ref(false)

const setActiveTab = (id: WheelsetTabId | string) => {
  if (!tabs.some((tab) => tab.id === id)) return
  const next = id as WheelsetTabId
  activeTab.value = next
  if (typeof window !== 'undefined') {
    const url = new URL(window.location.href)
    url.hash = `#${next}`
    window.history.replaceState(null, '', url.toString())
  }
}

const { openChat } = useChatWidget()
const router = useRouter()
const route = useRoute()
const localePath = useLocalePath()

const syncTabWithHash = (hash: string | null | undefined) => {
  if (!hash) return
  const clean = hash.startsWith('#') ? hash.slice(1) : hash
  if (tabs.some((tab) => tab.id === clean)) {
    activeTab.value = clean as WheelsetTabId
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

const openWhatsAppChat = () => {
  openChat({ showAgentList: true })
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'whatsapp-chat' } }))
  }
}

const goToTubelessInstallation = async () => {
  await router.push(`${localePath('/guides/tireguides')}#installation`)
}

const goToWarranty = async () => {
  await router.push(localePath('/support/warranty'))
}

const goToTechnicalTension = async () => {
  await router.push(localePath('/support/test-report') + '#tension')
}

const goToTechnicalSpokePattern = async () => {
  setActiveTab('wheel-components')
}

const goToHolePatterns = async () => {
  await router.push(`${localePath('/company/about')}#hole-patterns`)
}

const openQuickBuy = () => {
  quickOpen.value = true
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'wheelset-quick' } }))
  }
}

const goToAboutAppearance = async () => {
  await router.push(`${localePath('/company/about')}#appearance`)
}
</script>

<style scoped>
.products-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: #f9fafb;
}

.products-page__title--sr-only {
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

.products-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

.wheelset-page {
  margin: 0.25rem auto 0;
  width: 100%;
  max-width: none;
}

/* Page-level tabs are handled by PageTabBar. */
.wheelset-section--sample > p {
  color: #f9fafb;
  text-align: center;
}

.wheelset-safety-item {
  margin: 0.25rem 0 0;
}

.wheelset-section .sizecharts-section__list ul {
  margin-top: 0.15rem;
  padding-left: 1.1rem;
  list-style-type: none;
}

.wheelset-appearance-images {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 0.75rem;
  margin-top: 0.75rem;
}

.wheelset-link {
  color: #2dd4bf;
  text-decoration: underline;
}

.wheelset-link:hover {
  color: #60a5fa;
}

.wheelset-inline-note {
  margin-left: 0.25rem;
  color: var(--tz-text-secondary);
}

.wheelset-inline-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem 0.65rem;
  margin: 0 0.25rem;
  border-radius: 9999px;
  border: none;
  background: rgba(31, 41, 55, 0.65);
  color: #ffffff;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  box-shadow:
    0 10px 22px rgba(0, 0, 0, 0.85);
  transition: background-color 0.15s ease, transform 0.08s ease;
}

.wheelset-inline-button:hover {
  background: rgba(51, 65, 85, 0.75);
}

.wheelset-inline-button:active {
  transform: scale(0.98);
}

.wheelset-safety-footer {
  margin-top: 0.85rem;
}

/* Make Mixed rim CTA text readable on dark background */
.wheelset-section .guide-section__cta-wrapper {
  color: #e5e7eb;
}

.wheelset-special-card {
  box-shadow:
    4px 6px 18px rgba(0, 0, 0, 0.9);
}

@media (min-width: 768px) {
  .wheelset-section {
    margin-top: 1rem;
  }
}

@media (max-width: 768px) {
  .wheelset-tabs {
    justify-content: flex-start;
  }
}
</style>
