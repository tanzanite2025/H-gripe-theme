<template>
  <div class="support-page">
    <h1 class="sr-only">Payment</h1>
    <div class="support-intro-banner">
      <p class="support-intro-banner__text">
        All pages on this site use HTTPS with an SSL certificate issued by a trusted certificate
        authority. All payment information is transmitted over encrypted channels to prevent it from
        being intercepted or tampered with. Actual charges are processed by reputable payment providers.
        We only receive the payment result and do not store your card number, CVV, or other sensitive data.
      </p>
      <div class="support-card__actions support-intro-banner__actions">
        <NuxtLink
          to="/company/ourstory"
          target="_blank"
          rel="noopener"
          class="premium-button"
        >
          Our Story
        </NuxtLink>
        <NuxtLink
          to="/support/shipping"
          target="_blank"
          rel="noopener"
          class="premium-button"
        >
          Shipping instructions
        </NuxtLink>
        <NuxtLink
          to="/support/test-report"
          target="_blank"
          rel="noopener"
          class="premium-button"
        >
          Test report
        </NuxtLink>
        <button type="button" class="premium-button">
          <span class="support-card__button-icon">📧</span>
          <span>Send email</span>
        </button>
      </div>
    </div>

    <section class="support-section support-section--compact">
      <h3 class="support-section__title">Product service fee</h3>
      <article class="support-card">
        <p class="support-card__body support-card__highlight">
          Wheel building: $30 USD per wheelset.
        </p>
        <p class="support-card__body">
          If you need to customize the wheel configuration, you can choose the configuration you like in
          the store, or the complete set of wheel configurations we regularly sell (you can also choose DIY
          spokes, etc.), or you can make a quick purchase directly from the button at the bottom of the
          webpage.
        </p>
        <div class="support-card__actions">
          <NuxtLink to="/shop" target="_blank" rel="noopener" class="premium-button">
            Go Shopping
          </NuxtLink>
          <button type="button" class="premium-button support-card__button--primary" @click="openQuickFromPayment">
            Quickbuy
          </button>
        </div>
      </article>
    </section>

    <QuickBuyModal v-if="quickOpen" :config="null" @close="quickOpen = false" />

    <section class="support-section">
      <h3 class="support-section__title">Supported payment methods</h3>
      <div class="support-section__grid">
        <article class="support-card">
          <h4 class="support-card__title">
            <span class="support-card__icon">
              <img src="/icons/payment/default.svg" alt="Credit & debit cards" class="support-card__icon-image" />
            </span>
            <span>Credit &amp; debit cards</span>
          </h4>
          <p class="support-card__body">
            We accept major cards such as Visa and MasterCard.
          </p>
          <div class="mt-4 pt-4 border-t border-slate-800/50 flex items-start gap-3">
             <div class="flex-shrink-0 pt-1">
                <img src="/icons/payment/stripe.svg?v=4" alt="Stripe" class="h-5 w-auto opacity-80" />
             </div>
             <p class="text-xs tz-text-secondary leading-relaxed">
               Payments are securely processed by <strong>Stripe</strong>. Your payment data is encrypted and secure. We do not store your card details.
             </p>
          </div>
        </article>
        <article class="support-card">
          <h4 class="support-card__title">
            <span class="support-card__icon">
              <img src="/icons/payment/paypal.svg?v=4" alt="PayPal or wallet payments" class="support-card__icon-image" />
            </span>
            <span>PayPal or wallet payments</span>
          </h4>
          <p class="support-card__body">
            Where available, you can choose PayPal or a supported wallet at checkout for faster payment
            and additional buyer protection. You will be redirected back to our site once the payment is
            approved.
          </p>
        </article>
        <article class="support-card">
          <h4 class="support-card__title">
            <span class="support-card__icon">
              <img src="/icons/payment/bank-transfer.svg?v=4" alt="Bank transfer" class="support-card__icon-image" />
            </span>
            <span>Bank transfer (on request)</span>
          </h4>
          <p class="support-card__body">
            For larger or custom orders we can sometimes arrange payment via bank transfer. Please
            contact support before placing the order so that we can provide the correct account details
            and reserve your items.
          </p>
        </article>
        <article class="support-card">
          <h4 class="support-card__title">
            <span class="support-card__icon-group">
              <span class="support-card__icon">
                <img src="/icons/payment/wechatpay.svg?v=4" alt="WeChat Pay" class="support-card__icon-image" />
              </span>
              <span class="support-card__icon">
                <img src="/icons/payment/alipay.svg?v=6" alt="Alipay" class="support-card__icon-image" />
              </span>
            </span>
            <span>WeChat / Alipay</span>
          </h4>
          <p class="support-card__body">
            If you are a Chinese user, you can also use WeChat or Alipay to pay. You can pay on the
            payment page, or ask customer service for the payment code.
          </p>
        </article>
      </div>
    </section>

    <section class="support-section">
      <h3 class="support-section__title">Currencies &amp; charges</h3>
      <div class="support-section__grid support-section__grid--two">
        <article class="support-card">
          <h4 class="support-card__title">Display currency vs. billing currency</h4>
          <p class="support-card__body">
            Product prices on the site are usually shown in a reference currency (for example USD).
            Your bank or payment provider may convert this amount into your local currency using their
            own exchange rate and possible conversion fees.
          </p>
          <p class="support-card__body">
            Depending on your bank or payment provider, additional currency conversion or processing charges may apply. These charges are set by your provider and are not collected by us.
          </p>
        </article>
        <article class="support-card">
          <h4 class="support-card__title">Taxes &amp; import duties</h4>
          <p class="support-card__body">
            Depending on your shipping country, local VAT or import duties may apply. These charges
            are handled according to the shipping option shown at checkout. Please review the order
            summary carefully before confirming payment.
          </p>
        </article>
      </div>
    </section>

    <section class="support-section">
      <h3 class="support-section__title">When is my payment captured?</h3>
      <ul class="support-list">
        <li class="support-list__item">
          <strong>Card &amp; wallet payments</strong> — the authorization is created when you confirm the
          order. The final capture normally happens when the order is accepted and prepared for
          shipment.
        </li>
        <li class="support-list__item">
          <strong>Bank transfer</strong> — we will start processing your order after the funds have been
          received and matched to your order reference.
        </li>
        <li class="support-list__item">
          If an order cannot be fulfilled, the payment authorization will be voided or refunded
          according to your original payment method.
        </li>
      </ul>
    </section>

    <section class="support-section">
      <h3 class="support-section__title">Troubleshooting common payment issues</h3>
      <div class="support-section__grid support-section__grid--two">
        <article class="support-card">
          <h4 class="support-card__title">Payment was declined</h4>
          <p class="support-card__body">
            If your card or wallet is declined, please check that your billing details match your bank
            records, ensure that sufficient funds are available, and try again. In many cases your bank
            can provide more detail via their app or customer service.
          </p>
        </article>
        <article class="support-card">
          <h4 class="support-card__title">Charged but order not created</h4>
          <p class="support-card__body">
            On rare occasions a network issue can interrupt the redirect back to our site after
            payment. If you see a charge but no order in your account, please contact support with
            your payment reference so that we can investigate.
          </p>
        </article>
        <article class="support-card">
          <h4 class="support-card__title">Multiple charges</h4>
          <p class="support-card__body">
            If you attempted payment several times, your bank may show multiple pending
            authorizations. Normally, any unused authorizations are released automatically after a
            short period. If this does not happen, please contact your bank or our support team.
          </p>
        </article>
      </div>
    </section>

    <section class="support-section">
      <h3 class="support-section__title">Need help with a payment?</h3>
      <p class="support-section__note">
        If you are unsure whether a payment was successful, or if you notice any unusual activity on
        your account, please stop using that payment method and contact us together with a recent
        statement or payment reference. Our support team will never ask for your full card number or
        card PIN.
      </p>
    </section>

    <!-- FAQ Section -->
    <section class="support-section">
      <PageFaq 
        page-id="support-payment"
        theme="dark"
        :show-categories="true"
      />
    </section>

    <section class="support-section">
      <UserFeedbackThread
        threadKey="support-payment"
        title="Share your feedback about payment &amp; billing"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import QuickBuyModal from '@/components/QuickBuy.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import PageFaq from '~/components/PageFaq.vue'

definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Payment',
})

const quickOpen = ref(false)

const openQuickFromPayment = () => {
  quickOpen.value = true
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'payment-quick' } }))
  }
}
</script>

<style scoped>
.support-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.support-page__intro {
  margin: 0 0 1.5rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

.support-intro-banner {
  margin: 0 0 1.25rem;
  padding: 1.5rem;
  border-radius: 1rem;
  background: #11151e;
  border: none;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.6);
}

.support-intro-banner__text {
  margin: 0;
  font-size: 0.9rem;
  line-height: 1.6;
  color: var(--tz-text-secondary);
}

.support-section {
  margin-top: 2rem;
}

.support-section__title {
  margin: 0 0 1rem;
  font-size: var(--tz-type-section-title);
  line-height: 1.35;
  font-weight: 700;
  color: var(--tz-text-primary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  text-align: center;
}

.support-section__grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1.5rem;
}

.support-section__grid--two {
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
}

/* Premium Card Styling */
.support-card {
  border-radius: 1rem;
  background: #11151e;
  border: none;
  padding: 1.5rem;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.6);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.support-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.8);
}

.support-card__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-card-title);
  line-height: 1.35;
  font-weight: 700;
  color: #f1f5f9;
  display: flex;
  align-items: center;
}

.support-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: flex-start;
  width: auto;
  height: auto;
  background: transparent;
  border: none;
  padding: 0;
  margin-right: 0.75rem;
}

.support-card__icon-image {
  width: 2rem;
  height: 2rem;
  object-fit: contain;
}

.support-card__icon-image + .support-card__icon-image {
  margin-left: 0.25rem;
}

.support-card__icon-group {
  display: inline-flex;
  align-items: center;
}

.support-card__icon-group .support-card__icon {
  margin-right: 0.25rem;
}

.support-card__body {
  margin: 0 0 0.5rem 0;
  font-size: 0.95rem;
  line-height: 1.6;
  color: var(--tz-text-secondary);
}

.support-card__highlight {
  color: #2dd4bf; /* Teal for highlights instead of amber */
  font-weight: 600;
}

.support-card__actions {
  margin-top: 1.5rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}



.support-card__button-icon {
  margin-right: 0.4rem;
  font-size: 1rem;
}

.support-card__button--primary {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #fff;
  border: none;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}

.support-card__button--primary:hover {
  box-shadow: 0 6px 16px rgba(45, 212, 191, 0.5);
  transform: translateY(-1px);
}

.support-list {
  margin: 0;
  padding-left: 1.25rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

.support-list__item + .support-list__item {
  margin-top: 0.5rem;
}

.support-section__note {
  margin: 0;
  font-size: 0.95rem;
  line-height: 1.6;
  color: var(--tz-text-secondary);
  text-align: center;
  max-width: 800px;
  margin: 0 auto;
}

@media (min-width: 768px) {
  .support-section {
    margin-top: 3rem;
  }
}
</style>
