<template>
  <div class="membership-page">
    <h2 class="sr-only">{{ $t('member.pageTitle', 'Membership and Points') }}</h2>
    <p class="company-page__intro">
      {{ $t('member.pageIntro', 'Manage your membership, view benefits, and redeem points.') }}
    </p>

    <!-- 保修查询入口卡片 -->
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

    <!-- 会员信息区域 -->
    <section class="membership-section">
      <div class="membership-grid">
        <!-- 左侧：会员信息 -->
        <div class="membership-info">
          <!-- 头像和登录状态 -->
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

          <!-- 会员等级卡片 -->
          <div class="member-card">
            <h4 class="card-title">{{ $t('member.myBenefits', 'My Benefits') }}</h4>
            <div class="member-stats">
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

            <!-- 优惠券和积分卡 -->
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

            <!-- 等级进度条 -->
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

          <!-- 个人资料 -->
          <div class="profile-info" v-if="isLogged && profileInfo">
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

        <!-- 右侧：等级说明和积分规则 -->
        <div class="membership-details">
          <!-- 等级说明 -->
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

          <!-- 积分规则 -->
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

          <!-- 礼品卡兑换 -->
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
                    :disabled="(isLogged && points < (card.points_spent || 0)) || redeemingCardId === card.id"
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
      </div>
    </section>

    <!-- FAQ Section -->
    <section class="company-section">
      <PageFaq 
        page-id="company-membership"
        theme="dark"
        :show-categories="true"
      />
    </section>

    <!-- Auth Modal -->
    <AuthModal
      v-model="showAuthModal"
      :default-mode="authMode"
      embedded
      @mode-change="authMode = $event"
      @success="handleAuthSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useLocalePath } from '#imports'
import { useMembership } from '~/composables/useMembership'
import PageFaq from '~/components/PageFaq.vue'
import BadgeAvatar from '~/components/BadgeAvatar.vue'
import AuthModal from '~/components/AuthModal.vue'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Membership and Points',
})

const localePath = useLocalePath()

// 使用会员 composable
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

// 认证弹窗
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
.membership-page {
  max-width: 1200px;
  margin: 0 auto;
}

.company-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.company-page__intro {
  margin: 0 0 1rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

/* 保修查询卡片 */
.warranty-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  margin-bottom: 1.5rem;
  background: linear-gradient(135deg, rgba(64, 255, 170, 0.1), rgba(107, 115, 255, 0.1));
  border: 1px solid rgba(107, 115, 255, 0.3);
  border-radius: 12px;
  transition: all 0.2s;
}

.warranty-card:hover {
  border-color: rgba(107, 115, 255, 0.5);
  box-shadow: 0 4px 20px rgba(107, 115, 255, 0.15);
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

.warranty-card__btn .arrow {
  transition: transform 0.2s;
}

.warranty-card__btn:hover .arrow {
  transform: translateX(3px);
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
  align-items: start;
}

@media (max-width: 900px) {
  .membership-grid {
    grid-template-columns: 1fr;
  }
}

/* 会员信息区域 */
.membership-info {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.member-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 1.5rem 1rem;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
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
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  border: 1px solid rgba(255, 255, 255, 0.2);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.2);
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
  background: rgba(15, 23, 42, 0.9);
  border-radius: 12px;
  padding: 1rem;
  box-shadow:
    0 10px 26px -14px rgba(0, 0, 0, 0.95),
    0 0 14px rgba(15, 23, 42, 0.9);
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

.stat-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
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

.asset-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
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
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1rem;
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

/* 等级表格 */
.tier-table {
  background: rgba(15, 23, 42, 0.9);
  border-radius: 12px;
  padding: 1rem;
  box-shadow:
    0 10px 26px -14px rgba(0, 0, 0, 0.95),
    0 0 14px rgba(15, 23, 42, 0.9);
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
  border-collapse: collapse;
}

.tier-table th,
.tier-table td {
  padding: 0.75rem;
  text-align: left;
  font-size: 13px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.tier-table th {
  color: rgba(255, 255, 255, 0.7);
  font-weight: 600;
  background: rgba(110, 110, 233, 0.1);
}

.tier-table td {
  color: rgba(255, 255, 255, 0.9);
}

.tier-table tr:last-child td {
  border-bottom: none;
}

/* 积分规则 */
.points-rules {
  background: rgba(15, 23, 42, 0.9);
  border-radius: 12px;
  padding: 1rem;
  box-shadow:
    0 10px 26px -14px rgba(0, 0, 0, 0.95),
    0 0 14px rgba(15, 23, 42, 0.9);
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
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
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
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1rem;
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
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 0.75rem;
  transition: all 0.2s;
}

.giftcard-item:hover {
  border-color: rgba(107, 115, 255, 0.5);
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
