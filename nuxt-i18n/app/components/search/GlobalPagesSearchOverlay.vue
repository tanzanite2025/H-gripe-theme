<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200 ease-out"
      leave-active-class="transition-opacity duration-150 ease-in"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <section
        v-if="isPagesSearchOpen"
        class="global-pages-search-overlay"
        role="dialog"
        aria-modal="true"
        :aria-label="title"
        @click.self="closePagesSearch('user')"
      >
        <div
          class="global-pages-search-overlay__backdrop"
          aria-hidden="true"
          @click="closePagesSearch('user')"
        />

        <div class="global-pages-search-overlay__viewport">
          <section class="global-pages-search-overlay__panel">
            <header class="global-pages-search-overlay__header">
              <div class="min-w-0">
                <div class="global-pages-search-overlay__eyebrow">
                  URL
                </div>
                <h2 class="global-pages-search-overlay__title">
                  {{ title }}
                </h2>
              </div>

              <button
                type="button"
                class="global-pages-search-overlay__close"
                :aria-label="t('common.close', 'Close')"
                @click="closePagesSearch('user')"
              >
                <Icon name="lucide:x" aria-hidden="true" />
              </button>
            </header>

            <div class="global-pages-search-overlay__search">
              <Icon name="lucide:search" aria-hidden="true" />
              <input
                ref="searchInputRef"
                v-model="pagesSearchQuery"
                type="search"
                :placeholder="searchPlaceholder"
                :aria-label="searchPlaceholder"
              >
              <button
                v-if="pagesSearchQuery"
                type="button"
                class="global-pages-search-overlay__clear"
                :aria-label="t('filter.clearSearch', 'Clear search')"
                @click="pagesSearchQuery = ''"
              >
                <Icon name="lucide:x" aria-hidden="true" />
              </button>
            </div>

            <div class="global-pages-search-overlay__body">
              <div
                v-if="pending"
                class="global-pages-search-overlay__state"
              >
                Loading URL catalog...
              </div>

              <div
                v-else-if="error"
                class="global-pages-search-overlay__state global-pages-search-overlay__state--error"
              >
                Failed to load page search data.
              </div>

              <GlobalUrlSearchResults
                v-else
                :items="displayedProfiles"
                :search-result-count="displayedCount"
                :section-kicker="sectionKicker"
                :section-title="sectionTitle"
                :empty-label="emptyLabel"
                @select="handleResultSelect"
              />
            </div>
          </section>
        </div>
      </section>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from '#imports'
import GlobalUrlSearchResults from '~/components/search/GlobalUrlSearchResults.vue'
import { usePagesSearchOverlayState } from '~/composables/usePagesSearchOverlayState'
import { useUrlSearch } from '~/composables/useUrlSearch'
import { useUrlSearchCatalog } from '~/composables/useUrlSearchCatalog'
import type { StorefrontURLSearchProfile } from '~/data/url-search/types'

const { t } = useI18n()
const { isPagesSearchOpen, pagesSearchQuery, closePagesSearch } = usePagesSearchOverlayState()
const { urlSearchProfiles, pending, error } = await useUrlSearchCatalog()
const { searchResults, searchResultCount } = useUrlSearch(urlSearchProfiles, pagesSearchQuery)

const searchInputRef = ref<HTMLInputElement | null>(null)
const normalizedQuery = computed(() => pagesSearchQuery.value.trim())
const hasSearchQuery = computed(() => normalizedQuery.value.length > 0)

const normalizeText = (value: string) => value.toLowerCase().replace(/\s+/g, ' ').trim()

const activeUrlProfiles = computed(() => (
  [...urlSearchProfiles.value]
    .filter(profile => profile.enabled !== false && Boolean(profile.route_entry?.path))
    .sort((left, right) => {
      const leftWeight = Number(left.search_weight || 0)
      const rightWeight = Number(right.search_weight || 0)
      if (rightWeight !== leftWeight) return rightWeight - leftWeight

      const leftPath = left.route_entry?.path || ''
      const rightPath = right.route_entry?.path || ''
      return normalizeText(leftPath).localeCompare(normalizeText(rightPath))
    })
))

const featuredProfiles = computed(() => activeUrlProfiles.value.slice(0, 6))
const displayedProfiles = computed(() => (
  hasSearchQuery.value ? searchResults.value : featuredProfiles.value
))
const displayedCount = computed(() => (
  hasSearchQuery.value ? searchResultCount.value : displayedProfiles.value.length
))
const title = computed(() => t('header.globalNavigationTransition.options.pages', 'Pages'))
const sectionKicker = 'URL'
const sectionTitle = computed(() => (
  hasSearchQuery.value
    ? t('header.globalNavigationTransition.options.pages', 'Pages')
    : t('search.popularTitle', 'Popular searches')
))
const searchPlaceholder = computed(() => t('search.placeholder', 'Search...'))
const emptyLabel = computed(() => (
  hasSearchQuery.value
    ? 'No matching pages.'
    : 'No pages have been indexed yet.'
))

const handleResultSelect = (_item: StorefrontURLSearchProfile) => {
  closePagesSearch('navigate')
}

