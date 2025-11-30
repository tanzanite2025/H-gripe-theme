<template>
  <div class="support-page">
    <h2 class="support-page__title">Payment</h2>
    <div class="support-intro-banner">
      <p class="support-intro-banner__text">
        All pages on this site use HTTPS with an SSL certificate issued by a trusted certificate
        authority. All payment information is transmitted over encrypted channels to prevent it from
        being intercepted or tampered with. Actual charges are processed by reputable payment providers
        such as PayPal, Stripe, Alipay / WeChat Pay / UnionPay. We only receive the payment result and do
        not store your card number, CVV, or other sensitive data.
      </p>
      <div class="support-card__actions support-intro-banner__actions">
        <NuxtLink to="/company/ourstory" class="support-card__button">
          Our Story
        </NuxtLink>
        <button type="button" class="support-card__button">
          Shipping instructions
        </button>
        <button type="button" class="support-card__button">
          Test report
        </button>
        <button type="button" class="support-card__button">
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
          <NuxtLink to="/shop" target="_blank" rel="noopener" class="support-card__button">
            Go Shopping
          </NuxtLink>
          <button type="button" class="support-card__button support-card__button--primary" @click="openQuickFromPayment">
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
              <img src="/icons/payment/credit-card.svg" alt="Credit &amp; debit cards" class="support-card__icon-image" />
            </span>
            <span>Credit &amp; debit cards</span>
          </h4>
          <p class="support-card__body">
            We accept major cards such as Visa and MasterCard. Your card will only be charged after the
            order has been confirmed. In some regions additional 3D Secure verification may be required.
          </p>
        </article>
        <article class="support-card">
          <h4 class="support-card__title">
            <span class="support-card__icon">
              <img src="/icons/payment/paypal.svg" alt="PayPal or wallet payments" class="support-card__icon-image" />
            </span>
            <span>PayPal or wallet payments</span>
          </h4>
          <p class="support-card__body">
            Where available, you can choose PayPal or a supported wallet at checkout for faster payment
            and additional buyer protection. You will be redirected back to our site once the payment is
            approved.
          </p>
          <p class="support-card__body support-card__highlight">
            PayPal and credit cards: 3.5% of the order amount.
          </p>
        </article>
        <article class="support-card">
          <h4 class="support-card__title">
            <span class="support-card__icon">
              <img src="/icons/payment/bank-transfer.svg" alt="Bank transfer" class="support-card__icon-image" />
            </span>
            <span>Bank transfer (on request)</span>
          </h4>
          <p class="support-card__body">
            For larger or custom orders we can sometimes arrange payment via bank transfer. Please
            contact support before placing the order so that we can provide the correct account details
            and reserve your items.
          </p>
          <p class="support-card__body support-card__highlight">
            Bank transfer: $45 USD bank fee added to your order. The fee from your bank will vary.
          </p>
        </article>
        <article class="support-card">
          <h4 class="support-card__title">
            <span class="support-card__icon-group">
              <span class="support-card__icon">
                <img src="/icons/payment/wechatpay.svg" alt="WeChat Pay" class="support-card__icon-image" />
              </span>
              <span class="support-card__icon">
                <img src="/icons/payment/alipay.svg" alt="Alipay" class="support-card__icon-image" />
              </span>
            </span>
            <span>WeChat / Alipay</span>
          </h4>
          <p class="support-card__body">
            If you are a Chinese user, you can also use WeChat or Alipay to pay. You can pay on the
            payment page, or ask customer service for the payment code.
          </p>
          <p class="support-card__body support-card__highlight">
            WeChat payment: 0.6% of the order amount.
          </p>
          <p class="support-card__body support-card__highlight">
            Alipay payment: 1% of the order amount.
          </p>
        </article>
        <article class="support-card">
          <h4 class="support-card__title">
            <span class="support-card__icon">
              <img src="/icons/payment/stripe.svg" alt="Stripe" class="support-card__icon-image" />
            </span>
            <span>Stripe</span>
          </h4>
          <p class="support-card__body">
            Pay securely with major credit and debit cards through Stripe. Stripe handles the payment
            processing on our behalf, and your card details are never stored on our servers. Additional
            verification (3D Secure) may be required depending on your card issuer.
          </p>
          <p class="support-card__body support-card__highlight">
            Stripe payment: 3.5% of the order amount.
          </p>
        </article>
        <article class="support-card">
          <h4 class="support-card__title">
            <span class="support-card__icon">🌍</span>
            <span>WorldFirst</span>
          </h4>
          <p class="support-card__body">
            For selected international or business orders we can accept payment via WorldFirst. You pay
            from your local bank to a WorldFirst account in a supported currency, and WorldFirst handles
            the currency conversion and settlement to us.
          </p>
          <p class="support-card__body support-card__highlight">
            WorldFirst payment: 1% of the order amount.
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
      <h3 class="support-section__title">Example total order cost</h3>
      <article class="support-card">
        <p class="support-card__body">
          The following example is for illustration only. Actual amounts will depend on your currency,
          shipping destination, payment method, and local taxes.
        </p>
        <ul class="support-list support-list--example">
          <li class="support-list__item">
            <strong>Product subtotal</strong>: $1,000
          </li>
          <li class="support-list__item">
            <strong>Payment fee (PayPal 3.5%)</strong>: $35
          </li>
          <li class="support-list__item">
            <strong>Shipping cost</strong>: $60
          </li>
          <li class="support-list__item">
            <strong>Estimated duties &amp; tax</strong>: $120
          </li>
        </ul>
        <p class="support-card__body support-card__highlight">
          Approximate total: $1,215
        </p>
      </article>
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
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.support-page__intro {
  margin: 0 0 1.5rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.support-intro-banner {
  margin: 0 0 1.25rem;
  padding: 0.85rem 1rem;
  border-radius: 0.9rem;
  background-image: linear-gradient(135deg, rgba(34, 211, 238, 0.26), rgba(59, 130, 246, 0.35));
  border: 1px solid rgba(56, 189, 248, 0.7);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.8);
}

