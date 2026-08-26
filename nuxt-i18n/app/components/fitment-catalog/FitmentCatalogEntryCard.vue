<template>
  <article class="fitment-catalog-entry">
    <header class="fitment-catalog-entry__header">
      <div class="fitment-catalog-entry__identity">
        <p class="fitment-catalog-entry__brand">{{ entry.brand_name }}</p>
        <h3>{{ entry.model_name }}</h3>
      </div>
      <span class="fitment-catalog-entry__year">
        {{ formatYear(entry) }}
      </span>
    </header>

    <dl class="fitment-catalog-entry__meta">
      <div v-if="entry.series_name">
        <dt>{{ t('fitmentCatalog.entry.series') }}</dt>
        <dd>{{ entry.series_name }}</dd>
      </div>
      <div v-if="entry.generation_name">
        <dt>{{ t('fitmentCatalog.entry.generation') }}</dt>
        <dd>{{ entry.generation_name }}</dd>
      </div>
      <div v-if="entry.market_code">
        <dt>{{ t('fitmentCatalog.entry.market') }}</dt>
        <dd>{{ entry.market_code }}</dd>
      </div>
    </dl>

    <p v-if="entry.notes" class="fitment-catalog-entry__notes">
      {{ entry.notes }}
    </p>

    <FitmentHubSpecificationList :specifications="entry.hub_specifications" />
  </article>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import FitmentHubSpecificationList from '~/components/fitment-catalog/FitmentHubSpecificationList.vue'
import type {
  FitmentFrameEntry,
  FitmentForkEntry,
} from '~/types/fitmentCatalog'

const props = defineProps<{
  entry: FitmentFrameEntry | FitmentForkEntry
}>()

const { t } = useI18n()

const formatYear = (entry: FitmentFrameEntry | FitmentForkEntry) => {
  switch (entry.year_mode) {
    case 'single':
      return entry.year_from ? String(entry.year_from) : t('fitmentCatalog.entry.unknownYear')
    case 'range':
      return entry.year_from && entry.year_to
        ? `${entry.year_from} - ${entry.year_to}`
        : t('fitmentCatalog.entry.unknownYear')
    case 'all':
      return t('fitmentCatalog.entry.allYears')
    default:
      return t('fitmentCatalog.entry.unknownYear')
  }
}

</script>

<style scoped>
.fitment-catalog-entry {
  display: grid;
  gap: 0.9rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  padding: 1rem;
  background: var(--tz-surface-card);
}

.fitment-catalog-entry__header {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.fitment-catalog-entry__identity {
  min-width: 0;
}

.fitment-catalog-entry__brand {
  margin: 0 0 0.2rem;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  font-weight: 800;
  text-transform: uppercase;
}

.fitment-catalog-entry h3 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1rem;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.fitment-catalog-entry__year {
  flex: 0 0 auto;
  color: var(--tz-action-primary);
  font-size: 0.78rem;
  font-weight: 800;
  text-align: right;
}

.fitment-catalog-entry__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1rem;
  margin: 0;
}

.fitment-catalog-entry__meta div {
  min-width: 0;
}

.fitment-catalog-entry__meta dt {
  color: var(--tz-text-muted);
  font-size: 0.68rem;
}

.fitment-catalog-entry__meta dd {
  margin: 0.15rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  font-weight: 700;
  overflow-wrap: anywhere;
}

.fitment-catalog-entry__notes {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.55;
  overflow-wrap: anywhere;
}

@media (max-width: 480px) {
  .fitment-catalog-entry__header {
    display: grid;
    gap: 0.35rem;
  }

  .fitment-catalog-entry__year {
    text-align: left;
  }
}
</style>
