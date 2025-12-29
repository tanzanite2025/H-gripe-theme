<template>
  <div class="membership-tabs" :class="{ 'membership-tabs--modal': isModal, 'membership-tabs--sticky': isModal }">
    <div class="nav-pill-tabs" role="tablist">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === tab.id }"
        @click="setActiveTab(tab.id)"
      >
        {{ $t(tab.labelKey, tab.fallback) }}
      </button>
    </div>

    <div class="membership-tabs__content" :class="{ 'membership-tabs__content--scroll': isModal }">
      <div v-show="activeTab === 'myinfo'">
      <div class="warranty-card">
        <div class="warranty-card__icon">🛡️</div>
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
                <span class="level-badge">{{ levelName }}</span>
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
            <div class="member-card">
              <h4 class="card-title">{{ $t('member.myBenefits', 'My Benefits') }}</h4>
              <div class="member-stats">
                <div class="stat-item">
                  <span class="stat-icon">🏷️</span>
                  <div class="stat-content">
                    <span class="stat-label">{{ $t('member.brief.level', 'Level') }}</span>
                    <span class="stat-value" :class="{ 'highlight': !isLogged }">{{ isLogged ? levelName : '?' }}</span>
                  </div>
                </div>
                <div class="stat-item">
                  <span class="stat-icon">🛍️</span>
                  <div class="stat-content">
                    <span class="stat-label">{{ $t('member.brief.productDiscount', 'Product Discount') }}</span>
                    <span class="stat-value" :class="{ 'highlight': !isLogged }">{{ isLogged ? (levelDiscounts.product + '%') : '?' }}</span>
                  </div>
                </div>
                <div class="stat-item">
                  <span class="stat-icon">💎</span>
                  <div class="stat-content">
                    <span class="stat-label">{{ $t('member.brief.pointsDiscount', 'Points') }}</span>
                    <span class="stat-value" :class="{ 'highlight': !isLogged }">{{ isLogged ? (levelDiscounts.points + '%') : '?' }}</span>
                  </div>
                </div>
                <div class="stat-item">
                  <span class="stat-icon">📊</span>
                  <div class="stat-content">
                    <span class="stat-label">{{ $t('member.brief.stackable', 'Stackable') }}</span>
                    <span class="stat-value" :class="{ 'highlight': !isLogged }">{{ isLogged ? (levelDiscounts.stackable ? '✓' : '✗') : '?' }}</span>
                  </div>
                </div>
              </div>

              <div class="member-assets">
                <div class="asset-item">
                  <span class="asset-icon">🎟️</span>
                  <div class="asset-content">
                    <span class="asset-label">{{ $t('member.coupons', 'Coupons') }}</span>
                    <span class="asset-value">{{ isLogged ? `× ${userCoupons}` : '?' }}</span>
                  </div>
                </div>
                <div class="asset-item">
                  <span class="asset-icon">💳</span>
                  <div class="asset-content">
                    <span class="asset-label">{{ $t('member.pointCards', 'Point Cards') }}</span>
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

      <AuthModal
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
                  <th>{{ $t('member.levels.header.productDiscount', 'Product Discount') }}</th>
                  <th>{{ $t('member.levels.header.pointsDiscount', 'Points Discount') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="tier in tierConfigs" :key="tier.key">
                  <td>{{ tier.name }}</td>
                  <td>{{ tier.min }}{{ tier.max !== null ? '–' + tier.max : '+' }}</td>
                  <td>{{ tier.discount }}%</td>
                  <td>{{ tier.pointsDiscount }}%</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="points-rules">
          <h4>{{ $t('member.points.title', 'How to Earn Points?') }}</h4>
          <div class="rule-list">
            <div class="rule-item">
              <div class="rule-title">{{ $t('member.points.invite', 'Invite new users') }}</div>
              <div class="rule-desc">{{ $t('member.points.inviteDesc', '50 Points (invitee gets 30 Points)') }}</div>
            </div>
            <div class="rule-item invite-action">
              <button class="btn-gradient" @click="handleCopyInviteLink" :disabled="inviteLoading">
                {{ inviteLoading ? '...' : $t('member.copyLink', 'Copy Invite Link') }}
              </button>
              <span class="invite-msg" v-if="inviteMsg">{{ inviteMsg }}</span>
            </div>
            <div class="rule-item">
              <div class="rule-title">{{ $t('member.points.consume', 'Consumption') }}</div>
              <div class="rule-desc">{{ $t('member.points.consumeDesc', '1 Dollar = 1 Point') }}</div>
            </div>
            <div class="rule-item">
              <div class="rule-title">{{ $t('member.points.daily', 'Daily login') }}</div>
              <div class="rule-desc">{{ $t('member.points.dailyDesc', '1 Point (30 days validity)') }}</div>
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
                <span class="giftcard-icon">💳</span>
                <div class="giftcard-info">
                  <div class="giftcard-code">{{ card.card_code }}</div>
                  <div class="giftcard-label">{{ $t('giftcards.balance', 'Balance') }}</div>
                </div>
                <div class="giftcard-value">${{ card.balance }}</div>
              </div>
              <div class="giftcard-footer">
                <span class="giftcard-points">
                  {{ $t('giftcards.pointsRequired', 'Points required') }}: {{ card.points_spent || 0 }}
                </span>
                <button
                  class="btn-redeem"
                  @click="handleRedeemGiftcard(card)"
                  :disabled="(isLogged && pointsNumber < (card.points_spent || 0)) || redeemingCardId === card.id"
                >
                  {{ redeemingCardId === card.id ? $t('giftcards.redeeming', 'Redeeming...') : $t('giftcards.redeem', 'Redeem') }}
                </button>
              </div>
            </div>
          </div>

          <div v-else class="empty-state">
            {{ $t('giftcards.noCards', 'No gift cards available') }}
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
import { useLocalePath } from '#imports'
import { useMembership } from '~/composables/useMembership'
import BadgeAvatar from '~/components/BadgeAvatar.vue'
import AuthModal from '~/components/AuthModal.vue'

const props = defineProps<{ variant?: 'page' | 'modal' }>()

const isModal = computed(() => props.variant === 'modal')

type MembershipTabId = 'myinfo' | 'levers' | 'exchange'

const tabs: { id: MembershipTabId; labelKey: string; fallback: string }[] = [
  { id: 'myinfo', labelKey: 'member.tabs.myInfo', fallback: 'My info' },
  { id: 'levers', labelKey: 'member.tabs.levers', fallback: 'Levers' },
  { id: 'exchange', labelKey: 'member.tabs.exchange', fallback: 'Exchange' },
]

const activeTab = ref<MembershipTabId>('myinfo')

const setActiveTab = (id: MembershipTabId) => {
  activeTab.value = id
}

const localePath = useLocalePath()

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
  giftcardsLoading,
  giftcardsError,
  redeemingCardId,
  redeemMessage,
  redeemSuccess,
  inviteLoading,
  inviteMsg,
  handleRedeemGiftcard,
  handleCopyInviteLink,
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

<style scoped>
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

 .membership-tabs--sticky .company-tabs {
   position: sticky;
   top: 0;
   z-index: 30;
   background: linear-gradient(180deg, rgba(15, 23, 42, 0.9), rgba(2, 6, 23, 0.7));
   backdrop-filter: blur(12px);
   -webkit-backdrop-filter: blur(12px);
   padding: 10px 56px 10px 16px;
   margin: 0 0 1rem;
   border: none;
   border-radius: 14px;
   box-shadow: 8px 8px 22px rgba(0, 0, 0, 0.92);
 }

 .membership-tabs--sticky .company-tabs {
   padding: 10px 56px 10px 16px;
   margin: 0 0 1rem;
   max-width: 100%;
 }
 /* Style clean up: .company-tabs and .company-tabs__item definitions are now in global nav.css */

 /* 保修查询卡片 */
 .warranty-card {
   display: flex;
   align-items: center;
   gap: 1rem;
   padding: 1rem 1.25rem;
   margin-bottom: 1.5rem;
   background: linear-gradient(135deg, rgba(64, 255, 170, 0.1), rgba(107, 115, 255, 0.1));
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
   background: linear-gradient(to right, #40ffaa, #6b73ff);
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
   background: linear-gradient(to right, #40ffaa, #6b73ff);
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
   background: linear-gradient(to right, #40ffaa, #6b73ff);
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
   background: linear-gradient(to right, #40ffaa, #6b73ff);
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

 .card-title {
   margin: 0 0 0.75rem;
   font-size: 14px;
   font-weight: 600;
   color: rgba(255, 255, 255, 0.9);
   padding-bottom: 0.75rem;
   border-bottom: 1px solid rgba(255, 255, 255, 0.1);
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

 .stat-value {
   font-size: 14px;
   font-weight: 600;
   color: rgba(255, 255, 255, 0.9);
 }

 .stat-value.highlight {
   color: #40ffaa;
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

 .asset-value {
   font-size: 14px;
   font-weight: 700;
   background: linear-gradient(to right, #40ffaa, #6b73ff);
   -webkit-background-clip: text;
   -webkit-text-fill-color: transparent;
   background-clip: text;
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
   background: linear-gradient(to right, #40ffaa, #6b73ff);
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
   background: rgba(110, 110, 233, 0.1);
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

 .invite-action {
   display: flex;
   align-items: center;
   gap: 0.75rem;
 }

 .invite-msg {
   font-size: 12px;
   color: #cfd6ff;
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
   background: linear-gradient(to right, #40ffaa, #6b73ff);
   -webkit-background-clip: text;
   -webkit-text-fill-color: transparent;
   background-clip: text;
 }

 .giftcard-footer {
   display: flex;
   align-items: center;
   justify-content: space-between;
   padding-top: 0.75rem;
   border-top: 1px solid rgba(255, 255, 255, 0.1);
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
   background: linear-gradient(to right, #40ffaa, #6b73ff);
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
   background: rgba(16, 185, 129, 0.2);
   color: #6ee7b7;
 }

 .redeem-message.error {
   background: rgba(239, 68, 68, 0.2);
   color: #fca5a5;
 }

 .company-section {
   margin-top: 2rem;
 }
</style>
