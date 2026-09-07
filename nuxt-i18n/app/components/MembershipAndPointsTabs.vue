<template>
  <div class="membership-tabs" :class="{ 'membership-tabs--modal': isModal, 'membership-tabs--sticky': isModal }">
    <div
      v-if="isModal"
      class="nav-pill-tabs"
      role="tablist"
      :aria-label="t('member.tabs.ariaLabel', 'Membership sections')"
    >
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
        {{ t(tab.labelKey || tab.id, tab.fallback || tab.label || tab.id) }}
      </button>
    </div>

    <div class="membership-tabs__content" :class="{ 'membership-tabs__content--scroll': isModal }">
      <MembershipMyInfoPanel
        v-if="activeTab === 'myinfo'"
        :is-modal="isModal"
        :is-logged="isLogged"
        :level-name="String(levelName)"
        :top-tier-image="String(topTierImage)"
        :points="points"
        :profile-info="profileInfo"
        :tier-info="tierInfo"
        :level-discounts="levelDiscounts"
        :user-coupons="userCoupons"
        :user-point-cards="userPointCards"
        @open-auth="openAuthForm"
        @logout="doLogout"
      />

      <MembershipLeversPanel
        v-else-if="activeTab === 'levers'"
        :tier-configs="tierConfigs"
        :loyalty-rules="loyaltyRules"
      />

      <MembershipGiftCardExchangePanel
        v-else-if="activeTab === 'exchange'"
        :is-logged="isLogged"
        :points="pointsNumber"
        :redemption-exchange-rate="loyaltyRules?.redemption_exchange_rate"
        :points-base-currency="loyaltyRules?.points_base_currency"
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

    <LazyAuthModal
      v-if="showAuthModal"
      v-model="showAuthModal"
      :default-mode="authMode"
      embedded
      @mode-change="authMode = $event"
      @success="handleAuthSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useI18n } from '#imports'
import { useMembership } from '~/composables/useMembership'
import MembershipMyInfoPanel from '~/components/membership/MembershipMyInfoPanel.vue'
import MembershipLeversPanel from '~/components/membership/MembershipLeversPanel.vue'
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

const { locale, t } = useI18n()
const pageMessagesByTab = {
  myinfo: usePageMessages('resourcesMembershipMyInfo'),
  levers: usePageMessages('resourcesMembershipLevers'),
  exchange: usePageMessages('resourcesMembershipExchange'),
} as const

const getPageMessageTab = (tabId: string): MembershipTabId =>
  isMembershipTabId(tabId) ? tabId : 'myinfo'

const loadActiveTabMessages = async (requestedLocale: string) => {
  await pageMessagesByTab[getPageMessageTab(activeTab.value)].loadPageMessages(requestedLocale)
}

await loadActiveTabMessages(locale.value)

watch([locale, activeTab], ([nextLocale]) => {
  void loadActiveTabMessages(nextLocale)
})

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

<style>
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

 .member-avatar .badge {
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
