<template>
  <div>
    <h1 class="sr-only">Spoke Calculator</h1>

    <div class="spoke-page">
      <div class="wheelset-tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="wheelset-tabs__item"
          :class="{ 'wheelset-tabs__item--active': activeTab === tab.id }"
          @click="setActiveTab(tab.id)"
        >
          {{ $t(tab.labelKey, tab.fallback) }}
        </button>
      </div>

      <section v-show="activeTab === 'calculator'">
        <div class="support-page__calculator-wrapper">
          <SpokeCalculatorCore />

          <div class="mt-8">
            <SpokeHistorySearch />
          </div>
        </div>

        <div class="mt-10">
          <UserFeedbackThread
            threadKey="products-spoke-calculator"
            title="Share your feedback about the Spoke Calculator"
          />
        </div>
      </section>

      <section v-show="activeTab === 'parameter'" class="spoke-parameter sizecharts-section">
        <h3 class="spoke-parameter__title">{{ $t('spokeCalculator.parameter.title', 'Parameter definitions') }}</h3>
        <div class="spoke-parameter__content">
          <p>{{ $t('spokeCalculator.parameter.intro', "Use these definitions to double-check your rim and hub data before calculating. Small measurement differences can change the spoke length.") }}</p>
          <ul class="sizecharts-section__list">
            <li>
              <strong>{{ $t('spokeCalculator.parameter.items.erd.title', 'ERD (Effective Rim Diameter)') }}</strong>: {{ $t('spokeCalculator.parameter.items.erd.desc', "The diameter at the spoke nipple seats inside the rim. Use the rim brand's ERD spec or measure it with two nipples and a caliper.") }}
            </li>
            <li>
              <strong>{{ $t('spokeCalculator.parameter.items.flangeDiameter.title', 'Flange diameter') }}</strong>: {{ $t('spokeCalculator.parameter.items.flangeDiameter.desc', 'The circle diameter through the spoke hole centers on the hub flange (left and right can be different).') }}
            </li>
            <li>
              <strong>{{ $t('spokeCalculator.parameter.items.centerToFlange.title', 'Center-to-flange') }}</strong>: {{ $t('spokeCalculator.parameter.items.centerToFlange.desc', 'The distance from the hub centerline to each flange (left and right). This affects dish and spoke length asymmetry.') }}
            </li>
            <li>
              <strong>{{ $t('spokeCalculator.parameter.items.holeCount.title', 'Spoke hole count') }}</strong>: {{ $t('spokeCalculator.parameter.items.holeCount.desc', 'Must match both rim and hub (e.g. 24/28/32). Make sure you select the same count for front and rear.') }}
            </li>
            <li>
              <strong>{{ $t('spokeCalculator.parameter.items.crossPattern.title', 'Cross pattern') }}</strong>: {{ $t('spokeCalculator.parameter.items.crossPattern.desc', 'How many times each spoke crosses other spokes (2x/3x/4x). Higher cross typically increases spoke length.') }}
            </li>
          </ul>
          <p class="spoke-parameter__note">{{ $t('spokeCalculator.parameter.note', "Tip: If you are unsure, use the manufacturer's published specs. If you measure yourself, measure twice and enter values to the same unit (mm).") }}</p>

          <!-- Detailed spoke length workflow moved from /guides/technical -->
          <h4 class="spoke-parameter__subtitle">Spoke length workflow</h4>

          <div class="spoke-parameter__image spoke-parameter__image--spoke-length">
            <GuideImage
              src="/public/technical/spoke-length.webp"
              alt="Overview illustration for calculating bicycle spoke length"
              :zoomOnClick="true"
              caption="Overview illustration showing which rim and hub measurements are needed to calculate bicycle spoke length."
            />
          </div>

          <ul>
            <li>
              <strong class="spoke-parameter__heading">Use a spoke length calculator</strong>
              <ul>
                <li>
                  Enter your specific rim and hub measurements into an online spoke length calculator to get precise results based on
                  established formulas.
                </li>
                <li>
                  TANZANITE provides an official spoke length calculator, with internal data for thousands of rims and hubs to
                  simplify input and avoid mistakes.
                </li>
              </ul>
            </li>
            <li>
              <strong class="spoke-parameter__heading">Required measurement data</strong>
              <ul>
                <li>
                  <strong>Effective Rim Diameter (ERD)</strong> &mdash; the diameter between the spoke junctions on opposite sides inside
                  the rim, where the spoke ends and nipples seat. This is a crucial measurement.
                </li>
                <li>
                  <strong>Flange Diameter</strong> &mdash; the diameter between the centers of the spoke holes on each hub flange. Left
                  and right sides may be different and should be measured separately.
                </li>
                <li>
                  <strong>Center to Flange</strong> &mdash; the horizontal distance from the hub centerline to the center of each
                  flange, measured separately for left and right.
                </li>
                <li>
                  <strong>Lacing pattern / cross count</strong> &mdash; how many times each spoke crosses other spokes (0-cross /
                  straight pull, 1-cross, 2-cross, 3-cross, etc.). 3-cross is the most common standard pattern.
                </li>
              </ul>
              <div class="spoke-parameter__image spoke-parameter__image--erd">
                <GuideImage
                  src="/public/technical/what-is-erd.webp"
                  alt="Diagram showing how Effective Rim Diameter (ERD) is measured"
                  :zoomOnClick="true"
                  caption="Diagram showing how Effective Rim Diameter (ERD) is measured using two spokes, nipples and a caliper."
                />
              </div>
            </li>
            <li>
              <strong class="spoke-parameter__heading">Measurement method (general)</strong>
              <ul>
                <li>
                  You can measure with a tape measure or calipers, but for best accuracy it is recommended to follow the dedicated ERD
                  and hub measurement steps below.
                </li>
              </ul>
            </li>
            <li>
              <strong class="spoke-parameter__heading">Measure ERD</strong>
              <ul>
                <li>
                  Prepare two old spokes of known length, two nipples, and a caliper.
                </li>
                <li>
                  Insert the two spokes through opposite holes in the rim.
                </li>
                <li>
                  Screw on the nipples until the spoke ends just reach the bottom of the nipple groove &mdash; the ideal final position
                  when the wheel is built.
                </li>
                <li>
                  Measure the distance between the inside of the J-bends of the two spokes.
                </li>
                <li>
                  Compute <strong>ERD = spoke&nbsp;1 length + spoke&nbsp;2 length + measured distance</strong>.
                </li>
                <li>
                  This method compensates for manufacturing tolerances between rims from different factories, materials, and wall
                  thicknesses, which can easily change ERD by 2&nbsp;mm or more and strongly affect the correct spoke length.
                </li>
              </ul>
            </li>
            <li>
              <strong class="spoke-parameter__heading">Measure hub size</strong>
              <ul>
                <li>
                  <strong>Flange Diameter</strong> &mdash; use calipers to measure the distance between opposing spoke hole centers on
                  the same flange.
                </li>
                <li>
                  <strong>Center to Flange distance</strong> &mdash; remove the hub axle if needed, or use a straight edge (for
                  example, the edge of a table) to help measure from the flange to the hub centerline.
                </li>
              </ul>
            </li>
            <li>
              <strong class="spoke-parameter__heading">Use an online calculator</strong>
              <ul>
                <li>
                  Once all measurements are collected, enter them into an online calculator.
                </li>
                <li>
                  Besides our own tool, many major spoke manufacturers (such as DT Swiss or Sapim) also provide free online spoke
                  length calculators.
                </li>
                <li>
                  Input your measurements, number of spokes, and desired lacing pattern; the calculator will output spoke lengths for the
                  left and right sides.
                </li>
                <li>
                  Spokes are usually sold in whole millimeters, so round to the nearest available length. Being about 1&nbsp;mm longer
                  is usually safer than 1&nbsp;mm shorter.
                </li>
                <li>
                  Following these steps will help you obtain reliable spoke lengths for a custom wheelset build.
                </li>
              </ul>
            </li>
          </ul>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import SpokeCalculatorCore from '~/components/SpokeCalculatorCore.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import SpokeHistorySearch from '~/components/SpokeHistorySearch.vue'
