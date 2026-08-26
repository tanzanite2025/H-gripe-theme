<template>
  <div>
    <h2 class="products-page__title products-page__title--sr-only">Wheelset Buyers Guide</h2>

    <div class="wheelset-page">
      <!-- Buying overview -->
      <section
        v-show="activeTab === 'overview'"
        id="overview"
        class="wheelset-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Wheelset buying overview</h2>
        <div class="wheelset-buying-action">
          <div class="wheelset-buying-action__copy">
            <strong>Ready to buy after the checks?</strong>
            <p>
              QUICKBUY lets you start selecting wheelset-related products from this guide, then
              review the selected items before adding them to cart or going to payment. You can also
              check component categories, freehub fit, appearance choices, add-ons, and safety notes
              before you place the order.
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
              Start QUICKBUY
            </button>
            <button
              type="button"
              class="wheelset-guide-button"
              :aria-label="ctaNotes.wheelComponentsTab"
              :data-cta-note="ctaNotes.wheelComponentsTab"
              @click="goToWheelComponents"
            >
              Wheel components
            </button>
          </div>
        </div>
        <div class="wheelset-policy-flow">
          <article class="wheelset-guide-step">
            <div class="wheelset-guide-step__number">1</div>
            <div class="wheelset-guide-step__content">
              <h3 class="wheelset-guide-step__title">Start with the riding use</h3>
              <p class="wheelset-guide-step__body">
                A useful wheelset choice starts with the bike and riding style, not with a single
                spec. Decide whether the build is mainly for road speed, climbing, gravel comfort,
                trail control, e-bike load, or a replacement wheel that must match an existing setup.
              </p>
              <div class="wheelset-guide-panel-grid">
                <div class="wheelset-guide-panel">
                  <strong>Road / climbing</strong>
                  <p>Prioritize lower weight, predictable handling, and a rim depth that suits local wind and terrain.</p>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>Gravel / all-road</strong>
                  <p>Prioritize tire volume, rim durability, and a setup that stays comfortable on mixed surfaces.</p>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>MTB / trail</strong>
                  <p>Prioritize impact resistance, spoke count, hub standard, and serviceable replacement parts.</p>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>Replacement wheel</strong>
                  <p>Match the existing axle, brake mount, freehub, rim depth, spoke count, and tire clearance first.</p>
                </div>
              </div>
            </div>
          </article>

          <article class="wheelset-guide-step">
            <div class="wheelset-guide-step__number">2</div>
            <div class="wheelset-guide-step__content">
              <h3 class="wheelset-guide-step__title">Check the specs that can block compatibility</h3>
              <p class="wheelset-guide-step__body">
                Use these quick tools to estimate the main compatibility choices before ordering.
                The detailed pages open in a new tab, so you can keep this buying path open while
                checking the full explanation.
              </p>
              <div class="wheelset-guide-panel-grid">
                <div class="wheelset-guide-panel">
                  <strong>Tire width and rim fit</strong>
                  <p>Confirm the tire size you want can work with the rim inner width and your frame clearance. Use the rim system toggle for DT-style TSS or TC estimates.</p>
                  <div class="wheelset-overview-tool">
                    <TireRimHelper
                      hideSearchButton
                      title="Tire and rim width quick check"
                      description="Enter the tire width you plan to use and choose the rim system. This is a width quick check, not a separate tube-vs-tubeless calculator."
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
                      Tire width chart
                    </button>
                  </div>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>Tube, tubeless and tubular check</strong>
                  <p>After the width estimate, confirm the actual tire/rim system. A rim-brake clincher, tubeless-ready TC rim, hookless TSS rim, and tubular rim can use different rim-bed designs; tubeless and hookless need tire approval and pressure limits, while tubular is a separate glued tire system.</p>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.tireSystemGuideNewTab"
                      :data-cta-note="ctaNotes.tireSystemGuideNewTab"
                      @click="goToTubelessSystemGuide"
                    >
                      Tire systems
                    </button>
                  </div>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>Freehub body</strong>
                  <p>Match the freehub body to the cassette standard before choosing hubs or a complete wheelset.</p>
                  <div class="wheelset-overview-tool">
                    <FreehubGroupsetHelper
                      title="Freehub quick finder"
                      description="Select your drivetrain brand and groupset to estimate the freehub body you may need before opening the full compatibility section."
                    />
                  </div>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.freehubGuideNewTab"
                      :data-cta-note="ctaNotes.freehubGuideNewTab"
                      @click="goToFreehubGuide"
                    >
                      Freehub guide
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </article>

          <article class="wheelset-guide-step">
            <div class="wheelset-guide-step__number">3</div>
            <div class="wheelset-guide-step__content">
              <h3 class="wheelset-guide-step__title">Then choose how to configure and buy</h3>
              <p class="wheelset-guide-step__body">
                Once the core compatibility checks are clear, move into the actual order path. Use
                QUICKBUY for a fast product selection flow, or open the wheelset sections below when
                you need to confirm parts, appearance, special orders, optional add-ons, or safety
                checks before installation.
              </p>
              <div class="wheelset-guide-panel-grid">
                <div class="wheelset-guide-panel">
                  <strong>Configure and buy with QUICKBUY</strong>
                  <p>Open the QUICKBUY builder to select available wheelset-related products, review the selected items, then add them to cart or go directly to payment.</p>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.quickBuyEntry"
                      :data-cta-note="ctaNotes.quickBuyEntry"
                      @click="openQuickBuy"
                    >
                      Start QUICKBUY
                    </button>
                  </div>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>What this order path can cover</strong>
                  <p>Use the guide sections to check rims, hubs, spokes, nipples, freehub bodies, appearance options, tubeless add-ons, spare spokes, and replacement-wheel requests before buying.</p>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.wheelComponentsTab"
                      :data-cta-note="ctaNotes.wheelComponentsTab"
                      @click="goToWheelComponents"
                    >
                      Component options
                    </button>
                  </div>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>Need custom work</strong>
                  <p>Use special order for mullet setups, single front or rear wheel replacement, or mixed rim profiles that need a custom confirmation before purchase.</p>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.specialOrderTab"
                      :data-cta-note="ctaNotes.specialOrderTab"
                      @click="goToSpecialOrder"
                    >
                      Custom orders
                    </button>
                  </div>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>Safety before riding</strong>
                  <p>Before installation or the first ride, review the wheelset-specific safety checks: plastic tire levers, torque and brake-pad checks, impact inspection, spoke tension, and stop-riding warning signs.</p>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.safetyInstructionsTab"
                      :data-cta-note="ctaNotes.safetyInstructionsTab"
                      @click="goToSafetyInstructions"
                    >
                      Safety checks
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
        v-show="activeTab === 'safety-instructions'"
        id="safety-instructions"
        class="wheelset-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Safety instructions</h2>
        <WheelsetSafetyInstructionsSection
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

        <div class="wheelset-policy-flow">
          <article class="wheelset-guide-step">
            <div class="wheelset-guide-step__number">1</div>
            <div class="wheelset-guide-step__content">
              <h3 class="wheelset-guide-step__title">Mullet Wheelsets (Mixed Size)</h3>
              <p class="wheelset-guide-step__body">
                A practical setup for modern MTB builds: <strong>29" front + 27.5" rear</strong>
                or <strong>27.5" front + 26" rear</strong>. Use this when rollover, grip, and rear-end
                agility matter more than a matched wheel diameter.
              </p>

              <div class="wheelset-guide-panel-grid wheelset-guide-panel-grid--two">
                <div class="wheelset-guide-panel">
                  <strong>Front wheel</strong>
                  <p>Prioritizes rollover capability, steering grip, and high-speed stability.</p>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>Rear wheel</strong>
                  <p>Improves acceleration, bike handling, and rear-frame clearance.</p>
                </div>
              </div>

              <div class="wheelset-guide-actions">
                <button
                  type="button"
                  class="wheelset-guide-button"
                  :aria-label="ctaNotes.mulletQuickBuyEntry"
                  :data-cta-note="ctaNotes.mulletQuickBuyEntry"
                  @click="openQuickBuy"
                >
                  Mullet QUICKBUY
                </button>
              </div>
            </div>
          </article>

          <article class="wheelset-guide-step">
            <div class="wheelset-guide-step__number">2</div>
            <div class="wheelset-guide-step__content">
              <h3 class="wheelset-guide-step__title">Single Wheel Customization</h3>
              <p class="wheelset-guide-step__body">
                Use this path when you only need a front or rear replacement, want to reuse a
                compatible hub, or need one wheel matched to an existing build.
              </p>

              <div class="wheelset-guide-panel">
                <strong>Before ordering</strong>
                <p>
                  Confirm axle standard, brake mount, spoke count, freehub body, rim depth, and tire
                  clearance so the replacement wheel works with the current bike.
                </p>
              </div>

              <div class="wheelset-guide-actions">
                <button
                  type="button"
                  class="wheelset-guide-button wheelset-guide-button--solid"
                  :aria-label="ctaNotes.singleWheelChat"
                  :data-cta-note="ctaNotes.singleWheelChat"
                  @click="openWhatsAppChat"
                >
                  Chat about one wheel
                </button>
              </div>
            </div>
          </article>

          <article class="wheelset-guide-step">
            <div class="wheelset-guide-step__number">3</div>
            <div class="wheelset-guide-step__content">
              <h3 class="wheelset-guide-step__title">Mixed Rim Profiles</h3>
              <WheelsetMixedRimSection
                :openQuickBuy="openQuickBuy"
                :openWhatsAppChat="openWhatsAppChat"
              />
            </div>
          </article>
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
        <div class="wheelset-policy-flow">
          <article class="wheelset-guide-step">
            <div class="wheelset-guide-step__number">1</div>
            <div class="wheelset-guide-step__content">
              <h3 class="wheelset-guide-step__title">Useful Add-ons</h3>
              <p class="wheelset-guide-step__body">
                Optional upgrades should solve a real fit, service, or riding problem. Use the list
                below as a quick check before adding parts to a custom wheelset order.
              </p>
              <div class="wheelset-guide-panel-grid">
                <div class="wheelset-guide-panel">
                  <strong>Tubeless setup</strong>
                  <p>Rim tape, valves, and sealant for riders who want lower pressure and puncture resistance.</p>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>Spare spokes</strong>
                  <p>Helpful for travel, racing, or less common spoke lengths that may be hard to source locally.</p>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>Freehub body</strong>
                  <p>Only add a spare or alternate freehub body after confirming cassette compatibility.</p>
                  <div class="wheelset-guide-actions">
                    <button
                      type="button"
                      class="wheelset-guide-button"
                      :aria-label="ctaNotes.freehubGuideNewTab"
                      :data-cta-note="ctaNotes.freehubGuideNewTab"
                      @click="goToFreehubGuide"
                    >
                      Freehub guide
                    </button>
                  </div>
                </div>
                <div class="wheelset-guide-panel">
                  <strong>Logo finish</strong>
                  <p>Laser engraving is the fastest low-maintenance choice; decals suit custom team colors.</p>
                </div>
              </div>
            </div>
          </article>
        </div>
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
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useLocalePath, useRouter } from '#imports'
import WheelsetSafetyInstructionsSection from '~/components/WheelsetSafetyInstructionsSection.vue'
import WheelsetSampleAssemblySection from '~/components/WheelsetSampleAssemblySection.vue'
import WheelsetAppearanceLogoSection from '~/components/WheelsetAppearanceLogoSection.vue'
import WheelsetChooseFreehubSection from '~/components/WheelsetChooseFreehubSection.vue'
import WheelsetMixedRimSection from '~/components/WheelsetMixedRimSection.vue'
import { useChatWidget } from '~/composables/useChatWidget'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import SmartAccordion from '~/components/ui/SmartAccordion.vue'
import AccordionItem from '~/components/ui/AccordionItem.vue'
import TechnicalHubsSection from '~/components/TechnicalHubsSection.vue'
import TechnicalRimsSection from '~/components/TechnicalRimsSection.vue'
import TechnicalSpokesSection from '~/components/TechnicalSpokesSection.vue'
import TechnicalNipplesSection from '~/components/TechnicalNipplesSection.vue'
import TireRimHelper from '~/components/TireRimHelper.vue'
import FreehubGroupsetHelper from '~/components/FreehubGroupsetHelper.vue'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import { wheelsetBuyerTabs } from '~/utils/pageSubNavigation'


