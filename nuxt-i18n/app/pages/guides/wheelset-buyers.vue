<template>
  <div>
    <h2 class="products-page__title products-page__title--sr-only">Wheelset Buyers Guide</h2>

    <div class="wheelset-page">
      <div class="wheelset-tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="wheelset-tabs__item"
          :class="{ 'wheelset-tabs__item--active': activeTab === tab.id }"
          @click="setActiveTab(tab.id)"
        >
          {{ tab.label }}
        </button>
      </div>

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
          :goToAfterSales="goToAfterSales"
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

        <!-- Mullet wheelsets subsection -->
        <div class="mt-4 rounded-xl bg-slate-800/70 p-4 space-y-2 wheelset-special-card">
          <h3 class="sizecharts-section__subheading text-sky-300 font-semibold">
            Mullet wheelsets
          </h3>
          <p class="sizecharts-section__intro">
            Many mountain bikers have fallen in love with the combination of a 29-inch front wheel paired with a 27.5-inch rear
            wheel, or a 27.5-inch front wheel with a 26-inch rear wheel. This wheel size mix is particularly well-suited for
            off-road riding. It offers easy handling on steep descents, agile and responsive cornering, and the ability to roll
            smoothly over rough terrain. This setup is commonly known as the "mullet", and for good reason: the front wheel
            provides stability and reliability, while the rear wheel delivers agility and performance!
          </p>
          <p class="guide-section__cta-wrapper">
            <button
              type="button"
              class="wheelset-inline-button"
              @click="openQuickBuy"
            >
              You can use our quick customization to match
            </button>
          </p>
        </div>

        <!-- Custom Front & Rear Wheels subsection -->
        <div class="mt-4 rounded-xl bg-slate-800/70 p-4 space-y-2 wheelset-special-card">
          <h3 class="sizecharts-section__subheading text-sky-300 font-semibold">
            Custom Front &amp; Rear Wheels
          </h3>
          <p class="sizecharts-section__intro">
            We specialize in customized services, including building individual wheels tailored to your specific needs. While we only
            showcase complete wheelsets online, single wheels can be custom-made upon request. Please contact us at
            <a class="wheelset-link" href="mailto:support@tanzanite.site">support@tanzanite.site</a>
            for special orders or
            <button
              type="button"
              class="wheelset-inline-button"
              @click="openWhatsAppChat"
            >
              use our faster Instant Chat option
            </button>
            !
          </p>
        </div>

        <!-- Mixed rim subsection -->
        <div class="mt-4 rounded-xl bg-slate-800/70 p-4 space-y-2 wheelset-special-card">
          <h3 class="sizecharts-section__subheading text-sky-300 font-semibold">
            Mixed rim
          </h3>
          <WheelsetMixedRimSection
            :openQuickBuy="openQuickBuy"
            :openWhatsAppChat="openWhatsAppChat"
          />
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
        <p class="sizecharts-section__intro">
          Detailed guidance on rims, hubs, spokes, and other wheel components will be added here.
        </p>
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

const setActiveTab = (id: WheelsetTabId) => {
  activeTab.value = id
  if (typeof window !== 'undefined') {
    const url = new URL(window.location.href)
    url.hash = `#${id}`
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

const goToAfterSales = async () => {
  await router.push(localePath('/support/after-sales'))
}

const goToTechnicalTension = async () => {
  await router.push(`${localePath('/guides/technical')}#tension`)
}

const goToTechnicalSpokePattern = async () => {
  await router.push(`${localePath('/guides/technical')}#spoke-pattern`)
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
  font-size: 1.5rem;
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
  color: rgba(148, 163, 184, 0.9);
}

.wheelset-page {
  margin: 0.25rem auto 0;
  max-width: 900px;
}

.wheelset-tabs {
  display: flex;
  overflow-x: auto;
  gap: 12px;
  padding: 4px 16px;
  margin: 0 -16px 1rem;
  max-width: calc(100% + 32px);
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  touch-action: pan-x;
}

.wheelset-tabs::-webkit-scrollbar {
  display: none;
}

.wheelset-tabs__item {
  flex-shrink: 0;
  border: none;
  border-radius: 9999px;
  padding: 8px 18px;
  font-size: 0.85rem;
  font-weight: 500;
  color: #ffffff;
  background: rgba(31, 41, 55, 0.9);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  backdrop-filter: blur(4px);
  box-shadow:
    0 3px 9px -6px rgba(0, 0, 0, 0.9),
    0 0 9px rgba(0, 0, 0, 0.85);
}

.wheelset-tabs__item:active {
  transform: scale(0.96);
}

.wheelset-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.wheelset-tabs__item--active {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #000000;
  border: none;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}

@media (min-width: 768px) {
  .wheelset-tabs {
    flex-wrap: wrap;
    justify-content: center;
    margin: 0 0 1rem;
    padding: 4px 0;
    max-width: 100%;
  }
}


.wheelset-section {
  margin-top: 0.75rem;
}

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
  color: rgba(148, 163, 184, 0.9);
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
