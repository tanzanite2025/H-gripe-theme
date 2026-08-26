<template>
  <section id="home-shop-with-confidence" class="home-shop-with-confidence bg-transparent py-8 tz-text-primary sm:py-10 lg:py-12">
    <div class="page-content-shell px-0 md:px-4">
      <header class="home-shop-with-confidence__header">
        <span class="home-shop-with-confidence__eyebrow">
          <Icon name="lucide:shield-check" class="h-4 w-4" aria-hidden="true" />
          {{ t('home.shopWithConfidence.eyebrow') }}
        </span>
        <h2>{{ t('home.trust.title') }}</h2>
        <p>{{ t('home.shopWithConfidence.subtitle') }}</p>
      </header>

      <div class="home-shop-with-confidence__grid">
        <article
          v-for="card in cards"
          :key="card.id"
          class="home-shop-with-confidence__card premium-card"
        >
          <div class="home-shop-with-confidence__card-heading">
            <span class="home-shop-with-confidence__icon" aria-hidden="true">
              <Icon :name="card.icon" class="h-6 w-6" />
            </span>
            <h3>{{ t(card.titleKey) }}</h3>
          </div>

          <p class="home-shop-with-confidence__description">
            {{ t(card.descriptionKey) }}
          </p>

          <template v-if="card.isPayment">
            <div class="home-shop-with-confidence__payment-logos" :aria-label="t('home.shopWithConfidence.paymentMethodsLabel')">
              <img src="/icons/payment/visa.svg" alt="Visa" width="200" height="120" />
              <img src="/icons/payment/mastercard.svg" alt="Mastercard" width="200" height="120" />
              <img src="/icons/payment/amex.svg" alt="American Express" width="200" height="120" />
              <img src="/icons/payment/paypal.svg" alt="PayPal" width="200" height="120" />
              <img src="/icons/payment/wechatpay.svg" alt="WeChat Pay" width="200" height="120" />
              <img src="/icons/payment/alipay.svg" alt="Alipay" width="200" height="120" />
            </div>

            <ul class="home-shop-with-confidence__assurances">
              <li v-for="assuranceKey in paymentAssuranceKeys" :key="assuranceKey">
                <Icon name="lucide:check" class="h-3.5 w-3.5" aria-hidden="true" />
                <span>{{ t(assuranceKey) }}</span>
              </li>
            </ul>
          </template>

          <NuxtLink
            :to="localePath(card.to)"
            class="home-shop-with-confidence__action premium-button"
          >
            <Icon name="lucide:arrow-right" class="h-4 w-4" aria-hidden="true" />
            {{ t(card.actionKey) }}
          </NuxtLink>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n, useLocalePath } from '#imports'

type ShopWithConfidenceCard = {
  id: string
  icon: string
  titleKey: string
  descriptionKey: string
  actionKey: string
  to: string
  isPayment?: boolean
}

const { t } = useI18n()
const localePath = useLocalePath()

const cards: ShopWithConfidenceCard[] = [
  {
    id: 'secure-payment',
    icon: 'lucide:credit-card',
    titleKey: 'home.shopWithConfidence.cards.securePayment.title',
    descriptionKey: 'home.shopWithConfidence.cards.securePayment.description',
    actionKey: 'home.shopWithConfidence.cards.securePayment.action',
    to: '/support/payment',
    isPayment: true,
  },
  {
    id: 'payment-verification',
    icon: 'lucide:badge-check',
    titleKey: 'home.shopWithConfidence.cards.paymentVerification.title',
    descriptionKey: 'home.shopWithConfidence.cards.paymentVerification.description',
    actionKey: 'home.shopWithConfidence.cards.paymentVerification.action',
    to: '/support/payment',
  },
  {
    id: 'delivery-support',
    icon: 'lucide:package-check',
    titleKey: 'home.shopWithConfidence.cards.deliverySupport.title',
    descriptionKey: 'home.shopWithConfidence.cards.deliverySupport.description',
    actionKey: 'home.shopWithConfidence.cards.deliverySupport.action',
    to: '/support/shipping',
  },
  {
    id: 'after-sales-support',
    icon: 'lucide:shield-plus',
    titleKey: 'home.shopWithConfidence.cards.afterSalesSupport.title',
    descriptionKey: 'home.shopWithConfidence.cards.afterSalesSupport.description',
    actionKey: 'home.shopWithConfidence.cards.afterSalesSupport.action',
    to: '/support/warranty',
  },
]

const paymentAssuranceKeys = [
  'home.shopWithConfidence.paymentAssurances.0',
  'home.shopWithConfidence.paymentAssurances.1',
]
</script>

<style scoped>
#home-shop-with-confidence {
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}

.home-shop-with-confidence__header {
  max-width: 44rem;
  margin: 0 auto 1.25rem;
  text-align: center;
}

.home-shop-with-confidence__eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--tz-site-accent);
  font-size: 0.6875rem;
  font-weight: 800;
  letter-spacing: 0.14em;
  line-height: 1.25;
  text-transform: uppercase;
}

.home-shop-with-confidence__header h2 {
  margin: 0.5rem 0 0;
  color: var(--tz-text-primary);
  font-size: clamp(1.5rem, 2vw, 2rem);
  font-weight: 800;
  line-height: 1.2;
}

.home-shop-with-confidence__header p {
  margin: 0.625rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.9375rem;
  line-height: 1.55;
}

.home-shop-with-confidence__grid {
  display: grid;
  gap: 0.875rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.home-shop-with-confidence__card {
  display: flex;
  min-width: 0;
  min-height: 15.5rem;
  flex-direction: column;
  gap: 0.875rem;
  padding: 1.125rem;
  border-radius: 8px;
}

.home-shop-with-confidence__card-heading {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.75rem;
}

.home-shop-with-confidence__icon {
  display: grid;
  width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  place-items: center;
  color: var(--tz-site-accent);
}

.home-shop-with-confidence__card-heading h3 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.25;
}

.home-shop-with-confidence__description {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.875rem;
  line-height: 1.55;
}

.home-shop-with-confidence__payment-logos {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.45rem;
}

.home-shop-with-confidence__payment-logos img {
  width: auto;
  height: 1.375rem;
  object-fit: contain;
}

.home-shop-with-confidence__assurances {
  display: grid;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
  line-height: 1.4;
  list-style: none;
}

.home-shop-with-confidence__assurances li {
  display: flex;
  align-items: flex-start;
  gap: 0.45rem;
}

.home-shop-with-confidence__assurances svg {
  flex: 0 0 auto;
  margin-top: 0.1rem;
  color: var(--tz-site-accent);
}

.home-shop-with-confidence__action {
  width: 100%;
  justify-content: center;
  gap: 0.5rem;
  margin-top: auto;
}

@media (max-width: 640px) {
  .home-shop-with-confidence__header {
    margin-bottom: 1rem;
  }

  .home-shop-with-confidence__grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .home-shop-with-confidence__card {
    min-height: 0;
    padding: 1rem;
  }
}
</style>
