<template>
  <article class="schwalbe-tire-card">
    <details
      class="schwalbe-tire-card__accordion"
      :open="isAccordionExpanded"
      @toggle="handleAccordionToggle"
    >
      <summary class="schwalbe-tire-card__summary">
        <span class="schwalbe-tire-card__summary-main">
          <span class="schwalbe-tire-card__summary-model">{{ item.model_name }}</span>
          <span class="schwalbe-tire-card__summary-size">{{ sizeSummary }}</span>
        </span>
        <span class="schwalbe-tire-card__summary-weight">
          <span class="schwalbe-tire-card__summary-weight-label">{{ tx('fields.weight') }}</span>
          <span>{{ numberWithUnit(item.weight_g, 'g') }}</span>
        </span>
        <Icon name="lucide:chevron-down" class="schwalbe-tire-card__summary-icon" aria-hidden="true" />
      </summary>

      <div class="schwalbe-tire-card__content">
        <header class="schwalbe-tire-card__header">
          <div class="schwalbe-tire-card__heading">
            <p class="schwalbe-tire-card__eyebrow">{{ item.article_no }}</p>
            <h3 class="schwalbe-tire-card__title">{{ item.model_name }}</h3>
            <p class="schwalbe-tire-card__subtitle">{{ sizeSummary }}</p>
          </div>
          <span
            class="schwalbe-tire-card__status"
            :class="item.product_exists ? 'schwalbe-tire-card__status--listed' : 'schwalbe-tire-card__status--candidate'"
          >
            {{ item.product_exists ? tx('status.listed') : tx('status.candidate') }}
          </span>
        </header>

        <dl class="schwalbe-tire-card__facts">
          <div class="schwalbe-tire-card__fact schwalbe-tire-card__fact--weight">
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

        <div v-if="rimRecommendations.length" class="schwalbe-tire-card__rim-guidance">
          <strong>{{ tx('rimWidth.recommendationLabel') }}</strong>
          <div
            v-for="recommendation in rimRecommendations"
            :key="recommendation.rim_system"
            class="schwalbe-tire-card__rim-recommendation"
          >
            <span class="schwalbe-tire-card__rim-system">
              {{ recommendation.rim_system === 'hookless' ? tx('rimWidth.hookless') : tx('rimWidth.hooked') }}
            </span>
            <span>
              {{ formatTireRimWidthReferenceRanges(recommendation.rim_width_ranges) }} mm
            </span>
            <small v-if="recommendation.result_kind === 'possible_reference'">
              {{ tx('rimWidth.possibleReference') }}
            </small>
            <small v-else-if="recommendation.result_kind === 'interpolated'">
              {{ tx('rimWidth.calculated') }}
            </small>
            <small
              v-if="recommendation.result_kind === 'interpolated' && recommendation.source_rows.some(sourceRow => sourceRow.kind === 'possible_reference')"
            >
              {{ tx('rimWidth.calculatedFromPossibleReference') }}
            </small>
          </div>
          <p>{{ tx('rimWidth.sourceNote') }}</p>
        </div>

        <footer class="schwalbe-tire-card__footer">
          <span>{{ tx('source.checked', { date: checkedDate }) }}</span>
        </footer>
      </div>
    </details>
  </article>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from '#imports'
import { formatTireRimWidthReferenceRanges, type TireRimWidthReferenceSuggestion } from '~/data/tireguides/tireRimWidthReferencePresentation'
import type { SchwalbeTireCatalogItem } from '~/data/tireguides/schwalbeCatalog'

const props = defineProps<{
  item: SchwalbeTireCatalogItem
}>()

const isAccordionExpanded = ref(true)
let mobileAccordionMediaQuery: MediaQueryList | null = null
let previousViewportWasMobile = false

const handleAccordionToggle = (event: Event) => {
  const details = event.currentTarget as HTMLDetailsElement | null
  if (!details) return

  if (mobileAccordionMediaQuery && !mobileAccordionMediaQuery.matches && !details.open) {
    details.open = true
    isAccordionExpanded.value = true
    return
  }

  isAccordionExpanded.value = details.open
}

const handleViewportChange = () => {
  if (!mobileAccordionMediaQuery) return

  const viewportIsMobile = mobileAccordionMediaQuery.matches
  if (viewportIsMobile === previousViewportWasMobile) return

  previousViewportWasMobile = viewportIsMobile
  isAccordionExpanded.value = !viewportIsMobile
}

