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
          <p class="warranty-card__desc">{{ $t('warranty.cardDesc', 'Enter your product code to check warranty status and history.') }}</p>
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

    <section v-show="activeTab === 'exchange'" class="company-section membership-section">
      <div class="membership-details">
        <div class="giftcard-section">
          <h4>{{ $t('giftcards.title', 'Redeem Points for Gift Cards') }}</h4>

          <div v-if="giftcardsLoading" class="loading-state">
            {{ $t('common.loading', 'Loading...') }}
          </div>

          <div v-else-if="giftcardsError" class="error-state">
            {{ giftcardsError }}
          </div>

          <div v-else-if="availableGiftcards.length > 0" class="giftcard-grid">
            <div v-for="card in availableGiftcards" :key="card.id" class="giftcard-item">
              <div class="giftcard-header">
                <Icon name="lucide:credit-card" class="giftcard-icon" aria-hidden="true" />
                <div class="giftcard-info">
                  <div class="giftcard-code">{{ card.label }}</div>
                  <div class="giftcard-label">{{ $t('giftcards.balance', 'Balance') }}</div>
                </div>
                <div class="giftcard-value">{{ card.currency }} {{ card.giftcard_value }}</div>
              </div>
              <div class="giftcard-footer">
                <span class="giftcard-points">
                  {{ $t('giftcards.pointsRequired', 'Points required') }}: {{ card.points_required || 0 }}
                  <span class="giftcard-stock">· {{ $t('giftcards.remaining', 'Remaining') }}: {{ card.remaining_quantity }}</span>
                </span>
                <button
                  class="btn-redeem"
                  @click="handleRedeemGiftcard(card)"
                  :disabled="(isLogged && pointsNumber < (card.points_required || 0)) || redeemingCardId === card.id"
                >
                  {{ redeemingCardId === card.id ? $t('giftcards.redeeming', 'Redeeming...') : $t('giftcards.redeem', 'Redeem') }}
                </button>
              </div>
            </div>
          </div>

          <div v-else class="empty-state">
            {{ $t('giftcards.noCards', 'No gift cards available') }}
          </div>

          <div v-if="isLogged && userGiftCards.length > 0" class="owned-giftcard-list">
            <h5>{{ $t('giftcards.myCards', 'My Gift Cards') }}</h5>
            <div v-for="card in userGiftCards" :key="card.id" class="owned-giftcard-item">
              <div>
                <strong>{{ card.code }}</strong>
                <span>{{ card.currency }} {{ Number(card.balance ?? 0).toFixed(2) }}</span>
              </div>
              <small>{{ card.status }}</small>
            </div>
          </div>

          <div v-if="redeemMessage" class="redeem-message" :class="{ success: redeemSuccess, error: !redeemSuccess }">
            {{ redeemMessage }}
          </div>
        </div>
      </div>
    </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { useMembership } from '~/composables/useMembership'
import BadgeAvatar from '~/components/BadgeAvatar.vue'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import { membershipAndPointsTabs, type MembershipTabId } from '~/utils/pageSubNavigation'

const props = defineProps<{ variant?: 'page' | 'modal' }>()

const isModal = computed(() => props.variant === 'modal')

const tabs = membershipAndPointsTabs
const isMembershipTabId = (id: string): id is MembershipTabId => {
  return membershipAndPointsTabs.some(tab => tab.id === id)
}

