<template>
  <div class="support-page pb-8">
    <h1 class="sr-only">{{ t('supportPayment.title') }}</h1>

<section class="rounded-2xl tz-surface-card px-5 py-6 shadow-[0_16px_42px_rgba(15,23,42,0.12)] md:px-8 md:py-8">
      <div class="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
        <div class="max-w-3xl">
          <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#059669]">
            {{ t('supportPayment.hero.eyebrow') }}
          </p>
          <h2 class="mt-2 text-xl font-semibold leading-tight tz-text-primary md:text-2xl">
            {{ t('supportPayment.hero.title') }}
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 tz-text-secondary">
            {{ t('supportPayment.hero.description') }}
          </p>
        </div>

<div class="flex shrink-0 items-center gap-3 rounded-xl tz-surface-muted px-4 py-3 text-sm tz-text-secondary">
          <span class="h-2.5 w-2.5 rounded-full bg-[#059669] shadow-[0_0_14px_rgba(5, 150, 105,0.75)]" />
          <span>{{ t('supportPayment.hero.https') }}</span>
        </div>
      </div>

      <div class="mt-6 flex flex-wrap gap-3">
        <NuxtLink
          :to="localePath('/company/ourstory')"
          target="_blank"
          rel="noopener"
class="inline-flex items-center justify-center rounded-full tz-surface-muted px-4 py-2.5 text-sm font-semibold tz-text-secondary shadow-[0_8px_20px_rgba(15,23,42,0.1)] transition-colors hover:tz-surface-subtle hover:tz-text-primary"
        >
          {{ t('supportPayment.hero.links.ourStory') }}
        </NuxtLink>
        <NuxtLink
          :to="localePath('/support/shipping')"
          target="_blank"
          rel="noopener"
          class="inline-flex items-center justify-center rounded-full tz-surface-muted px-4 py-2.5 text-sm font-semibold tz-text-secondary shadow-[0_8px_20px_rgba(20,32,43,0.1)] transition-colors hover:tz-surface-subtle hover:tz-text-primary"
        >
          {{ t('supportPayment.hero.links.shipping') }}
        </NuxtLink>
        <NuxtLink
          :to="localePath('/support/test-report')"
          target="_blank"
          rel="noopener"
          class="inline-flex items-center justify-center rounded-full tz-surface-muted px-4 py-2.5 text-sm font-semibold tz-text-secondary shadow-[0_8px_20px_rgba(20,32,43,0.1)] transition-colors hover:tz-surface-subtle hover:tz-text-primary"
        >
          {{ t('supportPayment.hero.links.testReport') }}
        </NuxtLink>
      </div>
    </section>

    <section class="mt-10">
      <div class="mb-5 tz-mobile-text-inset">
        <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#059669]">
          {{ t('supportPayment.methods.eyebrow') }}
        </p>
        <h2 class="mt-2 text-lg font-semibold tz-text-primary md:text-xl">
          {{ t('supportPayment.methods.title') }}
        </h2>
        <p class="mt-2 max-w-3xl text-sm leading-7 tz-text-secondary">
          {{ t('supportPayment.methods.description') }}
        </p>
      </div>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
<article class="flex h-full flex-col rounded-2xl tz-surface-card p-5 shadow-[0_12px_30px_rgba(15,23,42,0.1)]">
          <div class="flex min-h-10 items-center gap-3">
<div class="flex h-10 items-center gap-1 rounded-xl tz-surface-muted px-2.5">
              <img src="/icons/payment/visa.svg" alt="Visa" width="200" height="120" class="h-5 w-auto" />
              <img src="/icons/payment/mastercard.svg" alt="Mastercard" width="200" height="120" class="h-5 w-auto" />
              <img src="/icons/payment/amex.svg" alt="American Express" width="200" height="120" class="h-5 w-auto" />
            </div>
          </div>
          <h3 class="mt-5 text-base font-semibold tz-text-primary">
            {{ t('supportPayment.methods.cards.creditDebit.title') }}
          </h3>
          <p class="mt-2 text-sm leading-6 tz-text-secondary">
            {{ t('supportPayment.methods.cards.creditDebit.body') }}
          </p>
          <div class="mt-auto pt-5">
