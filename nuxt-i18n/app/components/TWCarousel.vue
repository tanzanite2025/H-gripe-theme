<template>
  <section class="carousel-section w-full max-w-[1600px] mx-auto">
    <div class="relative">
      <div
        ref="track"
        class="flex gap-6 max-md:gap-4 overflow-x-auto scroll-smooth snap-x snap-mandatory [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden pb-2.5 items-center"
      >
        <div
          v-for="(slide, slideIndex) in slides"
          :key="slideIndex"
          :class="[
            'flex-none w-full snap-center transition-[transform,opacity,box-shadow] duration-[350ms] ease-in-out',
            slideIndex === activeIndex
              ? 'scale-100 opacity-100 shadow-[0_0_30px_rgba(107,115,255,0.3)]'
              : 'scale-95 opacity-70'
          ]"
        >
          <div class="grid grid-cols-1 lg:grid-cols-3 gap-4 lg:gap-6">
            <div
              v-for="(card, cardIndex) in slide"
              :key="cardIndex"
              class="h-[60vw] lg:h-[400px] rounded-2xl lg:rounded-3xl overflow-hidden bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_4px_14px_-10px_rgba(0,0,0,0.9),0_0_12px_rgba(15,23,42,0.85)]"
            >
              <div
                class="w-full h-full flex items-center justify-center text-white/70 text-sm bg-[radial-gradient(circle_at_top,rgba(56,189,248,0.16),transparent_60%)]"
                aria-label="carousel item placeholder"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <!-- 按钮容器：覆盖在左右两侧，中线位置 -->
      <div class="button-container pointer-events-none absolute inset-y-0 left-0 right-0 flex items-center justify-between px-2 md:px-4">
        <button
          class="pointer-events-auto w-9 h-9 md:w-10 md:h-10 rounded-full bg-black/55 text-white shadow-[0_0_18px_rgba(37,99,235,0.6)] flex items-center justify-center leading-none p-0 hover:bg-black/75 transition-colors"
          type="button"
          :disabled="activeIndex === 0"
          @click="scrollPrev"
        >
          <span class="sr-only">Prev</span>
          <svg viewBox="0 0 24 24" class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2"><path d="M15 18l-6-6 6-6"/></svg>
        </button>
        <button
          class="pointer-events-auto w-9 h-9 md:w-10 md:h-10 rounded-full bg-black/55 text-white shadow-[0_0_18px_rgba(37,99,235,0.6)] flex items-center justify-center leading-none p-0 hover:bg-black/75 transition-colors"
          type="button"
          :disabled="activeIndex === totalSlides - 1"
          @click="scrollNext"
        >
          <span class="sr-only">Next</span>
          <svg viewBox="0 0 24 24" class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 6l6 6-6 6"/></svg>
        </button>
      </div>

      <!-- 分页小圆点指示器（仅作为位置提示） -->
      <div class="pointer-events-none absolute bottom-3 left-1/2 -translate-x-1/2 flex gap-1.5">
        <span
          v-for="(_, idx) in slides"
          :key="idx"
          :class="[
            'w-1.5 h-1.5 rounded-full transition-all',
            idx === activeIndex ? 'bg-white opacity-100 scale-110' : 'bg-white/30 opacity-70 scale-100',
          ]"
        />
      </div>

      <div v-if="$slots.footer" class="mt-8 max-md:mt-5 min-[1024px]:aspect-[21/9]:mt-7 w-full flex justify-center">
        <slot name="footer" />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'

// Placeholder items; you can later replace the slot content by injecting images/content
const items = ref<Record<string, unknown>[]>([{}, {}, {}, {}])

const track = ref<HTMLElement | null>(null)
const activeIndex = ref(0)

const isDesktop = ref(false)

const updateIsDesktop = () => {
  if (typeof window === 'undefined') return
  isDesktop.value = window.innerWidth >= 1024
}

const slides = computed(() => {
  const perSlide = isDesktop.value ? 3 : 1
  const source = items.value
  const result: Record<string, unknown>[][] = []
  if (!source.length) return result
  for (let i = 0; i < source.length; i += perSlide) {
    result.push(source.slice(i, i + perSlide))
  }
  return result
})

const totalSlides = computed(() => slides.value.length)

const getCardElements = () => {
  const el = track.value
  return el ? Array.from(el.querySelectorAll<HTMLElement>('.flex-none')) : []
}

const centerAt = (index: number, behavior: ScrollBehavior = 'smooth') => {
  const el = track.value
  if (!el) return
  const cards = getCardElements()
  const total = cards.length
  if (!total) return

  const clamped = Math.max(0, Math.min(index, total - 1))
  const target = cards[clamped]
  if (!target) return

  const offset = target.offsetLeft - (el.clientWidth - target.offsetWidth) / 2
  activeIndex.value = clamped
  el.scrollTo({ left: offset, behavior })
}

const scrollNext = () => {
  const total = getCardElements().length
  if (activeIndex.value >= total - 1) return
  centerAt(activeIndex.value + 1)
}

const scrollPrev = () => {
  if (activeIndex.value <= 0) return
  centerAt(activeIndex.value - 1)
}

const updateActiveHighlight = () => {
  const el = track.value
  if (!el) return
  const cards = getCardElements()
  if (!cards.length) return
  const center = el.scrollLeft + el.clientWidth / 2
  let closest = activeIndex.value
  let min = Number.POSITIVE_INFINITY
  cards.forEach((card, idx) => {
    const cardCenter = card.offsetLeft + card.offsetWidth / 2
    const distance = Math.abs(cardCenter - center)
    if (distance < min) {
      min = distance
      closest = idx
    }
  })
  activeIndex.value = closest
}

const snapToNearest = () => {
  const el = track.value
  if (!el) return
  const cards = getCardElements()
  if (!cards.length) return
  const center = el.scrollLeft + el.clientWidth / 2
  let closest = activeIndex.value
  let min = Number.POSITIVE_INFINITY
  cards.forEach((card, idx) => {
    const cardCenter = card.offsetLeft + card.offsetWidth / 2
    const distance = Math.abs(cardCenter - center)
    if (distance < min) {
      min = distance
      closest = idx
    }
  })
  centerAt(closest)
}

let scrollTimer: number | null = null

const handleScroll = () => {
  updateActiveHighlight()
  if (scrollTimer) window.clearTimeout(scrollTimer)
  scrollTimer = window.setTimeout(() => {
    snapToNearest()
  }, 140)
}

onMounted(async () => {
  updateIsDesktop()
  await nextTick()
  centerAt(0, 'auto')
  const el = track.value
  if (el) {
    el.addEventListener('scroll', handleScroll, { passive: true })
  }
  if (typeof window !== 'undefined') {
    window.addEventListener('resize', updateIsDesktop)
  }
})

onBeforeUnmount(() => {
  const el = track.value
  if (el) {
    el.removeEventListener('scroll', handleScroll)
  }
  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', updateIsDesktop)
  }
  if (scrollTimer) {
    window.clearTimeout(scrollTimer)
    scrollTimer = null
  }
})
</script>

<style scoped>
/* 轮播组件上边距 - 使用CSS媒体查询避免SSR跳动 */
/* 需要留出 SiteHeader 的高度空间 */
.carousel-section {
  margin-top: 100px;
}

@media (max-width: 768px) {
  .carousel-section {
    margin-top: 115px;
  }
}

/* 按钮容器上边距 */
.button-container {
  margin-top: 1px;
}

@media (max-width: 768px) {
  .button-container {
    margin-top: -3px;
  }
}
</style>
