<template>
  <div class="min-h-screen bg-black text-white py-20 px-4">
    <div class="max-w-4xl mx-auto">
      <!-- Header -->
      <div class="text-center mb-12">
        <h1 class="text-4xl md:text-5xl font-bold mb-4 bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] bg-clip-text text-transparent">
          {{ $t('faq.title', 'Frequently Asked Questions') }}
        </h1>
        <p class="text-white/70 text-lg">
          {{ $t('faq.subtitle', 'Find answers to common questions about our products and services') }}
        </p>
      </div>

      <!-- Loading State -->
      <div v-if="pending" class="text-center py-20">
        <div class="inline-block animate-spin rounded-full h-12 w-12 border-4 border-[#6b73ff] border-t-transparent"></div>
        <p class="mt-4 text-white/60">Loading FAQs...</p>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="text-center py-20">
        <div class="text-red-500 text-xl mb-4">⚠️ Failed to load FAQs</div>
        <p class="text-white/60">{{ error.message }}</p>
      </div>

      <!-- FAQ Content -->
      <div v-else-if="faqData" class="space-y-8">
        <!-- Category Loop -->
        <div v-for="category in faqData.categories" :key="category.id" class="space-y-4">
          <!-- Category Header -->
          <div class="flex items-center gap-3 mb-6">
            <span class="text-3xl">{{ category.icon }}</span>
            <h2 class="text-2xl font-bold text-white">{{ category.name }}</h2>
            <span class="text-white/40 text-sm">({{ category.items.length }})</span>
          </div>

          <!-- FAQ Items -->
          <div class="space-y-3">
            <div
              v-for="item in category.items"
              :key="item.id"
              class="bg-white/[0.06] border border-white/10 rounded-xl overflow-hidden hover:border-[#6b73ff] transition-all"
            >
              <button
                @click="toggleItem(item.id)"
                class="w-full text-left p-6 flex items-start justify-between gap-4"
              >
                <div class="flex-1">
                  <h3 class="text-lg font-semibold text-white pr-4">
                    {{ item.question }}
                  </h3>
                </div>
                <div class="flex-shrink-0">
                  <svg
                    class="w-6 h-6 text-[#6b73ff] transition-transform"
                    :class="{ 'rotate-180': openItems.includes(item.id) }"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                  </svg>
                </div>
              </button>
              
              <!-- Answer (Collapsible) -->
              <div
                v-show="openItems.includes(item.id)"
                class="px-6 pb-6 text-white/80 leading-relaxed"
                v-html="item.answer"
              ></div>
            </div>
          </div>
        </div>

        <!-- No FAQs -->
        <div v-if="faqData.categories.length === 0" class="text-center py-20">
          <div class="text-white/40 text-xl">No FAQs available yet</div>
        </div>
      </div>

      <!-- Back to Home -->
      <div class="mt-12 text-center">
        <NuxtLink
          to="/"
          class="inline-flex items-center gap-2 px-6 py-3 bg-[#6b73ff] text-white rounded-full hover:brightness-110 transition-all"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          {{ $t('common.backToHome', 'Back to Home') }}
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup>
const { locale } = useI18n()
const config = useRuntimeConfig()

// Track open/closed items
const openItems = ref([])

// Fetch FAQ data
const { data: faqData, pending, error } = await useFetch(
  () => `${config.public.wordpressUrl}/wp-content/uploads/faq/${locale.value}.json`,
  {
    key: `faq-${locale.value}`,
    // Cache for 1 hour
    getCachedData: (key) => {
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

// Toggle FAQ item
const toggleItem = (id) => {
  const index = openItems.value.indexOf(id)
  if (index > -1) {
    openItems.value.splice(index, 1)
  } else {
    openItems.value.push(id)
  }
}

// SEO Meta Tags
useHead({
  title: () => `FAQ - ${locale.value.toUpperCase()}`,
  meta: [
    {
      name: 'description',
      content: 'Find answers to frequently asked questions about our products and services.'
    }
  ]
})

// FAQ Schema.org structured data
useHead({
  script: [
    {
      type: 'application/ld+json',
      children: () => {
        if (!faqData.value || !faqData.value.categories) return '{}'
        
        const allItems = faqData.value.categories.flatMap(cat => cat.items)
        
        return JSON.stringify({
          '@context': 'https://schema.org',
          '@type': 'FAQPage',
          'mainEntity': allItems.map(item => ({
            '@type': 'Question',
            'name': item.question,
            'acceptedAnswer': {
              '@type': 'Answer',
              'text': item.answer.replace(/<[^>]*>/g, '') // Strip HTML tags
            }
          }))
        })
      }
    }
  ]
})
</script>

<style scoped>
/* Answer content styling */
:deep(.px-6.pb-6) p {
  margin-bottom: 1em;
}

:deep(.px-6.pb-6) ul,
:deep(.px-6.pb-6) ol {
  margin-left: 1.5em;
  margin-bottom: 1em;
}

:deep(.px-6.pb-6) li {
  margin-bottom: 0.5em;
}

:deep(.px-6.pb-6) strong {
  color: #fff;
  font-weight: 600;
}

:deep(.px-6.pb-6) a {
  color: #6b73ff;
  text-decoration: underline;
}

:deep(.px-6.pb-6) a:hover {
  color: #40ffaa;
}
</style>
