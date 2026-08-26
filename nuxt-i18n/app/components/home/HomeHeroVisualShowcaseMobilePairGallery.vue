<template>
  <div class="home-hero-visual-showcase-mobile">
    <div class="tz-carousel-pagination home-hero-visual-showcase-mobile__pagination" role="tablist" :aria-label="ariaLabel">
      <button
        v-for="pairNumber in mobilePairCount"
        :key="pairNumber"
        type="button"
        class="tz-carousel-pagination__dot"
        :class="{ 'is-active': pairNumber - 1 === activePairIndex }"
        :aria-label="`${ariaLabel} ${pairNumber}`"
        :aria-selected="pairNumber - 1 === activePairIndex"
        role="tab"
        @click="setActivePair(pairNumber - 1)"
      ></button>
    </div>

    <div class="home-hero-visual-showcase-mobile__grid" role="tabpanel" aria-live="polite">
      <button
        v-for="({ item, index }) in activePair"
        :key="`${activePairIndex}-${item.id}`"
        type="button"
        class="home-hero-visual-showcase-mobile__card"
        :class="{ 'is-active': index === activeItemIndex }"
        :aria-label="`${item.title} ${index + 1}`"
        :aria-pressed="index === activeItemIndex"
        @click="setActiveItem(index)"
      >
        <HomeHeroVisualShowcaseFigure
          :item="item"
          sizes="xs:50vw sm:50vw md:50vw"
          :loading="activePairIndex === 0 ? 'eager' : 'lazy'"
          :fetchpriority="activePairIndex === 0 && index === 0 ? 'high' : 'low'"
          caption-visibility="sr-only"
        />
      </button>
    </div>

    <div
      v-if="activeItem"
      class="home-hero-visual-showcase-mobile__detail"
      role="region"
      aria-live="polite"
      :aria-label="activeItem.title"
    >
      <div class="home-hero-visual-showcase-mobile__detail-copy">
        <h3 class="home-hero-visual-showcase-mobile__detail-title">{{ activeItem.title }}</h3>
        <p v-if="activeItem.caption" class="home-hero-visual-showcase-mobile__detail-description">
          {{ activeItem.caption }}
        </p>
      </div>
      <NuxtLink
        v-if="activeItem.targetUrl && activeItem.targetLabel"
        :to="activeItem.targetUrl"
        class="home-hero-visual-showcase-mobile__detail-link"
      >
        {{ activeItem.targetLabel }}
        <Icon name="lucide:arrow-up-right" aria-hidden="true" />
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import HomeHeroVisualShowcaseFigure from '~/components/home/HomeHeroVisualShowcaseFigure.vue'
import type { HomeHeroVisualShowcaseItem } from '~/types/homeHeroVisualShowcase'

const props = defineProps<{
  items: HomeHeroVisualShowcaseItem[]
  ariaLabel: string
}>()

const mobileItems = computed(() => props.items.slice(0, 8))
const activePairIndex = ref(0)
const activeItemIndex = ref(0)
const mobilePairCount = computed(() => Math.ceil(mobileItems.value.length / 2))
const activePair = computed(() => (
  mobileItems.value
    .slice(activePairIndex.value * 2, activePairIndex.value * 2 + 2)
    .map((item, pairOffset) => ({
      item,
      index: activePairIndex.value * 2 + pairOffset,
    }))
))
const activeItem = computed(() => mobileItems.value[activeItemIndex.value] ?? null)

const setActivePair = (pairIndex: number) => {
  if (pairIndex < 0 || pairIndex >= mobilePairCount.value) return
  activePairIndex.value = pairIndex
  activeItemIndex.value = Math.min(pairIndex * 2, Math.max(0, mobileItems.value.length - 1))
}

const setActiveItem = (index: number) => {
  if (index < 0 || index >= mobileItems.value.length) return
  activeItemIndex.value = index
}
</script>

<style scoped>
.home-hero-visual-showcase-mobile__pagination {
  position: static;
  margin: 0 0 0.35rem;
  padding: 0;
}

.home-hero-visual-showcase-mobile__grid {
  display: grid;
  width: 100%;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: clamp(0.45rem, 0.8vw, 0.75rem);
}

.home-hero-visual-showcase-mobile__card {
  display: block;
  min-width: 0;
  padding: 0;
  border: 0;
  border-radius: 0.75rem;
  background: transparent;
  cursor: pointer;
}

.home-hero-visual-showcase-mobile__card.is-active {
  box-shadow: 0 0 0 2px rgba(5, 150, 105, 0.9);
}

.home-hero-visual-showcase-mobile__card:focus-visible {
  outline: 2px solid rgba(5, 150, 105, 0.95);
  outline-offset: 3px;
}

.home-hero-visual-showcase-mobile__detail {
  display: flex;
  min-height: 4.8rem;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: 0.55rem;
  padding: 0.75rem 0.85rem;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 0.85rem;
  background: rgba(255, 255, 255, 0.96);
  color: #111318;
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.24);
  text-align: left;
}

.home-hero-visual-showcase-mobile__detail-copy {
  min-width: 0;
}

.home-hero-visual-showcase-mobile__detail-title {
  margin: 0;
  font-size: 0.82rem;
  font-weight: 800;
  line-height: 1.2;
}

.home-hero-visual-showcase-mobile__detail-description {
  margin: 0.3rem 0 0;
  color: rgba(17, 19, 24, 0.7);
  font-size: 0.68rem;
  line-height: 1.4;
}

.home-hero-visual-showcase-mobile__detail-link {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.25rem;
  color: #111318;
  font-size: 0.68rem;
  font-weight: 800;
  text-decoration: none;
  white-space: nowrap;
}

.home-hero-visual-showcase-mobile__detail-link :deep(svg) {
  width: 0.8rem;
  height: 0.8rem;
}
</style>
