<template>
  <div class="wheelset-policy-flow">
    <aside class="wheelset-guide-panel wheelset-guide-panel--notice" role="note">
      <strong>{{ t('guidesWheelsetBuyersSample.serviceArea.title') }}</strong>
      <p>{{ t('guidesWheelsetBuyersSample.serviceArea.body') }}</p>
    </aside>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">1</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">
          {{ t('guidesWheelsetBuyersSample.professional.title') }}
        </h3>
        <p class="wheelset-guide-step__body">
          {{ t('guidesWheelsetBuyersSample.professional.body') }}
        </p>

        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSample.professional.feeTitle') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSample.professional.feeBody') }}</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSample.professional.bestForTitle') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSample.professional.bestForBody') }}</p>
          </div>
        </div>

        <div class="wheelset-guide-actions">
          <button
            type="button"
            class="wheelset-guide-button wheelset-guide-button--solid"
            @click="openWhatsAppChat"
          >
              {{ t('guidesWheelsetBuyersSample.professional.button') }}
          </button>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">2</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">
          {{ t('guidesWheelsetBuyersSample.process.title') }}
        </h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSample.process.shipping.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSample.process.shipping.body') }}</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSample.process.confirmation.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSample.process.confirmation.body') }}</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSample.process.assembly.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSample.process.assembly.body') }}</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSample.process.delivery.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSample.process.delivery.body') }}</p>
          </div>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">3</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">
          {{ t('guidesWheelsetBuyersSample.scope.title') }}
        </h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSample.scope.acceptedTitle') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSample.scope.acceptedBody') }}</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSample.scope.spareSpokesTitle') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSample.scope.spareSpokesBody') }}</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSample.scope.importantTitle') }}</strong>
            <ul>
              <li v-for="(item, index) in scopeItems" :key="index">
                {{ rt(item) }}
              </li>
            </ul>
          </div>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">4</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">
          {{ t('guidesWheelsetBuyersSample.ownBuild.title') }}
        </h3>
        <p class="wheelset-guide-step__body">
          {{ t('guidesWheelsetBuyersSample.ownBuild.body') }}
        </p>
        <div class="wheelset-guide-actions">
          <button
            type="button"
            class="wheelset-guide-button"
            @click="goToHolePatterns"
          >
            {{ t('guidesWheelsetBuyersSample.ownBuild.holeButton') }}
          </button>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">5</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">
          {{ t('guidesWheelsetBuyersSample.customBuilds.mullet.title') }}
        </h3>
        <p class="wheelset-guide-step__body">
          {{ t('guidesWheelsetBuyersSample.customBuilds.mullet.body') }}
        </p>

        <div class="wheelset-guide-panel-grid wheelset-guide-panel-grid--two">
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSample.customBuilds.mullet.frontTitle') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSample.customBuilds.mullet.frontBody') }}</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSample.customBuilds.mullet.rearTitle') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSample.customBuilds.mullet.rearBody') }}</p>
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
            {{ t('guidesWheelsetBuyersSample.customBuilds.mullet.button') }}
          </button>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">6</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">
          {{ t('guidesWheelsetBuyersSample.customBuilds.singleWheel.title') }}
        </h3>
        <p class="wheelset-guide-step__body">
          {{ t('guidesWheelsetBuyersSample.customBuilds.singleWheel.body') }}
        </p>

        <div class="wheelset-guide-panel">
          <strong>{{ t('guidesWheelsetBuyersSample.customBuilds.singleWheel.beforeTitle') }}</strong>
          <p>
            {{ t('guidesWheelsetBuyersSample.customBuilds.singleWheel.beforeBody') }}
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
            {{ t('guidesWheelsetBuyersSample.customBuilds.singleWheel.button') }}
          </button>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">7</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">
          {{ t('guidesWheelsetBuyersSample.customBuilds.mixedRim.title') }}
        </h3>
        <WheelsetMixedRimSection
          :openQuickBuy="openQuickBuy"
          :openWhatsAppChat="openWhatsAppChat"
        />
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import WheelsetMixedRimSection from '~/components/WheelsetMixedRimSection.vue'

const { locale, t, tm, rt } = useI18n()
const { loadPageMessages } = usePageMessages('guidesWheelsetBuyersSample')
await loadPageMessages(locale.value)
watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const scopeItems = computed<string[]>(() => (
  tm('guidesWheelsetBuyersSample.scope.items') as string[]
))

const { openWhatsAppChat, openQuickBuy, ctaNotes, goToHolePatterns } = defineProps<{
  openWhatsAppChat: () => void | Promise<void>
  openQuickBuy: () => void
  ctaNotes: {
    mulletQuickBuyEntry: string
    singleWheelChat: string
  }
  goToHolePatterns: () => void | Promise<void>
}>()
</script>
