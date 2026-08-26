<template>
  <section class="giftcard-exchange" aria-labelledby="giftcard-exchange-title">
    <div class="giftcard-exchange__shell">
      <div class="giftcard-exchange__header">
        <div class="giftcard-exchange__copy">
          <p class="giftcard-exchange__eyebrow">{{ t('giftcards.eyebrow', 'Gift cards') }}</p>
          <h3 id="giftcard-exchange-title" class="giftcard-exchange__title">
            {{ t('giftcards.title', 'Redeem Points for Gift Cards') }}
          </h3>
          <p class="giftcard-exchange__intro">
            {{
              t(
                'giftcards.exchangeIntro',
                'Convert loyalty points into published gift card options, then track redeemed cards from your member center.'
              )
            }}
          </p>
        </div>

        <div class="giftcard-exchange__stats" aria-label="Gift card exchange summary">
          <div class="giftcard-exchange__stat">
            <span>{{ t('giftcards.stats.points', 'Points balance') }}</span>
            <strong>{{ pointsDisplay }}</strong>
          </div>
          <div class="giftcard-exchange__stat">
            <span>{{ t('giftcards.stats.redemption', 'Redemption rule') }}</span>
            <strong>{{ redemptionRuleDescription || t('member.points.notConfigured', 'Not configured') }}</strong>
          </div>
          <div class="giftcard-exchange__stat">
            <span>{{ t('giftcards.stats.owned', 'Owned cards') }}</span>
            <strong>{{ ownedCardsCount }}</strong>
          </div>
        </div>
      </div>

      <div v-if="!isLogged" class="giftcard-exchange__notice">
        {{ t('giftcards.signInNotice', 'Sign in to redeem points and see your redeemed gift cards.') }}
      </div>

      <div class="giftcard-exchange__body">
        <div v-if="loading" class="giftcard-exchange__state">
          <span class="giftcard-exchange__state-title">{{ t('common.loading', 'Loading...') }}</span>
          <span class="giftcard-exchange__state-copy">
            {{ t('giftcards.loadingCopy', 'Fetching active gift card redemption options.') }}
          </span>
        </div>

        <div v-else-if="error" class="giftcard-exchange__state giftcard-exchange__state--error">
          <span class="giftcard-exchange__state-title">
            {{ t('giftcards.loadErrorTitle', 'Could not load gift cards') }}
          </span>
          <span class="giftcard-exchange__state-copy">{{ error }}</span>
        </div>

        <template v-else>
          <div v-if="availableGiftcards.length > 0" class="giftcard-grid">
            <article v-for="card in availableGiftcards" :key="card.id" class="giftcard-item">
              <div class="giftcard-header">
                <Icon name="lucide:credit-card" class="giftcard-icon" aria-hidden="true" />
                <div class="giftcard-info">
                  <div class="giftcard-code">{{ card.label }}</div>
                  <div class="giftcard-label">{{ t('giftcards.balance', 'Balance') }}</div>
                </div>
                <div class="giftcard-value">{{ formatGiftcardValue(card) }}</div>
              </div>

              <div class="giftcard-footer">
                <span class="giftcard-points">
                  {{ t('giftcards.pointsRequired', 'Points required') }}: {{ formatNumber(card.points_required || 0) }}
                  <span class="giftcard-stock">
                    · {{ t('giftcards.remaining', 'Remaining') }}: {{ formatNumber(card.remaining_quantity || 0) }}
                  </span>
                </span>
                <button
                  class="btn-redeem"
                  type="button"
                  :disabled="isRedeemDisabled(card)"
                  @click="emit('redeem', card)"
                >
                  {{ redeemButtonLabel(card) }}
                </button>
              </div>
            </article>
          </div>

          <div v-else class="giftcard-exchange__empty">
            <h4>{{ t('giftcards.noCardsTitle', 'No active gift cards right now') }}</h4>
            <p>
              {{
                t(
                  'giftcards.noCardsBody',
                  'When new reward options are released, they will appear here for points redemption.'
                )
              }}
            </p>
            <div class="giftcard-exchange__empty-actions">
<NuxtLink :to="localePath('/resources/membershipandpoints/levers')" class="giftcard-exchange__link giftcard-exchange__link--primary">
                {{ t('giftcards.viewPointRules', 'View point rules') }}
              </NuxtLink>
<NuxtLink :to="localePath('/resources/membershipandpoints/myinfo')" class="giftcard-exchange__link">
                {{ t('giftcards.backToMemberInfo', 'Member info') }}
              </NuxtLink>
            </div>
          </div>

          <div v-if="isLogged && userGiftCards.length > 0" class="owned-giftcard-list">
            <h4>{{ t('giftcards.myCards', 'My Gift Cards') }}</h4>
            <div v-for="card in userGiftCards" :key="card.id" class="owned-giftcard-item">
              <div>
                <strong>{{ card.code }}</strong>
                <span>{{ formatOwnedGiftcardBalance(card) }}</span>
              </div>
              <small>{{ card.status }}</small>
            </div>
          </div>
        </template>

        <div v-if="redeemMessage" class="redeem-message" :class="{ success: redeemSuccess, error: !redeemSuccess }">
          {{ redeemMessage }}
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useLocalePath } from '#imports'

