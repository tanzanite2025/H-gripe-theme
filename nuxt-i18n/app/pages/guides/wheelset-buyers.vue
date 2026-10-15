<template>
  <div>
    <GuideCategoryChildRouteNavigationCards
      v-if="isGuideCategoryLandingPage"
      id="wheelset-guide-navigation"
      eyebrow="Wheelset Guides"
      :heading="t('guidesWheelsetBuyers.title')"
      :description="wheelsetGuidePageDescription"
      open-label="Open guide"
      :cards="wheelsetGuideNavigationCards"
    />

    <div v-else class="wheelset-page">
      <!-- Buying overview -->
      <section
        v-if="activeTab === 'overview'"
        id="overview"
        class="wheelset-section sizecharts-section"
      >
        <h1 class="sizecharts-section__title">
          {{ t('guidesWheelsetBuyersOverview.title') }}
        </h1>
        <div class="wheelset-buying-action">
          <div class="wheelset-buying-action__copy">
            <strong>{{ t('guidesWheelsetBuyersOverview.readyTitle') }}</strong>
            <p>
              {{ t('guidesWheelsetBuyersOverview.readyBody') }}
            </p>
          </div>
          <div class="wheelset-buying-action__buttons">
            <button
              type="button"
              class="wheelset-guide-button"
              :aria-label="ctaNotes.quickBuyEntry"
              :data-cta-note="ctaNotes.quickBuyEntry"
              @click="openQuickBuy"
            >
              {{ t('guidesWheelsetBuyersOverview.startQuickBuy') }}
            </button>
          </div>
        </div>
        <div class="wheelset-policy-flow">
          <article class="wheelset-guide-step">
            <div class="wheelset-guide-step__number">1</div>
            <div class="wheelset-guide-step__content">
              <h2 class="wheelset-guide-step__title">
                {{ t('guidesWheelsetBuyersOverview.step1.title') }}
              </h2>
              <p class="wheelset-guide-step__body">
                {{ t('guidesWheelsetBuyersOverview.step1.body') }}
              </p>
              <div class="wheelset-guide-panel-grid">
                <div class="wheelset-guide-panel">
                  <strong>{{ t('guidesWheelsetBuyersOverview.step1.road.title') }}</strong>
                  <p>{{ t('guidesWheelsetBuyersOverview.step1.road.body') }}</p>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>{{ t('guidesWheelsetBuyersOverview.step1.gravel.title') }}</strong>
                  <p>{{ t('guidesWheelsetBuyersOverview.step1.gravel.body') }}</p>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>{{ t('guidesWheelsetBuyersOverview.step1.mtb.title') }}</strong>
                  <p>{{ t('guidesWheelsetBuyersOverview.step1.mtb.body') }}</p>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>{{ t('guidesWheelsetBuyersOverview.step1.replacement.title') }}</strong>
                  <p>{{ t('guidesWheelsetBuyersOverview.step1.replacement.body') }}</p>
                </div>
              </div>
            </div>
          </article>

          <article class="wheelset-guide-step">
            <div class="wheelset-guide-step__number">2</div>
            <div class="wheelset-guide-step__content">
              <h2 class="wheelset-guide-step__title">
                {{ t('guidesWheelsetBuyersOverview.step2.title') }}
              </h2>
              <p class="wheelset-guide-step__body">
                {{ t('guidesWheelsetBuyersOverview.step2.body') }}
              </p>
              <div class="wheelset-guide-panel-grid">
                <div class="wheelset-guide-panel">
                  <strong>{{ t('guidesWheelsetBuyersOverview.step2.tireWidth.title') }}</strong>
                  <p>{{ t('guidesWheelsetBuyersOverview.step2.tireWidth.body') }}</p>
                  <div class="wheelset-overview-tool">
                    <TireRimHelper
                      hideSearchButton
                      :title="t('guidesWheelsetBuyersOverview.step2.tireWidth.helperTitle')"
                      :description="t('guidesWheelsetBuyersOverview.step2.tireWidth.helperDescription')"
                    />
                  </div>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.tireWidthGuideNewTab"
                      :data-cta-note="ctaNotes.tireWidthGuideNewTab"
                      @click="goToHookedHooklessGuide"
                    >
                      {{ t('guidesWheelsetBuyersOverview.step2.tireWidth.button') }}
                    </button>
                  </div>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>{{ t('guidesWheelsetBuyersOverview.step2.tireSystem.title') }}</strong>
                  <p>{{ t('guidesWheelsetBuyersOverview.step2.tireSystem.body') }}</p>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.tireSystemGuideNewTab"
                      :data-cta-note="ctaNotes.tireSystemGuideNewTab"
                      @click="goToTubelessSystemGuide"
                    >
                      {{ t('guidesWheelsetBuyersOverview.step2.tireSystem.button') }}
                    </button>
                  </div>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>{{ t('guidesWheelsetBuyersOverview.step2.freehub.title') }}</strong>
                  <p>{{ t('guidesWheelsetBuyersOverview.step2.freehub.body') }}</p>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.freehubGuideNewTab"
                      :data-cta-note="ctaNotes.freehubGuideNewTab"
                      @click="goToFreehubGuide"
                    >
                      {{ t('guidesWheelsetBuyersOverview.step2.freehub.button') }}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </article>

          <article class="wheelset-guide-step">
            <div class="wheelset-guide-step__number">3</div>
            <div class="wheelset-guide-step__content">
              <h2 class="wheelset-guide-step__title">
                {{ t('guidesWheelsetBuyersOverview.step3.title') }}
              </h2>
              <p class="wheelset-guide-step__body">
                {{ t('guidesWheelsetBuyersOverview.step3.body') }}
              </p>
              <div class="wheelset-guide-panel-grid">
                <div class="wheelset-guide-panel">
                  <strong>{{ t('guidesWheelsetBuyersOverview.step3.quickBuy.title') }}</strong>
                  <p>{{ t('guidesWheelsetBuyersOverview.step3.quickBuy.body') }}</p>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.quickBuyEntry"
                      :data-cta-note="ctaNotes.quickBuyEntry"
                      @click="openQuickBuy"
                    >
                      {{ t('guidesWheelsetBuyersOverview.step3.quickBuy.button') }}
                    </button>
                  </div>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>{{ t('guidesWheelsetBuyersOverview.step3.custom.title') }}</strong>
                  <p>{{ t('guidesWheelsetBuyersOverview.step3.custom.body') }}</p>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.sampleAssemblyTab"
                      :data-cta-note="ctaNotes.sampleAssemblyTab"
                      @click="goToSampleAssembly"
                    >
                      {{ t('guidesWheelsetBuyersOverview.step3.custom.button') }}
                    </button>
                  </div>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>{{ t('guidesWheelsetBuyersOverview.step3.safety.title') }}</strong>
                  <p>{{ t('guidesWheelsetBuyersOverview.step3.safety.body') }}</p>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.safetyInstructionsTab"
                      :data-cta-note="ctaNotes.safetyInstructionsTab"
                      @click="goToSafetyInstructions"
                    >
                      {{ t('guidesWheelsetBuyersOverview.step3.safety.button') }}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </article>
        </div>
      </section>

      <!-- Safety instructions -->
      <section
        v-if="activeTab === 'safety-instructions'"
        id="safety-instructions"
        class="wheelset-section sizecharts-section"
      >
        <h1 class="sizecharts-section__title">
          {{ t('guidesWheelsetBuyers.tabs.safetyInstructions.label') }}
        </h1>
        <WheelsetSafetyInstructionsSection
          :goToTubelessInstallation="goToTubelessInstallation"
          :goToTechnicalTension="goToTechnicalTension"
          :goToWarranty="goToWarranty"
        />
      </section>

      <!-- Sample assembly -->
      <section
        v-if="activeTab === 'sample-assembly'"
        id="sample-assembly"
        class="wheelset-section sizecharts-section wheelset-section--sample"
      >
        <h1 class="sizecharts-section__title">
          {{ t('guidesWheelsetBuyers.tabs.sampleAssembly.label') }}
        </h1>
        <WheelsetSampleAssemblySection
          :openWhatsAppChat="openWhatsAppChat"
          :openQuickBuy="openQuickBuy"
          :ctaNotes="ctaNotes"
          :goToHolePatterns="goToHolePatterns"
        />
      </section>

      <!-- Appearance Logo -->
      <section
        v-if="activeTab === 'appearance-logo'"
        id="appearance-logo"
        class="wheelset-section sizecharts-section"
      >
        <h1 class="sizecharts-section__title">
          {{ t('guidesWheelsetBuyers.tabs.appearanceLogo.label') }}
        </h1>
        <WheelsetAppearanceLogoSection :goToAboutAppearance="goToAboutAppearance" />
      </section>

      <!-- Feedback / Leave a message -->
      <section class="wheelset-feedback">
        <UserFeedbackThread
          threadKey="guides-wheelset-buyers"
          :title="t('guidesWheelsetBuyers.feedbackTitle')"
        />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useHead, useI18n, useLocalePath, useRoute, useRouter } from '#imports'
