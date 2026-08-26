<template>
  <div class="home-hero-visual-showcase-desktop" :aria-label="ariaLabel">
    <div class="home-hero-visual-showcase-desktop__content">
      <div class="home-hero-visual-showcase-desktop__stage">
        <button
          v-for="(item, index) in desktopItems"
          :key="item.id"
          type="button"
          class="home-hero-visual-showcase-desktop__card"
          :class="{ 'is-active': index === activeIndex }"
          :style="cardStyle(index)"
          :aria-label="`${item.title} ${index + 1}`"
          :aria-pressed="index === activeIndex"
          @click="setActiveIndex(index)"
        >
          <HomeHeroVisualShowcaseFigure
            :item="item"
            :loading="'eager'"
            :fetchpriority="index === activeIndex ? 'high' : 'low'"
            sizes="xs:100vw sm:50vw lg:22vw xl:22vw"
            caption-visibility="sr-only"
            class="home-hero-visual-showcase-desktop__figure"
          />
        </button>
      </div>

      <div
        v-if="activeItem"
        class="home-hero-visual-showcase-desktop__detail"
        role="region"
        aria-live="polite"
        :aria-label="activeItem.title"
      >
        <div class="home-hero-visual-showcase-desktop__detail-copy">
          <h3 class="home-hero-visual-showcase-desktop__detail-title">{{ activeItem.title }}</h3>
          <p v-if="activeItem.caption" class="home-hero-visual-showcase-desktop__detail-description">
            {{ activeItem.caption }}
          </p>
        </div>
        <NuxtLink
          v-if="activeItem.targetUrl && activeItem.targetLabel"
          :to="activeItem.targetUrl"
          class="home-hero-visual-showcase-desktop__detail-link"
        >
          {{ activeItem.targetLabel }}
          <Icon name="lucide:arrow-up-right" aria-hidden="true" />
        </NuxtLink>
      </div>

      <div class="tz-carousel-pagination home-hero-visual-showcase-desktop__pagination" role="tablist" :aria-label="ariaLabel">
        <button
          v-for="(item, index) in desktopItems"
          :key="`${item.id}-dot`"
          type="button"
          class="tz-carousel-pagination__dot"
          :class="{ 'is-active': index === activeIndex }"
          :aria-label="`${ariaLabel} ${index + 1}`"
          :aria-selected="index === activeIndex"
          role="tab"
          @click="setActiveIndex(index)"
        ></button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import HomeHeroVisualShowcaseFigure from '~/components/home/HomeHeroVisualShowcaseFigure.vue'
import type { HomeHeroVisualShowcaseItem } from '~/types/homeHeroVisualShowcase'

type FanSlot = {
  x: number
  y: number
  widthFactor: number
  rotation: number
  scale: number
  opacity: number
  zIndex: number
  blur: number
}

const FAN_SLOTS: FanSlot[] = [
  { x: 14, y: 26, widthFactor: 0.62, rotation: -30, scale: 0.82, opacity: 0.42, zIndex: 1, blur: 0.6 },
  { x: 22, y: 15, widthFactor: 0.68, rotation: -21, scale: 0.87, opacity: 0.58, zIndex: 2, blur: 0.35 },
  { x: 30, y: 7, widthFactor: 0.78, rotation: -13, scale: 0.92, opacity: 0.74, zIndex: 3, blur: 0 },
  { x: 40, y: 2, widthFactor: 0.88, rotation: -5, scale: 0.97, opacity: 0.9, zIndex: 4, blur: 0 },
  { x: 50, y: 0, widthFactor: 1, rotation: 0, scale: 1.06, opacity: 1, zIndex: 6, blur: 0 },
  { x: 60, y: 2, widthFactor: 0.88, rotation: 5, scale: 0.97, opacity: 0.9, zIndex: 5, blur: 0 },
  { x: 70, y: 7, widthFactor: 0.78, rotation: 13, scale: 0.92, opacity: 0.74, zIndex: 4, blur: 0 },
  { x: 78, y: 15, widthFactor: 0.68, rotation: 21, scale: 0.87, opacity: 0.58, zIndex: 2, blur: 0.35 },
  { x: 86, y: 26, widthFactor: 0.62, rotation: 30, scale: 0.82, opacity: 0.42, zIndex: 1, blur: 0.6 },
]
const CENTER_FAN_SLOT = FAN_SLOTS[Math.floor(FAN_SLOTS.length / 2)]!

const props = defineProps<{
  items: HomeHeroVisualShowcaseItem[]
  ariaLabel: string
}>()

const desktopItems = computed(() => props.items.slice(0, FAN_SLOTS.length))
const centerIndex = Math.min(4, Math.max(0, desktopItems.value.length - 1))
const activeIndex = ref(centerIndex)
const activeItem = computed(() => desktopItems.value[activeIndex.value] ?? null)
const ACTIVE_CARD_WIDTH_MULTIPLIER = 1.15
const ACTIVE_CARD_VERTICAL_LIFT = 2

