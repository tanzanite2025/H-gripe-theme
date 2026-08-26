<template>
  <div class="warranty-check__container">
    <!-- 标题 -->
    <div class="warranty-check__header">
      <h1 class="warranty-check__title">{{ $t('warranty.title') }}</h1>
      <p class="warranty-check__subtitle">{{ $t('warranty.subtitle') }}</p>
    </div>

    <!-- 未登录状态 -->
    <div v-if="!isLoggedIn" class="warranty-check__login-required">
      <div class="warranty-check__lock-icon">🔒</div>
      <p>{{ $t('warranty.login_required') }}</p>
      <button type="button" class="warranty-check__login-btn" @click="handleLoginClick">
        {{ $t('warranty.login_button') }}
      </button>
    </div>

    <!-- 已登录 - 查询表单 -->
    <div v-else class="warranty-check__content">
      <!-- 查询表单 -->
      <div class="warranty-check__form">
        <label for="order-number" class="warranty-check__label">
          {{ $t('warranty.input_label') }}
        </label>
        <div class="warranty-check__input-group">
          <input
            id="order-number"
            v-model="orderNumber"
            type="text"
            class="warranty-check__input"
            :placeholder="$t('warranty.input_placeholder')"
            @keypress.enter="checkWarranty"
          />
          <button
            type="button"
            class="warranty-check__submit-btn"
            :disabled="loading || !orderNumber.trim()"
            @click="checkWarranty"
          >
            <span v-if="loading" class="warranty-check__spinner"></span>
            <span v-else>{{ $t('warranty.check_button') }}</span>
          </button>
        </div>
        <p class="warranty-check__help">{{ $t('warranty.help_text') }}</p>
      </div>

      <!-- 错误提示 -->
      <div v-if="error" class="warranty-check__error">
        <div class="warranty-check__error-icon">❌</div>
        <h3>{{ $t('warranty.result.not_found') }}</h3>
        <p>{{ $t('warranty.errors.not_found_message', { code: searchedOrderNumber }) }}</p>
        <ul class="warranty-check__tips">
          <li>{{ $t('warranty.errors.check_tips.0') }}</li>
          <li>{{ $t('warranty.errors.check_tips.1') }}</li>
        </ul>
        <p class="warranty-check__error-contact">{{ $t('warranty.errors.error_contact') }}</p>
        <NuxtLink :to="localePath('/company/contact')" class="warranty-check__contact-btn">
          {{ $t('warranty.actions.contact_support') }}
        </NuxtLink>
      </div>

      <!-- 查询结果 -->
      <div v-if="result" class="warranty-check__result">
        <!-- 状态标题 -->
        <div
          class="warranty-check__status"
          :class="result.status === 'valid' ? 'warranty-check__status--valid' : 'warranty-check__status--expired'"
        >
          <span class="warranty-check__status-icon">{{ result.status === 'valid' ? '✅' : '❌' }}</span>
          <span>
            {{ result.status === 'valid' ? $t('warranty.result.valid') : $t('warranty.result.expired') }}
          </span>
        </div>

        <!-- 产品信息 -->
        <div class="warranty-check__info">
          <div class="warranty-check__info-row">
            <span class="warranty-check__info-label">{{ $t('warranty.fields.order_number') }}</span>
            <span class="warranty-check__info-value">{{ result.order_number || '-' }}</span>
          </div>
          <div class="warranty-check__info-row">
            <span class="warranty-check__info-label">{{ $t('warranty.fields.product_type') }}</span>
            <span class="warranty-check__info-value">
              {{ useChineseWarrantyLabels ? result.product_type.name_zh : result.product_type.name }}
            </span>
          </div>
          <div v-if="result.product_name" class="warranty-check__info-row">
            <span class="warranty-check__info-label">{{ $t('warranty.fields.product_name') }}</span>
            <span class="warranty-check__info-value">{{ result.product_name }}</span>
          </div>
          <div class="warranty-check__info-row">
            <span class="warranty-check__info-label">{{ $t('warranty.fields.ship_date') }}</span>
            <span class="warranty-check__info-value">{{ formatDate(result.ship_date) }}</span>
          </div>
          <div class="warranty-check__info-row">
            <span class="warranty-check__info-label">{{ $t('warranty.fields.warranty_period') }}</span>
            <span class="warranty-check__info-value">{{ result.warranty_months }} {{ $t('warranty.months') }}</span>
          </div>
          <div class="warranty-check__info-row">
            <span class="warranty-check__info-label">{{ $t('warranty.fields.warranty_until') }}</span>
            <span class="warranty-check__info-value">{{ formatDate(result.warranty_end) }}</span>
          </div>
        </div>

        <!-- 剩余时间 -->
        <div
          class="warranty-check__remaining"
          :class="result.status === 'valid' ? 'warranty-check__remaining--valid' : 'warranty-check__remaining--expired'"
        >
          <span class="warranty-check__remaining-icon">⏱️</span>
          <span v-if="result.status === 'valid'">
            {{ $t('warranty.fields.remaining') }}:
            {{ result.remaining.months }} {{ $t('warranty.months') }}
            {{ result.remaining.days }} {{ $t('warranty.days') }}
          </span>
          <span v-else>
            {{ $t('warranty.fields.expired_ago', { days: result.remaining.expired_days }) }}
          </span>
        </div>

        <!-- 服务记录 -->
        <div v-if="result.records && result.records.length > 0" class="warranty-check__records">
          <h4>{{ $t('warranty.records.title') }}</h4>
          <ul class="warranty-check__records-list">
            <li
              v-for="record in result.records"
              :key="record.date"
              class="warranty-check__record-item"
            >
              <span class="warranty-check__record-type">
                {{ useChineseWarrantyLabels ? record.type_name_zh : record.type_name }}
              </span>
              <span class="warranty-check__record-date">{{ record.date }}</span>
              <span v-if="record.description" class="warranty-check__record-desc">
                {{ record.description }}
              </span>
            </li>
          </ul>
        </div>
        <div v-else class="warranty-check__no-records">
          <p>{{ $t('warranty.records.no_records') }}</p>
        </div>

        <!-- 操作按钮 -->
        <div class="warranty-check__actions">
          <button type="button" class="warranty-check__action-btn" @click="resetForm">
            {{ $t('warranty.actions.check_another') }}
          </button>
          <NuxtLink
            :to="localePath('/company/contact')"
            class="warranty-check__action-btn warranty-check__action-btn--secondary"
          >
            {{ $t('warranty.actions.contact_support') }}
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useWarrantyCheck } from '~/composables/useWarrantyCheck'
import { isSimplifiedChineseStorefrontLocale } from '~/utils/storefrontLocales'

