<template>
  <div class="w-full pt-4 md:pt-8 pb-6">
    <h1 class="sr-only">{{ t('company.nav.contact') }}</h1>

    <div class="max-w-6xl mx-auto">
      
      <!-- Hero / Introduction -->
      <div class="text-center mb-8 md:mb-10">
        <h2 class="text-xl md:text-2xl font-bold text-white mb-2">
          Global <span class="text-teal-400">Connectivity</span>
        </h2>
        <p class="text-slate-300 max-w-2xl mx-auto leading-relaxed text-sm">
          From our manufacturing base in Xiamen to our global headquarters in Hong Kong, we are positioned to serve partners worldwide with efficiency and precision.
        </p>
      </div>

      <!-- Address Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-8 max-w-6xl mx-auto">
        <!-- HK HQ -->
        <div class="relative bg-slate-900/50 backdrop-blur-sm border border-white/10 rounded-2xl md:rounded-3xl p-6 md:p-8 shadow-2xl overflow-hidden group hover:border-teal-500/30 transition-all duration-300 flex flex-col">
            <h3 class="text-lg font-bold text-white mb-1">Global Headquarters</h3>
            <div class="text-xs uppercase tracking-wider text-indigo-400 font-semibold mb-2">Hong Kong</div>
            <p class="text-slate-300 mb-4 flex-grow leading-relaxed text-sm">
                Strategic hub for international trade, logistics, and global customer support.
            </p>
            <div class="text-sm text-slate-400 font-mono bg-slate-800/50 p-3 rounded-lg border border-slate-700/50">
                Flat 1602, 16/F, Lucky Centre,<br>
                No.165-171 Wan Chai Road, Wan Chai,<br>
                Hong Kong
            </div>
        </div>

        <!-- China Factory -->
         <div class="group bg-[#11151e] rounded-2xl p-6 border border-slate-700/50 hover:border-teal-500/30 transition-all duration-300 flex flex-col">
            <h3 class="text-lg font-bold text-white mb-1">Manufacturing & R&D Base</h3>
            <div class="text-xs uppercase tracking-wider text-teal-400 font-semibold mb-2">Xiamen, China</div>
            <p class="text-slate-300 mb-4 flex-grow leading-relaxed text-sm">
                State-of-the-art facility for carbon fiber layup, precision molding, and rigorous testing.
            </p>
            <div class="text-sm text-slate-400 font-mono bg-slate-800/50 p-3 rounded-lg border border-slate-700/50">
                Building 6, No. 639 Tongji South Road,<br>
                Industrial Concentration Zone, Tong'an Dist,<br>
                Xiamen, Fujian, China
            </div>
        </div>
      </div>

      <!-- Bank Info (Folded/Compact) -->
      <div class="bg-gradient-to-br from-slate-900 via-[#11151e] to-slate-900 rounded-2xl border border-slate-700/50 overflow-hidden mb-12">
        <div class="p-4 md:p-8">
            <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
                <div>
                     <h3 class="text-lg font-bold text-white flex items-center gap-2">
                        <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" /></svg>
                        Wire Transfer Details (B2B/Wholesale)
                     </h3>
                     <p class="text-sm text-slate-400 mt-1">For wholesale partners and OEM/ODM clients.</p>
                </div>
                 <div class="px-3 py-1 rounded-full bg-yellow-500/10 border border-yellow-500/20 text-yellow-400/90 text-xs font-semibold flex items-center gap-2">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
                    Secure Information
                 </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-y-4 gap-x-8 text-sm border-t border-slate-800 pt-6">
                <div>
                    <span class="block text-slate-500 text-xs uppercase tracking-wider mb-1">Beneficiary Name</span>
                    <span class="text-slate-200 font-mono bg-slate-800/30 px-2 py-1 rounded inline-block">Top Sports Co., Limited</span>
                </div>
                <div>
                    <span class="block text-slate-500 text-xs uppercase tracking-wider mb-1">Bank Name</span>
                    <span class="text-slate-200 font-medium">Deutsche Bank AG, Hong Kong</span>
                </div>
                 <div class="md:col-span-2">
                    <span class="block text-slate-500 text-xs uppercase tracking-wider mb-1">Account & SWIFT Details</span>
                    <div class="flex items-center gap-2 text-slate-400 italic bg-slate-950/30 p-3 rounded-lg border border-dashed border-slate-800">
                        <svg class="w-5 h-5 text-slate-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        <span>For security reasons, full wire transfer details (Account No. & SWIFT) will be provided exclusively in your official <strong>Proforma Invoice (PI)</strong>.</span>
                    </div>
                </div>
            </div>
            
            <!-- Anti-Fraud Warning -->
            <div class="mt-6 pt-6 border-t border-red-900/30 flex items-start gap-3">
                 <svg class="w-5 h-5 text-red-400 shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                 <p class="text-xs text-red-300 leading-relaxed">
                    <strong>SECURITY NOTICE:</strong> We will <u>NEVER</u> change our bank account via email notification. If you receive any email claiming our bank details have changed, please contact us immediately via phone or WhatsApp for verification before making any payment.
                 </p>
            </div>
        </div>
      </div>

      <!-- Map Component -->
      <div class="mt-12">
        <ContactLocationMap />
      </div>

    </div>

    <!-- FAQ Section -->
    <div class="max-w-[960px] mx-auto mt-4 px-0">
      <PageFaq pageId="company-contact" />
    </div>

    <!-- Feedback Section -->
    <div class="max-w-4xl mx-auto px-0 mt-4">
      <UserFeedbackThread
        threadKey="company-contact"
        title="Share your feedback about our Service"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { definePageMeta, useHead, useI18n } from '#imports'
import ContactLocationMap from '~/components/ContactLocationMap.vue'
import { contactLocation } from '~/utils/contactLocation'
import PageFaq from '~/components/PageFaq.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

const { t } = useI18n()

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Contact Us',
  script: [
    {
      type: 'application/ld+json',
      children: JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'LocalBusiness',
        name: contactLocation.name,
        email: 'support@tanzanite.site',
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
      }),
    } as any,
  ],
})
</script>
