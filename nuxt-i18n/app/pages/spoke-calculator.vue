<template>
  <div>
    <h1 class="sr-only">Spoke Calculator</h1>

    <div class="spoke-page">
      <section v-show="activeTab === 'calculator'">
        <div class="support-page__calculator-wrapper">
          <SpokeCalculatorCore />

          <div class="spoke-smart-search-section mt-16 pt-10">
             <div class="text-center mb-8">
               <h3 class="spoke-smart-search-section__title">Smart Search</h3>
                <p class="text-sm tz-text-secondary mt-2">Instantly find spoke lengths for verified official builds</p>
             </div>
             <SpokeSmartSearch />
          </div>


        </div>

        <div class="mt-10">
          <UserFeedbackThread
            threadKey="products-spoke-calculator"
            title="Share your feedback about the Spoke Calculator"
          />
        </div>
      </section>

      <section
        v-show="activeTab === 'parameter'"
        class="spoke-parameter sizecharts-section rounded-2xl p-6 bg-[var(--tz-card-surface)] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)]"
      >
        <h3 class="spoke-parameter__title text-lg font-bold tz-text-primary mb-2">{{ $t('spokeCalculator.parameter.title', 'Parameter definitions') }}</h3>
        
        <div class="spoke-parameter__content text-left">
          <p class="tz-text-secondary text-sm mb-6 text-center max-w-2xl mx-auto">
            {{ $t('spokeCalculator.parameter.intro', "Use these definitions to double-check your rim and hub data before calculating. Small measurement differences can change the spoke length.") }}
          </p>

          <!-- Definitions Grid -->
          <div class="grid gap-4 md:grid-cols-2 mb-8">
            <div class="spoke-parameter__definition-card">
              <h4 class="spoke-parameter__definition-title">
                {{ $t('spokeCalculator.parameter.items.erd.title', 'ERD (Effective Rim Diameter)') }}
              </h4>
              <p class="text-sm tz-text-secondary leading-relaxed">
                {{ $t('spokeCalculator.parameter.items.erd.desc', "The diameter at the spoke nipple seats inside the rim. Use the rim brand's ERD spec or measure it with two nipples and a caliper.") }}
              </p>
            </div>
            
            <div class="spoke-parameter__definition-card">
              <h4 class="spoke-parameter__definition-title">
                {{ $t('spokeCalculator.parameter.items.flangeDiameter.title', 'Flange diameter') }}
              </h4>
              <p class="text-sm tz-text-secondary leading-relaxed">
                {{ $t('spokeCalculator.parameter.items.flangeDiameter.desc', 'The circle diameter through the spoke hole centers on the hub flange (left and right can be different).') }}
              </p>
            </div>

            <div class="spoke-parameter__definition-card">
              <h4 class="spoke-parameter__definition-title">
                {{ $t('spokeCalculator.parameter.items.centerToFlange.title', 'Center-to-flange') }}
              </h4>
              <p class="text-sm tz-text-secondary leading-relaxed">
                {{ $t('spokeCalculator.parameter.items.centerToFlange.desc', 'The distance from the hub centerline to each flange (left and right). This affects dish and spoke length asymmetry.') }}
              </p>
            </div>

            <div class="spoke-parameter__definition-card">
              <h4 class="spoke-parameter__definition-title">
                {{ $t('spokeCalculator.parameter.items.holeCount.title', 'Spoke hole count') }}
              </h4>
              <p class="text-sm tz-text-secondary leading-relaxed">
                {{ $t('spokeCalculator.parameter.items.holeCount.desc', 'Must match both rim and hub (e.g. 24/28/32). Make sure you select the same count for front and rear.') }}
              </p>
            </div>

            <div class="spoke-parameter__definition-card md:col-span-2">
              <h4 class="spoke-parameter__definition-title">
                {{ $t('spokeCalculator.parameter.items.crossPattern.title', 'Cross pattern') }}
              </h4>
              <p class="text-sm tz-text-secondary leading-relaxed">
                {{ $t('spokeCalculator.parameter.items.crossPattern.desc', 'How many times each spoke crosses other spokes (2x/3x/4x). Higher cross typically increases spoke length.') }}
              </p>
            </div>
          </div>

          <!-- Note Alert -->
           <div class="spoke-parameter__note-card">
             <span class="text-lg">💡</span>
             <p>{{ $t('spokeCalculator.parameter.note', "Tip: If you are unsure, use the manufacturer's published specs. If you measure yourself, measure twice and enter values to the same unit (mm).") }}</p>
          </div>

          <!-- Workflow Section -->
          <div class="spoke-parameter__workflow">
            <h4 class="spoke-parameter__workflow-title">Spoke length workflow</h4>

             <div class="spoke-parameter__workflow-visual">
              <GuideImage
                src="/public/technical/spoke-length.webp"
                alt="Overview illustration for calculating bicycle spoke length"
                :zoomOnClick="true"
                caption="Overview illustration showing which rim and hub measurements are needed to calculate bicycle spoke length."
              />
            </div>

            <div class="space-y-8">
              <!-- Step 1: Measure ERD -->
              <div class="spoke-parameter__step-card">
                <h5 class="tz-text-primary font-bold mb-3 flex items-center gap-2">
                  <span class="spoke-parameter__step-badge">1</span>
                  Measure ERD (Effective Rim Diameter)
                </h5>
                
                <div class="grid md:grid-cols-2 gap-6 items-start">
                    <div class="text-sm tz-text-secondary space-y-2 leading-relaxed">
                      <p>Compute <strong>ERD = spoke 1 length + spoke 2 length + measured distance</strong>.</p>
                       <ul class="list-disc list-inside space-y-1 ml-1 tz-text-muted">
                        <li>Prepare two old spokes of known length, two nipples, and a caliper.</li>
                        <li>Insert spokes through opposite holes in the rim.</li>
                        <li>Screw nipples until flush with the nipple groove bottom (ideal final position).</li>
                        <li>Measure distance between the J-bends.</li>
                      </ul>
                       <p class="text-xs italic mt-2 tz-text-muted">
                        * This method compensates for rim manufacturing tolerances.
                      </p>
                   </div>
                   <div class="spoke-parameter__step-illustration">
                      <GuideImage
                        src="/public/technical/what-is-erd.webp"
                        alt="Diagram showing how Effective Rim Diameter (ERD) is measured"
                        :zoomOnClick="true"
                        caption="Diagram showing how Effective Rim Diameter (ERD) is measured using two spokes, nipples and a caliper."
                      />
                   </div>
                </div>
              </div>

               <!-- Step 2: Measure Hub -->
              <div class="spoke-parameter__step-card">
                <h5 class="tz-text-primary font-bold mb-3 flex items-center gap-2">
                  <span class="spoke-parameter__step-badge">2</span>
                  Measure Hub Dimensions
                </h5>
                 <div class="text-sm tz-text-secondary space-y-3 leading-relaxed">
                  <div>
                     <strong class="tz-text-primary">Flange Diameter:</strong> Use calipers to measure the distance between opposing spoke hole centers on the same flange.
                  </div>
                  <div>
                     <strong class="tz-text-primary">Center to Flange:</strong> Remove the hub axle if needed. Measure from the flange center to the hub centerline (or locknut face and subtract).
                  </div>
                </div>
              </div>

               <!-- Step 3: Calculation -->
              <div class="spoke-parameter__step-card">
                <h5 class="tz-text-primary font-bold mb-3 flex items-center gap-2">
                  <span class="spoke-parameter__step-badge">3</span>
                  Calculate & Round
                </h5>
                 <div class="text-sm tz-text-secondary space-y-2 leading-relaxed">
                  <p>
                    Enter measurements into the calculator. If the result is between standard sizes (e.g. 288.4mm), you usually round to the nearest available 1mm increment.
                  </p>
                  <p>
                     <strong class="tz-text-primary">Tip:</strong> Being 1mm longer is generally safer than 1mm shorter to ensure full thread engagement.
                  </p>
                </div>
              </div>

            </div>
          </div>
        </div>
      </section>

    </div>
  </div>
