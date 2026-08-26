<template>
  <div class="membership-tabs" :class="{ 'membership-tabs--modal': isModal, 'membership-tabs--sticky': isModal }">
    <div v-if="isModal" class="nav-pill-tabs" role="tablist" :aria-label="$t('member.tabs.ariaLabel', 'Membership sections')">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === tab.id }"
        role="tab"
        :aria-selected="activeTab === tab.id"
        @click="setActiveTab(tab.id)"
      >
        {{ $t(tab.labelKey || tab.id, tab.fallback || tab.label || tab.id) }}
      </button>
    </div>

    <div class="membership-tabs__content" :class="{ 'membership-tabs__content--scroll': isModal }">
      <div v-show="activeTab === 'myinfo'">
      <div class="warranty-card">
        <Icon name="lucide:shield-check" class="warranty-card__icon" aria-hidden="true" />
        <div class="warranty-card__content">
          <h3 class="warranty-card__title">{{ $t('warranty.title', 'Warranty Check') }}</h3>
          <p class="warranty-card__desc">{{ $t('warranty.cardDesc', 'Enter your order number to check shipped items, warranty time, and service history.') }}</p>
        </div>
        <NuxtLink :to="localePath('/support/warranty-check')" class="warranty-card__btn">
          {{ $t('warranty.checkNow', 'Check Now') }}
          <span class="arrow">→</span>
        </NuxtLink>
      </div>

      <section class="membership-section">
        <div class="membership-grid" :class="{ 'membership-grid--modal': isModal }">
          <div class="membership-col">
            <div class="member-header">
              <div class="member-avatar">
                <BadgeAvatar :logged="isLogged" :level="String(levelName)" :topTierImageUrl="String(topTierImage)" />
              </div>
              <div class="member-name" v-if="isLogged && profileInfo?.fullName">
                {{ profileInfo.fullName }}
              </div>
              <div class="member-level" v-if="isLogged">
                <span class="level-badge">{{ displayLevelName }}</span>
                <span class="level-points">{{ points }} {{ $t('member.points.unit', 'pts') }}</span>
              </div>
              <div class="member-actions">
                <template v-if="!isLogged">
                  <button class="btn-primary" @click="openAuthForm('register')">
                    {{ $t('user.register') }}
                  </button>
                  <button class="btn-secondary" @click="openAuthForm('login')">
                    {{ $t('user.login') }}
                  </button>
                </template>
                <template v-else>
                  <button class="btn-danger" @click="doLogout">
                    {{ $t('user.logout') }}
                  </button>
                </template>
              </div>
            </div>
          </div>

          <div class="membership-col">
            <div class="member-card member-benefits-card">
              <h4 class="card-title">{{ $t('member.myBenefits', 'My Benefits') }}</h4>
              <div class="member-stats">
                <div class="stat-item">
                  <Icon name="lucide:tag" class="stat-icon" aria-hidden="true" />
                  <div class="stat-content">
                    <span class="stat-copy">
                      <span class="stat-label">{{ $t('member.brief.level', 'Level') }}</span>
                      <span class="stat-desc">{{ $t('member.brief.levelDesc', 'Your membership tier reflects accumulated activity and unlocks the benefit rules connected to that tier.') }}</span>
                    </span>
                    <span class="stat-value" :class="{ 'highlight': !isLogged }">{{ isLogged ? displayLevelName : '?' }}</span>
                  </div>
                </div>
                <div class="stat-item">
                  <Icon name="lucide:badge-percent" class="stat-icon" aria-hidden="true" />
                  <div class="stat-content">
                    <span class="stat-copy">
                      <span class="stat-label">{{ $t('member.brief.discountRate', 'Discount Rate') }}</span>
                      <span class="stat-desc">{{ $t('member.brief.discountRateDesc', 'The member-level price discount configured in the backend.') }}</span>
                    </span>
                    <span class="stat-value" :class="{ 'highlight': !isLogged }">{{ isLogged ? formatDiscountRate(levelDiscounts.discountRate) : '?' }}</span>
                  </div>
                </div>
              </div>

              <div class="member-assets">
                <div class="asset-item">
                  <Icon name="lucide:ticket-percent" class="asset-icon" aria-hidden="true" />
                  <div class="asset-content">
                    <span class="asset-label">{{ $t('member.coupons', 'Coupons') }}</span>
                    <span class="asset-value">{{ isLogged ? `× ${userCoupons}` : '?' }}</span>
                  </div>
                </div>
                <div class="asset-item">
                  <Icon name="lucide:credit-card" class="asset-icon" aria-hidden="true" />
                  <div class="asset-content">
                    <span class="asset-label">{{ $t('member.giftCards', 'Gift Cards') }}</span>
                    <span class="asset-value">{{ isLogged ? `× ${userPointCards}` : '?' }}</span>
                  </div>
                </div>
              </div>

              <div class="tier-progress" v-if="isLogged">
                <div class="progress-bar">
                  <div class="progress-fill" :style="{ width: tierInfo.pct + '%' }"></div>
                </div>
                <div class="progress-labels">
                  <span>{{ tierInfo.current ? tierInfo.current.min : 0 }}</span>
                  <span class="progress-pct">{{ tierInfo.pct }}%</span>
                  <span>{{ tierInfo.next ? tierInfo.next.min : 'MAX' }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="membership-col membership-col--full" v-if="isLogged && profileInfo">
            <div class="profile-info">
              <h4>{{ $t('profile.title', 'Profile') }}</h4>
              <div class="profile-item" v-if="profileInfo.fullName">
                <span class="profile-label">{{ $t('profile.fullName', 'Full Name') }}</span>
                <span class="profile-value">{{ profileInfo.fullName }}</span>
              </div>
              <div class="profile-item" v-if="profileInfo.company">
                <span class="profile-label">{{ $t('profile.company', 'Company') }}</span>
                <span class="profile-value">{{ profileInfo.company }}</span>
              </div>
              <div class="profile-item" v-if="profileInfo.country">
                <span class="profile-label">{{ $t('profile.country', 'Country/Region') }}</span>
                <span class="profile-value">{{ profileInfo.country }}</span>
              </div>
              <div class="profile-item" v-if="profileInfo.phone">
                <span class="profile-label">{{ $t('profile.phone', 'Phone') }}</span>
                <span class="profile-value">{{ profileInfo.phone }}</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <LazyAuthModal
        v-model="showAuthModal"
        :default-mode="authMode"
        embedded
        @mode-change="authMode = $event"
        @success="handleAuthSuccess"
      />
      </div>

    <section v-show="activeTab === 'levers'" class="company-section membership-section membership-levers">
      <div class="membership-details">
        <div class="tier-table">
          <h4>{{ $t('member.levels.title', 'Membership Levels') }}</h4>
          <div class="table-wrapper">
            <table>
              <thead>
                <tr>
                  <th>{{ $t('member.levels.header.level', 'Level') }}</th>
                  <th>{{ $t('member.levels.header.pointsRequired', 'Points Required') }}</th>
                  <th>{{ $t('member.levels.header.discountRate', 'Discount Rate') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="tier in tierConfigs" :key="tier.key">
                  <td>{{ displayTierName(tier) }}</td>
                  <td>{{ tier.min }}{{ tier.max !== null ? '–' + tier.max : '+' }}</td>
                  <td>{{ formatDiscountRate(tier.discount) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="points-rules">
          <h4>{{ $t('member.points.title', 'Points Rules') }}</h4>
          <div class="rule-list">
            <div v-for="rule in pointRuleItems" :key="rule.key" class="rule-item">
              <div class="rule-title">{{ rule.title }}</div>
              <div class="rule-desc">{{ rule.description }}</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <MembershipGiftCardExchangePanel
      v-show="activeTab === 'exchange'"
      :is-logged="isLogged"
      :points="pointsNumber"
      :redemption-rule-description="redemptionRuleDescription"
      :available-giftcards="availableGiftcards"
      :user-gift-cards="userGiftCards"
      :loading="giftcardsLoading"
      :error="giftcardsError"
      :redeeming-card-id="redeemingCardId"
      :redeem-message="redeemMessage"
      :redeem-success="redeemSuccess"
      @redeem="handleRedeemGiftcard"
    />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { useMembership } from '~/composables/useMembership'
import BadgeAvatar from '~/components/BadgeAvatar.vue'
import MembershipGiftCardExchangePanel from '~/components/membership/MembershipGiftCardExchangePanel.vue'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import {
  membershipAndPointsTabs,
  type MembershipTabId,
  type PageSubNavigationTab,
} from '~/utils/pageSubNavigation'

const props = defineProps<{ variant?: 'page' | 'modal' }>()

const isModal = computed(() => props.variant === 'modal')

const tabs: readonly PageSubNavigationTab[] = membershipAndPointsTabs
const isMembershipTabId = (id: string): id is MembershipTabId => {
  return membershipAndPointsTabs.some(tab => tab.id === id)
}

const { activeTab, setActiveTab: setPageActiveTab } = usePageSubNavigationTab({
  tabs,
  basePath: '/resources/membershipandpoints',
  defaultValue: 'myinfo',
  syncWithUrl: () => !isModal.value,
})

const setActiveTab = (id: MembershipTabId | string) => {
  if (!isMembershipTabId(id)) return
  setPageActiveTab(id)
}

const localePath = useLocalePath()
const { t, te } = useI18n()

const {
  isLogged,
  levelName,
  topTierImage,
  points,
  profileInfo,
  tierInfo,
  tierConfigs,
  levelDiscounts,
  userCoupons,
  userPointCards,
  availableGiftcards,
  userGiftCards,
  loyaltyRules,
  giftcardsLoading,
  giftcardsError,
  redeemingCardId,
  redeemMessage,
  redeemSuccess,
  handleRedeemGiftcard,
  doLogout,
  initMembership,
  refreshData
} = useMembership()

const pointsNumber = computed(() => Number(points.value ?? 0))

const normalizeLevelKey = (value: string) =>
  value.toLowerCase().trim().replace(/\s+/g, '-')

const displayTierName = (tier: { key: string; name: string }) =>
  t(`member.levels.rows.${tier.key}`, tier.name)

const displayLevelName = computed(() => {
  const rawLevel = String(levelName.value || '')
  if (!rawLevel || rawLevel === '—') return rawLevel
  return t(`member.levels.rows.${normalizeLevelKey(rawLevel)}`, rawLevel)
})

const formatBenefitNumber = (value: number) => {
  const numericValue = Number(value)
  if (!Number.isFinite(numericValue)) return '0'
  return Number.isInteger(numericValue) ? String(numericValue) : numericValue.toFixed(2).replace(/\.?0+$/, '')
}

const formatDiscountRate = (value: number) => `${formatBenefitNumber(value)}%`

const hasRuleNumber = (value: number | null | undefined): value is number =>
  typeof value === 'number' && Number.isFinite(value)

const formatRuleNumber = (value: number) =>
  Number.isInteger(value) ? String(value) : value.toFixed(2).replace(/\.?0+$/, '')

const translateWithFallback = (key: string, params: Record<string, string | number>, fallback: string) =>
  te(key) ? t(key, params) : fallback

const formatPoints = (value: number) => {
  const points = formatRuleNumber(value)
  if (value === 1) {
    return translateWithFallback('member.points.pointValueSingular', { points }, `${points} Point`)
  }
  return translateWithFallback('member.points.pointsValue', { points }, `${points} Points`)
}

const notConfiguredText = computed(() =>
  t('member.points.notConfigured', 'Not configured')
)

const pointsBaseCurrency = computed(() =>
  String(loyaltyRules.value?.points_base_currency || 'USD').trim().toUpperCase() || 'USD'
)

const purchaseEarnRuleDescription = computed(() => {
  const rules = loyaltyRules.value
  const pointsPerUnit = rules?.purchase_earn_points_per_currency_unit

  if (!hasRuleNumber(pointsPerUnit)) return notConfiguredText.value
  if (pointsPerUnit <= 0) return t('member.points.purchaseEarnDisabled', 'Purchase earning is not enabled')

  const points = formatPoints(pointsPerUnit)
  const currency = pointsBaseCurrency.value
  return translateWithFallback(
    'member.points.purchaseEarnRule',
    { points, currency },
    `${points} per 1 ${currency} of product amount after discounts, awarded after order completion`
  )
})

const referralRuleDescription = computed(() => {
  const rules = loyaltyRules.value
  const referrer = rules?.referral_referrer_points
  const referee = rules?.referral_referee_points

  if (!hasRuleNumber(referrer) || !hasRuleNumber(referee)) return notConfiguredText.value
  if (referrer <= 0 && referee <= 0) return t('member.points.referralDisabled', 'Referral points are not enabled')

  const referrerText = formatPoints(referrer)
  const refereeText = formatPoints(referee)
  return translateWithFallback(
    'member.points.referralDisplayRule',
    { referrer: referrerText, referee: refereeText },
    `Inviter gets ${referrerText}; invitee gets ${refereeText}`
  )
})

const redemptionRuleDescription = computed(() => {
  const rules = loyaltyRules.value
  const exchangeRate = rules?.redemption_exchange_rate

  if (!hasRuleNumber(exchangeRate)) return notConfiguredText.value
  if (exchangeRate <= 0) return t('member.points.redemptionDisabled', 'Points redemption is not enabled')

  const pointsText = formatPoints(exchangeRate)
  const currency = pointsBaseCurrency.value
  return translateWithFallback(
    'member.points.redemptionDisplayRule',
    { points: pointsText, amount: 1, currency },
    `${pointsText} = ${currency} 1`
  )
})

const checkInRuleDescription = computed(() => {
  const rules = loyaltyRules.value
  const base = rules?.checkin_base_points

  if (!hasRuleNumber(base)) return notConfiguredText.value
  if (base <= 0) return t('member.points.checkinDisabled', 'Daily check-in points are not enabled')

  const parts = [translateWithFallback('member.points.checkinBaseRule', { base: formatPoints(base) }, `${formatPoints(base)} per check-in`)]
  const max = rules?.checkin_max_points
  const bonus = rules?.checkin_streak_bonus_points
  const interval = rules?.checkin_streak_interval_days

  if (hasRuleNumber(max) && max > base) {
    parts.push(translateWithFallback('member.points.checkinMaxRule', { max: formatPoints(max) }, `up to ${formatPoints(max)}`))
  }

  if (hasRuleNumber(bonus) && hasRuleNumber(interval) && bonus > 0 && interval > 0) {
    parts.push(translateWithFallback(
      'member.points.checkinStreakRule',
      { bonus: formatPoints(bonus), interval: formatRuleNumber(interval) },
      `+${formatPoints(bonus)} every ${formatRuleNumber(interval)} consecutive days`
    ))
  }

  return parts.join('; ')
})

const pointRuleItems = computed(() => [
  {
    key: 'purchase',
    title: t('member.points.purchaseEarn', 'Order completion'),
    description: purchaseEarnRuleDescription.value,
  },
  {
    key: 'redemption',
    title: t('member.points.redeem', 'Redemption rate'),
    description: redemptionRuleDescription.value,
  },
  {
    key: 'referral',
    title: t('member.points.invite', 'Invite new users'),
    description: referralRuleDescription.value,
  },
  {
    key: 'checkin',
    title: t('member.points.checkin', 'Daily check-in'),
    description: checkInRuleDescription.value,
  },
])

const showAuthModal = ref(false)
const authMode = ref<'login' | 'register'>('login')

const openAuthForm = (mode: 'login' | 'register') => {
  authMode.value = mode
  showAuthModal.value = true
}

const handleAuthSuccess = async () => {
  showAuthModal.value = false
  await refreshData()
}

onMounted(() => {
  initMembership()
})
</script>

<style scoped>
 .membership-tabs {
   --membership-card-surface: var(--tz-card-surface);
   --membership-card-subtle: var(--tz-surface-subtle);
   --membership-card-border: rgba(5, 150, 105, 0.22);
   --membership-card-shadow: 0 16px 34px rgba(15, 23, 42, 0.12);
   --membership-accent: var(--tz-site-accent, #059669);
 }

 .membership-tabs--modal {
   height: 100%;
   display: flex;
   flex-direction: column;
 }

 .membership-tabs__content--scroll {
   flex: 1;
   min-height: 0;
   overflow-y: auto;
   padding-left: 12px;
   padding-right: 12px;
 }

 /* 保修查询卡片 */
 .warranty-card {
   display: flex;
   align-items: center;
   gap: 1rem;
   padding: 1rem 1.25rem;
   margin-bottom: 1.5rem;
    background: var(--membership-card-subtle);
    border: 1px solid var(--tz-border-subtle);
   border-radius: 12px;
    box-shadow: 0 3px 9px rgba(15, 23, 42, 0.1);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
   transition: all 0.2s;
 }

 .warranty-card:hover {
    box-shadow: 0 4px 12px rgba(15, 23, 42, 0.12);
 }

 .warranty-card__icon {
   font-size: 2rem;
   flex-shrink: 0;
 }

 .warranty-card__content {
   flex: 1;
   min-width: 0;
 }

 .warranty-card__title {
   margin: 0 0 0.25rem;
   font-size: 16px;
   font-weight: 600;
    color: var(--tz-text-primary);
 }

 .warranty-card__desc {
   margin: 0;
   font-size: 13px;
    color: var(--tz-text-secondary);
   line-height: 1.4;
 }

 .warranty-card__btn {
   display: inline-flex;
   align-items: center;
   gap: 0.5rem;
   padding: 0.5rem 1rem;
   background: var(--tz-site-accent);
   border-radius: 999px;
   font-size: 13px;
   font-weight: 600;
    color: #ffffff;
   text-decoration: none;
   white-space: nowrap;
   transition: all 0.2s;
   flex-shrink: 0;
 }

 .warranty-card__btn:hover {
   filter: brightness(1.1);
   transform: translateX(2px);
 }

 @media (max-width: 600px) {
   .warranty-card {
     flex-direction: column;
     text-align: center;
     gap: 0.75rem;
   }

   .warranty-card__btn {
     width: 100%;
     justify-content: center;
   }
 }

 .membership-section {
   margin-bottom: 2rem;
 }

 .membership-grid {
   display: grid;
   grid-template-columns: 380px 1fr;
   gap: 1.5rem;
   align-items: stretch;
 }

 .membership-tabs--modal .membership-grid {
   grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
 }

 .membership-col {
   min-width: 0;
 }

 .membership-col--full {
   grid-column: 1 / -1;
 }

 .member-header,
 .member-card {
   height: 100%;
 }

 @media (max-width: 900px) {
   .membership-grid {
     grid-template-columns: 1fr;
   }

   .membership-tabs--modal .membership-grid {
     grid-template-columns: 1fr;
   }
 }

 /* 会员信息区域 */
 .membership-info {
   display: flex;
   flex-direction: column;
   gap: 1rem;
 }

 .membership-info--full {
 	grid-column: 1 / -1;
 }

  .member-header {
   display: flex;
   flex-direction: column;
   align-items: center;
   justify-content: center;
   gap: 0.75rem;
   padding: 1.5rem 1rem;
    background: var(--membership-card-surface);
   border-radius: 12px;
    box-shadow: var(--membership-card-shadow);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
   text-align: center;
 }

 .member-avatar {
   width: 96px;
   height: 96px;
   display: flex;
   align-items: center;
   justify-content: center;
 }

 .member-avatar :deep(.badge) {
   width: 96px;
   height: 96px;
 }

  .member-name {
   font-size: 18px;
   font-weight: 600;
    color: var(--tz-text-primary);
 }

 .member-level {
   display: flex;
   align-items: center;
   gap: 0.75rem;
 }

 .level-badge {
   padding: 4px 12px;
   background: var(--tz-site-accent);
   border-radius: 999px;
   font-size: 12px;
   font-weight: 700;
   color: #000;
   text-transform: uppercase;
 }

  .level-points {
   font-size: 14px;
    color: var(--tz-text-secondary);
 }

 .member-actions {
   display: flex;
   gap: 0.5rem;
   flex-wrap: wrap;
   justify-content: center;
 }

 /* 按钮样式 */
  .btn-primary {
   height: 40px;
   padding: 0 20px;
   border-radius: 999px;
   background: var(--tz-site-accent);
    color: #ffffff;
   font-size: 14px;
   font-weight: 600;
   border: none;
   cursor: pointer;
   transition: all 0.2s;
 }

 .btn-primary:hover {
   filter: brightness(1.1);
 }

  .btn-secondary {
   height: 40px;
   padding: 0 20px;
   border-radius: 999px;
    background: var(--membership-card-subtle);
    color: var(--tz-text-primary);
   font-size: 14px;
   font-weight: 600;
   border: none;
   cursor: pointer;
   box-shadow:
      0 2px 6px -3px rgba(20, 32, 43, 0.12),
      0 0 6px rgba(20, 32, 43, 0.06);
   transition: all 0.2s;
 }

  .btn-secondary:hover {
    background: var(--tz-surface-muted);
   box-shadow:
      0 4px 12px -4px rgba(20, 32, 43, 0.16),
      0 0 8px rgba(20, 32, 43, 0.08);
 }

 .btn-danger {
   height: 40px;
   padding: 0 20px;
   border-radius: 999px;
   background: #dc2626;
   color: #fff;
   font-size: 14px;
   font-weight: 600;
   border: none;
   cursor: pointer;
   transition: all 0.2s;
 }

 .btn-danger:hover {
   background: #b91c1c;
 }

 /* 会员卡片 */
 .member-card {
   background: var(--membership-card-surface);
   border-radius: 12px;
   padding: 1rem;
   box-shadow: var(--membership-card-shadow);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
 }

 .member-benefits-card {
   color: var(--tz-text-primary);
 }

  .card-title {
   margin: 0 0 0.75rem;
   font-size: 14px;
   font-weight: 600;
    color: var(--tz-text-primary);
   padding-bottom: 0.75rem;
    border-bottom: 1px solid var(--tz-border-subtle);
 }

  .member-benefits-card .card-title {
   font-size: 16px;
    color: var(--tz-text-primary);
 }

 .member-stats {
   display: flex;
   flex-direction: column;
   gap: 0.5rem;
 }

 .membership-tabs--modal .member-stats {
   display: grid;
   grid-template-columns: 1fr 1fr;
 }

  .stat-item {
   display: flex;
   align-items: center;
   gap: 0.75rem;
   padding: 0.75rem;
    background: var(--membership-card-subtle);
   border-radius: 8px;
    box-shadow: 0 6px 16px rgba(20, 32, 43, 0.08);
   backdrop-filter: blur(12px);
   -webkit-backdrop-filter: blur(12px);
 }

 .stat-icon {
   font-size: 1.25rem;
   flex-shrink: 0;
 }

 .stat-content {
   display: flex;
   align-items: center;
   justify-content: space-between;
   flex: 1;
   gap: 0.5rem;
 }

 .stat-label {
   font-size: 13px;
   color: var(--tz-text-secondary);
 }

  .member-benefits-card .stat-label {
    font-size: 15px;
    color: var(--tz-text-primary);
 }

 .stat-copy {
   min-width: 0;
   display: flex;
   flex-direction: column;
   gap: 0.2rem;
 }

  .stat-desc {
   max-width: 36rem;
   font-size: 11px;
   line-height: 1.38;
    color: var(--tz-text-muted);
 }

  .member-benefits-card .stat-desc {
   font-size: 13px;
   line-height: 1.48;
    color: var(--tz-text-secondary);
 }

  .stat-value {
   font-size: 14px;
   font-weight: 600;
    color: var(--tz-text-primary);
   flex: none;
 }

  .member-benefits-card .stat-value {
   font-size: 16px;
    color: var(--tz-text-primary);
 }

 .stat-value.highlight {
   color: #059669;
 }

  .member-benefits-card .stat-value.highlight {
    color: var(--tz-text-primary);
 }

 /* 资产 */
  .member-assets {
   display: flex;
   flex-direction: column;
   gap: 0.5rem;
   margin-top: 0.75rem;
   padding-top: 0.75rem;
    border-top: 1px solid var(--tz-border-subtle);
 }

 .membership-tabs--modal .member-assets {
   display: grid;
   grid-template-columns: 1fr 1fr;
   margin-top: 1rem;
 }

 @media (max-width: 600px) {
   .membership-tabs--modal .member-stats,
   .membership-tabs--modal .member-assets {
     grid-template-columns: 1fr;
   }
 }

  .asset-item {
   display: flex;
   align-items: center;
   gap: 0.75rem;
   padding: 0.75rem;
    background: var(--membership-card-subtle);
   border-radius: 8px;
    box-shadow: 0 6px 16px rgba(20, 32, 43, 0.08);
   backdrop-filter: blur(12px);
   -webkit-backdrop-filter: blur(12px);
 }

 .asset-icon {
   font-size: 1.25rem;
   flex-shrink: 0;
 }

 .asset-content {
   display: flex;
   align-items: center;
   justify-content: space-between;
   flex: 1;
 }

  .asset-label {
   font-size: 13px;
    color: var(--tz-text-secondary);
 }

  .member-benefits-card .asset-label {
   font-size: 15px;
    color: var(--tz-text-primary);
 }

 .asset-value {
   font-size: 14px;
   font-weight: 700;
   color: var(--tz-site-accent);
   background: none;
 }

  .member-benefits-card .asset-value {
   font-size: 16px;
    color: var(--tz-text-primary);
 }

 /* 进度条 */
 .tier-progress {
   margin-top: 1rem;
 }

  .progress-bar {
   height: 8px;
    background: var(--tz-border-subtle);
   border-radius: 999px;
   overflow: hidden;
 }

 .progress-fill {
   height: 100%;
   background: var(--tz-site-accent);
   transition: width 0.3s;
 }

  .progress-labels {
   display: flex;
   justify-content: space-between;
   margin-top: 4px;
   font-size: 12px;
    color: var(--tz-text-secondary);
 }

  .progress-pct {
   font-weight: 600;
    color: var(--tz-text-primary);
 }

  .member-benefits-card .progress-labels {
   font-size: 14px;
    color: var(--tz-text-primary);
 }

  .member-benefits-card .progress-pct {
    color: var(--tz-text-primary);
 }

 /* 个人资料 */
  .profile-info {
    background: var(--membership-card-surface);
   border-radius: 12px;
   padding: 1rem;
    box-shadow: var(--membership-card-shadow);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
 }

  .profile-info h4 {
   margin: 0 0 0.75rem;
   font-size: 14px;
   font-weight: 600;
    color: var(--tz-text-primary);
 }

  .profile-item {
   display: flex;
   justify-content: space-between;
   padding: 0.5rem 0.75rem;
    background: var(--membership-card-subtle);
   border-radius: 8px;
   margin-bottom: 0.5rem;
 }

  .profile-label {
   font-size: 14px;
    color: var(--tz-text-secondary);
 }

  .profile-value {
   font-size: 14px;
    color: var(--tz-text-primary);
 }

 /* 右侧详情 */
 .membership-details {
   display: flex;
   flex-direction: column;
   gap: 1.5rem;
 }

 .membership-levers .membership-details {
   display: grid;
   grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
   align-items: stretch;
 }

 .membership-levers .tier-table,
 .membership-levers .points-rules {
   height: 100%;
 }

 @media (max-width: 900px) {
   .membership-levers .membership-details {
     grid-template-columns: 1fr;
   }

   .membership-levers .membership-details {
     justify-items: stretch;
   }

   .membership-levers .tier-table,
   .membership-levers .points-rules {
     width: 100%;
     box-sizing: border-box;
   }
 }

 /* 等级表格 */
  .tier-table {
    background: var(--membership-card-surface);
   border-radius: 12px;
   padding: 1rem;
    box-shadow: var(--membership-card-shadow);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
 }

  .tier-table h4 {
   margin: 0 0 0.75rem;
   font-size: 14px;
   font-weight: 600;
    color: var(--tz-text-primary);
 }

 .table-wrapper {
   overflow-x: auto;
 }

 .tier-table table {
   width: 100%;
   border-collapse: separate;
   border-spacing: 0 6px;
 }

 .tier-table th,
 .tier-table td {
   padding: 0.75rem;
   text-align: left;
   font-size: 13px;
 }

  .tier-table th {
    color: var(--tz-text-secondary);
   font-weight: 600;
   background: rgba(5, 150, 105, 0.08);
 }

  .tier-table td {
    color: var(--tz-text-primary);
 }

  .tier-table thead tr,
  .tier-table tbody tr {
    background: var(--membership-card-subtle);
   border-radius: 999px;
    box-shadow: 0 6px 16px rgba(20, 32, 43, 0.08);
   backdrop-filter: blur(12px);
   -webkit-backdrop-filter: blur(12px);
 }

 .tier-table thead tr th:first-child,
 .tier-table tbody tr td:first-child {
   border-top-left-radius: 999px;
   border-bottom-left-radius: 999px;
 }

 .tier-table thead tr th:last-child,
 .tier-table tbody tr td:last-child {
   border-top-right-radius: 999px;
   border-bottom-right-radius: 999px;
 }

 /* 积分规则 */
  .points-rules {
    background: var(--membership-card-surface);
   border-radius: 12px;
   padding: 1rem;
    box-shadow: var(--membership-card-shadow);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
 }

  .points-rules h4 {
   margin: 0 0 0.75rem;
   font-size: 14px;
   font-weight: 600;
    color: var(--tz-text-primary);
 }

 .rule-list {
   display: flex;
   flex-direction: column;
   gap: 0.75rem;
 }

  .rule-item {
   padding: 0.75rem;
    background: var(--membership-card-subtle);
   border-radius: 8px;
    box-shadow: 0 6px 16px rgba(20, 32, 43, 0.08);
   backdrop-filter: blur(12px);
   -webkit-backdrop-filter: blur(12px);
 }

  .rule-title {
   font-size: 13px;
   font-weight: 600;
    color: var(--tz-text-primary);
   margin-bottom: 4px;
 }

  .rule-desc {
    font-size: 13px;
    color: var(--tz-text-secondary);
 }

 .warranty-card,
 .member-header,
 .member-card,
 .stat-item,
 .asset-item,
 .profile-info,
 .tier-table,
 .tier-table thead tr,
 .tier-table tbody tr,
 .points-rules,
 .rule-item {
   background: var(--membership-card-surface);
   border: none;
   box-shadow: var(--membership-card-shadow);
 }

 .stat-item,
 .asset-item,
 .rule-item,
 .profile-item {
   background: var(--membership-card-subtle);
 }

 .warranty-card__btn,
 .btn-primary {
   background: var(--membership-accent);
  color: var(--tz-text-primary);
 }

  .level-badge {
    background: var(--tz-surface-inset);
  color: var(--tz-text-primary);
    border: 1px solid var(--tz-border-subtle);
  }

 .btn-secondary {
   background: var(--membership-card-subtle);
   border-color: var(--membership-card-border);
 }

 .tier-table th {
   background: rgba(5, 150, 105, 0.08);
   color: var(--tz-text-secondary);
 }

 .progress-fill {
   background: var(--membership-accent);
 }

 .stat-icon,
 .asset-icon,
 .warranty-card__icon {
   color: var(--tz-text-secondary);
   stroke-width: 1.8;
 }

 .company-section {
   margin-top: 2rem;
 }
</style>
