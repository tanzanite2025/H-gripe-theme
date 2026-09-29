<template>
  <section class="spoke-catalog-system">
    <header class="spoke-catalog-system__header">
      <div>
        <p class="spoke-catalog-system__eyebrow">
          {{ t('resourcesSpokeCalculator.catalog.eyebrow') }}
        </p>
        <h2 class="spoke-catalog-system__title">
          {{ t('resourcesSpokeCalculator.catalog.title') }}
        </h2>
      </div>
      <p class="spoke-catalog-system__description">
        {{ t('resourcesSpokeCalculator.catalog.description') }}
      </p>
    </header>

    <div class="spoke-catalog-system__selection-grid">
      <div class="spoke-catalog-system__selection-card">
        <h3 class="spoke-catalog-system__selection-title">
          {{ t('resourcesSpokeCalculator.calculator.frontWheel') }}
        </h3>

        <div class="spoke-catalog-system__fields">
          <div class="spoke-catalog-system__field">
            <label :for="fieldId('front', 'rim-brand')">
              {{ t('resourcesSpokeCalculator.calculator.buildSettings.rimBrand') }}
            </label>
            <SpokeCalculatorSelect
              :id="fieldId('front', 'rim-brand')"
              v-model="frontSelection.rimBrandId"
              :options="frontOptions.rimBrandOptions"
              :placeholder="t('resourcesSpokeCalculator.calculator.buildSettings.selectBrand')"
            />
          </div>

          <div class="spoke-catalog-system__field">
            <label :for="fieldId('front', 'rim-model')">
              {{ t('resourcesSpokeCalculator.calculator.buildSettings.rimModel') }}
            </label>
            <SpokeCalculatorSelect
              :id="fieldId('front', 'rim-model')"
              v-model="frontSelection.rimModelId"
              :disabled="frontOptions.rimModelOptions.length === 0"
              :options="frontOptions.rimModelOptions"
              :placeholder="t('resourcesSpokeCalculator.calculator.buildSettings.selectModel')"
            />
          </div>

          <div class="spoke-catalog-system__field">
            <label :for="fieldId('front', 'hub-brand')">
              {{ t('resourcesSpokeCalculator.calculator.buildSettings.hubBrand') }}
            </label>
            <SpokeCalculatorSelect
              :id="fieldId('front', 'hub-brand')"
              v-model="frontSelection.hubBrandId"
              :options="frontOptions.hubBrandOptions"
              :placeholder="t('resourcesSpokeCalculator.calculator.buildSettings.selectBrand')"
            />
          </div>

          <div class="spoke-catalog-system__field">
            <label :for="fieldId('front', 'hub-model')">
              {{ t('resourcesSpokeCalculator.calculator.buildSettings.hubModel') }}
            </label>
            <SpokeCalculatorSelect
              :id="fieldId('front', 'hub-model')"
              v-model="frontSelection.hubModelId"
              :disabled="frontOptions.hubModelOptions.length === 0"
              :options="frontOptions.hubModelOptions"
              :placeholder="t('resourcesSpokeCalculator.calculator.buildSettings.selectModel')"
            />
          </div>
        </div>
      </div>

      <div class="spoke-catalog-system__selection-card">
        <h3 class="spoke-catalog-system__selection-title">
          {{ t('resourcesSpokeCalculator.calculator.rearWheel') }}
        </h3>

        <div class="spoke-catalog-system__fields">
          <div class="spoke-catalog-system__field">
            <label :for="fieldId('rear', 'rim-brand')">
              {{ t('resourcesSpokeCalculator.calculator.buildSettings.rimBrand') }}
            </label>
            <SpokeCalculatorSelect
              :id="fieldId('rear', 'rim-brand')"
              v-model="rearSelection.rimBrandId"
              :options="rearOptions.rimBrandOptions"
              :placeholder="t('resourcesSpokeCalculator.calculator.buildSettings.selectBrand')"
            />
          </div>

          <div class="spoke-catalog-system__field">
            <label :for="fieldId('rear', 'rim-model')">
              {{ t('resourcesSpokeCalculator.calculator.buildSettings.rimModel') }}
            </label>
            <SpokeCalculatorSelect
              :id="fieldId('rear', 'rim-model')"
              v-model="rearSelection.rimModelId"
              :disabled="rearOptions.rimModelOptions.length === 0"
              :options="rearOptions.rimModelOptions"
              :placeholder="t('resourcesSpokeCalculator.calculator.buildSettings.selectModel')"
            />
          </div>

          <div class="spoke-catalog-system__field">
            <label :for="fieldId('rear', 'hub-brand')">
              {{ t('resourcesSpokeCalculator.calculator.buildSettings.hubBrand') }}
            </label>
            <SpokeCalculatorSelect
              :id="fieldId('rear', 'hub-brand')"
              v-model="rearSelection.hubBrandId"
              :options="rearOptions.hubBrandOptions"
              :placeholder="t('resourcesSpokeCalculator.calculator.buildSettings.selectBrand')"
            />
          </div>

          <div class="spoke-catalog-system__field">
            <label :for="fieldId('rear', 'hub-model')">
              {{ t('resourcesSpokeCalculator.calculator.buildSettings.hubModel') }}
            </label>
            <SpokeCalculatorSelect
              :id="fieldId('rear', 'hub-model')"
              v-model="rearSelection.hubModelId"
              :disabled="rearOptions.hubModelOptions.length === 0"
              :options="rearOptions.hubModelOptions"
              :placeholder="t('resourcesSpokeCalculator.calculator.buildSettings.selectModel')"
            />
          </div>
        </div>
      </div>
    </div>

    <p class="spoke-catalog-system__note">
      {{ t('resourcesSpokeCalculator.catalog.note') }}
    </p>

    <div class="spoke-catalog-system__search">
      <div class="spoke-catalog-system__search-header">
        <h3>{{ t('resourcesSpokeCalculator.search.title') }}</h3>
        <p>{{ t('resourcesSpokeCalculator.search.intro') }}</p>
      </div>
      <SpokeSmartSearch
        :front-selection="frontSelection"
        :rear-selection="rearSelection"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import SpokeCalculatorSelect from '~/components/SpokeCalculatorSelect.vue'
