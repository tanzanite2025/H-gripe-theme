<template>
  <div class="w-full pb-6">
    <h1 class="sr-only">{{ t('company.nav.contact') }}</h1>

    <div class="w-full max-w-none space-y-8">
      <!-- Location introduction -->
      <div class="text-center">
        <h2 class="text-xl font-bold tz-text-primary md:text-2xl">
          {{ t('companyContact.hero.title') }}
          {{ t('companyContact.hero.joiner') }}
          <span class="text-emerald-600">{{ t('companyContact.hero.highlight') }}</span>
        </h2>
        <p class="mx-auto max-w-2xl text-sm leading-relaxed tz-text-secondary">
          {{ t('companyContact.hero.description') }}
        </p>
      </div>

      <ContactServiceEntry />

      <!-- Address Grid -->
      <div class="grid w-full grid-cols-1 gap-6 md:grid-cols-2 md:gap-8">
        <!-- HK HQ -->
        <div class="premium-card group flex flex-col rounded-2xl p-6 transition-all duration-300 hover:-translate-y-0.5 md:p-8">
            <h3 class="text-lg font-bold tz-text-primary mb-1">{{ t('companyContact.locations.hongKong.title') }}</h3>
            <div class="text-xs uppercase tracking-wider text-emerald-600 font-semibold mb-2">{{ t('companyContact.locations.hongKong.region') }}</div>
            <p class="tz-text-secondary mb-4 flex-grow leading-relaxed text-sm">
                {{ t('companyContact.locations.hongKong.description') }}
            </p>
            <div class="rounded-lg border tz-border-subtle bg-[var(--tz-form-panel-surface)] p-3 text-sm tz-text-secondary">
                Flat 1602, 16/F, Lucky Centre,<br>
                No.165-171 Wan Chai Road, Wan Chai,<br>
                Hong Kong
            </div>
        </div>

        <!-- China Factory -->
         <div class="premium-card group flex flex-col rounded-2xl p-6 transition-all duration-300 hover:-translate-y-0.5 md:p-8">
            <h3 class="text-lg font-bold tz-text-primary mb-1">{{ t('companyContact.locations.xiamen.title') }}</h3>
            <div class="text-xs uppercase tracking-wider text-emerald-600 font-semibold mb-2">{{ t('companyContact.locations.xiamen.region') }}</div>
            <p class="tz-text-secondary mb-4 flex-grow leading-relaxed text-sm">
                {{ t('companyContact.locations.xiamen.description') }}
            </p>
            <div class="rounded-lg border tz-border-subtle bg-[var(--tz-form-panel-surface)] p-3 text-sm tz-text-secondary">
                Building 6, No. 639 Tongji South Road,<br>
                Industrial Concentration Zone, Tong'an Dist,<br>
                Xiamen, Fujian, China
            </div>
        </div>
      </div>

      <!-- Bank Info (Folded/Compact) -->
      <section class="premium-card overflow-hidden rounded-2xl">
        <div class="p-4 md:p-8">
            <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
                <div>
                     <h3 class="text-lg font-bold tz-text-primary flex items-center gap-2">
        <svg class="w-5 h-5 tz-text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" /></svg>
                        {{ t('companyContact.bank.title') }}
                     </h3>
                     <p class="text-sm tz-text-secondary mt-1">{{ t('companyContact.bank.description') }}</p>
                </div>
                 <div class="px-3 py-1 rounded-full bg-yellow-500/10 border border-yellow-500/20 text-yellow-400/90 text-xs font-semibold flex items-center gap-2">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
                    {{ t('companyContact.bank.secureInfo') }}
                 </div>
            </div>

            <div class="grid grid-cols-1 gap-x-8 gap-y-4 border-t tz-border-subtle pt-6 text-sm md:grid-cols-2">
                <div>
                    <span class="block tz-text-muted text-xs uppercase tracking-wider mb-1">{{ t('companyContact.bank.beneficiaryName') }}</span>
        <span class="inline-block rounded border tz-border-subtle bg-[var(--tz-form-panel-surface)] px-2 py-1 tz-text-primary">Top Sports Co., Limited</span>
                </div>
                <div>
                    <span class="block tz-text-muted text-xs uppercase tracking-wider mb-1">{{ t('companyContact.bank.bankName') }}</span>
        <span class="tz-text-primary font-medium">Deutsche Bank AG, Hong Kong</span>
                </div>
                 <div class="md:col-span-2">
                    <span class="block tz-text-muted text-xs uppercase tracking-wider mb-1">{{ t('companyContact.bank.accountDetails') }}</span>
                    <div class="flex items-center gap-2 rounded-lg border border-dashed tz-border-subtle bg-[var(--tz-form-control-surface)] p-3 italic tz-text-secondary">
                        <svg class="w-5 h-5 tz-text-muted shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        <span>{{ t('companyContact.bank.accountDetailsIntro') }}<strong>{{ t('companyContact.bank.proformaInvoice') }}</strong>{{ t('companyContact.bank.accountDetailsEnding') }}</span>
                    </div>
                </div>
            </div>
            
            <!-- Anti-Fraud Warning -->
            <div class="mt-6 pt-6 border-t border-red-900/30 flex items-start gap-3">
                 <svg class="w-5 h-5 text-red-400 shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                 <p class="text-xs text-red-300 leading-relaxed">
                    <strong>{{ t('companyContact.bank.securityNoticeTitle') }}:</strong>
                    {{ t('companyContact.bank.securityNoticeBefore') }}
                    <u>{{ t('companyContact.bank.securityNoticeNever') }}</u>
                    {{ t('companyContact.bank.securityNoticeAfter') }}
                 </p>
            </div>
        </div>
      </section>

      <!-- Map Component -->
      <section class="premium-card rounded-2xl p-4 md:p-6">
        <ContactLocationMap />
      </section>

    </div>

  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { definePageMeta, useHead, useI18n } from '#imports'
import ContactLocationMap from '~/components/ContactLocationMap.vue'
import ContactServiceEntry from '~/components/company/ContactServiceEntry.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import { contactLocation } from '~/utils/contactLocation'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('companyContact')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

definePageMeta({
  layout: 'products',
  footerLabelKey: 'company.nav.contact',
  footerLabelFallback: 'Contact Us',
})

useHead(() => ({
  title: t('companyContact.title'),
  script: [createSeoJsonLdScript({
    '@context': 'https://schema.org',
    '@type': 'LocalBusiness',
    name: contactLocation.name,
    address: {
      '@type': 'PostalAddress',
      streetAddress: contactLocation.addressText,
      addressLocality: 'Xiamen',
      addressRegion: 'Fujian',
      addressCountry: 'CN',
    },
    ...(contactLocation.lat && contactLocation.lng
      ? {
          geo: {
            '@type': 'GeoCoordinates',
            latitude: contactLocation.lat,
            longitude: contactLocation.lng,
          },
        }
      : {}),
  })],
}))
</script>
