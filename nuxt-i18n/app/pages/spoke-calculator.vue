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

      <section v-show="activeTab === 'parameter'" class="spoke-parameter">
        <h3 class="spoke-parameter__title">{{ $t('spokeCalculator.parameter.title', 'Parameter definitions') }}</h3>
        <div class="spoke-parameter__content">
          <p>{{ $t('spokeCalculator.parameter.intro', "Use these definitions to double-check your rim and hub data before calculating. Small measurement differences can change the spoke length.") }}</p>
          <ul>
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
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import SpokeCalculatorCore from '~/components/SpokeCalculatorCore.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import SpokeHistorySearch from '~/components/SpokeHistorySearch.vue'
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
 }

 .spoke-parameter__title {
   margin: 0 0 0.35rem;
   font-size: 1rem;
   font-weight: 600;
   color: #e5e7eb;
 }

 .spoke-parameter__content {
   font-size: 0.88rem;
   color: rgba(148, 163, 184, 0.9);
 }

 .spoke-parameter__content p {
   margin: 0 0 0.6rem;
 }

 .spoke-parameter__content ul {
   margin: 0;
   padding-left: 1.1rem;
   list-style-type: disc;
 }

 .spoke-parameter__content li + li {
   margin-top: 0.35rem;
 }

 .spoke-parameter__note {
   margin-top: 0.75rem;
 }
</style>