<span class="inline-flex w-fit items-center rounded-full border tz-border-subtle tz-surface-muted px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.08em] tz-text-secondary">
              {{ t('supportPayment.methods.cards.creditDebit.badge') }}
            </span>
          </div>
        </article>

<article class="flex h-full flex-col rounded-2xl tz-surface-card p-5 shadow-[0_12px_30px_rgba(15,23,42,0.1)]">
          <div class="flex min-h-10 items-center">
<div class="flex h-10 items-center rounded-xl tz-surface-muted px-3">
              <img src="/icons/payment/paypal.svg?v=4" alt="PayPal" width="200" height="120" class="h-6 w-auto" />
            </div>
          </div>
          <h3 class="mt-5 text-base font-semibold tz-text-primary">
            {{ t('supportPayment.methods.cards.paypal.title') }}
          </h3>
          <p class="mt-2 text-sm leading-6 tz-text-secondary">
            {{ t('supportPayment.methods.cards.paypal.body') }}
          </p>
          <div class="mt-auto pt-5">
<span class="inline-flex w-fit items-center rounded-full border tz-border-subtle tz-surface-muted px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.08em] tz-text-secondary">
              {{ t('supportPayment.methods.cards.paypal.badge') }}
            </span>
          </div>
        </article>

<article class="flex h-full flex-col rounded-2xl tz-surface-card p-5 shadow-[0_12px_30px_rgba(15,23,42,0.1)]">
          <div class="flex min-h-10 items-center">
<div class="flex h-10 items-center rounded-xl tz-surface-muted px-3">
              <img src="/icons/payment/alipay.svg?v=6" alt="Alipay" width="200" height="120" class="h-6 w-auto" />
            </div>
          </div>
          <h3 class="mt-5 text-base font-semibold tz-text-primary">
            {{ t('supportPayment.methods.cards.alipay.title') }}
          </h3>
          <p class="mt-2 text-sm leading-6 tz-text-secondary">
            {{ t('supportPayment.methods.cards.alipay.body') }}
          </p>
          <div class="mt-auto pt-5">
<span class="inline-flex w-fit items-center rounded-full border tz-border-subtle tz-surface-muted px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.08em] tz-text-secondary">
              {{ t('supportPayment.methods.cards.alipay.badge') }}
            </span>
          </div>
        </article>

<article class="flex h-full flex-col rounded-2xl tz-surface-card p-5 shadow-[0_12px_30px_rgba(15,23,42,0.1)]">
          <div class="flex min-h-10 items-center">
<div class="flex h-10 items-center rounded-xl tz-surface-muted px-3">
              <img src="/icons/payment/wechatpay.svg?v=4" alt="WeChat Pay" width="200" height="120" class="h-6 w-auto" />
            </div>
          </div>
          <h3 class="mt-5 text-base font-semibold tz-text-primary">
            {{ t('supportPayment.methods.cards.wechatPay.title') }}
          </h3>
          <p class="mt-2 text-sm leading-6 tz-text-secondary">
            {{ t('supportPayment.methods.cards.wechatPay.body') }}
          </p>
          <div class="mt-auto pt-5">
<span class="inline-flex w-fit items-center rounded-full border tz-border-subtle tz-surface-muted px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.08em] tz-text-secondary">
              {{ t('supportPayment.methods.cards.wechatPay.badge') }}
            </span>
          </div>
        </article>
      </div>
    </section>

    <section class="mt-10">
      <div class="mb-5 tz-mobile-text-inset">
        <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#059669]">
          {{ t('supportPayment.security.eyebrow') }}
        </p>
        <h2 class="mt-2 text-lg font-semibold tz-text-primary md:text-xl">
          {{ t('supportPayment.security.title') }}
        </h2>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
<article class="rounded-2xl tz-surface-card p-5 shadow-[0_12px_30px_rgba(15,23,42,0.12)] md:p-6">
          <h3 class="text-base font-semibold tz-text-primary">
            {{ t('supportPayment.security.flow.title') }}
          </h3>
          <div class="mt-5 grid gap-3">
<div class="rounded-xl tz-surface-muted p-4">
              <h4 class="text-sm font-semibold tz-text-secondary">
                {{ t('supportPayment.security.flow.encrypted.title') }}
              </h4>
              <p class="mt-1.5 text-sm leading-6 tz-text-secondary">
                {{ t('supportPayment.security.flow.encrypted.body') }}
              </p>
            </div>
