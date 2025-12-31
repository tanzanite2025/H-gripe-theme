<template>
  <section class="bg-transparent text-white">
    <div class="mx-auto max-w-6xl px-1 sm:px-4 pt-[3px] pb-0">
      <div class="mb-[9px] text-center">
        <h2 class="text-lg font-semibold text-white/90 lg:text-xl">{{ t('home.trust.title') }}</h2>
        <div
          class="mt-[6px] mb-[9px] h-1 w-14 mx-auto rounded-full bg-gradient-to-r from-[#2dd4bf] to-[#3b82f6] shadow-[0_0_18px_rgba(45,212,191,0.25)]"
        ></div>
      </div>

      <div class="grid grid-cols-1 gap-[9px] sm:grid-cols-2 lg:grid-cols-4">
        <div
          v-for="(card, index) in cards"
          :key="card.titleKey"
          class="rounded-2xl bg-white/5 p-5 shadow-[6px_6px_0_rgba(0,0,0,1)]"
        >
          <h3 class="text-base font-semibold text-white/95">{{ t(card.titleKey) }}</h3>
          <p class="mt-2 text-sm leading-relaxed text-white/70">{{ t(card.descriptionKey) }}</p>

          <ul v-if="index === 0" class="mt-4 space-y-3 text-sm text-white/80">
            <!-- 1. SSL -->
            <li class="flex items-center justify-center gap-2">
              <span class="text-teal-300">•</span>
              <span>{{ t('home.shopWithConfidence.items.0.bullets.0') }}</span>
              <span class="text-lg leading-none">🔒</span>
            </li>
            <!-- 2. Cards -->
            <li class="flex items-center justify-center gap-2 flex-wrap">
              <span class="text-teal-300">•</span>
              <span class="mr-1">{{ t('home.shopWithConfidence.items.0.bullets.1') }}</span>
              <div class="flex items-center gap-2 transition-all">
                <img src="/icons/payment/visa.svg" alt="Visa" class="h-5 w-auto" />
                <img src="/icons/payment/mastercard.svg" alt="Mastercard" class="h-5 w-auto" />
                <img src="/icons/payment/amex.svg" alt="Amex" class="h-5 w-auto" />
              </div>
            </li>
            <!-- 3. Wallets -->
            <li class="flex items-center justify-center gap-2 flex-wrap">
              <span class="text-teal-300">•</span>
              <span class="mr-1">{{ t('home.shopWithConfidence.items.0.bullets.2') }}</span>
              <div class="flex items-center gap-2 transition-all">
                <img src="/icons/payment/paypal.svg" alt="PayPal" class="h-5 w-auto" />
                <img src="/icons/payment/wechatpay.svg" alt="WeChat" class="h-5 w-auto" />
                <img src="/icons/payment/alipay.svg" alt="Alipay" class="h-5 w-auto" />
              </div>
            </li>
          </ul>

          <div v-if="index === 0" class="mt-5 text-center">
            <NuxtLink
              to="/support/payment"
              class="inline-flex items-center justify-center rounded-full border-none bg-[rgba(15,23,42,0.6)] shadow-[0_4px_16px_rgba(0,0,0,0.5)] px-6 py-1.5 text-xs font-semibold text-white/90 transition-all hover:translate-y-[-1px] hover:shadow-[0_6px_20px_rgba(0,0,0,0.6)]"
            >
              View More
            </NuxtLink>
          </div>

          <ul v-else class="mt-4 space-y-2 text-sm text-white/80">
            <li v-for="bulletKey in card.bullets" :key="bulletKey" class="flex gap-2">
              <span class="text-teal-300">•</span>
              <span>{{ t(bulletKey) }}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'

const { t } = useI18n()

const cards = computed(() => {
  return Array.from({ length: 4 }, (_, index) => ({
    titleKey: `home.shopWithConfidence.items.${index}.title`,
    descriptionKey: `home.shopWithConfidence.items.${index}.description`,
    bullets: [0, 1, 2].map((bulletIndex) => `home.shopWithConfidence.items.${index}.bullets.${bulletIndex}`),
  }))
})
</script>
