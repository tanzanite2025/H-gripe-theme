<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200 ease-out"
      leave-active-class="transition-opacity duration-300 ease-in"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <section
        v-if="open"
        class="global-all-faqs-search-overlay"
        role="dialog"
        aria-modal="true"
        :aria-label="t('faq.title')"
        @click.self="emit('close')"
      >
        <div
          class="global-all-faqs-search-overlay__backdrop"
          aria-hidden="true"
          @click="emit('close')"
        />
        <div class="global-all-faqs-search-overlay__viewport">
          <GlobalAllFaqsDesktopSearchOverlayPanel
            :pending="pending"
            :search-query="searchQuery"
            :featured-topics="featuredTopics"
            :active-topic-id="activeTopicId"
            :active-topic="activeTopic"
            :topic-items="topicItems"
            :search-results="searchResults"
            :search-result-count="searchResultCount"
            :expanded-items="expandedItems"
            @update:search-query="searchQuery = $event"
            @select-topic="selectTopic"
            @toggle-item="toggleItem"
            @close="emit('close')"
          />
          <GlobalAllFaqsMobileSearchOverlayPanel
            :pending="pending"
            :search-query="searchQuery"
            :featured-topics="featuredTopics"
            :active-topic-id="activeTopicId"
            :active-topic="activeTopic"
            :topic-items="topicItems"
            :search-results="searchResults"
            :search-result-count="searchResultCount"
            :expanded-items="expandedItems"
            @update:search-query="searchQuery = $event"
            @select-topic="selectTopic"
            @toggle-item="toggleItem"
            @close="emit('close')"
          />
        </div>
      </section>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import { watch } from 'vue'
import GlobalAllFaqsDesktopSearchOverlayPanel from '~/components/faq/GlobalAllFaqsDesktopSearchOverlayPanel.vue'
import GlobalAllFaqsMobileSearchOverlayPanel from '~/components/faq/GlobalAllFaqsMobileSearchOverlayPanel.vue'
import { useGlobalAllFaqsSearchAndGroupedResults } from '~/composables/useGlobalAllFaqsSearchAndGroupedResults'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const {
  pending,
  refreshAllFaqData,
  featuredItems,
  featuredTopics,
  activeTopicId,
  activeTopic,
  topicItems,
  searchQuery,
  searchResults,
  searchResultCount,
  expandedItems,
  toggleItem,
  selectTopic,
  resetSearchOverlayState,
} = await useGlobalAllFaqsSearchAndGroupedResults({
  mode: 'search-overlay',
})

watch(
  () => props.open,
  (open, previousOpen) => {
    if (open && previousOpen === false) {
      resetSearchOverlayState()
    }
    if (!open || previousOpen !== false || pending.value || featuredItems.value.length > 0) return
    void refreshAllFaqData()
  },
  { flush: 'post' },
)
</script>

<style scoped>
.global-all-faqs-search-overlay {
  position: fixed;
  inset: 0;
  z-index: 10004;
  overflow: hidden;
}

.global-all-faqs-search-overlay__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(20, 32, 43, 0.46);
  backdrop-filter: blur(5px);
}

.global-all-faqs-search-overlay__viewport {
  position: relative;
  z-index: 1;
  display: flex;
  width: 100%;
  height: 100%;
  justify-content: center;
  align-items: center;
  padding: 1rem;
  overflow: hidden;
  animation: global-all-faqs-search-overlay-slide-up 520ms
    cubic-bezier(0.22, 1, 0.36, 1) both;
}

@media (max-width: 767px) {
  .global-all-faqs-search-overlay__viewport {
    align-items: stretch;
    padding: 0.5rem 1px 0.5rem;
  }
}

@keyframes global-all-faqs-search-overlay-slide-up {
  from {
    opacity: 0;
    transform: translateY(2.25rem);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .global-all-faqs-search-overlay__viewport {
    animation: none;
  }
}
</style>