<div class="rounded-xl tz-surface-muted p-4">
              <h4 class="text-sm font-semibold tz-text-secondary">
                {{ t('supportPayment.security.flow.trusted.title') }}
              </h4>
              <p class="mt-1.5 text-sm leading-6 tz-text-secondary">
                {{ t('supportPayment.security.flow.trusted.body') }}
              </p>
            </div>
<div class="rounded-xl tz-surface-muted p-4">
              <h4 class="text-sm font-semibold tz-text-secondary">
                {{ t('supportPayment.security.flow.server.title') }}
              </h4>
              <p class="mt-1.5 text-sm leading-6 tz-text-secondary">
                {{ t('supportPayment.security.flow.server.body') }}
              </p>
            </div>
          </div>
        </article>

<article class="rounded-2xl tz-surface-card p-5 shadow-[0_12px_30px_rgba(15,23,42,0.12)] md:p-6">
          <h3 class="text-base font-semibold tz-text-primary">
            {{ t('supportPayment.security.storage.title') }}
          </h3>
          <p class="mt-3 text-sm leading-7 tz-text-secondary">
            {{ t('supportPayment.security.storage.body') }}
          </p>

<div class="mt-6 rounded-xl tz-surface-card p-4">
            <p class="text-sm font-semibold tz-text-secondary">
              {{ t('supportPayment.security.storage.verification.title') }}
            </p>
            <p class="mt-1.5 text-sm leading-6 tz-text-secondary">
              {{ t('supportPayment.security.storage.verification.body') }}
            </p>
          </div>

<div class="mt-6 rounded-xl tz-surface-muted p-4">
            <p class="text-sm font-semibold tz-text-secondary">
              {{ t('supportPayment.security.storage.pin.title') }}
            </p>
            <p class="mt-1.5 text-sm leading-6 tz-text-secondary">
              {{ t('supportPayment.security.storage.pin.body') }}
            </p>
          </div>
        </article>
      </div>
    </section>

    <section class="mt-10">
      <div class="mb-5 tz-mobile-text-inset">
        <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#059669]">
          {{ t('supportPayment.concerns.eyebrow') }}
        </p>
        <h2 class="mt-2 text-lg font-semibold tz-text-primary md:text-xl">
          {{ t('supportPayment.concerns.title') }}
        </h2>
      </div>

      <div class="grid gap-4 md:grid-cols-2">
<article class="rounded-2xl tz-surface-card p-5 shadow-[0_12px_30px_rgba(15,23,42,0.1)]">
          <h3 class="text-base font-semibold tz-text-primary">
            {{ t('supportPayment.concerns.declined.title') }}
          </h3>
          <p class="mt-2 text-sm leading-7 tz-text-secondary">
            {{ t('supportPayment.concerns.declined.body') }}
          </p>
        </article>
<article class="rounded-2xl tz-surface-card p-5 shadow-[0_12px_30px_rgba(15,23,42,0.1)]">
          <h3 class="text-base font-semibold tz-text-primary">
            {{ t('supportPayment.concerns.completed.title') }}
          </h3>
          <p class="mt-2 text-sm leading-7 tz-text-secondary">
            {{ t('supportPayment.concerns.completed.body') }}
          </p>
        </article>
      </div>
    </section>

<section class="mt-10 rounded-2xl tz-surface-card p-5 text-center shadow-[0_12px_30px_rgba(15,23,42,0.1)] md:p-7">
      <h2 class="text-lg font-semibold tz-text-primary">
        {{ t('supportPayment.help.title') }}
      </h2>
      <p class="mx-auto mt-2 max-w-2xl text-sm leading-7 tz-text-secondary">
        {{ t('supportPayment.help.body') }}
      </p>
    </section>

    <section class="mt-10">
      <UserFeedbackThread
        threadKey="support-payment"
        :title="t('supportPayment.feedbackTitle')"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { definePageMeta, useHead, useI18n, useLocalePath } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'support',
  footerLabelKey: 'support.nav.payment',
  footerLabelFallback: 'Payment',
})

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages } = usePageMessages('supportPayment')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

useHead(() => ({
  title: t('supportPayment.title'),
}))
</script>
