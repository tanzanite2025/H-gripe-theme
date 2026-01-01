<template>
  <div class="relative w-full max-w-4xl mx-auto px-0">
    <div
      class="relative rounded-3xl pt-9 sm:pt-[60px]"
      tabindex="0"
      @keydown.left.prevent="prev"
      @keydown.right.prevent="next"
    >
      <!-- Navigation Dots -->
      <div class="absolute inset-x-0 top-2 sm:top-[14px] z-40 flex justify-center gap-2 px-4">
        <button
          v-for="(_, index) in items"
          :key="index"
          type="button"
          class="h-1.5 rounded-full transition-[width,background-color] duration-200"
          :class="index === activeIndex ? 'w-7 bg-sky-400' : 'w-1.5 bg-slate-700 hover:bg-slate-600'"
          @click="goTo(index)"
          :aria-label="`Go to slide ${index + 1}`"
        ></button>
      </div>

      <!-- Carousel Stack -->
      <div class="relative aspect-[16/10] sm:aspect-[16/9] overflow-visible">
        <ul class="absolute inset-0 overflow-visible p-0 m-0 list-none">
          <li
            v-for="(item, index) in items"
            :key="index"
            class="absolute inset-0 origin-center will-change-transform transition-[transform,filter,opacity] duration-[400ms] ease-[cubic-bezier(0.2,0,0.2,1)]"
            :class="cardClass(index)"
          >
            <slot name="card" :item="item" :index="index">
              <img
                v-if="item.src"
                :src="item.src"
                :alt="item.alt || ''"
                class="h-full w-full object-cover rounded-2xl bg-slate-900 shadow-2xl"
                loading="lazy"
              />
              <!-- Caption Overlay -->
              <div
                v-if="item.caption"
                class="absolute bottom-0 inset-x-0 p-6 pt-12 bg-gradient-to-t from-black/90 via-black/60 to-transparent text-white text-center rounded-b-2xl"
              >
                <p class="text-sm sm:text-base font-medium text-slate-100 drop-shadow-md">
                  {{ item.caption }}
                </p>
              </div>
            </slot>
          </li>
        </ul>

        <!-- Prev Button -->
        <button
          type="button"
          class="absolute left-2 sm:left-4 top-1/2 z-50 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-slate-800/80 text-white backdrop-blur transition-all hover:bg-slate-700 hover:scale-110 shadow-lg border border-slate-700/50"
          @click="prev"
          aria-label="Previous image"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" class="opacity-90">
            <path d="M15 19L8 12L15 5" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>

        <!-- Next Button -->
        <button
          type="button"
          class="absolute right-2 sm:right-4 top-1/2 z-50 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-slate-800/80 text-white backdrop-blur transition-all hover:bg-slate-700 hover:scale-110 shadow-lg border border-slate-700/50"
          @click="next"
          aria-label="Next image"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" class="opacity-90">
            <path d="M9 5L16 12L9 19" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

export interface CarouselItem {
  src?: string
  alt?: string
  caption?: string
  [key: string]: any
}

const props = defineProps<{
  items: CarouselItem[]
  modelValue?: number
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number): void
}>()

const activeIndex = ref(props.modelValue ?? 0)

watch(() => props.modelValue, (val) => {
  if (val !== undefined) activeIndex.value = val
})

const updateIndex = (val: number) => {
  activeIndex.value = val
  emit('update:modelValue', val)
}

const next = () => {
  if (props.items.length === 0) return
  const nextVal = (activeIndex.value + 1) % props.items.length
  updateIndex(nextVal)
}

const prev = () => {
  if (props.items.length === 0) return
  const nextVal = (activeIndex.value - 1 + props.items.length) % props.items.length
  updateIndex(nextVal)
}

const goTo = (index: number) => {
  updateIndex(index)
}

const relativeSlot = (index: number) => {
  if (props.items.length === 0) return 0
  return (index - activeIndex.value + props.items.length) % props.items.length
}

const cardClass = (index: number) => {
  const count = props.items.length
  if (count === 0) return ''
  
  const slot = relativeSlot(index)

  if (slot === 0) {
    // Front card
    return 'z-30 opacity-100 translate-y-0 scale-100 brightness-100'
  }

  if (slot === 1) {
    // Second card
    return 'z-20 opacity-80 -translate-y-[8%] scale-[0.92] brightness-[0.7]'
  }

  if (slot === count - 1) {
    // Previous card (hidden left/behind transition)
     return 'z-10 opacity-0 translate-y-[10%] scale-[0.9] brightness-[0.5]'
  }

  // Others (stacked at back)
  return 'z-10 opacity-0 -translate-y-[15%] scale-[0.85] brightness-[0.5]'
}

defineExpose({
  next,
  prev,
  goTo
})
</script>
