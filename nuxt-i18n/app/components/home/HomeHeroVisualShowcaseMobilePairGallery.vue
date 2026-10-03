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

    <div
      class="home-hero-visual-showcase-mobile__grid"
      :class="{ 'home-hero-visual-showcase-mobile__grid--single': activePair.length === 1 }"
      role="tabpanel"
      aria-live="polite"
    >
      <button
        v-for="({ item, index }) in activePair"
        :key="`${activePairIndex}-${item.id}`"
        type="button"
        class="home-hero-visual-showcase-mobile__card"
        :class="{ 'is-active': item.src && index === activeItemIndex, 'is-empty': !item.src }"
        :aria-label="`${item.title || ariaLabel} ${index + 1}`"
        :aria-pressed="item.src ? index === activeItemIndex : false"
        :disabled="!item.src"
        @click="setActiveItem(index)"
      >
        <HomeHeroVisualShowcaseFigure
          :item="item"
          :loading="index === 0 ? 'eager' : 'lazy'"
          :fetchpriority="index === 0 ? 'high' : 'low'"
          :preload="index === 0 ? { fetchPriority: 'high', media: '(max-width: 1023px)' } : false"
          caption-visibility="sr-only"
        />
      </button>
    </div>

    <div
      v-if="activeItem"
      class="home-hero-visual-showcase-mobile__detail"
      role="region"
      aria-live="polite"
      :aria-label="activeItem.title || ariaLabel"
    >
      <div class="home-hero-visual-showcase-mobile__detail-copy">
        <p v-if="activeItem.title" class="home-hero-visual-showcase-mobile__detail-title">{{ activeItem.title }}</p>
        <p v-if="activeItem.caption" class="home-hero-visual-showcase-mobile__detail-description">
          {{ activeItem.caption }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import HomeHeroVisualShowcaseFigure from '~/components/home/HomeHeroVisualShowcaseFigure.vue'
import {
  HOME_HERO_VISUAL_SHOWCASE_MAXIMUM_ITEM_COUNT,
  type HomeHeroVisualShowcaseItem,
} from '~/types/homeHeroVisualShowcase'

const props = defineProps<{
  items: HomeHeroVisualShowcaseItem[]
  ariaLabel: string
}>()

const mobileItems = computed(() => props.items.slice(0, HOME_HERO_VISUAL_SHOWCASE_MAXIMUM_ITEM_COUNT))
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

const closestConfiguredItemIndex = (preferredIndex: number): number => {
  const configuredIndices = mobileItems.value
    .map((item, index) => item.src ? index : -1)
    .filter((index) => index >= 0)
  if (configuredIndices.length === 0) return 0

  return configuredIndices.reduce((closestIndex, index) => (
    Math.abs(index - preferredIndex) < Math.abs(closestIndex - preferredIndex)
      ? index
      : closestIndex
  ))
}

watch(
  () => mobileItems.value.map((item) => item.src).join('|'),
  () => {
    const length = mobileItems.value.length
    if (length <= 0) {
      activePairIndex.value = 0
      activeItemIndex.value = 0
      return
    }

    const pairCount = Math.ceil(length / 2)
    activePairIndex.value = Math.min(activePairIndex.value, pairCount - 1)
    if (!mobileItems.value[activeItemIndex.value]?.src) {
      activeItemIndex.value = closestConfiguredItemIndex(activePairIndex.value * 2)
      activePairIndex.value = Math.floor(activeItemIndex.value / 2)
    }
  },
  { immediate: true },
)

const setActivePair = (pairIndex: number) => {
  if (pairIndex < 0 || pairIndex >= mobilePairCount.value) return
  activePairIndex.value = pairIndex
  activeItemIndex.value = Math.min(pairIndex * 2, Math.max(0, mobileItems.value.length - 1))
}

const setActiveItem = (index: number) => {
  if (index < 0 || index >= mobileItems.value.length || !mobileItems.value[index]?.src) return
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

.home-hero-visual-showcase-mobile__grid--single {
  grid-template-columns: minmax(0, 1fr);
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

.home-hero-visual-showcase-mobile__card.is-empty {
  cursor: default;
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

</style>