import WheelsetSafetyInstructionsSection from '~/components/WheelsetSafetyInstructionsSection.vue'
import WheelsetSampleAssemblySection from '~/components/WheelsetSampleAssemblySection.vue'
import WheelsetAppearanceLogoSection from '~/components/WheelsetAppearanceLogoSection.vue'
import { useChatWidget } from '~/composables/useChatWidget'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import TireRimHelper from '~/components/TireRimHelper.vue'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import { pageSubNavigationChildPath, wheelsetBuyerTabs } from '~/utils/pageSubNavigation'
import { usePageMessages } from '~/composables/usePageMessages'
import GuideCategoryChildRouteNavigationCards from '~/components/GuideCategoryChildRouteNavigationCards.vue'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'products.nav.wheelsetBuyersGuide',
  footerLabelFallback: 'Wheelset Guide',
})

const { locale, t } = useI18n()
const route = useRoute()
const { canonicalUrl } = useStorefrontSeoLinks()
const localePath = useLocalePath()
const tabs = wheelsetBuyerTabs
const normalizeGuideCategoryLandingRoutePath = (path: string) => path.replace(/\/+$/, '') || '/'
const isGuideCategoryLandingPage = computed(() => (
  normalizeGuideCategoryLandingRoutePath(route.path) ===
  normalizeGuideCategoryLandingRoutePath(localePath('/guides/wheelset-buyers'))
))
const getWheelsetGuideCardRoute = (tab: (typeof tabs)[number]) => {
  const explicitRoute = (tab as { to?: string }).to
  return explicitRoute || pageSubNavigationChildPath('/guides/wheelset-buyers', tab.id)
}
const wheelsetGuideNavigationCards = computed(() => tabs.map(tab => ({
  id: tab.id,
  label: t(tab.labelKey, tab.fallback),
  description: t(tab.descriptionKey, tab.description),
  to: getWheelsetGuideCardRoute(tab),
})))
const wheelsetMessages = usePageMessages('guidesWheelsetBuyers')
const overviewMessages = usePageMessages('guidesWheelsetBuyersOverview')
await Promise.all([
  wheelsetMessages.loadPageMessages(locale.value),
  overviewMessages.loadPageMessages(locale.value),
])