const props = defineProps<{
  isLoggedIn: boolean
}>()

const emit = defineEmits<{
  'login-request': []
}>()

const { locale } = useI18n()
const localePath = useLocalePath()
const useChineseWarrantyLabels = computed(() => isSimplifiedChineseStorefrontLocale(locale.value))

const {
  orderNumber,
  searchedOrderNumber,
  loading,
  error,
  result,
  checkWarranty,
  reset,
  formatDate,
} = useWarrantyCheck()

const handleLoginClick = () => {
  emit('login-request')
}

const resetForm = () => {
  reset()
}
</script>

<style scoped>
.warranty-check__container {
  width: 100%;
  max-width: none;
  margin: 0 auto;
}

.warranty-check__header {
  text-align: center;
  margin-bottom: 2rem;
}

.warranty-check__title {
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: var(--tz-text-primary);
  margin: 0 0 0.5rem;
}

.warranty-check__subtitle {
  font-size: 0.9rem;
  color: var(--tz-text-secondary);
  margin: 0;
}

/* 未登录状态 */
.warranty-check__login-required {
  text-align: center;
  padding: 3rem 2rem;
  background: var(--tz-surface-subtle);
  border-radius: 1rem;
  border: none;
  box-shadow:
    0 6px 20px rgba(20, 32, 43, 0.08),
    0 14px 40px rgba(20, 32, 43, 0.08);
}