watch(
  () => desktopItems.value.length,
  (length) => {
    if (length <= 0) return
    activeIndex.value = Math.min(activeIndex.value, length - 1)
  },
  { immediate: true },
)

const relativeIndex = (index: number): number => {
  const total = desktopItems.value.length
  if (total <= 0) return 0

  const delta = (index - activeIndex.value + total) % total
  return delta > total / 2 ? delta - total : delta
}

const slotForRelativeIndex = (relativeIndexValue: number): FanSlot => {
  const clamped = Math.max(-(FAN_SLOTS.length - 1) / 2, Math.min((FAN_SLOTS.length - 1) / 2, relativeIndexValue))
  const slotIndex = Math.round(clamped) + Math.floor(FAN_SLOTS.length / 2)
  return FAN_SLOTS[slotIndex] ?? CENTER_FAN_SLOT
}

const cardStyle = (index: number): Record<string, string | number> => {
  const slot = slotForRelativeIndex(relativeIndex(index))
  const isActive = index === activeIndex.value
  const scale = slot.scale * (isActive ? 1.05 : 1)

  return {
    left: `${slot.x}%`,
    top: `${slot.y - (isActive ? ACTIVE_CARD_VERTICAL_LIFT : 0)}%`,
    width: `calc(var(--home-hero-fan-center-card-width) * ${slot.widthFactor} * ${isActive ? ACTIVE_CARD_WIDTH_MULTIPLIER : 1})`,
    zIndex: isActive ? 10 : slot.zIndex,
    opacity: slot.opacity,
    filter: `blur(${slot.blur}px) saturate(${isActive ? 1.08 : 0.96})`,
    transform: `translate(-50%, 0) rotate(${slot.rotation}deg) scale(${scale})`,
  }
}

const setActiveIndex = (index: number): void => {
  if (index < 0 || index >= desktopItems.value.length) return
  activeIndex.value = index
}
</script>

<style scoped>
.home-hero-visual-showcase-desktop {
  --home-hero-fan-center-card-width: clamp(10rem, 23cqw, 14.75rem);

  position: relative;
  width: 100%;
  height: clamp(34rem, 39vw, 38rem);
  container-type: inline-size;
  overflow: hidden;
  isolation: isolate;
  display: flex;
  align-items: center;
  justify-content: center;
}

.home-hero-visual-showcase-desktop__content {
  position: relative;
  display: flex;
  width: 100%;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
}

.home-hero-visual-showcase-desktop__stage {
  position: relative;
  width: 100%;
  height: clamp(22.5rem, 26vw, 24.5rem);
  transform: translateY(2rem);
  overflow: visible;
}

.home-hero-visual-showcase-desktop__detail {
  display: flex;
  width: min(100%, 30rem);
  min-height: 5.4rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin: 0.5rem auto 0;
  padding: 0.9rem 1.1rem;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 0.85rem;
  background: rgba(255, 255, 255, 0.96);
  color: #111318;
  box-shadow: 0 14px 30px rgba(0, 0, 0, 0.28);
  text-align: left;
}

.home-hero-visual-showcase-desktop__detail-copy {
  min-width: 0;
}

.home-hero-visual-showcase-desktop__detail-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 800;
  line-height: 1.2;
}

.home-hero-visual-showcase-desktop__detail-description {
  margin: 0.35rem 0 0;
  color: rgba(17, 19, 24, 0.7);
  font-size: 0.82rem;
  line-height: 1.45;
}

.home-hero-visual-showcase-desktop__detail-link {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.25rem;
  color: #111318;
  font-size: 0.75rem;
  font-weight: 800;
  text-decoration: none;
  white-space: nowrap;
}

.home-hero-visual-showcase-desktop__detail-link:hover {
  color: #4b7f1f;
}

.home-hero-visual-showcase-desktop__detail-link :deep(svg) {
  width: 0.9rem;
  height: 0.9rem;
}

.home-hero-visual-showcase-desktop__card {
  position: absolute;
  display: block;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
  transform-origin: 50% 100%;
  transition:
    transform 420ms cubic-bezier(0.22, 1, 0.36, 1),
    opacity 280ms ease,
    filter 280ms ease;
  will-change: transform, opacity, filter;
}

.home-hero-visual-showcase-desktop__card:focus-visible {
  outline: 2px solid rgba(5, 150, 105, 0.95);
  outline-offset: 4px;
  border-radius: 1rem;
}

.home-hero-visual-showcase-desktop__card.is-active {
  cursor: default;
}

.home-hero-visual-showcase-desktop__figure {
  display: block;
  width: 100%;
}

.home-hero-visual-showcase-desktop__pagination {
  position: relative;
  margin-top: 0.7rem;
}

@media (max-width: 1100px) {
  .home-hero-visual-showcase-desktop__detail {
    width: min(100%, 26rem);
  }

  .home-hero-visual-showcase-desktop__detail-description {
    font-size: 0.76rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .home-hero-visual-showcase-desktop__card,
  .home-hero-visual-showcase-desktop__pagination .tz-carousel-pagination__dot {
    transition: none;
  }
}
</style>
