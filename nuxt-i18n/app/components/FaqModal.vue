<template>
  <div class="max-w-[1400px] w-full h-[90vh] md:h-[700px] max-h-[80vh] md:max-h-[85vh] bg-black rounded-2xl border-2 border-[#6b73ff] shadow-[0_0_30px_rgba(107,115,255,0.3)] overflow-hidden flex flex-col">
    <!-- Header -->
    <div class="flex items-center justify-between px-6 py-3 md:py-4 border-b border-white/10">
      <h2 class="text-xl md:text-2xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-[#40ffaa] to-[#6b73ff]">
        {{ $t('faq.title', 'Quick Help') }}
      </h2>
      <button 
        @click="$emit('close')"
        class="w-10 h-10 rounded-full bg-white/5 hover:bg-white/10 border border-white/20 flex items-center justify-center transition-all duration-200 text-white/60 hover:text-white"
        aria-label="Close FAQ"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto flex flex-col">
      <!-- Search Filter - 固定在顶部 -->
      <div class="sticky top-0 z-10 bg-black border-b border-white/10 px-6 py-3 md:py-4">
        <ContentSearchFilter
          :options="{
            showCategories: true,
            showTags: false,
            showSortBy: false,
            showDateRange: false
          }"
          :categories="faqCategories"
          theme="dark"
          compact
          @update:filters="handleFiltersUpdate"
          @search="handleSearch"
        />
      </div>

      <!-- FAQ Content Area -->
      <div class="flex-1 px-6 py-6">
        <!-- Loading State -->
      <div v-if="pending" class="flex flex-col items-center justify-center py-20">
        <div class="inline-block animate-spin rounded-full h-12 w-12 border-4 border-[#6b73ff] border-t-transparent"></div>
        <p class="mt-4 text-white/60">{{ $t('common.loading', 'Loading') }}...</p>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="flex flex-col items-center justify-center py-20">
        <div class="text-red-500 text-xl mb-4">⚠️ {{ $t('faq.loadError', 'Failed to load FAQs') }}</div>
        <p class="text-white/60">{{ error.message }}</p>
      </div>

      <!-- FAQ Content -->
      <div v-else-if="filteredFaqData && filteredFaqData.categories && filteredFaqData.categories.length > 0" class="space-y-6">
        <!-- Category Loop -->
        <div v-for="category in filteredFaqData.categories" :key="category.id" class="space-y-3">
          <!-- Category Header -->
          <div class="flex items-center gap-3 mb-4">
            <span class="text-2xl">{{ category.icon }}</span>
            <h3 class="text-xl font-bold text-white">{{ category.name }}</h3>
            <span class="text-white/40 text-sm">({{ category.items?.length || 0 }})</span>
          </div>

          <!-- FAQ Items -->
          <div class="space-y-2">
            <div
              v-for="item in category.items"
              :key="item.id"
              class="bg-white/[0.03] border border-white/10 rounded-lg overflow-hidden hover:border-[#6b73ff]/50 transition-all"
            >
              <button
                @click="toggleItem(item.id)"
                class="w-full text-left p-4 flex items-start justify-between gap-4"
              >
                <div class="flex-1">
                  <h4 class="text-base font-semibold text-white pr-4">
                    {{ item.question }}
                  </h4>
                </div>
                <div class="flex-shrink-0">
                  <svg
                    class="w-5 h-5 text-[#6b73ff] transition-transform duration-200"
                    :class="{ 'rotate-180': openItems.includes(item.id) }"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                  </svg>
                </div>
              </button>
              
              <!-- Answer (Expandable) -->
              <transition
                enter-active-class="transition-all duration-200 ease-out"
                leave-active-class="transition-all duration-200 ease-in"
                enter-from-class="max-h-0 opacity-0"
                leave-to-class="max-h-0 opacity-0"
              >
                <div v-if="openItems.includes(item.id)" class="px-4 pb-4">
                  <div class="text-white/70 text-sm leading-relaxed prose prose-invert max-w-none" v-html="item.answer"></div>
                </div>
              </transition>
            </div>
          </div>
        </div>
      </div>

      <!-- No Results State (after filtering) -->
      <div v-else-if="faqData && faqData.categories && faqData.categories.length > 0" class="flex flex-col items-center justify-center py-20">
        <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-white/20 mb-4">
          <circle cx="11" cy="11" r="8"/>
          <path d="m21 21-4.35-4.35"/>
        </svg>
        <p class="text-white/40 text-center">{{ $t('faq.noResults', 'No FAQs match your search or filters') }}</p>
        <button @click="resetFilters" class="mt-4 px-4 py-2 bg-white/10 hover:bg-white/20 rounded-full text-white text-sm transition-colors">
          {{ $t('filter.reset', 'Reset Filters') }}
        </button>
      </div>
      
      <!-- Empty State (no data) -->
      <div v-else class="flex flex-col items-center justify-center py-20">
        <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-white/20 mb-4">
          <circle cx="12" cy="12" r="10"/>
          <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
          <line x1="12" y1="17" x2="12.01" y2="17"/>
        </svg>
        <p class="text-white/40">{{ $t('faq.noContent', 'No FAQ content available') }}</p>
      </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="px-6 py-4 border-t border-white/10 bg-white/[0.02]">
      <div class="flex items-center justify-between gap-3">
        <p class="text-sm text-white/40">
          {{ $t('faq.needHelp', 'Still need help?') }}
        </p>
        <div class="flex items-center gap-3">
          <button
            @click="openAllFaq"
            class="px-4 py-2 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black font-medium text-sm hover:shadow-[0_0_20px_rgba(107,115,255,0.5)] transition-all duration-200"
          >
            All F.A.Q
          </button>
          <button 
            @click="openWhatsApp"
            class="px-4 py-2 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-black font-medium text-sm hover:shadow-[0_0_20px_rgba(107,115,255,0.5)] transition-all duration-200"
          >
            {{ $t('common.contactUs', 'Contact Us') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import ContentSearchFilter from '~/components/ContentSearchFilter.vue'
import { useLocalePath } from '#imports'

const { locale } = useI18n()
const config = useRuntimeConfig()
const localePath = useLocalePath()

// Track open/closed items
const openItems = ref([])

// Search and filter state
const searchText = ref('')
const selectedCategories = ref([])

// Fetch FAQ data
const { data: faqData, pending, error } = await useFetch(
  () => `${config.public.wpApiBase?.replace('/wp-json', '') || 'https://tanzanite.site'}/wp-content/uploads/faq/${locale.value}.json`,
  {
    key: `faq-modal-${locale.value}`,
    // Watch locale changes
    watch: [locale],
    // Cache for 1 hour
    getCachedData: (key) => {
      if (typeof window === 'undefined') return
      const nuxtApp = useNuxtApp()
      const data = nuxtApp.payload.data[key] || nuxtApp.static.data[key]
      if (!data) return
      
      // Check if cached data is still fresh (1 hour)
      const expirationDate = new Date(data.fetchedAt)
      expirationDate.setHours(expirationDate.getHours() + 1)
      
      if (expirationDate.getTime() < Date.now()) return
      
      return data
    }
  }
)

// Fixed categories from backend settings (filter buttons only, no counts)
const faqCategories = [
  { id: 1, name: 'Product Questions' },
  { id: 2, name: 'Shipping & Delivery' },
  { id: 3, name: 'Returns & Refunds' },
  { id: 4, name: 'Payment Methods' }
]

// Filtered FAQ data based on search and selected categories
const filteredFaqData = computed(() => {
  if (!faqData.value?.categories) return null
  
  let categories = faqData.value.categories
  
  // 1. Filter by selected categories
  if (selectedCategories.value.length > 0) {
    categories = categories.filter(cat => 
      selectedCategories.value.includes(cat.id)
    )
  }
  
  // 2. Filter by search text
  if (searchText.value.trim()) {
    const searchLower = searchText.value.toLowerCase()
    categories = categories.map(cat => ({
      ...cat,
      items: cat.items?.filter(item => 
        item.question.toLowerCase().includes(searchLower) ||
        item.answer.toLowerCase().includes(searchLower)
      )
    })).filter(cat => cat.items && cat.items.length > 0)
  }
  
  return { categories }
})

// Handle search
const handleSearch = (text) => {
  searchText.value = text
}

// Handle filters update
const handleFiltersUpdate = (filters) => {
  searchText.value = filters.searchText || ''
  selectedCategories.value = filters.selectedCategories || []
}

// Reset all filters
const resetFilters = () => {
  searchText.value = ''
  selectedCategories.value = []
}

// Toggle FAQ item
const toggleItem = (id) => {
  const index = openItems.value.indexOf(id)
  if (index > -1) {
    openItems.value.splice(index, 1)
  } else {
    openItems.value.push(id)
  }
}

// Open WhatsApp Chat Modal
const emit = defineEmits(['close', 'openWhatsApp'])

const openAllFaq = () => {
  try {
    const target = localePath('/faq')
    if (typeof window !== 'undefined' && target) {
      window.open(String(target), '_blank')
    }
  } catch (error) {
    console.error('Failed to open FAQ page:', error)
  }
}

const openWhatsApp = () => {
  // Close FAQ modal
  emit('close')
  // Trigger WhatsApp modal
  emit('openWhatsApp')
}
</script>

<style scoped>
/* Custom scrollbar */
.overflow-y-auto::-webkit-scrollbar {
  width: 8px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: rgba(107, 115, 255, 0.5);
  border-radius: 4px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: rgba(107, 115, 255, 0.7);
}
</style>
