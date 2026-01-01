<template>
  <div>
    <h1 class="sr-only">Spoke Calculator</h1>

    <div class="spoke-page">
      <div class="nav-pill-tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="nav-pill-item"
          :class="{ 'nav-pill-item--active': activeTab === tab.id }"
          @click="setActiveTab(tab.id)"
        >
          {{ $t(tab.labelKey, tab.fallback) }}
        </button>
      </div>

      <section v-show="activeTab === 'calculator'">
        <div class="support-page__calculator-wrapper">
          <SpokeCalculatorCore />

          <div class="mt-16 pt-10 border-t border-slate-800/50">
             <div class="text-center mb-8">
               <h3 class="text-xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-sky-400 to-emerald-400">Smart Search</h3>
               <p class="text-sm text-slate-400 mt-2">Instantly find spoke lengths for verified official builds</p>
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
        class="spoke-parameter sizecharts-section rounded-2xl p-6 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_10px_26px_-14px_rgba(0,0,0,0.95)]"
      >
        <h3 class="spoke-parameter__title text-lg font-bold text-slate-200 mb-2">{{ $t('spokeCalculator.parameter.title', 'Parameter definitions') }}</h3>
        
        <div class="spoke-parameter__content text-left">
          <p class="text-slate-400 text-sm mb-6 text-center max-w-2xl mx-auto">
            {{ $t('spokeCalculator.parameter.intro', "Use these definitions to double-check your rim and hub data before calculating. Small measurement differences can change the spoke length.") }}
          </p>

          <!-- Definitions Grid -->
          <div class="grid gap-4 md:grid-cols-2 mb-8">
            <div class="bg-slate-900/50 rounded-xl p-4 border border-slate-800/50 hover:border-slate-700/50 transition-colors">
              <h4 class="text-sky-400 font-semibold mb-2 text-sm uppercase tracking-wide">
                {{ $t('spokeCalculator.parameter.items.erd.title', 'ERD (Effective Rim Diameter)') }}
              </h4>
              <p class="text-sm text-slate-400 leading-relaxed">
                {{ $t('spokeCalculator.parameter.items.erd.desc', "The diameter at the spoke nipple seats inside the rim. Use the rim brand's ERD spec or measure it with two nipples and a caliper.") }}
              </p>
            </div>
            
            <div class="bg-slate-900/50 rounded-xl p-4 border border-slate-800/50 hover:border-slate-700/50 transition-colors">
              <h4 class="text-sky-400 font-semibold mb-2 text-sm uppercase tracking-wide">
                {{ $t('spokeCalculator.parameter.items.flangeDiameter.title', 'Flange diameter') }}
              </h4>
              <p class="text-sm text-slate-400 leading-relaxed">
                {{ $t('spokeCalculator.parameter.items.flangeDiameter.desc', 'The circle diameter through the spoke hole centers on the hub flange (left and right can be different).') }}
              </p>
            </div>

            <div class="bg-slate-900/50 rounded-xl p-4 border border-slate-800/50 hover:border-slate-700/50 transition-colors">
              <h4 class="text-sky-400 font-semibold mb-2 text-sm uppercase tracking-wide">
                {{ $t('spokeCalculator.parameter.items.centerToFlange.title', 'Center-to-flange') }}
              </h4>
              <p class="text-sm text-slate-400 leading-relaxed">
                {{ $t('spokeCalculator.parameter.items.centerToFlange.desc', 'The distance from the hub centerline to each flange (left and right). This affects dish and spoke length asymmetry.') }}
              </p>
            </div>

            <div class="bg-slate-900/50 rounded-xl p-4 border border-slate-800/50 hover:border-slate-700/50 transition-colors">
              <h4 class="text-sky-400 font-semibold mb-2 text-sm uppercase tracking-wide">
                {{ $t('spokeCalculator.parameter.items.holeCount.title', 'Spoke hole count') }}
              </h4>
              <p class="text-sm text-slate-400 leading-relaxed">
                {{ $t('spokeCalculator.parameter.items.holeCount.desc', 'Must match both rim and hub (e.g. 24/28/32). Make sure you select the same count for front and rear.') }}
              </p>
            </div>

            <div class="bg-slate-900/50 rounded-xl p-4 border border-slate-800/50 hover:border-slate-700/50 transition-colors md:col-span-2">
              <h4 class="text-sky-400 font-semibold mb-2 text-sm uppercase tracking-wide">
                {{ $t('spokeCalculator.parameter.items.crossPattern.title', 'Cross pattern') }}
              </h4>
              <p class="text-sm text-slate-400 leading-relaxed">
                {{ $t('spokeCalculator.parameter.items.crossPattern.desc', 'How many times each spoke crosses other spokes (2x/3x/4x). Higher cross typically increases spoke length.') }}
              </p>
            </div>
          </div>

          <!-- Note Alert -->
          <div class="bg-slate-800/40 rounded-lg p-4 mb-10 flex gap-3 text-sm text-slate-300 border border-slate-700/30">
             <span class="text-lg">💡</span>
             <p>{{ $t('spokeCalculator.parameter.note', "Tip: If you are unsure, use the manufacturer's published specs. If you measure yourself, measure twice and enter values to the same unit (mm).") }}</p>
          </div>

          <!-- Workflow Section -->
          <div class="border-t border-slate-800/50 pt-8">
            <h4 class="text-emerald-400 font-bold mb-6 text-center text-base uppercase tracking-wider">Spoke length workflow</h4>

            <div class="mb-8 rounded-xl overflow-hidden shadow-lg border border-slate-800/50 bg-slate-950">
              <GuideImage
                src="/public/technical/spoke-length.webp"
                alt="Overview illustration for calculating bicycle spoke length"
                :zoomOnClick="true"
                caption="Overview illustration showing which rim and hub measurements are needed to calculate bicycle spoke length."
              />
            </div>

            <div class="space-y-8">
              <!-- Step 1: Measure ERD -->
              <div class="bg-slate-900/30 rounded-xl p-5 border border-slate-800/30">
                <h5 class="text-slate-200 font-bold mb-3 flex items-center gap-2">
                  <span class="bg-sky-500/10 text-sky-400 w-6 h-6 rounded-full flex items-center justify-center text-xs border border-sky-500/20">1</span>
                  Measure ERD (Effective Rim Diameter)
                </h5>
                
                <div class="grid md:grid-cols-2 gap-6 items-start">
                   <div class="text-sm text-slate-400 space-y-2 leading-relaxed">
                      <p>Compute <strong>ERD = spoke 1 length + spoke 2 length + measured distance</strong>.</p>
                      <ul class="list-disc list-inside space-y-1 ml-1 text-slate-500">
                        <li>Prepare two old spokes of known length, two nipples, and a caliper.</li>
                        <li>Insert spokes through opposite holes in the rim.</li>
                        <li>Screw nipples until flush with the nipple groove bottom (ideal final position).</li>
                        <li>Measure distance between the J-bends.</li>
                      </ul>
                      <p class="text-xs italic mt-2 text-slate-500">
                        * This method compensates for rim manufacturing tolerances.
                      </p>
                   </div>
                   <div class="rounded-lg overflow-hidden border border-slate-800/50">
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
              <div class="bg-slate-900/30 rounded-xl p-5 border border-slate-800/30">
                <h5 class="text-slate-200 font-bold mb-3 flex items-center gap-2">
                  <span class="bg-sky-500/10 text-sky-400 w-6 h-6 rounded-full flex items-center justify-center text-xs border border-sky-500/20">2</span>
                  Measure Hub Dimensions
                </h5>
                <div class="text-sm text-slate-400 space-y-3 leading-relaxed">
                  <div>
                    <strong class="text-slate-300">Flange Diameter:</strong> Use calipers to measure the distance between opposing spoke hole centers on the same flange.
                  </div>
                  <div>
                    <strong class="text-slate-300">Center to Flange:</strong> Remove the hub axle if needed. Measure from the flange center to the hub centerline (or locknut face and subtract).
                  </div>
                </div>
              </div>

               <!-- Step 3: Calculation -->
              <div class="bg-slate-900/30 rounded-xl p-5 border border-slate-800/30">
                <h5 class="text-slate-200 font-bold mb-3 flex items-center gap-2">
                  <span class="bg-sky-500/10 text-sky-400 w-6 h-6 rounded-full flex items-center justify-center text-xs border border-sky-500/20">3</span>
                  Calculate & Round
                </h5>
                <div class="text-sm text-slate-400 space-y-2 leading-relaxed">
                  <p>
                    Enter measurements into the calculator. If the result is between standard sizes (e.g. 288.4mm), you usually round to the nearest available 1mm increment.
                  </p>
                  <p>
                    <strong class="text-slate-300">Tip:</strong> Being 1mm longer is generally safer than 1mm shorter to ensure full thread engagement.
                  </p>
                </div>
              </div>

            </div>
          </div>
        </div>
      </section>

      <!-- FAQ Section -->
      <section class="spoke-faq mt-8">
        <PageFaq
          page-id="products-spoke-calculator"
          theme="dark"
          :show-categories="true"
        />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import SpokeCalculatorCore from '~/components/SpokeCalculatorCore.vue'