definePageMeta({
  layout: 'products',
  footerLabelKey: 'products.nav.wheelsetBuyersGuide',
  footerLabelFallback: 'Wheelset Guide',
})

useHead({
  title: 'Wheelset Buyers Guide',
})

const tabs = wheelsetBuyerTabs

const { activeTab, localizedTabPath, setActiveTab } = usePageSubNavigationTab({
  tabs,
  basePath: '/guides/wheelset-buyers',
  defaultValue: 'overview',
})
const { openChat } = useChatWidget()
const router = useRouter()
const localePath = useLocalePath()

const ctaNotes = {
  quickBuyEntry: 'Opens the bottom QUICKBUY menu; the direct-select option opens the existing wheelset product selection builder.',
  wheelComponentsTab: 'Switches /guides/wheelset-buyers to the Wheel Components tab with hubs, rims, spokes, and nipples.',
  tireWidthGuideNewTab: 'Opens /guides/tireguides/choose in a new tab for tire width, hooked TC, and hookless TSS rim chart details.',
  tireSystemGuideNewTab: 'Opens /guides/tireguides/tubeless in a new tab for tube, tubeless, hookless, and tubular tire system guidance.',
  freehubGuideNewTab: 'Opens the /guides/wheelset-buyers choose-freehub tab in a new tab for freehub compatibility guidance.',
  specialOrderTab: 'Switches /guides/wheelset-buyers to the Special Order tab for mullet wheelsets, single-wheel customization, and mixed rim profiles.',
  safetyInstructionsTab: 'Switches /guides/wheelset-buyers to the Safety Instructions tab for wheelset installation and pre-ride safety checks.',
  mulletQuickBuyEntry: 'Opens the QUICKBUY entry popover from the Mullet Wheelsets section for a mixed-size wheelset order path.',
  singleWheelChat: 'Opens the chat widget for a single front or rear wheel replacement/customization inquiry.',
} as const

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
  openLocalizedRouteInNewTab(localePath('/guides/tireguides/installation'))
}

const goToHookedHooklessGuide = () => {
  openLocalizedRouteInNewTab(localePath('/guides/tireguides/choose'))
}

const goToTubelessSystemGuide = () => {
  openLocalizedRouteInNewTab(localePath('/guides/tireguides/tubeless'))
}

const goToFreehubGuide = () => {
  openLocalizedRouteInNewTab(localizedTabPath('choose-freehub'))
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

const goToTechnicalSpokePattern = async () => {
  setActiveTab('wheel-components')
}

const goToWheelComponents = () => {
  setActiveTab('wheel-components')
}

const goToSpecialOrder = () => {
  setActiveTab('special-order')
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
.products-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: var(--tz-text-primary);
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

.wheelset-overview-tool {
  margin-top: 1rem;
}

.wheelset-guide-button,
.wheelset-page :deep(.wheelset-guide-button) {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--tz-action-primary);
  border-radius: 9999px;
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  padding: 0.6rem 1rem;
  font-size: 0.82rem;
  font-weight: 700;
  line-height: 1.25;
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
