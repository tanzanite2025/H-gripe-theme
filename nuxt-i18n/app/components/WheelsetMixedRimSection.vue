<template>
  <div class="wheelset-guide-subsection">
    <p class="wheelset-guide-step__body">
      {{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.body') }}
    </p>

    <div class="wheelset-guide-actions">
      <button
        type="button"
        class="wheelset-guide-button"
        @click="openQuickBuy"
      >
        {{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.builderButton') }}
      </button>
      <button
        type="button"
        class="wheelset-guide-button wheelset-guide-button--solid"
        @click="openWhatsAppChat"
      >
        {{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.expertButton') }}
      </button>
    </div>

    <div class="wheelset-guide-panel-grid">
      <div class="wheelset-guide-panel">
        <strong>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.differentModels.title') }}</strong>
        <p>
          {{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.differentModels.body') }}
        </p>
      </div>

      <div class="wheelset-guide-panel">
        <strong>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.differentHeights.title') }}</strong>
        <p>
          {{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.differentHeights.body') }}
        </p>
      </div>
    </div>

    <div>
      <h4 class="wheelset-guide-kicker">
        {{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.cheatsheet') }}
      </h4>

      <div class="wheelset-guide-panel-grid">
        <div class="wheelset-guide-panel">
          <strong>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.climbing.title') }}</strong>
          <p>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.climbing.sizes') }}</p>
          <p>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.climbing.body') }}</p>
        </div>
        <div class="wheelset-guide-panel">
          <strong>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.aero.title') }}</strong>
          <p>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.aero.sizes') }}</p>
          <p>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.aero.body') }}</p>
        </div>
        <div class="wheelset-guide-panel">
          <strong>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.training.title') }}</strong>
          <p>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.training.sizes') }}</p>
          <p>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.training.body') }}</p>
        </div>
        <div class="wheelset-guide-panel">
          <strong>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.versatile.title') }}</strong>
          <p>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.versatile.sizes') }}</p>
          <p>{{ t('guidesWheelsetBuyersSpecialOrder.mixedRim.versatile.body') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'

const { t } = useI18n()

const { openQuickBuy, openWhatsAppChat } = defineProps<{
  openQuickBuy: () => void | Promise<void>
  openWhatsAppChat: () => void | Promise<void>
}>()
</script>

<style scoped>
.guide-section__cta-wrapper {
  color: var(--tz-text-primary);
}

.wheelset-link {
  color: #059669;
  text-decoration: underline;
}

.wheelset-link:hover {
  color: var(--tz-site-accent-hover);
}

.wheelset-inline-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem 0.65rem;
  margin: 0 0.25rem;
  border-radius: 9999px;
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

.wheelset-guide-kicker {
  margin: 0 0 0.8rem;
  color: var(--tz-text-muted);
  font-size: var(--tz-type-micro-label);
  font-weight: 700;
  letter-spacing: 0.12em;
  line-height: 1.25;
  text-transform: uppercase;
}

.wheelset-guide-panel span {
  color: var(--tz-text-muted);
  font-weight: 650;
}
</style>
