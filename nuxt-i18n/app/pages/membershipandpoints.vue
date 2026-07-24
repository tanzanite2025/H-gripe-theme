<template>
  <div class="membership-page">
    <h1 class="sr-only">{{ $t('member.pageTitle', 'Membership and Points') }}</h1>

    <MembershipAndPointsTabs />
  </div>
</template>

<script setup lang="ts">
import MembershipAndPointsTabs from '~/components/MembershipAndPointsTabs.vue'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Membership and Points',
})
</script>

<style scoped>
.membership-page {
  width: 100%;
  max-width: none;
  margin: 0 auto;
}

.company-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: #f9fafb;
}

.company-page__intro {
  margin: 0 0 1rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

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