const { activeTab, setActiveTab } = usePageSubNavigationTab({
  tabs,
  basePath: '/guides/wheelset-buyers',
  defaultValue: 'overview',
})
const { openChat } = useChatWidget()
const router = useRouter()

const wheelsetGuidePageTitle = computed(() => {
  const guideTitle = t('guidesWheelsetBuyers.title')
  if (isGuideCategoryLandingPage.value) return guideTitle

  const activeSection = tabs.find(tab => tab.id === activeTab.value) || tabs[0]
  const sectionTitle = activeTab.value === 'overview'
    ? t('guidesWheelsetBuyersOverview.title')
    : t(activeSection.labelKey, activeSection.fallback)
  return `${sectionTitle} | ${guideTitle}`
})
const wheelsetGuidePageDescription = computed(() => t(
  isGuideCategoryLandingPage.value
    ? 'guidesWheelsetBuyers.seo.description'
    : `guidesWheelsetBuyers.seo.sections.${activeTab.value}.description`,
))

useHead(() => ({
  title: wheelsetGuidePageTitle.value,
  meta: [
    { name: 'description', content: wheelsetGuidePageDescription.value },
    { property: 'og:title', content: wheelsetGuidePageTitle.value },
    { property: 'og:description', content: wheelsetGuidePageDescription.value },
    { property: 'og:type', content: 'website' },
    { property: 'og:url', content: canonicalUrl.value },
    { name: 'twitter:card', content: 'summary' },
    { name: 'twitter:title', content: wheelsetGuidePageTitle.value },
    { name: 'twitter:description', content: wheelsetGuidePageDescription.value },
  ],
}))

