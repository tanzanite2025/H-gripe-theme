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
        <label for="product-code" class="warranty-check__label">
          {{ $t('warranty.input_label') }}
        </label>
        <div class="warranty-check__input-group">
          <input
            id="product-code"
            v-model="productCode"
            type="text"
            class="warranty-check__input"
            :placeholder="$t('warranty.input_placeholder')"
            @keypress.enter="checkWarranty"
          />
          <button
            type="button"
            class="warranty-check__submit-btn"
            :disabled="loading || !productCode.trim()"
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
        <p>{{ $t('warranty.errors.not_found_message', { code: searchedCode }) }}</p>
        <ul class="warranty-check__tips">
          <li>{{ $t('warranty.errors.check_tips.0') }}</li>
          <li>{{ $t('warranty.errors.check_tips.1') }}</li>
        </ul>
        <p class="warranty-check__error-contact">{{ $t('warranty.errors.error_contact') }}</p>
        <NuxtLink to="/support" class="warranty-check__contact-btn">
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
            <span class="warranty-check__info-label">{{ $t('warranty.fields.product_code') }}</span>
            <span class="warranty-check__info-value">{{ result.product_code }}</span>
          </div>
          <div class="warranty-check__info-row">
            <span class="warranty-check__info-label">{{ $t('warranty.fields.product_type') }}</span>
            <span class="warranty-check__info-value">
              {{ String(resultLocale).startsWith('zh') ? result.product_type.name_zh : result.product_type.name }}
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
                {{ String(resultLocale).startsWith('zh') ? record.type_name_zh : record.type_name }}
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
            to="/support"
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

const props = defineProps<{
  isLoggedIn: boolean
}>()

const emit = defineEmits<{
  'login-request': []
}>()

const { locale } = useI18n()
const resultLocale = computed(() => locale.value)

const {
  productCode,
  searchedCode,
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
