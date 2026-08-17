<template>
  <section class="bg-transparent text-white pt-6 pb-2">
    <div class="page-content-shell px-0 md:px-4">
      
      <!-- Tab Navigation -->
      <div class="flex justify-center mb-4 sm:mb-8">
        <div class="tz-segmented-tabs" role="tablist">
          <button
            v-for="(tab, index) in tabs"
            :key="tab.id"
            type="button"
            @click="selectTab(tab.id)"
            @keydown="handleTabKeydown($event, index)"
            class="tz-segmented-tabs__button"
            :class="{ 'is-active': currentTabId === tab.id }"
            role="tab"
            :aria-selected="currentTabId === tab.id"
            :tabindex="currentTabId === tab.id ? 0 : -1"
          >
            {{ $t(tab.labelKey) }}
          </button>
        </div>
      </div>

      <!-- Main Content Area -->
      <div class="grid lg:grid-cols-12 gap-3 sm:gap-6 items-start">
        
        <!-- Left Column: Interactive List (Summary) -->
        <div class="lg:col-span-4 lg:pr-4 order-1 lg:self-center">
          <ul class="space-y-1 sm:space-y-2">
            <li
              v-for="(item, index) in currentTabItems"
              :key="index"
              class="tz-selectable-list-item p-2 sm:p-3"
              :class="{ 'is-active': activeIndex === index }"
              @click="activeIndex = index"
              @mouseenter="activeIndex = index"
            >
              <div class="flex items-center gap-3">
                <!-- Active Indicator Dot -->
                <div 
                  class="tz-selectable-list-item__indicator"
                ></div>
                
                <span 
                  class="tz-selectable-list-item__label text-sm font-medium"
                >
                  {{ $t(item.titleKey) }}
                </span>
              </div>
            </li>
          </ul>
        </div>

        <!-- Right Column: Visual Carousel -->
        <div class="lg:col-span-8 order-2">
          <StackedCarousel 
            :items="currentTabItems" 
            v-model="activeIndex"
          >
            <template #card="{ item }">
              <div class="h-full w-full rounded-2xl premium-card p-6 flex flex-col justify-center items-center text-center border border-white/10 bg-[var(--tz-card-surface)] relative overflow-hidden">
                
                <!-- Background decor for visual interest (since no distinct images) -->
                <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(181,255,109,0.05),transparent_60%)] pointer-events-none"></div>

                <!-- Card Content -->
                <h3 class="text-xl font-bold text-white mb-3 relative z-10">{{ $t(item.titleKey) }}</h3>
                <p class="tz-text-secondary text-sm leading-relaxed max-w-lg mx-auto mb-6 relative z-10">
                  {{ $t(item.descriptionKey) }}
                </p>

                <!-- Innovation Bullets -->
                <ul v-if="currentTabId === 'innovation' && item.bullets" class="space-y-2 text-sm text-left inline-block mx-auto mb-6 relative z-10">
                   <li v-for="bulletKey in item.bullets" :key="bulletKey" class="flex gap-3">
                      <span class="text-teal-400 mt-1">•</span>
                      <span class="tz-text-secondary">{{ $t(bulletKey) }}</span>
                   </li>
                </ul>

                <!-- Action Buttons -->
                <div class="relative z-10">
                  <NuxtLink 
                    v-if="currentTabId === 'stories'" 
                    to="/blog" 
                    class="premium-button text-sm px-6 py-2 inline-flex items-center"
                  >
                    {{ $t('home.factoryStories.readMore') }}
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" class="ml-2 h-4 w-4" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 6l6 6-6 6" /></svg>
                  </NuxtLink>

                  <button 
                    v-else 
                    type="button" 
                    class="premium-button text-sm px-6 py-2 inline-flex items-center"
                  >
                    {{ $t('home.innovationRd.readMore') }}
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" class="ml-2 h-4 w-4" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 6l6 6-6 6" /></svg>
                  </button>
                </div>

              </div>
            </template>
          </StackedCarousel>
        </div>

      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch } from 'vue'
import { useI18n } from '#imports'
import StackedCarousel from '~/components/StackedImageCarousel.vue'

const { t } = useI18n()

type TabId = 'stories' | 'innovation'

const tabs: { id: TabId; labelKey: string }[] = [
  { id: 'stories', labelKey: 'home.factoryStories.title' },
  { id: 'innovation', labelKey: 'home.innovationRd.title' }
]

const currentTabId = ref<TabId>('stories')
const activeIndex = ref(0) // Syncs with carousel

const selectTab = (id: TabId) => {
  currentTabId.value = id
}

const handleTabKeydown = (event: KeyboardEvent, index: number) => {
  const tabCount = tabs.length
  let nextIndex = index

  switch (event.key) {
    case 'ArrowRight':
    case 'ArrowDown':
      nextIndex = (index + 1) % tabCount
      break
    case 'ArrowLeft':
    case 'ArrowUp':
      nextIndex = (index - 1 + tabCount) % tabCount
      break
    case 'Home':
      nextIndex = 0
      break
    case 'End':
      nextIndex = tabCount - 1
      break
    default:
      return
  }

  event.preventDefault()
  const nextTab = tabs[nextIndex]
  if (!nextTab) return

  selectTab(nextTab.id)
  const tabsContainer = event.currentTarget instanceof HTMLElement
    ? event.currentTarget.parentElement
    : null
  nextTick(() => {
    tabsContainer
      ?.querySelectorAll<HTMLElement>('[role="tab"]')
      .item(nextIndex)
      ?.focus()
  })
}

// Reset index when tab changes
watch(currentTabId, () => {
  activeIndex.value = 0
})

// --- Data Definition ---

// Factory Stories Data
const storiesCards = computed(() => {
  return [
    {
      titleKey: 'home.factoryStories.items.0.title',
      descriptionKey: 'home.factoryStories.items.0.description',
    },
    {
      titleKey: 'home.factoryStories.items.1.title',
      descriptionKey: 'home.factoryStories.items.1.description',
    },
    {
      titleKey: 'home.factoryStories.items.2.title',
      descriptionKey: 'home.factoryStories.items.2.description',
    },
    {
      titleKey: 'home.factoryStories.items.3.title',
      descriptionKey: 'home.factoryStories.items.3.description',
    },
  ]
})

// Innovation & R&D Data
const innovationCards = computed(() => {
  return Array.from({ length: 3 }, (_, index) => ({
    titleKey: `home.innovationRd.items.${index}.title`,
    descriptionKey: `home.innovationRd.items.${index}.description`,
    bullets: [0, 1, 2].map((bulletIndex) => `home.innovationRd.items.${index}.bullets.${bulletIndex}`),
  }))
})

const currentTabItems = computed(() => {
  return currentTabId.value === 'stories' ? storiesCards.value : innovationCards.value
})

</script>
