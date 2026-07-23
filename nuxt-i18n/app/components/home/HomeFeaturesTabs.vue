<template>
  <section class="bg-transparent text-white pt-6 pb-2">
    <div class="page-content-shell px-0 md:px-4">
      
      <!-- Tab Navigation -->
      <div class="flex justify-center mb-4 sm:mb-8">
        <div class="bg-slate-900/50 p-1 rounded-full border border-white/10 inline-flex">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            @click="currentTabId = tab.id"
            class="px-6 py-2 rounded-full text-sm font-medium transition-all duration-300"
            :class="currentTabId === tab.id ? 'bg-slate-800 tz-text-primary shadow-lg' : 'tz-text-secondary hover:text-white'"
          >
            {{ $t(tab.labelKey) }}
          </button>
        </div>
      </div>

      <!-- Main Content Area -->
      <div class="grid lg:grid-cols-12 gap-3 sm:gap-6 items-start">
        
        <!-- Left Column: Interactive List (Summary) -->
        <div class="lg:col-span-4 lg:pr-4 order-1 lg:self-center">
          <!-- Mobile Title (optional, since tabs show it) -->
          
          <ul class="space-y-1 sm:space-y-2">
            <li
              v-for="(item, index) in currentTabItems"
              :key="index"
              class="group cursor-pointer rounded-xl p-2 sm:p-3 transition-all duration-200 border border-transparent"
              :class="activeIndex === index ? 'bg-slate-800/60 border-slate-700/50 shadow-sm' : 'hover:bg-slate-800/30'"
              @click="activeIndex = index"
              @mouseenter="activeIndex = index"
            >
              <div class="flex items-center gap-3">
                <!-- Active Indicator Dot -->
                <div 
                  class="w-2 h-2 rounded-full transition-colors duration-200"
                  :class="activeIndex === index ? 'bg-teal-400 shadow-[0_0_8px_rgba(45,212,191,0.6)]' : 'bg-slate-600 group-hover:bg-slate-500'"
                ></div>
                
                <span 
                  class="text-sm font-medium transition-colors duration-200"
                  :class="activeIndex === index ? 'tz-text-primary' : 'tz-text-secondary group-hover:text-slate-200'"
                >
                  {{ $t(item.titleKey) }}
                </span>
              </div>
            </li>
          </ul>
        </div>

        <!-- Right Column: Visual Carousel -->
        <div class="lg:col-span-8 order-2">
          <StackedImageCarousel 
            :items="currentTabItems" 
            v-model="activeIndex"
          >
            <template #card="{ item, index }">
              <div class="h-full w-full rounded-2xl premium-card p-6 flex flex-col justify-center items-center text-center border border-white/10 bg-[#11151e]">
                
                <!-- Card Content -->
                <h3 class="text-xl font-bold text-white mb-3">{{ $t(item.titleKey) }}</h3>
                <p class="tz-text-secondary text-sm leading-relaxed max-w-lg mx-auto mb-6">
                  {{ $t(item.descriptionKey) }}
                </p>

                <!-- Special Content for Payment Card (Trust Tab Index 0) -->
                <div v-if="currentTabId === 'trust' && index === 0" class="w-full">
                  <div class="flex flex-wrap justify-center gap-4 mb-4">
                     <!-- Icons Row 1 -->
                     <div class="flex gap-2">
                        <img src="/icons/payment/visa.svg" alt="Visa" class="h-6 w-auto opacity-90" />
                        <img src="/icons/payment/mastercard.svg" alt="Mastercard" class="h-6 w-auto opacity-90" />
                        <img src="/icons/payment/amex.svg" alt="Amex" class="h-6 w-auto opacity-90" />
                     </div>
                     <!-- Icons Row 2 -->
                     <div class="flex gap-2">
                        <img src="/icons/payment/paypal.svg" alt="PayPal" class="h-6 w-auto opacity-90" />
                        <img src="/icons/payment/wechatpay.svg" alt="WeChat" class="h-6 w-auto opacity-90" />
                        <img src="/icons/payment/alipay.svg" alt="Alipay" class="h-6 w-auto opacity-90" />
                     </div>
                  </div>
                  <div class="flex items-center justify-center gap-2 text-teal-300 text-xs font-mono uppercase tracking-widest mb-4">
                    <span>SSL Secure Encyption</span>
                    <span>🔒</span>
                  </div>
                  
                  <NuxtLink to="/support/payment" class="premium-button text-sm px-6 py-2">
                    View Payment Details
                  </NuxtLink>
                </div>

                <!-- Standard Bullets for other cards -->
                <ul v-else class="space-y-2 text-sm text-left inline-block mx-auto">
                   <li v-for="bulletKey in item.bullets" :key="bulletKey" class="flex gap-3">
                      <span class="text-teal-400 mt-1">•</span>
                      <span class="tz-text-secondary">{{ $t(bulletKey) }}</span>
                   </li>
                </ul>

              </div>
            </template>
          </StackedImageCarousel>
        </div>

      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from '#imports'
import StackedImageCarousel from '~/components/StackedImageCarousel.vue'

const { t } = useI18n()

type TabId = 'trust' | 'why_us'

const tabs: { id: TabId; labelKey: string }[] = [
  { id: 'trust', labelKey: 'home.trust.title' },
  { id: 'why_us', labelKey: 'home.whyChooseUs.title' }
]

const currentTabId = ref<TabId>('trust')
const activeIndex = ref(0) // Syncs with carousel

// Reset index when tab changes
watch(currentTabId, () => {
  activeIndex.value = 0
})

// --- Data Definition ---

// Trust (Shop With Confidence) Data
const trustCards = computed(() => {
  return Array.from({ length: 4 }, (_, index) => ({
    titleKey: `home.shopWithConfidence.items.${index}.title`,
    descriptionKey: `home.shopWithConfidence.items.${index}.description`,
    bullets: [0, 1, 2].map((bulletIndex) => `home.shopWithConfidence.items.${index}.bullets.${bulletIndex}`),
    // Additional metadata for list view if needed
  }))
})

// Why Choose Us Data
const whyUsCards = computed(() => {
  return Array.from({ length: 6 }, (_, index) => ({
    titleKey: `home.whyChooseUs.items.${index}.title`,
    descriptionKey: `home.whyChooseUs.items.${index}.description`,
    bullets: [0, 1, 2].map((bulletIndex) => `home.whyChooseUs.items.${index}.bullets.${bulletIndex}`),
  }))
})

const currentTabItems = computed(() => {
  return currentTabId.value === 'trust' ? trustCards.value : whyUsCards.value
})

</script>

<style scoped>
/* Optional: specific visual tweaks if premium-card needs adjustment in this context */
</style>