import SpokeSmartSearch from '~/components/SpokeSmartSearch.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

import PageFaq from '~/components/PageFaq.vue'
import GuideImage from '~/components/GuideImage.vue'
import { ref, watch } from 'vue'
import { useRoute } from '#imports'

type SpokeCalculatorTabId = 'calculator' | 'parameter'

const tabs: { id: SpokeCalculatorTabId; labelKey: string; fallback: string }[] = [
  { id: 'calculator', labelKey: 'spokeCalculator.tabs.calculator', fallback: 'Calculator' },
  { id: 'parameter', labelKey: 'spokeCalculator.tabs.parameter', fallback: 'Parameter' },
]

const activeTab = ref<SpokeCalculatorTabId>('calculator')
const route = useRoute()

const getTabFromHash = (hash: string | null | undefined): SpokeCalculatorTabId | null => {
  if (!hash) return null
  const raw = hash.startsWith('#') ? hash.slice(1) : hash
  const allowed: SpokeCalculatorTabId[] = ['calculator', 'parameter']
  return (allowed as string[]).includes(raw) ? (raw as SpokeCalculatorTabId) : null
}

watch(
  () => route.hash,
  (hash) => {
    const next = getTabFromHash(hash)
    if (next) activeTab.value = next
  },
  { immediate: true }
)

const setActiveTab = (id: SpokeCalculatorTabId) => {
  activeTab.value = id
  if (typeof window !== 'undefined') {
    const url = new URL(window.location.href)
    url.hash = `#${id}`
    window.history.replaceState(null, '', url.toString())
  }
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

/* .wheelset-tabs styles removed in favor of global .nav-pill-tabs */

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