const ctaNotes = computed(() => ({
  quickBuyEntry: t('guidesWheelsetBuyersOverview.ctaNotes.quickBuyEntry'),
  tireWidthGuideNewTab: t('guidesWheelsetBuyersOverview.ctaNotes.tireWidthGuideNewTab'),
  tireSystemGuideNewTab: t('guidesWheelsetBuyersOverview.ctaNotes.tireSystemGuideNewTab'),
  freehubGuideNewTab: t('guidesWheelsetBuyersOverview.ctaNotes.freehubGuideNewTab'),
  sampleAssemblyTab: t('guidesWheelsetBuyersOverview.ctaNotes.sampleAssemblyTab'),
  safetyInstructionsTab: t('guidesWheelsetBuyersOverview.ctaNotes.safetyInstructionsTab'),
  mulletQuickBuyEntry: t('guidesWheelsetBuyersOverview.ctaNotes.mulletQuickBuyEntry'),
  singleWheelChat: t('guidesWheelsetBuyersOverview.ctaNotes.singleWheelChat'),
}))

watch(locale, (nextLocale) => {
  void Promise.all([
    wheelsetMessages.loadPageMessages(nextLocale),
    overviewMessages.loadPageMessages(nextLocale),
  ])
})

const openWhatsAppChat = () => {
  openChat({ showAgentList: true })
}

const openLocalizedRouteInNewTab = (target: string) => {
  if (typeof window === 'undefined') return
  const href = router.resolve(target).href
  const opened = window.open(href, '_blank', 'noopener,noreferrer')
  if (opened) {
    opened.opener = null
  }
}

const goToTubelessInstallation = () => {
  openLocalizedRouteInNewTab(localePath('/guides/tireguides/tubeless'))
}

const goToHookedHooklessGuide = () => {
  openLocalizedRouteInNewTab(localePath('/tools/tire-rim-width-compatibility-engine'))
}

const goToTubelessSystemGuide = () => {
  openLocalizedRouteInNewTab(localePath('/guides/tireguides/tubeless'))
}

const goToFreehubGuide = () => {
  openLocalizedRouteInNewTab(localePath('/tools/freehub-cassette-fitment-engine'))
}

const goToSafetyInstructions = () => {
  setActiveTab('safety-instructions')
}

const goToWarranty = async () => {
  await router.push(localePath('/support/warranty'))
}

const goToTechnicalTension = async () => {
  await router.push(localePath('/support/test-report/tension'))
}

const goToSampleAssembly = () => {
  setActiveTab('sample-assembly')
}

const goToHolePatterns = async () => {
  await router.push(localePath('/company/about/hole-patterns'))
}

const openQuickBuy = () => {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent('quickbuy:open-entry'))
}

const goToAboutAppearance = async () => {
  await router.push(localePath('/company/about/appearance'))
}
</script>

<style src="~/assets/css/guide-sections.css"></style>

<style scoped>
#wheelset-guide-navigation {
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}

.products-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

.wheelset-page {
  --wheelset-accent: var(--tz-site-accent);
  --wheelset-accent-strong: var(--tz-site-accent-hover);
  --wheelset-accent-soft: rgba(5, 150, 105, 0.1);
  --wheelset-accent-softer: rgba(5, 150, 105, 0.06);
  --wheelset-border: rgba(203, 213, 225, 0.72);
  --wheelset-border-strong: var(--tz-border-strong);
  --wheelset-card-tint: var(--tz-surface-subtle);
  --wheelset-card-tint-strong: var(--tz-surface-muted);
  --wheelset-caution: var(--tz-status-warning-text);
  --wheelset-warning: var(--tz-status-warning-text);
  --wheelset-critical: var(--tz-status-danger-text);
  --tz-form-panel-surface: var(--tz-surface-muted);
  --tz-text-accent: var(--wheelset-accent);
  margin: 0 auto;
  width: 100%;
  max-width: none;
}

