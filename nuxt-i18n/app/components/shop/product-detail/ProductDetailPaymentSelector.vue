<template>
  <div v-if="options.length" class="product-payment-selector" aria-label="Payment method">
    <div class="product-payment-selector__header">
      <span>{{ t('checkout.steps.payment', 'Choose payment method') }}</span>
      <small v-if="loading">{{ t('common.loading', 'Loading...') }}</small>
    </div>
    <div class="product-payment-options">
      <button
        v-for="option in options"
        :key="paymentKey(option)"
        type="button"
        class="product-payment-option"
        :class="{
          'product-payment-option--selected': selectedMethod === paymentMethod(option),
          'product-payment-option--unavailable': !isAvailable(option),
        }"
        :disabled="!isAvailable(option)"
        :aria-disabled="!isAvailable(option)"
        :aria-pressed="selectedMethod === paymentMethod(option)"
        @click="selectPaymentMethod(paymentMethod(option))"
      >
        <span class="product-payment-option__logos" aria-hidden="true">
          <img
            v-for="logo in paymentLogos(option)"
            :key="logo.src"
            :src="logo.src"
            :alt="logo.alt"
            :width="logo.width"
            :height="logo.height"
            :class="logo.className"
            loading="lazy"
          />
        </span>
        <span class="product-payment-option__body">
          <span class="product-payment-option__title-row">
            <span class="product-payment-option__title">{{ paymentTitle(option) }}</span>
            <small v-if="!isAvailable(option)" class="product-payment-option__status">
              {{ unavailableLabel(option) }}
            </small>
          </span>
          <span class="product-payment-option__description">{{ paymentDescription(option) }}</span>
        </span>
        <Icon
          v-if="selectedMethod === paymentMethod(option)"
          name="lucide:check"
          class="product-payment-option__check"
          aria-hidden="true"
        />
      </button>
    </div>
    <p v-if="error" class="product-payment-status">
      {{ error }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import type { CheckoutPaymentOption } from '~/types/payment'
import {
  isPaymentOptionAvailable,
  paymentMethodFromOption,
  paymentPresentation,
  type PaymentLogoAsset,
  type StorefrontPaymentMethod,
} from '~/utils/paymentPresentation'

defineProps<{
  options: CheckoutPaymentOption[]
  selectedMethod: StorefrontPaymentMethod
  loading: boolean
  error?: string | null
}>()

const emit = defineEmits<{
  (event: 'update:selectedMethod', method: StorefrontPaymentMethod): void
}>()

const { t } = useI18n()

const paymentMethod = (option: CheckoutPaymentOption) => paymentMethodFromOption(option)
const isAvailable = isPaymentOptionAvailable

const paymentKey = (option: CheckoutPaymentOption) =>
  `${paymentMethod(option)}-${option.id || option.code || option.provider || 'payment'}`

const paymentTitle = (option: CheckoutPaymentOption) => {
  const method = paymentMethod(option)
  if (!method) return option.title || option.code || option.id
  const presentation = paymentPresentation(method)
  return t(presentation.titleKey, presentation.title)
}

const paymentDescription = (option: CheckoutPaymentOption) => {
  if (option.description) return option.description
  const method = paymentMethod(option)
  if (!method) return option.subtitle || ''
  const presentation = paymentPresentation(method)
  return t(presentation.descriptionKey, presentation.description)
}

const paymentLogos = (option: CheckoutPaymentOption): PaymentLogoAsset[] => {
  const method = paymentMethod(option)
  return method
    ? paymentPresentation(method).logos
    : [{ src: '/icons/payment/default.svg', alt: paymentTitle(option), width: 750, height: 471 }]
}

const unavailableLabel = (option: CheckoutPaymentOption) => {
  const reason = String(option.unavailableReason || option.unavailable_reason || '').trim()
  if (reason === 'gateway_not_configured' || reason === 'gateway_config_invalid' || reason === 'disabled') {
    return t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
  }
  return reason
    ? reason.replace(/_/g, ' ')
    : t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
}

const selectPaymentMethod = (method: StorefrontPaymentMethod | '') => {
  if (method) emit('update:selectedMethod', method)
}
</script>

<style scoped>
.product-payment-selector {
  display: grid;
  gap: 0.65rem;
}

.product-payment-selector__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.85rem;
  font-weight: 700;
  text-transform: uppercase;
}

.product-payment-selector__header small {
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: none;
}

.product-payment-options {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.55rem;
}

.product-payment-option {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  min-height: 4.35rem;
  align-items: center;
  gap: 0.75rem;
  box-sizing: border-box;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.85rem;
  background: var(--tz-surface-card);
  color: var(--tz-text-primary);
  cursor: pointer;
  padding: 0.75rem;
  text-align: left;
  transition: background-color 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
}

.product-payment-option:hover:not(:disabled) {
  border-color: rgba(5, 150, 105, 0.62);
  background: rgba(5, 150, 105, 0.1);
  transform: translateY(-1px);
}

.product-payment-option--selected {
  border-color: rgba(5, 150, 105, 0.82);
  background: rgba(5, 150, 105, 0.14);
}

.product-payment-option--unavailable {
  cursor: not-allowed;
  opacity: 0.52;
}

.product-payment-option__logos {
  display: inline-flex;
  min-width: 3.1rem;
  max-width: 4.9rem;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.18rem;
}

.product-payment-option__logos img {
  display: block;
  width: auto;
  max-width: 2.35rem;
  height: 1rem;
  object-fit: contain;
}

.product-payment-option__logos img.payment-logo--alipay {
  max-width: 2.75rem;
}

.product-payment-option__body {
  display: grid;
  min-width: 0;
  gap: 0.24rem;
}

.product-payment-option__title-row {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
}

.product-payment-option__title {
  min-width: 0;
  overflow: hidden;
  font-size: 0.86rem;
  font-weight: 800;
  line-height: 1.1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-payment-option__status {
  display: inline-flex;
  height: 1rem;
  align-items: center;
  border-radius: 999px;
  background: var(--tz-status-warning-bg);
  color: var(--tz-status-warning-text);
  font-size: 0.62rem;
  font-weight: 700;
  line-height: 1;
  padding: 0 0.34rem;
  white-space: nowrap;
}

.product-payment-option__description {
  display: -webkit-box;
  overflow: hidden;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  font-weight: 600;
  line-height: 1.35;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.product-payment-option__check {
  width: 0.95rem;
  height: 0.95rem;
  color: #059669;
}

.product-payment-status {
  flex-basis: 100%;
  margin: 0;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
}

.product-payment-option:focus-visible {
  outline: 2px solid #059669;
  outline-offset: 3px;
}

@media (max-width: 640px) {
  .product-payment-options {
    grid-template-columns: 1fr;
  }
}
</style>
