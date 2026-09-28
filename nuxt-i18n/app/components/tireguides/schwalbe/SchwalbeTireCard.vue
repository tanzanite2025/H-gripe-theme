<template>
  <article class="schwalbe-tire-card">
    <header class="schwalbe-tire-card__header">
      <div class="schwalbe-tire-card__heading">
        <p class="schwalbe-tire-card__eyebrow">{{ item.article_no }}</p>
        <h3 class="schwalbe-tire-card__title">{{ item.model_name }}</h3>
        <p v-if="item.inch_designation" class="schwalbe-tire-card__subtitle">
          {{ item.etrto }} · {{ item.inch_designation }}
        </p>
        <p v-else class="schwalbe-tire-card__subtitle">{{ item.etrto }}</p>
      </div>
      <span
        class="schwalbe-tire-card__status"
        :class="item.product_exists ? 'schwalbe-tire-card__status--listed' : 'schwalbe-tire-card__status--candidate'"
      >
        {{ item.product_exists ? tx('status.listed') : tx('status.candidate') }}
      </span>
    </header>

    <dl class="schwalbe-tire-card__facts">
      <div class="schwalbe-tire-card__fact">
        <dt>{{ tx('fields.weight') }}</dt>
        <dd>{{ numberWithUnit(item.weight_g, 'g') }}</dd>
      </div>
      <div class="schwalbe-tire-card__fact">
        <dt>{{ tx('fields.pressure') }}</dt>
        <dd>{{ pressureRange }}</dd>
      </div>
      <div class="schwalbe-tire-card__fact">
        <dt>{{ tx('fields.load') }}</dt>
        <dd>{{ numberWithUnit(item.load_kg, 'kg') }}</dd>
      </div>
      <div class="schwalbe-tire-card__fact">
        <dt>{{ tx('fields.epi') }}</dt>
        <dd>{{ item.epi ?? tx('common.notProvided') }}</dd>
      </div>
      <div class="schwalbe-tire-card__fact">
        <dt>{{ tx('fields.version') }}</dt>
        <dd>{{ item.version_label || tx('common.notProvided') }}</dd>
      </div>
      <div class="schwalbe-tire-card__fact">
        <dt>{{ tx('fields.compound') }}</dt>
        <dd>{{ item.compound || tx('common.notProvided') }}</dd>
      </div>
      <div class="schwalbe-tire-card__fact">
        <dt>{{ tx('fields.bead') }}</dt>
        <dd>{{ item.bead || tx('common.notProvided') }}</dd>
      </div>
      <div class="schwalbe-tire-card__fact">
        <dt>{{ tx('fields.eBike') }}</dt>
        <dd>{{ item.e_bike_rating || tx('common.notProvided') }}</dd>
      </div>
    </dl>

    <div class="schwalbe-tire-card__secondary">
      <span v-if="item.color">{{ tx('fields.color') }}: {{ item.color }}</span>
      <span v-if="item.seal">{{ tx('fields.seal') }}: {{ item.seal }}</span>
      <span v-if="item.tread">{{ tx('fields.tread') }}: {{ item.tread }}</span>
      <span v-if="item.ean">{{ tx('fields.ean') }}: {{ item.ean }}</span>
    </div>

    <footer class="schwalbe-tire-card__footer">
      <span>{{ tx('source.checked', { date: checkedDate }) }}</span>
    </footer>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import type { SchwalbeTireCatalogItem } from '~/data/tireguides/schwalbeCatalog'

const props = defineProps<{
  item: SchwalbeTireCatalogItem
}>()

const { t: translate, locale } = useI18n()
const tx = (key: string, params?: Record<string, unknown>) => translate(`guidesSchwalbeTireSelector.${key}`, params || {})
const intlLocale = computed(() => locale.value.replace(/_/g, '-'))

const numberWithUnit = (value: number | undefined, unit: string) => {
  if (value === undefined || !Number.isFinite(value)) return tx('common.notProvided')
  return `${new Intl.NumberFormat(intlLocale.value, { maximumFractionDigits: 2 }).format(value)} ${unit}`
}

const pressureRange = computed(() => {
  const barMin = props.item.min_pressure_bar
  const barMax = props.item.max_pressure_bar
  const psiMin = props.item.min_pressure_psi
  const psiMax = props.item.max_pressure_psi
  if (barMin !== undefined || barMax !== undefined) {
    const bar = `${barMin ?? '—'}–${barMax ?? '—'} bar`
    const psi = psiMin !== undefined || psiMax !== undefined
      ? ` (${psiMin ?? '—'}–${psiMax ?? '—'} PSI)`
      : ''
    return `${bar}${psi}`
  }
  if (psiMin !== undefined || psiMax !== undefined) {
    return `${psiMin ?? '—'}–${psiMax ?? '—'} PSI`
  }
  return tx('common.notProvided')
})

const checkedDate = computed(() => {
  const parsed = new Date(props.item.source_checked_at)
  if (Number.isNaN(parsed.getTime())) return props.item.source_checked_at
  return new Intl.DateTimeFormat(intlLocale.value, { dateStyle: 'medium' }).format(parsed)
})
</script>

<style scoped>
.schwalbe-tire-card {
  display: grid;
  gap: 1rem;
  min-width: 0;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: var(--tz-card-surface);
  padding: 1.1rem;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.06);
}

.schwalbe-tire-card__header {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
  justify-content: space-between;
}

.schwalbe-tire-card__heading {
  min-width: 0;
}

.schwalbe-tire-card__eyebrow {
  margin: 0 0 0.25rem;
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.schwalbe-tire-card__title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1.05rem;
  font-weight: 750;
  line-height: 1.35;
}

.schwalbe-tire-card__subtitle {
  margin: 0.2rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.84rem;
}

.schwalbe-tire-card__status {
  flex: 0 0 auto;
  border-radius: 999px;
  padding: 0.28rem 0.6rem;
  font-size: 0.7rem;
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
}

.schwalbe-tire-card__status--listed {
  background: var(--tz-status-success-bg);
  color: var(--tz-status-success-text);
}

.schwalbe-tire-card__status--candidate {
  background: var(--tz-surface-muted);
  color: var(--tz-text-secondary);
}

.schwalbe-tire-card__facts {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.65rem;
  margin: 0;
}

.schwalbe-tire-card__fact {
  min-width: 0;
  border-radius: 0.65rem;
  background: var(--tz-surface-subtle);
  padding: 0.55rem 0.65rem;
}

.schwalbe-tire-card__fact dt {
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  line-height: 1.25;
}

.schwalbe-tire-card__fact dd {
  margin: 0.2rem 0 0;
  overflow-wrap: anywhere;
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 650;
  line-height: 1.35;
}

.schwalbe-tire-card__secondary {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.74rem;
  line-height: 1.5;
}

.schwalbe-tire-card__footer {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 0.75rem;
  align-items: center;
  justify-content: flex-start;
  border-top: 1px solid var(--tz-border-subtle);
  padding-top: 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
}

@media (max-width: 480px) {
  .schwalbe-tire-card__header {
    display: grid;
  }

  .schwalbe-tire-card__status {
    justify-self: start;
  }
}
</style>