.wheelset-section :deep(.sizecharts-section__title),
.wheelset-page .sizecharts-section__title {
  margin: 0 0 2rem;
  color: var(--tz-text-primary);
  font-weight: 650;
  text-align: center;
}

.wheelset-policy-flow,
.wheelset-page :deep(.wheelset-policy-flow) {
  display: grid;
  gap: 2rem;
}

.wheelset-buying-action {
  display: grid;
  grid-template-columns: minmax(0, 1.25fr) minmax(16rem, 0.75fr);
  gap: 1rem;
  align-items: center;
  margin: 0 0 2rem;
  border-radius: 0.5rem;
  background: var(--tz-card-surface);
  padding: 1rem;
}

.wheelset-buying-action__copy {
  min-width: 0;
}

.wheelset-buying-action__copy strong {
  display: inline-flex;
  width: fit-content;
  max-width: 100%;
  margin-bottom: 0.65rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 9999px;
  color: var(--tz-text-primary);
  padding: 0.28rem 0.65rem;
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.35;
}

.wheelset-buying-action__copy p {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.86rem;
  line-height: 1.7;
  text-align: left;
}

.wheelset-buying-action__buttons {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  gap: 0.65rem;
  justify-content: flex-end;
}

.wheelset-guide-step,
.wheelset-page :deep(.wheelset-guide-step) {
  position: relative;
  display: block;
}

.wheelset-guide-step__number,
.wheelset-page :deep(.wheelset-guide-step__number) {
  position: absolute;
  top: 0;
  left: 0;
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--tz-border-strong);
  border-radius: 999px;
  background: var(--tz-surface-muted);
  color: var(--tz-text-secondary);
  font-size: 0.82rem;
  font-weight: 700;
  line-height: 1;
  box-shadow: 0 8px 22px rgba(0, 0, 0, 0.35);
}

.wheelset-guide-step__content,
.wheelset-page :deep(.wheelset-guide-step__content) {
  width: 100%;
  min-width: 0;
  padding-top: 0;
}

.wheelset-guide-step__title,
.wheelset-page :deep(.wheelset-guide-step__title) {
  display: flex;
  min-height: 2rem;
  align-items: center;
  margin: 0 0 0.55rem;
  padding-left: 3rem;
  color: var(--tz-text-primary);
  font-size: 1.05rem;
  font-weight: 700;
  line-height: 1.35;
}

.wheelset-guide-step__body,
.wheelset-page :deep(.wheelset-guide-step__body) {
  margin: 0 0 1rem;
  color: var(--tz-text-secondary);
  font-size: 0.9rem;
  line-height: 1.7;
  text-align: left;
}

.wheelset-guide-step__body strong,
.wheelset-page :deep(.wheelset-guide-step__body strong),
.wheelset-page :deep(.wheelset-guide-panel strong) {
  color: var(--tz-text-primary);
  font-weight: 650;
}

.wheelset-guide-panel-grid,
.wheelset-page :deep(.wheelset-guide-panel-grid) {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.85rem;
  margin-top: 1rem;
}

.wheelset-guide-panel,
.wheelset-page :deep(.wheelset-guide-panel) {
  border: none;
  border-radius: 0.5rem;
  background: var(--tz-card-surface);
  padding: 1rem;
  box-shadow: none;
}

.wheelset-page :deep(.wheelset-guide-panel--notice) {
  border: 1px solid color-mix(in srgb, var(--tz-status-warning-text) 35%, transparent);
  background: var(--tz-status-warning-bg);
}

.wheelset-guide-panel strong,
.wheelset-page :deep(.wheelset-guide-panel strong) {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  max-width: 100%;
  margin-bottom: 0.65rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 9999px;
  background: transparent;
  color: var(--tz-text-primary);
  padding: 0.28rem 0.65rem;
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.35;
  white-space: normal;
}

.wheelset-page :deep(.wheelset-guide-panel--notice strong) {
  color: var(--tz-status-warning-text);
}

.wheelset-guide-panel p,
.wheelset-page :deep(.wheelset-guide-panel p),
.wheelset-page :deep(.wheelset-guide-panel li) {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.65;
  text-align: left;
}

