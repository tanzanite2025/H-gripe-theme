<template>
  <article class="tire-pressure-product-standards-placeholder">
    <span class="tire-pressure-product-standards-placeholder__eyebrow">
      {{ t('guidesTirePressure.standardsPlaceholder.status') }}
    </span>
    <h3>{{ t('guidesTirePressure.standardsPlaceholder.title') }}</h3>
    <p>{{ t('guidesTirePressure.standardsPlaceholder.body') }}</p>
    <NuxtLink
      class="tire-pressure-product-standards-placeholder__calculator-link"
      :to="localePath('/guides/tireguides/tire-pressure-calculator')"
    >
      {{ t('guidesTirePressure.standardsPlaceholder.openCalculator') }}
    </NuxtLink>
  </article>
</template>

<script setup lang="ts">
import { useI18n, useLocalePath } from '#imports'

const { t } = useI18n()
const localePath = useLocalePath()
</script>

<style scoped>
.tire-pressure-product-standards-placeholder {
  display: grid;
  justify-items: center;
  gap: 0.65rem;
  align-content: center;
  padding: 1rem;
  border: 1px dashed var(--tz-border-strong);
  border-radius: 0.75rem;
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
  text-align: center;
}

.tire-pressure-product-standards-placeholder__eyebrow {
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.tire-pressure-product-standards-placeholder h3 {
  margin: 0;
  font-size: 1rem;
  line-height: 1.3;
}

.tire-pressure-product-standards-placeholder p {
  max-width: 38rem;
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.5;
}

.tire-pressure-product-standards-placeholder__calculator-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-top: 0.25rem;
  border-radius: 999px;
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  font-size: 0.72rem;
  font-weight: 700;
  padding: 0.42rem 0.85rem;
  text-decoration: none;
}

.tire-pressure-product-standards-placeholder__calculator-link:hover,
.tire-pressure-product-standards-placeholder__calculator-link:focus-visible {
  background: var(--tz-action-primary-hover);
  outline: none;
}
</style>