watch(
  () => isPagesSearchOpen.value,
  (open, previousOpen) => {
    if (!open || previousOpen) return

    pagesSearchQuery.value = ''
    nextTick(() => {
      searchInputRef.value?.focus()
    })
  },
  { flush: 'post' },
)

onMounted(() => {
  nextTick(() => {
    searchInputRef.value?.focus()
  })
})

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && isPagesSearchOpen.value) {
    closePagesSearch('user')
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.global-pages-search-overlay {
  position: fixed;
  inset: 0;
  z-index: 10005;
  overflow: hidden;
}

.global-pages-search-overlay__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(20, 32, 43, 0.46);
  backdrop-filter: blur(5px);
}

.global-pages-search-overlay__viewport {
  position: relative;
  z-index: 1;
  display: flex;
  width: 100%;
  height: 100%;
  justify-content: center;
  align-items: center;
  padding: 1rem;
  overflow: hidden;
}

.global-pages-search-overlay__panel {
  display: flex;
  flex-direction: column;
  width: min(56rem, calc(100vw - 1.5rem));
  max-height: min(84vh, calc(100dvh - 1.5rem));
  overflow: hidden;
  border-radius: 0.9rem;
  background: var(--tz-card-surface, #ffffff);
  box-shadow: 0 24px 70px rgba(20, 32, 43, 0.18);
}

.global-pages-search-overlay__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.2rem 0.85rem;
  border-bottom: 1px solid rgba(20, 32, 43, 0.12);
  background: var(--tz-card-surface);
}

.global-pages-search-overlay__eyebrow {
  margin-bottom: 0.3rem;
  color: var(--tz-text-accent);
  font-size: 0.6rem;
  font-weight: 900;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.global-pages-search-overlay__title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1.15rem;
  font-weight: 900;
  line-height: 1.2;
}

.global-pages-search-overlay__close {
  display: inline-grid;
  width: 2.2rem;
  height: 2.2rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid rgba(20, 32, 43, 0.14);
  border-radius: 0.6rem;
  background: #f3f6f8;
  color: var(--tz-text-secondary);
  cursor: pointer;
}

.global-pages-search-overlay__close:hover,
.global-pages-search-overlay__close:focus-visible {
  border-color: rgba(20, 32, 43, 0.3);
  color: var(--tz-text-primary);
}

.global-pages-search-overlay__close :deep(svg) {
  width: 1rem;
  height: 1rem;
}

.global-pages-search-overlay__search {
  display: flex;
  min-height: 2.75rem;
  align-items: center;
  gap: 0.6rem;
  margin: 0.8rem 1.2rem 0.75rem;
  padding: 0 0.8rem;
  border: 1px solid rgba(20, 32, 43, 0.18);
  border-radius: 0.7rem;
  background: var(--tz-input-surface, #ffffff);
  color: var(--tz-text-secondary);
}

.global-pages-search-overlay__search > :deep(svg) {
  width: 0.95rem;
  height: 0.95rem;
  flex: 0 0 auto;
}

.global-pages-search-overlay__search input {
  min-width: 0;
  flex: 1;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--tz-text-primary);
  font-size: 0.8rem;
}

.global-pages-search-overlay__search input::placeholder {
  color: var(--tz-text-muted);
}

.global-pages-search-overlay__clear {
  display: inline-grid;
  width: 1.55rem;
  height: 1.55rem;
  flex: 0 0 auto;
  place-items: center;
  border: 0;
  border-radius: 999px;
  background: #e4eaee;
  color: var(--tz-text-secondary);
  cursor: pointer;
}

.global-pages-search-overlay__clear :deep(svg) {
  width: 0.75rem;
  height: 0.75rem;
}

.global-pages-search-overlay__body {
  min-height: 0;
  flex: 1 1 auto;
  overflow: hidden;
  padding: 0 1.2rem 1.2rem;
}

.global-pages-search-overlay__state {
  display: grid;
  min-height: 12rem;
  place-items: center;
  border: 1px dashed rgba(20, 32, 43, 0.18);
  border-radius: 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.5;
  text-align: center;
}

.global-pages-search-overlay__state--error {
  color: #b42318;
}

@media (max-width: 767px) {
  .global-pages-search-overlay__viewport {
    align-items: stretch;
    padding: 0.5rem 1px 0.5rem;
  }

  .global-pages-search-overlay__panel {
    width: 100%;
    max-height: calc(100vh - 1rem);
    border-radius: 0.8rem;
  }

  .global-pages-search-overlay__header {
    padding: 0.8rem 0.75rem 0.65rem;
  }

  .global-pages-search-overlay__title {
    font-size: 0.96rem;
  }

  .global-pages-search-overlay__close {
    width: 2rem;
    height: 2rem;
    border-radius: 0.55rem;
  }

  .global-pages-search-overlay__search {
    min-height: 2.45rem;
    gap: 0.48rem;
    margin: 0.6rem 0.65rem 0.55rem;
    padding: 0 0.65rem;
    border-radius: 0.65rem;
  }

  .global-pages-search-overlay__search input {
    font-size: 0.76rem;
  }

  .global-pages-search-overlay__body {
    padding: 0 0.65rem 0.75rem;
  }
}
</style>
