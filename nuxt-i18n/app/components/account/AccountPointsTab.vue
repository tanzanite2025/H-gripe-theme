<template>
  <section class="account-tab-panel">
    <div class="points-hero">
      <div>
        <p>{{ t('accountSidebar.points.currentLevel', 'Current level') }}</p>
        <h3>{{ levelName }}</h3>
      </div>
      <div class="points-hero__score">
        <strong>{{ pointsNumber }}</strong>
        <span>{{ t('member.points.unit', 'pts') }}</span>
      </div>
    </div>

    <div class="points-progress" aria-label="Membership progress">
      <div class="points-progress__bar">
        <span :style="{ width: progressPct + '%' }"></span>
      </div>
      <div class="points-progress__meta">
        <span>{{ t('accountSidebar.points.progress', 'Progress') }}</span>
        <strong>{{ progressPct }}%</strong>
      </div>
    </div>

    <div class="points-grid">
      <div class="points-card">
        <span>{{ t('member.brief.discountRate', 'Discount Rate') }}</span>
        <strong>{{ discountText }}</strong>
      </div>
      <div class="points-card">
        <span>{{ t('member.coupons', 'Coupons') }}</span>
        <strong>{{ coupons }}</strong>
      </div>
      <div class="points-card">
        <span>{{ t('member.pointCards', 'Point Cards') }}</span>
        <strong>{{ pointCards }}</strong>
      </div>
    </div>

    <div class="points-actions">
      <button type="button" @click="$emit('refresh')" :disabled="loading">
        <Icon name="lucide:refresh-cw" :class="{ 'points-actions__spin': loading }" />
        {{ loading ? t('common.loading', 'Loading...') : t('accountSidebar.actions.refresh', 'Refresh') }}
      </button>
<NuxtLink :to="localePath('/resources/membershipandpoints')" @click="$emit('close')">
        {{ t('accountSidebar.points.viewFull', 'Full membership center') }}
      </NuxtLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useLocalePath } from '#imports'

const props = withDefaults(defineProps<{
  levelName?: string
  points?: number | string
  tierInfo?: { pct?: number }
  levelDiscounts?: {
    discountRate?: number
  }
  coupons?: number
  pointCards?: number
  loading?: boolean
}>(), {
  levelName: '—',
  points: 0,
  coupons: 0,
  pointCards: 0,
  loading: false,
})

defineEmits<{
  (event: 'refresh'): void
  (event: 'close'): void
}>()

const { t } = useI18n()
const localePath = useLocalePath()

const pointsNumber = computed(() => Number(props.points || 0))
const progressPct = computed(() => Math.max(0, Math.min(100, Number(props.tierInfo?.pct ?? 0))))
const formatBenefitNumber = (value: number) => {
  const numericValue = Number(value)
  if (!Number.isFinite(numericValue)) return '0'
  return Number.isInteger(numericValue) ? String(numericValue) : numericValue.toFixed(2).replace(/\.?0+$/, '')
}
const discountText = computed(() => `${formatBenefitNumber(Number(props.levelDiscounts?.discountRate ?? 0))}%`)
</script>

<style scoped>
.account-tab-panel {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
}

.points-hero {
  display: flex;
  align-items: stretch;
  justify-content: space-between;
  gap: 0.8rem;
  border-radius: 1.35rem;
  border: 1px solid var(--tz-border-subtle);
  background: var(--tz-card-surface);
  padding: 1rem;
}

.points-hero p {
  margin: 0;
  color: var(--tz-text-muted);
  font-size: var(--tz-type-micro-label);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.points-hero h3 {
  margin: 0.25rem 0 0;
  color: var(--tz-text-primary);
  font-size: 1.25rem;
  font-weight: 850;
}

.points-hero__score {
  display: flex;
  min-width: 5.2rem;
  flex-direction: column;
  align-items: flex-end;
  justify-content: center;
}

.points-hero__score strong {
  color: #059669;
  font-size: 1.7rem;
  font-weight: 900;
  line-height: 1;
}

.points-hero__score span {
  margin-top: 0.25rem;
  color: var(--tz-text-secondary);
  font-size: var(--tz-type-micro-label);
  font-weight: 700;
}

.points-progress {
  border-radius: 1rem;
  border: 1px solid var(--tz-border-subtle);
  background: var(--tz-input-surface);
  padding: 0.85rem;
}

.points-progress__bar {
  height: 0.52rem;
  overflow: hidden;
  border-radius: 999px;
  background: var(--tz-border-subtle);
}

.points-progress__bar span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: #059669;
}

.points-progress__meta {
  display: flex;
  justify-content: space-between;
  margin-top: 0.5rem;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
}

.points-progress__meta strong {
  color: var(--tz-text-primary);
}

.points-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.7rem;
}

.points-card {
  min-height: 4.2rem;
  border-radius: 1rem;
  border: 1px solid var(--tz-border-subtle);
  background: var(--tz-input-surface);
  padding: 0.8rem;
}

.points-card span,
.points-card strong {
  display: block;
}

.points-card span {
  color: var(--tz-text-secondary);
  font-size: var(--tz-type-micro-label);
  line-height: 1.35;
}

.points-card strong {
  margin-top: 0.4rem;
  color: var(--tz-text-primary);
  font-size: 1.05rem;
  font-weight: 850;
}

.points-actions {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.65rem;
}

.points-actions button,
.points-actions a {
  display: inline-flex;
  min-height: 2.55rem;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border-radius: 999px;
  font-size: 0.82rem;
  font-weight: 800;
  text-decoration: none;
}

.points-actions button {
  border: 1px solid var(--tz-border-strong);
  background: var(--tz-surface-subtle);
  color: var(--tz-text-primary);
}

.points-actions button:disabled {
  opacity: 0.68;
}

.points-actions a {
  background: var(--tz-site-accent);
  color: #ffffff;
}

.points-actions svg {
  width: 1rem;
  height: 1rem;
}

.points-actions__spin {
  animation: points-spin 0.85s linear infinite;
}

@keyframes points-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>