.warranty-check__lock-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.warranty-check__login-required p {
  color: var(--tz-text-secondary);
  margin-bottom: 1.5rem;
}

.warranty-check__login-btn {
  padding: 0.75rem 2rem;
  background: var(--tz-action-primary);
  border: 1px solid var(--tz-action-primary);
  border-radius: 9999px;
  color: var(--tz-action-primary-foreground);
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 12px rgb(15 23 42 / 0.16);
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}

.warranty-check__login-btn:hover {
  background: var(--tz-action-primary-hover);
  border-color: var(--tz-action-primary-hover);
  transform: translateY(-2px);
  box-shadow: 0 6px 18px rgb(15 23 42 / 0.22);
}

/* 查询表单 */
.warranty-check__form {
  margin-bottom: 2rem;
}

.warranty-check__label {
  display: block;
  font-size: 0.9rem;
  color: var(--tz-text-primary);
  margin-bottom: 0.5rem;
}

.warranty-check__input-group {
  display: flex;
  gap: 0.5rem;
}

.warranty-check__input {
  flex: 1;
  padding: 0.75rem 1rem;
  background: var(--tz-form-control-surface);
  border: 1px solid var(--tz-form-control-border);
  border-radius: 9999px;
  color: var(--tz-text-primary);
  font-size: 1rem;
  box-shadow:
    0 2px 6px rgba(20, 32, 43, 0.08);
}

.warranty-check__input:focus {
  outline: none;
  box-shadow:
    0 0 0 1px rgba(5, 150, 105, 0.75),
    0 0 14px rgba(5, 150, 105, 0.35);
}

.warranty-check__input::placeholder {
  color: var(--tz-text-muted);
}

.warranty-check__submit-btn {
  padding: 0.75rem 1.5rem;
  background: var(--tz-action-primary);
  border: 1px solid var(--tz-action-primary);
  border-radius: 9999px;
  color: var(--tz-action-primary-foreground);
  font-weight: 600;
  cursor: pointer;
  min-width: 100px;
  box-shadow: 0 4px 12px rgb(15 23 42 / 0.16);
  transition: opacity 0.2s, transform 0.18s ease, box-shadow 0.18s ease;
}

.warranty-check__submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.warranty-check__submit-btn:not(:disabled):hover {
  background: var(--tz-action-primary-hover);
  border-color: var(--tz-action-primary-hover);
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgb(15 23 42 / 0.22);
}

.warranty-check__spinner {
  display: inline-block;
  width: 1rem;
  height: 1rem;
  border: 2px solid #000;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.warranty-check__help {
  font-size: 0.8rem;
  color: var(--tz-text-muted);
  margin-top: 0.5rem;
}

/* 错误状态 */
.warranty-check__error {
  text-align: center;
  padding: 2rem;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 1rem;
}

.warranty-check__error-icon {
  font-size: 2.5rem;
  margin-bottom: 0.5rem;
}

.warranty-check__error h3 {
  color: #fca5a5;
  margin: 0 0 0.5rem;
}

.warranty-check__error p {
  color: var(--tz-text-secondary);
  margin: 0 0 1rem;
}

.warranty-check__tips {
  text-align: left;
  padding-left: 1.5rem;
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
  margin-bottom: 1rem;
}

.warranty-check__tips li {
  margin: 0.25rem 0;
}

.warranty-check__error-contact {
  font-size: 0.85rem;
  margin-bottom: 1rem;
}