const { activeTab, setActiveTab: setPageActiveTab } = usePageSubNavigationTab({
  tabs,
  basePath: '/membershipandpoints',
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
   --membership-card-surface: var(--tz-card-surface, #111116);
   --membership-card-subtle: rgba(255, 255, 255, 0.035);
   --membership-card-border: rgba(181, 255, 109, 0.14);
   --membership-card-shadow: 0 16px 34px rgba(0, 0, 0, 0.34);
   --membership-accent: var(--tz-brand-primary, #b5ff6d);
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
   background: rgba(255, 255, 255, 0.035);
   border: none;
   border-radius: 12px;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
   transition: all 0.2s;
 }

 .warranty-card:hover {
   box-shadow: 0 4px 12px rgba(0, 0, 0, 0.9);
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
   color: #fff;
 }

 .warranty-card__desc {
   margin: 0;
   font-size: 13px;
   color: rgba(255, 255, 255, 0.6);
   line-height: 1.4;
 }

 .warranty-card__btn {
   display: inline-flex;
   align-items: center;
   gap: 0.5rem;
   padding: 0.5rem 1rem;
   background: var(--tz-brand-primary);
   border-radius: 999px;
   font-size: 13px;
   font-weight: 600;
   color: #000;
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
   background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
   border-radius: 12px;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
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
   color: #fff;
 }

 .member-level {
   display: flex;
   align-items: center;
   gap: 0.75rem;
 }

 .level-badge {
   padding: 4px 12px;
   background: var(--tz-brand-primary);
   border-radius: 999px;
   font-size: 12px;
   font-weight: 700;
   color: #000;
   text-transform: uppercase;
 }

 .level-points {
   font-size: 14px;
   color: rgba(255, 255, 255, 0.7);
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
   background: var(--tz-brand-primary);
   color: #000;
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
   background: linear-gradient(135deg, rgba(15, 23, 42, 0.98), rgba(15, 23, 42, 0.96));
   color: #fff;
   font-size: 14px;
   font-weight: 600;
   border: none;
   cursor: pointer;
   box-shadow:
     0 2px 6px -3px rgba(0, 0, 0, 0.9),
     0 0 6px rgba(0, 0, 0, 0.7);
   transition: all 0.2s;
 }

 .btn-secondary:hover {
   background: linear-gradient(135deg, rgba(31, 41, 55, 0.98), rgba(15, 23, 42, 0.98));
   box-shadow:
     0 4px 12px -4px rgba(0, 0, 0, 0.95),
     0 0 8px rgba(0, 0, 0, 0.9);
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

 .btn-gradient {
   height: 40px;
   padding: 0 18px;
   border-radius: 999px;
   background: var(--tz-brand-primary);
   color: #fff;
   font-size: 14px;
   font-weight: 700;
   border: none;
   cursor: pointer;
   transition: all 0.2s;
 }

 .btn-gradient:hover {
   filter: brightness(1.1);
 }

 .btn-gradient:disabled {
   opacity: 0.5;
   cursor: not-allowed;
 }

 /* 会员卡片 */
 .member-card {
   background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
   border-radius: 12px;
   padding: 1rem;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
 }

 .member-benefits-card {
   color: #fff;
 }

 .card-title {
   margin: 0 0 0.75rem;
   font-size: 14px;
   font-weight: 600;
   color: rgba(255, 255, 255, 0.9);
   padding-bottom: 0.75rem;
   border-bottom: 1px solid rgba(255, 255, 255, 0.1);
 }

 .member-benefits-card .card-title {
   font-size: 16px;
   color: #fff;
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
   background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
   border-radius: 8px;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
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
   color: rgba(255, 255, 255, 0.7);
 }

 .member-benefits-card .stat-label {
   font-size: 15px;
   color: #fff;
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
   color: rgba(255, 255, 255, 0.52);
 }

 .member-benefits-card .stat-desc {
   font-size: 13px;
   line-height: 1.48;
   color: #fff;
 }

 .stat-value {
   font-size: 14px;
   font-weight: 600;
   color: rgba(255, 255, 255, 0.9);
   flex: none;
 }

 .member-benefits-card .stat-value {
   font-size: 16px;
   color: #fff;
 }

 .stat-value.highlight {
   color: #B5FF6D;
 }

 .member-benefits-card .stat-value.highlight {
   color: #fff;
 }

 /* 资产 */
 .member-assets {
   display: flex;
   flex-direction: column;
   gap: 0.5rem;
   margin-top: 0.75rem;
   padding-top: 0.75rem;
   border-top: 1px solid rgba(255, 255, 255, 0.1);
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
   background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
   border-radius: 8px;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
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
   color: rgba(255, 255, 255, 0.7);
 }

 .member-benefits-card .asset-label {
   font-size: 15px;
   color: #fff;
 }

 .asset-value {
   font-size: 14px;
   font-weight: 700;
   color: var(--tz-brand-primary);
   background: none;
 }

 .member-benefits-card .asset-value {
   font-size: 16px;
   color: #fff;
 }

 /* 进度条 */
 .tier-progress {
   margin-top: 1rem;
 }

 .progress-bar {
   height: 8px;
   background: rgba(255, 255, 255, 0.1);
   border-radius: 999px;
   overflow: hidden;
 }

 .progress-fill {
   height: 100%;
   background: var(--tz-brand-primary);
   transition: width 0.3s;
 }

 .progress-labels {
   display: flex;
   justify-content: space-between;
   margin-top: 4px;
   font-size: 12px;
   color: rgba(255, 255, 255, 0.7);
 }

 .progress-pct {
   font-weight: 600;
   color: rgba(255, 255, 255, 0.9);
 }

 .member-benefits-card .progress-labels {
   font-size: 14px;
   color: #fff;
 }

 .member-benefits-card .progress-pct {
   color: #fff;
 }

 /* 个人资料 */
 .profile-info {
   background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
   border-radius: 12px;
   padding: 1rem;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
 }

 .profile-info h4 {
   margin: 0 0 0.75rem;
   font-size: 14px;
   font-weight: 600;
   color: rgba(255, 255, 255, 0.9);
 }

 .profile-item {
   display: flex;
   justify-content: space-between;
   padding: 0.5rem 0.75rem;
   background: rgba(255, 255, 255, 0.05);
   border-radius: 8px;
   margin-bottom: 0.5rem;
 }

 .profile-label {
   font-size: 14px;
   color: rgba(255, 255, 255, 0.7);
 }

 .profile-value {
   font-size: 14px;
   color: rgba(255, 255, 255, 0.9);
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
   background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
   border-radius: 12px;
   padding: 1rem;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
 }

 .tier-table h4 {
   margin: 0 0 0.75rem;
   font-size: 14px;
   font-weight: 600;
   color: rgba(255, 255, 255, 0.9);
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
   color: rgba(255, 255, 255, 0.7);
   font-weight: 600;
   background: rgba(181, 255, 109, 0.08);
 }

 .tier-table td {
   color: rgba(255, 255, 255, 0.9);
 }

 .tier-table thead tr,
 .tier-table tbody tr {
   background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.98), rgba(15, 23, 42, 0.98));
   border-radius: 999px;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
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
   background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
   border-radius: 12px;
   padding: 1rem;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
 }

 .points-rules h4 {
   margin: 0 0 0.75rem;
   font-size: 14px;
   font-weight: 600;
   color: rgba(255, 255, 255, 0.9);
 }

 .rule-list {
   display: flex;
   flex-direction: column;
   gap: 0.75rem;
 }

 .rule-item {
   padding: 0.75rem;
   background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
   border-radius: 8px;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
   backdrop-filter: blur(12px);
   -webkit-backdrop-filter: blur(12px);
 }

 .rule-title {
   font-size: 13px;
   font-weight: 600;
   color: rgba(255, 255, 255, 0.85);
   margin-bottom: 4px;
 }

 .rule-desc {
   font-size: 13px;
   color: rgba(255, 255, 255, 0.7);
 }

 /* 礼品卡 */
 .giftcard-section {
   background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
   border-radius: 12px;
   padding: 1rem;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
   backdrop-filter: blur(14px);
   -webkit-backdrop-filter: blur(14px);
 }

 .giftcard-section h4 {
   margin: 0 0 0.75rem;
   font-size: 14px;
   font-weight: 600;
   color: rgba(255, 255, 255, 0.9);
 }

 .giftcard-grid {
   display: flex;
   flex-direction: column;
   gap: 0.75rem;
 }

 .giftcard-item {
   background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
   border-radius: 12px;
   padding: 0.75rem;
   box-shadow: 0 3px 9px rgba(0, 0, 0, 0.9);
   backdrop-filter: blur(12px);
   -webkit-backdrop-filter: blur(12px);
   transition: all 0.2s;
 }

 .giftcard-item:hover {
   box-shadow: 0 4px 12px rgba(0, 0, 0, 0.9);
 }

 .giftcard-header {
   display: flex;
   align-items: center;
   gap: 0.75rem;
   margin-bottom: 0.75rem;
 }

 .giftcard-icon {
   font-size: 1.5rem;
 }

 .giftcard-info {
   flex: 1;
 }

 .giftcard-code {
   font-size: 14px;
   font-weight: 600;
   color: rgba(255, 255, 255, 0.9);
 }

 .giftcard-label {
   font-size: 12px;
   color: rgba(255, 255, 255, 0.5);
 }

 .giftcard-value {
   font-size: 18px;
   font-weight: 700;
   color: var(--tz-brand-primary);
   background: none;
 }

.giftcard-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 0.75rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.owned-giftcard-list {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.owned-giftcard-list h5 {
  margin: 0 0 0.65rem;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.84);
}

.owned-giftcard-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.65rem 0;
  color: rgba(255, 255, 255, 0.72);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.owned-giftcard-item > div {
  display: grid;
  gap: 0.2rem;
}

.owned-giftcard-item strong {
  color: rgba(255, 255, 255, 0.92);
  font-size: 12px;
}

.owned-giftcard-item span,
.owned-giftcard-item small {
  font-size: 11px;
}

 .giftcard-points {
   font-size: 12px;
   color: rgba(255, 255, 255, 0.7);
 }

 .btn-redeem {
   padding: 6px 12px;
   font-size: 12px;
   font-weight: 600;
   border-radius: 8px;
   background: var(--tz-brand-primary);
   color: #fff;
   border: none;
   cursor: pointer;
   transition: all 0.2s;
 }

 .btn-redeem:hover {
   filter: brightness(1.1);
 }

 .btn-redeem:disabled {
   opacity: 0.5;
   cursor: not-allowed;
 }

 /* 状态 */
 .loading-state,
 .error-state,
 .empty-state {
   text-align: center;
   padding: 1.5rem;
   font-size: 14px;
   color: rgba(255, 255, 255, 0.5);
 }

 .error-state {
   color: #f87171;
 }

 .redeem-message {
   margin-top: 0.75rem;
   padding: 0.5rem;
   border-radius: 8px;
   text-align: center;
   font-size: 14px;
 }

 .redeem-message.success {
   background: rgba(181, 255, 109, 0.2);
   color: #6ee7b7;
 }

 .redeem-message.error {
   background: rgba(239, 68, 68, 0.2);
   color: #fca5a5;
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
 .rule-item,
 .giftcard-section,
 .giftcard-item {
   background: var(--membership-card-surface);
   border: none;
   box-shadow: var(--membership-card-shadow);
 }

 .stat-item,
 .asset-item,
 .rule-item,
 .profile-item,
 .giftcard-item {
   background: var(--membership-card-subtle);
 }

 .warranty-card__btn,
 .btn-primary,
 .btn-redeem {
   background: var(--membership-accent);
   color: #050505;
 }

 .level-badge {
   background: #f8fafc;
   color: #050505;
   border: 1px solid rgba(255, 255, 255, 0.72);
 }

 .btn-secondary {
   background: var(--membership-card-subtle);
   border-color: var(--membership-card-border);
 }

 .tier-table th {
   background: rgba(181, 255, 109, 0.08);
   color: var(--tz-text-secondary);
 }

 .progress-fill {
   background: var(--membership-accent);
 }

 .stat-icon,
 .asset-icon,
 .giftcard-icon,
 .warranty-card__icon {
   color: var(--tz-text-secondary);
   stroke-width: 1.8;
 }

 .company-section {
   margin-top: 2rem;
 }
</style>
