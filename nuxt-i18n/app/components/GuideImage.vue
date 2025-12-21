<template>
  <figure
    class="guide-image"
    :class="{ 'guide-image--clickable': props.zoomOnClick }"
    @click="handleClick"
    :tabindex="props.zoomOnClick ? 0 : -1"
    :role="props.zoomOnClick ? 'button' : undefined"
  >
    <div class="guide-image__frame">
      <img
        :src="props.src"
        :alt="props.alt || ''"
        class="guide-image__img"
        loading="lazy"
      />
    </div>
    <figcaption
      v-if="props.caption"
      class="guide-image__caption"
    >
      {{ props.caption }}
    </figcaption>
  </figure>

  <Teleport to="body" v-if="props.zoomOnClick && isOpen">
    <div class="guide-image__overlay" @click.self="close">
      <button type="button" class="guide-image__close" @click="close">
        ×
      </button>
      <img
        :src="props.src"
        :alt="props.alt || ''"
        class="guide-image__overlay-img"
      />
    </div>
  </Teleport>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'

interface Props {
  src: string
  alt?: string
  zoomOnClick?: boolean
  caption?: string
}

const props = withDefaults(defineProps<Props>(), {
  zoomOnClick: false,
})

onMounted(() => {
  if (!props.alt) {
    // eslint-disable-next-line no-console
    console.warn('[GuideImage] Missing alt text for image with src:', props.src)
  }
})

const isOpen = ref(false)

const handleClick = () => {
  if (!props.zoomOnClick) return
  isOpen.value = true
}

const close = () => {
  isOpen.value = false
}
</script>

<style scoped>
.guide-image {
  margin: 0;
}

.guide-image__frame {
  position: relative;
  width: 100%;
  padding-top: 56.25%; /* 16:9 aspect ratio */
}

.guide-image__img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  border-radius: 0.5rem;
  border: none;
  box-shadow: 2px 2px 4px rgba(0, 0, 0, 0.9);
  object-fit: contain;
}

.guide-image--clickable {
  cursor: zoom-in;
}

.guide-image__overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.guide-image__overlay-img {
  max-width: 90vw;
  max-height: 90vh;
  object-fit: contain;
  border-radius: 0.75rem;
  box-shadow: 0 18px 60px rgba(0, 0, 0, 1);
}

.guide-image__close {
  position: absolute;
  top: 1.5rem;
  right: 1.5rem;
  background: transparent;
  border: none;
  color: #e5e7eb;
  font-size: 1.4rem;
  cursor: pointer;
}

.guide-image__caption {
  margin-top: 0.35rem;
  font-size: 0.8rem;
  line-height: 1.4;
  color: rgba(148, 163, 184, 0.9);
}
</style>
