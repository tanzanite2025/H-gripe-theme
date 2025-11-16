<template>
  <section class="carousel-section w-full max-w-[1600px] mx-auto min-[1024px]:aspect-[21/9]:h-[45vh]">
    <div class="relative min-[1024px]:aspect-[21/9]:h-full">
      <div
        ref="track"
        class="flex gap-6 max-md:gap-4 overflow-x-auto scroll-smooth snap-x snap-mandatory [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden pb-2.5 min-[1024px]:aspect-[21/9]:items-center"
      >
        <div
          v-for="(card, i) in extendedItems"
          :key="i"
          :class="[
            'flex-none w-[calc((100%-48px)/3)] max-md:w-full h-[300px] max-md:h-[25vh] min-[1024px]:aspect-[21/9]:h-[calc(45vh-60px)] rounded-3xl max-md:rounded-2xl overflow-hidden bg-[#111111] border border-white snap-center transition-[transform,opacity,box-shadow] duration-[350ms] ease-in-out',
            i === activeIndex 
              ? 'scale-100 opacity-100 shadow-[0_20px_45px_rgba(0,0,0,0.45)]' 
              : 'scale-[0.88] opacity-65'
          ]"
        >
          <div class="w-full h-full bg-gradient-to-br from-[rgba(110,110,233,0.18)] to-[rgba(172,105,238,0.18)] flex items-center justify-center text-white/60 text-sm" aria-label="carousel item placeholder"></div>
        </div>
      </div>

      <!-- 按钮容器 -->
      <div class="button-container flex items-center justify-center gap-2 max-md:gap-1.5">
        <button
          class="w-[72px] max-md:w-[52px] h-[35px] max-md:h-[30px] rounded-full bg-white/10 text-white border border-white inline-flex items-center justify-center leading-none p-0 hover:bg-white/20 transition-colors"
          type="button"
          @click="scrollPrev"
        >
          <span class="sr-only">Prev</span>
          <svg viewBox="0 0 24 24" class="w-[18px] h-[18px] max-md:w-4 max-md:h-4" fill="none" stroke="currentColor" stroke-width="2"><path d="M15 18l-6-6 6-6"/></svg>
        </button>
        <button
          class="w-[72px] max-md:w-[52px] h-[35px] max-md:h-[30px] rounded-full bg-white/10 text-white border border-white inline-flex items-center justify-center leading-none p-0 hover:bg-white/20 transition-colors"
          type="button"
          @click="scrollNext"
        >
          <span class="sr-only">Next</span>
          <svg viewBox="0 0 24 24" class="w-[18px] h-[18px] max-md:w-4 max-md:h-4" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 6l6 6-6 6"/></svg>
        </button>
      </div>

      <div v-if="$slots.footer" class="mt-8 max-md:mt-5 min-[1024px]:aspect-[21/9]:mt-7 w-full flex justify-center">
        <slot name="footer" />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'

// Keep placeholder items; you can later replace the slot content by injecting images/content
const items = ref<Record<string, unknown>[]>([{}, {}, {}, {}])
// extended list: [lastClone, ...items, firstClone]
const extendedItems = computed(() => {
  const arr = items.value
  if (!arr.length) return []
  const first = arr[0]
  const last = arr[arr.length - 1]
  return [last, ...arr, first]
})

const track = ref<HTMLElement | null>(null)
const activeIndex = ref(1)

const getCardElements = () => {
  const el = track.value
  return el ? Array.from(el.querySelectorAll<HTMLElement>('.flex-none')) : []
}

const centerAt = (index: number, behavior: ScrollBehavior = 'smooth', skipNormalize = false) => {
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

  if (skipNormalize || total <= 2) return

  if (clamped === 0) {
    window.setTimeout(() => centerAt(total - 2, 'auto', true), 360)
  } else if (clamped === total - 1) {
    window.setTimeout(() => centerAt(1, 'auto', true), 360)
  }
}

const scrollNext = () => {
  centerAt(activeIndex.value + 1)
}

const scrollPrev = () => {
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
  await nextTick()
  centerAt(1, 'auto')
  const el = track.value
  if (el) {
    el.addEventListener('scroll', handleScroll, { passive: true })
  }
})

onBeforeUnmount(() => {
  const el = track.value
  if (el) {
    el.removeEventListener('scroll', handleScroll)
  }
  if (scrollTimer) {
    window.clearTimeout(scrollTimer)
    scrollTimer = null
  }
})
</script>

<style scoped>
/* 轮播组件上边距 - 使用CSS媒体查询避免SSR跳动 */
.carousel-section {
  margin-top: 80px;
}

@media (max-width: 768px) {
  .carousel-section {
    margin-top: 125px;
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
