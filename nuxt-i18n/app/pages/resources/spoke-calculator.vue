<template>
  <div>
    <h1 class="sr-only">{{ t('resourcesSpokeCalculator.title') }}</h1>

    <div class="spoke-page">
      <section v-show="activeTab === 'calculator'">
        <div class="support-page__calculator-wrapper">
          <SpokeCalculatorBlueprint />

          <div class="spoke-smart-search-section mt-16 pt-10">
             <div class="text-center mb-8">
               <h3 class="spoke-smart-search-section__title">{{ t('resourcesSpokeCalculator.search.title') }}</h3>
                <p class="text-sm tz-text-secondary mt-2">{{ t('resourcesSpokeCalculator.search.intro') }}</p>
             </div>
             <SpokeSmartSearch />
          </div>


        </div>

        <div class="mt-10">
          <UserFeedbackThread
            threadKey="products-spoke-calculator"
            :title="t('resourcesSpokeCalculator.feedbackTitle')"
          />
        </div>
      </section>

      <section
        v-show="activeTab === 'parameter'"
        class="spoke-parameter sizecharts-section rounded-2xl p-6 bg-[var(--tz-card-surface)] shadow-md"
      >
        <h3 class="spoke-parameter__title text-lg font-bold tz-text-primary mb-2">{{ t('resourcesSpokeCalculator.parameter.title') }}</h3>
        
        <div class="spoke-parameter__content text-left">
          <p class="tz-text-secondary text-sm mb-6 text-center max-w-2xl mx-auto">
            {{ t('resourcesSpokeCalculator.parameter.intro') }}
          </p>

          <!-- Definitions Grid -->
          <div class="grid gap-4 md:grid-cols-2 mb-8">
            <div class="spoke-parameter__definition-card">
              <h4 class="spoke-parameter__definition-title">
                {{ t('resourcesSpokeCalculator.parameter.items.erd.title') }}
              </h4>
              <p class="text-sm tz-text-secondary leading-relaxed">
                {{ t('resourcesSpokeCalculator.parameter.items.erd.desc') }}
              </p>
            </div>
            
            <div class="spoke-parameter__definition-card">
              <h4 class="spoke-parameter__definition-title">
                {{ t('resourcesSpokeCalculator.parameter.items.flangeDiameter.title') }}
              </h4>
              <p class="text-sm tz-text-secondary leading-relaxed">
                {{ t('resourcesSpokeCalculator.parameter.items.flangeDiameter.desc') }}
              </p>
            </div>

            <div class="spoke-parameter__definition-card">
              <h4 class="spoke-parameter__definition-title">
                {{ t('resourcesSpokeCalculator.parameter.items.centerToFlange.title') }}
              </h4>
              <p class="text-sm tz-text-secondary leading-relaxed">
                {{ t('resourcesSpokeCalculator.parameter.items.centerToFlange.desc') }}
              </p>
            </div>

            <div class="spoke-parameter__definition-card">
              <h4 class="spoke-parameter__definition-title">
                {{ t('resourcesSpokeCalculator.parameter.items.holeCount.title') }}
              </h4>
              <p class="text-sm tz-text-secondary leading-relaxed">
                {{ t('resourcesSpokeCalculator.parameter.items.holeCount.desc') }}
              </p>
            </div>

            <div class="spoke-parameter__definition-card md:col-span-2">
              <h4 class="spoke-parameter__definition-title">
                {{ t('resourcesSpokeCalculator.parameter.items.crossPattern.title') }}
              </h4>
              <p class="text-sm tz-text-secondary leading-relaxed">
                {{ t('resourcesSpokeCalculator.parameter.items.crossPattern.desc') }}
              </p>
            </div>
          </div>

          <!-- Note Alert -->
           <div class="spoke-parameter__note-card">
             <span class="text-lg">💡</span>
             <p>{{ t('resourcesSpokeCalculator.parameter.note') }}</p>
          </div>

          <!-- Workflow Section -->
          <div class="spoke-parameter__workflow">
            <h4 class="spoke-parameter__workflow-title">{{ t('resourcesSpokeCalculator.parameter.workflow.title') }}</h4>

             <div class="spoke-parameter__workflow-visual">
              <GuideImage
                src="/public/technical/spoke-length.webp"
                :alt="t('resourcesSpokeCalculator.parameter.workflow.overviewAlt')"
                :zoomOnClick="true"
                :caption="t('resourcesSpokeCalculator.parameter.workflow.overviewCaption')"
              />
            </div>

            <div class="space-y-8">
              <!-- Step 1: Measure ERD -->
              <div class="spoke-parameter__step-card">
                <h5 class="tz-text-primary font-bold mb-3 flex items-center gap-2">
                  <span class="spoke-parameter__step-badge">1</span>
                  {{ t('resourcesSpokeCalculator.parameter.workflow.stepOneTitle') }}
                </h5>
                
                <div class="grid md:grid-cols-2 gap-6 items-start">
                    <div class="text-sm tz-text-secondary space-y-2 leading-relaxed">
                      <p>{{ t('resourcesSpokeCalculator.parameter.workflow.stepOneFormula') }} <strong>{{ t('resourcesSpokeCalculator.parameter.workflow.stepOneFormulaValue') }}</strong>.</p>
                       <ul class="list-disc list-inside space-y-1 ml-1 tz-text-muted">
                        <li
                          v-for="index in 4"
                          :key="`step-one-item-${index}`"
                        >
                          {{ t(`resourcesSpokeCalculator.parameter.workflow.stepOneItems.${index - 1}`) }}
                        </li>
                      </ul>
                       <p class="text-xs italic mt-2 tz-text-muted">
                         {{ t('resourcesSpokeCalculator.parameter.workflow.stepOneNote') }}
                      </p>
                   </div>
                   <div class="spoke-parameter__step-illustration">
                      <GuideImage
                        src="/public/technical/what-is-erd.webp"
                        :alt="t('resourcesSpokeCalculator.parameter.workflow.stepOneAlt')"
                        :zoomOnClick="true"
                        :caption="t('resourcesSpokeCalculator.parameter.workflow.stepOneCaption')"
                      />
                   </div>
                </div>
              </div>

               <!-- Step 2: Measure Hub -->
              <div class="spoke-parameter__step-card">
                <h5 class="tz-text-primary font-bold mb-3 flex items-center gap-2">
                  <span class="spoke-parameter__step-badge">2</span>
                  {{ t('resourcesSpokeCalculator.parameter.workflow.stepTwoTitle') }}
                </h5>
                 <div class="text-sm tz-text-secondary space-y-3 leading-relaxed">
                  <div>
                     <strong class="tz-text-primary">{{ t('resourcesSpokeCalculator.parameter.workflow.flangeDiameterLabel') }}</strong>
                     {{ t('resourcesSpokeCalculator.parameter.workflow.flangeDiameterBody') }}
                  </div>
                  <div>
                     <strong class="tz-text-primary">{{ t('resourcesSpokeCalculator.parameter.workflow.centerToFlangeLabel') }}</strong>
                     {{ t('resourcesSpokeCalculator.parameter.workflow.centerToFlangeBody') }}
                  </div>
                </div>
              </div>

               <!-- Step 3: Calculation -->
              <div class="spoke-parameter__step-card">
                <h5 class="tz-text-primary font-bold mb-3 flex items-center gap-2">
                  <span class="spoke-parameter__step-badge">3</span>
                  {{ t('resourcesSpokeCalculator.parameter.workflow.stepThreeTitle') }}
                </h5>
                 <div class="text-sm tz-text-secondary space-y-2 leading-relaxed">
                  <p>
                    {{ t('resourcesSpokeCalculator.parameter.workflow.stepThreeBody') }}
                  </p>
                  <p>
                     <strong class="tz-text-primary">{{ t('resourcesSpokeCalculator.parameter.workflow.tipLabel') }}</strong>
                     {{ t('resourcesSpokeCalculator.parameter.workflow.tipBody') }}
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
import SpokeCalculatorBlueprint from '~/components/SpokeCalculatorBlueprint.vue'
import SpokeSmartSearch from '~/components/SpokeSmartSearch.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

