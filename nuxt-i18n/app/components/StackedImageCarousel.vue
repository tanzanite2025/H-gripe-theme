<template>
  <div class="relative w-full max-w-4xl mx-auto px-0">
    <div
      class="tz-image-carousel relative rounded-3xl pt-[56px] sm:pt-[68px]"
      tabindex="0"
      @keydown.left.prevent="prev"
      @keydown.right.prevent="next"
    >
      <!-- Navigation Dots -->
      <div class="tz-carousel-pagination absolute inset-x-0 top-2 z-40 px-4 sm:top-[14px]">
        <button
          v-for="(_, index) in items"
          :key="index"
          type="button"
          class="tz-carousel-pagination__dot"
          :class="{ 'is-active': index === activeIndex }"
          @click="goTo(index)"
          :aria-label="`Go to slide ${index + 1}`"
          :aria-current="index === activeIndex ? 'true' : undefined"
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
              <StorefrontImage
                v-if="item.src"
                :src="item.src"
                :alt="item.alt || ''"
                :width="item.width"
                :height="item.height"
                class="h-full w-full object-cover rounded-2xl bg-[var(--tz-image-loading-surface)] shadow-2xl"
                preset="gallery"
                :sizes="item.sizes || 'xs:100vw sm:100vw md:768px lg:1024px'"
                :densities="item.densities || '1x 2x'"
                loading="lazy"
                decoding="async"
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
  width?: number
  height?: number
  sizes?: string
  densities?: string
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