</template>

<script setup lang="ts">
import SpokeCalculatorCore from '~/components/SpokeCalculatorCore.vue'
import SpokeSmartSearch from '~/components/SpokeSmartSearch.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

import GuideImage from '~/components/GuideImage.vue'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import { spokeCalculatorTabs } from '~/utils/pageSubNavigation'

const { activeTab } = usePageSubNavigationTab({
  tabs: spokeCalculatorTabs,
  basePath: '/spoke-calculator',
  defaultValue: 'calculator',
})

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Spoke Calculator',
})
</script>

<style src="~/assets/css/guide-sections.css"></style>

<style scoped>
.support-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: #f9fafb;
}

.support-page__calculator-wrapper {
  margin-top: 1.5rem;
}

 .spoke-page {
   margin: 0 auto;
   width: 100%;
   max-width: none;
 }

/* Page-level tab entry points are rendered by the header/mobile mega menu. */

 .spoke-parameter {
   margin-top: 0.75rem;
   text-align: center;
 }

 .spoke-parameter__title {
    margin: 0 0 0.35rem;
    font-size: var(--tz-type-section-title);
    line-height: 1.35;
    font-weight: 600;
     color: var(--tz-text-primary);
    text-align: center;
 }

 .spoke-parameter__content {
   font-size: 0.88rem;
   color: var(--tz-text-secondary);
   text-align: center;
 }

 .spoke-parameter__content p {
   margin: 0 0 0.6rem;
 }

 .spoke-parameter__note {
   margin-top: 0.75rem;
 }

