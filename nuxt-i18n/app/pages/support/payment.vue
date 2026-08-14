<template>
  <div class="support-page pb-8">
    <h1 class="sr-only">Payment</h1>

    <section class="rounded-2xl bg-[#111116] px-5 py-6 shadow-[0_16px_42px_rgba(0,0,0,0.45)] md:px-8 md:py-8">
      <div class="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
        <div class="max-w-3xl">
          <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#B5FF6D]">
            Secure payment
          </p>
          <h2 class="mt-2 text-xl font-semibold leading-tight text-slate-50 md:text-2xl">
            Pay with a method you recognize, with protection at every step.
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 tz-text-secondary">
            We support trusted payment providers and verify payment results on the server before an
            order is treated as paid. Sensitive card details stay with the payment provider instead
            of being stored by our store.
          </p>
        </div>

        <div class="flex shrink-0 items-center gap-3 rounded-xl bg-[#0b0b0e] px-4 py-3 text-sm text-slate-200">
          <span class="h-2.5 w-2.5 rounded-full bg-[#B5FF6D] shadow-[0_0_14px_rgba(181,255,109,0.75)]" />
          <span>HTTPS encrypted checkout</span>
        </div>
      </div>

      <div class="mt-6 flex flex-wrap gap-3">
        <NuxtLink
          to="/company/ourstory"
          target="_blank"
          rel="noopener"
          class="inline-flex items-center justify-center rounded-full bg-[#1a1a20] px-4 py-2.5 text-sm font-semibold text-slate-200 shadow-[0_8px_20px_rgba(0,0,0,0.3)] transition-colors hover:bg-[#24242c] hover:text-white"
        >
          Our Story
        </NuxtLink>
        <NuxtLink
          to="/support/shipping"
          target="_blank"
          rel="noopener"
          class="inline-flex items-center justify-center rounded-full bg-[#1a1a20] px-4 py-2.5 text-sm font-semibold text-slate-200 shadow-[0_8px_20px_rgba(0,0,0,0.3)] transition-colors hover:bg-[#24242c] hover:text-white"
        >
          Shipping instructions
        </NuxtLink>
        <NuxtLink
          to="/support/test-report"
          target="_blank"
          rel="noopener"
          class="inline-flex items-center justify-center rounded-full bg-[#1a1a20] px-4 py-2.5 text-sm font-semibold text-slate-200 shadow-[0_8px_20px_rgba(0,0,0,0.3)] transition-colors hover:bg-[#24242c] hover:text-white"
        >
          Test report
        </NuxtLink>
      </div>
    </section>

    <section class="mt-10">
      <div class="mb-5">
        <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#B5FF6D]">
          Supported methods
        </p>
        <h2 class="mt-2 text-lg font-semibold text-slate-50 md:text-xl">
          Choose the payment option shown at checkout
        </h2>
        <p class="mt-2 max-w-3xl text-sm leading-7 tz-text-secondary">
          Payment availability can depend on the order and region. The checkout page is the final
          source for the methods currently enabled for your purchase.
        </p>
      </div>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <article class="flex h-full flex-col rounded-2xl bg-[#111116] p-5 shadow-[0_12px_30px_rgba(0,0,0,0.38)]">
          <div class="flex min-h-10 items-center gap-3">
            <div class="flex h-10 items-center gap-1 rounded-xl bg-[#17171d] px-2.5">
              <img src="/icons/payment/visa.svg" alt="Visa" class="h-5 w-auto" />
              <img src="/icons/payment/mastercard.svg" alt="Mastercard" class="h-5 w-auto" />
              <img src="/icons/payment/amex.svg" alt="American Express" class="h-5 w-auto" />
            </div>
          </div>
          <h3 class="mt-5 text-base font-semibold text-slate-50">Credit and debit cards</h3>
          <p class="mt-2 text-sm leading-6 tz-text-secondary">
            Visa, Mastercard, American Express, and other major cards are processed through Stripe.
          </p>
          <div class="mt-auto pt-5">
            <span class="inline-flex w-fit items-center rounded-full border border-white/10 bg-[#17171d] px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.08em] text-slate-400">
              Stripe card checkout
            </span>
          </div>
        </article>

        <article class="flex h-full flex-col rounded-2xl bg-[#111116] p-5 shadow-[0_12px_30px_rgba(0,0,0,0.38)]">
          <div class="flex min-h-10 items-center">
            <div class="flex h-10 items-center rounded-xl bg-[#17171d] px-3">
              <img src="/icons/payment/paypal.svg?v=4" alt="PayPal" class="h-6 w-auto" />
            </div>
          </div>
          <h3 class="mt-5 text-base font-semibold text-slate-50">PayPal</h3>
          <p class="mt-2 text-sm leading-6 tz-text-secondary">
            Continue to PayPal's hosted checkout, approve the payment there, and return to our site
            after confirmation.
          </p>
          <div class="mt-auto pt-5">
            <span class="inline-flex w-fit items-center rounded-full border border-white/10 bg-[#17171d] px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.08em] text-slate-400">
              Hosted PayPal checkout
            </span>
          </div>
        </article>

        <article class="flex h-full flex-col rounded-2xl bg-[#111116] p-5 shadow-[0_12px_30px_rgba(0,0,0,0.38)]">
          <div class="flex min-h-10 items-center">
            <div class="flex h-10 items-center rounded-xl bg-[#17171d] px-3">
              <img src="/icons/payment/alipay.svg?v=6" alt="Alipay" class="h-6 w-auto" />
            </div>
          </div>
          <h3 class="mt-5 text-base font-semibold text-slate-50">Alipay</h3>
          <p class="mt-2 text-sm leading-6 tz-text-secondary">
            Continue to Alipay to complete payment. The provider notification is verified before the
            order is marked as paid.
          </p>
          <div class="mt-auto pt-5">
            <span class="inline-flex w-fit items-center rounded-full border border-white/10 bg-[#17171d] px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.08em] text-slate-400">
              Verified provider notification
            </span>
          </div>
        </article>

        <article class="flex h-full flex-col rounded-2xl bg-[#111116] p-5 shadow-[0_12px_30px_rgba(0,0,0,0.38)]">
          <div class="flex min-h-10 items-center">
            <div class="flex h-10 items-center rounded-xl bg-[#17171d] px-3">
              <img src="/icons/payment/wechatpay.svg?v=4" alt="WeChat Pay" class="h-6 w-auto" />
            </div>
          </div>
          <h3 class="mt-5 text-base font-semibold text-slate-50">WeChat Pay</h3>
          <p class="mt-2 text-sm leading-6 tz-text-secondary">
            Scan the WeChat Pay QR code shown during checkout. Payment status is confirmed with the
            provider before the order is completed.
          </p>
          <div class="mt-auto pt-5">
            <span class="inline-flex w-fit items-center rounded-full border border-white/10 bg-[#17171d] px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.08em] text-slate-400">
              Native QR payment
            </span>
          </div>
        </article>
      </div>
    </section>

    <section class="mt-10">
      <div class="mb-5">
        <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#B5FF6D]">
          Security measures
        </p>
        <h2 class="mt-2 text-lg font-semibold text-slate-50 md:text-xl">
          What we do to protect your payment
        </h2>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <article class="rounded-2xl bg-[#0b0b0e] p-5 shadow-[0_12px_30px_rgba(0,0,0,0.52)] md:p-6">
          <h3 class="text-base font-semibold text-slate-50">Protection built into the payment flow</h3>
          <div class="mt-5 grid gap-3">
            <div class="rounded-xl bg-[#17171d] p-4">
              <h4 class="text-sm font-semibold text-slate-100">Encrypted connection</h4>
              <p class="mt-1.5 text-sm leading-6 tz-text-secondary">
                Every page uses HTTPS/TLS so payment-related data is encrypted while it travels
                between your browser and our services.
              </p>
            </div>
            <div class="rounded-xl bg-[#17171d] p-4">
              <h4 class="text-sm font-semibold text-slate-100">Trusted payment providers</h4>
              <p class="mt-1.5 text-sm leading-6 tz-text-secondary">
                Card and wallet payment details are entered into Stripe, PayPal, Alipay, or WeChat
                Pay rather than being handled as raw card data by our storefront.
              </p>
            </div>
            <div class="rounded-xl bg-[#17171d] p-4">
              <h4 class="text-sm font-semibold text-slate-100">Server-side verification</h4>
              <p class="mt-1.5 text-sm leading-6 tz-text-secondary">
                Provider webhooks or verified status checks are matched to the order amount and
                currency before payment is finalized.
              </p>
            </div>
          </div>
        </article>

        <article class="rounded-2xl bg-[#0b0b0e] p-5 shadow-[0_12px_30px_rgba(0,0,0,0.52)] md:p-6">
          <h3 class="text-base font-semibold text-slate-50">What we do not store</h3>
          <p class="mt-3 text-sm leading-7 tz-text-secondary">
            We do not store your full card number, CVV, or card PIN. We receive the payment result
            and the provider reference needed to reconcile your order and provide support.
          </p>

          <div class="mt-6 rounded-xl bg-[#0b0b0e] p-4">
            <p class="text-sm font-semibold text-slate-100">Extra verification may appear</p>
            <p class="mt-1.5 text-sm leading-6 tz-text-secondary">
              Your bank or card issuer may request 3D Secure verification in its app, by one-time
              code, or through another identity check. That step is controlled by the issuer and
              helps prevent unauthorized card use.
            </p>
          </div>

          <div class="mt-6 rounded-xl bg-[#17171d] p-4">
            <p class="text-sm font-semibold text-slate-100">We will never ask for your PIN</p>
            <p class="mt-1.5 text-sm leading-6 tz-text-secondary">
              Do not send your full card number, CVV, password, or card PIN through chat or email.
              Contact support if a payment message looks suspicious.
            </p>
          </div>
        </article>
      </div>
    </section>

    <section class="mt-10">
      <div class="mb-5">
        <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#B5FF6D]">
          Common concerns
        </p>
        <h2 class="mt-2 text-lg font-semibold text-slate-50 md:text-xl">
          If something does not look right
        </h2>
      </div>

      <div class="grid gap-4 md:grid-cols-2">
        <article class="rounded-2xl bg-[#111116] p-5 shadow-[0_12px_30px_rgba(0,0,0,0.38)]">
          <h3 class="text-base font-semibold text-slate-50">Payment declined</h3>
          <p class="mt-2 text-sm leading-7 tz-text-secondary">
            Check your billing details and available funds, then try again or choose another method
            shown at checkout. Your bank or payment provider can usually explain a decline in more
            detail.
          </p>
        </article>
        <article class="rounded-2xl bg-[#111116] p-5 shadow-[0_12px_30px_rgba(0,0,0,0.38)]">
          <h3 class="text-base font-semibold text-slate-50">Payment completed but the page did not update</h3>
          <p class="mt-2 text-sm leading-7 tz-text-secondary">
            Do not submit the payment repeatedly. Keep the provider reference or payment receipt
            and contact support so we can verify the order status on the server.
          </p>
        </article>
      </div>
    </section>

    <section class="mt-10 rounded-2xl bg-[#111116] p-5 text-center shadow-[0_12px_30px_rgba(0,0,0,0.38)] md:p-7">
      <h2 class="text-lg font-semibold text-slate-50">Need help with a payment?</h2>
      <p class="mx-auto mt-2 max-w-2xl text-sm leading-7 tz-text-secondary">
        Contact us with your order number or provider reference. Please do not include your full
        card number, CVV, password, or PIN.
      </p>
    </section>

    <section class="mt-10">
      <UserFeedbackThread
        threadKey="support-payment"
        title="Share your feedback about payment &amp; billing"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Payment',
})
</script>