import SpokeSmartSearch from '~/components/SpokeSmartSearch.vue'
import { useSpokeCalculatorCatalogSelection } from '~/composables/useSpokeCalculatorCatalogSelection'
import { useSpokeCalculatorWheelCatalog } from '~/composables/useSpokeCalculatorWheelCatalog'

const { t } = useI18n()
const { front: frontSelection, rear: rearSelection } = useSpokeCalculatorCatalogSelection()
const { frontOptions, rearOptions } = useSpokeCalculatorWheelCatalog(frontSelection, rearSelection)

const fieldId = (side: 'front' | 'rear', field: string) => `catalog-${side}-${field}`
</script>

<style scoped>
.spoke-catalog-system {
  display: grid;
  gap: 1rem;
  margin-top: 2rem;
  padding: 1.25rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
  box-shadow: 0 10px 26px -14px rgba(20, 32, 43, 0.12);
  color: var(--tz-text-primary);
}

.spoke-catalog-system__header {
  display: grid;
  gap: 0.4rem;
}

.spoke-catalog-system__eyebrow {
  margin: 0;
  color: var(--tz-text-accent);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.14em;
  line-height: 1.2;
  text-transform: uppercase;
}

.spoke-catalog-system__title,
.spoke-catalog-system__selection-title,
.spoke-catalog-system__search-header h3 {
  margin: 0;
  color: var(--tz-text-primary);
  font-weight: 700;
}

.spoke-catalog-system__title {
  font-size: 1rem;
  line-height: 1.35;
}

.spoke-catalog-system__description,
.spoke-catalog-system__note,
.spoke-catalog-system__search-header p {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.5;
}

.spoke-catalog-system__selection-grid {
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.spoke-catalog-system__selection-card {
  display: grid;
  gap: 0.75rem;
  min-width: 0;
  padding: 0.9rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.625rem;
  background: var(--tz-form-panel-surface);
}

.spoke-catalog-system__selection-title {
  color: var(--tz-text-accent);
  font-size: 0.82rem;
  line-height: 1.3;
}

.spoke-catalog-system__fields {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.spoke-catalog-system__field {
  display: grid;
  gap: 0.3rem;
  min-width: 0;
}

.spoke-catalog-system__field label {
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  font-weight: 600;
  line-height: 1.3;
}

.spoke-catalog-system__note {
  padding: 0.65rem 0.75rem;
  border: 1px solid rgba(5, 150, 105, 0.22);
  border-radius: 0.5rem;
  background: rgba(5, 150, 105, 0.06);
}

.spoke-catalog-system__search {
  display: grid;
  gap: 0.85rem;
  border-top: 1px solid var(--tz-border-subtle);
  padding-top: 1rem;
}

.spoke-catalog-system__search-header {
  display: grid;
  gap: 0.25rem;
  text-align: center;
}

.spoke-catalog-system__search-header h3 {
  font-size: 0.95rem;
  line-height: 1.35;
}

@media (max-width: 767px) {
  .spoke-catalog-system {
    padding: 0.85rem;
  }

  .spoke-catalog-system__selection-grid,
  .spoke-catalog-system__fields {
    grid-template-columns: 1fr;
  }
}
</style>