.support-intro-banner__text {
  margin: 0;
  font-size: 0.9rem;
  line-height: 1.6;
  color: #e5e7eb;
}

.support-section {
  margin-top: 1.75rem;
}

.support-section__title {
  margin: 0 0 0.75rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: #e5e7eb;
}

.support-section__grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.9rem;
}

.support-section__grid--two {
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
}

.support-card {
  border-radius: 0.75rem;
  background: rgba(15, 23, 42, 0.85);
  border: 1px solid rgba(148, 163, 184, 0.35);
  padding: 0.9rem 1rem;
  box-shadow: 0 10px 25px rgba(15, 23, 42, 0.7);
}

.support-card__title {
  margin: 0 0 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: #f9fafb;
  display: flex;
  align-items: center;
}

.support-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.2rem;
  height: 2.2rem;
  border-radius: 9999px;
  background: rgba(15, 23, 42, 0.9);
  border: none;
  font-size: 0.8rem;
  margin-right: 0.4rem;
}

.support-card__icon-image {
  width: 1.6rem;
  height: 1.6rem;
  object-fit: contain;
}

.support-card__icon-image + .support-card__icon-image {
  margin-left: 0.12rem;
}

.support-card__icon-group {
  display: inline-flex;
  align-items: center;
}

.support-card__icon-group .support-card__icon {
  margin-right: 0.2rem;
}

.support-card__icon-group .support-card__icon:last-child {
  margin-right: 0.4rem;
}

.support-card__body {
  margin: 0;
  font-size: 0.9rem;
  line-height: 1.55;
  color: rgba(148, 163, 184, 0.9);
}

.support-card__highlight {
  color: #fbbf24;
  font-weight: 600;
}

.support-card__actions {
  margin-top: 0.8rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.support-card__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  padding: 0.4rem 0.9rem;
  font-size: 0.85rem;
  font-weight: 600;
  border: 1px solid rgba(148, 163, 184, 0.6);
  background: rgba(15, 23, 42, 0.9);
  color: #e5e7eb;
  cursor: pointer;
  text-decoration: none;
}

.support-card__button-icon {
  margin-right: 0.3rem;
  font-size: 1rem;
}

.support-card__button--primary {
  border-color: rgba(56, 189, 248, 0.9);
  background-image: linear-gradient(135deg, rgba(56, 189, 248, 0.9), rgba(59, 130, 246, 0.95));
  color: #0b1020;
}

.support-list {
  margin: 0;
  padding-left: 1.1rem;
  font-size: 0.9rem;
  color: rgba(148, 163, 184, 0.9);
}

.support-list__item + .support-list__item {
  margin-top: 0.45rem;
}

.support-section__note {
  margin: 0;
  font-size: 0.9rem;
  line-height: 1.6;
  color: rgba(148, 163, 184, 0.9);
}

@media (min-width: 768px) {
  .support-section {
    margin-top: 2rem;
  }
}
</style>
