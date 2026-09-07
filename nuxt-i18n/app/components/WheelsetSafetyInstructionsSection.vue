<template>
  <div class="wheelset-policy-flow">
    <article class="wheelset-guide-step wheelset-safety-notice">
      <div class="wheelset-guide-step__number wheelset-safety-notice__icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>
      </div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title wheelset-safety-notice__title">
          {{ t('guidesWheelsetBuyersSafety.notice.title') }}
        </h3>
        <p class="wheelset-guide-step__body">
          {{ t('guidesWheelsetBuyersSafety.notice.body') }}
        </p>
        <div class="wheelset-guide-panel">
          <p>{{ t('guidesWheelsetBuyersSafety.notice.experienceNote') }}</p>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">1</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">
          {{ t('guidesWheelsetBuyersSafety.installation.title') }}
        </h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.installation.professional.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSafety.installation.professional.body') }}</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.installation.tire.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSafety.installation.tire.body') }}</p>
            <div class="wheelset-guide-actions">
              <button type="button" class="wheelset-guide-button" @click="goToTubelessInstallation">
                  {{ t('guidesWheelsetBuyersSafety.installation.tire.button') }}
              </button>
            </div>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.installation.torque.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSafety.installation.torque.body') }}</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.installation.brakePads.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSafety.installation.brakePads.body') }}</p>
          </div>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">2</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">
          {{ t('guidesWheelsetBuyersSafety.riding.title') }}
        </h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.riding.weight.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSafety.riding.weight.body') }}</p>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.riding.road.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSafety.riding.road.body') }}</p>
            <div class="wheelset-guide-note">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
              <span>{{ t('guidesWheelsetBuyersSafety.riding.road.note') }}</span>
            </div>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.riding.braking.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSafety.riding.braking.body') }}</p>
          </div>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">3</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">
          {{ t('guidesWheelsetBuyersSafety.maintenance.title') }}
        </h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.maintenance.inspection.title') }}</strong>
            <ul>
              <li v-for="(item, index) in inspectionItems" :key="index">
                {{ rt(item) }}
              </li>
            </ul>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.maintenance.tension.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSafety.maintenance.tension.body') }}</p>
            <div class="wheelset-guide-actions">
              <button type="button" class="wheelset-guide-button" @click="goToTechnicalTension">
                      {{ t('guidesWheelsetBuyersSafety.maintenance.tension.button') }}
              </button>
            </div>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.maintenance.cleaning.title') }}</strong>
            <p>{{ t('guidesWheelsetBuyersSafety.maintenance.cleaning.body') }}</p>
          </div>
        </div>
      </div>
    </article>

    <article class="wheelset-guide-step">
      <div class="wheelset-guide-step__number">4</div>
      <div class="wheelset-guide-step__content">
        <h3 class="wheelset-guide-step__title">
          {{ t('guidesWheelsetBuyersSafety.emergency.title') }}
        </h3>
        <div class="wheelset-guide-panel-grid">
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.emergency.prohibited.title') }}</strong>
            <ul>
              <li v-for="(item, index) in prohibitedItems" :key="index">
                {{ rt(item) }}
              </li>
            </ul>
          </div>
          <div class="wheelset-guide-panel">
            <strong>{{ t('guidesWheelsetBuyersSafety.emergency.protocol.title') }}</strong>
            <p>
              {{ t('guidesWheelsetBuyersSafety.emergency.protocol.bodyBefore') }}
              <strong>{{ t('guidesWheelsetBuyersSafety.emergency.protocol.stop') }}</strong>
            </p>
            <p class="mt-2">{{ t('guidesWheelsetBuyersSafety.emergency.protocol.bodyAfter') }}</p>
          </div>
        </div>

        <div class="wheelset-guide-actions">
          <button
            type="button"
            class="wheelset-guide-button"
            @click="goToWarranty"
          >
                    {{ t('guidesWheelsetBuyersSafety.emergency.protocol.button') }}
          </button>
        </div>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t, tm, rt } = useI18n()
const { loadPageMessages } = usePageMessages('guidesWheelsetBuyersSafety')
await loadPageMessages(locale.value)
watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const inspectionItems = computed<string[]>(() => (
  tm('guidesWheelsetBuyersSafety.maintenance.inspection.items') as string[]
))
const prohibitedItems = computed<string[]>(() => (
  tm('guidesWheelsetBuyersSafety.emergency.prohibited.items') as string[]
))

const { goToTubelessInstallation, goToTechnicalTension, goToWarranty } = defineProps<{
  goToTubelessInstallation: () => void | Promise<void>
  goToTechnicalTension: () => void | Promise<void>
  goToWarranty: () => void | Promise<void>
  }>()
</script>
