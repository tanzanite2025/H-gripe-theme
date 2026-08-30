<template>
  <section class="global-url-search-results">
    <div class="global-url-search-results__section-heading">
      <div>
        <span class="global-url-search-results__section-kicker">
          {{ sectionKicker }}
        </span>
        <h3>{{ sectionTitle }}</h3>
      </div>
      <span class="global-url-search-results__result-count">
        {{ searchResultCount }}
      </span>
    </div>

    <div v-if="items.length > 0" class="global-url-search-results__list">
      <NuxtLink
        v-for="item in items"
        :key="item.id"
        :to="resolveTarget(item)"
        class="global-url-search-results__item"
        @click="emit('select', item)"
      >
        <div class="global-url-search-results__item-head">
          <span class="global-url-search-results__title">
            {{ titleFor(item) }}
          </span>
          <span class="global-url-search-results__path font-mono">
            {{ pathFor(item) }}
          </span>
        </div>

        <p v-if="summaryFor(item)" class="global-url-search-results__summary">
          {{ summaryFor(item) }}
        </p>

        <div v-if="keywordsFor(item).length" class="global-url-search-results__keywords">
          <span
            v-for="keyword in keywordsFor(item)"
            :key="keyword"
            class="global-url-search-results__keyword"
          >
            {{ keyword }}
          </span>
        </div>
      </NuxtLink>
    </div>

    <div v-else class="global-url-search-results__empty">
      {{ emptyLabel }}
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useLocalePath } from '#imports'
import type { StorefrontURLSearchProfile } from '~/data/url-search/types'

const props = defineProps<{
  items: StorefrontURLSearchProfile[]
  searchResultCount: number
  sectionKicker?: string
  sectionTitle?: string
  emptyLabel?: string
}>()

const emit = defineEmits<{
  select: [item: StorefrontURLSearchProfile]
}>()

const localePath = useLocalePath()
const sectionKicker = computed(() => props.sectionKicker || 'URL')
const sectionTitle = computed(() => props.sectionTitle || 'URL 搜索结果')
const emptyLabel = computed(() => props.emptyLabel || '没有找到匹配的 URL')

const routeEntryPath = (item: StorefrontURLSearchProfile): string => {
  return String(item.route_entry?.path || '').trim() || '/'
}

const resolveTarget = (item: StorefrontURLSearchProfile) => localePath(routeEntryPath(item))

const titleFor = (item: StorefrontURLSearchProfile): string => {
  return item.display_title || item.route_entry?.title || routeEntryPath(item)
}

const summaryFor = (item: StorefrontURLSearchProfile): string => {
  return item.display_summary || item.route_entry?.summary || ''
}

const keywordsFor = (item: StorefrontURLSearchProfile): string[] => {
  return Array.isArray(item.keywords) ? item.keywords : []
}

const pathFor = (item: StorefrontURLSearchProfile): string => {
  return item.route_entry?.path || routeEntryPath(item)
}
</script>

<style scoped>
.global-url-search-results {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.global-url-search-results__section-heading {
  display: flex;
  min-height: 2rem;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding-bottom: 0.35rem;
}

.global-url-search-results__section-kicker {
  display: block;
  margin-bottom: 0.18rem;
  color: var(--tz-text-muted);
  font-size: 0.55rem;
  font-weight: 900;
  letter-spacing: 0.13em;
  line-height: 1.2;
  text-transform: uppercase;
}

.global-url-search-results__section-heading h3 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 900;
  line-height: 1.25;
}

.global-url-search-results__result-count {
  display: inline-grid;
  min-width: 1.7rem;
  height: 1.7rem;
  place-items: center;
  border: 1px solid rgba(20, 32, 43, 0.16);
  border-radius: 999px;
  color: var(--tz-text-accent);
  font-size: 0.62rem;
  font-weight: 900;
}

.global-url-search-results__list {
  display: grid;
  gap: 0.55rem;
}

.global-url-search-results__item {
  display: grid;
  gap: 0.45rem;
  padding: 0.75rem 0.8rem;
  border: 1px solid rgba(20, 32, 43, 0.12);
  border-radius: 0.72rem;
  background: var(--tz-card-surface, #ffffff);
  color: var(--tz-text-primary);
  text-decoration: none;
  transition: border-color 0.2s ease, background 0.2s ease;
}

.global-url-search-results__item:hover,
.global-url-search-results__item:focus-visible {
  border-color: rgba(5, 150, 105, 0.45);
  background:
    linear-gradient(0deg, rgba(5, 150, 105, 0.06), rgba(5, 150, 105, 0.06)),
    var(--tz-card-surface, #ffffff);
}

.global-url-search-results__item-head {
  display: flex;
  min-width: 0;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.65rem;
}

.global-url-search-results__title {
  min-width: 0;
  overflow: hidden;
  font-size: 0.76rem;
  font-weight: 900;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.global-url-search-results__path {
  flex: 0 0 auto;
  color: var(--tz-text-muted);
  font-size: 0.58rem;
  letter-spacing: 0.05em;
}

.global-url-search-results__summary {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  line-height: 1.55;
}

.global-url-search-results__keywords {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.global-url-search-results__keyword {
  display: inline-flex;
  align-items: center;
  padding: 0.22rem 0.45rem;
  border-radius: 999px;
  background: rgba(5, 150, 105, 0.08);
  color: var(--tz-text-accent);
  font-size: 0.56rem;
  font-weight: 800;
}

.global-url-search-results__empty {
  display: grid;
  min-height: 4rem;
  place-items: center;
  border: 1px dashed rgba(20, 32, 43, 0.16);
  border-radius: 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
}
</style>
