<template>
  <section class="bg-transparent text-white py-8 sm:py-12 lg:py-20">
    <div class="mx-auto max-w-6xl px-4 sm:px-6">
      
      <div class="grid lg:grid-cols-12 gap-5 sm:gap-10 lg:gap-16 items-start">
        
        <!-- Left Column: Header & Context -->
        <div class="lg:col-span-3 lg:sticky lg:top-32 self-start space-y-3 sm:space-y-6">
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-300 text-xs font-medium uppercase tracking-wider">
            <span>Our Collection</span>
          </div>

          <h2 class="text-2xl sm:text-3xl font-bold text-white leading-tight">
             {{ t('home.featuredProducts.title') }}
          </h2>
          
          <div class="h-1 w-20 rounded-full bg-gradient-to-r from-blue-400 to-purple-500 shadow-[0_0_12px_rgba(59,130,246,0.5)]"></div>
          
          <p class="text-base text-slate-400 leading-relaxed max-w-md">
            {{ t('home.featuredProducts.subtitle') }}
          </p>

          <div class="pt-2 sm:pt-4">
             <NuxtLink to="/shop" class="premium-button inline-flex items-center px-6 py-3 text-sm font-medium">
                {{ t('home.featuredProducts.viewAll') }}
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" class="ml-2 h-5 w-5"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3" /></svg>
             </NuxtLink>
          </div>
        </div>

        <!-- Right Column: Product Grid -->
        <div class="lg:col-span-9 grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6">
           <NuxtLink 
              v-for="card in cards" 
              :key="card.titleKey" 
              to="/shop" 
              class="group block overflow-hidden rounded-2xl premium-card relative hover:shadow-2xl hover:shadow-blue-900/20 transition-all duration-500"
           >
              <!-- Image Aspect -->
              <div class="relative aspect-[4/3] bg-slate-800 overflow-hidden">
                 <!-- Placeholder Gradient / Image Slot -->
                 <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(59,130,246,0.15),transparent_60%)] group-hover:scale-110 transition-transform duration-700 ease-out"></div>
                 
                 <!-- Gradient Overlay -->
                 <div class="absolute inset-0 bg-gradient-to-t from-black/90 via-black/40 to-transparent"></div>
                 
                 <div class="absolute bottom-0 inset-x-0 p-5">
                    <h3 class="text-lg font-bold text-white mb-1 group-hover:text-blue-300 transition-colors">{{ t(card.titleKey) }}</h3>
                    <p class="text-white/70 text-sm line-clamp-2 mb-3">{{ t(card.descriptionKey) }}</p>
                    <div class="inline-block px-3 py-1 rounded-lg bg-white/10 backdrop-blur text-xs font-medium text-white/90 border border-white/10 group-hover:bg-blue-500/20 group-hover:border-blue-500/30 transition-colors">
                       {{ t(card.priceKey) }}
                    </div>
                 </div>
              </div>
           </NuxtLink>
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
  return [
    {
      titleKey: 'home.featuredProducts.items.0.title',
      descriptionKey: 'home.featuredProducts.items.0.description',
      priceKey: 'home.featuredProducts.items.0.price',
    },
    {
      titleKey: 'home.featuredProducts.items.1.title',
      descriptionKey: 'home.featuredProducts.items.1.description',
      priceKey: 'home.featuredProducts.items.1.price',
    },
    {
      titleKey: 'home.featuredProducts.items.2.title',
      descriptionKey: 'home.featuredProducts.items.2.description',
      priceKey: 'home.featuredProducts.items.2.price',
    },
    {
      titleKey: 'home.featuredProducts.items.3.title',
      descriptionKey: 'home.featuredProducts.items.3.description',
      priceKey: 'home.featuredProducts.items.3.price',
    },
  ]
})
</script>