interface GiftCardOption {
  id: number
  label: string
  giftcard_value: number
  giftcard_value_cents: number
  currency: string
  points_required: number
  remaining_quantity: number
  status: string
  cover_image?: string
}

interface UserGiftCard {
  id: number
  code: string
  balance?: number | null
  balance_cents?: number
  currency: string
  status: string
}

const props = withDefaults(defineProps<{
  isLogged: boolean
  points: number
  redemptionRuleDescription?: string
  availableGiftcards: GiftCardOption[]
  userGiftCards: UserGiftCard[]
  loading?: boolean
  error?: string
  redeemingCardId?: number | null
  redeemMessage?: string
  redeemSuccess?: boolean
}>(), {
  redemptionRuleDescription: '',
  loading: false,
  error: '',
  redeemingCardId: null,
  redeemMessage: '',
  redeemSuccess: false,
})

const emit = defineEmits<{
  (event: 'redeem', card: GiftCardOption): void
}>()

const { t } = useI18n()
const localePath = useLocalePath()

const formatNumber = (value: number) => {
  const numericValue = Number(value)
  if (!Number.isFinite(numericValue)) return '0'
  return new Intl.NumberFormat().format(Math.max(0, Math.round(numericValue)))
}

const formatAmount = (value: number) => {
  const numericValue = Number(value)
  if (!Number.isFinite(numericValue)) return '0'
  return Number.isInteger(numericValue) ? String(numericValue) : numericValue.toFixed(2).replace(/\.?0+$/, '')
}

const pointsDisplay = computed(() => formatNumber(props.points || 0))
const ownedCardsCount = computed(() => formatNumber(props.userGiftCards.length))

const formatGiftcardValue = (card: GiftCardOption) => {
  return `${card.currency} ${formatAmount(Number(card.giftcard_value ?? 0))}`
}

const formatOwnedGiftcardBalance = (card: UserGiftCard) => {
  return `${card.currency} ${Number(card.balance ?? 0).toFixed(2)}`
}

const isRedeemDisabled = (card: GiftCardOption) => {
  return (props.isLogged && Number(props.points || 0) < Number(card.points_required || 0))
    || props.redeemingCardId === card.id
}

const redeemButtonLabel = (card: GiftCardOption) => {
  if (props.redeemingCardId === card.id) return t('giftcards.redeeming', 'Redeeming...')
  if (!props.isLogged) return t('giftcards.signInToRedeem', 'Sign in to redeem')
  return t('giftcards.redeem', 'Redeem')
}
</script>

<style scoped>
.giftcard-exchange {
  margin-top: 2rem;
  margin-bottom: 2rem;
}

.giftcard-exchange__shell {
  display: grid;
  gap: 1.1rem;
  border-radius: 1rem;
  background: var(--membership-card-surface, var(--tz-card-surface));
  padding: 1rem;
  box-shadow: var(--membership-card-shadow, 0 16px 34px rgba(0, 0, 0, 0.34));
}

.giftcard-exchange__header {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(18rem, 0.9fr);
  gap: 1rem;
  align-items: start;
}

.giftcard-exchange__copy {
  min-width: 0;
}