import GuideImage from '~/components/GuideImage.vue'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import { spokeCalculatorTabs } from '~/utils/pageSubNavigation'
import { usePageMessages } from '~/composables/usePageMessages'
import { definePageMeta, useHead, useI18n } from '#imports'
import { watch } from 'vue'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('resourcesSpokeCalculator')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const { activeTab } = usePageSubNavigationTab({
  tabs: spokeCalculatorTabs,
  basePath: '/resources/spoke-calculator',
  defaultValue: 'calculator',
  redirectBasePathToDefaultTab: true,
})

definePageMeta({
  layout: 'products',
  footerLabelKey: 'support.nav.spokeCalculator',
  footerLabelFallback: 'Spoke Calculator',
})

useHead(() => ({
  title: t('resourcesSpokeCalculator.title'),
}))
</script>

<style src="~/assets/css/guide-sections.css"></style>

<style scoped>
.support-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: var(--tz-text-primary);
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
  border-color: rgba(5, 150, 105, 0.42);
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
  border: 1px solid rgba(5, 150, 105, 0.24);
  border-radius: 0.5rem;
  background: rgba(5, 150, 105, 0.06);
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
  border: 1px solid rgba(5, 150, 105, 0.32);
  border-radius: 50%;
  background: rgba(5, 150, 105, 0.08);
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
