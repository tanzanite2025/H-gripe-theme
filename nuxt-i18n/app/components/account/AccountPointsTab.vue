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
        <span>{{ t('member.brief.productDiscount', 'Product Discount') }}</span>
        <strong>{{ discountText }}</strong>
      </div>
      <div class="points-card">
        <span>{{ t('member.brief.pointsDiscount', 'Points Discount') }}</span>
        <strong>{{ pointsDiscountText }}</strong>
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
      <NuxtLink :to="localePath('/membershipandpoints')" @click="$emit('close')">
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
    product?: number
    points?: number
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
const discountText = computed(() => `${Number(props.levelDiscounts?.product ?? 0)}%`)
const pointsDiscountText = computed(() => `${Number(props.levelDiscounts?.points ?? 0)}%`)
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
  background:
    radial-gradient(circle at top left, rgba(64, 255, 170, 0.16), transparent 52%),
    rgba(255, 255, 255, 0.055);
  padding: 1rem;
}

.points-hero p {
  margin: 0;
  color: rgba(203, 213, 225, 0.78);
  font-size: var(--tz-type-micro-label);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.points-hero h3 {
  margin: 0.25rem 0 0;
  color: #ffffff;
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
  color: #40ffaa;
  font-size: 1.7rem;
  font-weight: 900;
  line-height: 1;
}

.points-hero__score span {
  margin-top: 0.25rem;
  color: rgba(226, 232, 240, 0.75);
  font-size: var(--tz-type-micro-label);
  font-weight: 700;
}

.points-progress {
  border-radius: 1rem;
  background: rgba(255, 255, 255, 0.045);
  padding: 0.85rem;
}

.points-progress__bar {
  height: 0.52rem;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.22);
}

.points-progress__bar span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #40ffaa, #60a5fa);
}

.points-progress__meta {
  display: flex;
  justify-content: space-between;
  margin-top: 0.5rem;
  color: rgba(226, 232, 240, 0.78);
  font-size: 0.75rem;
}

.points-progress__meta strong {
  color: #ffffff;
}

.points-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.7rem;
}

.points-card {
  min-height: 4.2rem;
  border-radius: 1rem;
  background: rgba(255, 255, 255, 0.055);
  padding: 0.8rem;
}

.points-card span,
.points-card strong {
  display: block;
}

.points-card span {
  color: rgba(203, 213, 225, 0.82);
  font-size: var(--tz-type-micro-label);
  line-height: 1.35;
}

.points-card strong {
  margin-top: 0.4rem;
  color: #ffffff;
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
  background: rgba(255, 255, 255, 0.075);
  color: #ffffff;
}

.points-actions button:disabled {
  opacity: 0.68;
}

.points-actions a {
  background: linear-gradient(135deg, #4efce7, #60a5fa);
  color: #020617;
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