.warranty-check__contact-btn {
  display: inline-block;
  padding: 0.5rem 1.5rem;
  background: var(--tz-surface-subtle);
  border: none;
  border-radius: 9999px;
  color: var(--tz-text-primary);
  text-decoration: none;
  font-size: 0.85rem;
  box-shadow: 0 6px 16px rgba(20, 32, 43, 0.08);
  transition: all 0.18s ease;
}

.warranty-check__contact-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 18px rgba(20, 32, 43, 0.12);
}

/* 查询结果 */
.warranty-check__result {
  background: var(--tz-card-surface);
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  overflow: hidden;
}

.warranty-check__status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem;
  font-size: 1.1rem;
  font-weight: 600;
}

.warranty-check__status--valid {
  background: rgba(5, 150, 105, 0.15);
  color: var(--tz-site-accent);
}

.warranty-check__status--expired {
  background: rgba(239, 68, 68, 0.15);
  color: #fca5a5;
}

.warranty-check__status-icon {
  font-size: 1.25rem;
}

.warranty-check__info {
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.warranty-check__info-row {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.warranty-check__info-row:last-child {
  border-bottom: none;
}

.warranty-check__info-label {
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
}

.warranty-check__info-value {
  color: var(--tz-text-primary);
  font-weight: 500;
}

.warranty-check__remaining {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem;
  font-weight: 500;
}

.warranty-check__remaining--valid {
  background: rgba(5, 150, 105, 0.1);
  color: var(--tz-site-accent);
}

.warranty-check__remaining--expired {
  background: rgba(239, 68, 68, 0.1);
  color: #fca5a5;
}

.warranty-check__remaining-icon {
  font-size: 1.1rem;
}

/* 服务记录 */
.warranty-check__records {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.warranty-check__records h4 {
  font-size: 0.9rem;
  color: var(--tz-text-primary);
  margin: 0 0 0.75rem;
}

.warranty-check__records-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.warranty-check__record-item {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--tz-border-subtle);
  font-size: 0.85rem;
}

.warranty-check__record-item:last-child {
  border-bottom: none;
}

.warranty-check__record-type {
  padding: 0.15rem 0.5rem;
  background: var(--tz-site-accent-soft-surface);
  color: var(--tz-site-accent);
  border-radius: 9999px;
  font-size: 0.75rem;
}

.warranty-check__record-date {
  color: var(--tz-text-secondary);
}

.warranty-check__record-desc {
  width: 100%;
  color: var(--tz-text-secondary);
}

.warranty-check__no-records {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.warranty-check__no-records p {
  color: var(--tz-text-muted);
  font-size: 0.85rem;
  margin: 0;
}

/* 操作按钮 */
.warranty-check__actions {
  display: flex;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.warranty-check__action-btn {
  flex: 1;
  padding: 0.6rem 1rem;
  border-radius: 9999px;
  font-size: 0.85rem;
  font-weight: 500;
  text-align: center;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.18s ease;
}

.warranty-check__action-btn:not(.warranty-check__action-btn--secondary) {
  background: var(--tz-action-primary);
  border: 1px solid var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  box-shadow: 0 4px 12px rgb(15 23 42 / 0.16);
}

.warranty-check__action-btn:not(.warranty-check__action-btn--secondary):hover {
  background: var(--tz-action-primary-hover);
  border-color: var(--tz-action-primary-hover);
}

.warranty-check__action-btn--secondary {
  background: var(--tz-surface-subtle);
  border: none;
  color: var(--tz-text-primary);
  box-shadow: 0 6px 16px rgba(20, 32, 43, 0.08);
}

.warranty-check__action-btn--secondary:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 18px rgba(20, 32, 43, 0.12);
}

/* 响应式 */
@media (max-width: 480px) {
  .warranty-check__input-group {
    flex-direction: column;
  }

  .warranty-check__submit-btn {
    width: 100%;
  }

  .warranty-check__actions {
    flex-direction: column;
  }
}
</style>