.giftcard-exchange__eyebrow {
  margin: 0 0 0.35rem;
  color: var(--membership-accent, var(--tz-site-accent, #059669));
  font-size: var(--tz-type-micro-label);
  font-weight: 850;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.giftcard-exchange__title {
  margin: 0;
  color: #ffffff;
  font-size: 1.15rem;
  font-weight: 850;
  line-height: 1.2;
}

.giftcard-exchange__intro {
  max-width: 46rem;
  margin: 0.55rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.9rem;
  line-height: 1.55;
}

.giftcard-exchange__stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.65rem;
}

.giftcard-exchange__stat {
  min-width: 0;
  border-radius: 0.85rem;
  background: var(--membership-card-subtle, var(--tz-surface-subtle));
  padding: 0.75rem;
}

.giftcard-exchange__stat span,
.giftcard-exchange__stat strong {
  display: block;
  min-width: 0;
}

.giftcard-exchange__stat span {
  color: var(--tz-text-muted);
  font-size: var(--tz-type-micro-label);
  font-weight: 750;
}

.giftcard-exchange__stat strong {
  margin-top: 0.35rem;
  color: #ffffff;
  font-size: 0.9rem;
  font-weight: 850;
  line-height: 1.3;
  overflow-wrap: anywhere;
}

.giftcard-exchange__notice {
  border-radius: 0.85rem;
  background: rgba(5, 150, 105, 0.08);
  padding: 0.75rem 0.85rem;
  color: var(--tz-text-secondary);
  font-size: 0.84rem;
  line-height: 1.45;
}

.giftcard-exchange__body {
  display: grid;
  gap: 1rem;
}

.giftcard-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.giftcard-item {
  min-width: 0;
  border-radius: 0.9rem;
  background: var(--membership-card-subtle, var(--tz-surface-subtle));
  padding: 0.85rem;
}

.giftcard-header {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.giftcard-icon {
  width: 1.5rem;
  height: 1.5rem;
  flex: 0 0 auto;
  color: var(--tz-text-secondary);
  stroke-width: 1.8;
}

.giftcard-info {
  min-width: 0;
  flex: 1;
}

.giftcard-code {
  overflow: hidden;
  color: var(--tz-text-primary);
  font-size: 0.9rem;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.giftcard-label {
  margin-top: 0.15rem;
  color: var(--tz-text-muted);
  font-size: 0.75rem;
}

.giftcard-value {
  flex: 0 0 auto;
  color: var(--membership-accent, var(--tz-site-accent, #059669));
  font-size: 1.05rem;
  font-weight: 850;
}

.giftcard-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.giftcard-points {
  min-width: 0;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.45;
}

.giftcard-stock {
  white-space: nowrap;
}

.btn-redeem {
  display: inline-flex;
  min-height: 2.25rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 9999px;
  background: var(--membership-accent, var(--tz-site-accent, #059669));
  color: #ffffff;
  cursor: pointer;
  font-size: 0.78rem;
  font-weight: 800;
  padding: 0.45rem 0.85rem;
  transition:
    filter 0.18s ease,
    opacity 0.18s ease;
  white-space: nowrap;
}

.btn-redeem:hover {
  filter: brightness(1.08);
}

.btn-redeem:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.giftcard-exchange__state,
.giftcard-exchange__empty {
  display: grid;
  gap: 0.55rem;
  min-height: 8rem;
  place-items: center;
  border-radius: 0.9rem;
  background: var(--membership-card-subtle, var(--tz-surface-subtle));
  padding: 1.25rem;
  text-align: center;
}

.giftcard-exchange__state-title,
.giftcard-exchange__empty h4 {
  margin: 0;
  color: #ffffff;
  font-size: 0.95rem;
  font-weight: 850;
}

.giftcard-exchange__state-copy,
.giftcard-exchange__empty p {
  max-width: 34rem;
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.84rem;
  line-height: 1.55;
}

.giftcard-exchange__state--error .giftcard-exchange__state-title,
.giftcard-exchange__state--error .giftcard-exchange__state-copy {
  color: #fca5a5;
}

.giftcard-exchange__empty-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.65rem;
  margin-top: 0.25rem;
}

.giftcard-exchange__link {
  display: inline-flex;
  min-height: 2.35rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
  font-size: 0.8rem;
  font-weight: 780;
  padding: 0.5rem 0.9rem;
  text-decoration: none;
}

.giftcard-exchange__link--primary {
  background: #ffffff;
  color: var(--tz-text-primary);
}

.owned-giftcard-list {
  display: grid;
  gap: 0.4rem;
  border-top: 1px solid var(--tz-border-subtle);
  padding-top: 1rem;
}

.owned-giftcard-list h4 {
  margin: 0 0 0.2rem;
  color: var(--tz-text-secondary);
  font-size: 0.86rem;
  font-weight: 800;
}

.owned-giftcard-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-radius: 0.75rem;
  background: var(--membership-card-subtle, var(--tz-surface-subtle));
  color: var(--tz-text-secondary);
  padding: 0.65rem 0.75rem;
}

.owned-giftcard-item > div {
  display: grid;
  gap: 0.18rem;
  min-width: 0;
}

.owned-giftcard-item strong {
  overflow: hidden;
  color: var(--tz-text-primary);
  font-size: 0.78rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.owned-giftcard-item span,
.owned-giftcard-item small {
  color: var(--tz-text-muted);
  font-size: 0.72rem;
}

.redeem-message {
  border-radius: 0.75rem;
  padding: 0.65rem 0.75rem;
  text-align: center;
  font-size: 0.84rem;
}

.redeem-message.success {
  background: rgba(5, 150, 105, 0.14);
  color: #a7f3d0;
}

.redeem-message.error {
  background: rgba(239, 68, 68, 0.18);
  color: #fca5a5;
}

@media (max-width: 900px) {
  .giftcard-exchange__header,
  .giftcard-grid {
    grid-template-columns: 1fr;
  }

  .giftcard-exchange__stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .giftcard-exchange {
    margin-top: 1.25rem;
  }

  .giftcard-exchange__shell {
    padding: 0.85rem;
  }

  .giftcard-exchange__stats {
    grid-template-columns: 1fr;
  }

  .giftcard-footer {
    align-items: stretch;
    flex-direction: column;
  }

  .btn-redeem {
    width: 100%;
  }
}
</style>