import GuideImage from '~/components/GuideImage.vue'
import { ref } from 'vue'

type SpokeCalculatorTabId = 'calculator' | 'parameter'

const tabs: { id: SpokeCalculatorTabId; labelKey: string; fallback: string }[] = [
  { id: 'calculator', labelKey: 'spokeCalculator.tabs.calculator', fallback: 'Calculator' },
  { id: 'parameter', labelKey: 'spokeCalculator.tabs.parameter', fallback: 'Parameter' },
]

const activeTab = ref<SpokeCalculatorTabId>('calculator')

const setActiveTab = (id: SpokeCalculatorTabId) => {
  activeTab.value = id
}

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Spoke Calculator',
})
</script>

<style scoped>
.support-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.support-page__calculator-wrapper {
  margin-top: 1.5rem;
}

 .spoke-page {
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

 @media (max-width: 768px) {
   .wheelset-tabs {
     justify-content: flex-start;
   }
 }

 .spoke-parameter {
   margin-top: 0.75rem;
   text-align: center;
 }

 .spoke-parameter__title {
   margin: 0 0 0.35rem;
   font-size: 1rem;
   font-weight: 600;
   color: #e5e7eb;
   text-align: center;
 }

 .spoke-parameter__content {
   font-size: 0.88rem;
   color: rgba(148, 163, 184, 0.9);
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
  color: #38bdf8;
  text-align: center;
}

.spoke-parameter__heading {
  color: #38bdf8;
}

.spoke-parameter__image {
  margin: 0.75rem 0 1rem;
}

/* On this page we do not want leading bullets before items; keep guides blue dots only in /guides */
.spoke-parameter .sizecharts-section__list > li {
  padding-left: 0;
}

.spoke-parameter .sizecharts-section__list > li::before {
  content: none;
}
</style>
