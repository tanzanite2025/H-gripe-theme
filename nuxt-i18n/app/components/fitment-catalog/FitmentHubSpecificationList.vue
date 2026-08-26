<template>
  <section class="fitment-hub-specifications" :aria-label="t('fitmentCatalog.entry.hubSpecifications')">
    <h4 class="fitment-hub-specifications__title">
      {{ t('fitmentCatalog.entry.hubSpecifications') }}
    </h4>

    <p v-if="!specifications.length" class="fitment-hub-specifications__empty">
      {{ t('fitmentCatalog.states.noHubSpecifications') }}
    </p>

    <div v-else class="fitment-hub-specifications__list">
      <article
        v-for="specification in specifications"
        :key="specification.id"
        class="fitment-hub-specification"
      >
        <div class="fitment-hub-specification__heading">
          <strong>{{ specification.display_name }}</strong>
          <span>{{ specification.spec_code }}</span>
        </div>

        <dl class="fitment-hub-specification__facts">
          <div>
            <dt>{{ t('fitmentCatalog.hub.position') }}</dt>
            <dd>{{ formatPosition(specification.position) }}</dd>
          </div>
          <div>
            <dt>{{ t('fitmentCatalog.hub.axleType') }}</dt>
            <dd>{{ formatAxleType(specification.axle_type) }}</dd>
          </div>
          <div>
            <dt>{{ t('fitmentCatalog.hub.axleSpacing') }}</dt>
            <dd>{{ specification.axle_spacing_mm }} {{ t('fitmentCatalog.hub.millimetres') }}</dd>
          </div>
        </dl>

        <p v-if="specification.notes" class="fitment-hub-specification__notes">
          {{ specification.notes }}
        </p>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import type {
  FitmentHubAxleType,
  FitmentHubPosition,
  FitmentHubSpecification,
} from '~/types/fitmentCatalog'

defineProps<{
  specifications: FitmentHubSpecification[]
}>()

const { t } = useI18n()

const formatPosition = (position: FitmentHubPosition) => {
  return t(`fitmentCatalog.hub.positions.${position}`)
}

const formatAxleType = (axleType: FitmentHubAxleType) => {
  const keyByAxleType: Record<FitmentHubAxleType, string> = {
    quick_release: 'quickRelease',
    thru_axle: 'thruAxle',
    bolt_on: 'boltOn',
    other: 'other',
  }
  return t(`fitmentCatalog.hub.axleTypes.${keyByAxleType[axleType]}`)
}
</script>

<style scoped>
.fitment-hub-specifications {
  display: grid;
  gap: 0.65rem;
}

.fitment-hub-specifications__title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.85rem;
  font-weight: 800;
}

.fitment-hub-specifications__empty {
  margin: 0;
  color: var(--tz-text-muted);
  font-size: 0.8rem;
}

.fitment-hub-specifications__list {
  display: grid;
  gap: 0.55rem;
}

.fitment-hub-specification {
  display: grid;
  gap: 0.55rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  padding: 0.7rem;
  background: var(--tz-surface-subtle);
}

.fitment-hub-specification__heading {
  display: flex;
  min-width: 0;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
}

.fitment-hub-specification__heading strong {
  min-width: 0;
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  overflow-wrap: anywhere;
}

.fitment-hub-specification__heading span {
  flex: 0 0 auto;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  overflow-wrap: anywhere;
}

.fitment-hub-specification__facts {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.5rem;
  margin: 0;
}

.fitment-hub-specification__facts div {
  min-width: 0;
}

.fitment-hub-specification__facts dt {
  color: var(--tz-text-muted);
  font-size: 0.68rem;
}

.fitment-hub-specification__facts dd {
  margin: 0.15rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
  font-weight: 700;
  overflow-wrap: anywhere;
}

.fitment-hub-specification__notes {
  margin: 0;
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

@media (max-width: 480px) {
  .fitment-hub-specification__heading {
    display: grid;
    gap: 0.15rem;
  }

  .fitment-hub-specification__facts {
    grid-template-columns: 1fr;
  }
}
</style>