onMounted(() => {
  mobileAccordionMediaQuery = window.matchMedia('(max-width: 760.5px)')
  previousViewportWasMobile = mobileAccordionMediaQuery.matches
  isAccordionExpanded.value = !previousViewportWasMobile
  mobileAccordionMediaQuery.addEventListener('change', handleViewportChange)
})

onBeforeUnmount(() => {
  mobileAccordionMediaQuery?.removeEventListener('change', handleViewportChange)
})

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

const rimRecommendations = computed<TireRimWidthReferenceSuggestion[]>(() => (
  props.item.rim_width_guidance || []
))

const sizeSummary = computed(() => {
  if (props.item.wheel_size) {
    return tx('wheelSizeCard', {
      diameter: props.item.wheel_size.wheelDiameterIn,
      bsd: props.item.wheel_size.beadSeatDiameterMm,
      etrto: props.item.etrto,
    })
  }

  return props.item.inch_designation
    ? `${props.item.etrto} · ${props.item.inch_designation}`
    : props.item.etrto
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

.schwalbe-tire-card__accordion {
  display: grid;
  gap: 1rem;
}

.schwalbe-tire-card__summary {
  display: none;
}

.schwalbe-tire-card__content {
  display: grid;
  gap: 1rem;
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

.schwalbe-tire-card__rim-guidance {
  display: grid;
  gap: 0.35rem;
  border-radius: 0.65rem;
  background: color-mix(in srgb, var(--tz-action-primary) 8%, var(--tz-card-surface));
  color: var(--tz-text-secondary);
  padding: 0.55rem 0.65rem;
  font-size: 0.72rem;
  line-height: 1.4;
}

.schwalbe-tire-card__rim-guidance strong {
  color: var(--tz-action-primary);
  font-weight: 750;
}

.schwalbe-tire-card__rim-recommendation {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem 0.55rem;
  align-items: baseline;
  color: var(--tz-text-primary);
  font-weight: 650;
}

.schwalbe-tire-card__rim-system {
  min-width: 4.8rem;
  color: var(--tz-text-secondary);
  font-weight: 750;
}

.schwalbe-tire-card__rim-recommendation small {
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  font-weight: 500;
}

.schwalbe-tire-card__rim-guidance p {
  margin: 0.15rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  line-height: 1.45;
}

@media (max-width: 760.5px) {
  .schwalbe-tire-card {
    padding: 0.75rem;
  }

  .schwalbe-tire-card__summary {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto auto;
    gap: 0.65rem;
    align-items: center;
    cursor: pointer;
    list-style: none;
  }

  .schwalbe-tire-card__summary::-webkit-details-marker {
    display: none;
  }

  .schwalbe-tire-card__summary::marker {
    content: '';
  }

  .schwalbe-tire-card__summary-main {
    display: grid;
    min-width: 0;
    gap: 0.2rem;
  }

  .schwalbe-tire-card__summary-model,
  .schwalbe-tire-card__summary-size {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .schwalbe-tire-card__summary-model {
    color: var(--tz-text-primary);
    font-size: 0.92rem;
    font-weight: 750;
    line-height: 1.25;
  }

  .schwalbe-tire-card__summary-size {
    color: var(--tz-text-secondary);
    font-size: 0.76rem;
    line-height: 1.25;
  }

  .schwalbe-tire-card__summary-weight {
    display: grid;
    justify-items: end;
    gap: 0.15rem;
    color: var(--tz-text-primary);
    font-size: 0.8rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .schwalbe-tire-card__summary-weight-label {
    color: var(--tz-text-secondary);
    font-size: 0.68rem;
    font-weight: 500;
  }

  .schwalbe-tire-card__summary-icon {
    width: 1rem;
    height: 1rem;
    color: var(--tz-text-secondary);
    transition: transform 160ms ease;
  }

  .schwalbe-tire-card__accordion[open] .schwalbe-tire-card__summary-icon {
    transform: rotate(180deg);
  }

  .schwalbe-tire-card__heading .schwalbe-tire-card__title,
  .schwalbe-tire-card__heading .schwalbe-tire-card__subtitle {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    clip-path: inset(50%);
    white-space: nowrap;
  }

  .schwalbe-tire-card__fact--weight {
    display: none;
  }

  .schwalbe-tire-card__header {
    display: grid;
  }

  .schwalbe-tire-card__status {
    justify-self: start;
  }
}
</style>