.spoke-parameter__subtitle {
  margin: 1rem 0 0.4rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--tz-text-accent);
  text-align: center;
}

.spoke-parameter__heading {
  color: var(--tz-text-accent);
}

.spoke-parameter__definition-card,
.spoke-parameter__step-card {
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.75rem;
  background: var(--tz-form-panel-surface);
}

.spoke-parameter__definition-card {
  padding: 1rem;
  transition: border-color 0.18s ease, background-color 0.18s ease;
}

.spoke-parameter__definition-card:hover {
  border-color: rgba(181, 255, 109, 0.42);
  background: var(--tz-card-surface);
}

.spoke-parameter__definition-title {
  margin: 0 0 0.5rem;
  color: var(--tz-text-accent);
  font-size: 0.875rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  line-height: 1.35;
  text-transform: uppercase;
}

.spoke-parameter__note-card {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 2.5rem;
  border: 1px solid rgba(181, 255, 109, 0.24);
  border-radius: 0.5rem;
  background: rgba(181, 255, 109, 0.06);
  color: var(--tz-text-secondary);
  font-size: 0.875rem;
  padding: 1rem;
}

.spoke-parameter__workflow {
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  padding-top: 2rem;
}

.spoke-parameter__workflow-title {
  margin: 0 0 1.5rem;
  color: var(--tz-text-accent);
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  line-height: 1.35;
  text-align: center;
  text-transform: uppercase;
}

.spoke-parameter__workflow-visual,
.spoke-parameter__step-illustration {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: var(--tz-form-panel-surface);
}

.spoke-parameter__workflow-visual {
  margin-bottom: 2rem;
  border-radius: 0.75rem;
  box-shadow: 0 1rem 1.5rem -1rem rgba(0, 0, 0, 0.9);
}

.spoke-parameter__step-card {
  padding: 1.25rem;
}

.spoke-parameter__step-badge {
  display: inline-flex;
  width: 1.5rem;
  height: 1.5rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(181, 255, 109, 0.32);
  border-radius: 50%;
  background: rgba(181, 255, 109, 0.08);
  color: var(--tz-text-accent);
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1;
}

.spoke-parameter__step-illustration {
  border-radius: 0.5rem;
}

.spoke-parameter__image {
  margin: 0.75rem 0 1rem;
}

.spoke-smart-search-section {
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.spoke-smart-search-section__title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: var(--tz-type-section-title);
  line-height: 1.35;
  font-weight: 600;
}

/* On this page we do not want leading bullets before items. */
.spoke-parameter .sizecharts-section__list > li {
  padding-left: 0;
}

.spoke-parameter .sizecharts-section__list > li::before {
  content: none;
}
</style>
