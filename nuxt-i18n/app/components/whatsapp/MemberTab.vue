<template>
  <div class="h-full overflow-y-auto px-1 pt-1 pb-3 md:p-6">
    <div class="member-tab-card w-full md:max-w-md md:mx-auto rounded-2xl p-3 md:p-4 space-y-3 md:space-y-4">
      <!-- 顶部：当前等级 / 提示 -->
      <div class="flex items-center gap-3">
        <div class="member-tab-avatar w-9 h-9 md:w-10 md:h-10 rounded-full flex items-center justify-center tz-caption md:text-xs font-semibold tz-text-primary">
          {{ isMemberLogged ? (levelName || '—') : 'Guest' }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="tz-micro-label tz-text-muted truncate">
            {{ isMemberLogged ? 'Your membership' : 'Membership program' }}
          </div>
          <div class="text-sm font-semibold tz-text-primary truncate">
            <span v-if="isMemberLogged">Level {{ levelName }}</span>
            <span v-else>Log in to unlock member prices</span>
          </div>
        </div>
      </div>

      <!-- 核心指标网格 -->
      <div class="grid grid-cols-2 gap-2 md:gap-3 tz-caption">
        <div class="member-tab-metric rounded-xl px-2.5 md:px-3 py-2">
          <div class="tz-text-muted">Points</div>
          <div class="text-sm font-semibold tz-text-primary">
            {{ isMemberLogged ? points : '—' }}
          </div>
        </div>
        <div class="member-tab-metric rounded-xl px-2.5 md:px-3 py-2">
          <div class="tz-text-muted">Discount rate</div>
          <div class="text-sm font-semibold tz-text-primary">
            {{ isMemberLogged ? formatDiscountRate(levelDiscounts.discountRate) : '—' }}
          </div>
        </div>
        <div class="member-tab-metric rounded-xl px-2.5 md:px-3 py-2">
          <div class="tz-text-muted">Coupons / Cards</div>
          <div class="text-sm font-semibold tz-text-primary">
            {{ isMemberLogged ? `× ${userCoupons} / × ${userPointCards}` : '—' }}
          </div>
        </div>
      </div>

      <!-- 等级进度条 -->
      <div v-if="isMemberLogged" class="space-y-1.5">
        <div class="h-1.5 rounded-full tz-surface-subtle overflow-hidden">
          <div
            class="member-tab-progress__bar h-full"
            :style="{ width: tierInfo.pct + '%' }"
          ></div>
        </div>
        <div class="flex items-center justify-between tz-micro-label md:text-xs tz-text-secondary">
          <span>{{ tierInfo.current ? tierInfo.current.min : 0 }}</span>
          <span class="font-semibold tz-text-primary">{{ tierInfo.pct }}%</span>
          <span>
            {{
              tierInfo.next
                ? tierInfo.next.min
                : (tierInfo.current && tierInfo.current.max !== -1 ? tierInfo.current.max : 'MAX')
            }}
          </span>
        </div>
      </div>

      <div v-else class="tz-caption tz-text-secondary space-y-2 md:space-y-3">
        <p>Log in or sign up to see your member prices, points and progress.</p>
        <div class="flex gap-1.5 md:gap-2">
          <button
            type="button"
            class="member-tab-primary-action flex-1 h-8 md:h-9 rounded-full tz-caption font-semibold transition-all"
            @click="$emit('openAuth', 'register')"
          >
            Sign up
          </button>
          <button
            type="button"
            class="member-tab-secondary-action flex-1 h-8 md:h-9 rounded-full tz-caption font-semibold transition-all"
            @click="$emit('openAuth', 'login')"
          >
            Log in
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  isMemberLogged: boolean
  levelName: string | number
  points: number | string
  tierInfo: any
  levelDiscounts: {
    discountRate?: number
  }
  userCoupons: number
  userPointCards: number
}>()

defineEmits<{
  'openAuth': [mode: 'login' | 'register']
}>()

const formatBenefitNumber = (value: number) => {
  const numericValue = Number(value)
  if (!Number.isFinite(numericValue)) return '0'
  return Number.isInteger(numericValue) ? String(numericValue) : numericValue.toFixed(2).replace(/\.?0+$/, '')
}

const formatDiscountRate = (value?: number) => `${formatBenefitNumber(Number(value ?? 0))}%`
</script>

<style scoped>
.member-tab-card {
  background: var(--tz-card-surface);
  border: none;
  box-shadow: 0 8px 20px rgba(20, 32, 43, 0.1);
  backdrop-filter: blur(12px);
}

.member-tab-avatar {
  background: var(--tz-surface-muted);
  border: 1px solid var(--tz-border-subtle);
}

.member-tab-metric {
  background: var(--tz-surface-subtle);
  border: 1px solid var(--tz-border-subtle);
}

.member-tab-progress__bar {
  background: var(--tz-site-accent);
}

.member-tab-primary-action {
  background: var(--tz-action-primary);
  border: none;
  color: var(--tz-action-primary-foreground);
  box-shadow: 0 4px 12px -4px rgb(15 23 42 / 0.24);
}

.member-tab-primary-action:hover {
  background: var(--tz-action-primary-hover);
  transform: translateY(-1px);
}

.member-tab-secondary-action {
  background: var(--tz-surface-muted);
  border: 1px solid var(--tz-border-subtle);
  color: var(--tz-text-primary);
  box-shadow:
    0 2px 6px -3px rgba(20, 32, 43, 0.12),
    0 0 6px rgba(20, 32, 43, 0.06);
}

.member-tab-secondary-action:hover {
  background: var(--tz-surface-subtle);
  transform: translateY(-1px);
  box-shadow:
    0 4px 12px -4px rgba(20, 32, 43, 0.16),
    0 0 8px rgba(20, 32, 43, 0.08);
}
</style>