.wheelset-guide-panel ul,
.wheelset-page :deep(.wheelset-guide-panel ul) {
  margin: 0.55rem 0 0;
  padding-left: 1rem;
}

.wheelset-guide-actions,
.wheelset-page :deep(.wheelset-guide-actions) {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  margin-top: 1rem;
}

.wheelset-guide-button,
.wheelset-page :deep(.wheelset-guide-button) {
  display: inline-flex;
  min-height: var(--tz-pill-button-min-height);
  align-items: center;
  justify-content: center;
  border: 1px solid var(--tz-action-primary);
  border-radius: 9999px;
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  padding: var(--tz-pill-button-padding-block) 1rem;
  font-size: 0.82rem;
  font-weight: 700;
  line-height: var(--tz-pill-button-line-height);
  letter-spacing: 0;
  text-transform: none;
  white-space: normal;
  transition: background-color 0.16s ease, border-color 0.16s ease, color 0.16s ease;
}

.wheelset-guide-button:hover,
.wheelset-page :deep(.wheelset-guide-button:hover) {
  border-color: var(--tz-action-primary-hover);
  background: var(--tz-action-primary-hover);
  color: #ffffff;
}

.wheelset-guide-button--solid,
.wheelset-page :deep(.wheelset-guide-button--solid) {
  border-color: var(--tz-action-primary);
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
}

.wheelset-guide-note,
.wheelset-page :deep(.wheelset-guide-note) {
  display: flex;
  gap: 0.55rem;
  align-items: flex-start;
  margin-top: 0.8rem;
  border: none;
  padding-left: 0;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.55;
  text-align: left;
}

.wheelset-page :deep(.wheelset-guide-note svg) {
  flex: 0 0 auto;
  margin-top: 0.15rem;
}

.wheelset-page :deep(.wheelset-guide-subsection) {
  display: grid;
  gap: 1rem;
}

/* Page-level tab entry points are rendered by the header/mobile mega menu. */
.wheelset-section--sample > p {
  color: var(--tz-text-secondary);
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
  color: var(--wheelset-accent);
  text-decoration: underline;
}

.wheelset-link:hover {
  color: var(--wheelset-accent-strong);
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
  border: 1px solid var(--tz-action-primary);
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 8px 18px rgba(5, 150, 105, 0.16);
  transition: background-color 0.15s ease, transform 0.08s ease;
}

.wheelset-inline-button:hover {
  border-color: var(--tz-action-primary-hover);
  background: var(--tz-action-primary-hover);
}

.wheelset-inline-button:active {
  transform: scale(0.98);
}

.wheelset-safety-footer {
  margin-top: 0.85rem;
}

/* Make Mixed rim CTA text readable on dark background */
.wheelset-section .guide-section__cta-wrapper {
  color: var(--tz-text-secondary);
}

.wheelset-special-card {
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.08);
}

@media (min-width: 768px) {
  .wheelset-section {
    margin-top: 1rem;
  }
}

@media (max-width: 768px) {
  .wheelset-policy-flow,
  .wheelset-page :deep(.wheelset-policy-flow) {
    gap: 1.5rem;
  }

  .wheelset-buying-action {
    grid-template-columns: 1fr;
  }

  .wheelset-buying-action__buttons {
    justify-content: stretch;
  }

  .wheelset-guide-step,
  .wheelset-page :deep(.wheelset-guide-step) {
    display: block;
  }

  .wheelset-guide-step__number,
  .wheelset-page :deep(.wheelset-guide-step__number) {
    width: 1.75rem;
    height: 1.75rem;
  }

  .wheelset-guide-step__title,
  .wheelset-page :deep(.wheelset-guide-step__title) {
    min-height: 1.75rem;
    padding-left: 2.5rem;
  }

  .wheelset-guide-panel-grid,
  .wheelset-page :deep(.wheelset-guide-panel-grid) {
    grid-template-columns: 1fr;
  }

  .wheelset-guide-button,
  .wheelset-page :deep(.wheelset-guide-button) {
    width: 100%;
  }

  .wheelset-tabs {
    justify-content: flex-start;
  }
}
</style>
